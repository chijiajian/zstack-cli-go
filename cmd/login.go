// cmd/login.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
	"golang.org/x/term"
)

// loginCmd 表示登录命令
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
			// 从命令行参数获取 endpoint
			endpoint = args[0]
		} else {
			// 从配置文件获取 endpoint
			endpoint = viper.GetString("endpoint")
			if endpoint == "" {
				fmt.Println("Error: No endpoint specified")
				return
			}
		}

		// 确保 endpoint 有 http:// 或 https:// 前缀
		if !hasProtocolPrefix(endpoint) {
			endpoint = endpoint
		}

		// 获取用户名和密码
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		savePassword, _ := cmd.Flags().GetBool("save-password")

		// 如果命令行没有提供用户名，尝试从配置文件获取或提示用户输入
		if username == "" {
			username = viper.GetString("username")
			if username == "" {
				fmt.Print("Username: ")
				fmt.Scanln(&username)
			}
		}

		// 如果命令行没有提供密码，尝试从配置文件获取或提示用户输入
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

		// 创建登录配置
		fmt.Printf("Logging in to ZStack API server: %s\n", endpoint)
		zsConfig := client.DefaultZSConfig(endpoint).
			LoginAccount(username, password).
			Debug(viper.GetBool("debug")).
			ReadOnly(false)

		zsClient := client.NewZSClient(zsConfig)

		// 尝试登录
		sessionInfo, err := zsClient.Login()
		if err != nil {
			fmt.Printf("Login failed: %s\n", err)
			return
		}

		// 登录成功，保存配置
		viper.Set("endpoint", endpoint)
		viper.Set("username", username)

		// 根据用户选择是否保存密码
		if savePassword {
			viper.Set("password", password)
		} else {
			// 如果用户选择不保存密码，但配置中已有密码，则清除它
			if viper.GetString("password") != "" {
				viper.Set("password", "")
			}
		}

		viper.Set("session_uuid", sessionInfo.UUID)

		// 确保配置目录存在
		configDir := filepath.Dir(viper.ConfigFileUsed())
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			fmt.Printf("Failed to create config directory: %s\n", err)
			return
		}

		// 保存配置
		err = viper.WriteConfig()
		if err != nil {
			// 如果配置文件不存在，尝试创建它
			if os.IsNotExist(err) {
				err = viper.SafeWriteConfig()
			}
			if err != nil {
				fmt.Printf("Failed to save configuration: %s\n", err)
				return
			}
		}

		fmt.Println("Login successful!")
		fmt.Printf("Account: %s\n", sessionInfo.AccountUuid)
		fmt.Printf("User: %s\n", sessionInfo.UserUuid)
		fmt.Printf("Session UUID: %s\n", sessionInfo.UUID)
		fmt.Println("Session saved. You can now use other commands.")
	},
}

// hasProtocolPrefix 检查 URL 是否有协议前缀
func hasProtocolPrefix(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// 添加命令行标志
	loginCmd.Flags().StringP("username", "u", "", "Username for authentication")
	loginCmd.Flags().StringP("password", "p", "", "Password for authentication")
	loginCmd.Flags().Bool("save-password", true, "Save password in config file (default: true)")
}
