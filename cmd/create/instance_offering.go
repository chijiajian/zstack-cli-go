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
	"io/ioutil"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
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
	Use:   "instance-offering NAME",
	Short: "Create a new instance offering",
	Long: `Create a new instance offering in ZStack cloud platform.
	
Examples:
  # Create a basic instance offering
  zstack-cli create instance-offering small-vm --cpu 1 --memory 1G

  # Create an instance offering with more options
  zstack-cli create instance-offering medium-vm --cpu 2 --memory 4G --description "Medium VM" --allocator-strategy "LastHostPreferredAllocatorStrategy"
  
  # Create an instance offering from a YAML or JSON file
  zstack-cli create instance-offering my-offering -f offering-spec.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			createInstanceOfferingFromFile(cmd, name, filePath)
			return
		}

		createInstanceOfferingFromFlags(cmd, name)
	},
}

func createInstanceOfferingFromFile(cmd *cobra.Command, name string, filePath string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	var instanceOfferingSpec InstanceOfferingSpec
	var resourceSpec utils.ResourceSpec

	isGenericFormat := false
	if err := yaml.Unmarshal(data, &resourceSpec); err == nil {
		if resourceSpec.Kind == utils.KindInstanceOffering && resourceSpec.Spec != nil {
			isGenericFormat = true

			specData, err := json.Marshal(resourceSpec.Spec)
			if err != nil {
				fmt.Printf("Error converting resource spec: %v\n", err)
				return
			}
			if err := json.Unmarshal(specData, &instanceOfferingSpec); err != nil {
				fmt.Printf("Error parsing instance offering spec from generic format: %v\n", err)
				return
			}

			if name == "" && resourceSpec.Metadata.Name != "" {
				name = resourceSpec.Metadata.Name
			}
		}
	}

	if !isGenericFormat {
		if strings.HasSuffix(filePath, ".json") {
			if err := json.Unmarshal(data, &instanceOfferingSpec); err != nil {
				fmt.Printf("Error parsing JSON file: %v\n", err)
				return
			}
		} else {
			if err := yaml.Unmarshal(data, &instanceOfferingSpec); err != nil {
				fmt.Printf("Error parsing YAML file: %v\n", err)
				return
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
		fmt.Println("Error: name is required in instance offering specification")
		return
	}

	if instanceOfferingSpec.CpuNum <= 0 {
		fmt.Println("Error: cpuNum must be greater than 0")
		return
	}

	memoryBytes, err := utils.ParseMemorySize(instanceOfferingSpec.MemorySize)
	if err != nil {
		fmt.Printf("Error parsing memory size: %v\n", err)
		return
	}

	cli := client.GetClient()
	if cli == nil {
		fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
		return
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

	var sortKey *int
	if instanceOfferingSpec.SortKey != 0 {
		sortKey = &instanceOfferingSpec.SortKey
	}

	offeringParam := param.CreateInstanceOfferingParam{
		Params: param.CreateInstanceOfferingDetailParam{
			Name:              instanceOfferingSpec.Name,
			Description:       description,
			CpuNum:            instanceOfferingSpec.CpuNum,
			MemorySize:        memoryBytes,
			AllocatorStrategy: allocatorStrategy,
			SortKey:           sortKey,
			Type:              offeringType,
			ResourceUuid:      resourceUuid,
			TagUuids:          instanceOfferingSpec.TagUUIDs,
		},
	}

	if dryRunFlag {
		utils.PrintDryRun(offeringParam, outputFlag)
		return
	}

	fmt.Printf("Creating instance offering '%s' from file...\n", instanceOfferingSpec.Name)
	result, err := cli.CreateInstanceOffering(&offeringParam)
	if err != nil {
		fmt.Printf("Error creating instance offering: %s\n", err)
		return
	}

	utils.PrintOperationResult("InstanceOffering", result, outputFlag)
}

func createInstanceOfferingFromFlags(cmd *cobra.Command, name string) {

	cpuNum, _ := cmd.Flags().GetInt("cpu")
	memoryStr, _ := cmd.Flags().GetString("memory")

	if cpuNum <= 0 {
		fmt.Println("Error: --cpu must be greater than 0")
		cmd.Help()
		return
	}

	if memoryStr == "" {
		fmt.Println("Error: required flag --memory not set")
		cmd.Help()
		return
	}

	memoryBytes, err := utils.ParseMemorySize(memoryStr)
	if err != nil {
		fmt.Printf("Error parsing memory size: %v\n", err)
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

	var sortKeyPtr *int
	if cmd.Flags().Changed("sort-key") {
		sortKeyPtr = &sortKey
	}

	offeringParam := param.CreateInstanceOfferingParam{
		Params: param.CreateInstanceOfferingDetailParam{
			Name:              name,
			Description:       descPtr,
			CpuNum:            cpuNum,
			MemorySize:        memoryBytes,
			AllocatorStrategy: allocatorPtr,
			SortKey:           sortKeyPtr,
			Type:              typePtr,
			ResourceUuid:      resourceUuidPtr,
			TagUuids:          tagUuids,
		},
	}

	if dryRunFlag {
		utils.PrintDryRun(offeringParam, outputFlag)
		return
	}

	fmt.Printf("Creating instance offering '%s'...\n", name)
	result, err := cli.CreateInstanceOffering(&offeringParam)
	if err != nil {
		fmt.Printf("Error creating instance offering: %s\n", err)
		return
	}

	utils.PrintOperationResult("InstanceOffering", result, outputFlag)
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
