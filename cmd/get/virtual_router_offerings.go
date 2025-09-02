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

type FormattedVirtualRouterOffering struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`

	CpuNum                int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	CpuSpeed              int    `json:"cpuSpeed" yaml:"cpuSpeed" header:"CPU SPEED"`
	MemorySize            string `json:"memorySize" yaml:"memorySize" header:"MEMORY SIZE"`
	Type                  string `json:"type" yaml:"type" header:"TYPE"`
	AllocatorStrategy     string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	SortKey               int    `json:"sortKey" yaml:"sortKey" header:"SORT KEY"`
	State                 string `json:"state" yaml:"state" header:"STATE"`
	ManagementNetworkUuid string `json:"managementNetworkUuid" yaml:"managementNetworkUuid" header:"MANAGEMENT NETWORK UUID"`
	PublicNetworkUuid     string `json:"publicNetworkUuid" yaml:"publicNetworkUuid" header:"PUBLIC NETWORK UUID"`
	ZoneUuid              string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	ImageUuid             string `json:"imageUuid" yaml:"imageUuid" header:"IMAGE UUID"`
	IsDefault             bool   `json:"isDefault" yaml:"isDefault" header:"IS DEFAULT"`
	ReservedMemorySize    string `json:"reservedMemorySize" yaml:"reservedMemorySize" header:"RESERVED MEMORY SIZE"`
}

var virtualRouterOfferingsCmd = &cobra.Command{
	Use:   "virtual-router-offerings [name]",
	Short: "List virtual router offerings",
	Long:  `List all virtual router offerings in the ZStack cloud platform.`,
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

		offerings, err := zsClient.QueryVirtualRouterOffering(*queryParam)
		if err != nil {
			fmt.Printf("Error querying virtual router offerings: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedVirtualRouterOffering
		for _, offering := range offerings {
			formatted := FormattedVirtualRouterOffering{
				Name:        offering.Name,
				UUID:        offering.UUID,
				Description: offering.Description,

				CpuNum:                offering.CpuNum,
				CpuSpeed:              offering.CpuSpeed,
				MemorySize:            utils.FormatMemorySize(offering.MemorySize),
				Type:                  offering.Type,
				AllocatorStrategy:     offering.AllocatorStrategy,
				SortKey:               offering.SortKey,
				State:                 offering.State,
				ManagementNetworkUuid: offering.ManagementNetworkUuid,
				PublicNetworkUuid:     offering.PublicNetworkUuid,
				ZoneUuid:              offering.ZoneUuid,
				ImageUuid:             offering.ImageUuid,
				IsDefault:             offering.IsDefault,
				ReservedMemorySize:    offering.ReservedMemorySize,
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
	GetCmd.AddCommand(virtualRouterOfferingsCmd)
	common.AddQueryFlags(virtualRouterOfferingsCmd)
}
