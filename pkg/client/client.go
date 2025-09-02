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

// pkg/client/client.go
package client

import (
	"fmt"
	"sync"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	zsclient "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
)

var clientMutex sync.Mutex
var globalClient *zsclient.ZSClient

func GetClient() *zsclient.ZSClient {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if globalClient != nil {
		return globalClient
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error: failed to load config: %s\n", err)
		return nil
	}

	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		fmt.Println("Error: current context not found. Please run 'zstack-cli login' first.")
		return nil
	}

	if ctx.Endpoint == "" || ctx.Username == "" || ctx.Password == "" {
		fmt.Println("Error: endpoint, username or password missing in current context. Please run 'zstack-cli login' first.")
		return nil
	}

	zsCfg := zsclient.DefaultZSConfig(ctx.Endpoint).
		LoginAccount(ctx.Username, ctx.Password).
		ReadOnly(false)

	client := zsclient.NewZSClient(zsCfg)
	session, err := client.Login()
	if err != nil {
		fmt.Printf("Error: login failed: %s\n", err)
		return nil
	}

	ctx.SessionUUID = session.UUID
	cfg.Contexts[cfg.CurrentContext] = ctx
	_ = config.SaveConfig(cfg)

	globalClient = client
	return client

}

func ResetClient() {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	globalClient = nil
}

func Logout() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if globalClient != nil {

		err := globalClient.Logout()
		if err != nil {
			return fmt.Errorf("failed to logout: %s", err)
		}
		globalClient = nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if ok {
		ctx.SessionUUID = ""
		ctx.Password = ""
		cfg.Contexts[cfg.CurrentContext] = ctx
		return config.SaveConfig(cfg)
	}
	return nil

}

func GetSessionUUID() string {

	cfg, _ := config.LoadConfig()
	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if ok {
		return ctx.SessionUUID
	}
	return ""
}

func IsLoggedIn() bool {
	cfg, _ := config.LoadConfig()
	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	return ok && ctx.Endpoint != "" && ctx.SessionUUID != ""

}
