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

// cmd/create/image.go
package create

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"gopkg.in/yaml.v2"
)

type ImageSpec struct {
	URL                string   `json:"url" yaml:"url"`
	BackupStorageNames []string `json:"imageStorageName" yaml:"imageStorageName"`
	Description        string   `json:"description" yaml:"description"`
	MediaType          string   `json:"mediaType" yaml:"mediaType"`
	GuestOsType        string   `json:"guestOsType" yaml:"guestOsType"`
	System             bool     `json:"system" yaml:"system"`
	Format             string   `json:"format" yaml:"format"`
	Platform           string   `json:"platform" yaml:"platform"`
	Architecture       string   `json:"architecture" yaml:"architecture"`
	Virtio             bool     `json:"virtio" yaml:"virtio"`
	ResourceUUID       string   `json:"resourceUuid" yaml:"resourceUuid"`
	TagUUIDs           []string `json:"tagUuids" yaml:"tagUuids"`
	SystemTags         []string `json:"systemTags" yaml:"systemTags"`
	UserTags           []string `json:"userTags" yaml:"userTags"`
}

var imageCmd = &cobra.Command{
	Use:   "image [NAME]",
	Short: "Create a new image",
	Long: `Create a new image in ZStack cloud platform.

Examples:
  # Create a basic image from URL
  zstack-cli create image my-image --url http://example.com/image.qcow2 --image-storage bs-uuid1

  # Create a Windows image with more options
  zstack-cli create image windows-image --url http://example.com/windows.qcow2 --image-storage bs-uuid1 \
    --media-type RootVolumeTemplate --format qcow2 --platform Windows --guest-os-type "Windows 10" \
    --architecture x86_64

  # Create an image from a YAML or JSON file
  zstack-cli create image -f image-spec.yaml

  # Create an image from a YAML or JSON file with a different name
  zstack-cli create image override-name -f image-spec.yaml`,
	Args: exactOneArgUnlessFile,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			return createImageFromFile(cmd, name, filePath)
		}

		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			cmd.Help()
			return fmt.Errorf("required flag --url not set")
		}

		imageStorage, _ := cmd.Flags().GetString("image-storage")
		if imageStorage == "" {
			cmd.Help()
			return fmt.Errorf("required flag --image-storage not set")
		}

		return createImageFromFlags(cmd, name)
	},
}

func createImageFromFlags(cmd *cobra.Command, name string) error {

	url, _ := cmd.Flags().GetString("url")
	backupStorageStr, _ := cmd.Flags().GetString("image-storage")

	if url == "" {
		return fmt.Errorf("Error: required flag --url not set")
	}

	if backupStorageStr == "" {
		return fmt.Errorf("Error: required flag --image-storage not set")
	}

	backupStorageNames := strings.Split(backupStorageStr, ",")
	backupStorageUuids := make([]string, 0, len(backupStorageNames))

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Not logged in. Please run 'zstack-cli login' first.")
	}

	for _, nameOrUUID := range backupStorageNames {
		uuid, err := client.GetBackupStorageUUIDByName(cli, nameOrUUID)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}
		backupStorageUuids = append(backupStorageUuids, uuid)
	}

	description, _ := cmd.Flags().GetString("description")
	mediaType, _ := cmd.Flags().GetString("media-type")
	guestOsType, _ := cmd.Flags().GetString("guest-os-type")
	system, _ := cmd.Flags().GetBool("system")
	format, _ := cmd.Flags().GetString("format")
	platform, _ := cmd.Flags().GetString("platform")
	architecture, _ := cmd.Flags().GetString("architecture")
	virtio, _ := cmd.Flags().GetBool("virtio")
	resourceUuid, _ := cmd.Flags().GetString("resource-uuid")
	tagUuids, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	imageParam := param.AddImageParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.AddImageParamDetail{
			Name:               name,
			Description:        stringPtrIfNotEmpty(description),
			Url:                url,
			MediaType:          stringPtrIfNotEmpty(mediaType),
			GuestOsType:        stringPtrIfNotEmpty(guestOsType),
			System:             system,
			Format:             stringPtrIfNotEmpty(format),
			Platform:           stringPtrIfNotEmpty(platform),
			BackupStorageUuids: backupStorageUuids,
			ResourceUuid:       stringPtrIfNotEmpty(resourceUuid),
			Architecture:       stringPtrIfNotEmpty(architecture),
			Virtio:             boolPtr(virtio),
			TagUuids:           tagUuids,
		},
	}

	if dryRunFlag {
		return utils.PrintDryRun(imageParam, outputFlag)
	}

	if err := printCreateProgress(cmd, "Creating image '%s'...\n", name); err != nil {
		return err
	}
	result, err := cli.AddImage(imageParam)
	if err != nil {
		return fmt.Errorf("Error creating image: %w", err)
	}

	return utils.PrintOperationResult("Image", result, outputFlag)
}

