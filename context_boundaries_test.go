package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"gopkg.in/yaml.v3"
)

func TestCommandPinsContextAcrossFilterResolution(t *testing.T) {
	for _, mode := range []string{"switch", "replace"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			var wrong, queries atomic.Int32
			b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { wrong.Add(1); inventory(w) }))
			defer b.Close()
			a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/zstack/v1/zones" {
					args := []string{"config", "use-context", "b"}
					if mode == "replace" {
						args = []string{"login", strings.TrimPrefix(b.URL, "http://"), "--ak", "a-key", "--sk", "a-secret", "--context-name", "a", "--no-switch"}
					}
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					changed := <-concurrentCLI(ctx, path, args...)
					if changed.err != nil {
						t.Errorf("change context: %v %s", changed.err, changed.output)
					}
					inventory(w, map[string]any{"uuid": "zone-a", "name": "zone"})
					return
				}
				if r.URL.Path != "/zstack/v1/vm-instances" || r.URL.Query().Get("q") != "zoneUuid=zone-a" {
					t.Errorf("bad final query %s", r.URL)
				}
				queries.Add(1)
				inventory(w)
			}))
			defer a.Close()
			initial := config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{
				"a": {Endpoint: strings.TrimPrefix(a.URL, "http://"), AuthType: "access-key", AccessKeyID: "a-key", AccessKeySecret: "a-secret"},
				"b": {Endpoint: strings.TrimPrefix(b.URL, "http://"), AuthType: "access-key", AccessKeyID: "b-key", AccessKeySecret: "b-secret"},
			}}
			data, _ := yaml.Marshal(initial)
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			out, stderr, err := invokeCLI(t, path, "", "get", "instances", "--zone", "zone", "-o", "json")
			if wrong.Load() != 0 {
				t.Fatal("command crossed environments")
			}
			if mode == "switch" {
				if err != nil || queries.Load() != 1 || readContexts(t, path).CurrentContext != "b" {
					t.Fatalf("%v %s %s", err, out, stderr)
				}
			} else if err == nil || queries.Load() != 0 || !strings.Contains(stderr, "changed during command") {
				t.Fatalf("%v %s %s", err, out, stderr)
			}
		})
	}
}

func TestSessionValidationRequiresExplicitBoolean(t *testing.T) {
	for _, response := range []string{`{}`, `{"valid":null}`, `{"valid":"true"}`} {
		var logins atomic.Int32
		cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/valid") {
				fmt.Fprint(w, response)
				return
			}
			logins.Add(1)
			http.Error(w, "must not authenticate or query", 500)
		})
		accountConfig(t, cfg, true)
		before, _ := os.ReadFile(cfg)
		out, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "-o", "json")
		after, _ := os.ReadFile(cfg)
		if err == nil || out != "" || logins.Load() != 0 || string(before) != string(after) {
			t.Fatalf("response=%s err=%v out=%s stderr=%s", response, err, out, stderr)
		}
	}
}
