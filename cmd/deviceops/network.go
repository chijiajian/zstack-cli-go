package deviceops

import (
	"bytes"
	"net"
	"net/http"
	"net/netip"
	"net/url"

	"github.com/chijiajian/zstack-cli-go/cmd/vmops"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

const (
	nics   = "v1/vm-instances/nics"
	cdroms = "v1/vm-instances/cdroms"
)

func AttachNIC(cmd *cobra.Command, vmID, networkID string, body param.AttachL3NetworkToVmParam) error {
	if !idPattern.MatchString(networkID) {
		return vmops.Invalid("an exact L3 network UUID is required")
	}
	var requestedMAC net.HardwareAddr
	if body.Params.CustomMac != nil {
		mac, err := net.ParseMAC(*body.Params.CustomMac)
		if err != nil || len(mac) != 6 {
			return vmops.Invalid("a valid 48-bit MAC address is required")
		}
		requestedMAC = mac
		canonical := mac.String()
		body.Params.CustomMac = &canonical
	}
	var requestedIP netip.Addr
	if body.Params.StaticIp != nil {
		ip, err := netip.ParseAddr(*body.Params.StaticIp)
		if err != nil || ip.Zone() != "" {
			return vmops.Invalid("a valid static IP address without a zone is required")
		}
		requestedIP = ip.Unmap()
	}
	return run(cmd, "attach-nic", vmID, func(op operation) (string, bool, error) {
		if _, err := op.vm(vmID, false); err != nil {
			return "", false, err
		}
		if _, err := op.exact("v1/l3-networks", networkID); err != nil {
			return "", false, err
		}
		before, err := op.read(nics, 512, "vmInstanceUuid="+vmID, "l3NetworkUuid="+networkID)
		if err != nil {
			return "", false, err
		}
		if len(before) >= 512 {
			return "", false, conflict("cannot safely enumerate adapters on this network")
		}
		seen := make(map[string]bool, len(before))
		for _, nic := range before {
			seen[nic.UUID] = true
		}
		var after struct {
			UUID   string
			VmNics []resource
		}
		path := "v1/vm-instances/" + vmID + "/l3-networks/" + networkID
		if err := op.api.Mutate(op.ctx, http.MethodPost, path, body, &after); err != nil {
			return "", false, err
		}
		if after.UUID != vmID {
			return "", false, unverified("adapter attachment returned an unexpected VM")
		}
		var added []resource
		for _, nic := range after.VmNics {
			if nic.VmInstanceUuid == vmID && nic.L3NetworkUuid == networkID && !seen[nic.UUID] {
				added = append(added, nic)
			}
		}
		if len(added) != 1 {
			return "", false, unverified("new network adapter identity was ambiguous")
		}
		nic, err := op.exact(nics, added[0].UUID)
		if err != nil || nic.VmInstanceUuid != vmID || nic.L3NetworkUuid != networkID {
			return "", false, unverified("new network adapter was not verified")
		}
		if requestedMAC != nil {
			actual, err := net.ParseMAC(nic.Mac)
			if err != nil || !bytes.Equal(requestedMAC, actual) {
				return "", false, unverified("requested network adapter MAC address was not verified")
			}
		}
		if requestedIP.IsValid() && !hasNICAddress(nic, requestedIP) {
			return "", false, unverified("requested network adapter static IP was not verified")
		}
		return nic.UUID, false, nil
	})
}

func hasNICAddress(nic resource, requested netip.Addr) bool {
	matches := func(value string) bool {
		ip, err := netip.ParseAddr(value)
		return err == nil && ip.Unmap() == requested
	}
	if matches(nic.Ip) {
		return true
	}
	// On a dual-stack NIC, the requested address may be an additional used IP.
	for _, used := range nic.UsedIps {
		if (used.L3NetworkUuid == "" || used.L3NetworkUuid == nic.L3NetworkUuid) && matches(used.Ip) {
			return true
		}
	}
	return false
}

func DetachNIC(cmd *cobra.Command, vmID, nicID, networkID string) error {
	if !idPattern.MatchString(vmID) || networkID != "" && !idPattern.MatchString(networkID) {
		return vmops.Invalid("detach requires an exact VM UUID and, if supplied, an exact L3 UUID")
	}
	return run(cmd, "detach-nic", nicID, func(op operation) (string, bool, error) {
		if _, err := op.vm(vmID, false); err != nil {
			return nicID, false, err
		}
		nic, err := op.exact(nics, nicID)
		if err != nil {
			return nicID, false, err
		}
		if nic.VmInstanceUuid != vmID || nic.L3NetworkUuid == "" || networkID != "" && nic.L3NetworkUuid != networkID {
			return nicID, false, conflict("adapter does not belong to the specified VM and network")
		}
		if err := op.api.Delete(op.ctx, nics+"/"+nicID, "Permissive"); err != nil {
			return nicID, false, err
		}
		return nicID, false, op.removed(nics, nicID)
	})
}

func SetMedia(cmd *cobra.Command, vmID, driveID, isoID string, eject bool) error {
	if !idPattern.MatchString(vmID) || !eject && !idPattern.MatchString(isoID) || eject && isoID != "" && !idPattern.MatchString(isoID) {
		return vmops.Invalid("media changes require exact VM and drive UUIDs and a valid ISO UUID when supplied")
	}
	action := "attach-iso"
	if eject {
		action = "detach-iso"
	}
	return run(cmd, action, driveID, func(op operation) (string, bool, error) {
		if _, err := op.vm(vmID, false); err != nil {
			return driveID, false, err
		}
		drive, err := op.exact(cdroms, driveID)
		if err != nil {
			return driveID, false, err
		}
		if drive.VmInstanceUuid != vmID {
			return driveID, false, conflict("drive does not belong to the specified VM")
		}
		if eject {
			if drive.IsoUuid == "" {
				return driveID, true, nil
			}
			if isoID != "" && isoID != drive.IsoUuid {
				return driveID, false, conflict("drive media changed")
			}
			matches, err := op.read(cdroms, 2, "vmInstanceUuid="+vmID, "isoUuid="+drive.IsoUuid)
			if err != nil {
				return driveID, false, err
			}
			if len(matches) != 1 || matches[0].UUID != driveID || matches[0].VmInstanceUuid != vmID || matches[0].IsoUuid != drive.IsoUuid {
				return driveID, false, conflict("ISO is not uniquely associated with the selected drive")
			}
			if err := op.eject(vmID, drive.IsoUuid); err != nil {
				return driveID, false, err
			}
			isoID = ""
		} else {
			if drive.IsoUuid == isoID {
				return driveID, true, nil
			}
			if drive.IsoUuid != "" {
				return driveID, false, conflict("eject existing media before inserting a new ISO")
			}
			image, err := op.exact("v1/images", isoID)
			if err != nil {
				return driveID, false, err
			}
			if image.MediaType != "ISO" || image.Status != "Ready" || image.State != "Enabled" {
				return driveID, false, conflict("selected image must be an enabled, Ready ISO")
			}
			body := param.AttachIsoToVmInstanceParam{BaseParam: param.BaseParam{SystemTags: []string{"cdromUuid::" + driveID}}}
			path := "v1/vm-instances/" + vmID + "/iso/" + isoID
			if err := op.api.Mutate(op.ctx, http.MethodPost, path, body, nil); err != nil {
				return driveID, false, err
			}
		}
		after, err := op.exact(cdroms, driveID)
		if err != nil || after.VmInstanceUuid != vmID || after.IsoUuid != isoID {
			return driveID, false, unverified("selected drive media change was not verified")
		}
		return driveID, false, nil
	})
}

func (op operation) eject(vmID, isoID string) error {
	if !idPattern.MatchString(vmID) || !idPattern.MatchString(isoID) {
		return vmops.Invalid("valid VM and ISO UUIDs are required")
	}
	return op.api.DeleteWithQuery(op.ctx, "v1/vm-instances/"+vmID+"/iso", url.Values{"isoUuid": {isoID}})
}
