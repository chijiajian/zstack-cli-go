package resources

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
	sdkView "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

var StopInstanceCmd = &cobra.Command{
	Use:   "stop <name-or-uuid>",
	Short: "Stop a virtual machine instance (name or uuid).",
	Long: `Stop a VM instance by name or UUID. 
Supports batch stop if multiple matches are found. You will be prompted
for confirmation unless -y/--yes is provided.

Example:
  zstack-cli instance stop my-vm`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStopInstance(cmd, args[0])
	},
}

func init() {
	InstanceCmd.AddCommand(StopInstanceCmd)
	StopInstanceCmd.Flags().Bool("stop-ha", true, "Completely shut down HA VM if applicable")
	//StopInstanceCmd.Flags().String("stop-type", "grace", "grace stop or not")
}

func runStopInstance(cmd *cobra.Command, identifier string) {
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

	var toStop, skipped []sdkView.VmInstanceInventoryView
	for _, vm := range vms {
		if vm.State == "Running" {
			toStop = append(toStop, vm)
		} else {
			skipped = append(skipped, vm)
		}
	}

	if len(toStop) == 0 {
		fmt.Println("No matched VMs are in 'Running' state to stop.")
		if len(skipped) > 0 {
			fmt.Println("Matched but skipped (not Running):")
			for _, s := range skipped {
				fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
			}
		}
		return
	}

	fmt.Printf("Matched %d VM(s); %d will be stopped, %d will be skipped.\n", len(vms), len(toStop), len(skipped))
	fmt.Println("Will stop:")
	for _, s := range toStop {
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

	stopHA, _ := cmd.Flags().GetBool("stop-ha")
	//stopType, _ := cmd.Flags().GetString("stop-type")
	var successes []sdkView.VmInstanceInventoryView
	var failures []struct {
		VM  sdkView.VmInstanceInventoryView
		Err error
	}

	p := param.StopVmInstanceParam{
		StopVmInstance: param.StopVmInstanceDetailParam{
			Type:   "grace",
			StopHA: stopHA, //bug if true bu restart auto
		},
	}

	for _, vm := range toStop {

		resp, err := cli.StopVmInstance(vm.UUID, p)
		if err != nil {
			failures = append(failures, struct {
				VM  sdkView.VmInstanceInventoryView
				Err error
			}{VM: vm, Err: err})
			fmt.Printf("Failed to stop %s (%s): %v\n", vm.Name, vm.UUID, err)
			continue
		}
		successes = append(successes, *resp)
		fmt.Printf("Stopped %s (%s)\n", resp.Name, resp.UUID)
	}

	fmt.Printf("\nSummary: %d stopped, %d failed, %d skipped\n", len(successes), len(failures), len(skipped))
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
