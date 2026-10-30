// Package iam registers identity operations without coupling parent command packages.
package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type resource struct{ kind, path, update string }

var resources = []resource{
	{"account", "v1/accounts", "updateAccount"}, {"user", "v1/accounts/users", "updateUser"},
	{"role", "v1/identities/roles", "updateRole"}, {"policy", "v1/accounts/policies", ""},
	{"project", "v1/iam2/projects", "updateIAM2Project"}, {"organization", "v1/iam2/organizations", "updateIAM2Organization"},
}

type inventory struct {
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty" header:"NAME"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" header:"DESCRIPTION"`
	State       string `json:"state,omitempty" yaml:"state,omitempty" header:"STATE"`
}

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

func Register(create, update, delete *cobra.Command) {
	for _, spec := range resources {
		create.AddCommand(command(spec, "create"))
		create.ValidArgs = append(create.ValidArgs, spec.kind)
		delete.AddCommand(command(spec, "delete"))
		delete.ValidArgs = append(delete.ValidArgs, spec.kind)
		if spec.update != "" {
			update.AddCommand(command(spec, "update"))
			update.ValidArgs = append(update.ValidArgs, spec.kind)
		}
	}
}
func command(spec resource, action string) *cobra.Command {
	cmd := &cobra.Command{Use: spec.kind + " <name-or-uuid>", Short: action + " an IAM " + spec.kind, Args: cobra.ExactArgs(1)}
	if action == "create" {
		cmd.Use = spec.kind + " <name>"
	}
	if spec.kind == "user" {
		cmd.Long = "Manage a user belonging to the account of the selected context. Password files contain plaintext; request previews redact passwords."
	}
	cmd.Flags().Duration("request-timeout", 2*time.Minute, "Deadline for authentication, resolution, mutation and API job polling")
	if action == "delete" {
		cmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
		cmd.Flags().String("delete-mode", "Permissive", "Deletion mode: Permissive or Enforcing")
	} else {
		cmd.Flags().StringP("description", "d", "", "Description (explicit empty clears it)")
		if action == "create" {
			cmd.Flags().String("resource-uuid", "", "UUID for the new resource")
		} else {
			cmd.Flags().String("name", "", "New resource name")
		}
		if spec.kind == "account" || spec.kind == "user" {
			cmd.Flags().String("password-file", "", "Read the password from a local file; previews redact it")
		}
		if spec.kind == "account" && action == "create" || spec.kind == "organization" {
			cmd.Flags().String("type", "", "Account or organization type")
		}
		if spec.kind == "role" || spec.kind == "policy" {
			cmd.Flags().String("statements-file", "", "JSON array of policy statements")
		}
		if spec.kind == "role" {
			cmd.Flags().StringSlice("policy", nil, "Policy names or UUIDs")
			if action == "create" {
				cmd.Flags().String("identity", "", "Role identity")
			}
		}
		if spec.kind == "organization" {
			cmd.Flags().String("parent", "", "Parent organization name or UUID")
		}
		if spec.kind == "project" && action == "create" {
			cmd.Flags().String("organization", "", "Organization name or UUID")
			cmd.Flags().String("link-account", "", "Account name or UUID to link")
		}
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		preview, _ := cmd.Flags().GetBool("dry-run")
		if action == "create" && text(cmd, "file") != "" {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("IAM commands use flags and --statements-file; --file is unsupported"))
		}
		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(fields) > 0 {
			if err := utils.ValidateOutputFields([]inventory{}, fields); err != nil {
				return clierr.Wrap(clierr.Usage, err)
			}
		}
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("request-timeout must be positive"))
		}
		if err := nonempty(args[0], "resource name or UUID"); err != nil {
			return clierr.Wrap(clierr.Usage, err)
		}
		if cmd.Flags().Changed("resource-uuid") && !uuidPattern.MatchString(text(cmd, "resource-uuid")) {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("resource-uuid must contain 32 hexadecimal characters"))
		}
		method, path := http.MethodPost, spec.path
		var request any
		if action == "delete" {
			mode := text(cmd, "delete-mode")
			if mode != "Permissive" && mode != "Enforcing" {
				return clierr.Wrap(clierr.Usage, fmt.Errorf("delete-mode must be Permissive or Enforcing"))
			}
			method = http.MethodDelete
		} else {
			var err error
			request, err = build(cmd, spec.kind, action, args[0], preview)
			if err != nil {
				return clierr.Wrap(clierr.Usage, err)
			}
			if action == "update" {
				method = http.MethodPut
			}
		}
		if preview {
			format := text(cmd, "output")
			if format == "" || format == "text" {
				format = "yaml"
			}
			return utils.PrintDryRun(map[string]any{"resource": spec.kind, "operation": action, "target": args[0], "request": request}, format)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		cli, err := client.NewMutationClient(ctx)
		if err != nil {
			return err
		}
		var target client.ResolvedResource
		if action != "create" {
			target, err = cli.Reader().Resolve(ctx, spec.kind, args[0], client.ResolveAuto)
			if err != nil {
				return err
			}
			path += "/" + url.PathEscape(target.UUID)
			if action == "update" && spec.kind != "account" {
				path += "/actions"
			}
			if action == "delete" {
				yes, err := common.Confirm(cmd, fmt.Sprintf("Delete %s %s (%s)? Type 'yes' to confirm: ", spec.kind, target.Name, target.UUID))
				if err != nil {
					return err
				}
				if !yes {
					_, err = fmt.Fprintln(cmd.OutOrStdout(), "Aborted by user.")
					return err
				}
			}
		}
		// Resolve references only after offline validation; preserve the SDK wire keys.
		if request != nil {
			data, err := json.Marshal(request)
			if err != nil {
				return err
			}
			var envelope map[string]map[string]any
			if err = json.Unmarshal(data, &envelope); err != nil {
				return err
			}
			key := "params"
			if action == "update" {
				key = spec.update
			}
			params := envelope[key]
			for field, kind := range map[string]string{"organizationUuid": "organization", "parentUuid": "organization", "linkAccountUuid": "account"} {
				if name, ok := params[field].(string); ok {
					resolved, err := cli.Reader().Resolve(ctx, kind, name, client.ResolveAuto)
					if err != nil {
						return err
					}
					params[field] = resolved.UUID
				}
			}
			if values, ok := params["policyUuids"].([]any); ok {
				for i, value := range values {
					resolved, err := cli.Reader().Resolve(ctx, "policy", value.(string), client.ResolveAuto)
					if err != nil {
						return err
					}
					values[i] = resolved.UUID
				}
			}
			if spec.kind == "user" && action == "update" {
				path = spec.path + "/actions"
				params["uuid"] = target.UUID
			}
			request = envelope
		}
		var result inventory
		if action == "delete" {
			// The API accepts deleteMode as a query parameter; use the typed transport path.
			err = cli.Delete(ctx, path, text(cmd, "delete-mode"))
			if err == nil {
				result = inventory{UUID: target.UUID, Name: target.Name}
			}
		} else {
			err = cli.Mutate(ctx, method, path, request, &result)
		}
		if err != nil {
			return err
		}
		if result.UUID == "" {
			return clierr.Wrap(clierr.Remote, fmt.Errorf("mutation response missing resource UUID"))
		}
		expectedUUID := target.UUID
		if action == "create" {
			expectedUUID = strings.ToLower(text(cmd, "resource-uuid"))
		}
		if expectedUUID != "" && result.UUID != expectedUUID {
			return clierr.Wrap(clierr.Remote, fmt.Errorf("mutation response identity mismatch"))
		}
		format := text(cmd, "output")
		if format == "" || format == "text" || format == "name" {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s/%s %s\n", spec.kind, result.UUID, map[string]string{"create": "created", "update": "updated", "delete": "deleted"}[action])
			return err
		}
		if format == "wide" {
			return utils.PrintOperationResult(spec.kind, result, format)
		}
		return utils.PrintWithFields(result, utils.ParseFormat(format), fields)
	}
	return cmd
}
