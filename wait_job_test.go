package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestWaitJobLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name, state, target, output string
		fail                        bool
	}{
		{"success", "Succeeded", "", "json", false},
		{"failed", "Failed", "", "json", true},
		{"canceled", "Canceled", "", "yaml", true},
		{"explicit-failure", "Failed", "Failed", "yaml", false},
		{"unreachable-target", "Succeeded", "Suspended", "text", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/zstack/v1/longjobs" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
				}
				n := calls.Add(1)
				if n > 1 && r.URL.Query().Get("q") != "uuid=id1" {
					t.Errorf("poll must pin UUID: %s", r.URL)
				}
				state := "Running"
				if n > 1 {
					state = tc.state
				}
				inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": state, "jobResult": "storage unavailable"})
			})
			args := []string{"wait", "job", "demo", "--interval", "1ms", "--timeout", "3s", "--progress", "-o", tc.output}
			if tc.target != "" {
				args = append(args, "--for", tc.target)
			}
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if (err != nil) != tc.fail || calls.Load() != 2 {
				t.Fatalf("err=%v calls=%d out=%s stderr=%s", err, calls.Load(), out, stderr)
			}
			if !strings.Contains(stderr, "state=Running") || !strings.Contains(stderr, "state="+tc.state) {
				t.Fatal(stderr)
			}
			if tc.fail {
				if out != "" || !strings.Contains(stderr, "storage unavailable") {
					t.Fatalf("out=%s stderr=%s", out, stderr)
				}
			} else {
				var result map[string]any
				if tc.output == "json" {
					err = json.Unmarshal([]byte(out), &result)
				} else {
					err = yaml.Unmarshal([]byte(out), &result)
				}
				if err != nil || result["state"] != tc.state || result["jobResult"] != "storage unavailable" {
					t.Fatalf("out=%s err=%v", out, err)
				}
			}
		})
	}
}

func TestJobDetailAndListPreserveResult(t *testing.T) {
	for _, args := range [][]string{{"describe", "job", "id1", "-o", "json"}, {"get", "jobs", "-o", "json", "--fields", "uuid,jobResult"}} {
		t.Run(args[0], func(t *testing.T) {
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				want := "/zstack/v1/longjobs"
				if args[0] == "describe" {
					want += "/id1"
				}
				if r.URL.Path != want {
					t.Errorf("wrong URL %s", r.URL)
				}
				inventory(w, map[string]any{"uuid": "id1", "state": "Failed", "jobResult": `{"error":{"details":"disk full"}}`})
			})
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if err != nil || !json.Valid([]byte(out)) || !strings.Contains(out, "disk full") {
				t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
			}
		})
	}
}

func TestDescribeJobDeadline(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "describe", "job", "id1", "--request-timeout", "500ms", "-o", "json")
	if err == nil || out != "" || !strings.Contains(stderr, "deadline exceeded") {
		t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
	}
}

func TestWaitPollErrorIncludesLastState(t *testing.T) {
	var calls atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			inventory(w, map[string]any{"uuid": "id1", "state": "Stopped"})
			return
		}
		http.Error(w, "backend unavailable", 503)
	})
	out, stderr, err := invokeCLI(t, cfg, "", "wait", "instance", "id1", "--for", "Running", "--interval", "1ms", "-o", "json")
	if err == nil || out != "" || !strings.Contains(stderr, "state=Stopped") || !strings.Contains(stderr, "503") {
		t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
	}
}

func TestWaitLongJobExactUUIDAndTerminalResult(t *testing.T) {
	for _, tc := range []struct {
		name, uuid, state, code string
		exit                    int
	}{
		{"failed", "job-id", "Failed", "conflict", 5},
		{"canceled", "job-id", "Canceled", "conflict", 5},
		{"mismatched-uuid", "other-id", "Succeeded", "api_failed", 8},
		{"no-name-fallback", "", "", "not_found", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != "GET" || r.URL.Query().Get("q") != "uuid=job-id" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
				}
				if tc.uuid == "" {
					w.Write([]byte(`{"inventories":[]}`))
					return
				}
				inventory(w, map[string]any{"uuid": tc.uuid, "name": "fixture", "state": tc.state, "jobResult": "private payload"})
			})
			out, stderr, err := invokeCLI(t, cfg, "", "wait", "long-job", "job-id", "--uuid-only", "-o", "json", "--error-format", "json")
			assertDiagnostic(t, stderr, err, tc.code, tc.exit)
			if calls.Load() != 1 || strings.Contains(stderr, "private payload") {
				t.Fatalf("unexpected calls=%d diagnostic=%s", calls.Load(), stderr)
			}
			if tc.exit == 5 {
				var result map[string]any
				if err := json.Unmarshal([]byte(out), &result); err != nil || result["uuid"] != tc.uuid || result["state"] != tc.state {
					t.Fatalf("invalid terminal observation: %s (%v)", out, err)
				}
			} else if out != "" {
				t.Fatalf("unexpected result %s", out)
			}
		})
	}
}
