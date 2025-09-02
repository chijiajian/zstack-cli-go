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

	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type FormattedL3Network struct {
	Name          string `json:"name" yaml:"name" header:"NAME"`
	UUID          string `json:"uuid" yaml:"uuid" header:"UUID"`
	Type          string `json:"type" yaml:"type" header:"TYPE"`
	State         string `json:"state" yaml:"state" header:"STATE"`
	L2NetworkUuid string `json:"l2NetworkUuid" yaml:"l2NetworkUuid" header:"L2 NETWORK UUID"`
	ZoneUuid      string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	IpVersion     int    `json:"ipVersion" yaml:"ipVersion" header:"IP VERSION"`
	DnsDomain     string `json:"dnsDomain" yaml:"dnsDomain" header:"DNS DOMAIN"`
	Dns           string `json:"dns" yaml:"dns" header:"DNS"`
	System        bool   `json:"system" yaml:"system" header:"SYSTEM"`
	Category      string `json:"category" yaml:"category" header:"CATEGORY"`
	EnableIPAM    bool   `json:"enableIPAM" yaml:"enableIPAM" header:"ENABLE IPAM"`
}

var l3NetworksCmd = &cobra.Command{
	Use:   "l3-networks [name]",
	Short: "List L3 networks",
	Long:  `List all Layer 3 networks in the ZStack cloud platform.`,
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

		l3Networks, err := zsClient.QueryL3Network(*queryParam)
		if err != nil {
			fmt.Printf("Error querying L3 networks: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedL3Network
		for _, network := range l3Networks {
			formatted := FormattedL3Network{
				Name:          network.Name,
				UUID:          network.UUID,
				Type:          network.Type,
				State:         network.State,
				L2NetworkUuid: network.L2NetworkUuid,
				ZoneUuid:      network.ZoneUuid,
				IpVersion:     network.IpVersion,
				DnsDomain:     network.DnsDomain,
				Dns:           strings.Join(network.Dns, ","),
				System:        network.System,
				Category:      network.Category,
				EnableIPAM:    network.EnableIPAM,
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
	GetCmd.AddCommand(l3NetworksCmd)

	common.AddQueryFlags(l3NetworksCmd)

}
