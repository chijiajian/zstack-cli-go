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
	"strconv"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type FormattedHost struct {
	Name            string `json:"name" yaml:"name" header:"NAME"`
	UUID            string `json:"uuid" yaml:"uuid" header:"UUID"`
	ManagementIp    string `json:"managementIp" yaml:"managementIp" header:"MANAGEMENT IP"`
	HypervisorType  string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	State           string `json:"state" yaml:"state" header:"STATE"`
	Status          string `json:"status" yaml:"status" header:"STATUS"`
	ClusterUuid     string `json:"clusterUuid" yaml:"clusterUuid" header:"CLUSTER UUID"`
	ZoneUuid        string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	TotalCpu        string `json:"totalCpu" yaml:"totalCpu" header:"TOTAL CPU"`
	AvailableCpu    string `json:"availableCpu" yaml:"availableCpu" header:"AVAILABLE CPU"`
	TotalMemory     string `json:"totalMemory" yaml:"totalMemory" header:"TOTAL MEMORY"`
	AvailableMemory string `json:"availableMemory" yaml:"availableMemory" header:"AVAILABLE MEMORY"`
	CpuSockets      int    `json:"cpuSockets" yaml:"cpuSockets" header:"CPU SOCKETS"`
	CpuNum          int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	Architecture    string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	Description     string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

var hostsCmd = &cobra.Command{
	Use:   "hosts [name]",
	Short: "List physical hosts",
	Long:  `List all physical hosts in the ZStack cloud platform.`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")

		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.HostInventoryView](cobraCmd, queryParam, "v1/hosts")
		hosts := result.Items
		if err != nil {
			return fmt.Errorf("Error querying hosts: %w", err)
		}

		var formattedResults []FormattedHost
		for _, host := range hosts {
			formatted := FormattedHost{
				Name:            host.Name,
				UUID:            host.UUID,
				ManagementIp:    host.ManagementIp,
				HypervisorType:  host.HypervisorType,
				State:           host.State,
				Status:          host.Status,
				ClusterUuid:     host.ClusterUuid,
				ZoneUuid:        host.ZoneUuid,
				TotalCpu:        utils.FormatCpuCapacity(host.TotalCpuCapacity),
				AvailableCpu:    utils.FormatCpuCapacity(host.AvailableCpuCapacity),
				TotalMemory:     utils.FormatMemorySize(host.TotalMemoryCapacity),
				AvailableMemory: utils.FormatMemorySize(host.AvailableMemoryCapacity),
				CpuSockets:      host.CpuSockets,
				CpuNum:          host.CpuNum,
				Architecture:    host.Architecture,
				Description:     host.Description,
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

func formatCpuCapacity(cpuHz int64) string {
	cpuGHz := float64(cpuHz) / 1000000000
	return strconv.FormatFloat(cpuGHz, 'f', 2, 64) + " GHz"
}

func init() {
	GetCmd.AddCommand(hostsCmd)
	common.AddQueryFlags(hostsCmd)
}
