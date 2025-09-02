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
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"

	"github.com/spf13/cobra"
)

var ExpungeInstancesCmd = &cobra.Command{
	Use:   "instance [name-or-uuid]",
	Short: "Expunge one or many VM instances",
	Long: `Expunge one or many ZStack VM instances.

Examples:
  # Delete a single VM instance by name or UUID
  zstack-cli expunge instance my-vm

  # Delete multiple VM instances by name
  zstack-cli expunge instance my-vm 

Note: This is a dangerous operation. You must confirm by typing 'yes'.`,
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

	vms, err := client.GetDestroyedVMsByNameOrUUID(cli, nameOrUUID)
	if err != nil {
		return fmt.Errorf("query VM instances failed: %s", err)
	}

	if len(vms) == 0 {
		return fmt.Errorf("no Destroyed VM instances found with name or UUID: %s", nameOrUUID)
	}

	fmt.Printf("The following VM instances will be Expunge:\n")
	for _, vm := range vms {
		fmt.Printf("  - %s (%s)\n", vm.Name, vm.UUID)
	}

	var input string
	fmt.Print("Are you sure you want to Expunge the above VM instances? Type 'yes' to confirm: ")
	fmt.Scanln(&input)
	if input != "yes" {
		fmt.Println("Aborted by user.")
		return nil
	}

	for _, vm := range vms {
		err := cli.ExpungeVmInstance(vm.UUID)
		if err != nil {
			fmt.Printf("Failed to Expunge VM instance %s (%s): %s\n", vm.Name, vm.UUID, err)
			continue
		}
		fmt.Printf("Expunged VM instance: %s (%s)\n", vm.Name, vm.UUID)

	}

	return nil
}

func init() {
	ExpungeCmd.AddCommand(ExpungeInstancesCmd)

}
