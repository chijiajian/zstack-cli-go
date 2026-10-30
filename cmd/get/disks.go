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
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type FormattedVolume struct {
	Name               string `json:"name" yaml:"name" header:"NAME"`
	UUID               string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description        string `json:"description" yaml:"description" header:"DESCRIPTION"`
	PrimaryStorageUUID string `json:"primaryStorageUuid" yaml:"primaryStorageUuid" header:"PRIMARY STORAGE UUID"`
	VMInstanceUUID     string `json:"vmInstanceUuid" yaml:"vmInstanceUuid" header:"VM INSTANCE UUID"`
	LastVmInstanceUuid string `json:"lastVmInstanceUuid" yaml:"lastVmInstanceUuid" header:"LAST VM INSTANCE UUID"`
	DiskOfferingUUID   string `json:"diskOfferingUuid" yaml:"diskOfferingUuid" header:"DISK OFFERING UUID"`
	RootImageUUID      string `json:"rootImageUuid" yaml:"rootImageUuid" header:"ROOT IMAGE UUID"`
	InstallPath        string `json:"installPath" yaml:"installPath" header:"INSTALL PATH"`
	Type               string `json:"type" yaml:"type" header:"TYPE"`
	Format             string `json:"format" yaml:"format" header:"FORMAT"`
	Size               string `json:"size" yaml:"size" header:"SIZE"`
	ActualSize         string `json:"actualSize" yaml:"actualSize" header:"ACTUAL SIZE"`
	DeviceID           int    `json:"deviceId" yaml:"deviceId" header:"DEVICE ID"`
	State              string `json:"state" yaml:"state" header:"STATE"`
	Status             string `json:"status" yaml:"status" header:"STATUS"`
	IsShareable        bool   `json:"isShareable" yaml:"isShareable" header:"IS SHAREABLE"`
	LastDetachDate     string `json:"lastDetachDate" yaml:"lastDetachDate" header:"LAST DETACH DATE"`
	Attached           string `json:"attached" yaml:"attached" header:"ATTACHED"`
}

var volumesCmd = &cobra.Command{
	Use:     "disks [name]",
	Aliases: []string{"volumes"},
	Short:   "List disks",
	Long:    `List all volumes (disks) in the ZStack cloud platform.`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.VolumeInventoryView](cobraCmd, queryParam, "v1/volumes")
		volumes := result.Items
		if err != nil {
			return err
		}

		var formattedResults []FormattedVolume
		for _, volume := range volumes {
			attached := "No"
			if volume.VmInstanceUuid != "" {
				attached = "Yes"
			}

			lastDetachDate := ""
			var zeroTime time.Time
			if volume.LastDetachDate != zeroTime {
				lastDetachDate = volume.LastDetachDate.Format("2006-01-02 15:04:05")
			}

			formatted := FormattedVolume{
				Name:               volume.Name,
				UUID:               volume.UUID,
				Description:        volume.Description,
				PrimaryStorageUUID: volume.PrimaryStorageUuid,
				VMInstanceUUID:     volume.VmInstanceUuid,
				LastVmInstanceUuid: volume.LastVmInstanceUuid,
				DiskOfferingUUID:   volume.DiskOfferingUuid,
				RootImageUUID:      volume.RootImageUuid,
				InstallPath:        volume.InstallPath,
				Type:               volume.Type,
				Format:             volume.Format,
				Size:               utils.FormatDiskSize(int64(volume.Size)),
				ActualSize:         utils.FormatDiskSize(int64(volume.ActualSize)),
				DeviceID:           volume.DeviceId,
				State:              volume.State,
				Status:             volume.Status,
				IsShareable:        volume.IsShareable,
				LastDetachDate:     lastDetachDate,
				Attached:           attached,
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
	GetCmd.AddCommand(volumesCmd)
	common.AddQueryFlags(volumesCmd)
	volumesCmd.Flags().Bool("pagination", false, "Use pagination when querying volumes")
}
