// pkg/client/client.go
package client

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	zsclient "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
)

// 使用互斥锁确保线程安全
var clientMutex sync.Mutex
var globalClient *zsclient.ZSClient

// GetClient 创建并返回一个已认证的客户端实例
// 每次调用都会使用用户名和密码重新登录，确保会话有效
func GetClient() *zsclient.ZSClient {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// 从配置文件中读取必要信息
	endpoint := viper.GetString("endpoint")
	username := viper.GetString("username")
	password := viper.GetString("password")

	// 验证配置完整性
	if endpoint == "" {
		fmt.Println("Error: No endpoint configured. Please run 'zstack-cli login' first.")
		return nil
	}

	if username == "" || password == "" {
		fmt.Println("Error: Missing credentials. Please run 'zstack-cli login' first.")
		return nil
	}

	isDebug := viper.GetBool("debug")
	if isDebug {
		fmt.Printf("Debug: Creating client - endpoint=%s, username=%s\n", endpoint, username)
	}

	// 创建登录配置
	zsConfig := zsclient.DefaultZSConfig(endpoint).
		LoginAccount(username, password).
		Debug(isDebug).
		ReadOnly(false) // 设置为 false 以允许修改操作

	// 创建新的客户端实例
	zsClient := zsclient.NewZSClient(zsConfig)

	// 尝试登录
	if isDebug {
		fmt.Println("Debug: Logging in to ZStack API...")
	}

	sessionInfo, err := zsClient.Login()
	if err != nil {
		fmt.Printf("Error: Login failed: %s\n", err)
		return nil
	}

	if isDebug {
		fmt.Printf("Debug: Login successful, session UUID: %s\n", sessionInfo.UUID)
	}

	// 更新会话 UUID
	viper.Set("session_uuid", sessionInfo.UUID)
	err = viper.WriteConfig()
	if err != nil && isDebug {
		fmt.Printf("Debug: Failed to save session: %s\n", err)
	}

	// 更新全局客户端实例
	globalClient = zsClient
	return zsClient
}

// ResetClient 重置全局客户端实例，强制下次调用 GetClient 时重新创建
func ResetClient() {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	globalClient = nil
}

// Logout 注销当前会话并清除凭据
func Logout() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if globalClient != nil {
		// 尝试注销当前会话
		err := globalClient.Logout()
		if err != nil {
			return fmt.Errorf("failed to logout: %s", err)
		}
		globalClient = nil
	}

	// 清除会话信息
	viper.Set("session_uuid", "")
	err := viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to update config: %s", err)
	}

	return nil
}

// GetSessionUUID 返回当前会话的 UUID
func GetSessionUUID() string {
	return viper.GetString("session_uuid")
}

// IsLoggedIn 检查用户是否已登录
func IsLoggedIn() bool {
	return viper.GetString("username") != "" &&
		viper.GetString("password") != "" &&
		viper.GetString("endpoint") != ""
}
