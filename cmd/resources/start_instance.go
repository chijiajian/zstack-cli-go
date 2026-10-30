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

var StartInstanceCmd = &cobra.Command{
	Use:   "start <name-or-uuid>",
	Short: "Start virtual machine instance",
	Long: `Start a virtual machine instance by specifying its name or UUID.
If the identifier matches multiple VMs, all in 'Stopped' state will be started (with confirmation).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStartInstance(cmd, args[0])
	},
}

func init() {
	InstanceCmd.AddCommand(StartInstanceCmd)

	/*
		StartInstanceCmd.Flags().Bool("dry-run", false, "Preview the API request without sending it")
		StartInstanceCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts")
		StartInstanceCmd.Flags().StringP("output", "o", "", "Output format: json, yaml, table, wide, or name")
		StartInstanceCmd.Flags().StringSlice("fields", nil, "Custom fields to display in table output")
	*/
}

func runStartInstance(cmd *cobra.Command, identifier string) error {
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

	var toStart, skipped []sdkView.VmInstanceInventoryView

	for _, vm := range vms {
		if types.IsVMRunnable(vm.State) {
			toStart = append(toStart, vm)
		} else {
			skipped = append(skipped, vm)
		}
	}

	if len(toStart) == 0 {
		fmt.Println("No matched VMs are in 'Stopped' state to start.")
		if len(skipped) > 0 {
			fmt.Println("Matched but skipped (not Stopped):")
			for _, s := range skipped {
				fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
			}
		}
		return nil
	}

	fmt.Printf("Matched %d VM(s); %d will be started, %d will be skipped.\n", len(vms), len(toStart), len(skipped))
	fmt.Println("Will start:")
	for _, s := range toStart {
		fmt.Printf("  - %s (%s)\n", s.Name, s.UUID)
	}
	if len(skipped) > 0 {
		fmt.Println("Skipped (not Stopped):")
		for _, s := range skipped {
			fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
		}
	}

	// dry-run
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

	for _, s := range toStart {
		resp, err := cli.StartVmInstance(s.UUID, param.StartVmInstanceParam{})
		if err != nil {
			failures = append(failures, struct {
				VM  sdkView.VmInstanceInventoryView
				Err error
			}{VM: s, Err: err})
			fmt.Fprintf(os.Stderr, "Failed to start %s (%s): %v\n", s.Name, s.UUID, err)
			continue
		}
		successes = append(successes, *resp)
		fmt.Printf("Started %s (%s)\n", resp.Name, resp.UUID)
	}

	fmt.Printf("\nSummary: %d started, %d failed, %d skipped\n", len(successes), len(failures), len(skipped))
	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "Failures:")
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "  - %s (%s): %v\n", f.VM.Name, f.VM.UUID, f.Err)
		}
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	format := utils.ParseFormat(outputFormat)

	fields, _ := cmd.Flags().GetStringSlice("fields")

	if len(successes) > 0 {
		if err := utils.PrintVMs(successes, format, fields); err != nil {
			return fmt.Errorf("Error formatting output : %w", err)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("%d VM operations failed", len(failures))
	}
	return nil
}
