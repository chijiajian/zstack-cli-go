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
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type FormattedPrimaryStorage struct {
	Name                           string `json:"name" yaml:"name" header:"NAME"`
	UUID                           string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description                    string `json:"description" yaml:"description" header:"DESCRIPTION"`
	ZoneUuid                       string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	Url                            string `json:"url" yaml:"url" header:"URL"`
	TotalCapacity                  string `json:"totalCapacity" yaml:"totalCapacity" header:"TOTAL CAPACITY"`
	TotalCapacityBytes             *int64 `json:"totalCapacityBytes" yaml:"totalCapacityBytes" header:"TOTAL CAPACITY BYTES" table:"opt-in"`
	AvailableCapacity              string `json:"availableCapacity" yaml:"availableCapacity" header:"AVAILABLE CAPACITY"`
	AvailableCapacityBytes         *int64 `json:"availableCapacityBytes" yaml:"availableCapacityBytes" header:"AVAILABLE CAPACITY BYTES" table:"opt-in"`
	TotalPhysicalCapacity          string `json:"totalPhysicalCapacity" yaml:"totalPhysicalCapacity" header:"TOTAL PHYSICAL CAPACITY"`
	TotalPhysicalCapacityBytes     *int64 `json:"totalPhysicalCapacityBytes" yaml:"totalPhysicalCapacityBytes" header:"TOTAL PHYSICAL CAPACITY BYTES" table:"opt-in"`
	AvailablePhysicalCapacity      string `json:"availablePhysicalCapacity" yaml:"availablePhysicalCapacity" header:"AVAILABLE PHYSICAL CAPACITY"`
	AvailablePhysicalCapacityBytes *int64 `json:"availablePhysicalCapacityBytes" yaml:"availablePhysicalCapacityBytes" header:"AVAILABLE PHYSICAL CAPACITY BYTES" table:"opt-in"`
	SystemUsedCapacity             string `json:"systemUsedCapacity" yaml:"systemUsedCapacity" header:"SYSTEM USED CAPACITY"`
	SystemUsedCapacityBytes        *int64 `json:"systemUsedCapacityBytes" yaml:"systemUsedCapacityBytes" header:"SYSTEM USED CAPACITY BYTES" table:"opt-in"`
	Type                           string `json:"type" yaml:"type" header:"TYPE"`
	State                          string `json:"state" yaml:"state" header:"STATE"`
	Status                         string `json:"status" yaml:"status" header:"STATUS"`
	AttachedClusterUuids           string `json:"attachedClusterUuids" yaml:"attachedClusterUuids" header:"ATTACHED CLUSTER UUIDS"`
}

type primaryStorageQueryInventory struct {
	view.PrimaryStorageInventoryView
	TotalCapacityBytes             *int64 `json:"totalCapacity"`
	AvailableCapacityBytes         *int64 `json:"availableCapacity"`
	TotalPhysicalCapacityBytes     *int64 `json:"totalPhysicalCapacity"`
	AvailablePhysicalCapacityBytes *int64 `json:"availablePhysicalCapacity"`
	SystemUsedCapacityBytes        *int64 `json:"systemUsedCapacity"`
}

var primaryStorageCmd = &cobra.Command{
	Use:   "primary-storages [name]",
	Short: "List primary storages",
	Long: `List all primary storages in the ZStack cloud platform.
Structured output includes totalCapacityBytes, availableCapacityBytes, totalPhysicalCapacityBytes, availablePhysicalCapacityBytes, systemUsedCapacityBytes, or null if not reported.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[primaryStorageQueryInventory](cobraCmd, queryParam, "v1/primary-storage")
		primaryStorages := result.Items
		if err != nil {
			return fmt.Errorf("Error querying primary storages: %w", err)
		}

		var formattedResults []FormattedPrimaryStorage
		for _, ps := range primaryStorages {
			formatted := FormattedPrimaryStorage{
				Name:                           ps.Name,
				UUID:                           ps.UUID,
				Description:                    ps.Description,
				ZoneUuid:                       ps.ZoneUuid,
				Url:                            ps.Url,
				TotalCapacity:                  utils.FormatDiskSize(inventoryNumber(ps.TotalCapacityBytes)),
				TotalCapacityBytes:             ps.TotalCapacityBytes,
				AvailableCapacity:              utils.FormatDiskSize(inventoryNumber(ps.AvailableCapacityBytes)),
				AvailableCapacityBytes:         ps.AvailableCapacityBytes,
				TotalPhysicalCapacity:          utils.FormatDiskSize(inventoryNumber(ps.TotalPhysicalCapacityBytes)),
				TotalPhysicalCapacityBytes:     ps.TotalPhysicalCapacityBytes,
				AvailablePhysicalCapacity:      utils.FormatDiskSize(inventoryNumber(ps.AvailablePhysicalCapacityBytes)),
				AvailablePhysicalCapacityBytes: ps.AvailablePhysicalCapacityBytes,
				SystemUsedCapacity:             utils.FormatDiskSize(inventoryNumber(ps.SystemUsedCapacityBytes)),
				SystemUsedCapacityBytes:        ps.SystemUsedCapacityBytes,
				Type:                           ps.Type,
				State:                          ps.State,
				Status:                         ps.Status,
				AttachedClusterUuids:           strings.Join(ps.AttachedClusterUuids, ","),
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
	GetCmd.AddCommand(primaryStorageCmd)

	common.AddQueryFlags(primaryStorageCmd)

}
