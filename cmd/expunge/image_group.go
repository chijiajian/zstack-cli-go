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

type expungeImageGroupRequest struct {
	ImageGroup string `json:"imageGroup" yaml:"imageGroup"`
}

var ExpungeImageGroupsCmd = &cobra.Command{
	Use:     "image-group <name-or-uuid>",
	Aliases: []string{"image-groups"},
	Short:   "Expunge an image group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(expungeImageGroupRequest{ImageGroup: args[0]}, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		group, err := zstackclient.ResolveOneResource(cli, "image-group", args[0])
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes && !confirmImageGroupExpunge(group.Name, group.UUID) {
			fmt.Println("Aborted by user.")
			return nil
		}

		if err := cli.ExpungeImageGroup(group.UUID); err != nil {
			return err
		}
		fmt.Printf("Expunged image-group: %s (%s)\n", group.Name, group.UUID)
		return nil
	},
}

func init() {
	ExpungeCmd.AddCommand(ExpungeImageGroupsCmd)
	ExpungeImageGroupsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
}

func confirmImageGroupExpunge(name string, uuid string) bool {
	fmt.Printf("The following image-group will be expunged:\n")
	fmt.Printf("  - %s (%s)\n", name, uuid)
	fmt.Printf("Are you sure you want to expunge this image-group? Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes" || input == "y"
}
