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
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.EipInventoryView](cobraCmd, queryParam, "v1/eips")
		eips := result.Items
		if err != nil {
			return fmt.Errorf("Error querying elastic IPs: %w", err)
		}

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

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(eipsCmd)

	common.AddQueryFlags(eipsCmd)
}
