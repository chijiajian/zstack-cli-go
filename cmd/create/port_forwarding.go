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

package create

import (
	"fmt"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

var portForwardingCmd = &cobra.Command{
	Use:     "port-forwarding <name>",
	Aliases: []string{"pf"},
	Short:   "Create a port forwarding rule",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pfParam, err := buildCreatePortForwardingParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(pfParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vip, err := zstackclient.ResolveOneResource(cli, "vip", pfParam.Params.VipUuid)
		if err != nil {
			return err
		}
		pfParam.Params.VipUuid = vip.UUID

		if pfParam.Params.VmNicUuid != nil {
			nic, err := zstackclient.ResolveOneResource(cli, "nic", *pfParam.Params.VmNicUuid)
			if err != nil {
				return err
			}
			pfParam.Params.VmNicUuid = &nic.UUID
		}

		result, err := cli.CreatePortForwardingRule(pfParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("port-forwarding", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(portForwardingCmd)
	portForwardingCmd.Flags().StringP("description", "d", "", "Detailed description of the rule")
	portForwardingCmd.Flags().String("vip", "", "VIP name or UUID")
	portForwardingCmd.Flags().String("nic", "", "VM NIC name or UUID")
	portForwardingCmd.Flags().String("protocol", "", "Protocol type, for example TCP or UDP")
	portForwardingCmd.Flags().Int("vip-port", 0, "VIP port start")
	portForwardingCmd.Flags().Int("vip-port-end", 0, "VIP port end")
	portForwardingCmd.Flags().Int("private-port", 0, "Private port start")
	portForwardingCmd.Flags().Int("private-port-end", 0, "Private port end")
	portForwardingCmd.Flags().String("allowed-cidr", "", "Allowed CIDR")
	portForwardingCmd.Flags().String("resource-uuid", "", "Resource UUID")
	portForwardingCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	portForwardingCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	portForwardingCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreatePortForwardingParam(cmd *cobra.Command, name string) (param.CreatePortForwardingRuleParam, error) {
	vip, _ := cmd.Flags().GetString("vip")
	if vip == "" {
		return param.CreatePortForwardingRuleParam{}, fmt.Errorf("required flag --vip not set")
	}
	protocol, _ := cmd.Flags().GetString("protocol")
	if protocol == "" {
		return param.CreatePortForwardingRuleParam{}, fmt.Errorf("required flag --protocol not set")
	}
	vipPort, _ := cmd.Flags().GetInt("vip-port")
	if vipPort == 0 {
		return param.CreatePortForwardingRuleParam{}, fmt.Errorf("required flag --vip-port not set")
	}

	description, _ := cmd.Flags().GetString("description")
	nic, _ := cmd.Flags().GetString("nic")
	vipPortEnd, _ := cmd.Flags().GetInt("vip-port-end")
	privatePort, _ := cmd.Flags().GetInt("private-port")
	privatePortEnd, _ := cmd.Flags().GetInt("private-port-end")
	allowedCIDR, _ := cmd.Flags().GetString("allowed-cidr")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreatePortForwardingRuleParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreatePortForwardingRuleParamDetail{
			Name:             name,
			Description:      stringPtrIfNotEmpty(description),
			VipUuid:          vip,
			VmNicUuid:        stringPtrIfNotEmpty(nic),
			ProtocolType:     protocol,
			VipPortStart:     vipPort,
			VipPortEnd:       intPtrIfNotZero(vipPortEnd),
			PrivatePortStart: intPtrIfNotZero(privatePort),
			PrivatePortEnd:   intPtrIfNotZero(privatePortEnd),
			AllowedCidr:      stringPtrIfNotEmpty(allowedCIDR),
			ResourceUuid:     stringPtrIfNotEmpty(resourceUUID),
			TagUuids:         tagUUIDs,
		},
	}, nil
}
