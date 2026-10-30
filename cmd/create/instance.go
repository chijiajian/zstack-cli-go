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

type VmInstanceSpec struct {
	Name                 string   `json:"name" yaml:"name"`
	InstanceOffering     string   `json:"instanceOffering" yaml:"instanceOffering"`
	CpuNum               int64    `json:"cpuNum" yaml:"cpuNum"`
	MemorySize           string   `json:"memorySize" yaml:"memorySize"`
	Image                string   `json:"image" yaml:"image"`
	L3Networks           []string `json:"l3Networks" yaml:"l3Networks"`
	Type                 string   `json:"type" yaml:"type"`
	RootDiskOffering     string   `json:"rootDiskOffering" yaml:"rootDiskOffering"`
	RootDiskSize         string   `json:"rootDiskSize" yaml:"rootDiskSize"`
	DataDiskOfferings    []string `json:"dataDiskOfferings" yaml:"dataDiskOfferings"`
	DataDiskSizes        []string `json:"dataDiskSizes" yaml:"dataDiskSizes"`
	Zone                 string   `json:"zone" yaml:"zone"`
	Cluster              string   `json:"cluster" yaml:"cluster"`
	Host                 string   `json:"host" yaml:"host"`
	PrimaryStorage       string   `json:"primaryStorage" yaml:"primaryStorage"`
	Description          string   `json:"description" yaml:"description"`
	DefaultL3Network     string   `json:"defaultL3Network" yaml:"defaultL3Network"`
	ResourceUuid         string   `json:"resourceUuid" yaml:"resourceUuid"`
	TagUuids             []string `json:"tagUuids" yaml:"tagUuids"`
	Strategy             string   `json:"strategy" yaml:"strategy"`
	RootVolumeSystemTags []string `json:"rootVolumeSystemTags" yaml:"rootVolumeSystemTags"`
	DataVolumeSystemTags []string `json:"dataVolumeSystemTags" yaml:"dataVolumeSystemTags"`
	SystemTags           []string `json:"systemTags" yaml:"systemTags"`
	UserTags             []string `json:"userTags" yaml:"userTags"`
}

var instanceCmd = &cobra.Command{
	Use:   "instance [name]",
	Short: "Create a virtual machine instance",
	Long: `Create a virtual machine instance with specified parameters.

Examples:
  # Create VM instance with required parameters
  zstack-cli create instance my-vm --image 2162b130d30c49f2a3aad8585517e668 --instance-offering 2162b130d30c49f2a3aad8585517e668 --l3-network 2162b130d30c49f2a3aad8585517e668

  # Create VM instance with custom CPU and memory
  zstack-cli create instance my-vm --image 2162b130d30c49f2a3aad8585517e668 --cpu 4 --memory 8G --l3-network 2162b130d30c49f2a3aad8585517e668

  # Create VM instance from configuration file
  zstack-cli create instance -f vm-spec.yaml

  # Create VM instance in stopped state
  zstack-cli create instance my-vm --image 2162b130d30c49f2a3aad8585517e668 --instance-offering 2162b130d30c49f2a3aad8585517e668 --l3-network 2162b130d30c49f2a3aad8585517e668 --strategy CreateStopped`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if filePath != "" {
			return createVmInstanceFromFile(cmd, name, filePath)
		} else {
			if name == "" {
				cmd.Help()
				return fmt.Errorf("VM instance name is required when not using --file")
			}

			imageStr, _ := cmd.Flags().GetString("image")
			if imageStr == "" {
				return fmt.Errorf("--image is required")
			}

			l3NetworkStrs, _ := cmd.Flags().GetStringSlice("l3-network")
			if len(l3NetworkStrs) == 0 {
				return fmt.Errorf("at least one --l3-network is required")
			}

			instanceOfferingStr, _ := cmd.Flags().GetString("instance-offering")
			cpuNum, _ := cmd.Flags().GetInt64("cpu")
			memorySize, _ := cmd.Flags().GetString("memory")
			if instanceOfferingStr == "" && (cpuNum == 0 || memorySize == "") {
				return fmt.Errorf("either --instance-offering or both --cpu and --memory must be specified")
			}
			return createVmInstanceFromFlags(cmd, name)
		}
	},
}

