package resources

import (
	"github.com/spf13/cobra"
)

var InstanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage VM instances",
	Long:  "Perform operations on VM instances, such as start, stop, restart.",
}

func init() {
	InstanceCmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	InstanceCmd.PersistentFlags().BoolP("yes", "y", false, "Automatic yes to prompts")
	InstanceCmd.PersistentFlags().StringP("output", "o", "", "Output format: json, yaml, table, wide, or name")
	InstanceCmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")
}
