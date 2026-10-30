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

// cmd/create/disk_offering.go
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
	"gopkg.in/yaml.v3"
)

type DiskOfferingSpec struct {
	Name              string   `json:"name" yaml:"name"`
	Description       string   `json:"description" yaml:"description"`
	DiskSize          string   `json:"diskSize" yaml:"diskSize"`
	AllocatorStrategy string   `json:"allocatorStrategy" yaml:"allocatorStrategy"`
	SortKey           int      `json:"sortKey" yaml:"sortKey"`
	Type              string   `json:"type" yaml:"type"`
	ResourceUUID      string   `json:"resourceUuid" yaml:"resourceUuid"`
	SystemTags        []string `json:"systemTags" yaml:"systemTags"`
	UserTags          []string `json:"userTags" yaml:"userTags"`
}

var diskOfferingCmd = &cobra.Command{
	Use:   "disk-offering [NAME]",
	Short: "Create a new disk offering",
	Long: `Create a new disk offering in ZStack cloud platform.
	
Examples:
  # Create a basic disk offering
  zstack-cli create disk-offering small-disk --size 20G

  # Create a disk offering with more options
  zstack-cli create disk-offering medium-disk --size 100G --description "Medium Disk" --allocator-strategy "DefaultHostAllocatorStrategy"
  
  # Create a disk offering from a YAML or JSON file
  zstack-cli create disk-offering my-disk-offering -f disk-spec.yaml`,
	Args: exactOneArgUnlessFile,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			return createDiskOfferingFromFile(cmd, name, filePath)
		}

		diskSize, _ := cmd.Flags().GetString("size")
		if diskSize == "" {
			cmd.Help()
			return fmt.Errorf("required flag --size not set")
		}

		return createDiskOfferingFromFlags(cmd, name)
	},
}

func createDiskOfferingFromFile(cmd *cobra.Command, name string, filePath string) error {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %w", filePath, err)
	}

	var diskOfferingSpec DiskOfferingSpec

	if strings.HasSuffix(filePath, ".json") {
		if err := json.Unmarshal(data, &diskOfferingSpec); err != nil {
			return fmt.Errorf("Error parsing JSON file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &diskOfferingSpec); err != nil {
			return fmt.Errorf("Error parsing YAML file: %w", err)
		}
	}

	if name != "" {
		diskOfferingSpec.Name = name
	} else if diskOfferingSpec.Name != "" {
		name = diskOfferingSpec.Name
	} else {
		return fmt.Errorf("Error: name is required in disk offering specification")
	}

	var diskSizeBytes int64
	var parseErr error
	if diskOfferingSpec.DiskSize != "" {
		diskSizeBytes, parseErr = utils.ParseMemorySize(diskOfferingSpec.DiskSize)
		if parseErr != nil {
			return fmt.Errorf("Error parsing disk size: %w", parseErr)
		}
	} else {
		return fmt.Errorf("Error: diskSize is required in disk offering specification")
	}

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Not logged in. Please run 'zstack-cli login' first.")
	}

	var description *string
	if diskOfferingSpec.Description != "" {
		description = &diskOfferingSpec.Description
	}

	var allocatorStrategy *string
	if diskOfferingSpec.AllocatorStrategy != "" {
		allocatorStrategy = &diskOfferingSpec.AllocatorStrategy
	}

	var offeringType *string
	if diskOfferingSpec.Type != "" {
		offeringType = &diskOfferingSpec.Type
	}

	var resourceUuid *string
	if diskOfferingSpec.ResourceUUID != "" {
		resourceUuid = &diskOfferingSpec.ResourceUUID
	}

	offeringParam := param.CreateDiskOfferingParam{
		BaseParam: param.BaseParam{
			SystemTags: diskOfferingSpec.SystemTags,
			UserTags:   diskOfferingSpec.UserTags,
		},
		Params: param.CreateDiskOfferingParamDetail{
			Name:               diskOfferingSpec.Name,
			Description:        description,
			DiskSize:           diskSizeBytes,
			AllocationStrategy: allocatorStrategy,
			SortKey:            diskOfferingSpec.SortKey,
			Type:               offeringType,
			ResourceUuid:       resourceUuid,
		},
	}

	if dryRunFlag {
		return utils.PrintDryRun(offeringParam, outputFlag)
	}

	if err := printCreateProgress(cmd, "Creating disk offering '%s' from file...\n", diskOfferingSpec.Name); err != nil {
		return err
	}
	result, err := cli.CreateDiskOffering(offeringParam)
	if err != nil {
		return fmt.Errorf("Error creating disk offering: %w", err)
	}

	return utils.PrintOperationResult("DiskOffering", result, outputFlag)
}

func createDiskOfferingFromFlags(cmd *cobra.Command, name string) error {
	diskSizeStr, _ := cmd.Flags().GetString("size")

	if diskSizeStr == "" {
		return fmt.Errorf("Error: required flag --size not set")
	}

	diskSizeBytes, err := utils.ParseMemorySize(diskSizeStr)
	if err != nil {
		return fmt.Errorf("Error parsing disk size: %w", err)
	}

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Not logged in. Please run 'zstack-cli login' first.")
	}
	description, _ := cmd.Flags().GetString("description")
	allocatorStrategy, _ := cmd.Flags().GetString("allocator-strategy")
	sortKey, _ := cmd.Flags().GetInt("sort-key")
	offeringType, _ := cmd.Flags().GetString("type")
	resourceUuid, _ := cmd.Flags().GetString("resource-uuid")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	var allocatorPtr *string
	if allocatorStrategy != "" {
		allocatorPtr = &allocatorStrategy
	}

	var typePtr *string
	if offeringType != "" {
		typePtr = &offeringType
	}

	var resourceUuidPtr *string
	if resourceUuid != "" {
		resourceUuidPtr = &resourceUuid
	}

	offeringParam := param.CreateDiskOfferingParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateDiskOfferingParamDetail{
			Name:               name,
			Description:        descPtr,
			DiskSize:           diskSizeBytes,
			AllocationStrategy: allocatorPtr,
			SortKey:            sortKey,
			Type:               typePtr,
			ResourceUuid:       resourceUuidPtr,
		},
	}

	if dryRunFlag {
		return utils.PrintDryRun(offeringParam, outputFlag)
	}

	if err := printCreateProgress(cmd, "Creating disk offering '%s'...\n", name); err != nil {
		return err
	}
	result, err := cli.CreateDiskOffering(offeringParam)
	if err != nil {
		return fmt.Errorf("Error creating disk offering: %w", err)
	}
	return utils.PrintOperationResult("DiskOffering", result, outputFlag)
}

func init() {
	CreateCmd.AddCommand(diskOfferingCmd)

	diskOfferingCmd.Flags().StringP("file", "f", "", "Path to YAML or JSON file containing disk offering specification")
	diskOfferingCmd.Flags().StringP("size", "s", "", "Disk size (e.g., 20G, 20480M) (required when not using --file)")
	diskOfferingCmd.Flags().StringP("description", "d", "", "Detailed description of the disk offering")
	diskOfferingCmd.Flags().String("allocator-strategy", "", "Disk allocation strategy")
	diskOfferingCmd.Flags().Int("sort-key", 0, "Sort key")
	diskOfferingCmd.Flags().String("type", "", "Disk offering type")
	diskOfferingCmd.Flags().String("resource-uuid", "", "Resource UUID. If specified, the disk offering will use this value as UUID")
	diskOfferingCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	diskOfferingCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}
