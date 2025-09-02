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

		zones, err := zsClient.QueryZone(*queryParam)
		if err != nil {
			fmt.Printf("Error querying zones: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

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

		err = utils.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
	},
}

func init() {
	GetCmd.AddCommand(zonesCmd)
	common.AddQueryFlags(zonesCmd)
}
