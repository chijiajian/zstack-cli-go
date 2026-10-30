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

var securityGroupCmd = &cobra.Command{
	Use:     "security-group <name>",
	Aliases: []string{"sg"},
	Short:   "Create a security group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		securityGroupParam := buildCreateSecurityGroupParam(cmd, args[0])

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(securityGroupParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		result, err := cli.CreateSecurityGroup(securityGroupParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("security-group", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(securityGroupCmd)
	securityGroupCmd.Flags().StringP("description", "d", "", "Detailed description of the security group")
	securityGroupCmd.Flags().Int("ip-version", 0, "IP version")
	securityGroupCmd.Flags().String("vswitch-type", "", "vSwitch type")
	securityGroupCmd.Flags().String("resource-uuid", "", "Resource UUID")
	securityGroupCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	securityGroupCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	securityGroupCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateSecurityGroupParam(cmd *cobra.Command, name string) param.CreateSecurityGroupParam {
	description, _ := cmd.Flags().GetString("description")
	ipVersion, _ := cmd.Flags().GetInt("ip-version")
	vSwitchType, _ := cmd.Flags().GetString("vswitch-type")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateSecurityGroupParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateSecurityGroupParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			IpVersion:    intPtrIfNotZero(ipVersion),
			VSwitchType:  stringPtrIfNotEmpty(vSwitchType),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}
}
