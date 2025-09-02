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
	"strconv"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
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

		hosts, err := zsClient.QueryHost(*queryParam)
		if err != nil {
			fmt.Printf("Error querying hosts: %s\n", err)
			return
		}

		if len(hosts) == 0 {
			fmt.Println("No hosts found.")
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

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

		err = utils.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
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
