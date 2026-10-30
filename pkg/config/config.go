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

// pkg/cmd/config.go
package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Context struct {
	Endpoint        string `yaml:"endpoint"`
	AuthType        string `yaml:"auth_type,omitempty"`
	Username        string `yaml:"username,omitempty"`
	Password        string `yaml:"password,omitempty"`
	AccessKeyID     string `yaml:"access_key_id,omitempty"`
	AccessKeySecret string `yaml:"access_key_secret,omitempty"`
	SessionUUID     string `yaml:"session_uuid,omitempty"`
}

const (
	AuthTypeAccount   = "account"
	AuthTypeAccessKey = "access-key"
)

func (ctx Context) EffectiveAuthType() string {
	if ctx.AuthType != "" {
		return ctx.AuthType
	}
	if ctx.AccessKeyID != "" || ctx.AccessKeySecret != "" {
		return AuthTypeAccessKey
	}
	return AuthTypeAccount
}

type ZStackConfig struct {
	CurrentContext string             `yaml:"current-context"`
	Contexts       map[string]Context `yaml:"contexts"`
}

func getDefaultConfigFile() string {
	if envCfg := os.Getenv("ZSTACK_CONFIG"); envCfg != "" {
		return envCfg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zstack-cli", "config.yaml")
}

func LoadConfig() (*ZStackConfig, error) {
	return loadConfigFile(getDefaultConfigFile())
}

func loadConfigFile(file string) (*ZStackConfig, error) {
	cfg := &ZStackConfig{
		Contexts: make(map[string]Context),
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return cfg, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveConfig explicitly replaces a complete snapshot. Read-modify-write callers
// must use UpdateConfig so that the read happens inside the transaction.
func SaveConfig(cfg *ZStackConfig) error {
	return withConfigLock(context.Background(), func(file string) error { return saveConfigFile(file, cfg) })
}

func saveConfigFile(file string, cfg *ZStackConfig) error {
	dir := filepath.Dir(file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".config-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), file)
}

func GetCurrentContext(cfg *ZStackConfig) (*Context, error) {
	name := SelectedContextName(cfg)
	if name == "" {
		return nil, clierr.Wrap(clierr.Auth, errors.New("no current context, please login first"))
	}
	ctx, ok := cfg.Contexts[name]
	if !ok {
		return nil, clierr.Wrap(clierr.NotFound, fmt.Errorf("context %s not found", name))
	}
	if err := validatePinnedContext(name, ctx); err != nil {
		return nil, err
	}
	return &ctx, nil
}

func SetCurrentContext(cfg *ZStackConfig, name string, ctx Context) {
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]Context)
	}
	cfg.Contexts[name] = ctx
	cfg.CurrentContext = name
}
