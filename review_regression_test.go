package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"gopkg.in/yaml.v3"
)

func TestMalformedQueryResponsesFail(t *testing.T) {
	for _, body := range []string{`{}`, `{"inventories":null}`, `{"inventories":[],"total":-1}`} {
		cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
		out, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "-o", "json")
		if err == nil || out != "" || stderr == "" {
			t.Fatalf("body=%s err=%v out=%s stderr=%s", body, err, out, stderr)
		}
	}
}

func TestActionProjectionValidatedBeforeMutation(t *testing.T) {
	var calls atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method == "GET" {
			inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": "Destroyed"})
			return
		}
		fmt.Fprint(w, `{"inventory":{"uuid":"id1","name":"demo","state":"Stopped"}}`)
	})
	out, stderr, err := invokeCLI(t, cfg, "", "instance", "recover", "id1", "--fields", "typo", "-o", "json")
	if err == nil || calls.Load() != 0 || out != "" {
		t.Fatalf("invalid field caused IO: %v %s %s", err, out, stderr)
	}
	out, stderr, err = invokeCLI(t, cfg, "", "instance", "recover", "id1", "--fields", "uuid", "-o", "json")
	var result map[string]any
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || len(result) != 1 || result["uuid"] != "id1" {
		t.Fatalf("%v %s %s", err, out, stderr)
	}
}

func TestJobDetailProjectionIncludesEmbeddedFields(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": "Failed"})
	})
	out, stderr, err := invokeCLI(t, cfg, "", "describe", "job", "id1", "--fields", "uuid,state", "-o", "json")
	var result map[string]any
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || len(result) != 2 || result["uuid"] != "id1" {
		t.Fatalf("%v %s %s", err, out, stderr)
	}
}

func TestRefreshDoesNotOverwriteConfigChangesDuringLogin(t *testing.T) {
	for _, changeCredentials := range []bool{false, true} {
		t.Run(fmt.Sprint(changeCredentials), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/valid"):
					fmt.Fprint(w, `{"valid":false}`)
				case strings.HasSuffix(r.URL.Path, "/login"):
					latest := readContexts(t, path)
					latest.CurrentContext = "b"
					if changeCredentials {
						c := latest.Contexts["a"]
						c.Password = "replacement"
						latest.Contexts["a"] = c
					}
					data, _ := yaml.Marshal(latest)
					if err := os.WriteFile(path, data, 0600); err != nil {
						t.Error(err)
					}
					fmt.Fprint(w, `{"inventory":{"uuid":"renewed"}}`)
				default:
					inventory(w)
				}
			}))
			defer server.Close()
			initial := config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{
				"a": {Endpoint: strings.TrimPrefix(server.URL, "http://"), Username: "a", Password: "old", SessionUUID: "expired"},
				"b": {Endpoint: "b", AccessKeyID: "b", AccessKeySecret: "b-secret"},
			}}
			data, marshalErr := yaml.Marshal(initial)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			out, stderr, err := invokeCLI(t, path, "", "get", "zones", "-o", "json")
			if (err != nil) != changeCredentials {
				t.Fatalf("%v %s %s", err, out, stderr)
			}
			saved := readContexts(t, path)
			if saved.CurrentContext != "b" {
				t.Fatal("session refresh reverted context switch")
			}
			if changeCredentials {
				if saved.Contexts["a"].Password != "replacement" || saved.Contexts["a"].SessionUUID != "expired" {
					t.Fatal("renewal overwrote replaced credentials")
				}
			} else if saved.Contexts["a"].SessionUUID != "renewed" {
				t.Fatal("session not saved")
			}
		})
	}
}

func TestConfigMachineOutputAndInvalidFormat(t *testing.T) {
	cfg := writeContexts(t, config.ZStackConfig{CurrentContext: "a", Contexts: map[string]config.Context{"a": {Endpoint: "a"}, "b": {Endpoint: "b"}}})
	before, _ := os.ReadFile(cfg)
	out, stderr, err := invokeCLI(t, cfg, "", "config", "use-context", "b", "-o", "nonsense")
	after, _ := os.ReadFile(cfg)
	if err == nil || out != "" || string(before) != string(after) {
		t.Fatalf("bad format mutated config: %v %s %s", err, out, stderr)
	}
	for _, args := range [][]string{{"config", "use-context", "b", "-o", "json"}, {"config", "current-context", "-o", "json"}} {
		out, stderr, err = invokeCLI(t, cfg, "", args...)
		if err != nil || !json.Valid([]byte(out)) {
			t.Fatalf("%v %s %s", err, out, stderr)
		}
	}
}
