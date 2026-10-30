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

package resources

import (
	"fmt"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type snapshotCreateRequest struct {
	Volume string                          `json:"volume" yaml:"volume"`
	Params param.CreateVolumeSnapshotParam `json:"params" yaml:"params"`
}

type snapshotRevertRequest struct {
	Snapshot string `json:"snapshot" yaml:"snapshot"`
}

var SnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage volume snapshots",
	Long:  "Create and revert ZStack volume snapshots.",
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create <volume-name-or-uuid>",
	Short: "Create a volume snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotName, _ := cmd.Flags().GetString("name")
		if snapshotName == "" {
			return fmt.Errorf("required flag --name not set")
		}

		description, _ := cmd.Flags().GetString("description")
		resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
		tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
		systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
		userTags, _ := cmd.Flags().GetStringSlice("user-tag")

		snapshotParam := param.CreateVolumeSnapshotParam{
			BaseParam: param.BaseParam{
				SystemTags: systemTags,
				UserTags:   userTags,
			},
			Params: param.CreateVolumeSnapshotParamDetail{
				Name:         snapshotName,
				Description:  stringPtrIfNotEmpty(description),
				ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
				TagUuids:     tagUUIDs,
			},
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return printSnapshotDryRun(cmd, snapshotCreateRequest{Volume: args[0], Params: snapshotParam})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		volume, err := zstackclient.ResolveOneResource(cli, "volume", args[0])
		if err != nil {
			return err
		}

		result, err := cli.CreateVolumeSnapshot(volume.UUID, snapshotParam)
		if err != nil {
			return err
		}

		return printSnapshotResult(cmd, fmt.Sprintf("snapshot/%s created", result.Name), result)
	},
}

var snapshotRevertCmd = &cobra.Command{
	Use:   "revert <snapshot-name-or-uuid>",
	Short: "Revert a volume from a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return printSnapshotDryRun(cmd, snapshotRevertRequest{Snapshot: args[0]})
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		snapshot, err := zstackclient.ResolveOneResource(cli, "snapshot", args[0])
		if err != nil {
			return err
		}

		confirmed, err := common.Confirm(cmd, fmt.Sprintf("Revert volume from snapshot %s (%s)? Type 'yes' to confirm: ", snapshot.Name, snapshot.UUID))
		if err != nil {
			return err
		}
		if !confirmed {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Aborted by user.")
			return err
		}
		result, err := cli.RevertVolumeFromSnapshot(snapshot.UUID, param.RevertVolumeFromSnapshotParam{})
		if err != nil {
			return err
		}

		return printSnapshotResult(cmd, fmt.Sprintf("snapshot/%s reverted", snapshot.Name), result)
	},
}

func init() {
	snapshotRevertCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	SnapshotCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	SnapshotCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	SnapshotCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")

	snapshotCreateCmd.Flags().String("name", "", "Snapshot name")
	snapshotCreateCmd.Flags().StringP("description", "d", "", "Detailed description of the snapshot")
	snapshotCreateCmd.Flags().String("resource-uuid", "", "Resource UUID")
	snapshotCreateCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	snapshotCreateCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	snapshotCreateCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	SnapshotCmd.AddCommand(snapshotCreateCmd, snapshotRevertCmd)
}

func printSnapshotDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printSnapshotResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}

func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
