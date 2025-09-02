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

type FormattedPrimaryStorage struct {
	Name                      string `json:"name" yaml:"name" header:"NAME"`
	UUID                      string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description               string `json:"description" yaml:"description" header:"DESCRIPTION"`
	ZoneUuid                  string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	Url                       string `json:"url" yaml:"url" header:"URL"`
	TotalCapacity             string `json:"totalCapacity" yaml:"totalCapacity" header:"TOTAL CAPACITY"`
	AvailableCapacity         string `json:"availableCapacity" yaml:"availableCapacity" header:"AVAILABLE CAPACITY"`
	TotalPhysicalCapacity     string `json:"totalPhysicalCapacity" yaml:"totalPhysicalCapacity" header:"TOTAL PHYSICAL CAPACITY"`
	AvailablePhysicalCapacity string `json:"availablePhysicalCapacity" yaml:"availablePhysicalCapacity" header:"AVAILABLE PHYSICAL CAPACITY"`
	SystemUsedCapacity        string `json:"systemUsedCapacity" yaml:"systemUsedCapacity" header:"SYSTEM USED CAPACITY"`
	Type                      string `json:"type" yaml:"type" header:"TYPE"`
	State                     string `json:"state" yaml:"state" header:"STATE"`
	Status                    string `json:"status" yaml:"status" header:"STATUS"`
	AttachedClusterUuids      string `json:"attachedClusterUuids" yaml:"attachedClusterUuids" header:"ATTACHED CLUSTER UUIDS"`

	MonCount  int `json:"monCount" yaml:"monCount" header:"MON COUNT"`
	PoolCount int `json:"poolCount" yaml:"poolCount" header:"POOL COUNT"`
}

var primaryStorageCmd = &cobra.Command{
	Use:   "primary-storages [name]",
	Short: "List primary storages",
	Long:  `List all primary storages in the ZStack cloud platform.`,
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

		primaryStorages, err := zsClient.QueryPrimaryStorage(*queryParam)
		if err != nil {
			fmt.Printf("Error querying primary storages: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedPrimaryStorage
		for _, ps := range primaryStorages {
			formatted := FormattedPrimaryStorage{
				Name:                      ps.Name,
				UUID:                      ps.UUID,
				Description:               ps.Description,
				ZoneUuid:                  ps.ZoneUuid,
				Url:                       ps.Url,
				TotalCapacity:             utils.FormatDiskSize(ps.TotalCapacity),
				AvailableCapacity:         utils.FormatDiskSize(ps.AvailableCapacity),
				TotalPhysicalCapacity:     utils.FormatDiskSize(ps.TotalPhysicalCapacity),
				AvailablePhysicalCapacity: utils.FormatDiskSize(ps.AvailablePhysicalCapacity),
				SystemUsedCapacity:        utils.FormatDiskSize(ps.SystemUsedCapacity),
				Type:                      ps.Type,
				State:                     ps.State,
				Status:                    ps.Status,
				AttachedClusterUuids:      strings.Join(ps.AttachedClusterUuids, ","),

				MonCount:  len(ps.Mons),
				PoolCount: len(ps.Pools),
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
	GetCmd.AddCommand(primaryStorageCmd)

	common.AddQueryFlags(primaryStorageCmd)

}
