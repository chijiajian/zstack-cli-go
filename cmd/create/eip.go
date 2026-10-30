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

var eipCmd = &cobra.Command{
	Use:   "eip <name>",
	Short: "Create an EIP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eipParam, err := buildCreateEipParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(eipParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vip, err := zstackclient.ResolveOneResource(cli, "vip", eipParam.Params.VipUuid)
		if err != nil {
			return err
		}
		eipParam.Params.VipUuid = vip.UUID

		if eipParam.Params.VmNicUuid != nil {
			nic, err := zstackclient.ResolveOneResource(cli, "nic", *eipParam.Params.VmNicUuid)
			if err != nil {
				return err
			}
			eipParam.Params.VmNicUuid = &nic.UUID
		}

		result, err := cli.CreateEip(eipParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("eip", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(eipCmd)
	eipCmd.Flags().StringP("description", "d", "", "Detailed description of the EIP")
	eipCmd.Flags().String("vip", "", "VIP name or UUID")
	eipCmd.Flags().String("nic", "", "VM NIC name or UUID")
	eipCmd.Flags().String("used-ip", "", "Used IP UUID")
	eipCmd.Flags().String("resource-uuid", "", "Resource UUID")
	eipCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	eipCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	eipCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateEipParam(cmd *cobra.Command, name string) (param.CreateEipParam, error) {
	vip, _ := cmd.Flags().GetString("vip")
	if vip == "" {
		return param.CreateEipParam{}, fmt.Errorf("required flag --vip not set")
	}
	description, _ := cmd.Flags().GetString("description")
	nic, _ := cmd.Flags().GetString("nic")
	usedIP, _ := cmd.Flags().GetString("used-ip")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateEipParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateEipParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			VipUuid:      vip,
			VmNicUuid:    stringPtrIfNotEmpty(nic),
			UsedIpUuid:   stringPtrIfNotEmpty(usedIP),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}
