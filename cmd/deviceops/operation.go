// Package deviceops implements bounded, exact-UUID automation for VM devices.
package deviceops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/chijiajian/zstack-cli-go/cmd/vmops"
	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/spf13/cobra"
)

var idPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,128}$`)

type operation struct {
	ctx context.Context
	api *client.MutationClient
}

type action func(operation) (id string, skipped bool, err error)

func run(cmd *cobra.Command, name, target string, execute action) error {
	format, _ := cmd.Flags().GetString("output")
	yes, _ := cmd.Flags().GetBool("yes")
	preview, _ := cmd.Flags().GetBool("dry-run")
	timeout, _ := cmd.Flags().GetDuration("request-timeout")
	if format != "json" || cmd.Flags().Changed("fields") || !yes && !preview || timeout <= 0 || !idPattern.MatchString(target) {
		return vmops.Invalid("exact UUID mode requires a UUID, -o json, --yes or --dry-run, and a positive timeout")
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()
	result := struct {
		Operation string `json:"operation"`
		Target    string `json:"targetUuid"`
		UUID      string `json:"uuid,omitempty"`
		Status    string `json:"status"`
	}{Operation: name, Target: target, Status: "skipped"}
	if !preview {
		api, err := client.NewMutationClient(ctx)
		if err != nil {
			return err
		}
		id, skipped, err := execute(operation{ctx, api})
		if err != nil {
			return err
		}
		if !idPattern.MatchString(id) {
			return unverified("operation returned an invalid resource UUID")
		}
		result.UUID = id
		if !skipped {
			result.Status = "succeeded"
		}
	}
	if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
		return clierr.Wrap(clierr.Output, err)
	}
	return nil
}

func conflict(message string) error   { return clierr.Wrap(clierr.Conflict, fmt.Errorf("%s", message)) }
func unverified(message string) error { return clierr.Wrap(clierr.Remote, fmt.Errorf("%s", message)) }

type resource struct {
	UUID, Name, Type, State, Status, MediaType         string
	VmInstanceUuid, VolumeUuid, L3NetworkUuid, IsoUuid string
	DiskOfferingUuid, PrimaryStorageUuid, Mac, Ip      string
	Size, DiskSize                                     int64
	UsedIps                                            []struct{ Ip, L3NetworkUuid string }
}

func (op operation) read(path string, limit int, conditions ...string) ([]resource, error) {
	var result struct{ Inventories []resource }
	q := url.Values{"q": conditions, "limit": {fmt.Sprint(limit)}}
	err := op.api.Reader().Read(op.ctx, path, q, &result)
	return result.Inventories, err
}

func (op operation) exact(path, id string) (resource, error) {
	if !idPattern.MatchString(id) {
		return resource{}, vmops.Invalid("a valid resource UUID is required")
	}
	rows, err := op.read(path, 2, "uuid="+id)
	if err != nil {
		return resource{}, err
	}
	if len(rows) == 0 {
		return resource{}, clierr.Wrap(clierr.NotFound, fmt.Errorf("resource UUID was not found"))
	}
	if len(rows) != 1 || rows[0].UUID != id {
		return resource{}, unverified("lookup returned an unexpected resource identity")
	}
	return rows[0], nil
}

func (op operation) absent(path, id string) error {
	_, err := op.exact(path, id)
	if clierr.Describe(err).Code == clierr.NotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return conflict("resource UUID already exists")
}

func (op operation) removed(path, id string) error {
	if err := op.absent(path, id); err != nil {
		return unverified("resource deletion was not observed")
	}
	return nil
}

func (op operation) vm(id string, stopped bool) (resource, error) {
	vm, err := op.exact("v1/vm-instances", id)
	if err != nil {
		return vm, err
	}
	if vm.State != "Stopped" && (stopped || vm.State != "Running") {
		return vm, conflict("operation requires a supported VM power state")
	}
	return vm, nil
}

func (op operation) policy(category string) error {
	var result struct {
		Inventories []struct{ Category, Name, Value string }
	}
	err := op.api.Reader().Read(op.ctx, "v1/global-configurations", url.Values{"q": {"category=" + category, "name=deletionPolicy"}, "limit": {"2"}}, &result)
	if err != nil {
		return err
	}
	if len(result.Inventories) != 1 {
		return conflict("deletion policy could not be verified")
	}
	config := result.Inventories[0]
	if config.Category != category || config.Name != "deletionPolicy" || config.Value != "Delay" && config.Value != "Never" {
		return conflict("recoverable deletion requires Delay or Never policy")
	}
	return nil
}

func LegacyDetachISO(ctx context.Context, vm, iso string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	api, err := client.NewMutationClient(ctx)
	if err != nil {
		return err
	}
	return (operation{ctx, api}).eject(vm, iso)
}
