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
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"

	"github.com/spf13/cobra"
)

type FormattedVmInstance struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`

	ZoneUUID             string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	ClusterUUID          string `json:"clusterUuid" yaml:"clusterUuid" header:"CLUSTER UUID"`
	ImageUUID            string `json:"imageUuid" yaml:"imageUuid" header:"IMAGE UUID"`
	HostUUID             string `json:"hostUuid" yaml:"hostUuid" header:"HOST UUID"`
	LastHostUUID         string `json:"lastHostUuid" yaml:"lastHostUuid" header:"LAST HOST UUID"`
	InstanceOfferingUUID string `json:"instanceOfferingUuid" yaml:"instanceOfferingUuid" header:"INSTANCE OFFERING UUID"`
	RootVolumeUUID       string `json:"rootVolumeUuid" yaml:"rootVolumeUuid" header:"ROOT VOLUME UUID"`
	Platform             string `json:"platform" yaml:"platform" header:"PLATFORM"`
	Architecture         string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	GuestOsType          string `json:"guestOsType" yaml:"guestOsType" header:"GUEST OS TYPE"`
	DefaultL3NetworkUUID string `json:"defaultL3NetworkUuid" yaml:"defaultL3NetworkUuid" header:"DEFAULT L3 NETWORK UUID"`
	Type                 string `json:"type" yaml:"type" header:"TYPE"`
	HypervisorType       string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	MemorySize           string `json:"memorySize" yaml:"memorySize" header:"MEMORY SIZE"`
	CPUNum               int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	CPUSpeed             int64  `json:"cpuSpeed" yaml:"cpuSpeed" header:"CPU SPEED"`
	AllocatorStrategy    string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State                string `json:"state" yaml:"state" header:"STATE"`
	IPs                  string `json:"ips" yaml:"ips" header:"IPS"`
	Volumes              string `json:"volumes" yaml:"volumes" header:"VOLUMES"`
}

var vmInstancesCmd = &cobra.Command{
	Use:   "instances [name]",
	Short: "List VM instances",
	Long:  `List all VM instances in the ZStack cloud platform.`,
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

		err = common.ProcessBasicContextFlags(cobraCmd, queryParam)
		if err != nil {
			fmt.Printf("Error processing context filters: %s\n", err)
			return
		}

		usePagination, _ := cobraCmd.Flags().GetBool("pagination")
		var vmInstances []view.VmInstanceInventoryView
		var total int

		if usePagination {

			vmInstances, total, err = zsClient.PageVmInstance(*queryParam)
			if err != nil {
				fmt.Printf("Error querying VM instances: %s\n", err)
				return
			}
			fmt.Printf("Total: %d\n", total)
		} else {

			vmInstances, err = zsClient.QueryVmInstance(*queryParam)
			if err != nil {
				fmt.Printf("Error querying VM instances: %s\n", err)
				return
			}
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedVmInstance
		for _, vm := range vmInstances {

			var ips []string
			for _, nic := range vm.VMNics {
				if nic.IP != "" {
					ips = append(ips, nic.IP)
				}
			}

			var volumes []string
			for _, vol := range vm.AllVolumes {
				volumes = append(volumes, vol.Name)
			}

			formatted := FormattedVmInstance{
				Name:        vm.Name,
				UUID:        vm.UUID,
				Description: vm.Description,

				ZoneUUID:             vm.ZoneUUID,
				ClusterUUID:          vm.ClusterUUID,
				ImageUUID:            vm.ImageUUID,
				HostUUID:             vm.HostUUID,
				LastHostUUID:         vm.LastHostUUID,
				InstanceOfferingUUID: vm.InstanceOfferingUUID,
				RootVolumeUUID:       vm.RootVolumeUUID,
				Platform:             vm.Platform,
				Architecture:         vm.Architecture,
				GuestOsType:          vm.GuestOsType,
				DefaultL3NetworkUUID: vm.DefaultL3NetworkUUID,
				Type:                 vm.Type,
				HypervisorType:       vm.HypervisorType,
				MemorySize:           utils.FormatMemorySize(vm.MemorySize),
				CPUNum:               vm.CPUNum,
				CPUSpeed:             vm.CPUSpeed,
				AllocatorStrategy:    vm.AllocatorStrategy,
				State:                vm.State,
				IPs:                  strings.Join(ips, ", "),
				Volumes:              strings.Join(volumes, ", "),
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
	GetCmd.AddCommand(vmInstancesCmd)
	common.AddQueryFlags(vmInstancesCmd)
	vmInstancesCmd.Flags().Bool("pagination", false, "Use pagination when querying VM instances")
	vmInstancesCmd.Flags().StringP("zone", "z", "", "Filter by zone name or UUID")
	vmInstancesCmd.Flags().StringP("cluster", "c", "", "Filter by cluster name or UUID")
	vmInstancesCmd.Flags().StringP("host", "H", "", "Filter by host name or UUID")
}
