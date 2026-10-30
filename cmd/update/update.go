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

package update

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	sdkClient "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type updateRequest struct {
	Resource string      `json:"resource" yaml:"resource"`
	Target   string      `json:"target" yaml:"target"`
	Params   interface{} `json:"params" yaml:"params"`
}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update resource metadata",
	Long:  "Update ZStack resource metadata such as name and description.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return nil
	},
}

func init() {
	UpdateCmd.PersistentFlags().Bool("dry-run", false, "Only print the request that would be sent, without sending it")
	UpdateCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	UpdateCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	UpdateCmd.AddCommand(
		newInstanceCmd(),
		newImageCmd(),
		newDiskOfferingCmd(),
		newInstanceOfferingCmd(),
		newVolumeCmd(),
		newL3NetworkCmd(),
		newVipCmd(),
		newEipCmd(),
		newSecurityGroupCmd(),
		newVmScriptCmd(),
		newZoneCmd(),
		newClusterCmd(),
		newHostCmd(),
		newPrimaryStorageCmd(),
		newBackupStorageCmd(),
		newL2NetworkCmd(),
	)
}

func newInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance <name-or-uuid>",
		Short: "Update a VM instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateVmInstanceParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "instance", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				if params.Params.DefaultL3NetworkUuid != nil {
					resolvedL3, err := zstackclient.ResolveOneResource(cli, "l3-network", *params.Params.DefaultL3NetworkUuid)
					if err != nil {
						return nil, err
					}
					params.Params.DefaultL3NetworkUuid = &resolvedL3.UUID
				}
				return cli.UpdateVmInstance(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("state", "", "VM state")
	cmd.Flags().String("default-l3-network", "", "Default L3 network name or UUID")
	cmd.Flags().String("platform", "", "VM platform")
	cmd.Flags().Int("cpu", 0, "CPU count")
	cmd.Flags().String("memory", "", "Memory size, for example 4G")
	cmd.Flags().String("reserved-memory", "", "Reserved memory size")
	cmd.Flags().String("guest-os-type", "", "Guest OS type")
	return cmd
}

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image <name-or-uuid>",
		Short: "Update an image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateImageParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "image", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateImage(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("guest-os-type", "", "Guest OS type")
	cmd.Flags().String("media-type", "", "Image media type")
	cmd.Flags().String("format", "", "Image format")
	cmd.Flags().String("platform", "", "Image platform")
	cmd.Flags().String("architecture", "", "Image architecture")
	cmd.Flags().Bool("system", false, "Whether this is a system image")
	cmd.Flags().Bool("virtio", false, "Whether virtio is enabled")
	return cmd
}

func newDiskOfferingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk-offering <name-or-uuid>",
		Short: "Update a disk offering",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateDiskOfferingParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "disk-offering", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateDiskOffering(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newInstanceOfferingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance-offering <name-or-uuid>",
		Short: "Update an instance offering",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateInstanceOfferingParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "instance-offering", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateInstanceOffering(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("allocator-strategy", "", "Host allocator strategy")
	return cmd
}

func newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume <name-or-uuid>",
		Short: "Update a volume",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateVolumeParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "volume", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateVolume(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newL3NetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "l3-network <name-or-uuid>",
		Short: "Update an L3 network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateL3NetworkParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "l3-network", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateL3Network(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("dns-domain", "", "DNS domain")
	cmd.Flags().String("category", "", "Network category")
	cmd.Flags().Bool("system", false, "Whether this is a system network")
	return cmd
}

func newVipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vip <name-or-uuid>",
		Short: "Update a VIP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateVipParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "vip", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateVip(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newEipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eip <name-or-uuid>",
		Short: "Update an EIP",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateEipParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "eip", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateEip(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newSecurityGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "security-group <name-or-uuid>",
		Aliases: []string{"sg"},
		Short:   "Update a security group",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateSecurityGroupParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "security-group", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateSecurityGroup(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newVmScriptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm-script <name-or-uuid>",
		Short: "Update a VM script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateVmScriptParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "vm-script", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateGuestVmScript(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("file", "", "Script file to read")
	cmd.Flags().String("content", "", "Inline script content")
	cmd.Flags().String("platform", "", "Script platform, for example Linux or Windows")
	cmd.Flags().String("encoding", "", "Script encoding type")
	cmd.Flags().String("type", "", "Script type")
	cmd.Flags().Int("timeout", 0, "Script timeout in seconds")
	cmd.Flags().String("render-params", "", "Render parameters")
	return cmd
}

func newZoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zone <name-or-uuid>",
		Short: "Update a zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateZoneParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "zone", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateZone(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().Bool("default", false, "Make this the default zone")
	return cmd
}

func newClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster <name-or-uuid>",
		Short: "Update a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateClusterParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "cluster", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateCluster(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newHostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host <name-or-uuid>",
		Short: "Update a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateHostParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "host", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateHost(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("management-ip", "", "Host management IP")
	return cmd
}

func newPrimaryStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "primary-storage <name-or-uuid>",
		Aliases: []string{"ps"},
		Short:   "Update a primary storage",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdatePrimaryStorageParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "primary-storage", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdatePrimaryStorage(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	cmd.Flags().String("url", "", "Primary storage URL")
	return cmd
}

func newBackupStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backup-storage <name-or-uuid>",
		Aliases: []string{"bs", "image-storage"},
		Short:   "Update a backup storage",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateBackupStorageParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "backup-storage", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateBackupStorage(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func newL2NetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "l2-network <name-or-uuid>",
		Short: "Update an L2 network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := buildUpdateL2NetworkParam(cmd)
			if err != nil {
				return err
			}
			return runUpdate(cmd, "l2-network", args[0], &params, func(cli *sdkClient.ZSClient, uuid string) (interface{}, error) {
				return cli.UpdateL2Network(uuid, params)
			})
		},
	}
	addNameDescriptionFlags(cmd)
	return cmd
}

func addNameDescriptionFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "New resource name")
	cmd.Flags().StringP("description", "d", "", "New resource description")
}

func runUpdate(cmd *cobra.Command, resourceType string, target string, params interface{}, updateFn func(cli *sdkClient.ZSClient, uuid string) (interface{}, error)) error {
	if isEmptyUpdate(params) {
		return fmt.Errorf("at least one update field must be specified")
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		output, _ := cmd.Flags().GetString("output")
		if output == "" || output == "text" {
			output = "yaml"
		}
		return utils.PrintDryRun(updateRequest{Resource: resourceType, Target: target, Params: params}, output)
	}

	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}

	resolved, err := zstackclient.ResolveOneResource(cli, resourceType, target)
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("name") {
		preserveCurrentName(params, resolved.Name)
	}

	result, err := updateFn(cli, resolved.UUID)
	if err != nil {
		return err
	}
	return printUpdateResult(cmd, resourceType, result)
}

func printUpdateResult(cmd *cobra.Command, resourceType string, result interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	fields, _ := cmd.Flags().GetStringSlice("fields")

	switch strings.ToLower(output) {
	case "", "text":
		fmt.Printf("%s/%s updated\n", resourceType, resultName(result))
		return nil
	case "name":
		fmt.Printf("%s/%s\n", resourceType, resultName(result))
		return nil
	default:
		return utils.PrintWithFields(result, utils.ParseFormat(output), fields)
	}
}

func resultName(result interface{}) string {
	value := reflect.ValueOf(result)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return ""
	}
	field := value.FieldByName("Name")
	if field.IsValid() && field.Kind() == reflect.String && field.String() != "" {
		return field.String()
	}
	field = value.FieldByName("UUID")
	if field.IsValid() && field.Kind() == reflect.String {
		return field.String()
	}
	return ""
}

func preserveCurrentName(params interface{}, name string) {
	if name == "" {
		return
	}
	value := reflect.ValueOf(params)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}
	paramsField := value.FieldByName("Params")
	if paramsField.Kind() == reflect.Ptr {
		paramsField = paramsField.Elem()
	}
	if !paramsField.IsValid() || paramsField.Kind() != reflect.Struct {
		return
	}
	nameField := paramsField.FieldByName("Name")
	if !nameField.IsValid() || !nameField.CanSet() || nameField.Kind() != reflect.String || nameField.String() != "" {
		return
	}
	nameField.SetString(name)
}

