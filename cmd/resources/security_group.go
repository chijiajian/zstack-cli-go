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

type securityGroupTargetRequest struct {
	SecurityGroup string      `json:"securityGroup" yaml:"securityGroup"`
	Target        string      `json:"target,omitempty" yaml:"target,omitempty"`
	Params        interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

var SecurityGroupCmd = &cobra.Command{
	Use:     "security-group",
	Aliases: []string{"sg"},
	Short:   "Manage security groups",
	Long:    "Attach networks, manage rules, and manage NIC membership for ZStack security groups.",
}

var securityGroupAttachL3Cmd = &cobra.Command{
	Use:   "attach-l3 <name-or-uuid>",
	Short: "Attach a security group to an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityGroupL3(cmd, args[0], true)
	},
}

var securityGroupDetachL3Cmd = &cobra.Command{
	Use:   "detach-l3 <name-or-uuid>",
	Short: "Detach a security group from an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityGroupL3(cmd, args[0], false)
	},
}

var securityGroupAddRuleCmd = &cobra.Command{
	Use:   "add-rule <name-or-uuid>",
	Short: "Add a security group rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ruleType, _ := cmd.Flags().GetString("type")
		protocol, _ := cmd.Flags().GetString("protocol")
		if ruleType == "" {
			return fmt.Errorf("required flag --type not set")
		}
		if protocol == "" {
			return fmt.Errorf("required flag --protocol not set")
		}

		ruleParam, err := buildSecurityGroupRuleParam(cmd, ruleType, protocol)
		if err != nil {
			return err
		}
		if isSecurityGroupDryRun(cmd) {
			return printSecurityGroupDryRun(cmd, securityGroupTargetRequest{SecurityGroup: args[0], Params: ruleParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		sg, err := zstackclient.ResolveOneResource(cli, "security-group", args[0])
		if err != nil {
			return err
		}
		remoteSG, _ := cmd.Flags().GetString("remote-security-group")
		if remoteSG != "" {
			remote, err := zstackclient.ResolveOneResource(cli, "security-group", remoteSG)
			if err != nil {
				return err
			}
			ruleParam.Params.Rules[0].RemoteSecurityGroupUuid = &remote.UUID
		}

		result, err := cli.AddSecurityGroupRule(sg.UUID, ruleParam)
		if err != nil {
			return err
		}
		return printSecurityGroupResult(cmd, fmt.Sprintf("security-group/%s rule added", sg.Name), result)
	},
}

var securityGroupDeleteRuleCmd = &cobra.Command{
	Use:   "delete-rule <rule-uuid>",
	Short: "Delete a security group rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isSecurityGroupDryRun(cmd) {
			return printSecurityGroupDryRun(cmd, map[string]string{"rule": args[0]})
		}
		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		if err := cli.DeleteSecurityGroupRule(args[0], param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("security-group-rule/%s deleted\n", args[0])
		return nil
	},
}

var securityGroupAddNicCmd = &cobra.Command{
	Use:   "add-nic <name-or-uuid>",
	Short: "Add a VM NIC to a security group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityGroupNic(cmd, args[0], true)
	},
}

var securityGroupRemoveNicCmd = &cobra.Command{
	Use:   "remove-nic <name-or-uuid>",
	Short: "Remove a VM NIC from a security group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityGroupNic(cmd, args[0], false)
	},
}

func init() {
	SecurityGroupCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	SecurityGroupCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	SecurityGroupCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	securityGroupAttachL3Cmd.Flags().String("l3-network", "", "L3 network name or UUID")
	securityGroupDetachL3Cmd.Flags().String("l3-network", "", "L3 network name or UUID")

	securityGroupAddRuleCmd.Flags().String("type", "", "Rule type, for example Ingress or Egress")
	securityGroupAddRuleCmd.Flags().String("protocol", "", "Protocol, for example TCP, UDP, ICMP, or ALL")
	securityGroupAddRuleCmd.Flags().Int("start-port", 0, "Start port")
	securityGroupAddRuleCmd.Flags().Int("end-port", 0, "End port")
	securityGroupAddRuleCmd.Flags().String("cidr", "", "Allowed CIDR")
	securityGroupAddRuleCmd.Flags().String("action", "", "Rule action")
	securityGroupAddRuleCmd.Flags().String("state", "", "Rule state")
	securityGroupAddRuleCmd.Flags().StringP("description", "d", "", "Rule description")
	securityGroupAddRuleCmd.Flags().Int("ip-version", 0, "IP version")
	securityGroupAddRuleCmd.Flags().String("remote-security-group", "", "Remote security group name or UUID")
	securityGroupAddRuleCmd.Flags().Int("priority", 0, "Rule priority")

	securityGroupAddNicCmd.Flags().String("nic", "", "VM NIC name or UUID")
	securityGroupRemoveNicCmd.Flags().String("nic", "", "VM NIC name or UUID")

	SecurityGroupCmd.AddCommand(
		securityGroupAttachL3Cmd,
		securityGroupDetachL3Cmd,
		securityGroupAddRuleCmd,
		securityGroupDeleteRuleCmd,
		securityGroupAddNicCmd,
		securityGroupRemoveNicCmd,
	)
}

