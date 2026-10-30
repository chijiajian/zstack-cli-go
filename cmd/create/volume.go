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

var volumeCmd = &cobra.Command{
	Use:   "volume <name>",
	Short: "Create a data volume",
	Long: `Create a ZStack data volume.

Examples:
  zstack-cli create volume data-01 --size 10G
  zstack-cli create volume data-01 --disk-offering small-disk`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeParam, err := buildCreateVolumeParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(volumeParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		diskOffering, _ := cmd.Flags().GetString("disk-offering")
		if diskOffering != "" {
			resolved, err := zstackclient.ResolveOneResource(cli, "disk-offering", diskOffering)
			if err != nil {
				return err
			}
			volumeParam.Params.DiskOfferingUuid = &resolved.UUID
		}

		primaryStorage, _ := cmd.Flags().GetString("primary-storage")
		if primaryStorage != "" {
			primaryStorageUUID, err := zstackclient.GetPrimaryStorageUUIDByName(cli, primaryStorage)
			if err != nil {
				return err
			}
			volumeParam.Params.PrimaryStorageUuid = &primaryStorageUUID
		}

		result, err := cli.CreateDataVolume(volumeParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("volume", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(volumeCmd)
	volumeCmd.Flags().StringP("description", "d", "", "Detailed description of the volume")
	volumeCmd.Flags().String("size", "", "Volume size, for example 10G")
	volumeCmd.Flags().String("disk-offering", "", "Disk offering name or UUID")
	volumeCmd.Flags().String("primary-storage", "", "Primary storage name or UUID")
	volumeCmd.Flags().String("resource-uuid", "", "Resource UUID")
	volumeCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	volumeCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	volumeCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateVolumeParam(cmd *cobra.Command, name string) (param.CreateDataVolumeParam, error) {
	size, _ := cmd.Flags().GetString("size")
	diskOffering, _ := cmd.Flags().GetString("disk-offering")
	if size == "" && diskOffering == "" {
		return param.CreateDataVolumeParam{}, fmt.Errorf("either --size or --disk-offering must be specified")
	}

	var diskSize *int64
	if size != "" {
		parsed, err := utils.ParseMemorySize(size)
		if err != nil {
			return param.CreateDataVolumeParam{}, fmt.Errorf("invalid --size: %w", err)
		}
		diskSize = &parsed
	}

	description, _ := cmd.Flags().GetString("description")
	primaryStorage, _ := cmd.Flags().GetString("primary-storage")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateDataVolumeParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateDataVolumeParamDetail{
			Name:               name,
			Description:        stringPtrIfNotEmpty(description),
			DiskOfferingUuid:   stringPtrIfNotEmpty(diskOffering),
			DiskSize:           diskSize,
			PrimaryStorageUuid: stringPtrIfNotEmpty(primaryStorage),
			ResourceUuid:       stringPtrIfNotEmpty(resourceUUID),
			TagUuids:           tagUUIDs,
		},
	}, nil
}
