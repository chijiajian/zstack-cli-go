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
	"reflect"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type instanceUpdateRequest struct {
	Instance string                      `json:"instance" yaml:"instance"`
	Params   param.UpdateVmInstanceParam `json:"params" yaml:"params"`
}

var updateInstanceCmd = &cobra.Command{
	Use:   "update <name-or-uuid>",
	Short: "Update a VM instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		updateParam, err := buildInstanceUpdateParam(cmd)
		if err != nil {
			return err
		}
		if reflect.ValueOf(updateParam.Params).IsZero() {
			return fmt.Errorf("at least one update field must be specified")
		}

		if isInstanceDryRun(cmd) {
			return printInstanceDryRun(cmd, instanceUpdateRequest{Instance: args[0], Params: updateParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vm, err := zstackclient.ResolveOneResource(cli, "instance", args[0])
		if err != nil {
			return err
		}

		if updateParam.Params.DefaultL3NetworkUuid != nil {
			l3, err := zstackclient.ResolveOneResource(cli, "l3-network", *updateParam.Params.DefaultL3NetworkUuid)
			if err != nil {
				return err
			}
			updateParam.Params.DefaultL3NetworkUuid = &l3.UUID
		}

		result, err := cli.UpdateVmInstance(vm.UUID, updateParam)
		if err != nil {
			return err
		}
		return printInstanceResult(cmd, fmt.Sprintf("instance/%s updated", result.Name), result)
	},
}

func init() {
	updateInstanceCmd.Flags().String("name", "", "New VM name")
	updateInstanceCmd.Flags().StringP("description", "d", "", "New VM description")
	updateInstanceCmd.Flags().String("state", "", "VM state")
	updateInstanceCmd.Flags().String("default-l3-network", "", "Default L3 network name or UUID")
	updateInstanceCmd.Flags().String("platform", "", "VM platform")
	updateInstanceCmd.Flags().Int("cpu", 0, "CPU count")
	updateInstanceCmd.Flags().String("memory", "", "Memory size, for example 4G")
	updateInstanceCmd.Flags().String("reserved-memory", "", "Reserved memory size")
	updateInstanceCmd.Flags().String("guest-os-type", "", "Guest OS type")

	InstanceCmd.AddCommand(updateInstanceCmd)
}

func buildInstanceUpdateParam(cmd *cobra.Command) (param.UpdateVmInstanceParam, error) {
	memory, err := instanceMemoryFlag(cmd, "memory")
	if err != nil {
		return param.UpdateVmInstanceParam{}, err
	}
	reservedMemory, err := instanceMemoryFlag(cmd, "reserved-memory")
	if err != nil {
		return param.UpdateVmInstanceParam{}, err
	}
	cpu, _ := cmd.Flags().GetInt("cpu")

	return param.UpdateVmInstanceParam{
		Params: param.UpdateVmInstanceParamDetail{
			Name:                 changedInstanceString(cmd, "name"),
			Description:          changedInstanceStringPtr(cmd, "description"),
			State:                changedInstanceStringPtr(cmd, "state"),
			DefaultL3NetworkUuid: changedInstanceStringPtr(cmd, "default-l3-network"),
			Platform:             changedInstanceStringPtr(cmd, "platform"),
			CpuNum:               intPtrIfNotZero(cpu),
			MemorySize:           int64PtrIfNotZero(memory),
			ReservedMemorySize:   int64PtrIfNotZero(reservedMemory),
			GuestOsType:          changedInstanceStringPtr(cmd, "guest-os-type"),
		},
	}, nil
}

func changedInstanceString(cmd *cobra.Command, flag string) string {
	value, _ := cmd.Flags().GetString(flag)
	if !cmd.Flags().Changed(flag) {
		return ""
	}
	return value
}

func changedInstanceStringPtr(cmd *cobra.Command, flag string) *string {
	value, _ := cmd.Flags().GetString(flag)
	if !cmd.Flags().Changed(flag) {
		return nil
	}
	return &value
}

func intPtrIfNotZero(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func int64PtrIfNotZero(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func instanceMemoryFlag(cmd *cobra.Command, flag string) (int64, error) {
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
