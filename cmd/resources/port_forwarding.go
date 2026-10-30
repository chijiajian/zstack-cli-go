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

type portForwardingAttachRequest struct {
	Rule string `json:"rule" yaml:"rule"`
	NIC  string `json:"nic,omitempty" yaml:"nic,omitempty"`
}

var PortForwardingCmd = &cobra.Command{
	Use:     "port-forwarding",
	Aliases: []string{"pf"},
	Short:   "Manage port forwarding rules",
	Long:    "Attach and detach ZStack port forwarding rules.",
}

var portForwardingAttachCmd = &cobra.Command{
	Use:   "attach <name-or-uuid>",
	Short: "Attach a port forwarding rule to a VM NIC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nicName, _ := cmd.Flags().GetString("nic")
		if nicName == "" {
			return fmt.Errorf("required flag --nic not set")
		}
		if isPortForwardingDryRun(cmd) {
			return printPortForwardingDryRun(cmd, portForwardingAttachRequest{Rule: args[0], NIC: nicName})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		rule, err := zstackclient.ResolveOneResource(cli, "port-forwarding", args[0])
		if err != nil {
			return err
		}
		nic, err := zstackclient.ResolveOneResource(cli, "nic", nicName)
		if err != nil {
			return err
		}

		result, err := cli.AttachPortForwardingRule(rule.UUID, nic.UUID, param.AttachPortForwardingRuleParam{})
		if err != nil {
			return err
		}
		return printPortForwardingResult(cmd, fmt.Sprintf("port-forwarding/%s attached nic/%s", result.Name, nic.UUID), result)
	},
}

var portForwardingDetachCmd = &cobra.Command{
	Use:   "detach <name-or-uuid>",
	Short: "Detach a port forwarding rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isPortForwardingDryRun(cmd) {
			return printPortForwardingDryRun(cmd, portForwardingAttachRequest{Rule: args[0]})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		rule, err := zstackclient.ResolveOneResource(cli, "port-forwarding", args[0])
		if err != nil {
			return err
		}

		if err := cli.DetachPortForwardingRule(rule.UUID, param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("port-forwarding/%s detached\n", rule.Name)
		return nil
	},
}

func init() {
	PortForwardingCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	PortForwardingCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	PortForwardingCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	portForwardingAttachCmd.Flags().String("nic", "", "VM NIC name or UUID")

	PortForwardingCmd.AddCommand(portForwardingAttachCmd, portForwardingDetachCmd)
}

func isPortForwardingDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printPortForwardingDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printPortForwardingResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
