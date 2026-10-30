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
	"fmt"
	"os"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

var vmScriptCmd = &cobra.Command{
	Use:   "vm-script <name>",
	Short: "Create a VM script",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptParam, err := buildCreateVmScriptParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(scriptParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		result, err := cli.CreateGuestVmScript(scriptParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("vm-script", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(vmScriptCmd)
	vmScriptCmd.Flags().String("content", "", "Inline script content")
	vmScriptCmd.Flags().String("platform", "", "Script platform, for example Linux or Windows")
	vmScriptCmd.Flags().String("encoding", "PlainText", "Script encoding type")
	vmScriptCmd.Flags().String("type", "Shell", "Script type")
	vmScriptCmd.Flags().Int("timeout", 0, "Script timeout in seconds")
	vmScriptCmd.Flags().StringP("description", "d", "", "Detailed description of the script")
	vmScriptCmd.Flags().String("render-params", "", "Render parameters")
	vmScriptCmd.Flags().String("resource-uuid", "", "Resource UUID")
	vmScriptCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	vmScriptCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	vmScriptCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateVmScriptParam(cmd *cobra.Command, name string) (param.CreateGuestVmScriptParam, error) {
	platform, _ := cmd.Flags().GetString("platform")
	if platform == "" {
		return param.CreateGuestVmScriptParam{}, fmt.Errorf("required flag --platform not set")
	}
	scriptContent, err := readVmScriptContent(cmd)
	if err != nil {
		return param.CreateGuestVmScriptParam{}, err
	}

	description, _ := cmd.Flags().GetString("description")
	encodingType, _ := cmd.Flags().GetString("encoding")
	scriptType, _ := cmd.Flags().GetString("type")
	timeout, _ := cmd.Flags().GetInt("timeout")
	renderParams, _ := cmd.Flags().GetString("render-params")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateGuestVmScriptParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateGuestVmScriptParamDetail{
			Name:          name,
			Description:   stringPtrIfNotEmpty(description),
			EncodingType:  encodingType,
			ScriptContent: scriptContent,
			RenderParams:  stringPtrIfNotEmpty(renderParams),
			Platform:      platform,
			ScriptType:    scriptType,
			ScriptTimeout: intPtrIfNotZero(timeout),
			ResourceUuid:  stringPtrIfNotEmpty(resourceUUID),
			TagUuids:      tagUUIDs,
		},
	}, nil
}

func readVmScriptContent(cmd *cobra.Command) (string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	content, _ := cmd.Flags().GetString("content")
	if filePath != "" && content != "" {
		return "", fmt.Errorf("use only one of --file or --content")
	}
	if content != "" {
		return content, nil
	}
	if filePath == "" {
		return "", fmt.Errorf("required flag --file or --content not set")
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read script file %q: %w", filePath, err)
	}
	return string(data), nil
}
