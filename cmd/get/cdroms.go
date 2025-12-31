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

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

type FormattedCdRom struct {
	UUID           string  `json:"uuid" yaml:"uuid" header:"UUID"`
	VMInstanceUUID string  `json:"vmInstanceUuid" yaml:"vmInstanceUuid" header:"VM UUID"`
	DeviceID       float64 `json:"deviceId" yaml:"deviceId" header:"DEVICE"`
	IsoUUID        string  `json:"isoUuid" yaml:"isoUuid" header:"ISO UUID"`
	IsoInstallPath string  `json:"isoInstallPath" yaml:"isoInstallPath" header:"ISO PATH"`
	Name           string  `json:"name" yaml:"name" header:"NAME"`
	Description    string  `json:"description" yaml:"description" header:"DESCRIPTION"`
	CreateDate     string  `json:"createDate" yaml:"createDate" header:"CREATED"`
}

var cdromsCmd = &cobra.Command{
	Use:     "cdroms",
	Aliases: []string{"cdrom", "cd-roms"},
	Short:   "List VM CD-ROM devices",
	Long:    `List all VM CD-ROM devices in the ZStack cloud platform.`,
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
		var cdroms []view.VMCDRomView
		var total int

		if usePagination {
			cdroms, total, err = zsClient.PageVmCdRom(*queryParam)
			if err != nil {
				fmt.Printf("Error querying CD-ROMs: %s\n", err)
				return
			}
			fmt.Printf("Total: %d\n", total)
		} else {
			cdroms, err = zsClient.QueryVmCdRom(*queryParam)
			if err != nil {
				fmt.Printf("Error querying CD-ROMs: %s\n", err)
				return
			}
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedCdRom
		for _, cdrom := range cdroms {
			formatted := FormattedCdRom{
				UUID:           cdrom.UUID,
				VMInstanceUUID: cdrom.VmInstanceUuid,
				DeviceID:       cdrom.DeviceId,
				IsoUUID:        cdrom.IsoUuid,
				IsoInstallPath: cdrom.IsoInstallPath,
				Name:           cdrom.Name,
				Description:    cdrom.Description,
				CreateDate:     cdrom.CreateDate.Format("2006-01-02 15:04:05"),
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
	GetCmd.AddCommand(cdromsCmd)
	common.AddQueryFlags(cdromsCmd)
	cdromsCmd.Flags().Bool("pagination", false, "Use pagination when querying CD-ROMs")
}
