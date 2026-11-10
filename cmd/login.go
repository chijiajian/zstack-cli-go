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
	"context"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	cliclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login [endpoint]",
	Short: "Login to ZStack API server",
	Long: `Login to ZStack API server and save the session for future commands.
Example:
  zstack-cli login 10.255.255.23 --username admin --password password
  zstack-cli login 10.255.255.23 --access-key-id AK --access-key-secret SK`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var endpoint string
		if len(args) > 0 {
			endpoint = args[0]
		} else {
			endpoint = viper.GetString("endpoint")
			if endpoint == "" {
				return fmt.Errorf("no endpoint specified")
			}
		}

		contextName, _ := cmd.Flags().GetString("context-name")
		if contextName == "" {
			contextName = endpoint
		}
		if err := config.ValidateContextName(contextName); err != nil {
			return err
		}
		if _, err := cliclient.NewSDKConfig(endpoint); err != nil {
			return err
		}
		noSwitch, _ := cmd.Flags().GetBool("no-switch")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		savePassword, _ := cmd.Flags().GetBool("save-password")
		accessKeyID, _ := cmd.Flags().GetString("access-key-id")
		accessKeySecret, _ := cmd.Flags().GetString("access-key-secret")
		accessKeyIDAlias, _ := cmd.Flags().GetString("ak")
		accessKeySecretAlias, _ := cmd.Flags().GetString("sk")
		accessKeySpecified := accessKeyID != "" || accessKeySecret != "" || accessKeyIDAlias != "" || accessKeySecretAlias != ""
		passwordStdin, _ := cmd.Flags().GetBool("password-stdin")
		accessKeySecretStdin, _ := cmd.Flags().GetBool("access-key-secret-stdin")
		if accessKeySecretStdin {
			for _, flag := range []string{"username", "password", "password-stdin", "access-key-secret", "sk"} {
				if cmd.Flags().Changed(flag) {
					return clierr.Wrap(clierr.Usage, fmt.Errorf("--access-key-secret-stdin cannot be combined with --%s", flag))
				}
			}
			if accessKeyID == "" {
				accessKeyID = accessKeyIDAlias
			}
			if accessKeyID == "" {
				return clierr.Wrap(clierr.Usage, fmt.Errorf("--access-key-secret-stdin requires --access-key-id"))
			}
			secret, err := readLoginSecret(cmd.InOrStdin(), "access-key-secret-stdin")
			if err != nil {
				return err
			}
			return saveAccessKeyContext(cmd.Context(), endpoint, accessKeyID, secret, contextName, !noSwitch)
		}
		if passwordStdin {
			for _, flag := range []string{"password", "access-key-id", "access-key-secret", "ak", "sk"} {
				if cmd.Flags().Changed(flag) {
					return clierr.Wrap(clierr.Usage, fmt.Errorf("--password-stdin cannot be combined with --%s", flag))
				}
			}
			if username == "" {
				return clierr.Wrap(clierr.Usage, fmt.Errorf("--password-stdin requires --username"))
			}
			if cmd.Flags().Changed("save-password") && savePassword {
				return clierr.Wrap(clierr.Usage, fmt.Errorf("--password-stdin cannot save the password"))
			}
			savePassword = false
			var err error
			password, err = readLoginSecret(cmd.InOrStdin(), "password-stdin")
			if err != nil {
				return err
			}
		}
		if accessKeyID == "" {
			accessKeyID = accessKeyIDAlias
		}
		if accessKeySecret == "" {
			accessKeySecret = accessKeySecretAlias
		}
		if !accessKeySpecified && username == "" && password == "" && accessKeyID == "" {
			accessKeyID = os.Getenv("ZSTACK_ACCESS_KEY_ID")
		}
		if !accessKeySpecified && username == "" && password == "" && accessKeySecret == "" {
			accessKeySecret = os.Getenv("ZSTACK_ACCESS_KEY_SECRET")
		}

		if accessKeyID != "" || accessKeySecret != "" {
			if accessKeyID == "" || accessKeySecret == "" {
				return fmt.Errorf("both --access-key-id and --access-key-secret are required for access key authentication")
			}
			if username != "" || password != "" {
				return fmt.Errorf("use either username/password or access key credentials, not both")
			}
			return saveAccessKeyContext(cmd.Context(), endpoint, accessKeyID, accessKeySecret, contextName, !noSwitch)
		}

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
					return fmt.Errorf("reading password: %w", err)
				}
				password = string(bytePassword)
			}
		}

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required")
		}

		fmt.Printf("Logging in to ZStack API server: %s\n", endpoint)
		sessionInfo, err := cliclient.LoginAccountContext(cmd.Context(), endpoint, username, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		ctxName := contextName

		ctx := config.Context{
			Endpoint:    endpoint,
			AuthType:    config.AuthTypeAccount,
			Username:    username,
			Password:    "",
			SessionUUID: sessionInfo.Uuid,
		}

		if savePassword {
			ctx.Password = password
		}

		if err := config.UpdateConfig(cmd.Context(), func(cfg *config.ZStackConfig) error {
			return config.StoreContext(cfg, ctxName, ctx, !noSwitch)
		}); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Login successful!")
		fmt.Printf("Account: %s\n", sessionInfo.AccountUuid)
		fmt.Printf("User: %s\n", sessionInfo.UserUuid)
		if !passwordStdin {
			fmt.Printf("Session UUID: %s\n", sessionInfo.Uuid)
		}
		fmt.Println("Session saved. You can now use other commands.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("context-name", "", "Name for saved context (default: endpoint)")
	loginCmd.Flags().Bool("no-switch", false, "Save credentials without changing the default context")

	loginCmd.Flags().StringP("username", "u", "", "Username for authentication")
	loginCmd.Flags().StringP("password", "p", "", "Password for authentication")
	loginCmd.Flags().Bool("password-stdin", false, "Read a UTF-8 password from stdin until EOF (requires --username, never saves the password)")
	loginCmd.Flags().String("access-key-id", "", "Access key ID for request signing")
	loginCmd.Flags().String("access-key-secret", "", "Access key secret for request signing")
	loginCmd.Flags().Bool("access-key-secret-stdin", false, "Read a UTF-8 access key secret from stdin until EOF (requires --access-key-id)")
	loginCmd.Flags().String("ak", "", "Alias for --access-key-id")
	loginCmd.Flags().String("sk", "", "Alias for --access-key-secret")
	loginCmd.Flags().Bool("save-password", true, "Save password in config file (default: true)")
}

func readLoginSecret(input io.Reader, flag string) (string, error) {
	const maxSecretBytes = 64 * 1024
	data, err := io.ReadAll(io.LimitReader(input, maxSecretBytes+1))
	defer clear(data)
	if err != nil || len(data) == 0 || len(data) > maxSecretBytes || !utf8.Valid(data) {
		return "", clierr.Wrap(clierr.Usage, fmt.Errorf("--%s requires 1 to 65536 bytes of valid UTF-8 input", flag))
	}
	return string(data), nil
}

func saveAccessKeyContext(requestCtx context.Context, endpoint, accessKeyID, accessKeySecret, contextName string, activate bool) error {

	ctx := config.Context{
		Endpoint:        endpoint,
		AuthType:        config.AuthTypeAccessKey,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	}

	if err := config.UpdateConfig(requestCtx, func(cfg *config.ZStackConfig) error {
		return config.StoreContext(cfg, contextName, ctx, activate)
	}); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Saved access key authentication for ZStack API server: %s\n", endpoint)
	fmt.Println("Requests will be signed with the configured access key.")
	return nil
}
