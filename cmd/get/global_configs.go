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
	"github.com/chijiajian/zstack-cli-go/pkg/query"
	"github.com/spf13/cobra"
)

type FormattedGlobalConfig struct {
	ID           int64  `json:"id" yaml:"id" header:"ID"`
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
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		// Filter by category if specified
		category, _ := cobraCmd.Flags().GetString("category")
		if category != "" {
			queryParam.AddQ(fmt.Sprintf("category=%s", category))
		}

		result, err := common.Fetch[view.GlobalConfigInventoryView](cobraCmd, queryParam, "v1/global-configurations", query.WithDefaultSort("+id"))
		configs := result.Items
		if err != nil {
			return fmt.Errorf("Error querying global configs: %w", err)
		}

		var formattedResults []FormattedGlobalConfig
		for _, config := range configs {
			formatted := FormattedGlobalConfig{
				ID:           config.Id,
				Category:     config.Category,
				Name:         config.Name,
				Description:  config.Description,
				DefaultValue: config.DefaultValue,
				Value:        config.Value,
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
	GetCmd.AddCommand(globalConfigsCmd)
	common.AddQueryFlags(globalConfigsCmd)
	globalConfigsCmd.Flags().String("category", "", "Filter by configuration category")
}
