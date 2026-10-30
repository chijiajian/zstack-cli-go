package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

type concurrentResult struct {
	output string
	err    error
}

func concurrentCLI(ctx context.Context, path string, args ...string) <-chan concurrentResult {
	result := make(chan concurrentResult, 1)
	go func() {
		cmd := exec.CommandContext(ctx, cliBinary, args...)
		cmd.Env = append(os.Environ(), "ZSTACK_CONFIG="+path, "ZSTACK_ACCESS_KEY_ID=", "ZSTACK_ACCESS_KEY_SECRET=")
		output, err := cmd.CombinedOutput()
		result <- concurrentResult{string(output), err}
	}()
	return result
}
func TestConcurrentContextWritersPreserveEveryChange(t *testing.T) {
	path := writeContexts(t, config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{"a": {Endpoint: "a"}, "b": {Endpoint: "b"}}})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	results := []<-chan concurrentResult{}
	for i := 0; i < 16; i++ {
		results = append(results, concurrentCLI(ctx, path, "login", "localhost:8080", "--ak", fmt.Sprint(i), "--sk", "secret", "--context-name", fmt.Sprintf("env-%d", i), "--no-switch"))
	}
	results = append(results, concurrentCLI(ctx, path, "config", "use-context", "b"))
	for _, result := range results {
		r := <-result
		if r.err != nil {
			t.Fatalf("%v %s", r.err, r.output)
		}
	}
	saved := readContexts(t, path)
	if len(saved.Contexts) != 18 || saved.CurrentContext != "b" {
		t.Fatalf("lost concurrent changes: count=%d current=%s", len(saved.Contexts), saved.CurrentContext)
	}
	for i := 0; i < 16; i++ {
		if saved.Contexts[fmt.Sprintf("env-%d", i)].AccessKeyID != fmt.Sprint(i) {
			t.Fatal("lost credentials")
		}
	}
}

func TestConcurrentRefreshAndContextSwitch(t *testing.T) {
	arrived := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer unblock()
	var logins atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/valid"):
			fmt.Fprint(w, `{"valid":false}`)
		case strings.HasSuffix(r.URL.Path, "/login"):
			id := logins.Add(1)
			arrived <- struct{}{}
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			fmt.Fprintf(w, `{"inventory":{"uuid":"session-%d"}}`, id)
		default:
			if !strings.HasPrefix(r.Header.Get("Authorization"), "OAuth session-") {
				t.Errorf("wrong refreshed credentials: %s", r.Header.Get("Authorization"))
			}
			inventory(w)
		}
	}))
	defer server.Close()
	defer unblock()
	path := writeContexts(t, config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{
		"a": {Endpoint: strings.TrimPrefix(server.URL, "http://"), Username: "admin", Password: "password", SessionUUID: "expired"}, "b": {Endpoint: "b"},
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	first := concurrentCLI(ctx, path, "get", "zones", "--context", "a", "-o", "json")
	second := concurrentCLI(ctx, path, "get", "zones", "--context", "a", "-o", "json")
	for i := 0; i < 2; i++ {
		select {
		case <-arrived:
		case <-ctx.Done():
			t.Fatal("both refreshes did not reach login")
		}
	}
	// This must finish before network logins are released: no lock is held over HTTP.
	switched := <-concurrentCLI(ctx, path, "config", "use-context", "b")
	if switched.err != nil {
		t.Fatalf("%v %s", switched.err, switched.output)
	}
	unblock()
	for _, pending := range []<-chan concurrentResult{first, second} {
		r := <-pending
		if r.err != nil {
			t.Fatalf("%v %s", r.err, r.output)
		}
	}
	saved := readContexts(t, path)
	if saved.CurrentContext != "b" || !strings.HasPrefix(saved.Contexts["a"].SessionUUID, "session-") {
		t.Fatal("refresh lost switch or session")
	}
}

func TestRefreshCannotResurrectDeletedOrReplacedContext(t *testing.T) {
	for _, action := range []string{"delete", "replace"} {
		t.Run(action, func(t *testing.T) {
			arrived := make(chan struct{}, 1)
			release := make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/valid") {
					fmt.Fprint(w, `{"valid":false}`)
					return
				}
				if strings.HasSuffix(r.URL.Path, "/login") {
					arrived <- struct{}{}
					select {
					case <-release:
					case <-r.Context().Done():
						return
					}
					fmt.Fprint(w, `{"inventory":{"uuid":"stale-new-session"}}`)
					return
				}
				t.Error("query must not run after context replacement")
				inventory(w)
			}))
			defer server.Close()
			defer unblock()
			path := writeContexts(t, config.ZStackConfig{CurrentContext: "b", Contexts: map[string]config.Context{
				"a": {Endpoint: strings.TrimPrefix(server.URL, "http://"), Username: "admin", Password: "password", SessionUUID: "expired"}, "b": {Endpoint: "b"},
			}})
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			refreshing := concurrentCLI(ctx, path, "get", "zones", "--context", "a", "-o", "json")
			select {
			case <-arrived:
			case <-ctx.Done():
				t.Fatal("login did not start")
			}
			args := []string{"config", "delete-context", "a"}
			if action == "replace" {
				args = []string{"login", "localhost:8080", "--ak", "new-key", "--sk", "new-secret", "--context-name", "a", "--no-switch"}
			}
			changed := <-concurrentCLI(ctx, path, args...)
			if changed.err != nil {
				t.Fatalf("%v %s", changed.err, changed.output)
			}
			unblock()
			result := <-refreshing
			if result.err == nil || !strings.Contains(result.output, "changed during login") {
				t.Fatalf("refresh should reject conflict: %v %s", result.err, result.output)
			}
			saved := readContexts(t, path)
			if saved.CurrentContext != "b" {
				t.Fatal("default changed")
			}
			if action == "delete" {
				if _, ok := saved.Contexts["a"]; ok {
					t.Fatal("deleted context resurrected")
				}
			} else if saved.Contexts["a"].AccessKeyID != "new-key" || saved.Contexts["a"].SessionUUID != "" {
				t.Fatal("replacement credentials overwritten")
			}
		})
	}
}
