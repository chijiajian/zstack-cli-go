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

type FormattedVirtualRouter struct {
	Name                      string `json:"name" yaml:"name" header:"NAME"`
	UUID                      string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description               string `json:"description" yaml:"description" header:"DESCRIPTION"`
	ApplianceVmType           string `json:"applianceVmType" yaml:"applianceVmType" header:"APPLIANCE VM TYPE"`
	ManagementNetworkUuid     string `json:"managementNetworkUuid" yaml:"managementNetworkUuid" header:"MANAGEMENT NETWORK UUID"`
	DefaultRouteL3NetworkUuid string `json:"defaultRouteL3NetworkUuid" yaml:"defaultRouteL3NetworkUuid" header:"DEFAULT ROUTE L3 NETWORK UUID"`
	Status                    string `json:"status" yaml:"status" header:"STATUS"`
	AgentPort                 int    `json:"agentPort" yaml:"agentPort" header:"AGENT PORT"`
	ZoneUuid                  string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	ClusterUUID               string `json:"clusterUuid" yaml:"clusterUuid" header:"CLUSTER UUID"`
	ImageUUID                 string `json:"imageUuid" yaml:"imageUuid" header:"IMAGE UUID"`
	HostUuid                  string `json:"hostUuid" yaml:"hostUuid" header:"HOST UUID"`
	LastHostUUID              string `json:"lastHostUuid" yaml:"lastHostUuid" header:"LAST HOST UUID"`
	InstanceOfferingUUID      string `json:"instanceOfferingUuid" yaml:"instanceOfferingUuid" header:"INSTANCE OFFERING UUID"`
	RootVolumeUuid            string `json:"rootVolumeUuid" yaml:"rootVolumeUuid" header:"ROOT VOLUME UUID"`
	Platform                  string `json:"platform" yaml:"platform" header:"PLATFORM"`
	DefaultL3NetworkUuid      string `json:"defaultL3NetworkUuid" yaml:"defaultL3NetworkUuid" header:"DEFAULT L3 NETWORK UUID"`
	Type                      string `json:"type" yaml:"type" header:"TYPE"`
	HypervisorType            string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	MemorySize                string `json:"memorySize" yaml:"memorySize" header:"MEMORY SIZE"`
	CPUNum                    int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	CPUSpeed                  int64  `json:"cpuSpeed" yaml:"cpuSpeed" header:"CPU SPEED"`
	AllocatorStrategy         string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State                     string `json:"state" yaml:"state" header:"STATE"`
	HaStatus                  string `json:"haStatus" yaml:"haStatus" header:"HA STATUS"`
	Architecture              string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	IPs                       string `json:"ips" yaml:"ips" header:"IPS"`
}

var virtualRoutersCmd = &cobra.Command{
	Use:   "virtual-routers [name]",
	Short: "List virtual routers",
	Long:  `List all virtual routers in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.ApplianceVmInventoryView](cobraCmd, queryParam, "v1/vm-instances/appliances/virtual-routers")
		virtualRouters := result.Items
		if err != nil {
			return fmt.Errorf("Error querying virtual routers: %w", err)
		}

		var formattedResults []FormattedVirtualRouter
		for _, vr := range virtualRouters {

			var ips []string
			for _, nic := range vr.VmNics {
				if nic.Ip != "" {
					ips = append(ips, nic.Ip)
				}
			}

			formatted := FormattedVirtualRouter{
				Name:        vr.Name,
				UUID:        vr.UUID,
				Description: vr.Description,

				ApplianceVmType:           vr.ApplianceVmType,
				ManagementNetworkUuid:     vr.ManagementNetworkUuid,
				DefaultRouteL3NetworkUuid: vr.DefaultRouteL3NetworkUuid,
				Status:                    vr.Status,
				AgentPort:                 vr.AgentPort,
				ZoneUuid:                  vr.ZoneUuid,
				ClusterUUID:               vr.ClusterUuid,
				ImageUUID:                 vr.ImageUuid,
				HostUuid:                  vr.HostUuid,
				LastHostUUID:              vr.LastHostUuid,
				InstanceOfferingUUID:      vr.InstanceOfferingUuid,
				RootVolumeUuid:            vr.RootVolumeUuid,
				Platform:                  vr.Platform,
				DefaultL3NetworkUuid:      vr.DefaultL3NetworkUuid,
				Type:                      vr.Type,
				HypervisorType:            vr.HypervisorType,
				MemorySize:                utils.FormatMemorySize(vr.MemorySize),
				CPUNum:                    vr.CpuNum,
				CPUSpeed:                  vr.CpuSpeed,
				AllocatorStrategy:         vr.AllocatorStrategy,
				State:                     vr.State,
				HaStatus:                  vr.HaStatus,
				Architecture:              vr.Architecture,
				IPs:                       strings.Join(ips, ", "),
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
	GetCmd.AddCommand(virtualRoutersCmd)
	common.AddQueryFlags(virtualRoutersCmd)
}
