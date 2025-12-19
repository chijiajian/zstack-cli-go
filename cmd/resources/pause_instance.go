package resources

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/types"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	sdkView "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

var PauseInstanceCmd = &cobra.Command{
	Use:   "pause <name-or-uuid>",
	Short: "Pause a virtual machine instance (name or uuid).",
	Long: `Pause a VM instance by name or UUID. 
Supports batch pause if multiple matches are found. You will be prompted
for confirmation unless -y/--yes is provided.

Example:
  zstack-cli instance pause my-vm`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runPauseInstance(cmd, args[0])
	},
}

func init() {
	InstanceCmd.AddCommand(PauseInstanceCmd)
}

func runPauseInstance(cmd *cobra.Command, identifier string) {
	cli := client.GetClient()
	if cli == nil {
		fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
		return
	}

	vms, err := client.GetReadyVMsByNameOrUUID(cli, identifier)
	if err != nil {
		fmt.Printf("Error querying VMs: %v\n", err)
		return
	}
	if len(vms) == 0 {
		fmt.Printf("No VMs found with name or UUID '%s'.\n", identifier)
		return
	}

	var toPause, skipped []sdkView.VmInstanceInventoryView
	for _, vm := range vms {
		if vm.State == types.VMStateRunning {
			toPause = append(toPause, vm)
		} else {
			skipped = append(skipped, vm)
		}
	}

	if len(toPause) == 0 {
		fmt.Println("No matched VMs are in 'Running' state to pause.")
		if len(skipped) > 0 {
			fmt.Println("Matched but skipped (not Running):")
			for _, s := range skipped {
				fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
			}
		}
		return
	}

	fmt.Printf("Matched %d VM(s); %d will be paused, %d will be skipped.\n", len(vms), len(toPause), len(skipped))
	fmt.Println("Will pause:")
	for _, s := range toPause {
		fmt.Printf("  - %s (%s)\n", s.Name, s.UUID)
	}
	if len(skipped) > 0 {
		fmt.Println("Skipped (not Running):")
		for _, s := range skipped {
			fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
		}
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("Dry-run: no API calls will be made.")
		return
	}

	autoYes, _ := cmd.Flags().GetBool("yes")
	if !autoYes {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to continue? [y/N]: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Println("Aborted by user.")
			return
		}
	}

	var successes []sdkView.VmInstanceInventoryView
	var failures []struct {
		VM  sdkView.VmInstanceInventoryView
		Err error
	}

	for _, vm := range toPause {
		resp, err := cli.PauseVmInstance(vm.UUID)
		if err != nil {
			failures = append(failures, struct {
				VM  sdkView.VmInstanceInventoryView
				Err error
			}{VM: vm, Err: err})
			fmt.Printf("Failed to pause %s (%s): %v\n", vm.Name, vm.UUID, err)
			continue
		}
		successes = append(successes, *resp)
		fmt.Printf("Paused %s (%s)\n", resp.Name, resp.UUID)
	}

	fmt.Printf("\nSummary: %d paused, %d failed, %d skipped\n", len(successes), len(failures), len(skipped))
	if len(failures) > 0 {
		fmt.Println("Failures:")
		for _, f := range failures {
			fmt.Printf("  - %s (%s): %v\n", f.VM.Name, f.VM.UUID, f.Err)
		}
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	format := utils.ParseFormat(outputFormat)
	fields, _ := cmd.Flags().GetStringSlice("fields")

	if len(successes) > 0 {
		if err := utils.PrintVMs(successes, format, fields); err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
		}
	}
}
