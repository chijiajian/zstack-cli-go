package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/chijiajian/zstack-cli-go/pkg/diagnostics"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use: "doctor", Short: "Check selected configuration and API access without changing credentials", Args: cobra.NoArgs,
	Long: "Check endpoint syntax, saved credentials, session validity and a bounded zone query. Never logs in, renews a session, or writes configuration. Does not certify permissions for all APIs. Use --offline for local checks only.",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("request-timeout must be positive"))
		}
		format, _ := cmd.Flags().GetString("output")
		if format != "json" && format != "yaml" && format != "text" && format != "table" {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("unsupported doctor output format %q", format))
		}
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		selected, err := config.GetCurrentContext(cfg)
		if err != nil {
			return clierr.Wrap(clierr.Auth, err)
		}
		reader, err := client.NewProbeClient(*selected)
		if err != nil {
			return clierr.Wrap(clierr.Usage, err)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		offline, _ := cmd.Flags().GetBool("offline")
		report, checkErr := diagnostics.Run(ctx, config.SelectedContextName(cfg), *selected, reader, offline)
		if format == "table" || format == "text" {
			for _, check := range report.Checks {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", check.Name, check.Status, check.Message); err != nil {
					return clierr.Wrap(clierr.Output, err)
				}
			}
		} else if err := utils.Print(report, utils.ParseFormat(format)); err != nil {
			return clierr.Wrap(clierr.Output, err)
		}
		return checkErr
	},
}

func init() {
	doctorCmd.Flags().Bool("offline", false, "Check local configuration only; do not contact the server")
	doctorCmd.Flags().Duration("request-timeout", 30*time.Second, "Deadline for all remote diagnostic checks")
	rootCmd.AddCommand(doctorCmd)
}
