package iam

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var ProjectCmd = &cobra.Command{Use: "project", Short: "Manage the lifecycle of IAM2 projects"}

func init() {
	ProjectCmd.PersistentFlags().Bool("dry-run", false, "Preview the request without contacting the server")
	ProjectCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml or name")
	ProjectCmd.AddCommand(projectAction("recover"))
}
func RegisterExpunge(parent *cobra.Command) {
	parent.AddCommand(projectAction("expunge"))
	parent.ValidArgs = append(parent.ValidArgs, "project")
}
func projectAction(action string) *cobra.Command {
	use := action + " <name-or-uuid>"
	if action == "expunge" {
		use = "project <name-or-uuid>"
	}
	cmd := &cobra.Command{Use: use, Short: action + " an IAM2 project", Args: cobra.ExactArgs(1)}
	cmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	cmd.Flags().Duration("request-timeout", 2*time.Minute, "Deadline for resolution and project action")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("request-timeout must be positive"))
		}
		if err := nonempty(args[0], "project"); err != nil {
			return clierr.Wrap(clierr.Usage, err)
		}
		key := "recoverIAM2Project"
		if action == "expunge" {
			key = "expungeIAM2Project"
		}
		body := map[string]any{key: map[string]any{}}
		preview, _ := cmd.Flags().GetBool("dry-run")
		format := text(cmd, "output")
		if preview {
			if format == "" || format == "text" {
				format = "yaml"
			}
			return utils.PrintDryRun(map[string]any{"project": args[0], "request": body}, format)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		cli, err := client.NewMutationClient(ctx)
		if err != nil {
			return err
		}
		project, err := cli.Reader().Resolve(ctx, "project", args[0], client.ResolveAuto)
		if err != nil {
			return err
		}
		yes, err := common.Confirm(cmd, fmt.Sprintf("%s project %s (%s)? Type 'yes' to confirm: ", action, project.Name, project.UUID))
		if err != nil {
			return err
		}
		if !yes {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Aborted by user.")
			return err
		}
		var result inventory
		var target any = &result
		if action == "expunge" {
			result = inventory{UUID: project.UUID, Name: project.Name}
			target = nil
		}
		if err = cli.Mutate(ctx, http.MethodPut, "v1/iam2/projects/"+url.PathEscape(project.UUID)+"/actions", body, target); err != nil {
			return err
		}
		if result.UUID != project.UUID {
			return clierr.Wrap(clierr.Remote, fmt.Errorf("project response identity mismatch"))
		}
		if format == "" || format == "text" || format == "name" {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "project/%s %s requested\n", project.UUID, action)
			return err
		}
		return utils.Print(result, utils.ParseFormat(format))
	}
	return cmd
}
