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

// cmd/create/instance_offering.go
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

type InstanceOfferingSpec struct {
	Name              string   `json:"name" yaml:"name"`
	Description       string   `json:"description" yaml:"description"`
	CpuNum            int      `json:"cpuNum" yaml:"cpuNum"`
	MemorySize        string   `json:"memorySize" yaml:"memorySize"`
	AllocatorStrategy string   `json:"allocatorStrategy" yaml:"allocatorStrategy"`
	SortKey           int      `json:"sortKey" yaml:"sortKey"`
	Type              string   `json:"type" yaml:"type"`
	ResourceUUID      string   `json:"resourceUuid" yaml:"resourceUuid"`
	TagUUIDs          []string `json:"tagUuids" yaml:"tagUuids"`
	SystemTags        []string `json:"systemTags" yaml:"systemTags"`
	UserTags          []string `json:"userTags" yaml:"userTags"`
}

var instanceOfferingCmd = &cobra.Command{
	Use:   "instance-offering [NAME]",
	Short: "Create a new instance offering",
	Long: `Create a new instance offering in ZStack cloud platform.
	
Examples:
  # Create a basic instance offering
  zstack-cli create instance-offering small-vm --cpu 1 --memory 1G

  # Create an instance offering with more options
  zstack-cli create instance-offering medium-vm --cpu 2 --memory 4G --description "Medium VM" --allocator-strategy "LastHostPreferredAllocatorStrategy"
  
  # Create an instance offering from a YAML or JSON file
  zstack-cli create instance-offering my-offering -f offering-spec.yaml`,
	Args: exactOneArgUnlessFile,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			return createInstanceOfferingFromFile(cmd, name, filePath)
		}

		cpuNum, _ := cmd.Flags().GetInt("cpu")
		if cpuNum <= 0 {
			cmd.Help()
			return fmt.Errorf("--cpu must be greater than 0")
		}

		memory, _ := cmd.Flags().GetString("memory")
		if memory == "" {
			cmd.Help()
			return fmt.Errorf("required flag --memory not set")
		}

		return createInstanceOfferingFromFlags(cmd, name)
	},
}

func createInstanceOfferingFromFile(cmd *cobra.Command, name string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %w", filePath, err)
	}

	var instanceOfferingSpec InstanceOfferingSpec
	var resourceSpec utils.ResourceSpec

	isGenericFormat := false
	if err := yaml.Unmarshal(data, &resourceSpec); err == nil {
		if resourceSpec.Kind == utils.KindInstanceOffering && resourceSpec.Spec != nil {
			isGenericFormat = true

			specData, err := json.Marshal(resourceSpec.Spec)
			if err != nil {
				return fmt.Errorf("Error converting resource spec: %w", err)
			}
			if err := json.Unmarshal(specData, &instanceOfferingSpec); err != nil {
				return fmt.Errorf("Error parsing instance offering spec from generic format: %w", err)
			}

			if name == "" && resourceSpec.Metadata.Name != "" {
				name = resourceSpec.Metadata.Name
			}
		}
	}

	if !isGenericFormat {
		if strings.HasSuffix(filePath, ".json") {
			if err := json.Unmarshal(data, &instanceOfferingSpec); err != nil {
				return fmt.Errorf("Error parsing JSON file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &instanceOfferingSpec); err != nil {
				return fmt.Errorf("Error parsing YAML file: %w", err)
			}
		}

		if name == "" && instanceOfferingSpec.Name != "" {
			name = instanceOfferingSpec.Name
		}
	}

	if name != "" {
		instanceOfferingSpec.Name = name
	}

	if instanceOfferingSpec.Name == "" {
		return fmt.Errorf("Error: name is required in instance offering specification")
	}

	if instanceOfferingSpec.CpuNum <= 0 {
		return fmt.Errorf("Error: cpuNum must be greater than 0")
	}

	memoryBytes, err := utils.ParseMemorySize(instanceOfferingSpec.MemorySize)
	if err != nil {
		return fmt.Errorf("Error parsing memory size: %w", err)
	}

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Not logged in. Please run 'zstack-cli login' first.")
	}

	var description *string
	if instanceOfferingSpec.Description != "" {
		description = &instanceOfferingSpec.Description
	}

	var allocatorStrategy *string
	if instanceOfferingSpec.AllocatorStrategy != "" {
		allocatorStrategy = &instanceOfferingSpec.AllocatorStrategy
	}

	var offeringType *string
	if instanceOfferingSpec.Type != "" {
		offeringType = &instanceOfferingSpec.Type
	}

	var resourceUuid *string
	if instanceOfferingSpec.ResourceUUID != "" {
		resourceUuid = &instanceOfferingSpec.ResourceUUID
	}

	offeringParam := param.CreateInstanceOfferingParam{
		BaseParam: param.BaseParam{
			SystemTags: instanceOfferingSpec.SystemTags,
			UserTags:   instanceOfferingSpec.UserTags,
		},
		Params: param.CreateInstanceOfferingParamDetail{
			Name:              instanceOfferingSpec.Name,
			Description:       description,
			CpuNum:            instanceOfferingSpec.CpuNum,
			MemorySize:        memoryBytes,
			AllocatorStrategy: allocatorStrategy,
			SortKey:           instanceOfferingSpec.SortKey,
			Type:              offeringType,
			ResourceUuid:      resourceUuid,
			TagUuids:          instanceOfferingSpec.TagUUIDs,
		},
	}

	if dryRunFlag {
		return utils.PrintDryRun(offeringParam, outputFlag)
	}

	if err := printCreateProgress(cmd, "Creating instance offering '%s' from file...\n", instanceOfferingSpec.Name); err != nil {
		return err
	}
	result, err := cli.CreateInstanceOffering(offeringParam)
	if err != nil {
		return fmt.Errorf("Error creating instance offering: %w", err)
	}

	return utils.PrintOperationResult("InstanceOffering", result, outputFlag)
}

