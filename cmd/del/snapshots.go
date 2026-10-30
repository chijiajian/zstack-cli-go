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

package del

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type deleteSnapshotRequest struct {
	Snapshot string `json:"snapshot" yaml:"snapshot"`
}

var DeleteSnapshotsCmd = &cobra.Command{
	Use:     "snapshot <name-or-uuid>",
	Aliases: []string{"snapshots"},
	Short:   "Delete a volume snapshot",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(deleteSnapshotRequest{Snapshot: args[0]}, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		snapshot, err := zstackclient.ResolveOneResource(cli, "snapshot", args[0])
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes && !confirmSnapshotDelete(snapshot.Name, snapshot.UUID) {
			fmt.Println("Aborted by user.")
			return nil
		}

		if err := cli.DeleteVolumeSnapshot(snapshot.UUID, param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("Deleted snapshot: %s (%s)\n", snapshot.Name, snapshot.UUID)
		return nil
	},
}

func init() {
	DeleteCmd.AddCommand(DeleteSnapshotsCmd)
	DeleteSnapshotsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
}

func confirmSnapshotDelete(name string, uuid string) bool {
	fmt.Println("The following snapshot will be deleted:")
	fmt.Printf("  - %s (%s)\n", name, uuid)
	fmt.Print("Are you sure you want to delete this snapshot? Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes" || input == "y"
}
