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

var vipCmd = &cobra.Command{
	Use:   "vip <name>",
	Short: "Create a VIP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vipParam, err := buildCreateVipParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(vipParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", vipParam.Params.L3NetworkUuid)
		if err != nil {
			return err
		}
		vipParam.Params.L3NetworkUuid = l3.UUID

		result, err := cli.CreateVip(vipParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("vip", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(vipCmd)
	vipCmd.Flags().StringP("description", "d", "", "Detailed description of the VIP")
	vipCmd.Flags().String("l3-network", "", "L3 network name or UUID")
	vipCmd.Flags().String("allocator-strategy", "", "VIP allocator strategy")
	vipCmd.Flags().String("ip-range", "", "IP range UUID")
	vipCmd.Flags().String("required-ip", "", "Required IP address")
	vipCmd.Flags().Int("ip-version", 0, "IP version")
	vipCmd.Flags().String("resource-uuid", "", "Resource UUID")
	vipCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	vipCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	vipCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateVipParam(cmd *cobra.Command, name string) (param.CreateVipParam, error) {
	l3Network, _ := cmd.Flags().GetString("l3-network")
	if l3Network == "" {
		return param.CreateVipParam{}, fmt.Errorf("required flag --l3-network not set")
	}
	description, _ := cmd.Flags().GetString("description")
	allocatorStrategy, _ := cmd.Flags().GetString("allocator-strategy")
	ipRange, _ := cmd.Flags().GetString("ip-range")
	requiredIP, _ := cmd.Flags().GetString("required-ip")
	ipVersion, _ := cmd.Flags().GetInt("ip-version")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateVipParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateVipParamDetail{
			Name:              name,
			Description:       stringPtrIfNotEmpty(description),
			L3NetworkUuid:     l3Network,
			AllocatorStrategy: stringPtrIfNotEmpty(allocatorStrategy),
			IpRangeUuid:       stringPtrIfNotEmpty(ipRange),
			RequiredIp:        stringPtrIfNotEmpty(requiredIP),
			IpVersion:         intPtrIfNotZero(ipVersion),
			ResourceUuid:      stringPtrIfNotEmpty(resourceUUID),
			TagUuids:          tagUUIDs,
		},
	}, nil
}

func intPtrIfNotZero(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}
