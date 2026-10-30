package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func assertDiagnostic(t *testing.T, stderr string, err error, code string, exit int) {
	t.Helper()
	var process *exec.ExitError
	if !errors.As(err, &process) || process.ExitCode() != exit {
		t.Fatalf("exit=%v stderr=%s; want %d", err, stderr, exit)
	}
	var envelope struct {
		Error struct {
			Code     string `json:"code"`
			ExitCode int    `json:"exitCode"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(stderr), &envelope) != nil || envelope.Error.Code != code || envelope.Error.ExitCode != exit || envelope.Error.Message == "" {
		t.Fatalf("bad structured diagnostic: %s", stderr)
	}
}
func TestStructuredErrorsAndExitCodes(t *testing.T) {
	for _, tc := range []struct {
		status int
		code   string
		exit   int
	}{{400, "invalid_arguments", 2}, {401, "authentication_failed", 3}, {403, "authentication_failed", 3}, {404, "not_found", 4}, {409, "conflict", 5}, {504, "timeout", 6}, {500, "api_failed", 8}} {
		t.Run(tc.code+http.StatusText(tc.status), func(t *testing.T) {
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				w.Write([]byte(`{"error":{"details":"super-secret-server-echo"}}`))
			})
			out, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "--error-format", "json", "-o", "json")
			assertDiagnostic(t, stderr, err, tc.code, tc.exit)
			if out != "" || strings.Contains(stderr, "super-secret") {
				t.Fatal("diagnostics leaked data or polluted stdout")
			}
		})
	}
	cfg := filepath.Join(t.TempDir(), "config")
	for _, args := range [][]string{{"--error-format", "json", "unknown-command"}, {"--error-format", "json", "--unknown-flag"}, {"get", "zones", "--limit", "-1", "--error-format", "json"}, {"doctor", "extra", "--error-format", "json"}, {"volume", "resize", "target", "--error-format", "json"}, {"get", "zones", "--error-format", "json", "-o", "invalid"}} {
		_, stderr, err := invokeCLI(t, cfg, "", args...)
		assertDiagnostic(t, stderr, err, "invalid_arguments", 2)
	}
}
func TestDiagnosticDeadline(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	})
	_, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "--request-timeout", "20ms", "--error-format", "json")
	assertDiagnostic(t, stderr, err, "timeout", 6)
}
func TestDoctorDoesNotRenewOrWriteConfig(t *testing.T) {
	var calls int
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "GET" || r.URL.Path != "/zstack/v1/zones" || r.URL.Query().Get("limit") != "1" {
			t.Errorf("unexpected diagnostic request: %s %s", r.Method, r.URL)
		}
		inventory(w)
	})
	before, _ := os.ReadFile(cfg)
	for _, offline := range []bool{true, false} {
		args := []string{"doctor", "-o", "json"}
		if offline {
			args = append(args, "--offline")
		}
		out, stderr, err := invokeCLI(t, cfg, "", args...)
		if err != nil || stderr != "" || !strings.Contains(out, `"ok": true`) {
			t.Fatalf("doctor: %v %s %s", err, out, stderr)
		}
		if strings.Contains(out, "test-ak") || strings.Contains(out, "test-sk") {
			t.Fatal("doctor leaked credentials")
		}
	}
	after, _ := os.ReadFile(cfg)
	if string(before) != string(after) || calls != 1 {
		t.Fatal("doctor changed config or offline contacted server")
	}
}
func TestDoctorExpiredSessionNeverRenews(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/valid") {
			t.Errorf("doctor attempted renewal or unrelated request: %s %s", r.Method, r.URL)
		}
		w.Write([]byte(`{"valid":false}`))
	})
	data, _ := os.ReadFile(cfg)
	endpoint := strings.Split(strings.Split(string(data), "endpoint: ")[1], "\n")[0]
	os.WriteFile(cfg, []byte("current-context: mock\ncontexts:\n  mock:\n    endpoint: "+endpoint+"\n    username: admin\n    password: secret\n    session_uuid: expired\n"), 0600)
	before, _ := os.ReadFile(cfg)
	out, stderr, err := invokeCLI(t, cfg, "", "doctor", "-o", "json", "--error-format", "json")
	assertDiagnostic(t, stderr, err, "authentication_failed", 3)
	if !strings.Contains(out, `"ok": false`) || strings.Contains(out, "secret") || strings.Contains(out, "expired") {
		t.Fatal("unexpected doctor report")
	}
	after, _ := os.ReadFile(cfg)
	if string(before) != string(after) {
		t.Fatal("doctor rewrote credentials")
	}
}

func TestInstanceCreationProducesSingleJSONDocument(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			inventory(w, map[string]any{"uuid": "fixture", "name": "fixture", "state": "Enabled", "status": "Ready"})
			return
		}
		w.Write([]byte(`{"inventory":{"uuid":"created-vm","name":"test-vm"}}`))
	})
	out, stderr, err := invokeCLI(t, cfg, "", "create", "instance", "test-vm", "--image", "fixture", "--l3-network", "fixture", "--cpu", "1", "--memory", "1G", "-o", "json")
	var result map[string]any
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || result["uuid"] != "created-vm" {
		t.Fatalf("creation output: %v %s %s", err, out, stderr)
	}
}

func TestLegacySDKWriteUsesStructuredErrorContract(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"code":"SYS.1000","details":"secret-server-body"}}`))
	})
	_, stderr, err := invokeCLI(t, cfg, "", "create", "volume", "test", "--size", "1G", "--error-format", "json")
	assertDiagnostic(t, stderr, err, "api_failed", 8)
	if strings.Contains(stderr, "secret-server-body") {
		t.Fatal("raw SDK error leaked")
	}
}
