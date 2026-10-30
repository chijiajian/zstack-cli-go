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

type eipAttachRequest struct {
	EIP    string `json:"eip" yaml:"eip"`
	NIC    string `json:"nic" yaml:"nic"`
	UsedIP string `json:"usedIp,omitempty" yaml:"usedIp,omitempty"`
}

type eipDetachRequest struct {
	EIP string `json:"eip" yaml:"eip"`
}

var EipCmd = &cobra.Command{
	Use:     "eip",
	Aliases: []string{"eips"},
	Short:   "Manage EIPs",
	Long:    "Attach and detach ZStack elastic IP resources.",
}

var eipAttachCmd = &cobra.Command{
	Use:   "attach <name-or-uuid>",
	Short: "Attach an EIP to a VM NIC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nicName, _ := cmd.Flags().GetString("nic")
		if nicName == "" {
			return fmt.Errorf("required flag --nic not set")
		}
		usedIP, _ := cmd.Flags().GetString("used-ip")

		if isEipDryRun(cmd) {
			return printEipDryRun(cmd, eipAttachRequest{EIP: args[0], NIC: nicName, UsedIP: usedIP})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		eip, err := zstackclient.ResolveOneResource(cli, "eip", args[0])
		if err != nil {
			return err
		}
		nic, err := zstackclient.ResolveOneResource(cli, "nic", nicName)
		if err != nil {
			return err
		}

		result, err := cli.AttachEip(eip.UUID, nic.UUID, param.AttachEipParam{
			Params: param.AttachEipParamDetail{
				UsedIpUuid: stringPtrIfNotEmpty(usedIP),
			},
		})
		if err != nil {
			return err
		}

		return printEipResult(cmd, fmt.Sprintf("eip/%s attached to nic/%s", result.Name, nic.UUID), result)
	},
}

var eipDetachCmd = &cobra.Command{
	Use:   "detach <name-or-uuid>",
	Short: "Detach an EIP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isEipDryRun(cmd) {
			return printEipDryRun(cmd, eipDetachRequest{EIP: args[0]})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		eip, err := zstackclient.ResolveOneResource(cli, "eip", args[0])
		if err != nil {
			return err
		}

		if err := cli.DetachEip(eip.UUID, param.DeleteModePermissive); err != nil {
			return err
		}

		fmt.Printf("eip/%s detached\n", eip.Name)
		return nil
	},
}

func init() {
	EipCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	EipCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	EipCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	eipAttachCmd.Flags().String("nic", "", "VM NIC name or UUID")
	eipAttachCmd.Flags().String("used-ip", "", "Used IP UUID")

	EipCmd.AddCommand(eipAttachCmd, eipDetachCmd)
}

func isEipDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printEipDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printEipResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
