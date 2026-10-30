package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestJobActionsValidateStateAndUseBoundedTransport(t *testing.T) {
	for _, tc := range []struct {
		action, state, key string
		allowed            bool
	}{{"resume", "Suspended", "resumeLongJob", true}, {"resume", "Running", "resumeLongJob", false}, {"clean", "Succeeded", "cleanLongJob", true}, {"clean", "Running", "cleanLongJob", false}} {
		t.Run(tc.action+tc.state, func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					inventory(w, map[string]any{"uuid": "job-id", "name": "job", "state": tc.state})
					return
				}
				writes.Add(1)
				if r.Method != "PUT" || r.URL.Path != "/zstack/v1/longjobs/job-id/actions" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
				}
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				if _, ok := body[tc.key]; !ok {
					t.Error("wrong job action")
				}
				w.Write([]byte(`{"inventory":{"uuid":"job-id","state":"Succeeded"}}`))
			})
			out, stderr, err := invokeCLI(t, cfg, "", "job", tc.action, "job", "--yes", "-o", "json", "--error-format", "json")
			if tc.allowed {
				if err != nil || writes.Load() != 1 || !strings.Contains(out, "job-id") {
					t.Fatalf("%v %s %s", err, out, stderr)
				}
			} else {
				assertDiagnostic(t, stderr, err, "conflict", 5)
				if writes.Load() != 0 {
					t.Fatal("invalid job state mutated")
				}
			}
		})
	}
}
func TestProjectLifecycleSDKActions(t *testing.T) {
	for _, action := range []string{"recover", "expunge"} {
		t.Run(action, func(t *testing.T) {
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					inventory(w, map[string]any{"uuid": "project-id", "name": "project", "state": "Deleted"})
					return
				}
				if r.Method != "PUT" || r.URL.Path != "/zstack/v1/iam2/projects/project-id/actions" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
				}
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				if _, ok := body[action+"IAM2Project"]; !ok {
					t.Error("wrong project action")
				}
				if action == "recover" {
					w.Write([]byte(`{"inventory":{"uuid":"project-id","name":"project"}}`))
				} else {
					w.WriteHeader(204)
				}
			})
			args := []string{"project", action, "project", "--yes"}
			if action == "expunge" {
				args = []string{"expunge", "project", "project", "--yes"}
			}
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if err != nil || !strings.Contains(out, "project-id") {
				t.Fatalf("%v %s %s", err, out, stderr)
			}
		})
	}
}

func TestIAMMutationRejectsUnconfirmedIdentity(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		response string
	}{
		{"recover-missing", []string{"project", "recover", "example", "--yes"}, `{"inventory":{"name":"example"}}`},
		{"recover-wrong", []string{"project", "recover", "example", "--yes"}, `{"inventory":{"uuid":"other"}}`},
		{"update-wrong", []string{"update", "account", "example", "--description", "changed"}, `{"inventory":{"uuid":"other"}}`},
		{"create-wrong", []string{"create", "project", "example", "--resource-uuid", strings.Repeat("a", 32)}, `{"inventory":{"uuid":"other"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					inventory(w, map[string]any{"uuid": "target", "name": "example"})
					return
				}
				writes.Add(1)
				w.Write([]byte(tc.response))
			})
			out, stderr, err := invokeCLI(t, cfg, "", append(tc.args, "-o", "json", "--error-format", "json")...)
			assertDiagnostic(t, stderr, err, "api_failed", 8)
			if out != "" || writes.Load() != 1 {
				t.Fatalf("unexpected success output or retry: %q writes=%d", out, writes.Load())
			}
		})
	}
}
