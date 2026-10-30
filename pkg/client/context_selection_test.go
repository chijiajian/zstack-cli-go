package client

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

func TestClientCacheFollowsContextSelection(t *testing.T) {
	t.Setenv("ZSTACK_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	ResetClient()
	config.SetContextOverride("")
	t.Cleanup(func() { ResetClient(); config.SetContextOverride("") })
	cfg := &config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{
		"a": {Endpoint: "localhost:8080", AuthType: "access-key", AccessKeyID: "a", AccessKeySecret: "a-secret"},
		"b": {Endpoint: "localhost:8080", AuthType: "access-key", AccessKeyID: "b", AccessKeySecret: "b-secret"},
	}}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	a, err := GetClientContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	config.SetContextOverride("b")
	b, err := GetClientContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if a == b || clientContext.AccessKeyID != "b" {
		t.Fatal("client reused credentials from another context")
	}
	config.SetContextOverride("missing")
	if _, err = GetClientContext(context.Background()); err == nil {
		t.Fatal("unknown context fell back to cached client")
	}
	config.SetContextOverride("")
	_, err = GetClientContext(context.Background())
	if err != nil || clientContext.AccessKeyID != "a" {
		t.Fatalf("default not restored: %v", err)
	}
	saved, err := config.LoadConfig()
	if err != nil || saved.CurrentContext != "a" {
		t.Fatal("selection changed persisted default")
	}
}

func TestPinnedContextRejectsChangedIdentityButAllowsSessionRotation(t *testing.T) {
	t.Setenv("ZSTACK_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	config.SetContextOverride("")
	ResetClient()
	t.Cleanup(func() { config.SetContextOverride(""); ResetClient() })
	cfg := &config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{"a": {Endpoint: "a", Username: "admin", SessionUUID: "first"}}}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	config.PinContextSelection(cfg)
	if err := config.UpdateConfig(context.Background(), func(latest *config.ZStackConfig) error {
		c := latest.Contexts["a"]
		c.SessionUUID = "second"
		latest.Contexts["a"] = c
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if GetSessionUUID() != "second" || !IsLoggedIn() {
		t.Fatal("session rotation rejected")
	}
	if err := config.UpdateConfig(context.Background(), func(latest *config.ZStackConfig) error {
		c := latest.Contexts["a"]
		c.Endpoint = "b"
		latest.Contexts["a"] = c
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if GetSessionUUID() != "" || IsLoggedIn() {
		t.Fatal("identity guard bypassed")
	}
	if err := Logout(); err == nil {
		t.Fatal("logout accepted changed context")
	}
	saved, err := config.LoadConfig()
	if err != nil || saved.Contexts["a"].SessionUUID != "second" {
		t.Fatal("logout cleared replacement credentials")
	}
}
