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

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
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
	Run: func(cobraCmd *cobra.Command, args []string) {

		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		scripts, err := zsClient.QueryVmInstanceScript(*queryParam)
		if err != nil {
			fmt.Printf("Error querying VM instance scripts: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")
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

		err = utils.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
	},
}

func init() {
	GetCmd.AddCommand(vmScriptsCmd)
	common.AddQueryFlags(vmScriptsCmd)
	vmScriptsCmd.Flags().Bool("show-full-script", false, "Show full script content instead of truncated version")
}