func createImageFromFile(cmd *cobra.Command, name string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %w", filePath, err)
	}

	var imageSpec ImageSpec
	var resourceSpec utils.ResourceSpec

	isGenericFormat := false
	if err := yaml.Unmarshal(data, &resourceSpec); err == nil {
		if resourceSpec.Kind == utils.KindImage && resourceSpec.Spec != nil {
			isGenericFormat = true

			specData, err := json.Marshal(resourceSpec.Spec)
			if err != nil {
				return fmt.Errorf("Error converting resource spec: %w", err)
			}
			if err := json.Unmarshal(specData, &imageSpec); err != nil {
				return fmt.Errorf("Error parsing image spec from generic format: %w", err)
			}
			if name == "" {
				name = resourceSpec.Metadata.Name
			}
		}
	}

	if !isGenericFormat {
		if strings.HasSuffix(filePath, ".json") {
			if err := json.Unmarshal(data, &imageSpec); err != nil {
				return fmt.Errorf("Error parsing JSON file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &imageSpec); err != nil {
				return fmt.Errorf("Error parsing YAML file: %w", err)
			}
		}

		if name == "" {
			return fmt.Errorf("Error: image name is required")
		}
	}

	if imageSpec.URL == "" {
		return fmt.Errorf("Error: URL is required in image specification")
	}

	if len(imageSpec.BackupStorageNames) == 0 {
		return fmt.Errorf("Error: backupStorageUuids is required in image specification")
	}

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Not logged in. Please run 'zstack-cli login' first.")
	}

	backupStorageUuids := make([]string, 0, len(imageSpec.BackupStorageNames))
	for _, nameOrUUID := range imageSpec.BackupStorageNames {
		uuid, err := client.GetBackupStorageUUIDByName(cli, nameOrUUID)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}
		backupStorageUuids = append(backupStorageUuids, uuid)
	}

	imageParam := param.AddImageParam{
		BaseParam: param.BaseParam{
			SystemTags: imageSpec.SystemTags,
			UserTags:   imageSpec.UserTags,
		},
		Params: param.AddImageParamDetail{
			Name:               name,
			Description:        stringPtrIfNotEmpty(imageSpec.Description),
			Url:                imageSpec.URL,
			MediaType:          stringPtrIfNotEmpty(imageSpec.MediaType),
			GuestOsType:        stringPtrIfNotEmpty(imageSpec.GuestOsType),
			System:             imageSpec.System,
			Format:             stringPtrIfNotEmpty(imageSpec.Format),
			Platform:           stringPtrIfNotEmpty(imageSpec.Platform),
			BackupStorageUuids: backupStorageUuids,
			ResourceUuid:       stringPtrIfNotEmpty(imageSpec.ResourceUUID),
			Architecture:       stringPtrIfNotEmpty(imageSpec.Architecture),
			Virtio:             boolPtr(imageSpec.Virtio),
			TagUuids:           imageSpec.TagUUIDs,
		},
	}

	if dryRunFlag {
		return utils.PrintDryRun(imageParam, outputFlag)
	}

	if err := printCreateProgress(cmd, "Creating image '%s' from file...\n", name); err != nil {
		return err
	}
	result, err := cli.AddImage(imageParam)
	if err != nil {
		return fmt.Errorf("Error creating image: %w", err)
	}

	return utils.PrintOperationResult("Image", result, outputFlag)
}

func init() {
	CreateCmd.AddCommand(imageCmd)

	imageCmd.Flags().StringP("file", "f", "", "Path to YAML or JSON file containing image specification")

	imageCmd.Flags().String("url", "", "URL of the image to be added (required)")
	imageCmd.Flags().String("image-storage", "", "Comma-separated list of backup storage UUIDs (required)")

	imageCmd.Flags().String("description", "", "Detailed description of the image")
	imageCmd.Flags().String("media-type", "RootVolumeTemplate", "Image type (RootVolumeTemplate, ISO, DataVolumeTemplate)")
	imageCmd.Flags().String("guest-os-type", "", "Guest OS type corresponding to the image")
	imageCmd.Flags().Bool("system", false, "Whether it is a system image (e.g., cloud router image)")
	imageCmd.Flags().String("format", "qcow2", "Image format, e.g., raw, qcow2")
	imageCmd.Flags().String("platform", "", "Image system platform (Linux, Windows, WindowsVirtio, Other, Paravirtualization)")
	imageCmd.Flags().String("architecture", "x86_64", "CPU architecture (x86_64, aarch64, mips64el)")
	imageCmd.Flags().Bool("virtio", false, "Whether to use virtio drivers")
	imageCmd.Flags().String("resource-uuid", "", "Resource UUID. If specified, the image will use this value as UUID")
	imageCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	imageCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	imageCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}
