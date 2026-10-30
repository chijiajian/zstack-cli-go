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

type volumeAttachRequest struct {
	Volume   string `json:"volume" yaml:"volume"`
	Instance string `json:"instance" yaml:"instance"`
}

type volumeDetachRequest struct {
	Volume string `json:"volume" yaml:"volume"`
}

var VolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage volumes",
	Long:  "Attach, detach, recover, and resize ZStack volumes.",
}

var volumeAttachCmd = &cobra.Command{
	Use:   "attach <name-or-uuid>",
	Short: "Attach a data volume to a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instance, _ := cmd.Flags().GetString("instance")
		if instance == "" {
			return fmt.Errorf("required flag --instance not set")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return printVolumeDryRun(cmd, volumeAttachRequest{Volume: args[0], Instance: instance})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		volume, err := zstackclient.ResolveOneResource(cli, "volume", args[0])
		if err != nil {
			return err
		}
		vm, err := zstackclient.ResolveOneResource(cli, "instance", instance)
		if err != nil {
			return err
		}

		result, err := cli.AttachDataVolumeToVm(volume.UUID, vm.UUID, param.AttachDataVolumeToVmParam{})
		if err != nil {
			return err
		}

		return printVolumeResult(cmd, fmt.Sprintf("volume/%s attached to instance/%s", result.Name, vm.Name), result)
	},
}

var volumeDetachCmd = &cobra.Command{
	Use:   "detach <name-or-uuid>",
	Short: "Detach a data volume from its VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return printVolumeDryRun(cmd, volumeDetachRequest{Volume: args[0]})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		volume, err := zstackclient.ResolveOneResource(cli, "volume", args[0])
		if err != nil {
			return err
		}

		if _, err := cli.DetachDataVolumeFromVm(volume.UUID, param.DetachDataVolumeFromVmParam{}); err != nil {
			return err
		}

		fmt.Printf("volume/%s detached\n", volume.Name)
		return nil
	},
}

func init() {
	VolumeCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	VolumeCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	VolumeCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	volumeAttachCmd.Flags().String("instance", "", "VM instance name or UUID")

	VolumeCmd.AddCommand(volumeAttachCmd, volumeDetachCmd)
}

func printVolumeDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printVolumeResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
