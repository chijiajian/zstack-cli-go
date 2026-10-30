// Package jobs exposes explicit long-job actions, separate from read-only wait.
package jobs

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

var Cmd = &cobra.Command{Use: "job", Short: "Resume or clean long-running jobs"}

func init() {
	Cmd.PersistentFlags().Bool("dry-run", false, "Preview the request without contacting the server")
	Cmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, json, yaml or table")
	for _, action := range []string{"resume", "clean"} {
		Cmd.AddCommand(newAction(action))
	}
}
func newAction(action string) *cobra.Command {
	cmd := &cobra.Command{Use: action + " <name-or-uuid>", Short: action + " a long job", Args: cobra.ExactArgs(1)}
	cmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	cmd.Flags().Duration("request-timeout", 2*time.Minute, "Deadline for resolution and action completion")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("request-timeout must be positive"))
		}
		key := "resumeLongJob"
		if action == "clean" {
			key = "cleanLongJob"
		}
		body := map[string]any{key: map[string]any{}}
		format, _ := cmd.Flags().GetString("output")
		preview, _ := cmd.Flags().GetBool("dry-run")
		if preview {
			if format == "text" {
				format = "yaml"
			}
			return utils.PrintDryRun(map[string]any{"job": args[0], "request": body}, format)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		cli, err := client.NewMutationClient(ctx)
		if err != nil {
			return err
		}
		job, err := cli.Reader().Resolve(ctx, "long-job", args[0], client.ResolveAuto)
		if err != nil {
			return err
		}
		terminal := job.State == "Succeeded" || job.State == "Failed" || job.State == "Canceled"
		if action == "resume" && job.State != "Suspended" || action == "clean" && !terminal {
			return clierr.Wrap(clierr.Conflict, fmt.Errorf("cannot %s job in state %s", action, job.State))
		}
		yes, err := common.Confirm(cmd, fmt.Sprintf("%s job %s (%s)? Type 'yes' to confirm: ", action, job.Name, job.UUID))
		if err != nil {
			return err
		}
		if !yes {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Aborted by user.")
			return err
		}
		var result struct {
			UUID  string `json:"uuid" yaml:"uuid"`
			Name  string `json:"name" yaml:"name"`
			State string `json:"state" yaml:"state"`
		}
		if err = cli.Mutate(ctx, http.MethodPut, "v1/longjobs/"+url.PathEscape(job.UUID)+"/actions", body, &result); err != nil {
			return err
		}
		if result.UUID != job.UUID {
			return clierr.Wrap(clierr.Remote, fmt.Errorf("job response identity mismatch"))
		}
		if format == "text" || format == "name" {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "job/%s %s requested\n", job.UUID, action)
			return err
		}
		return utils.Print(result, utils.ParseFormat(format))
	}
	return cmd
}
