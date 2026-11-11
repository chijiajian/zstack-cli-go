// Package vmops implements the exact-UUID JSON mode of VM write commands.
package vmops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("uuid-only", false, "Use exact UUIDs with the fixed JSON operation result")
	cmd.Flags().Duration("request-timeout", 2*time.Minute, "Deadline for exact UUID lookup and mutation")
}

func Enabled(cmd *cobra.Command) bool {
	enabled, _ := cmd.Flags().GetBool("uuid-only")
	return enabled
}

var idPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,128}$`)

func Invalid(message string) error { return clierr.Wrap(clierr.Usage, fmt.Errorf("%s", message)) }

func Resolve(ctx context.Context, reader *client.ReadClient, kind, id string) (client.ResolvedResource, error) {
	return reader.ResolveExactUUID(ctx, kind, id)
}

type operation struct {
	ctx    context.Context
	cmd    *cobra.Command
	id     string
	name   string
	dryRun bool
}

func prepare(cmd *cobra.Command, name, id string) (operation, context.CancelFunc, error) {
	output, _ := cmd.Flags().GetString("output")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	timeout, _ := cmd.Flags().GetDuration("request-timeout")
	switch {
	case output != "json" || cmd.Flags().Changed("fields"):
		return operation{}, nil, Invalid("exact UUID mode requires -o json and a fixed result schema")
	case !dryRun && !yes:
		return operation{}, nil, Invalid("exact UUID mode requires --yes or --dry-run")
	case timeout <= 0:
		return operation{}, nil, Invalid("request timeout must be positive")
	case !idPattern.MatchString(id):
		return operation{}, nil, Invalid("a valid VM UUID is required")
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	return operation{ctx: ctx, cmd: cmd, id: id, name: name, dryRun: dryRun}, cancel, nil
}

func (op operation) print(status string, inventory *view.VmInstanceInventoryView, failure error) error {
	item := struct {
		UUID      string                        `json:"uuid"`
		Status    string                        `json:"status"`
		Inventory *view.VmInstanceInventoryView `json:"inventory,omitempty"`
		Error     *clierr.Diagnostic            `json:"error,omitempty"`
	}{UUID: op.id, Status: status, Inventory: inventory}
	if failure != nil {
		diagnostic := clierr.Describe(failure)
		item.Error = &diagnostic
	}
	if err := json.NewEncoder(op.cmd.OutOrStdout()).Encode(map[string]any{"operation": op.name, "items": []any{item}}); err != nil {
		return clierr.Wrap(clierr.Output, err)
	}
	return failure
}

func (op operation) finish(inventory *view.VmInstanceInventoryView, err error) error {
	if err == nil && inventory != nil && inventory.UUID != op.id {
		err = clierr.Wrap(clierr.Remote, fmt.Errorf("mutation response did not identify the requested VM"))
	}
	if err != nil {
		return op.print("failed", nil, err)
	}
	return op.print("succeeded", inventory, nil)
}

func Run(cmd *cobra.Command, name, id string, body any) error {
	op, cancel, err := prepare(cmd, name, id)
	if err != nil {
		return err
	}
	defer cancel()
	api, err := client.NewMutationClient(op.ctx)
	if err != nil {
		return err
	}
	vm, err := Resolve(op.ctx, api.Reader(), "instance", id)
	if err != nil {
		return err
	}
	allowed, satisfied := states(name, body)
	if satisfied != "" && vm.State == satisfied || op.dryRun {
		return op.print("skipped", nil, nil)
	}
	if !slices.Contains(allowed, vm.State) {
		return op.finish(nil, clierr.Wrap(clierr.Conflict, fmt.Errorf("VM state %q does not permit %s", vm.State, name)))
	}
	path := "v1/vm-instances/" + url.PathEscape(id)
	if name == "delete" {
		if err := checkDeletionPolicy(op.ctx, api.Reader()); err != nil {
			return op.finish(nil, err)
		}
		return op.finish(nil, api.Delete(op.ctx, path, "Permissive"))
	}
	var inventory view.VmInstanceInventoryView
	err = api.Mutate(op.ctx, http.MethodPut, path+"/actions", body, &inventory)
	return op.finish(&inventory, err)
}

func states(name string, body any) (allowed []string, satisfied string) {
	switch name {
	case "pause":
		return []string{"Running"}, "Paused"
	case "resume":
		return []string{"Paused"}, "Running"
	case "stop":
		return []string{"Running"}, "Stopped"
	case "restart":
		return []string{"Running"}, ""
	case "delete":
		return []string{"Stopped"}, "Destroyed"
	case "recover":
		return []string{"Destroyed"}, ""
	case "update":
		params := body.(param.UpdateVmInstanceParam).Params
		if params.CpuNum != nil || params.MemorySize != nil {
			return []string{"Stopped"}, ""
		}
		return []string{"Stopped", "Running", "Paused"}, ""
	}
	return nil, ""
}

func checkDeletionPolicy(ctx context.Context, reader *client.ReadClient) error {
	var config struct {
		Inventories []struct{ Category, Name, Value string }
	}
	query := url.Values{"q": {"category=vm", "name=deletionPolicy"}, "limit": {"2"}}
	if err := reader.Read(ctx, "v1/global-configurations", query, &config); err != nil {
		return err
	}
	if len(config.Inventories) != 1 {
		return clierr.Wrap(clierr.Conflict, fmt.Errorf("VM deletion policy could not be confirmed"))
	}
	policy := config.Inventories[0]
	if policy.Category != "vm" || policy.Name != "deletionPolicy" || policy.Value != "Delay" && policy.Value != "Never" {
		return clierr.Wrap(clierr.Conflict, fmt.Errorf("recoverable deletion requires VM deletion policy Delay or Never"))
	}
	return nil
}

func Create(cmd *cobra.Command, params param.CreateVmInstanceParam) error {
	p := params.Params
	op, cancel, err := prepare(cmd, "create", *p.ResourceUuid)
	if err != nil {
		return err
	}
	defer cancel()
	api, err := client.NewMutationClient(op.ctx)
	if err != nil {
		return err
	}
	if !idPattern.MatchString(*p.ImageUuid) {
		return Invalid("a valid image UUID is required")
	}
	var images struct {
		Inventories []struct{ UUID, MediaType string }
	}
	if err := api.Reader().Read(op.ctx, "v1/images/"+url.PathEscape(*p.ImageUuid), nil, &images); err != nil {
		return err
	}
	if len(images.Inventories) != 1 || images.Inventories[0].UUID != *p.ImageUuid || images.Inventories[0].MediaType != "RootVolumeTemplate" {
		return clierr.Wrap(clierr.Conflict, fmt.Errorf("the specified image must be a root volume template"))
	}
	if p.InstanceOfferingUuid != nil {
		if _, err := Resolve(op.ctx, api.Reader(), "instance-offering", *p.InstanceOfferingUuid); err != nil {
			return err
		}
	}
	for _, network := range p.L3NetworkUuids {
		if _, err := Resolve(op.ctx, api.Reader(), "l3-network", network); err != nil {
			return err
		}
	}
	if _, err := Resolve(op.ctx, api.Reader(), "instance", op.id); clierr.Describe(err).Code != clierr.NotFound {
		if err == nil {
			err = clierr.Wrap(clierr.Conflict, fmt.Errorf("VM UUID already exists"))
		}
		return err
	}
	if op.dryRun {
		return op.print("skipped", nil, nil)
	}
	var inventory view.VmInstanceInventoryView
	err = api.Mutate(op.ctx, http.MethodPost, "v1/vm-instances", params, &inventory)
	return op.finish(&inventory, err)
}
