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

package expunge

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type expungeVolumeRequest struct {
	Volume string `json:"volume" yaml:"volume"`
}

var ExpungeVolumesCmd = &cobra.Command{
	Use:     "volume <name-or-uuid>",
	Aliases: []string{"volumes"},
	Short:   "Expunge a deleted data volume",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(expungeVolumeRequest{Volume: args[0]}, format)
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
		if !yes && !confirmExpungeVolume(volume.Name, volume.UUID) {
			fmt.Println("Aborted by user.")
			return nil
		}

		if err := cli.ExpungeDataVolume(volume.UUID); err != nil {
			return err
		}
		fmt.Printf("Expunged volume: %s (%s)\n", volume.Name, volume.UUID)
		return nil
	},
}

func init() {
	ExpungeCmd.AddCommand(ExpungeVolumesCmd)
	ExpungeVolumesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
}

func confirmExpungeVolume(name string, uuid string) bool {
	fmt.Println("The following volume will be expunged:")
	fmt.Printf("  - %s (%s)\n", name, uuid)
	fmt.Print("Are you sure you want to expunge this volume? Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes" || input == "y"
}
