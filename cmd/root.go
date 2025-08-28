/*
Copyright © 2025 zstack.io
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chijiajian/zstack-cli-go/cmd/create"
	"github.com/chijiajian/zstack-cli-go/cmd/get"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	outputFlags struct {
		Format string
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zstack-cli",
	Short: "A CLI for interacting with the ZStack API",
	Long:  `zstack-cli is a command-line interface for managing ZStack resources.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 在这里调用 cobra.OnInitialize 来注册我们的配置初始化函数
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(get.GetCmd)

	// 这里可以定义全局标志，比如 --config
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zstack-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFlags.Format, "output", "o", "table", "Output format (table|json|yaml|text)")
	rootCmd.AddCommand(create.CreateCmd)
}

// initConfig 会在 cobra 初始化时被调用
// 这是解决问题的关键！
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".zstack-cli" (without extension).
		configDir := filepath.Join(home, ".zstack-cli")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// 重点：尝试读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Debug: Using config file:", viper.ConfigFileUsed())
	} else {
		// 如果文件不存在，这并非一个错误，因为 login 命令会创建它
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 如果是其他读取错误，打印出来
			fmt.Println("Debug: Error reading config file:", err)
		} else {
			fmt.Println("Debug: No config file found. Login might be required.")
		}
	}
}

func FormatOutput(data interface{}) error {
	format := output.ParseFormat(outputFlags.Format)
	return output.Print(data, format)
}
