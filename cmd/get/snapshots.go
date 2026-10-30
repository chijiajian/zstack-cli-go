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

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type FormattedVolumeSnapshot struct {
	Name             string `json:"name" yaml:"name" header:"NAME"`
	UUID             string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description      string `json:"description" yaml:"description" header:"DESCRIPTION"`
	Type             string `json:"type" yaml:"type" header:"TYPE"`
	VolumeUUID       string `json:"volumeUuid" yaml:"volumeUuid" header:"VOLUME UUID"`
	TreeUUID         string `json:"treeUuid" yaml:"treeUuid" header:"TREE UUID"`
	ParentUUID       string `json:"parentUuid" yaml:"parentUuid" header:"PARENT UUID"`
	PrimaryStorageID string `json:"primaryStorageUuid" yaml:"primaryStorageUuid" header:"PRIMARY STORAGE"`
	Size             string `json:"size" yaml:"size" header:"SIZE"`
	State            string `json:"state" yaml:"state" header:"STATE"`
	Status           string `json:"status" yaml:"status" header:"STATUS"`
	Latest           bool   `json:"latest" yaml:"latest" header:"LATEST"`
	CreateDate       string `json:"createDate" yaml:"createDate" header:"CREATED"`
}

var snapshotsCmd = &cobra.Command{
	Use:     "snapshots [name]",
	Aliases: []string{"snapshot", "snap"},
	Short:   "List volume snapshots",
	Long:    `List all volume snapshots in the ZStack cloud platform.`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.VolumeSnapshotInventoryView](cobraCmd, queryParam, "v1/volume-snapshots")
		snapshots := result.Items
		if err != nil {
			return err
		}

		var formattedResults []FormattedVolumeSnapshot
		for _, snapshot := range snapshots {
			formatted := FormattedVolumeSnapshot{
				Name:             snapshot.Name,
				UUID:             snapshot.UUID,
				Description:      snapshot.Description,
				Type:             snapshot.Type,
				VolumeUUID:       snapshot.VolumeUuid,
				TreeUUID:         snapshot.TreeUuid,
				ParentUUID:       snapshot.ParentUuid,
				PrimaryStorageID: snapshot.PrimaryStorageUuid,
				Size:             utils.FormatDiskSize(snapshot.Size),
				State:            snapshot.State,
				Status:           snapshot.Status,
				Latest:           snapshot.Latest,
				CreateDate:       snapshot.CreateDate.Format("2006-01-02 15:04:05"),
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
	GetCmd.AddCommand(snapshotsCmd)
	common.AddQueryFlags(snapshotsCmd)
	snapshotsCmd.Flags().Bool("pagination", false, "Use pagination when querying snapshots")
}
