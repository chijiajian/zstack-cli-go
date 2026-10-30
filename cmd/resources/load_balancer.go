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

type loadBalancerNicRequest struct {
	Listener string `json:"listener" yaml:"listener"`
	NIC      string `json:"nic" yaml:"nic"`
}

var LoadBalancerCmd = &cobra.Command{
	Use:     "load-balancer",
	Aliases: []string{"lb"},
	Short:   "Manage load balancer listeners",
	Long:    "Attach and remove VM NICs on ZStack load balancer listeners.",
}

var loadBalancerAddNicCmd = &cobra.Command{
	Use:   "add-nic <listener>",
	Short: "Add a VM NIC to a load balancer listener",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nicName, _ := cmd.Flags().GetString("nic")
		if nicName == "" {
			return fmt.Errorf("required flag --nic not set")
		}
		request := loadBalancerNicRequest{Listener: args[0], NIC: nicName}
		if isLoadBalancerDryRun(cmd) {
			return printLoadBalancerDryRun(cmd, request)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		listener, err := zstackclient.ResolveOneResource(cli, "load-balancer-listener", args[0])
		if err != nil {
			return err
		}
		nic, err := zstackclient.ResolveOneResource(cli, "nic", nicName)
		if err != nil {
			return err
		}

		params := param.AddVmNicToLoadBalancerParam{
			Params: param.AddVmNicToLoadBalancerParamDetail{VmNicUuids: []string{nic.UUID}},
		}
		result, err := cli.AddVmNicToLoadBalancer(listener.UUID, params)
		if err != nil {
			return err
		}
		return printLoadBalancerResult(cmd, fmt.Sprintf("load-balancer-listener/%s added nic/%s", listener.Name, nic.UUID), result)
	},
}

var loadBalancerRemoveNicCmd = &cobra.Command{
	Use:   "remove-nic <listener>",
	Short: "Remove a VM NIC from a load balancer listener",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nicName, _ := cmd.Flags().GetString("nic")
		if nicName == "" {
			return fmt.Errorf("required flag --nic not set")
		}
		request := loadBalancerNicRequest{Listener: args[0], NIC: nicName}
		if isLoadBalancerDryRun(cmd) {
			return printLoadBalancerDryRun(cmd, request)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		listener, err := zstackclient.ResolveOneResource(cli, "load-balancer-listener", args[0])
		if err != nil {
			return err
		}
		nic, err := zstackclient.ResolveOneResource(cli, "nic", nicName)
		if err != nil {
			return err
		}

		params := param.RemoveVmNicFromLoadBalancerParam{
			Params: param.RemoveVmNicFromLoadBalancerParamDetail{VmNicUuids: []string{nic.UUID}},
		}
		if err := cli.DeleteWithSpec("v1/load-balancers/listeners", listener.UUID, "vm-instances/nics", fmt.Sprintf("deleteMode=%s", param.DeleteModePermissive), params); err != nil {
			return err
		}
		fmt.Printf("load-balancer-listener/%s removed nic/%s\n", listener.Name, nic.UUID)
		return nil
	},
}

func init() {
	LoadBalancerCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	LoadBalancerCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	LoadBalancerCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	loadBalancerAddNicCmd.Flags().String("nic", "", "VM NIC name or UUID")
	loadBalancerRemoveNicCmd.Flags().String("nic", "", "VM NIC name or UUID")

	LoadBalancerCmd.AddCommand(loadBalancerAddNicCmd, loadBalancerRemoveNicCmd)
}

func isLoadBalancerDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printLoadBalancerDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printLoadBalancerResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
