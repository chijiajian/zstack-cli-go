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

type FormattedDiskOffering struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	DiskSize          string `json:"diskSize" yaml:"diskSize" header:"DISK SIZE"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	AllocatorStrategy string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State             string `json:"state" yaml:"state" header:"STATE"`
}

var diskOfferingsCmd = &cobra.Command{
	Use:   "disk-offerings [name]",
	Short: "List disk offerings",
	Long:  `List all disk offerings in the ZStack cloud platform.`,
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

		diskOfferings, err := zsClient.QueryDiskOffering(*queryParam)
		if err != nil {
			fmt.Printf("Error querying disk offerings: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedDiskOffering
		for _, offering := range diskOfferings {
			formatted := FormattedDiskOffering{
				Name:              offering.Name,
				UUID:              offering.UUID,
				DiskSize:          utils.FormatDiskSize(offering.DiskSize),
				Type:              offering.Type,
				AllocatorStrategy: offering.AllocatorStrategy,
				State:             offering.State,
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
	GetCmd.AddCommand(diskOfferingsCmd)

	common.AddQueryFlags(diskOfferingsCmd)
}