func createVmInstanceFromFile(cmd *cobra.Command, name string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %w", filePath, err)
	}

	var vmSpec VmInstanceSpec
	var resourceSpec utils.ResourceSpec
	isGenericFormat := false

	// 通用抽象结构解析
	if err := yaml.Unmarshal(data, &resourceSpec); err == nil {
		if resourceSpec.Kind == utils.KindInstance && resourceSpec.Spec != nil {
			isGenericFormat = true
			specData, err := json.Marshal(resourceSpec.Spec)
			if err != nil {
				return fmt.Errorf("Error converting resource spec: %w", err)
			}
			if err := json.Unmarshal(specData, &vmSpec); err != nil {
				return fmt.Errorf("Error parsing Instance spec: %w", err)
			}
			if name == "" {
				name = resourceSpec.Metadata.Name
			}
		}
	}

	// 直接解析
	if !isGenericFormat {
		if strings.HasSuffix(filePath, ".json") {
			if err := json.Unmarshal(data, &vmSpec); err != nil {
				return fmt.Errorf("Error parsing JSON file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &vmSpec); err != nil {
				return fmt.Errorf("Error parsing YAML file: %w", err)
			}
		}
	}

	// 默认名字
	if name != "" {
		vmSpec.Name = name
	}
	if vmSpec.Name == "" {
		return fmt.Errorf("Error: VM instance name is required")
	}

	// 解析内存/磁盘大小
	var memorySizeBytes int64
	if vmSpec.MemorySize != "" {
		parsed, err := utils.ParseMemorySize(vmSpec.MemorySize)
		if err != nil {
			return fmt.Errorf("Error parsing memory size: %w", err)
		}
		memorySizeBytes = parsed
	}

	var rootDiskSizeBytes *int64
	if vmSpec.RootDiskSize != "" {
		parsed, err := utils.ParseMemorySize(vmSpec.RootDiskSize)
		if err != nil {
			return fmt.Errorf("Error parsing root disk size: %w", err)
		}
		rootDiskSizeBytes = &parsed
	}

	var dataDiskSizesBytes []int64
	for _, size := range vmSpec.DataDiskSizes {
		parsed, err := utils.ParseMemorySize(size)
		if err != nil {
			return fmt.Errorf("Error parsing data disk size: %w", err)
		}
		dataDiskSizesBytes = append(dataDiskSizesBytes, parsed)
	}

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Client not initialized. Please run 'zstack-cli login' first.")
	}

	// Image
	if vmSpec.Image == "" {
		return fmt.Errorf("Error: image is required")
	}
	imageUuid, err := client.GetImageUUIDByName(cli, vmSpec.Image)
	if err != nil {
		return fmt.Errorf("Error finding image '%s': %w", vmSpec.Image, err)
	}

	// InstanceOffering
	var instanceOfferingUuid string
	if vmSpec.InstanceOffering != "" {
		instanceOfferingUuid, err = client.GetInstanceOfferingUUIDByName(cli, vmSpec.InstanceOffering)
		if err != nil {
			return fmt.Errorf("Error finding instance offering '%s': %w", vmSpec.InstanceOffering, err)
		}
	}

	// L3Networks
	if len(vmSpec.L3Networks) == 0 {
		return fmt.Errorf("Error: At least one L3 network is required")
	}
	l3NetworkUuids := make([]string, 0, len(vmSpec.L3Networks))
	for _, n := range vmSpec.L3Networks {
		uuid, err := client.GetL3NetworkUUIDByName(cli, n)
		if err != nil {
			return fmt.Errorf("Error finding L3 network '%s': %w", n, err)
		}
		l3NetworkUuids = append(l3NetworkUuids, uuid)
	}

	// Zone
	var zoneUuid string
	if vmSpec.Zone != "" {
		zoneUuid, err = client.GetZoneUUIDByName(cli, vmSpec.Zone)
		if err != nil {
			return fmt.Errorf("Error finding zone '%s': %w", vmSpec.Zone, err)
		}
	}

	// Cluster
	var clusterUuid string
	if vmSpec.Cluster != "" {
		clusterUuid, err = client.GetClusterUUIDByName(cli, vmSpec.Cluster)
		if err != nil {
			return fmt.Errorf("Error finding cluster '%s': %w", vmSpec.Cluster, err)
		}
	}

	// Host
	var hostUuid string
	if vmSpec.Host != "" {
		hostUuid, err = client.GetHostUUIDByName(cli, vmSpec.Host)
		if err != nil {
			return fmt.Errorf("Error finding host '%s': %w", vmSpec.Host, err)
		}
	}

	// PrimaryStorage
	var primaryStorageUuid string
	if vmSpec.PrimaryStorage != "" {
		primaryStorageUuid, err = client.GetPrimaryStorageUUIDByName(cli, vmSpec.PrimaryStorage)
		if err != nil {
			return fmt.Errorf("Error finding primary storage '%s': %w", vmSpec.PrimaryStorage, err)
		}
	}
	var primaryStoragePtr *string
	if primaryStorageUuid != "" {
		primaryStoragePtr = &primaryStorageUuid
	}

	// DefaultL3Network
	var defaultL3NetworkUuid string
	if vmSpec.DefaultL3Network != "" {
		defaultL3NetworkUuid, err = client.GetL3NetworkUUIDByName(cli, vmSpec.DefaultL3Network)
		if err != nil {
			return fmt.Errorf("Error finding default L3 network '%s': %w", vmSpec.DefaultL3Network, err)
		}
	}

	vmParam := param.CreateVmInstanceParam{
		BaseParam: param.BaseParam{
			SystemTags: vmSpec.SystemTags,
			UserTags:   vmSpec.UserTags,
		},
		Params: param.CreateVmInstanceParamDetail{
			Name:                            vmSpec.Name,
			InstanceOfferingUuid:            stringPtrIfNotEmpty(instanceOfferingUuid),
			CpuNum:                          intPtrFromInt64IfNotZero(vmSpec.CpuNum),
			MemorySize:                      int64PtrIfNotZero(memorySizeBytes),
			ImageUuid:                       stringPtrIfNotEmpty(imageUuid),
			L3NetworkUuids:                  l3NetworkUuids,
			Type:                            stringPtrIfNotEmpty(vmSpec.Type),
			RootDiskOfferingUuid:            stringPtrIfNotEmpty(vmSpec.RootDiskOffering),
			RootDiskSize:                    rootDiskSizeBytes,
			DataDiskOfferingUuids:           vmSpec.DataDiskOfferings,
			DataDiskSizes:                   dataDiskSizesBytes,
			ZoneUuid:                        stringPtrIfNotEmpty(zoneUuid),
			ClusterUuid:                     stringPtrIfNotEmpty(clusterUuid),
			HostUuid:                        stringPtrIfNotEmpty(hostUuid),
			PrimaryStorageUuidForRootVolume: primaryStoragePtr,
			Description:                     stringPtrIfNotEmpty(vmSpec.Description),
			DefaultL3NetworkUuid:            stringPtrIfNotEmpty(defaultL3NetworkUuid),
			ResourceUuid:                    stringPtrIfNotEmpty(vmSpec.ResourceUuid),
			TagUuids:                        vmSpec.TagUuids,
			Strategy:                        stringPtrIfNotEmpty(vmSpec.Strategy),
			RootVolumeSystemTags:            vmSpec.RootVolumeSystemTags,
			DataVolumeSystemTags:            vmSpec.DataVolumeSystemTags,
		},
	}

	// ==== dry-run 支持 ====
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		format, _ := cmd.Flags().GetString("output")
		if format == "" {
			format = "yaml"
		}
		return utils.PrintDryRun(vmParam, format)
	}

	// ==== 调用 API ====
	resp, err := cli.CreateVmInstance(vmParam)
	if err != nil {
		return fmt.Errorf("Error creating VM instance: %w", err)
	}

	if err := printCreateProgress(cmd, "VM instance created successfully: %s\n", resp.UUID); err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("output")
	if format == "" {
		format = "table"
	}
	return utils.PrintOperationResult("instance", resp, format)
}

