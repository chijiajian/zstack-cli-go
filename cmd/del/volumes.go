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

type deleteVolumeRequest struct {
	Volume string `json:"volume" yaml:"volume"`
}

var DeleteVolumesCmd = &cobra.Command{
	Use:     "volume <name-or-uuid>",
	Aliases: []string{"volumes"},
	Short:   "Delete a data volume",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(deleteVolumeRequest{Volume: args[0]}, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		volume, err := zstackclient.ResolveOneResource(cli, "volume", args[0])
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes && !confirmVolumeAction("delete", volume.Name, volume.UUID) {
			fmt.Println("Aborted by user.")
			return nil
		}

		if err := cli.DeleteDataVolume(volume.UUID, param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("Deleted volume: %s (%s)\n", volume.Name, volume.UUID)
		return nil
	},
}

func init() {
	DeleteCmd.AddCommand(DeleteVolumesCmd)
	DeleteVolumesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
}

func confirmVolumeAction(action string, name string, uuid string) bool {
	fmt.Printf("The following volume will be %sd:\n", action)
	fmt.Printf("  - %s (%s)\n", name, uuid)
	fmt.Printf("Are you sure you want to %s this volume? Type 'yes' to confirm: ", action)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes" || input == "y"
}
