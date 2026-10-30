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
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.L2NetworkInventoryView](cobraCmd, queryParam, "v1/l2-networks")
		l2Networks := result.Items
		if err != nil {
			return fmt.Errorf("Error querying L2 networks: %w", err)
		}

		var formattedResults []FormattedL2Network
		for _, network := range l2Networks {
			formatted := FormattedL2Network{
				Name:              network.Name,
				UUID:              network.UUID,
				Type:              network.Type,
				Vlan:              network.VirtualNetworkId,
				PhysicalInterface: network.PhysicalInterface,
				ZoneUuid:          network.ZoneUuid,
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
	GetCmd.AddCommand(l2NetworksCmd)

	common.AddQueryFlags(l2NetworksCmd)
}