func createVmInstanceFromFlags(cmd *cobra.Command, name string) error {
	imageStr, _ := cmd.Flags().GetString("image")
	instanceOfferingStr, _ := cmd.Flags().GetString("instance-offering")
	cpuNum, _ := cmd.Flags().GetInt64("cpu")
	memorySize, _ := cmd.Flags().GetString("memory")
	l3NetworkStrs, _ := cmd.Flags().GetStringSlice("l3-network")
	zoneStr, _ := cmd.Flags().GetString("zone")
	clusterStr, _ := cmd.Flags().GetString("cluster")
	hostStr, _ := cmd.Flags().GetString("host")
	description, _ := cmd.Flags().GetString("description")
	rootDiskOfferingStr, _ := cmd.Flags().GetString("root-disk-offering")
	rootDiskSize, _ := cmd.Flags().GetString("root-disk-size")
	dataDiskOfferingStrs, _ := cmd.Flags().GetStringSlice("data-disk-offering")
	dataDiskSizes, _ := cmd.Flags().GetStringSlice("data-disk-size")
	primaryStorageStr, _ := cmd.Flags().GetString("primary-storage")
	defaultL3NetworkStr, _ := cmd.Flags().GetString("default-l3-network")
	resourceUuid, _ := cmd.Flags().GetString("resource-uuid")
	strategy, _ := cmd.Flags().GetString("strategy")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	if imageStr == "" {
		return fmt.Errorf("Error: --image is required")
	}
	if len(l3NetworkStrs) == 0 {
		return fmt.Errorf("Error: At least one --l3-network is required")
	}
	if instanceOfferingStr == "" && (cpuNum == 0 || memorySize == "") {
		return fmt.Errorf("Error: Either --instance-offering or both --cpu and --memory must be specified")
	}

	var memorySizeBytes int64
	if memorySize != "" {
		parsed, err := utils.ParseMemorySize(memorySize)
		if err != nil {
			return fmt.Errorf("Error parsing memory size: %w", err)
		}
		memorySizeBytes = parsed
	}

	var rootDiskSizeBytes *int64
	if rootDiskSize != "" {
		parsed, err := utils.ParseMemorySize(rootDiskSize)
		if err != nil {
			return fmt.Errorf("Error parsing root disk size: %w", err)
		}
		rootDiskSizeBytes = &parsed
	}

	var dataDiskSizesBytes []int64
	if len(dataDiskSizes) > 0 {
		for _, size := range dataDiskSizes {
			parsed, err := utils.ParseMemorySize(size)
			if err != nil {
				return fmt.Errorf("Error parsing data disk size: %w", err)
			}
			dataDiskSizesBytes = append(dataDiskSizesBytes, parsed)
		}
	}

	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Error: Client not initialized. Please run 'zstack-cli login' first.")
	}

	imageUuidValue, err := client.GetImageUUIDByName(cli, imageStr)
	if err != nil {
		return fmt.Errorf("Error finding image '%s': %w", imageStr, err)
	}

	var instanceOfferingUuidValue string
	if instanceOfferingStr != "" {
		instanceOfferingUuidValue, err = client.GetInstanceOfferingUUIDByName(cli, instanceOfferingStr)
		if err != nil {
			return fmt.Errorf("Error finding instance offering '%s': %w", instanceOfferingStr, err)
		}
	}

	l3NetworkUuidValues := make([]string, 0, len(l3NetworkStrs))
	for _, nameOrUUID := range l3NetworkStrs {
		uuid, err := client.GetL3NetworkUUIDByName(cli, nameOrUUID)
		if err != nil {
			return fmt.Errorf("Error finding L3 network '%s': %w", nameOrUUID, err)
		}
		l3NetworkUuidValues = append(l3NetworkUuidValues, uuid)
	}

	var zoneUuidValue string
	if zoneStr != "" {
		zoneUuidValue, err = client.GetZoneUUIDByName(cli, zoneStr)
		if err != nil {
			return fmt.Errorf("Error finding zone '%s': %w", zoneStr, err)
		}
	}

	var clusterUuidValue string
	if clusterStr != "" {
		clusterUuidValue, err = client.GetClusterUUIDByName(cli, clusterStr)
		if err != nil {
			return fmt.Errorf("Error finding cluster '%s': %w", clusterStr, err)
		}
	}

	var hostUuidValue string
	if hostStr != "" {
		hostUuidValue, err = client.GetHostUUIDByName(cli, hostStr)
		if err != nil {
			return fmt.Errorf("Error finding host '%s': %w", hostStr, err)
		}
	}

	var primaryStorageUuidValue string
	if primaryStorageStr != "" {
		primaryStorageUuidValue, err = client.GetPrimaryStorageUUIDByName(cli, primaryStorageStr)
		if err != nil {
			return fmt.Errorf("Error finding primary storage '%s': %w", primaryStorageStr, err)
		}
	}
	var primaryStoragePtr *string
	if primaryStorageUuidValue != "" {
		primaryStoragePtr = &primaryStorageUuidValue
	}

	var defaultL3NetworkUuidValue string
	if defaultL3NetworkStr != "" {
		defaultL3NetworkUuidValue, err = client.GetL3NetworkUUIDByName(cli, defaultL3NetworkStr)
		if err != nil {
			return fmt.Errorf("Error finding default L3 network '%s': %w", defaultL3NetworkStr, err)
		}
	}

	vmParam := param.CreateVmInstanceParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateVmInstanceParamDetail{
			Name:                            name,
			InstanceOfferingUuid:            stringPtrIfNotEmpty(instanceOfferingUuidValue),
			CpuNum:                          intPtrFromInt64IfNotZero(cpuNum),
			MemorySize:                      int64PtrIfNotZero(memorySizeBytes),
			ImageUuid:                       stringPtrIfNotEmpty(imageUuidValue),
			L3NetworkUuids:                  l3NetworkUuidValues,
			RootDiskOfferingUuid:            stringPtrIfNotEmpty(rootDiskOfferingStr),
			RootDiskSize:                    rootDiskSizeBytes,
			DataDiskOfferingUuids:           dataDiskOfferingStrs,
			DataDiskSizes:                   dataDiskSizesBytes,
			ZoneUuid:                        stringPtrIfNotEmpty(zoneUuidValue),
			ClusterUuid:                     stringPtrIfNotEmpty(clusterUuidValue),
			HostUuid:                        stringPtrIfNotEmpty(hostUuidValue),
			PrimaryStorageUuidForRootVolume: primaryStoragePtr,
			Description:                     stringPtrIfNotEmpty(description),
			DefaultL3NetworkUuid:            stringPtrIfNotEmpty(defaultL3NetworkUuidValue),
			ResourceUuid:                    stringPtrIfNotEmpty(resourceUuid),
			Strategy:                        stringPtrIfNotEmpty(strategy),
		},
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		format, _ := cmd.Flags().GetString("output")
		if format == "" {
			format = "yaml"
		}
		return utils.PrintDryRun(vmParam, format)
	}

	resp, err := cli.CreateVmInstance(vmParam)
	if err != nil {
		return fmt.Errorf("Error creating VM instance: %w", err)
	}

	if err := printCreateProgress(cmd, "VM instance created successfully: %s\n", resp.UUID); err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("output")
	if format == "" {
		format = "table"
	}

	return utils.PrintOperationResult("instance", resp, format)
}

