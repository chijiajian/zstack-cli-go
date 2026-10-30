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

type FormattedImageGroup struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
	Status      string `json:"status" yaml:"status" header:"STATUS"`
	ImageCount  int    `json:"imageCount" yaml:"imageCount" header:"IMAGE COUNT"`
}

var imageGroupsCmd = &cobra.Command{
	Use:   "image-groups [name]",
	Short: "List image groups",
	Long:  `List image groups in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.ImageGroupInventoryView](cobraCmd, queryParam, "v1/imagegroups")
		groups := result.Items
		if err != nil {
			return fmt.Errorf("Error querying image groups: %w", err)
		}

		formattedResults := make([]FormattedImageGroup, 0, len(groups))
		for _, group := range groups {
			formattedResults = append(formattedResults, FormattedImageGroup{
				Name:        group.Name,
				UUID:        group.UUID,
				Description: group.Description,
				Status:      group.Status,
				ImageCount:  group.ImageCount,
			})
		}

		if err := common.PrintQuery(cobraCmd, formattedResults, result.Total); err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(imageGroupsCmd)
	common.AddQueryFlags(imageGroupsCmd)
}
