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
	Name                 string `json:"name" yaml:"name" header:"NAME"`
	UUID                 string `json:"uuid" yaml:"uuid" header:"UUID"`
	ManagementIp         string `json:"managementIp" yaml:"managementIp" header:"MANAGEMENT IP"`
	HypervisorType       string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	State                string `json:"state" yaml:"state" header:"STATE"`
	Status               string `json:"status" yaml:"status" header:"STATUS"`
	ClusterUuid          string `json:"clusterUuid" yaml:"clusterUuid" header:"CLUSTER UUID"`
	ZoneUuid             string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	TotalCpu             string `json:"totalCpu" yaml:"totalCpu" header:"TOTAL CPU"`
	TotalCpuHz           *int64 `json:"totalCpuHz" yaml:"totalCpuHz" header:"TOTAL CPU HZ" table:"opt-in"`
	AvailableCpu         string `json:"availableCpu" yaml:"availableCpu" header:"AVAILABLE CPU"`
	AvailableCpuHz       *int64 `json:"availableCpuHz" yaml:"availableCpuHz" header:"AVAILABLE CPU HZ" table:"opt-in"`
	TotalMemory          string `json:"totalMemory" yaml:"totalMemory" header:"TOTAL MEMORY"`
	TotalMemoryBytes     *int64 `json:"totalMemoryBytes" yaml:"totalMemoryBytes" header:"TOTAL MEMORY BYTES" table:"opt-in"`
	AvailableMemory      string `json:"availableMemory" yaml:"availableMemory" header:"AVAILABLE MEMORY"`
	AvailableMemoryBytes *int64 `json:"availableMemoryBytes" yaml:"availableMemoryBytes" header:"AVAILABLE MEMORY BYTES" table:"opt-in"`
	CpuSockets           int    `json:"cpuSockets" yaml:"cpuSockets" header:"CPU SOCKETS"`
	CpuSocketCount       *int   `json:"cpuSocketCount" yaml:"cpuSocketCount" header:"CPU SOCKET COUNT" table:"opt-in"`
	CpuNum               int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	CpuCount             *int   `json:"cpuCount" yaml:"cpuCount" header:"CPU COUNT" table:"opt-in"`
	Architecture         string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	Description          string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

type hostQueryInventory struct {
	view.HostInventoryView
	TotalCpuHz           *int64 `json:"totalCpuCapacity"`
	AvailableCpuHz       *int64 `json:"availableCpuCapacity"`
	TotalMemoryBytes     *int64 `json:"totalMemoryCapacity"`
	AvailableMemoryBytes *int64 `json:"availableMemoryCapacity"`
	CpuCount             *int   `json:"cpuNum"`
	CpuSocketCount       *int   `json:"cpuSockets"`
}

var hostsCmd = &cobra.Command{
	Use:   "hosts [name]",
	Short: "List physical hosts",
	Long: `List all physical hosts in the ZStack cloud platform.
Structured output includes totalCpuHz, availableCpuHz, totalMemoryBytes, availableMemoryBytes, cpuCount, cpuSocketCount, or null if not reported.`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")

		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[hostQueryInventory](cobraCmd, queryParam, "v1/hosts")
		hosts := result.Items
		if err != nil {
			return fmt.Errorf("Error querying hosts: %w", err)
		}

		var formattedResults []FormattedHost
		for _, host := range hosts {
			formatted := FormattedHost{
				Name:                 host.Name,
				UUID:                 host.UUID,
				ManagementIp:         host.ManagementIp,
				HypervisorType:       host.HypervisorType,
				State:                host.State,
				Status:               host.Status,
				ClusterUuid:          host.ClusterUuid,
				ZoneUuid:             host.ZoneUuid,
				TotalCpu:             utils.FormatCpuCapacity(inventoryNumber(host.TotalCpuHz)),
				TotalCpuHz:           host.TotalCpuHz,
				AvailableCpu:         utils.FormatCpuCapacity(inventoryNumber(host.AvailableCpuHz)),
				AvailableCpuHz:       host.AvailableCpuHz,
				TotalMemory:          utils.FormatMemorySize(inventoryNumber(host.TotalMemoryBytes)),
				TotalMemoryBytes:     host.TotalMemoryBytes,
				AvailableMemory:      utils.FormatMemorySize(inventoryNumber(host.AvailableMemoryBytes)),
				AvailableMemoryBytes: host.AvailableMemoryBytes,
				CpuSockets:           inventoryNumber(host.CpuSocketCount),
				CpuSocketCount:       host.CpuSocketCount,
				CpuNum:               inventoryNumber(host.CpuCount),
				CpuCount:             host.CpuCount,
				Architecture:         host.Architecture,
				Description:          host.Description,
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
