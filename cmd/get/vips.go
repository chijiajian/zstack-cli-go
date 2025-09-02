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

type FormattedVip struct {
	Name               string `json:"name" yaml:"name" header:"NAME"`
	UUID               string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description        string `json:"description" yaml:"description" header:"DESCRIPTION"`
	L3NetworkUUID      string `json:"l3NetworkUuid" yaml:"l3NetworkUuid" header:"L3 NETWORK UUID"`
	Ip                 string `json:"ip" yaml:"ip" header:"IP"`
	State              string `json:"state" yaml:"state" header:"STATE"`
	Gateway            string `json:"gateway" yaml:"gateway" header:"GATEWAY"`
	Netmask            string `json:"netmask" yaml:"netmask" header:"NETMASK"`
	PrefixLen          string `json:"prefixLen" yaml:"prefixLen" header:"PREFIX LENGTH"`
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

		vips, err := zsClient.QueryVip(*queryParam)
		if err != nil {
			fmt.Printf("Error querying virtual IPs: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

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

				L3NetworkUUID:      vip.L3NetworkUUID,
				Ip:                 vip.Ip,
				State:              vip.State,
				Gateway:            vip.Gateway,
				Netmask:            vip.Netmask,
				PrefixLen:          vip.PrefixLen,
				ServiceProvider:    vip.ServiceProvider,
				PeerL3NetworkUuids: vip.PeerL3NetworkUuids,
				UseFor:             vip.UseFor,
				System:             vip.System,
				ServicesTypes:      strings.Join(serviceTypes, ","),
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
	GetCmd.AddCommand(vipsCmd)
	common.AddQueryFlags(vipsCmd)
}
