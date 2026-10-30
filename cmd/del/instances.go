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
	"errors"
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"

	"github.com/spf13/cobra"
)

var DeleteInstancesCmd = &cobra.Command{
	Use:     "instance [name-or-uuid]",
	Aliases: []string{"instances"},
	Short:   "Delete one or many VM instances",
	Long: `Delete one or many ZStack VM instances.

Examples:
  # Delete a single VM instance by name or UUID
  zstack-cli delete instance my-vm

  # Delete multiple VM instances by name
  zstack-cli delete instance my-vm 

Note: This is a dangerous operation. Confirm interactively or use --yes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameOrUUID := args[0]
		return deleteVmInstance(cmd, nameOrUUID)
	},
}

func deleteVmInstance(cmd *cobra.Command, nameOrUUID string) error {
	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in, please run 'zstack-cli login' first")
	}

	vms, err := client.GetReadyVMsByNameOrUUID(cli, nameOrUUID)
	if err != nil {
		return fmt.Errorf("query VM instances failed: %s", err)
	}

	if len(vms) == 0 {
		return fmt.Errorf("no Ready VM instances found with name or UUID: %s", nameOrUUID)
	}

	fmt.Printf("The following VM instances will be deleted:\n")
	for _, vm := range vms {
		fmt.Printf("  - %s (%s)\n", vm.Name, vm.UUID)
	}

	if dryRunFlag {
		fmt.Println("Dry-run: no resources were changed.")
		return nil
	}

	confirmed, err := common.Confirm(cmd, "Are you sure you want to delete the above VM instances? Type 'yes' to confirm: ")
	if err != nil {
		return err
	}
	if !confirmed {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Aborted by user.")
		return err
	}

	var failures []error
	for _, vm := range vms {
		err := cli.DestroyVmInstance(vm.UUID, param.DeleteModePermissive)
		if err != nil {
			failures = append(failures, fmt.Errorf("Failed to delete VM instance %s (%s): %w", vm.Name, vm.UUID, err))
			continue
		}
		fmt.Printf("Deleted VM instance: %s (%s)\n", vm.Name, vm.UUID)

	}

	return errors.Join(failures...)
}

func init() {
	DeleteInstancesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteCmd.AddCommand(DeleteInstancesCmd)

}
