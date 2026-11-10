package resources

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
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
If the identifier matches multiple VMs, all in 'Stopped' state will be started (with confirmation).
Use --uuid-only for automation to disable name fallback. JSON output contains one operation/items result, including skipped and failed targets.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStartInstance(cmd, args[0])
	},
}

func init() {
	InstanceCmd.AddCommand(StartInstanceCmd)
	StartInstanceCmd.Flags().Bool("uuid-only", false, "Require a single exact UUID match without falling back to a name")
	StartInstanceCmd.Flags().Duration("request-timeout", 2*time.Minute, "Timeout for resolving targets and for each start request")
}

type startItem struct {
	UUID      string                           `json:"uuid"`
	Status    string                           `json:"status"`
	Inventory *sdkView.VmInstanceInventoryView `json:"inventory,omitempty"`
	Error     *clierr.Diagnostic               `json:"error,omitempty"`
}

func runStartInstance(cmd *cobra.Command, identifier string) error {
	output, _ := cmd.Flags().GetString("output")
	format := utils.ParseFormat(output)
	if format == utils.JSONFormat && cmd.Flags().Changed("fields") {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("start JSON has a fixed schema and does not support --fields"))
	}
	timeout, _ := cmd.Flags().GetDuration("request-timeout")
	if timeout <= 0 {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("request timeout must be positive"))
	}
	uuidOnly, _ := cmd.Flags().GetBool("uuid-only")
	options := client.ResolveOptions{AllowMultiple: true}
	if uuidOnly {
		options = client.ResolveOptions{Strategy: client.ResolveUUID}
	}
	resolveCtx, cancelResolve := context.WithTimeout(cmd.Context(), timeout)
	cli, err := client.NewMutationClient(resolveCtx)
	if err != nil {
		cancelResolve()
		return err
	}
	vms, err := cli.Reader().ResolveMatches(resolveCtx, "instance", identifier, options)
	cancelResolve()
	if err != nil {
		return err
	}
	if uuidOnly && vms[0].UUID != identifier {
		return clierr.Wrap(clierr.Remote, fmt.Errorf("UUID query returned a different instance"))
	}

	var toStart []client.ResolvedResource
	items := make([]startItem, 0, len(vms))
	var skipped []client.ResolvedResource
	for _, vm := range vms {
		if !uuidOnly && !types.IsVMActive(vm.State) {
			continue
		}
		if types.IsVMRunnable(vm.State) {
			toStart = append(toStart, vm)
		} else {
			skipped = append(skipped, vm)
			items = append(items, startItem{UUID: vm.UUID, Status: "skipped"})
		}
	}
	if len(toStart)+len(skipped) == 0 {
		return clierr.Wrap(clierr.NotFound, fmt.Errorf("no active instance matched %q", identifier))
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	autoYes, _ := cmd.Flags().GetBool("yes")
	var human io.Writer = cmd.OutOrStdout()
	if format == utils.JSONFormat {
		human = io.Discard
		if !dryRun && !autoYes && len(toStart) > 0 {
			human = cmd.ErrOrStderr()
		}
	}
	if len(toStart) == 0 {
		fmt.Fprintln(human, "No matched VMs are in 'Stopped' state to start.")
	} else {
		fmt.Fprintf(human, "Matched %d VM(s); %d will be started, %d will be skipped.\nWill start:\n", len(toStart)+len(skipped), len(toStart), len(skipped))
		for _, vm := range toStart {
			fmt.Fprintf(human, "  - %s (%s)\n", vm.Name, vm.UUID)
		}
	}
	if len(skipped) > 0 {
		fmt.Fprintln(human, "Skipped (not Stopped):")
		for _, vm := range skipped {
			fmt.Fprintf(human, "  - %s (%s) state=%s\n", vm.Name, vm.UUID, vm.State)
		}
	}
	if dryRun && len(toStart) > 0 {
		fmt.Fprintln(human, "Dry-run: no start requests will be sent.")
	}
	if !dryRun && !autoYes && len(toStart) > 0 {
		if _, err := fmt.Fprint(human, "Do you want to continue? [y/N]: "); err != nil {
			return clierr.Wrap(clierr.Output, err)
		}
		line, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(human, "Aborted by user.")
			dryRun = true
		}
	}
	if dryRun {
		for _, vm := range toStart {
			items = append(items, startItem{UUID: vm.UUID, Status: "skipped"})
		}
		toStart = nil
	}

	var successes []sdkView.VmInstanceInventoryView
	var failed int
	var firstFailure error
	for _, vm := range toStart {
		var resp sdkView.VmInstanceInventoryView
		requestCtx, cancel := context.WithTimeout(cmd.Context(), timeout)
		err := cli.Mutate(requestCtx, http.MethodPut, "v1/vm-instances/"+url.PathEscape(vm.UUID)+"/actions", param.StartVmInstanceParam{}, &resp)
		cancel()
		if err == nil && resp.UUID != vm.UUID {
			err = clierr.Wrap(clierr.Remote, fmt.Errorf("start response did not identify the requested instance"))
		}
		if err != nil {
			diagnostic := clierr.Describe(err)
			items = append(items, startItem{UUID: vm.UUID, Status: "failed", Error: &diagnostic})
			if firstFailure == nil {
				firstFailure = err
			}
			failed++
			if format != utils.JSONFormat {
				fmt.Fprintf(cmd.ErrOrStderr(), "Failed to start %s (%s): %v\n", vm.Name, vm.UUID, err)
			}
			continue
		}
		items = append(items, startItem{UUID: vm.UUID, Status: "succeeded", Inventory: &resp})
		successes = append(successes, resp)
		if format != utils.JSONFormat {
			fmt.Fprintf(human, "Started %s (%s)\n", resp.Name, resp.UUID)
		}
	}

	if format == utils.JSONFormat {
		if err := utils.Print(struct {
			Operation string      `json:"operation"`
			Items     []startItem `json:"items"`
		}{Operation: "start", Items: items}, format); err != nil {
			return fmt.Errorf("format start result: %w", err)
		}
	} else if len(toStart) > 0 {
		fmt.Fprintf(human, "\nSummary: %d started, %d failed, %d skipped\n", len(successes), failed, len(skipped))
		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(successes) > 0 {
			if err := utils.PrintVMs(successes, format, fields); err != nil {
				return fmt.Errorf("format start result: %w", err)
			}
		}
	}
	if failed == 1 && len(toStart) == 1 && format == utils.JSONFormat {
		return firstFailure
	}
	if failed > 0 {
		return fmt.Errorf("%d VM operations failed", failed)
	}
	return nil
}
