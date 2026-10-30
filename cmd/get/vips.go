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
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
)

type FormattedVip struct {
	Name               string `json:"name" yaml:"name" header:"NAME"`
	UUID               string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description        string `json:"description" yaml:"description" header:"DESCRIPTION"`
	L3NetworkUUID      string `json:"l3NetworkUuid" yaml:"l3NetworkUuid" header:"L3 NETWORK UUID"`
	Ip                 string `json:"ip" yaml:"ip" header:"IP"`
	State              string `json:"state" yaml:"state" header:"STATE"`
	Gateway            string `json:"gateway" yaml:"gateway" header:"GATEWAY"`
	Netmask            string `json:"netmask" yaml:"netmask" header:"NETMASK"`
	PrefixLen          int    `json:"prefixLen" yaml:"prefixLen" header:"PREFIX LENGTH"`
	ServiceProvider    string `json:"serviceProvider" yaml:"serviceProvider" header:"SERVICE PROVIDER"`
	PeerL3NetworkUuids string `json:"peerL3NetworkUuids" yaml:"peerL3NetworkUuids" header:"PEER L3 NETWORK UUIDS"`
	UseFor             string `json:"useFor" yaml:"useFor" header:"USE FOR"`
	System             bool   `json:"system" yaml:"system" header:"SYSTEM"`
	ServicesTypes      string `json:"servicesTypes" yaml:"servicesTypes" header:"SERVICES TYPES"`
}

var vipsCmd = &cobra.Command{
	Use:   "vips [name]",
	Short: "List virtual IPs",
	Long:  `List all virtual IPs in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.VipInventoryView](cobraCmd, queryParam, "v1/vips")
		vips := result.Items
		if err != nil {
			return fmt.Errorf("Error querying virtual IPs: %w", err)
		}

		var formattedResults []FormattedVip
		for _, vip := range vips {

			serviceTypes := make([]string, 0, len(vip.ServicesRefs))
			for _, ref := range vip.ServicesRefs {
				serviceTypes = append(serviceTypes, ref.ServiceType)
			}

			formatted := FormattedVip{
				Name:        vip.Name,
				UUID:        vip.UUID,
				Description: vip.Description,

				L3NetworkUUID:      vip.L3NetworkUuid,
				Ip:                 vip.Ip,
				State:              vip.State,
				Gateway:            vip.Gateway,
				Netmask:            vip.Netmask,
				PrefixLen:          vip.PrefixLen,
				ServiceProvider:    vip.ServiceProvider,
				PeerL3NetworkUuids: strings.Join(vip.PeerL3NetworkUuids, ","),
				UseFor:             vip.UseFor,
				System:             vip.System,
				ServicesTypes:      strings.Join(serviceTypes, ","),
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
	GetCmd.AddCommand(vipsCmd)
	common.AddQueryFlags(vipsCmd)
}
