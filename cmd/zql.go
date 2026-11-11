package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use: "zql <statement>", Short: "Execute a ZStack Query Language statement and print the JSON response",
		Example: `  zstack-cli zql "query vmInstance.uuid where name='web-01' limit 10" --context lab`,
		Args:    cobra.ExactArgs(1), RunE: runZQL,
	}
	cmd.Flags().StringP("output", "o", "json", "Output format: json (preserves the server's results and returnWith data)")
	cmd.Flags().Duration("request-timeout", 2*time.Minute, "Deadline for authentication and the ZQL request")
	rootCmd.AddCommand(cmd)
}

func runZQL(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("output")
	timeout, _ := cmd.Flags().GetDuration("request-timeout")
	if format != "json" || strings.TrimSpace(args[0]) == "" || timeout <= 0 {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("zql requires a nonempty statement, JSON output and a positive request-timeout"))
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()
	cli, err := client.NewReadClient(ctx)
	if err != nil {
		return err
	}
	reply, err := cli.ZQL(ctx, args[0])
	if err != nil {
		return fmt.Errorf("execute ZQL: %w", err)
	}
	var envelope struct {
		Results []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(reply, &envelope); err != nil || envelope.Results == nil {
		return clierr.Wrap(clierr.Remote, fmt.Errorf("ZQL response missing results array"))
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return clierr.Wrap(clierr.Output, encoder.Encode(reply))
}
