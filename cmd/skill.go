package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/skillinstall"
	"github.com/spf13/cobra"
)

func init() {
	group := &cobra.Command{Use: "skill", Short: "Install the bundled ZStack operations Skill"}
	install := &cobra.Command{Use: "install", Short: "Install for Codex, Claude Code, Cursor or Copilot", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		get := func(name string) string { v, _ := cmd.Flags().GetString(name); return v }
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		target, err := skillinstall.Target(get("agent"), get("scope"), get("project"), get("dest"), home)
		if err != nil {
			return clierr.Wrap(clierr.Usage, err)
		}
		format := get("output")
		if format != "json" && format != "text" && format != "table" {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("skill install supports text, table or json output"))
		}
		replace, _ := cmd.Flags().GetBool("replace")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		result, err := skillinstall.Install(target, replace, dryRun)
		if err != nil {
			return err
		}
		if format == "json" {
			err = json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		} else {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Skill directory: %s\n", result.Destination)
			if err == nil && result.Backup != "" {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Previous installation: %s\n", result.Backup)
			}
			if err == nil && result.DryRun {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "Preview only.")
			}
		}
		return clierr.Wrap(clierr.Output, err)
	}}
	install.Flags().String("agent", "", "Target agent: codex, claude-code, cursor or copilot")
	install.MarkFlagRequired("agent")
	install.Flags().String("scope", "user", "Installation scope: user or project")
	install.Flags().String("project", "", "Existing project directory for project scope")
	install.Flags().String("dest", "", "Override the parent Skill directory")
	install.Flags().Bool("replace", false, "Update existing installation and preserve a backup")
	install.Flags().Bool("dry-run", false, "Preview destination without installing")
	group.AddCommand(install)
	rootCmd.AddCommand(group)
}
