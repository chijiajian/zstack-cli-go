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
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type FormattedInstanceOffering struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	CpuNum            int    `json:"cpuNum" yaml:"cpuNum" header:"CPU"`
	MemorySize        string `json:"memorySize" yaml:"memorySize" header:"MEMORY"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	AllocatorStrategy string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State             string `json:"state" yaml:"state" header:"STATE"`
}

var instanceOfferingsCmd = &cobra.Command{
	Use:   "instance-offerings [name]",
	Short: "List instance offerings",
	Long:  `List all instance offerings in the ZStack cloud platform.`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.InstanceOfferingInventoryView](cobraCmd, queryParam, "v1/instance-offerings")
		instanceOfferings := result.Items
		if err != nil {
			return fmt.Errorf("Error querying instance offerings: %w", err)
		}

		var formattedResults []FormattedInstanceOffering
		for _, offering := range instanceOfferings {
			formatted := FormattedInstanceOffering{
				Name:              offering.Name,
				UUID:              offering.UUID,
				CpuNum:            offering.CpuNum,
				MemorySize:        utils.FormatMemorySize(offering.MemorySize),
				Type:              offering.Type,
				AllocatorStrategy: offering.AllocatorStrategy,
				State:             offering.State,
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
	GetCmd.AddCommand(instanceOfferingsCmd)

	common.AddQueryFlags(instanceOfferingsCmd)
}
