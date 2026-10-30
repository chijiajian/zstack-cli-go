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

type FormattedZone struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`

	State string `json:"state" yaml:"state" header:"STATE"`
	Type  string `json:"type" yaml:"type" header:"TYPE"`
}

var zonesCmd = &cobra.Command{
	Use:   "zones [name]",
	Short: "List zones",
	Long:  `List all zones in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.ZoneInventoryView](cobraCmd, queryParam, "v1/zones")
		zones := result.Items
		if err != nil {
			return fmt.Errorf("Error querying zones: %w", err)
		}

		var formattedResults []FormattedZone
		for _, zone := range zones {
			formatted := FormattedZone{
				Name:        zone.Name,
				UUID:        zone.UUID,
				Description: zone.Description,

				State: zone.State,
				Type:  zone.Type,
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
	GetCmd.AddCommand(zonesCmd)
	common.AddQueryFlags(zonesCmd)
}
