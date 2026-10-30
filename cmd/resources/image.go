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
	sdkClient "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type imageActionRequest struct {
	Image   string      `json:"image,omitempty" yaml:"image,omitempty"`
	Source  string      `json:"source,omitempty" yaml:"source,omitempty"`
	Storage string      `json:"storage,omitempty" yaml:"storage,omitempty"`
	Params  interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

var ImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage images and data volume templates",
	Long:  "Recover, sync, and create ZStack image templates.",
}

var ImageGroupCmd = &cobra.Command{
	Use:   "image-group",
	Short: "Manage image groups",
	Long:  "Create ZStack image groups from root volume templates.",
}

var imageRecoverCmd = &cobra.Command{
	Use:   "recover <name-or-uuid>",
	Short: "Recover a deleted image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		storages, _ := cmd.Flags().GetStringSlice("backup-storage")
		recoverParam := param.RecoverImageParam{
			Params: param.RecoverImageParamDetail{BackupStorageUuids: storages},
		}
		if isImageDryRun(cmd) {
			return printImageDryRun(cmd, imageActionRequest{Image: args[0], Params: recoverParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		image, err := zstackclient.ResolveOneResource(cli, "image", args[0])
		if err != nil {
			return err
		}
		recoverParam.Params.BackupStorageUuids, err = resolveBackupStorageList(cli, storages)
		if err != nil {
			return err
		}
		result, err := cli.RecoverImage(image.UUID, recoverParam)
		if err != nil {
			return err
		}
		return printImageResult(cmd, fmt.Sprintf("image/%s recovered", result.Name), result)
	},
}

var imageSyncCmd = &cobra.Command{
	Use:   "sync <name-or-uuid>",
	Short: "Sync image metadata on an image store backup storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageStorage, _ := cmd.Flags().GetString("image-storage")
		syncParam := param.SyncImageParam{}
		if isImageDryRun(cmd) {
			return printImageDryRun(cmd, imageActionRequest{Image: args[0], Storage: imageStorage, Params: syncParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		image, err := zstackclient.ResolveOneResource(cli, "image", args[0])
		if err != nil {
			return err
		}
		imageStoreUUID := ""
		if imageStorage != "" {
			storage, err := zstackclient.ResolveOneResource(cli, "backup-storage", imageStorage)
			if err != nil {
				return err
			}
			imageStoreUUID = storage.UUID
		} else {
			fullImage, err := cli.GetImage(image.UUID)
			if err != nil {
				return err
			}
			if len(fullImage.BackupStorageRefs) == 0 {
				return fmt.Errorf("image %q has no backup storage refs; pass --image-storage", args[0])
			}
			imageStoreUUID = fullImage.BackupStorageRefs[0].BackupStorageUuid
		}
		result, err := cli.SyncImage(imageStoreUUID, syncParam)
		if err != nil {
			return err
		}
		return printImageResult(cmd, fmt.Sprintf("image/%s synced", image.Name), result)
	},
}

var imageCreateFromVolumeCmd = &cobra.Command{
	Use:   "create-from-volume <volume>",
	Short: "Create a data volume template from a volume",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateParam, err := buildCreateDataVolumeTemplateFromVolumeParam(cmd)
		if err != nil {
			return err
		}
		if isImageDryRun(cmd) {
			return printImageDryRun(cmd, imageActionRequest{Source: args[0], Params: templateParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		volume, err := zstackclient.ResolveOneResource(cli, "volume", args[0])
		if err != nil {
			return err
		}
		templateParam.Params.BackupStorageUuids, err = resolveBackupStorageList(cli, templateParam.Params.BackupStorageUuids)
		if err != nil {
			return err
		}
		result, err := cli.CreateDataVolumeTemplateFromVolume(volume.UUID, templateParam)
		if err != nil {
			return err
		}
		return printImageResult(cmd, fmt.Sprintf("image/%s created from volume/%s", result.Name, volume.Name), result)
	},
}

var imageCreateFromSnapshotCmd = &cobra.Command{
	Use:   "create-from-snapshot <snapshot>",
	Short: "Create a data volume template from a volume snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateParam, err := buildCreateDataVolumeTemplateFromSnapshotParam(cmd)
		if err != nil {
			return err
		}
		if isImageDryRun(cmd) {
			return printImageDryRun(cmd, imageActionRequest{Source: args[0], Params: templateParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		snapshot, err := zstackclient.ResolveOneResource(cli, "volume-snapshot", args[0])
		if err != nil {
			return err
		}
		templateParam.Params.BackupStorageUuids, err = resolveBackupStorageList(cli, templateParam.Params.BackupStorageUuids)
		if err != nil {
			return err
		}
		result, err := cli.CreateDataVolumeTemplateFromVolumeSnapshot(snapshot.UUID, templateParam)
		if err != nil {
			return err
		}
		return printImageResult(cmd, fmt.Sprintf("image/%s created from snapshot/%s", result.Name, snapshot.Name), result)
	},
}

var imageGroupCreateFromImageCmd = &cobra.Command{
	Use:   "create-from-image <image>",
	Short: "Create an image group from a root volume template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupParam, err := buildCreateImageGroupFromImageParam(cmd)
		if err != nil {
			return err
		}
		if isImageDryRun(cmd) {
			return printImageDryRun(cmd, imageActionRequest{Image: args[0], Params: groupParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		image, err := zstackclient.ResolveOneResource(cli, "image", args[0])
		if err != nil {
			return err
		}
		groupParam.Params.DataVolumeTemplateUuids, err = resolveImageList(cli, groupParam.Params.DataVolumeTemplateUuids)
		if err != nil {
			return err
		}
		result, err := cli.CreateImageGroupFromImage(image.UUID, groupParam)
		if err != nil {
			return err
		}
		return printImageResult(cmd, fmt.Sprintf("image-group/%s created", result.Name), result)
	},
}

func init() {
	addImagePersistentFlags(ImageCmd)
	addImagePersistentFlags(ImageGroupCmd)

	imageRecoverCmd.Flags().StringSlice("backup-storage", []string{}, "Backup storage name or UUID list")
	imageSyncCmd.Flags().String("image-storage", "", "Image store backup storage name or UUID")

	addTemplateFlags(imageCreateFromVolumeCmd, false)
	addTemplateFlags(imageCreateFromSnapshotCmd, false)
	addImageGroupFlags(imageGroupCreateFromImageCmd)

	ImageCmd.AddCommand(imageRecoverCmd, imageSyncCmd, imageCreateFromVolumeCmd, imageCreateFromSnapshotCmd)
	ImageGroupCmd.AddCommand(imageGroupCreateFromImageCmd)
}

func addImagePersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	cmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	cmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")
}

func addTemplateFlags(cmd *cobra.Command, requireBackupStorage bool) {
	cmd.Flags().String("name", "", "Template name")
	cmd.Flags().StringP("description", "d", "", "Detailed description of the template")
	cmd.Flags().StringSlice("backup-storage", []string{}, "Backup storage name or UUID list")
	cmd.Flags().Bool("require-backup-storage", requireBackupStorage, "Require at least one backup storage")
	cmd.Flags().String("resource-uuid", "", "Resource UUID")
	cmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	cmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	cmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func addImageGroupFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Image group name")
	cmd.Flags().StringP("description", "d", "", "Detailed description of the image group")
	cmd.Flags().StringSlice("data-volume-template", []string{}, "Data volume template name or UUID list")
	cmd.Flags().String("resource-uuid", "", "Resource UUID")
	cmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	cmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	cmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateDataVolumeTemplateFromVolumeParam(cmd *cobra.Command) (param.CreateDataVolumeTemplateFromVolumeParam, error) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return param.CreateDataVolumeTemplateFromVolumeParam{}, fmt.Errorf("required flag --name not set")
	}
	description, _ := cmd.Flags().GetString("description")
	backupStorages, _ := cmd.Flags().GetStringSlice("backup-storage")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateDataVolumeTemplateFromVolumeParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateDataVolumeTemplateFromVolumeParamDetail{
			Name:               name,
			Description:        stringPtrIfNotEmpty(description),
			BackupStorageUuids: backupStorages,
			ResourceUuid:       stringPtrIfNotEmpty(resourceUUID),
			TagUuids:           tagUUIDs,
		},
	}, nil
}

func buildCreateDataVolumeTemplateFromSnapshotParam(cmd *cobra.Command) (param.CreateDataVolumeTemplateFromVolumeSnapshotParam, error) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return param.CreateDataVolumeTemplateFromVolumeSnapshotParam{}, fmt.Errorf("required flag --name not set")
	}
	backupStorages, _ := cmd.Flags().GetStringSlice("backup-storage")
	requireBackupStorage, _ := cmd.Flags().GetBool("require-backup-storage")
	if requireBackupStorage && len(backupStorages) == 0 {
		return param.CreateDataVolumeTemplateFromVolumeSnapshotParam{}, fmt.Errorf("required flag --backup-storage not set")
	}
	description, _ := cmd.Flags().GetString("description")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateDataVolumeTemplateFromVolumeSnapshotParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateDataVolumeTemplateFromVolumeSnapshotParamDetail{
			Name:               name,
			Description:        stringPtrIfNotEmpty(description),
			BackupStorageUuids: backupStorages,
			ResourceUuid:       stringPtrIfNotEmpty(resourceUUID),
			TagUuids:           tagUUIDs,
		},
	}, nil
}

func buildCreateImageGroupFromImageParam(cmd *cobra.Command) (param.CreateImageGroupFromImageParam, error) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return param.CreateImageGroupFromImageParam{}, fmt.Errorf("required flag --name not set")
	}
	description, _ := cmd.Flags().GetString("description")
	dataVolumeTemplates, _ := cmd.Flags().GetStringSlice("data-volume-template")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateImageGroupFromImageParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateImageGroupFromImageParamDetail{
			Name:                    name,
			Description:             stringPtrIfNotEmpty(description),
			DataVolumeTemplateUuids: dataVolumeTemplates,
			ResourceUuid:            stringPtrIfNotEmpty(resourceUUID),
			TagUuids:                tagUUIDs,
		},
	}, nil
}

func resolveBackupStorageList(cli *sdkClient.ZSClient, identifiers []string) ([]string, error) {
	return resolveResourceUUIDList(cli, "backup-storage", identifiers)
}

func resolveImageList(cli *sdkClient.ZSClient, identifiers []string) ([]string, error) {
	return resolveResourceUUIDList(cli, "image", identifiers)
}

func resolveResourceUUIDList(cli *sdkClient.ZSClient, resourceType string, identifiers []string) ([]string, error) {
	if len(identifiers) == 0 {
		return nil, nil
	}
	resolved := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		resource, err := zstackclient.ResolveOneResource(cli, resourceType, identifier)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, resource.UUID)
	}
	return resolved, nil
}

func isImageDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printImageDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printImageResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
