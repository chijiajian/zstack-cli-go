package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

type startResult struct {
	Operation string `json:"operation"`
	Items     []struct {
		UUID, Status string
		Error        *struct{ Code string }
	} `json:"items"`
}

func TestStartJSONContract(t *testing.T) {
	var writes atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("q") == "uuid=demo" || r.URL.Query().Get("q") == "uuid=gone" {
				inventory(w)
				return
			}
			inventory(w, map[string]any{"uuid": "ok", "state": "Stopped", "name": "demo"}, map[string]any{"uuid": "bad", "state": "Stopped", "name": "demo"}, map[string]any{"uuid": "skip", "state": "Running", "name": "demo"})
			return
		}
		writes.Add(1)
		switch r.URL.Path {
		case "/zstack/v1/vm-instances/ok/actions":
			var request map[string]json.RawMessage
			if r.Method != http.MethodPut || json.NewDecoder(r.Body).Decode(&request) != nil || string(request["startVmInstance"]) != "{}" {
				t.Error("unexpected start request")
			}
			fmt.Fprint(w, `{"inventory":{"uuid":"ok","state":"Running"}}`)
		case "/zstack/v1/vm-instances/bad/actions":
			http.Error(w, `{"error":{"code":"TEST.1","description":"synthetic-secret-echo"}}`, 500)
		default:
			t.Errorf("unexpected mutation: %s", r.URL.Path)
		}
	})
	for _, tc := range []struct {
		name, input string
		flags       []string
		wantWrites  int32
	}{
		{"partial", "", []string{"--yes"}, 2}, {"dry-run", "", []string{"--dry-run"}, 0}, {"declined", "n\n", nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := writes.Load()
			out, stderr, err := invokeCLI(t, cfg, tc.input, append([]string{"instance", "start", "demo", "-o", "json", "--error-format", "json"}, tc.flags...)...)
			if tc.wantWrites > 0 {
				assertDiagnostic(t, stderr, err, "operation_failed", 1)
			} else if err != nil {
				t.Fatal(err)
			}
			var result startResult
			if json.Unmarshal([]byte(out), &result) != nil || result.Operation != "start" || len(result.Items) != 3 || writes.Load()-before != tc.wantWrites || strings.Contains(out+stderr, "synthetic-secret-echo") {
				t.Fatalf("unexpected result %s %s", out, stderr)
			}
			want := map[string]string{"ok": "skipped", "bad": "skipped", "skip": "skipped"}
			if tc.wantWrites > 0 {
				want["ok"], want["bad"] = "succeeded", "failed"
			}
			for _, item := range result.Items {
				if item.Status != want[item.UUID] || (item.Status == "failed" && (item.Error == nil || item.Error.Code != "api_failed")) {
					t.Fatalf("unexpected item %+v", item)
				}
				delete(want, item.UUID)
			}
		})
	}
	before := writes.Load()
	out, stderr, err := invokeCLI(t, cfg, "", "instance", "start", "gone", "--uuid-only", "--yes", "-o", "json", "--error-format", "json")
	assertDiagnostic(t, stderr, err, "not_found", 4)
	if out != "" || writes.Load() != before {
		t.Fatal("missing UUID fell back to a VM name")
	}
}

func TestStartJSONVerifiesIdentity(t *testing.T) {
	var writes atomic.Int32
	for _, state := range []string{"Stopped", "Running"} {
		cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				inventory(w, map[string]any{"uuid": "vm-id", "state": state})
				return
			}
			writes.Add(1)
			fmt.Fprint(w, `{"inventory":{"uuid":"other-id"}}`)
		})
		out, stderr, err := invokeCLI(t, cfg, "", "instance", "start", "vm-id", "--uuid-only", "--yes", "-o", "json", "--error-format", "json")
		var result startResult
		if json.Unmarshal([]byte(out), &result) != nil || len(result.Items) != 1 || result.Items[0].UUID != "vm-id" {
			t.Fatalf("unexpected result: %s", out)
		}
		if state == "Stopped" {
			assertDiagnostic(t, stderr, err, "api_failed", 8)
			if result.Items[0].Status != "failed" {
				t.Fatal("mismatched inventory reported success")
			}
		} else if err != nil || result.Items[0].Status != "skipped" || writes.Load() != 1 {
			t.Fatalf("already running was not skipped: %v %s", err, out)
		}
	}
}