func isEmptyUpdate(params interface{}) bool {
	value := reflect.ValueOf(params)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	paramsField := value.FieldByName("Params")
	if !paramsField.IsValid() {
		return true
	}
	return paramsField.IsZero()
}

func buildUpdateVmInstanceParam(cmd *cobra.Command) (param.UpdateVmInstanceParam, error) {
	description, _ := cmd.Flags().GetString("description")
	defaultL3Network, _ := cmd.Flags().GetString("default-l3-network")
	memory, err := durationMemoryFlag(cmd, "memory")
	if err != nil {
		return param.UpdateVmInstanceParam{}, err
	}
	reservedMemory, err := durationMemoryFlag(cmd, "reserved-memory")
	if err != nil {
		return param.UpdateVmInstanceParam{}, err
	}
	cpu, _ := cmd.Flags().GetInt("cpu")

	detail := param.UpdateVmInstanceParamDetail{
		Name:                 changedString(cmd, "name"),
		Description:          changedStringPtr(description),
		State:                changedStringPtrFlag(cmd, "state"),
		DefaultL3NetworkUuid: changedStringPtr(defaultL3Network),
		Platform:             changedStringPtrFlag(cmd, "platform"),
		CpuNum:               changedIntPtr(cpu),
		MemorySize:           changedInt64Ptr(memory),
		ReservedMemorySize:   changedInt64Ptr(reservedMemory),
		GuestOsType:          changedStringPtrFlag(cmd, "guest-os-type"),
	}
	return param.UpdateVmInstanceParam{Params: detail}, nil
}

func buildUpdateImageParam(cmd *cobra.Command) (param.UpdateImageParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateImageParamDetail{
		Name:         changedString(cmd, "name"),
		Description:  changedStringPtr(description),
		GuestOsType:  changedStringPtrFlag(cmd, "guest-os-type"),
		MediaType:    changedStringPtrFlag(cmd, "media-type"),
		Format:       changedStringPtrFlag(cmd, "format"),
		Platform:     changedStringPtrFlag(cmd, "platform"),
		Architecture: changedStringPtrFlag(cmd, "architecture"),
		System:       changedBoolPtr(cmd, "system"),
		Virtio:       changedBoolPtr(cmd, "virtio"),
	}
	return param.UpdateImageParam{Params: detail}, nil
}

func buildUpdateDiskOfferingParam(cmd *cobra.Command) (param.UpdateDiskOfferingParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateDiskOfferingParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateDiskOfferingParam{Params: detail}, nil
}

func buildUpdateInstanceOfferingParam(cmd *cobra.Command) (param.UpdateInstanceOfferingParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateInstanceOfferingParamDetail{
		Name:              changedString(cmd, "name"),
		Description:       changedStringPtr(description),
		AllocatorStrategy: changedStringPtrFlag(cmd, "allocator-strategy"),
	}
	return param.UpdateInstanceOfferingParam{Params: detail}, nil
}

func buildUpdateVolumeParam(cmd *cobra.Command) (param.UpdateVolumeParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateVolumeParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateVolumeParam{Params: detail}, nil
}

func buildUpdateL3NetworkParam(cmd *cobra.Command) (param.UpdateL3NetworkParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateL3NetworkParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
		DnsDomain:   changedStringPtrFlag(cmd, "dns-domain"),
		Category:    changedStringPtrFlag(cmd, "category"),
		System:      changedBoolPtr(cmd, "system"),
	}
	return param.UpdateL3NetworkParam{Params: detail}, nil
}

func buildUpdateVipParam(cmd *cobra.Command) (param.UpdateVipParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateVipParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateVipParam{Params: detail}, nil
}

func buildUpdateEipParam(cmd *cobra.Command) (param.UpdateEipParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateEipParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateEipParam{Params: detail}, nil
}

func buildUpdateSecurityGroupParam(cmd *cobra.Command) (param.UpdateSecurityGroupParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateSecurityGroupParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateSecurityGroupParam{Params: detail}, nil
}