func runSecurityGroupL3(cmd *cobra.Command, securityGroup string, attach bool) error {
	l3Network, _ := cmd.Flags().GetString("l3-network")
	if l3Network == "" {
		return fmt.Errorf("required flag --l3-network not set")
	}
	if isSecurityGroupDryRun(cmd) {
		return printSecurityGroupDryRun(cmd, securityGroupTargetRequest{SecurityGroup: securityGroup, Target: l3Network})
	}

	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	sg, err := zstackclient.ResolveOneResource(cli, "security-group", securityGroup)
	if err != nil {
		return err
	}
	l3, err := zstackclient.ResolveOneResource(cli, "l3-network", l3Network)
	if err != nil {
		return err
	}

	if attach {
		result, err := cli.AttachSecurityGroupToL3Network(sg.UUID, l3.UUID, param.AttachSecurityGroupToL3NetworkParam{})
		if err != nil {
			return err
		}
		return printSecurityGroupResult(cmd, fmt.Sprintf("security-group/%s attached l3-network/%s", sg.Name, l3.Name), result)
	}

	if err := cli.DetachSecurityGroupFromL3Network(sg.UUID, l3.UUID, param.DeleteModePermissive); err != nil {
		return err
	}
	fmt.Printf("security-group/%s detached l3-network/%s\n", sg.Name, l3.Name)
	return nil
}

func buildSecurityGroupRuleParam(cmd *cobra.Command, ruleType string, protocol string) (param.AddSecurityGroupRuleParam, error) {
	startPort, _ := cmd.Flags().GetInt("start-port")
	endPort, _ := cmd.Flags().GetInt("end-port")
	cidr, _ := cmd.Flags().GetString("cidr")
	action, _ := cmd.Flags().GetString("action")
	state, _ := cmd.Flags().GetString("state")
	description, _ := cmd.Flags().GetString("description")
	ipVersion, _ := cmd.Flags().GetInt("ip-version")
	remoteSG, _ := cmd.Flags().GetString("remote-security-group")
	priority, _ := cmd.Flags().GetInt("priority")

	rule := param.AddSecurityGroupRule_SecurityGroupRuleAOParam{
		Type:        ruleType,
		Protocol:    stringPtrIfNotEmpty(protocol),
		StartPort:   intPtrIfNotZero(startPort),
		EndPort:     intPtrIfNotZero(endPort),
		AllowedCidr: stringPtrIfNotEmpty(cidr),
		Action:      stringPtrIfNotEmpty(action),
		State:       stringPtrIfNotEmpty(state),
		Description: stringPtrIfNotEmpty(description),
		IpVersion:   intPtrIfNotZero(ipVersion),
	}
	if remoteSG != "" {
		rule.RemoteSecurityGroupUuid = &remoteSG
	}

	return param.AddSecurityGroupRuleParam{
		Params: param.AddSecurityGroupRuleParamDetail{
			Rules:    []param.AddSecurityGroupRule_SecurityGroupRuleAOParam{rule},
			Priority: intPtrIfNotZero(priority),
		},
	}, nil
}

func runSecurityGroupNic(cmd *cobra.Command, securityGroup string, add bool) error {
	nicName, _ := cmd.Flags().GetString("nic")
	if nicName == "" {
		return fmt.Errorf("required flag --nic not set")
	}
	if isSecurityGroupDryRun(cmd) {
		return printSecurityGroupDryRun(cmd, securityGroupTargetRequest{SecurityGroup: securityGroup, Target: nicName})
	}

	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	sg, err := zstackclient.ResolveOneResource(cli, "security-group", securityGroup)
	if err != nil {
		return err
	}
	nic, err := zstackclient.ResolveOneResource(cli, "nic", nicName)
	if err != nil {
		return err
	}

	if add {
		result, err := cli.AddVmNicToSecurityGroup(sg.UUID, param.AddVmNicToSecurityGroupParam{
			Params: param.AddVmNicToSecurityGroupParamDetail{VmNicUuids: []string{nic.UUID}},
		})
		if err != nil {
			return err
		}
		return printSecurityGroupResult(cmd, fmt.Sprintf("security-group/%s added nic/%s", sg.Name, nic.UUID), result)
	}

	params := param.DeleteVmNicFromSecurityGroupParam{
		Params: param.DeleteVmNicFromSecurityGroupParamDetail{VmNicUuids: []string{nic.UUID}},
	}
	if err := cli.DeleteWithSpec("v1/security-groups", sg.UUID, "vm-instances/nics", fmt.Sprintf("deleteMode=%s", param.DeleteModePermissive), params); err != nil {
		return err
	}
	fmt.Printf("security-group/%s removed nic/%s\n", sg.Name, nic.UUID)
	return nil
}

func isSecurityGroupDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printSecurityGroupDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printSecurityGroupResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
