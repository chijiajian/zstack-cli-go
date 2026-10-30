package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

func setupLogout(t *testing.T, endpoint string) {
	t.Helper()
	t.Setenv("ZSTACK_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	config.SetContextOverride("selected")
	ResetClient()
	t.Cleanup(func() { config.SetContextOverride(""); ResetClient() })
	cfg := &config.ZStackConfig{CurrentContext: "default", Contexts: map[string]config.Context{
		"default":  {Endpoint: "other", SessionUUID: "keep", Password: "keep"},
		"selected": {Endpoint: endpoint, Username: "admin", Password: "password", SessionUUID: "secret-token"},
	}}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
}
func TestLogoutRevokesSavedSessionWithoutRenewal(t *testing.T) {
	for _, status := range []int{200, 204, 404, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodDelete || r.URL.Path != "/zstack/v1/accounts/sessions/secret-token" || r.Header.Get("Authorization") != "OAuth secret-token" {
					t.Errorf("unexpected auth/request: %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(status)
				if status == 500 {
					fmt.Fprint(w, "secret-token")
				}
			}))
			defer server.Close()
			setupLogout(t, strings.TrimPrefix(server.URL, "http://"))
			err := LogoutContext(context.Background(), false)
			if (err != nil) != (status == 500) || calls != 1 {
				t.Fatalf("status=%d calls=%d err=%v", status, calls, err)
			}
			if err != nil && strings.Contains(err.Error(), "secret-token") {
				t.Fatal("session leaked in error")
			}
			cfg, e := config.LoadConfig()
			if e != nil {
				t.Fatal(e)
			}
			if cfg.CurrentContext != "default" || cfg.Contexts["default"].SessionUUID != "keep" {
				t.Fatal("other context modified")
			}
			selected := cfg.Contexts["selected"]
			if status == 500 {
				if selected.SessionUUID != "secret-token" || selected.Password == "" {
					t.Fatal("failed logout cleared credentials")
				}
			} else if selected.SessionUUID != "" || selected.Password != "" {
				t.Fatal("credentials retained after success")
			}
		})
	}
}
func TestLogoutPreservesConcurrentReplacement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := config.UpdateConfig(context.Background(), func(cfg *config.ZStackConfig) error {
			c := cfg.Contexts["selected"]
			c.SessionUUID = "new-token"
			cfg.Contexts["selected"] = c
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	setupLogout(t, strings.TrimPrefix(server.URL, "http://"))
	if err := LogoutContext(context.Background(), false); err == nil {
		t.Fatal("concurrent update not detected")
	}
	cfg, _ := config.LoadConfig()
	if cfg.Contexts["selected"].SessionUUID != "new-token" {
		t.Fatal("new session erased")
	}
}
func TestLogoutCancellationAndLocalOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	setupLogout(t, strings.TrimPrefix(server.URL, "http://"))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := LogoutContext(ctx, false); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline not preserved: %v", err)
	}
	cfg, _ := config.LoadConfig()
	if cfg.Contexts["selected"].SessionUUID == "" {
		t.Fatal("canceled logout cleared credentials")
	}
	if err := LogoutContext(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	cfg, _ = config.LoadConfig()
	if cfg.Contexts["selected"].SessionUUID != "" {
		t.Fatal("local logout did not clear")
	}
}
func TestLogoutAccessKeyIsLocal(t *testing.T) {
	setupLogout(t, "unused.invalid")
	config.UpdateConfig(context.Background(), func(cfg *config.ZStackConfig) error {
		c := cfg.Contexts["selected"]
		c.AuthType = config.AuthTypeAccessKey
		c.AccessKeyID = "key"
		c.AccessKeySecret = "secret"
		cfg.Contexts["selected"] = c
		return nil
	})
	if err := LogoutContext(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.LoadConfig()
	c := cfg.Contexts["selected"]
	if c.AccessKeyID != "" || c.AccessKeySecret != "" || c.Password != "" || c.SessionUUID != "" {
		t.Fatal("local credentials not cleared")
	}
}
