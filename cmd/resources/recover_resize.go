package resources

import (
	"fmt"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	volumes "github.com/chijiajian/zstack-cli-go/pkg/volume"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type resourceActionPreview struct {
	ResourceType string `json:"resourceType" yaml:"resourceType"`
	Resource     string `json:"resource" yaml:"resource"`
	Action       string `json:"action" yaml:"action"`
	Size         int64  `json:"size,omitempty" yaml:"size,omitempty"`
}

// Shared command plumbing keeps validation and dry-run ahead of authentication.
func newRecoveryResizeCommand(kind, action string) *cobra.Command {
	command := &cobra.Command{Use: action + " <name-or-uuid>", Short: action + " a " + kind, Args: cobra.ExactArgs(1)}
	if action == "resize" {
		command.Flags().String("size", "", "New total volume size, for example 20G (growth only)")
		_ = command.MarkFlagRequired("size")
	}
	command.RunE = func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(args[0]) == "" {
			return fmt.Errorf("resource name or UUID must not be empty")
		}
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "", "text", "table", "json", "yaml", "name":
		default:
			return fmt.Errorf("unsupported output format %q", output)
		}
		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(fields) > 0 {
			var schema any = view.VolumeInventoryView{}
			if kind == "instance" {
				schema = view.VmInstanceInventoryView{}
			}
			if _, err := utils.ProjectObject(schema, fields); err != nil {
				return err
			}
		}
		var size int64
		if action == "resize" {
			raw, _ := cmd.Flags().GetString("size")
			var err error
			size, err = utils.ParseMemorySize(raw)
			if err != nil {
				return fmt.Errorf("invalid --size: %w", err)
			}
			if size <= 0 {
				return fmt.Errorf("--size must be positive")
			}
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			if output != "json" {
				output = "yaml"
			}
			return utils.Print(resourceActionPreview{ResourceType: kind, Resource: args[0], Action: action, Size: size}, utils.ParseFormat(output))
		}
		cli, err := client.GetClientContext(cmd.Context())
		if err != nil {
			return err
		}
		resource, err := client.ResolveOneResource(cli, kind, args[0])
		if err != nil {
			return err
		}
		var result any
		if kind == "instance" {
			if resource.State != "Destroyed" {
				return fmt.Errorf("instance must be Destroyed to recover, got %q", resource.State)
			}
			result, err = cli.RecoverVmInstance(resource.UUID, param.RecoverVmInstanceParam{})
		} else if action == "recover" {
			result, err = volumes.Recover(cli, cli, resource.UUID)
		} else {
			result, err = volumes.Resize(cli, cli, resource.UUID, size)
		}
		if err != nil {
			return err
		}
		verb := "recovered"
		if action == "resize" {
			verb = "resized"
		}
		if output == "name" {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s/%s\n", kind, resource.UUID)
			return err
		}
		if len(fields) > 0 && (output == "json" || output == "yaml" || output == "table") {
			projected, err := utils.ProjectObject(result, fields)
			if err != nil {
				return err
			}
			return utils.Print(projected, utils.ParseFormat(output))
		}
		return printVolumeResult(cmd, fmt.Sprintf("%s/%s %s", kind, resource.Name, verb), result)
	}
	return command
}

func init() {
	InstanceCmd.AddCommand(newRecoveryResizeCommand("instance", "recover"))
	VolumeCmd.AddCommand(newRecoveryResizeCommand("volume", "recover"), newRecoveryResizeCommand("volume", "resize"))
}
