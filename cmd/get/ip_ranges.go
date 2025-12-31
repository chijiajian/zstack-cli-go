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

type FormattedIpRange struct {
	Name          string `json:"name" yaml:"name" header:"NAME"`
	UUID          string `json:"uuid" yaml:"uuid" header:"UUID"`
	L3NetworkUUID string `json:"l3NetworkUuid" yaml:"l3NetworkUuid" header:"L3 NETWORK"`
	StartIP       string `json:"startIp" yaml:"startIp" header:"START IP"`
	EndIP         string `json:"endIp" yaml:"endIp" header:"END IP"`
	Netmask       string `json:"netmask" yaml:"netmask" header:"NETMASK"`
	Gateway       string `json:"gateway" yaml:"gateway" header:"GATEWAY"`
	NetworkCidr   string `json:"networkCidr" yaml:"networkCidr" header:"CIDR"`
	IPVersion     int    `json:"ipVersion" yaml:"ipVersion" header:"IP VER"`
	AddressMode   string `json:"addressMode" yaml:"addressMode" header:"MODE"`
	PrefixLen     int    `json:"prefixLen" yaml:"prefixLen" header:"PREFIX"`
}

var ipRangesCmd = &cobra.Command{
	Use:     "ip-ranges [name]",
	Aliases: []string{"ip-range", "iprange"},
	Short:   "List IP ranges",
	Long:    `List all IP ranges for L3 networks in the ZStack cloud platform.`,
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

		ipRanges, err := zsClient.QueryIpRange(*queryParam)
		if err != nil {
			fmt.Printf("Error querying IP ranges: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedIpRange
		for _, ipRange := range ipRanges {
			formatted := FormattedIpRange{
				Name:          ipRange.Name,
				UUID:          ipRange.Uuid,
				L3NetworkUUID: ipRange.L3NetworkUuid,
				StartIP:       ipRange.StartIp,
				EndIP:         ipRange.EndIp,
				Netmask:       ipRange.Netmask,
				Gateway:       ipRange.Gateway,
				NetworkCidr:   ipRange.NetworkCidr,
				IPVersion:     ipRange.IpVersion,
				AddressMode:   ipRange.AddressMode,
				PrefixLen:     ipRange.PrefixLen,
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
	GetCmd.AddCommand(ipRangesCmd)
	common.AddQueryFlags(ipRangesCmd)
}
