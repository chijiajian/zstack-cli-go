package deviceops

import (
	"net/http"
	"strings"

	"github.com/chijiajian/zstack-cli-go/cmd/vmops"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

const snapshots = "v1/volume-snapshots"

func CreateSnapshot(cmd *cobra.Command, volumeID string, body param.CreateVolumeSnapshotParam) error {
	p := body.Params
	if p.ResourceUuid == nil || !idPattern.MatchString(volumeID) || strings.TrimSpace(p.Name) == "" {
		return vmops.Invalid("create requires a snapshot name, resource UUID, and exact volume UUID")
	}
	return run(cmd, "create-snapshot", *p.ResourceUuid, func(op operation) (string, bool, error) {
		id := *p.ResourceUuid
		if err := op.absent(snapshots, id); err != nil {
			return id, false, err
		}
		v, err := op.exact(volumes, volumeID)
		if err != nil {
			return id, false, err
		}
		if v.Status != "Ready" || v.Type != "Data" && v.Type != "Root" {
			return id, false, conflict("snapshot requires a Ready volume")
		}
		if err := op.api.Mutate(op.ctx, http.MethodPost, volumes+"/"+volumeID+"/volume-snapshots", body, nil); err != nil {
			return id, false, err
		}
		created, err := op.exact(snapshots, id)
		if err != nil || created.VolumeUuid != volumeID || created.Status != "Ready" || created.Name != p.Name {
			return id, false, unverified("snapshot creation was not verified")
		}
		return id, false, nil
	})
}

func Snapshot(cmd *cobra.Command, action, id, volumeID string) error {
	if !idPattern.MatchString(volumeID) {
		return vmops.Invalid("an exact parent volume UUID is required")
	}
	return run(cmd, action+"-snapshot", id, func(op operation) (string, bool, error) {
		s, err := op.exact(snapshots, id)
		if err != nil {
			return id, false, err
		}
		if s.VolumeUuid != volumeID || s.Status != "Ready" {
			return id, false, conflict("snapshot must be Ready and belong to the specified volume")
		}
		switch action {
		case "revert":
			v, err := op.exact(volumes, volumeID)
			if err != nil {
				return id, false, err
			}
			if v.Status != "Ready" {
				return id, false, conflict("snapshot restore requires a Ready volume")
			}
			if v.VmInstanceUuid != "" {
				if _, err := op.vm(v.VmInstanceUuid, true); err != nil {
					return id, false, err
				}
			}
			err = op.api.Mutate(op.ctx, http.MethodPut, snapshots+"/"+id+"/actions", param.RevertVolumeFromSnapshotParam{}, nil)
			if err == nil {
				err = op.checkVolume(volumeID, func(after resource) bool { return after.Status == "Ready" && after.VmInstanceUuid == v.VmInstanceUuid })
			}
			return id, false, err
		case "delete":
			children, err := op.read(snapshots, 1, "parentUuid="+id)
			if err != nil {
				return id, false, err
			}
			if len(children) != 0 {
				return id, false, conflict("delete requires a leaf snapshot")
			}
			if err := op.api.Delete(op.ctx, snapshots+"/"+id, "Permissive"); err != nil {
				return id, false, err
			}
			return id, false, op.removed(snapshots, id)
		default:
			return id, false, vmops.Invalid("unsupported snapshot operation")
		}
	})
}
