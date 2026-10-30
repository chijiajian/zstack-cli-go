package config

import (
	"fmt"
	"sort"

	core "github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{Use: "config", Short: "Manage CLI configuration and contexts"}

type contextRow struct {
	Name     string `json:"name" yaml:"name" header:"NAME"`
	Current  bool   `json:"current" yaml:"current" header:"CURRENT"`
	Selected bool   `json:"selected" yaml:"selected" header:"SELECTED"`
	Endpoint string `json:"endpoint" yaml:"endpoint" header:"ENDPOINT"`
	AuthType string `json:"authType" yaml:"authType" header:"AUTH TYPE"`
	Username string `json:"username,omitempty" yaml:"username,omitempty" header:"USERNAME"`
}

func printConfig(cmd *cobra.Command, data any) error {
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "table", "text", "json", "yaml":
	default:
		return fmt.Errorf("unsupported output format %q", output)
	}
	return utils.Print(data, utils.ParseFormat(output))
}
func init() {
	ConfigCmd.AddCommand(&cobra.Command{Use: "view", Short: "Show configuration with credentials redacted", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := core.LoadConfig()
		if err != nil {
			return err
		}
		// Copy before redaction: never modify stored credentials.
		contexts := map[string]map[string]string{}
		for name, c := range cfg.Contexts {
			contexts[name] = map[string]string{"endpoint": c.Endpoint, "auth_type": c.EffectiveAuthType(), "username": c.Username, "password": mask(c.Password), "access_key_id": mask(c.AccessKeyID), "access_key_secret": mask(c.AccessKeySecret), "session_uuid": mask(c.SessionUUID)}
		}
		output, _ := cmd.Flags().GetString("output")
		data := map[string]any{"current-context": cfg.CurrentContext, "selected-context": core.SelectedContextName(cfg), "contexts": contexts}
		if output == "table" || output == "text" {
			return utils.Print(data, utils.YAMLFormat)
		}
		return printConfig(cmd, data)
	}})
	ConfigCmd.AddCommand(&cobra.Command{Use: "get-contexts", Aliases: []string{"contexts"}, Short: "List saved contexts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := core.LoadConfig()
		if err != nil {
			return err
		}
		names := []string{}
		for name := range cfg.Contexts {
			names = append(names, name)
		}
		sort.Strings(names)
		rows := []contextRow{}
		for _, name := range names {
			c := cfg.Contexts[name]
			rows = append(rows, contextRow{name, name == cfg.CurrentContext, name == core.SelectedContextName(cfg), c.Endpoint, c.EffectiveAuthType(), c.Username})
		}
		return printConfig(cmd, rows)
	}})
	ConfigCmd.AddCommand(&cobra.Command{Use: "current-context", Short: "Print the effective context name", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := core.LoadConfig()
		if err != nil {
			return err
		}
		if _, err = core.GetCurrentContext(cfg); err != nil {
			return err
		}
		output, _ := cmd.Flags().GetString("output")
		if output == "table" || output == "text" {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), core.SelectedContextName(cfg))
			return err
		}
		return printConfig(cmd, core.SelectedContextName(cfg))
	}})
	addContextMutation("use-context <name>", 1, func(cfg *core.ZStackConfig, args []string) error { return core.UseContext(cfg, args[0]) })
	addContextMutation("rename-context <old-name> <new-name>", 2, func(cfg *core.ZStackConfig, args []string) error { return core.RenameContext(cfg, args[0], args[1]) })
	addContextMutation("delete-context <name>", 1, func(cfg *core.ZStackConfig, args []string) error { return core.DeleteContext(cfg, args[0]) })
}
func addContextMutation(use string, n int, mutate func(*core.ZStackConfig, []string) error) {
	ConfigCmd.AddCommand(&cobra.Command{Use: use, Args: cobra.ExactArgs(n), RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "table", "text", "json", "yaml":
		default:
			return fmt.Errorf("unsupported output format %q", output)
		}
		var currentContext string
		err := core.UpdateConfig(cmd.Context(), func(cfg *core.ZStackConfig) error {
			if err := mutate(cfg, args); err != nil {
				return err
			}
			currentContext = cfg.CurrentContext
			return nil
		})
		if err != nil {
			return err
		}

		if output == "json" || output == "yaml" {
			return printConfig(cmd, map[string]string{"action": cmd.Name(), "context": args[0], "currentContext": currentContext})
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", cmd.Name(), args[0])
		return err
	}})
}
func mask(value string) string {
	if value == "" {
		return ""
	}
	return "*****"
}