func createInstanceOfferingFromFlags(cmd *cobra.Command, name string) error {

	cpuNum, _ := cmd.Flags().GetInt("cpu")
	memoryStr, _ := cmd.Flags().GetString("memory")

	if cpuNum <= 0 {
		return fmt.Errorf("Error: --cpu must be greater than 0")
	}

	if memoryStr == "" {
		return fmt.Errorf("Error: required flag --memory not set")
	}

	memoryBytes, err := utils.ParseMemorySize(memoryStr)
	if err != nil {
		return fmt.Errorf("Error parsing memory size: %w", err)
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
	tagUuids, _ := cmd.Flags().GetStringSlice("tag")

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

	offeringParam := param.CreateInstanceOfferingParam{
		Params: param.CreateInstanceOfferingParamDetail{
			Name:              name,
			Description:       descPtr,
			CpuNum:            cpuNum,
			MemorySize:        memoryBytes,
			AllocatorStrategy: allocatorPtr,
			SortKey:           sortKey,
			Type:              typePtr,
			ResourceUuid:      resourceUuidPtr,
			TagUuids:          tagUuids,
		},
	}

	if dryRunFlag {
		return utils.PrintDryRun(offeringParam, outputFlag)
	}

	if err := printCreateProgress(cmd, "Creating instance offering '%s'...\n", name); err != nil {
		return err
	}
	result, err := cli.CreateInstanceOffering(offeringParam)
	if err != nil {
		return fmt.Errorf("Error creating instance offering: %w", err)
	}

	return utils.PrintOperationResult("InstanceOffering", result, outputFlag)
}

func init() {
	CreateCmd.AddCommand(instanceOfferingCmd)

	instanceOfferingCmd.Flags().StringP("file", "f", "", "Path to YAML or JSON file containing instance offering specification")

	instanceOfferingCmd.Flags().IntP("cpu", "c", 0, "Number of CPUs (required when not using --file)")
	instanceOfferingCmd.Flags().StringP("memory", "m", "", "Memory size (e.g., 1G, 1024M) (required when not using --file)")

	instanceOfferingCmd.Flags().StringP("description", "d", "", "Detailed description of the instance offering")
	instanceOfferingCmd.Flags().String("allocator-strategy", "", "Host allocation strategy")
	instanceOfferingCmd.Flags().Int("sort-key", 0, "Sort key")
	instanceOfferingCmd.Flags().String("type", "", "Instance offering type")
	instanceOfferingCmd.Flags().String("resource-uuid", "", "Resource UUID. If specified, the instance offering will use this value as UUID")
	instanceOfferingCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
}
