package resources

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

func printResourceResult(cmd *cobra.Command, text string, result interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	fields, _ := cmd.Flags().GetStringSlice("fields")

	switch output {
	case "", "text", "name":
		_, err := fmt.Fprintln(cmd.OutOrStdout(), text)
		return err
	default:
		return utils.PrintWithFields(result, utils.ParseFormat(output), fields)
	}
}
