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

type FormattedEip struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	VipIp       string `json:"vipIp" yaml:"vipIp" header:"VIP IP"`
	GuestIp     string `json:"guestIp" yaml:"guestIp" header:"GUEST IP"`
	State       string `json:"state" yaml:"state" header:"STATE"`
	VmNicUuid   string `json:"vmNicUuid" yaml:"vmNicUuid" header:"VM NIC UUID"`
	VipUuid     string `json:"vipUuid" yaml:"vipUuid" header:"VIP UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

var eipsCmd = &cobra.Command{
	Use:   "eips [name]",
	Short: "List elastic IPs",
	Long:  `List all elastic IPs (EIPs) in the ZStack cloud platform.`,
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

		eips, err := zsClient.QueryEip(*queryParam)
		if err != nil {
			fmt.Printf("Error querying elastic IPs: %s\n", err)
			return
		}

		if len(eips) == 0 {
			fmt.Println("No elastic IPs found.")
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedEip
		for _, eip := range eips {
			formatted := FormattedEip{
				Name:        eip.Name,
				UUID:        eip.UUID,
				VipIp:       eip.VipIp,
				GuestIp:     eip.GuestIp,
				State:       eip.State,
				VmNicUuid:   eip.VmNicUuid,
				VipUuid:     eip.VipUuid,
				Description: eip.Description,
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
	GetCmd.AddCommand(eipsCmd)

	common.AddQueryFlags(eipsCmd)
}
