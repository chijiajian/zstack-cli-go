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
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

type FormattedVmNic struct {
	UUID           string `json:"uuid" yaml:"uuid" header:"UUID"`
	VMInstanceUUID string `json:"vmInstanceUuid" yaml:"vmInstanceUuid" header:"VM UUID"`
	L3NetworkUUID  string `json:"l3NetworkUuid" yaml:"l3NetworkUuid" header:"L3 NETWORK"`
	IP             string `json:"ip" yaml:"ip" header:"IP"`
	Mac            string `json:"mac" yaml:"mac" header:"MAC"`
	Netmask        string `json:"netmask" yaml:"netmask" header:"NETMASK"`
	Gateway        string `json:"gateway" yaml:"gateway" header:"GATEWAY"`
	IPVersion      int    `json:"ipVersion" yaml:"ipVersion" header:"IP VER"`
	DeviceID       int    `json:"deviceId" yaml:"deviceId" header:"DEVICE"`
	Type           string `json:"type" yaml:"type" header:"TYPE"`
	DriverType     string `json:"driverType" yaml:"driverType" header:"DRIVER"`
}

var nicsCmd = &cobra.Command{
	Use:     "nics",
	Aliases: []string{"nic", "vm-nics"},
	Short:   "List VM network interfaces",
	Long:    `List all VM network interfaces (NICs) in the ZStack cloud platform.`,
	Args:    cobra.MaximumNArgs(0),
	Run: func(cobraCmd *cobra.Command, args []string) {

		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		usePagination, _ := cobraCmd.Flags().GetBool("pagination")
		var nics []view.VmNicInventoryView
		var total int

		if usePagination {
			nics, total, err = zsClient.PageVmNic(*queryParam)
			if err != nil {
				fmt.Printf("Error querying VM NICs: %s\n", err)
				return
			}
			fmt.Printf("Total: %d\n", total)
		} else {
			nics, err = zsClient.QueryVmNic(*queryParam)
			if err != nil {
				fmt.Printf("Error querying VM NICs: %s\n", err)
				return
			}
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedVmNic
		for _, nic := range nics {
			// Get IP addresses from the nic
			var ips []string
			for _, usedIp := range nic.UsedIps {
				ips = append(ips, usedIp.Ip)
			}
			ipStr := strings.Join(ips, ", ")
			if ipStr == "" {
				ipStr = nic.IP
			}

			formatted := FormattedVmNic{
				UUID:           nic.UUID,
				VMInstanceUUID: nic.VMInstanceUUID,
				L3NetworkUUID:  nic.L3NetworkUUID,
				IP:             ipStr,
				Mac:            nic.Mac,
				Netmask:        nic.Netmask,
				Gateway:        nic.Gateway,
				IPVersion:      nic.IpVersion,
				DeviceID:       nic.DeviceID,
				Type:           nic.Type,
				DriverType:     nic.DriverType,
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
	GetCmd.AddCommand(nicsCmd)
	common.AddQueryFlags(nicsCmd)
	nicsCmd.Flags().Bool("pagination", false, "Use pagination when querying NICs")
}
