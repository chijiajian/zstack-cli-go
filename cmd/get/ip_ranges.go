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
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.IpRangeInventoryView](cobraCmd, queryParam, "v1/l3-networks/ip-ranges")
		ipRanges := result.Items
		if err != nil {
			return fmt.Errorf("Error querying IP ranges: %w", err)
		}

		var formattedResults []FormattedIpRange
		for _, ipRange := range ipRanges {
			formatted := FormattedIpRange{
				Name:          ipRange.Name,
				UUID:          ipRange.UUID,
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

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(ipRangesCmd)
	common.AddQueryFlags(ipRangesCmd)
}