func init() {
	CreateCmd.AddCommand(instanceCmd)

	instanceCmd.Flags().StringP("file", "f", "", "Path to YAML or JSON file containing VM instance specification")

	instanceCmd.Flags().String("image", "", "Image UUID (required)")
	instanceCmd.Flags().String("instance-offering", "", "Instance offering UUID")
	instanceCmd.Flags().StringSlice("l3-network", []string{}, "L3 network UUID(s) (required)")

	instanceCmd.Flags().Int64("cpu", 0, "Number of CPUs (alternative to instance-offering)")
	instanceCmd.Flags().String("memory", "", "Memory size (e.g., '8G', alternative to instance-offering)")

	instanceCmd.Flags().String("root-disk-offering", "", "Root disk offering UUID (required for ISO images)")
	instanceCmd.Flags().String("root-disk-size", "", "Root disk size (e.g., '100G')")
	instanceCmd.Flags().StringSlice("data-disk-offering", []string{}, "Data disk offering UUID(s)")
	instanceCmd.Flags().StringSlice("data-disk-size", []string{}, "Data disk size(s) (e.g., '100G,200G')")

	instanceCmd.Flags().String("zone", "", "Zone UUID")
	instanceCmd.Flags().String("cluster", "", "Cluster UUID")
	instanceCmd.Flags().String("host", "", "Host UUID")
	instanceCmd.Flags().String("primary-storage", "", "Primary storage UUID for root volume")

	instanceCmd.Flags().String("description", "", "Description for the VM instance")
	instanceCmd.Flags().String("default-l3-network", "", "Default L3 network UUID")
	instanceCmd.Flags().String("resource-uuid", "", "Resource UUID")
	instanceCmd.Flags().String("strategy", "InstantStart", "VM creation strategy: InstantStart or CreateStopped")

	instanceCmd.Flags().StringSlice("system-tag", []string{}, "System tag(s)")
	instanceCmd.Flags().StringSlice("user-tag", []string{}, "User tag(s)")

	instanceCmd.Flags().Bool("dry-run", false, "Preview the API request without sending it")
	instanceCmd.Flags().StringP("output", "o", "", "Output format: json, yaml, table, wide, or name")
}