func buildUpdateVmScriptParam(cmd *cobra.Command) (param.UpdateGuestVmScriptParam, error) {
	description, _ := cmd.Flags().GetString("description")
	scriptContent, err := changedScriptContent(cmd)
	if err != nil {
		return param.UpdateGuestVmScriptParam{}, err
	}
	timeout, _ := cmd.Flags().GetInt("timeout")

	detail := param.UpdateGuestVmScriptParamDetail{
		Name:          changedString(cmd, "name"),
		Description:   changedStringPtr(description),
		EncodingType:  changedStringPtrFlag(cmd, "encoding"),
		ScriptContent: scriptContent,
		RenderParams:  changedStringPtrFlag(cmd, "render-params"),
		Platform:      changedStringPtrFlag(cmd, "platform"),
		ScriptType:    changedStringPtrFlag(cmd, "type"),
		ScriptTimeout: changedIntPtr(timeout),
	}
	return param.UpdateGuestVmScriptParam{Params: detail}, nil
}

func buildUpdateZoneParam(cmd *cobra.Command) (param.UpdateZoneParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateZoneParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
		IsDefault:   changedBoolPtr(cmd, "default"),
	}
	return param.UpdateZoneParam{Params: detail}, nil
}

func buildUpdateClusterParam(cmd *cobra.Command) (param.UpdateClusterParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateClusterParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateClusterParam{Params: detail}, nil
}

func buildUpdateHostParam(cmd *cobra.Command) (param.UpdateHostParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateHostParamDetail{
		Name:         changedString(cmd, "name"),
		Description:  changedStringPtr(description),
		ManagementIp: changedStringPtrFlag(cmd, "management-ip"),
	}
	return param.UpdateHostParam{Params: detail}, nil
}

func buildUpdatePrimaryStorageParam(cmd *cobra.Command) (param.UpdatePrimaryStorageParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdatePrimaryStorageParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
		Url:         changedStringPtrFlag(cmd, "url"),
	}
	return param.UpdatePrimaryStorageParam{Params: detail}, nil
}

func buildUpdateBackupStorageParam(cmd *cobra.Command) (param.UpdateBackupStorageParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateBackupStorageParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateBackupStorageParam{Params: detail}, nil
}

func buildUpdateL2NetworkParam(cmd *cobra.Command) (param.UpdateL2NetworkParam, error) {
	description, _ := cmd.Flags().GetString("description")
	detail := param.UpdateL2NetworkParamDetail{
		Name:        changedString(cmd, "name"),
		Description: changedStringPtr(description),
	}
	return param.UpdateL2NetworkParam{Params: detail}, nil
}

func changedString(cmd *cobra.Command, flag string) string {
	value, _ := cmd.Flags().GetString(flag)
	if !cmd.Flags().Changed(flag) {
		return ""
	}
	return value
}

func changedStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func changedStringPtrFlag(cmd *cobra.Command, flag string) *string {
	value, _ := cmd.Flags().GetString(flag)
	if !cmd.Flags().Changed(flag) {
		return nil
	}
	return &value
}

func changedBoolPtr(cmd *cobra.Command, flag string) *bool {
	value, _ := cmd.Flags().GetBool(flag)
	if !cmd.Flags().Changed(flag) {
		return nil
	}
	return &value
}

func changedIntPtr(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func changedInt64Ptr(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func durationMemoryFlag(cmd *cobra.Command, flag string) (int64, error) {
	value, _ := cmd.Flags().GetString(flag)
	if value == "" {
		return 0, nil
	}
	parsed, err := utils.ParseMemorySize(value)
	if err != nil {
		return 0, fmt.Errorf("invalid --%s: %w", flag, err)
	}
	return parsed, nil
}

func changedScriptContent(cmd *cobra.Command) (*string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	content, _ := cmd.Flags().GetString("content")
	if cmd.Flags().Changed("file") && cmd.Flags().Changed("content") {
		return nil, fmt.Errorf("use only one of --file or --content")
	}
	if cmd.Flags().Changed("content") {
		return &content, nil
	}
	if !cmd.Flags().Changed("file") {
		return nil, nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file %q: %w", filePath, err)
	}
	value := string(data)
	return &value, nil
}
