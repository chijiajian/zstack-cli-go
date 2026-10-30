// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type instanceTargetRequest struct {
	Instance string      `json:"instance" yaml:"instance"`
	Target   string      `json:"target" yaml:"target"`
	Params   interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

var attachL3Cmd = &cobra.Command{
	Use:   "attach-l3 <name-or-uuid>",
	Short: "Attach an L3 network to a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l3Network, _ := cmd.Flags().GetString("l3-network")
		if l3Network == "" {
			return fmt.Errorf("required flag --l3-network not set")
		}

		params := param.AttachL3NetworkToVmParam{
			Params: param.AttachL3NetworkToVmParamDetail{
				StaticIp:   stringPtrIfNotEmptyFlag(cmd, "static-ip"),
				DriverType: stringPtrIfNotEmptyFlag(cmd, "driver-type"),
				CustomMac:  stringPtrIfNotEmptyFlag(cmd, "mac"),
			},
		}

		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceTargetRequest{Instance: args[0], Target: l3Network, Params: params})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}
		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", l3Network)
		if err != nil {
			return err
		}

		result, err := cli.AttachL3NetworkToVm(vm.UUID, l3.UUID, params)
		if err != nil {
			return err
		}
		return printInstanceResult(cmd, fmt.Sprintf("instance/%s attached l3-network/%s", result.Name, l3.Name), result)
	},
}

var detachL3Cmd = &cobra.Command{
	Use:   "detach-l3 <name-or-uuid>",
	Short: "Detach an L3 network from a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l3Network, _ := cmd.Flags().GetString("l3-network")
		if l3Network == "" {
			return fmt.Errorf("required flag --l3-network not set")
		}
		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceTargetRequest{Instance: args[0], Target: l3Network})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}
		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", l3Network)
		if err != nil {
			return err
		}

		vmInventory, err := cli.GetVmInstance(vm.UUID)
		if err != nil {
			return err
		}
		nicUUID := ""
		for _, nic := range vmInventory.VmNics {
			if nic.L3NetworkUuid == l3.UUID {
				nicUUID = nic.UUID
				break
			}
		}
		if nicUUID == "" {
			return fmt.Errorf("instance %q has no nic on l3-network %q", vm.Name, l3.Name)
		}

		if err := cli.DetachL3NetworkFromVm(nicUUID, param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("instance/%s detached l3-network/%s\n", vm.Name, l3.Name)
		return nil
	},
}

var attachIsoCmd = &cobra.Command{
	Use:   "attach-iso <name-or-uuid>",
	Short: "Attach an ISO image to a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		iso, _ := cmd.Flags().GetString("iso")
		if iso == "" {
			return fmt.Errorf("required flag --iso not set")
		}
		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceTargetRequest{Instance: args[0], Target: iso})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}
		image, err := zstackclient.ResolveOneResource(cli, "image", iso)
		if err != nil {
			return err
		}

		result, err := cli.AttachIsoToVmInstance(vm.UUID, image.UUID, param.AttachIsoToVmInstanceParam{})
		if err != nil {
			return err
		}
		return printInstanceResult(cmd, fmt.Sprintf("instance/%s attached iso/%s", result.Name, image.Name), result)
	},
}

var detachIsoCmd = &cobra.Command{
	Use:   "detach-iso <name-or-uuid>",
	Short: "Detach an ISO image from a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		iso, _ := cmd.Flags().GetString("iso")
		if iso == "" {
			return fmt.Errorf("required flag --iso not set")
		}
		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceTargetRequest{Instance: args[0], Target: iso})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}
		image, err := zstackclient.ResolveOneResource(cli, "image", iso)
		if err != nil {
			return err
		}

		vmInventory, err := cli.GetVmInstance(vm.UUID)
		if err != nil {
			return err
		}
		found := false
		for _, cdrom := range vmInventory.VmCdRoms {
			if cdrom.IsoUuid == image.UUID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("instance %q has no cdrom using iso %q", vm.Name, image.Name)
		}

		if err := cli.DetachIsoFromVmInstance(vm.UUID, param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("instance/%s detached iso/%s\n", vm.Name, image.Name)
		return nil
	},
}

var attachVolumeCmd = &cobra.Command{
	Use:   "attach-volume <name-or-uuid>",
	Short: "Attach a data volume to a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeName, _ := cmd.Flags().GetString("volume")
		if volumeName == "" {
			return fmt.Errorf("required flag --volume not set")
		}
		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceTargetRequest{Instance: args[0], Target: volumeName})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}
		volume, err := zstackclient.ResolveOneResource(cli, "volume", volumeName)
		if err != nil {
			return err
		}

		result, err := cli.AttachDataVolumeToVm(volume.UUID, vm.UUID, param.AttachDataVolumeToVmParam{})
		if err != nil {
			return err
		}
		return printInstanceResult(cmd, fmt.Sprintf("instance/%s attached volume/%s", vm.Name, result.Name), result)
	},
}

var detachVolumeCmd = &cobra.Command{
	Use:   "detach-volume <name-or-uuid>",
	Short: "Detach a data volume from a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeName, _ := cmd.Flags().GetString("volume")
		if volumeName == "" {
			return fmt.Errorf("required flag --volume not set")
		}
		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceTargetRequest{Instance: args[0], Target: volumeName})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}
		volume, err := zstackclient.ResolveOneResource(cli, "volume", volumeName)
		if err != nil {
			return err
		}

		if _, err := cli.DetachDataVolumeFromVm(volume.UUID, param.DetachDataVolumeFromVmParam{Params: param.DetachDataVolumeFromVmParamDetail{VmUuid: &vm.UUID}}); err != nil {
			return err
		}
		fmt.Printf("instance/%s detached volume/%s\n", vm.Name, volume.Name)
		return nil
	},
}

func init() {
	attachL3Cmd.Flags().String("l3-network", "", "L3 network name or UUID")
	attachL3Cmd.Flags().String("static-ip", "", "Static IP address for the new NIC")
	attachL3Cmd.Flags().String("driver-type", "", "NIC driver type")
	attachL3Cmd.Flags().String("mac", "", "Custom MAC address")
	detachL3Cmd.Flags().String("l3-network", "", "L3 network name or UUID")

	attachIsoCmd.Flags().String("iso", "", "ISO image name or UUID")
	detachIsoCmd.Flags().String("iso", "", "ISO image name or UUID")

	attachVolumeCmd.Flags().String("volume", "", "Data volume name or UUID")
	detachVolumeCmd.Flags().String("volume", "", "Data volume name or UUID")

	InstanceCmd.AddCommand(
		attachL3Cmd,
		detachL3Cmd,
		attachIsoCmd,
		detachIsoCmd,
		attachVolumeCmd,
		detachVolumeCmd,
	)
}

func isInstanceDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printInstanceDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printInstanceResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}

func stringPtrIfNotEmptyFlag(cmd *cobra.Command, flag string) *string {
	value, _ := cmd.Flags().GetString(flag)
	if value == "" {
		return nil
	}
	return &value
}
