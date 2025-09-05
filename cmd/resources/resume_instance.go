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

var ResumeInstanceCmd = &cobra.Command{
	Use:   "resume <name-or-uuid>",
	Short: "Resume a paused virtual machine instance (name or uuid).",
	Long: `Resume a VM instance that is in 'Paused' state by name or UUID. 
Supports batch resume if multiple matches are found. You will be prompted
for confirmation unless -y/--yes is provided.

Example:
  zstack-cli instance resume my-paused-vm`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runResumeInstance(cmd, args[0])
	},
}

func init() {
	InstanceCmd.AddCommand(ResumeInstanceCmd)
}

func runResumeInstance(cmd *cobra.Command, identifier string) {
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

	var toResume, skipped []sdkView.VmInstanceInventoryView
	for _, vm := range vms {
		if vm.State == "Paused" {
			toResume = append(toResume, vm)
		} else {
			skipped = append(skipped, vm)
		}
	}

	if len(toResume) == 0 {
		fmt.Println("No matched VMs are in 'Paused' state to resume.")
		if len(skipped) > 0 {
			fmt.Println("Matched but skipped (not Paused):")
			for _, s := range skipped {
				fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
			}
		}
		return
	}

	fmt.Printf("Matched %d VM(s); %d will be resumed, %d will be skipped.\n", len(vms), len(toResume), len(skipped))
	fmt.Println("Will resume:")
	for _, s := range toResume {
		fmt.Printf("  - %s (%s)\n", s.Name, s.UUID)
	}
	if len(skipped) > 0 {
		fmt.Println("Skipped (not Paused):")
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

	for _, vm := range toResume {
		resp, err := cli.ResumeVmInstance(vm.UUID)
		if err != nil {
			failures = append(failures, struct {
				VM  sdkView.VmInstanceInventoryView
				Err error
			}{VM: vm, Err: err})
			fmt.Printf("Failed to resume %s (%s): %v\n", vm.Name, vm.UUID, err)
			continue
		}
		successes = append(successes, *resp)
		fmt.Printf("Resumed %s (%s)\n", resp.Name, resp.UUID)
	}

	fmt.Printf("\nSummary: %d resumed, %d failed, %d skipped\n", len(successes), len(failures), len(skipped))
	if len(failures) > 0 {
		fmt.Println("Failures:")
		for _, f := range failures {
			fmt.Printf("  - %s (%s): %v\n", f.VM.Name, f.VM.UUID, f.Err)
		}
	}

	// 输出格式
	outputFormat, _ := cmd.Flags().GetString("output")
	format := utils.ParseFormat(outputFormat)
	fields, _ := cmd.Flags().GetStringSlice("fields")

	if len(successes) > 0 {
		if err := utils.PrintVMs(successes, format, fields); err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
		}
	}
}
