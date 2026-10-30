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

package get

import (
	"fmt"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
)

type FormattedVmInstanceScript struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`

	ScriptContent string `json:"scriptContent" yaml:"scriptContent" header:"SCRIPT CONTENT"`
	RenderParams  string `json:"renderParams" yaml:"renderParams" header:"RENDER PARAMS"`
	Platform      string `json:"platform" yaml:"platform" header:"PLATFORM"`
	ScriptType    string `json:"scriptType" yaml:"scriptType" header:"SCRIPT TYPE"`
	ScriptTimeout int    `json:"scriptTimeout" yaml:"scriptTimeout" header:"SCRIPT TIMEOUT (SEC)"`
	EncodingType  string `json:"encodingType" yaml:"encodingType" header:"ENCODING TYPE"`
}

var vmScriptsCmd = &cobra.Command{
	Use:   "vm-scripts [name]",
	Short: "List VM instance scripts",
	Long:  `List all VM instance scripts in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.GuestVmScriptInventoryView](cobraCmd, queryParam, "v1/scripts")
		scripts := result.Items
		if err != nil {
			return fmt.Errorf("Error querying VM instance scripts: %w", err)
		}

		showFullScript, _ := cobraCmd.Flags().GetBool("show-full-script")

		var formattedResults []FormattedVmInstanceScript
		for _, script := range scripts {

			scriptContent := script.ScriptContent
			if !showFullScript && len(scriptContent) > 50 {
				scriptContent = scriptContent[:50] + "..."
			}

			formatted := FormattedVmInstanceScript{
				Name:        script.Name,
				UUID:        script.UUID,
				Description: script.Description,

				ScriptContent: scriptContent,
				RenderParams:  script.RenderParams,
				Platform:      script.Platform,
				ScriptType:    script.ScriptType,
				ScriptTimeout: script.ScriptTimeout,
				EncodingType:  script.EncodingType,
			}
			formattedResults = append(formattedResults, formatted)
		}

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(vmScriptsCmd)
	common.AddQueryFlags(vmScriptsCmd)
	vmScriptsCmd.Flags().Bool("show-full-script", false, "Show full script content instead of truncated version")
}
