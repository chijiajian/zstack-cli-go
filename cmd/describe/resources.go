package describe

import (
	"context"
	"fmt"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/inspect"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newResourceDetail(kind string) *cobra.Command {
	cmd := &cobra.Command{Use: kind + " <name-or-uuid>", Short: "Describe " + kind + " and related resources", Args: cobra.ExactArgs(1)}
	cmd.Flags().Duration("request-timeout", 30*time.Second, "Deadline for detail and related resource queries")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return fmt.Errorf("request-timeout must be positive")
		}
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "table", "text", "json", "yaml":
		default:
			return fmt.Errorf("unsupported output format %q", output)
		}
		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(fields) > 0 {
			return fmt.Errorf("describe %s returns a complete diagnostic document; --fields is supported by get", kind)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		reader, err := client.NewReadClient(ctx)
		if err != nil {
			return err
		}
		resource, err := reader.Resolve(ctx, kind, args[0], client.ResolveAuto)
		if err != nil {
			return err
		}
		doc, err := (inspect.Service{Reader: reader}).Describe(ctx, kind, resource.UUID)
		if err != nil {
			return err
		}
		// Related-resource failures remain visible in machine output and on stderr.
		for _, warning := range doc.Warnings {
			if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "Warning:", warning); err != nil {
				return fmt.Errorf("write detail warning: %w", err)
			}
		}
		if output == "json" {
			return utils.Print(doc, utils.JSONFormat)
		}
		data, err := yaml.Marshal(doc)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(data))
		return err
	}
	return cmd
}

func init() { DescribeCmd.AddCommand(newResourceDetail("instance"), newResourceDetail("volume")) }
