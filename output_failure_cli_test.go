package main

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestQueryCLIReportsUnwritableStdout(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		inventory(w, map[string]any{"uuid": "id1", "name": "demo"})
	})
	for _, format := range []string{"json", "yaml", "text", "table"} {
		t.Run(format, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "output")
			if err := os.WriteFile(path, nil, 0600); err != nil {
				t.Fatal(err)
			}
			output, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			cmd := exec.Command(cliBinary, "get", "zones", "-o", format)
			cmd.Env = append(os.Environ(), "ZSTACK_CONFIG="+cfg)
			cmd.Stdout = output
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil || stderr.Len() == 0 {
				t.Fatalf("expected CLI output failure: err=%v stderr=%s", err, stderr.String())
			}
		})
	}
}

func TestDescribeReportsUnwritableWarningOutput(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/zstack/v1/hosts/host1" {
			http.Error(w, "unavailable", http.StatusForbidden)
			return
		}
		inventory(w, map[string]any{"uuid": "id1", "name": "demo", "hostUuid": "host1"})
	})
	path := filepath.Join(t.TempDir(), "stderr")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	cmd := exec.Command(cliBinary, "describe", "instance", "id1", "-o", "json")
	cmd.Env = append(os.Environ(), "ZSTACK_CONFIG="+cfg)
	cmd.Stderr = output
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil || stdout.Len() != 0 {
		t.Fatalf("warning output failure ignored: err=%v out=%s", err, stdout.String())
	}
}

func TestWaitReportsUnwritableProgress(t *testing.T) {
	for _, state := range []string{"Running", "Succeeded"} {
		t.Run(state, func(t *testing.T) {
			var calls atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				inventory(w, map[string]any{"uuid": "id1", "state": state})
			})
			path := filepath.Join(t.TempDir(), "stderr")
			if err := os.WriteFile(path, nil, 0600); err != nil {
				t.Fatal(err)
			}
			output, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			cmd := exec.Command(cliBinary, "wait", "job", "id1", "--progress", "--timeout", "5s", "--interval", "10ms", "-o", "json")
			cmd.Env = append(os.Environ(), "ZSTACK_CONFIG="+cfg)
			cmd.Stderr = output
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Run(); err == nil || stdout.Len() != 0 || calls.Load() != 1 {
				t.Fatalf("progress failure ignored: err=%v calls=%d out=%s", err, calls.Load(), stdout.String())
			}
		})
	}
}

func TestDryRunCLIReportsUnwritableStdout(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("dry-run contacted API: %s", r.URL)
		http.Error(w, "unexpected", 500)
	})
	path := filepath.Join(t.TempDir(), "stdout")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	cmd := exec.Command(cliBinary, "volume", "attach", "vol", "--instance", "vm", "--dry-run", "-o", "json")
	cmd.Env = append(os.Environ(), "ZSTACK_CONFIG="+cfg)
	cmd.Stdout = output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil || !strings.Contains(stderr.String(), "write") {
		t.Fatalf("dry-run failure ignored: err=%v stderr=%s", err, stderr.String())
	}
}
