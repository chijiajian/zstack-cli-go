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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"os"

	configcmd "github.com/chijiajian/zstack-cli-go/cmd/config"
	"github.com/chijiajian/zstack-cli-go/cmd/create"
	"github.com/chijiajian/zstack-cli-go/cmd/del"
	"github.com/chijiajian/zstack-cli-go/cmd/describe"
	"github.com/chijiajian/zstack-cli-go/cmd/expunge"
	"github.com/chijiajian/zstack-cli-go/cmd/get"
	"github.com/chijiajian/zstack-cli-go/cmd/iam"
	"github.com/chijiajian/zstack-cli-go/cmd/jobs"
	"github.com/chijiajian/zstack-cli-go/cmd/resources"
	"github.com/chijiajian/zstack-cli-go/cmd/update"
	"github.com/chijiajian/zstack-cli-go/cmd/wait"
	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
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

var errClientCommandFailed = errors.New("client command failed")

var rootCmd = &cobra.Command{
	Use:           "zstack-cli",
	Short:         "ZStack CLI - manage your ZStack resources",
	Long:          `zstack-cli is a command-line interface for managing ZStack resources.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if v, _ := cmd.Flags().GetBool("version"); v {
			printVersion()
			os.Exit(0)
		}
		zstackclient.ClearLastError()
		errorFormat, _ := cmd.Flags().GetString("error-format")
		if errorFormat != "text" && errorFormat != "json" {
			return clierr.Wrap(clierr.Usage, fmt.Errorf("unsupported error format %q", errorFormat))
		}
		if flag := cmd.Flags().Lookup("output"); flag != nil {
			format, _ := cmd.Flags().GetString("output")
			top := cmd
			for top.Parent() != nil && top.Parent().Parent() != nil {
				top = top.Parent()
			}
			valid := format == "" || format == "json" || format == "yaml" || format == "table" || format == "text" || format == "name" || format == "wide" && top.Name() == "create"
			if !valid {
				return clierr.Wrap(clierr.Usage, fmt.Errorf("unsupported output format %q", format))
			}
		}
		override, _ := cmd.Flags().GetString("context")
		config.SetContextOverride(override)
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		if override != "" {
			if cmd.Name() == "login" {
				return fmt.Errorf("login uses --context-name, not --context")
			}
			if _, err = config.GetCurrentContext(cfg); err != nil {
				return err
			}
		}
		config.PinContextSelection(cfg)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if zstackclient.LastError() != nil {
			return errClientCommandFailed
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			printVersion()
			return
		}
		cmd.Help()
	},
}

// prepareErrors marks argument/flag failures at their boundary, before execution.
func prepareErrors(command *cobra.Command) {
	command.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return clierr.Wrap(clierr.Usage, err) })
	validate := command.Args
	command.Args = func(cmd *cobra.Command, args []string) error {
		if validate != nil {
			if err := validate(cmd, args); err != nil {
				return clierr.Wrap(clierr.Usage, err)
			}
		}
		if err := cmd.ValidateRequiredFlags(); err != nil {
			return clierr.Wrap(clierr.Usage, err)
		}
		return clierr.Wrap(clierr.Usage, cmd.ValidateFlagGroups())
	}

	for _, child := range command.Commands() {
		prepareErrors(child)
	}
}

func Execute() {
	prepareErrors(rootCmd)
	_, _, resolveErr := rootCmd.Find(os.Args[1:])
	err := rootCmd.Execute()
	if err == nil {
		return
	}
	if errors.Is(err, errClientCommandFailed) {
		if cause := zstackclient.LastError(); cause != nil {
			err = cause
		}
	}
	if resolveErr != nil {
		err = clierr.Wrap(clierr.Usage, err)
	}
	err = zstackclient.NormalizeAPIError(err)
	diagnostic := clierr.Describe(err)
	format, _ := rootCmd.PersistentFlags().GetString("error-format")
	if format == "json" {
		_ = json.NewEncoder(os.Stderr).Encode(struct {
			Error clierr.Diagnostic `json:"error"`
		}{diagnostic})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
	os.Exit(diagnostic.ExitCode)
}

func init() {
	// Context selection must also run for commands with their own persistent hooks.
	cobra.EnableTraverseRunHooks = true

	rootCmd.AddCommand(get.GetCmd)

	rootCmd.PersistentFlags().String("error-format", "text", "Error diagnostics on stderr: text or json (stable code and exitCode)")
	rootCmd.PersistentFlags().String("context", "", "Use a saved context for this command without changing the default")
	rootCmd.PersistentFlags().Bool("version", false, "Show zstack-cli version")
	rootCmd.PersistentFlags().StringVarP(&outputFlags.Format, "output", "o", "table", "Output format (table|json|yaml|text)")
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(configcmd.ConfigCmd)
	rootCmd.AddCommand(del.DeleteCmd)
	rootCmd.AddCommand(describe.DescribeCmd)
	rootCmd.AddCommand(expunge.ExpungeCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(resources.BackupStorageCmd)
	rootCmd.AddCommand(resources.EipCmd)
	rootCmd.AddCommand(resources.GlobalConfigCmd)
	rootCmd.AddCommand(resources.HostCmd)
	rootCmd.AddCommand(resources.ImageCmd)
	rootCmd.AddCommand(resources.ImageGroupCmd)
	rootCmd.AddCommand(resources.InstanceCmd)
	rootCmd.AddCommand(resources.LoadBalancerCmd)
	rootCmd.AddCommand(resources.L2NetworkCmd)
	rootCmd.AddCommand(resources.L3NetworkCmd)
	rootCmd.AddCommand(resources.PortForwardingCmd)
	rootCmd.AddCommand(resources.PrimaryStorageCmd)
	rootCmd.AddCommand(resources.SecurityGroupCmd)
	rootCmd.AddCommand(resources.SnapshotCmd)
	rootCmd.AddCommand(resources.VmScriptCmd)
	rootCmd.AddCommand(resources.VolumeCmd)
	rootCmd.AddCommand(update.UpdateCmd)
	rootCmd.AddCommand(wait.WaitCmd)
	//rootCmd.AddCommand(cmdutil.ResourceCommand)

	rootCmd.ValidArgs = []string{"backup-storage", "create", "delete", "describe", "doctor", "eip", "expunge", "get", "global-config", "host", "image", "image-group", "instance", "l2-network", "l3-network", "load-balancer", "login", "logout", "config", "port-forwarding", "primary-storage", "security-group", "snapshot", "update", "vm-script", "volume", "wait"}
	rootCmd.Args = cobra.OnlyValidArgs

	create.CreateCmd.ValidArgs = []string{"alarm", "cluster", "disk-offering", "eip", "instance-offering", "image", "instance", "l2-network", "l3-network", "load-balancer", "lb", "load-balancer-listener", "lb-listener", "load-balancer-server-group", "lb-server-group", "port-forwarding", "pf", "security-group", "sg", "vip", "vm-script", "volume", "zone"}
	create.CreateCmd.Args = cobra.OnlyValidArgs

	describe.DescribeCmd.ValidArgs = []string{"instance", "volume", "long-job", "job"}
	describe.DescribeCmd.Args = cobra.OnlyValidArgs

	del.DeleteCmd.ValidArgs = []string{"alarm", "alarms", "backup-storage", "backup-storages", "bs", "cluster", "clusters", "eip", "eips", "host", "hosts", "image", "images", "image-storage", "image-storages", "instance", "instances", "ip-range", "ip-ranges", "l2-network", "l2-networks", "l3-network", "l3-networks", "load-balancer", "load-balancers", "lb", "load-balancer-listener", "load-balancer-listeners", "lb-listener", "load-balancer-server-group", "load-balancer-server-groups", "lb-server-group", "port-forwarding", "port-forwarding-rules", "pf", "primary-storage", "primary-storages", "ps", "security-group", "security-groups", "sg", "snapshot", "snapshots", "vip", "vips", "vm-script", "vm-scripts", "volume", "volumes", "zone", "zones"}
	del.DeleteCmd.Args = cobra.OnlyValidArgs

	expunge.ExpungeCmd.ValidArgs = []string{"image", "images", "image-group", "image-groups", "instance", "instances", "volume", "volumes"}
	expunge.ExpungeCmd.Args = cobra.OnlyValidArgs

	get.GetCmd.ValidArgs = []string{
		"volumes", "security-groups", "security-group-rules", "port-forwarding-rules", "load-balancers", "load-balancer-listeners", "load-balancer-server-groups",
		"accounts", "alarms", "audits", "cdroms", "clusters", "disk-offerings", "disks", "eips", "events",
		"global-config", "global-configs", "hosts", "images", "image-groups", "image-storages",
		"instance-offerings", "instances", "ip-ranges", "l2-networks", "l3-networks", "long-jobs", "jobs",
		"management-nodes", "nics", "organizations", "orgs", "iam2-organizations", "policies", "primary-storages",
		"projects", "iam2-projects", "roles", "snapshots", "tags", "users", "vips", "virtual-router-offerings",
		"virtual-routers", "vm-scripts", "vm-script-records", "zones",
	}

	get.GetCmd.Args = cobra.OnlyValidArgs

	update.UpdateCmd.ValidArgs = []string{
		"backup-storage", "bs", "cluster", "disk-offering", "eip", "host",
		"image", "image-storage", "instance", "instance-offering", "l2-network",
		"l3-network", "primary-storage", "ps", "security-group", "sg", "vip",
		"vm-script", "volume", "zone",
	}
	update.UpdateCmd.Args = cobra.OnlyValidArgs

	wait.WaitCmd.ValidArgs = []string{"image", "instance", "job", "long-job", "volume"}
	wait.WaitCmd.Args = cobra.OnlyValidArgs
	iam.Register(create.CreateCmd, update.UpdateCmd, del.DeleteCmd)
	iam.RegisterExpunge(expunge.ExpungeCmd)
	rootCmd.AddCommand(iam.ProjectCmd, jobs.Cmd)
	rootCmd.ValidArgs = append(rootCmd.ValidArgs, "project", "job")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show zstack-cli version",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
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

func printVersion() {
	fmt.Printf("zstack-cli version: %s, commit: %s\n", version, commit)
}
