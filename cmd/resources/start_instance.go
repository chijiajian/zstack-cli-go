package resources

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	sdkView "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

var StartInstanceCmd = &cobra.Command{
	Use:   "start <name-or-uuid>",
	Short: "Start virtual machine instance",
	Long: `Start a virtual machine instance by specifying its name or UUID.
If the identifier matches multiple VMs, all in 'Stopped' state will be started (with confirmation).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStartInstance(cmd, args[0])
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

func runStartInstance(cmd *cobra.Command, identifier string) {
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

	var toStart, skipped []sdkView.VmInstanceInventoryView

	for _, vm := range vms {
		if vm.State == "Stopped" {
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
		return
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

	for _, s := range toStart {
		resp, err := cli.StartVmInstance(s.UUID, nil)
		if err != nil {
			failures = append(failures, struct {
				VM  sdkView.VmInstanceInventoryView
				Err error
			}{VM: s, Err: err})
			fmt.Printf("Failed to start %s (%s): %v\n", s.Name, s.UUID, err)
			continue
		}
		successes = append(successes, *resp)
		fmt.Printf("Started %s (%s)\n", resp.Name, resp.UUID)
	}

	fmt.Printf("\nSummary: %d started, %d failed, %d skipped\n", len(successes), len(failures), len(skipped))
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
			fmt.Printf("Error formatting output : %s\n", err)
		}
	}
}
