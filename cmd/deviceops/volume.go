package deviceops

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/chijiajian/zstack-cli-go/cmd/vmops"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

const volumes = "v1/volumes"

func CreateVolume(cmd *cobra.Command, body param.CreateDataVolumeParam) error {
	p := body.Params
	if p.ResourceUuid == nil || strings.TrimSpace(p.Name) == "" || (p.DiskSize == nil) == (p.DiskOfferingUuid == nil) || p.DiskSize != nil && *p.DiskSize <= 0 {
		return vmops.Invalid("create requires a name, resource UUID, and exactly one positive size or disk offering")
	}
	return run(cmd, "create-volume", *p.ResourceUuid, func(op operation) (string, bool, error) {
		id := *p.ResourceUuid
		if err := op.absent(volumes, id); err != nil {
			return id, false, err
		}
		var size int64
		if p.DiskSize != nil {
			size = *p.DiskSize
		} else {
			offering, err := op.exact("v1/disk-offerings", *p.DiskOfferingUuid)
			if err != nil {
				return id, false, err
			}
			size = offering.DiskSize
			if size <= 0 {
				return id, false, unverified("disk offering size could not be verified")
			}
		}
		if p.PrimaryStorageUuid != nil {
			if _, err := op.exact("v1/primary-storage", *p.PrimaryStorageUuid); err != nil {
				return id, false, err
			}
		}
		if err := op.api.Mutate(op.ctx, http.MethodPost, volumes+"/data", body, nil); err != nil {
			return id, false, err
		}
		return id, false, op.checkVolume(id, func(v resource) bool {
			return v.Type == "Data" && v.Name == p.Name && v.VmInstanceUuid == "" &&
				(v.Status == "Ready" || v.Status == "NotInstantiated") && v.Size == size &&
				(p.DiskOfferingUuid == nil || v.DiskOfferingUuid == *p.DiskOfferingUuid) &&
				(p.PrimaryStorageUuid == nil || v.PrimaryStorageUuid == *p.PrimaryStorageUuid)
		})
	})
}

func (op operation) checkVolume(id string, check func(resource) bool) error {
	v, err := op.exact(volumes, id)
	if err != nil || !check(v) {
		return unverified("requested volume state could not be verified")
	}
	return nil
}

func Volume(cmd *cobra.Command, action, id, vmID string, size int64) error {
	if (action == "attach" || action == "detach") && !idPattern.MatchString(vmID) {
		return vmops.Invalid("attach and detach require an exact VM UUID")
	}
	if action == "resize" && size <= 0 {
		return vmops.Invalid("resize requires a positive total size")
	}
	return run(cmd, action+"-volume", id, func(op operation) (string, bool, error) {
		v, err := op.exact(volumes, id)
		if err != nil {
			return id, false, err
		}
		if v.Type != "Data" && !(action == "resize" && v.Type == "Root") {
			return id, false, conflict("the operation requires a data volume")
		}
		switch action {
		case "delete":
			if v.Status == "Deleted" {
				return id, true, nil
			}
			if v.VmInstanceUuid != "" || v.Status != "Ready" && v.Status != "NotInstantiated" {
				return id, false, conflict("delete requires a detached, available data volume")
			}
			if err := op.policy("volume"); err != nil {
				return id, false, err
			}
			err = op.api.Delete(op.ctx, volumes+"/"+id, "Permissive")
			if err == nil {
				err = op.checkVolume(id, func(after resource) bool { return after.Status == "Deleted" && after.VmInstanceUuid == "" })
			}
		case "recover":
			if v.Status != "Deleted" {
				return id, false, conflict("recover requires a Deleted data volume")
			}
			err = op.api.Mutate(op.ctx, http.MethodPut, volumes+"/"+id+"/actions", param.RecoverDataVolumeParam{}, nil)
			if err == nil {
				err = op.checkVolume(id, func(after resource) bool {
					return (after.Status == "Ready" || after.Status == "NotInstantiated") && after.VmInstanceUuid == ""
				})
			}
		case "resize":
			if v.Status != "Ready" || v.Size <= 0 || size <= v.Size {
				return id, false, conflict("resize requires a Ready volume and a total size greater than its current size")
			}
			if v.VmInstanceUuid != "" {
				if _, err := op.vm(v.VmInstanceUuid, true); err != nil {
					return id, false, err
				}
			}
			path := volumes + "/data/resize/" + id + "/actions"
			var body any = param.ResizeDataVolumeParam{Params: param.ResizeDataVolumeParamDetail{Size: size}}
			if v.Type == "Root" {
				if v.VmInstanceUuid == "" {
					return id, false, conflict("root volume must belong to a stopped VM")
				}
				path = volumes + "/resize/" + id + "/actions"
				body = param.ResizeRootVolumeParam{Params: param.ResizeRootVolumeParamDetail{Size: size, VmInstanceUuid: &v.VmInstanceUuid}}
			}
			err = op.api.Mutate(op.ctx, http.MethodPut, path, body, nil)
			if err == nil {
				err = op.checkVolume(id, func(after resource) bool {
					return after.Type == v.Type && after.VmInstanceUuid == v.VmInstanceUuid && after.Size == size
				})
			}
		case "attach", "detach":
			return op.mount(action, v, vmID)
		default:
			err = vmops.Invalid("unsupported volume operation")
		}
		return id, false, err
	})
}

func (op operation) mount(action string, v resource, vmID string) (string, bool, error) {
	id := v.UUID
	if _, err := op.vm(vmID, false); err != nil {
		return id, false, err
	}
	if v.Status != "Ready" && v.Status != "NotInstantiated" {
		return id, false, conflict("volume is not available for attachment changes")
	}
	if v.VmInstanceUuid != "" && v.VmInstanceUuid != vmID {
		return id, false, conflict("volume belongs to a different VM")
	}
	if action == "attach" && v.VmInstanceUuid == vmID || action == "detach" && v.VmInstanceUuid == "" {
		return id, true, nil
	}
	expected := vmID
	path := volumes + "/" + id + "/vm-instances"
	var err error
	if action == "attach" {
		err = op.api.Mutate(op.ctx, http.MethodPost, path+"/"+vmID, param.AttachDataVolumeToVmParam{}, nil)
	} else {
		expected = ""
		err = op.api.DeleteWithQuery(op.ctx, path, url.Values{"vmUuid": {vmID}})
	}
	if err == nil {
		err = op.checkVolume(id, func(after resource) bool { return after.Type == "Data" && after.VmInstanceUuid == expected })
	}
	return id, false, err
}
