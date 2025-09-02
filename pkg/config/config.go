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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View and modify ZStack CLI configuration and contexts.`,
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Println("Failed to load config:", err)
			return
		}
		fmt.Println("current-context:", cfg.CurrentContext)
		fmt.Println("contexts:")
		for name, ctx := range cfg.Contexts {
			fmt.Printf("    \"%s\":\n", name)
			fmt.Printf("        endpoint: %s\n", ctx.Endpoint)
			fmt.Printf("        username: %s\n", ctx.Username)
			fmt.Printf("        password: %s\n", maskPassword(ctx.Password))
			fmt.Printf("        session_uuid: %s\n", ctx.SessionUUID)
		}
	},
}

var useContextCmd = &cobra.Command{
	Use:   "use-context [name]",
	Short: "Switch to a specific context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Println("Failed to load config:", err)
			return
		}
		ctx, ok := cfg.Contexts[name]
		if !ok {
			fmt.Printf("Context %s not found\n", name)
			return
		}
		cfg.CurrentContext = name
		if err := SaveConfig(cfg); err != nil {
			fmt.Println("Failed to save config:", err)
			return
		}
		fmt.Printf("Switched to context \"%s\"\n", name)
		fmt.Printf("Endpoint: %s\n", ctx.Endpoint)
	},
}

func maskPassword(p string) string {
	if p == "" {
		return ""
	}
	return "*****"
}

func init() {
	ConfigCmd.AddCommand(viewCmd)
	ConfigCmd.AddCommand(useContextCmd)
}

type Context struct {
	Endpoint    string `yaml:"endpoint"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password,omitempty"`
	SessionUUID string `yaml:"session_uuid,omitempty"`
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
	file := getDefaultConfigFile()
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

func SaveConfig(cfg *ZStackConfig) error {
	file := getDefaultConfigFile()
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
	return os.WriteFile(file, data, 0o600)
}

func GetCurrentContext(cfg *ZStackConfig) (*Context, error) {
	if cfg.CurrentContext == "" {
		return nil, errors.New("no current context, please login first")
	}
	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		return nil, fmt.Errorf("context %s not found", cfg.CurrentContext)
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
