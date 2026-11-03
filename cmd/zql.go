package cmd

import (
	"fmt"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/util/jsonutils"
)

func init() {
	cmd := &cobra.Command{
		Use: "zql <statement>", Short: "Execute a ZStack Query Language statement and print the JSON response",
		Example: `  zstack-cli zql "query vmInstance.uuid where name='web-01' limit 10" --context lab`,
		Args:    cobra.ExactArgs(1), RunE: runZQL,
	}
	cmd.Flags().StringP("output", "o", "json", "Output format: json (preserves the server's results and returnWith data)")
	rootCmd.AddCommand(cmd)
}

func runZQL(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("output")
	if format != "json" || strings.TrimSpace(args[0]) == "" {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("zql requires a nonempty statement and JSON output"))
	}
	cli, err := client.GetClientContext(cmd.Context())
	if err != nil {
		return err
	}
	var reply *jsonutils.JSONDict
	if _, err := cli.Zql(cmd.Context(), args[0], &reply); err != nil {
		return fmt.Errorf("execute ZQL: %w", err)
	}
	if reply == nil {
		return clierr.Wrap(clierr.Remote, fmt.Errorf("empty ZQL response"))
	}
	results, _ := reply.Get("results")
	if _, ok := results.(*jsonutils.JSONArray); !ok {
		return clierr.Wrap(clierr.Remote, fmt.Errorf("ZQL response missing results array"))
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), reply.PrettyString())
	return clierr.Wrap(clierr.Output, err)
}
