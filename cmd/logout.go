package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use: "logout", Short: "Revoke the saved session and clear credentials for the selected context", Args: cobra.NoArgs,
	Long: "Log out the selected context without logging in or renewing its session. Access keys are only removed locally, not revoked on the server. Use --local to clear local credentials when the endpoint is unavailable.",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return fmt.Errorf("request-timeout must be positive")
		}
		local, _ := cmd.Flags().GetBool("local")
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		if err := client.LogoutContext(ctx, local); err != nil {
			return err
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Selected context credentials cleared.")
		return err
	},
}

func init() {
	logoutCmd.Flags().Bool("local", false, "Clear local credentials without revoking the remote session")
	logoutCmd.Flags().Duration("request-timeout", 30*time.Second, "Deadline for session revocation and configuration update")
	rootCmd.AddCommand(logoutCmd)
}
