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
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type vmScriptExecuteRequest struct {
	Script string                          `json:"script" yaml:"script"`
	Params param.ExecuteGuestVmScriptParam `json:"params" yaml:"params"`
}

var VmScriptCmd = &cobra.Command{
	Use:   "vm-script",
	Short: "Manage VM scripts",
	Long:  "Execute ZStack guest VM scripts.",
}

var vmScriptExecuteCmd = &cobra.Command{
	Use:   "execute <name-or-uuid>",
	Short: "Execute a VM script on an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName, _ := cmd.Flags().GetString("instance")
		if instanceName == "" {
			return fmt.Errorf("required flag --instance not set")
		}
		params := buildExecuteVmScriptParam(cmd, instanceName)
		if isVmScriptDryRun(cmd) {
			return printVmScriptDryRun(cmd, vmScriptExecuteRequest{Script: args[0], Params: params})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		script, err := zstackclient.ResolveOneResource(cli, "vm-script", args[0])
		if err != nil {
			return err
		}
		instance, err := zstackclient.ResolveOneResource(cli, "instance", instanceName)
		if err != nil {
			return err
		}
		params.Params.VmInstanceUuids = []string{instance.UUID}

		result, err := cli.ExecuteGuestVmScript(script.UUID, params)
		if err != nil {
			return err
		}
		return printVmScriptResult(cmd, fmt.Sprintf("vm-script/%s executed on instance/%s", script.Name, instance.UUID), result)
	},
}

func init() {
	VmScriptCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	VmScriptCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	VmScriptCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	vmScriptExecuteCmd.Flags().String("instance", "", "VM instance name or UUID")
	vmScriptExecuteCmd.Flags().Int("timeout", 0, "Script timeout in seconds")
	vmScriptExecuteCmd.Flags().String("log-path", "", "Log path")
	vmScriptExecuteCmd.Flags().String("record-uuid", "", "Record UUID")
	vmScriptExecuteCmd.Flags().String("runtime-params", "", "Runtime parameters")

	VmScriptCmd.AddCommand(vmScriptExecuteCmd)
}

func buildExecuteVmScriptParam(cmd *cobra.Command, instance string) param.ExecuteGuestVmScriptParam {
	timeout, _ := cmd.Flags().GetInt("timeout")
	logPath, _ := cmd.Flags().GetString("log-path")
	recordUUID, _ := cmd.Flags().GetString("record-uuid")
	runtimeParams, _ := cmd.Flags().GetString("runtime-params")

	return param.ExecuteGuestVmScriptParam{
		Params: param.ExecuteGuestVmScriptParamDetail{
			VmInstanceUuids: []string{instance},
			ScriptTimeout:   intPtrIfNotZero(timeout),
			LogPath:         stringPtrIfNotEmpty(logPath),
			RecordUuid:      stringPtrIfNotEmpty(recordUUID),
			RuntimeParams:   stringPtrIfNotEmpty(runtimeParams),
		},
	}
}

func isVmScriptDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printVmScriptDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printVmScriptResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
