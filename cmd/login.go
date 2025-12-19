// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// cmd/login.go

package cmd

import (
	"fmt"
	"os"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login [endpoint]",
	Short: "Login to ZStack API server",
	Long: `Login to ZStack API server and save the session for future commands.
Example:
  zstack-cli login 192.168.1.100 --username admin --password password`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var endpoint string
		if len(args) > 0 {
			endpoint = args[0]
		} else {
			endpoint = viper.GetString("endpoint")
			if endpoint == "" {
				fmt.Println("Error: No endpoint specified")
				return
			}
		}

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		savePassword, _ := cmd.Flags().GetBool("save-password")

		if username == "" {
			username = viper.GetString("username")
			if username == "" {
				fmt.Print("Username: ")
				fmt.Scanln(&username)
			}
		}

		if password == "" {
			password = viper.GetString("password")
			if password == "" {
				fmt.Print("Password: ")
				bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println() // 换行
				if err != nil {
					fmt.Printf("Error reading password: %s\n", err)
					return
				}
				password = string(bytePassword)
			}
		}

		if username == "" || password == "" {
			fmt.Println("Error: Username and password are required")
			return
		}

		fmt.Printf("Logging in to ZStack API server: %s\n", endpoint)
		zsConfig := client.DefaultZSConfig(endpoint).
			LoginAccount(username, password).
			Debug(viper.GetBool("debug")).
			ReadOnly(false)

		zsClient := client.NewZSClient(zsConfig)

		sessionInfo, err := zsClient.Login()
		if err != nil {
			fmt.Printf("Login failed: %s\n", err)
			return
		}

		cfg, loadErr := config.LoadConfig()
		if loadErr != nil {
			fmt.Printf("Failed to load config: %s\n", loadErr)
			return
		}

		ctxName := endpoint

		ctx := config.Context{
			Endpoint:    endpoint,
			Username:    username,
			Password:    "",
			SessionUUID: sessionInfo.UUID,
		}

		if savePassword {
			ctx.Password = password
		}

		cfg.Contexts[ctxName] = ctx
		cfg.CurrentContext = ctxName

		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Failed to save config: %s\n", err)
			return
		}

		fmt.Println("Login successful!")
		fmt.Printf("Account: %s\n", sessionInfo.AccountUuid)
		fmt.Printf("User: %s\n", sessionInfo.UserUuid)
		fmt.Printf("Session UUID: %s\n", sessionInfo.UUID)
		fmt.Println("Session saved. You can now use other commands.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "Username for authentication")
	loginCmd.Flags().StringP("password", "p", "", "Password for authentication")
	loginCmd.Flags().Bool("save-password", true, "Save password in config file (default: true)")
}
