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
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
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
	Use:   "disk-offering NAME",
	Short: "Create a new disk offering",
	Long: `Create a new disk offering in ZStack cloud platform.
	
Examples:
  # Create a basic disk offering
  zstack-cli create disk-offering small-disk --size 20G

  # Create a disk offering with more options
  zstack-cli create disk-offering medium-disk --size 100G --description "Medium Disk" --allocator-strategy "DefaultHostAllocatorStrategy"
  
  # Create a disk offering from a YAML or JSON file
  zstack-cli create disk-offering my-disk-offering -f disk-spec.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			createDiskOfferingFromFile(cmd, name, filePath)
			return
		}

		createDiskOfferingFromFlags(cmd, name)
	},
}

func createDiskOfferingFromFile(cmd *cobra.Command, name string, filePath string) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	var diskOfferingSpec DiskOfferingSpec

	if strings.HasSuffix(filePath, ".json") {
		if err := json.Unmarshal(data, &diskOfferingSpec); err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
			return
		}
	} else {
		if err := yaml.Unmarshal(data, &diskOfferingSpec); err != nil {
			fmt.Printf("Error parsing YAML file: %v\n", err)
			return
		}
	}

	if name != "" {
		diskOfferingSpec.Name = name
	} else if diskOfferingSpec.Name != "" {
		name = diskOfferingSpec.Name
	} else {
		fmt.Println("Error: name is required in disk offering specification")
		return
	}

	var diskSizeBytes int64
	var parseErr error
	if diskOfferingSpec.DiskSize != "" {
		diskSizeBytes, parseErr = utils.ParseMemorySize(diskOfferingSpec.DiskSize)
		if parseErr != nil {
			fmt.Printf("Error parsing disk size: %v\n", parseErr)
			return
		}
	} else {
		fmt.Println("Error: diskSize is required in disk offering specification")
		return
	}

	cli := client.GetClient()
	if cli == nil {
		fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
		return
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

	var sortKey *int
	if diskOfferingSpec.SortKey != 0 {
		sortKey = &diskOfferingSpec.SortKey
	}

	offeringParam := param.CreateDiskOfferingParam{
		Params: param.CreateDiskOfferingDetailParam{
			Name:              diskOfferingSpec.Name,
			Description:       description,
			DiskSize:          diskSizeBytes,
			AllocatorStrategy: allocatorStrategy,
			SortKey:           sortKey,
			Type:              offeringType,
			ResourceUuid:      resourceUuid,
			SystemTags:        diskOfferingSpec.SystemTags,
			UserTags:          diskOfferingSpec.UserTags,
		},
	}

	if dryRunFlag {
		utils.PrintDryRun(offeringParam, outputFlag)
		return
	}

	fmt.Printf("Creating disk offering '%s' from file...\n", diskOfferingSpec.Name)
	result, err := cli.CreateDiskOffering(&offeringParam)
	if err != nil {
		fmt.Printf("Error creating disk offering: %s\n", err)
		return
	}

	utils.PrintOperationResult("DiskOffering", result, outputFlag)
}

func createDiskOfferingFromFlags(cmd *cobra.Command, name string) {
	diskSizeStr, _ := cmd.Flags().GetString("size")

	if diskSizeStr == "" {
		fmt.Println("Error: required flag --size not set")
		cmd.Help()
		return
	}

	diskSizeBytes, err := utils.ParseMemorySize(diskSizeStr)
	if err != nil {
		fmt.Printf("Error parsing disk size: %v\n", err)
		return
	}

	cli := client.GetClient()
	if cli == nil {
		fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
		return
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

	var sortKeyPtr *int
	if cmd.Flags().Changed("sort-key") {
		sortKeyPtr = &sortKey
	}

	offeringParam := param.CreateDiskOfferingParam{
		Params: param.CreateDiskOfferingDetailParam{
			Name:              name,
			Description:       descPtr,
			DiskSize:          diskSizeBytes,
			AllocatorStrategy: allocatorPtr,
			SortKey:           sortKeyPtr,
			Type:              typePtr,
			ResourceUuid:      resourceUuidPtr,
			SystemTags:        systemTags,
			UserTags:          userTags,
		},
	}

	if dryRunFlag {
		utils.PrintDryRun(offeringParam, outputFlag)
		return
	}

	fmt.Printf("Creating disk offering '%s'...\n", name)
	result, err := cli.CreateDiskOffering(&offeringParam)
	if err != nil {
		fmt.Printf("Error creating disk offering: %s\n", err)
		return
	}
	utils.PrintOperationResult("DiskOffering", result, outputFlag)
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
