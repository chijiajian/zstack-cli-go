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
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	sdkView "github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

var RestartInstanceCmd = &cobra.Command{
	Use:   "restart <name-or-uuid>",
	Short: "Restart a virtual machine instance (name or uuid).",
	Long: `Restart a VM instance by name or UUID. 
Supports batch restart if multiple matches are found. You will be prompted
for confirmation unless -y/--yes is provided.

Example:
  zstack-cli instance restart my-vm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRestartInstance(cmd, args[0])
	},
}

func init() {
	InstanceCmd.AddCommand(RestartInstanceCmd)
}

func runRestartInstance(cmd *cobra.Command, identifier string) error {
	cli := client.GetClient()
	if cli == nil {
		return fmt.Errorf("Not logged in. Please run 'zstack-cli login' first.")
	}

	vms, err := client.GetReadyVMsByNameOrUUID(cli, identifier)
	if err != nil {
		return fmt.Errorf("Error querying VMs: %w", err)
	}
	if len(vms) == 0 {
		return fmt.Errorf("No VMs found with name or UUID '%s'.", identifier)
	}

	var toRestart, skipped []sdkView.VmInstanceInventoryView
	for _, vm := range vms {
		if vm.State == types.VMStateRunning {
			toRestart = append(toRestart, vm)
		} else {
			skipped = append(skipped, vm)
		}
	}

	if len(toRestart) == 0 {
		fmt.Println("No matched VMs are in 'Running' state to restart.")
		if len(skipped) > 0 {
			fmt.Println("Matched but skipped (not Running):")
			for _, s := range skipped {
				fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
			}
		}
		return nil
	}

	fmt.Printf("Matched %d VM(s); %d will be restarted, %d will be skipped.\n", len(vms), len(toRestart), len(skipped))
	fmt.Println("Will restart:")
	for _, s := range toRestart {
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
		return nil
	}

	autoYes, _ := cmd.Flags().GetBool("yes")
	if !autoYes {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to continue? [y/N]: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Println("Aborted by user.")
			return nil
		}
	}

	var successes []sdkView.VmInstanceInventoryView
	var failures []struct {
		VM  sdkView.VmInstanceInventoryView
		Err error
	}

	for _, vm := range toRestart {
		resp, err := cli.RebootVmInstance(vm.UUID, param.RebootVmInstanceParam{})
		if err != nil {
			failures = append(failures, struct {
				VM  sdkView.VmInstanceInventoryView
				Err error
			}{VM: vm, Err: err})
			fmt.Fprintf(os.Stderr, "Failed to restart %s (%s): %v\n", vm.Name, vm.UUID, err)
			continue
		}
		successes = append(successes, *resp)
		fmt.Printf("Restarted %s (%s)\n", resp.Name, resp.UUID)
	}

	fmt.Printf("\nSummary: %d restarted, %d failed, %d skipped\n", len(successes), len(failures), len(skipped))
	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "Failures:")
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "  - %s (%s): %v\n", f.VM.Name, f.VM.UUID, f.Err)
		}
	}

	// 输出格式
	outputFormat, _ := cmd.Flags().GetString("output")
	format := utils.ParseFormat(outputFormat)
	fields, _ := cmd.Flags().GetStringSlice("fields")

	if len(successes) > 0 {
		if err := utils.PrintVMs(successes, format, fields); err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("%d VM operations failed", len(failures))
	}
	return nil
}
