// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"fmt"
	"os"

	"github.com/chijiajian/zstack-cli-go/cmd/create"
	"github.com/chijiajian/zstack-cli-go/cmd/del"
	"github.com/chijiajian/zstack-cli-go/cmd/expunge"
	"github.com/chijiajian/zstack-cli-go/cmd/get"
	"github.com/chijiajian/zstack-cli-go/cmd/resources"
	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	outputFlags struct {
		Format string
	}
	version = "dev"
	commit  = "none"
)

var rootCmd = &cobra.Command{
	Use:   "zstack-cli",
	Short: "ZStack CLI - manage your ZStack resources",
	Long:  `zstack-cli is a command-line interface for managing ZStack resources.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("zstack-cli version: %s, commit: %s\n", version, commit)
			os.Exit(0)
		}
		_, err := config.LoadConfig()
		return err
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddCommand(get.GetCmd)

	rootCmd.PersistentFlags().StringVarP(&outputFlags.Format, "output", "o", "table", "Output format (table|json|yaml|text)")
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(del.DeleteCmd)
	rootCmd.AddCommand(expunge.ExpungeCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(resources.InstanceCmd)
	//rootCmd.AddCommand(cmdutil.ResourceCommand)

	rootCmd.ValidArgs = []string{"create", "delete", "expunge", "get", "login", "config"}
	rootCmd.Args = cobra.OnlyValidArgs

	create.CreateCmd.ValidArgs = []string{"disk-offering", "instance-offering", "image", "instance"}
	create.CreateCmd.Args = cobra.OnlyValidArgs

	del.DeleteCmd.ValidArgs = []string{"images", "instances"}
	del.DeleteCmd.Args = cobra.OnlyValidArgs

	expunge.ExpungeCmd.ValidArgs = []string{"images", "instances"}
	expunge.ExpungeCmd.Args = cobra.OnlyValidArgs

	get.GetCmd.ValidArgs = []string{
		"clusters", "disk-offerings", "disks", "eips", "hosts",
		"images", "image-storages", "instance-offerings", "instances",
		"l2-networks", "l3-networks", "management-nodes",
		"primary-storages", "vips", "virtual-router-offerings",
		"virtual-routers", "vm-scripts", "zones",
	}

	get.GetCmd.Args = cobra.OnlyValidArgs

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show zstack-cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("zstack-cli version: %s, commit: %s\n", version, commit)
		},
	})

}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:
  source <(zstack-cli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  zstack-cli completion bash > /etc/bash_completion.d/zstack-cli
  # macOS:
  zstack-cli completion bash > /usr/local/etc/bash_completion.d/zstack-cli

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute once:
  echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  zstack-cli completion zsh > "${fpath[1]}/_zstack-cli"

Fish:
  zstack-cli completion fish | source
  zstack-cli completion fish > ~/.config/fish/completions/zstack-cli.fish

PowerShell:
  zstack-cli completion powershell | Out-String | Invoke-Expression
  zstack-cli completion powershell > zstack-cli.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

func FormatOutput(data interface{}) error {
	format := utils.ParseFormat(outputFlags.Format)
	return utils.Print(data, format)
}
