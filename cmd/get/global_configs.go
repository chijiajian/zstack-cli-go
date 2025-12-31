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

type FormattedGlobalConfig struct {
	Category     string `json:"category" yaml:"category" header:"CATEGORY"`
	Name         string `json:"name" yaml:"name" header:"NAME"`
	Description  string `json:"description" yaml:"description" header:"DESCRIPTION"`
	DefaultValue string `json:"defaultValue" yaml:"defaultValue" header:"DEFAULT"`
	Value        string `json:"value" yaml:"value" header:"VALUE"`
}

var globalConfigsCmd = &cobra.Command{
	Use:     "global-configs [name]",
	Aliases: []string{"global-config", "configs", "config"},
	Short:   "List global configurations",
	Long:    `List all global configurations in the ZStack cloud platform.`,
	Args:    cobra.MaximumNArgs(1),
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

		// Filter by category if specified
		category, _ := cobraCmd.Flags().GetString("category")
		if category != "" {
			queryParam.AddQ(fmt.Sprintf("category=%s", category))
		}

		configs, err := zsClient.QueryGlobalConfig(*queryParam)
		if err != nil {
			fmt.Printf("Error querying global configs: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedGlobalConfig
		for _, config := range configs {
			formatted := FormattedGlobalConfig{
				Category:     config.Category,
				Name:         config.Name,
				Description:  config.Description,
				DefaultValue: config.DefaultValue,
				Value:        config.Value,
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
	GetCmd.AddCommand(globalConfigsCmd)
	common.AddQueryFlags(globalConfigsCmd)
	globalConfigsCmd.Flags().String("category", "", "Filter by configuration category")
}
