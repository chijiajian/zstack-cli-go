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
	"context"
	"fmt"
	"sync"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	zsclient "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
)

var clientMutex sync.Mutex
var globalClient *zsclient.ZSClient
var lastClientError error

// GetClient preserves the legacy command API. New callers should propagate the error.
func GetClient() *zsclient.ZSClient {
	cli, _ := GetClientContext(context.Background())
	return cli
}

var clientContext config.Context

// GetClientContext includes session validation and login in the caller's deadline.
func GetClientContext(requestCtx context.Context) (*zsclient.ZSClient, error) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	cli, err := loadClient(requestCtx)
	lastClientError = err
	return cli, err
}

func loadClient(requestCtx context.Context) (*zsclient.ZSClient, error) {
	if err := requestCtx.Err(); err != nil {
		return nil, err
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	selected, err := config.GetCurrentContext(cfg)
	if err != nil {
		return nil, err
	}
	ctx := *selected
	if ctx.Endpoint == "" {
		return nil, fmt.Errorf("current context or endpoint missing. Please run 'zstack-cli login' first")
	}
	if globalClient != nil && clientContext == ctx {
		return globalClient, nil
	}
	globalClient = nil
	zsCfg, err := NewSDKConfig(ctx.Endpoint)
	if err != nil {
		return nil, err
	}
	authenticated, err := newAuthenticatedHTTPClient(requestCtx, cfg)
	if err != nil {
		return nil, err
	}
	ctx = authenticated.credentials
	var cli *zsclient.ZSClient
	if ctx.EffectiveAuthType() == config.AuthTypeAccessKey {
		cli = zsclient.NewZSClient(zsCfg.AccessKey(ctx.AccessKeyID, ctx.AccessKeySecret).ReadOnly(false))
	} else {
		cli = zsclient.NewZSClient(zsCfg.LoginAccount(ctx.Username, ctx.Password).ReadOnly(false))
		cli.LoadSession(ctx.SessionUUID)
	}

	globalClient, clientContext = cli, ctx
	return cli, nil
}

func ClearLastError() {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	lastClientError = nil
}

func LastError() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	return lastClientError
}

func ResetClient() {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	globalClient = nil
}

// Logout retains the existing library entry point without implicitly logging in.
func Logout() error { return LogoutContext(context.Background(), false) }

func GetSessionUUID() string {
	cfg, err := config.LoadConfig()
	if err != nil {
		return ""
	}
	selected, err := config.GetCurrentContext(cfg)
	if err != nil {
		return ""
	}
	return selected.SessionUUID
}

func IsLoggedIn() bool {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}
	selected, err := config.GetCurrentContext(cfg)
	if err != nil || selected.Endpoint == "" {
		return false
	}
	if selected.EffectiveAuthType() == config.AuthTypeAccessKey {
		return selected.AccessKeyID != "" && selected.AccessKeySecret != ""
	}
	return selected.SessionUUID != ""
}
