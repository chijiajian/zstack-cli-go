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
	"strings"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type globalConfigSetRequest struct {
	Category string                        `json:"category" yaml:"category"`
	Name     string                        `json:"name" yaml:"name"`
	Params   param.UpdateGlobalConfigParam `json:"params" yaml:"params"`
}

var GlobalConfigCmd = &cobra.Command{
	Use:   "global-config",
	Short: "Manage global configurations",
	Long:  "Set ZStack global configuration values.",
}

var globalConfigSetCmd = &cobra.Command{
	Use:   "set <category.name>",
	Short: "Set a global configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		category, name, err := splitGlobalConfigName(args[0])
		if err != nil {
			return err
		}
		value, _ := cmd.Flags().GetString("value")
		if value == "" && !cmd.Flags().Changed("value") {
			return fmt.Errorf("required flag --value not set")
		}
		updateParam := param.UpdateGlobalConfigParam{
			Params: param.UpdateGlobalConfigParamDetail{Value: &value},
		}
		if isGlobalConfigDryRun(cmd) {
			return printGlobalConfigDryRun(cmd, globalConfigSetRequest{Category: category, Name: name, Params: updateParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		result, err := cli.UpdateGlobalConfig(category, name, updateParam)
		if err != nil {
			return err
		}
		return printGlobalConfigResult(cmd, fmt.Sprintf("global-config/%s.%s set", category, name), result)
	},
}

func init() {
	GlobalConfigCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	GlobalConfigCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	GlobalConfigCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	globalConfigSetCmd.Flags().String("value", "", "Configuration value")
	GlobalConfigCmd.AddCommand(globalConfigSetCmd)
}

func splitGlobalConfigName(value string) (string, string, error) {
	category, name, ok := strings.Cut(value, ".")
	if !ok || category == "" || name == "" {
		return "", "", fmt.Errorf("global config must be formatted as <category.name>")
	}
	return category, name, nil
}

func isGlobalConfigDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printGlobalConfigDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printGlobalConfigResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
