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

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
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
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.VmNicInventoryView](cobraCmd, queryParam, "v1/vm-instances/nics")
		nics := result.Items
		if err != nil {
			return err
		}

		var formattedResults []FormattedVmNic
		for _, nic := range nics {
			// Get IP addresses from the nic
			var ips []string
			for _, usedIp := range nic.UsedIps {
				ips = append(ips, usedIp.Ip)
			}
			ipStr := strings.Join(ips, ", ")
			if ipStr == "" {
				ipStr = nic.Ip
			}

			formatted := FormattedVmNic{
				UUID:           nic.UUID,
				VMInstanceUUID: nic.VmInstanceUuid,
				L3NetworkUUID:  nic.L3NetworkUuid,
				IP:             ipStr,
				Mac:            nic.Mac,
				Netmask:        nic.Netmask,
				Gateway:        nic.Gateway,
				IPVersion:      nic.IpVersion,
				DeviceID:       nic.DeviceId,
				Type:           nic.Type,
				DriverType:     nic.DriverType,
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
	GetCmd.AddCommand(nicsCmd)
	common.AddQueryFlags(nicsCmd)
	nicsCmd.Flags().Bool("pagination", false, "Use pagination when querying NICs")
}
