package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestZQLPassesStatementAndPreservesResult(t *testing.T) {
	const statement = "query vmInstance.uuid where name='app & db+01' limit 1"
	const response = `{"results":[{"inventories":[{"uuid":"vm"}],"total":9007199254740993,"returnWith":{"cpu":[{"value":0}]}}]}`
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/zstack/v1/zql" || r.URL.Query().Get("zql") != statement {
			t.Errorf("wrong request %s %v", r.Method, r.URL)
		}
		w.Write([]byte(response))
	})
	out, stderr, err := invokeCLI(t, cfg, "", "zql", statement)
	var result struct {
		Results []struct {
			Total      int64
			ReturnWith map[string][]struct{ Value int }
		}
	}
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || len(result.Results) != 1 || result.Results[0].Total != 9007199254740993 || len(result.Results[0].ReturnWith["cpu"]) != 1 {
		t.Fatalf("%v %s %s", err, out, stderr)
	}
}
func TestZQLResponseContract(t *testing.T) {
	for _, body := range []string{`{"results":[]}`, `{"results":"wrong"}`, `{}`} {
		t.Run(body, func(t *testing.T) {
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) })
			out, stderr, err := invokeCLI(t, cfg, "", "zql", "count vmInstance", "--error-format", "json")
			if body == `{"results":[]}` {
				if err != nil || !json.Valid([]byte(out)) {
					t.Fatalf("%v %s %s", err, out, stderr)
				}
				return
			}
			assertDiagnostic(t, stderr, err, "api_failed", 8)
			if out != "" {
				t.Fatal("invalid response emitted success output")
			}
		})
	}
}

func TestZQLPermissionAndRequestDeadline(t *testing.T) {
	t.Run("permission", func(t *testing.T) {
		cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "private response", http.StatusForbidden)
		})
		out, stderr, err := invokeCLI(t, cfg, "", "zql", "count vmInstance", "--error-format", "json")
		assertDiagnostic(t, stderr, err, "authentication_failed", 3)
		if out != "" || strings.Contains(stderr, "private response") {
			t.Fatalf("unexpected output: %s %s", out, stderr)
		}
	})
	t.Run("deadline", func(t *testing.T) {
		cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(3 * time.Second):
			}
		})
		out, stderr, err := invokeCLI(t, cfg, "", "zql", "count vmInstance", "--request-timeout", "500ms", "--error-format", "json")
		assertDiagnostic(t, stderr, err, "timeout", 6)
		if out != "" {
			t.Fatalf("unexpected output: %s", out)
		}
	})
}
