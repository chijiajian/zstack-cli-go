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

type FormattedL2Network struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	Vlan              int    `json:"vlan" yaml:"vlan" header:"VLAN"`
	PhysicalInterface string `json:"physicalInterface" yaml:"physicalInterface" header:"PHYSICAL INTERFACE"`
	ZoneUuid          string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
}

var l2NetworksCmd = &cobra.Command{
	Use:   "l2-networks [name]",
	Short: "List L2 networks",
	Long:  `List all Layer 2 networks in the ZStack cloud platform.`,
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

		l2Networks, err := zsClient.QueryL2Network(*queryParam)
		if err != nil {
			fmt.Printf("Error querying L2 networks: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedL2Network
		for _, network := range l2Networks {
			formatted := FormattedL2Network{
				Name:              network.Name,
				UUID:              network.UUID,
				Type:              network.Type,
				Vlan:              network.Vlan,
				PhysicalInterface: network.PhysicalInterface,
				ZoneUuid:          network.ZoneUuid,
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
	GetCmd.AddCommand(l2NetworksCmd)

	common.AddQueryFlags(l2NetworksCmd)
}
