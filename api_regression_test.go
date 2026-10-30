package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var cliBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "zstack-cli-test-")
	if err != nil {
		panic(err)
	}
	cliBinary = filepath.Join(dir, "zstack-cli")
	build := exec.Command("go", "build", "-o", cliBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		fmt.Fprintf(os.Stderr, "build: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func mockConfig(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := fmt.Sprintf("current-context: mock\ncontexts:\n  mock:\n    endpoint: %s\n    auth_type: access-key\n    access_key_id: test-ak\n    access_key_secret: test-sk\n", strings.TrimPrefix(server.URL, "http://"))
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func invokeCLI(t *testing.T, config, input string, args ...string) (string, string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, cliBinary, args...)
	command.Env = append(os.Environ(), "ZSTACK_CONFIG="+config, "ZSTACK_ACCESS_KEY_ID=", "ZSTACK_ACCESS_KEY_SECRET=")
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI hung: %v", args)
	}
	return stdout.String(), stderr.String(), err
}

func inventory(w http.ResponseWriter, objects ...map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	if objects == nil {
		objects = []map[string]any{}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"inventories": objects})
}

func TestDestructiveDryRunSendsNoWrites(t *testing.T) {
	for _, action := range []string{"delete", "expunge"} {
		for _, resource := range []string{"instance", "image"} {
			t.Run(action+"/"+resource, func(t *testing.T) {
				var reads, writes atomic.Int32
				cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
					if r.Method != "GET" {
						writes.Add(1)
						w.Write([]byte(`{}`))
						return
					}
					reads.Add(1)
					state, status := "Stopped", "Ready"
					if action == "expunge" {
						state, status = "Destroyed", "Deleted"
					}
					inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": state, "status": status})
				})
				out, stderr, err := invokeCLI(t, cfg, "yes\n", action, resource, "demo", "--dry-run")
				if err != nil || !strings.Contains(out, "Dry-run") || strings.Contains(out, "Are you sure") || reads.Load() == 0 || writes.Load() != 0 {
					t.Fatalf("err=%v reads=%d writes=%d stdout=%s stderr=%s", err, reads.Load(), writes.Load(), out, stderr)
				}
			})
		}
	}
}

func TestQueriesReturnErrorsOnStderr(t *testing.T) {
	resources := strings.Fields("zones clusters hosts images instances disks disk-offerings instance-offerings l2-networks l3-networks image-storages primary-storages management-nodes eips vips nics snapshots cdroms ip-ranges long-jobs vm-scripts vm-script-records virtual-routers virtual-router-offerings global-configs tags image-groups accounts users projects organizations roles policies alarms events audits")
	for _, resource := range resources {
		t.Run(resource, func(t *testing.T) {
			var calls atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				http.Error(w, `{"error":{"code":"mock.failure","description":"simulated failure"}}`, 500)
			})
			out, stderr, err := invokeCLI(t, cfg, "", "get", resource, "-o", "json")
			if err == nil || out != "" || !strings.Contains(stderr, "mock.failure") || calls.Load() == 0 {
				t.Fatalf("err=%v calls=%d stdout=%q stderr=%q", err, calls.Load(), out, stderr)
			}
		})
	}
}

func TestBatchWritePartialFailureReturnsNonZero(t *testing.T) {
	cases := []struct {
		args          []string
		state, status string
	}{
		{[]string{"delete", "instance", "demo"}, "Stopped", "Ready"},
		{[]string{"expunge", "instance", "demo"}, "Destroyed", "Deleted"},
		{[]string{"delete", "image", "demo"}, "Enabled", "Ready"},
		{[]string{"expunge", "image", "demo"}, "Enabled", "Deleted"},
		{[]string{"instance", "start", "demo", "-y"}, "Stopped", ""},
		{[]string{"instance", "stop", "demo", "-y"}, "Running", ""},
		{[]string{"instance", "restart", "demo", "-y"}, "Running", ""},
		{[]string{"instance", "pause", "demo", "-y"}, "Running", ""},
		{[]string{"instance", "resume", "demo", "-y"}, "Paused", ""},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": tc.state, "status": tc.status}, map[string]any{"uuid": "id2", "name": "demo", "state": tc.state, "status": tc.status})
					return
				}
				if writes.Add(1) == 1 {
					http.Error(w, `{"error":{"code":"mock.failure"}}`, 500)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"inventory":{"uuid":"id2","name":"demo"}}`))
			})
			out, stderr, err := invokeCLI(t, cfg, "yes\n", tc.args...)
			if err == nil || writes.Load() != 2 || stderr == "" {
				t.Fatalf("err=%v writes=%d stdout=%s stderr=%s", err, writes.Load(), out, stderr)
			}
		})
	}
}

func TestInstanceResolutionReachesRealActions(t *testing.T) {
	for _, args := range [][]string{{"wait", "instance", "demo", "--for", "Running"}, {"update", "instance", "demo", "--description", "updated"}} {
		t.Run(args[0], func(t *testing.T) {
			var reads, writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					reads.Add(1)
					inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": "Running"})
					return
				}
				writes.Add(1)
				w.Write([]byte(`{"inventory":{"uuid":"id1","name":"demo"}}`))
			})
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if err != nil || reads.Load() == 0 || (args[0] == "update" && writes.Load() != 1) {
				t.Fatalf("err=%v stdout=%s stderr=%s", err, out, stderr)
			}
		})
	}
}

func TestCreateInstanceAmbiguousImageDoesNotWrite(t *testing.T) {
	var writes atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes.Add(1)
			w.Write([]byte(`{}`))
			return
		}
		inventory(w, map[string]any{"uuid": "one", "name": "demo", "state": "Enabled", "status": "Ready"}, map[string]any{"uuid": "two", "name": "demo", "state": "Enabled", "status": "Ready"})
	})
	_, stderr, err := invokeCLI(t, cfg, "", "create", "instance", "vm", "--image", "demo", "--cpu", "1", "--memory", "1G", "--l3-network", "network")
	if err == nil || writes.Load() != 0 || !strings.Contains(stderr, "multiple image") {
		t.Fatalf("err=%v writes=%d stderr=%s", err, writes.Load(), stderr)
	}
}

func TestLoginWithoutSavedPasswordReusesSession(t *testing.T) {
	var logins, queries atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/login"):
			logins.Add(1)
			w.Write([]byte(`{"inventory":{"uuid":"saved-session","accountUuid":"account"}}`))
		case strings.HasSuffix(r.URL.Path, "/saved-session/valid"):
			w.Write([]byte(`{"valid":true}`))
		default:
			queries.Add(1)
			if r.Header.Get("Authorization") != "OAuth saved-session" {
				t.Errorf("unexpected auth header: %q", r.Header.Get("Authorization"))
			}
			inventory(w)
		}
	})
	// Recover the endpoint from the isolated config, never from the user's environment.
	data, _ := os.ReadFile(cfg)
	endpoint := strings.Fields(strings.Split(string(data), "endpoint: ")[1])[0]
	_, stderr, err := invokeCLI(t, cfg, "", "login", endpoint, "-u", "admin", "-p", "test-password", "--save-password=false")
	if err != nil {
		t.Fatalf("login: %v %s", err, stderr)
	}
	data, _ = os.ReadFile(cfg)
	if strings.Contains(string(data), "test-password") || !strings.Contains(string(data), "saved-session") {
		t.Fatalf("unexpected saved config: %s", data)
	}
	_, stderr, err = invokeCLI(t, cfg, "", "get", "zones", "-o", "json")
	if err != nil || logins.Load() != 1 || queries.Load() != 1 {
		t.Fatalf("err=%v logins=%d queries=%d stderr=%s", err, logins.Load(), queries.Load(), stderr)
	}
}

func accountConfig(t *testing.T, path string, password bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	endpoint := strings.Fields(strings.Split(string(data), "endpoint: ")[1])[0]
	body := fmt.Sprintf("current-context: mock\ncontexts:\n  mock:\n    endpoint: %s\n    auth_type: account\n    username: admin\n    session_uuid: expired-session\n", endpoint)
	if password {
		body += "    password: test-password\n"
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestExpiredSessionOnlyRelogsInWithCredentials(t *testing.T) {
	for _, scenario := range []string{"password", "no-password", "validation-failure"} {
		t.Run(scenario, func(t *testing.T) {
			var logins atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/valid"):
					if scenario == "validation-failure" {
						http.Error(w, `{"error":{"code":"validation.failure"}}`, 500)
						return
					}
					w.Write([]byte(`{"valid":false}`))
				case strings.Contains(r.URL.Path, "/login"):
					logins.Add(1)
					w.Write([]byte(`{"inventory":{"uuid":"new-session"}}`))
				default:
					inventory(w)
				}
			})
			accountConfig(t, cfg, scenario != "no-password")
			_, stderr, err := invokeCLI(t, cfg, "", "get", "zones")
			if scenario == "password" {
				data, _ := os.ReadFile(cfg)
				if err != nil || logins.Load() != 1 || !strings.Contains(string(data), "new-session") {
					t.Fatalf("err=%v logins=%d stderr=%s", err, logins.Load(), stderr)
				}
			} else if err == nil || logins.Load() != 0 {
				t.Fatalf("err=%v logins=%d stderr=%s", err, logins.Load(), stderr)
			}
		})
	}
}

func TestWaitDeadlineCoversEveryStage(t *testing.T) {
	for _, stage := range []string{"validation", "login", "resolution", "poll", "interval"} {
		t.Run(stage, func(t *testing.T) {
			var gets, stalls atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				stall := false
				switch {
				case strings.HasSuffix(r.URL.Path, "/valid"):
					if stage == "validation" {
						stall = true
					} else {
						w.Write([]byte(`{"valid":false}`))
						return
					}
				case strings.Contains(r.URL.Path, "/login"):
					stall = true
				default:
					n := gets.Add(1)
					stall = stage == "resolution" || (stage == "poll" && n > 1)
				}
				if stall {
					stalls.Add(1)
					select {
					case <-r.Context().Done():
					case <-time.After(6 * time.Second):
					}
					return
				}
				inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": "Stopped"})
			})
			if stage == "validation" || stage == "login" {
				accountConfig(t, cfg, true)
			}
			interval := "10s"
			if stage == "poll" {
				interval = "10ms"
			}
			start := time.Now()
			_, stderr, err := invokeCLI(t, cfg, "", "wait", "instance", "demo", "--for", "Running", "--timeout", "2s", "--interval", interval)
			elapsed := time.Since(start)
			if err == nil || elapsed > 4*time.Second || !strings.Contains(stderr, "deadline exceeded") {
				t.Fatalf("err=%v elapsed=%v stderr=%s", err, elapsed, stderr)
			}
			if stage != "interval" && stalls.Load() == 0 {
				t.Fatalf("stage %s not exercised", stage)
			}
		})
	}
}

func TestReleaseBuildEmbedsVersionAndCommit(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		t.Fatal(err)
	}
	// Use the actual release configuration, so a regression in the symbol path fails this test.
	var flags string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "-X ") {
			flags = strings.Trim(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")), "'")
			break
		}
	}
	if flags == "" {
		t.Fatal("release ldflags missing")
	}
	flags = strings.ReplaceAll(strings.ReplaceAll(flags, "{{ .Version }}", "9.8.7-review"), "{{ .Commit }}", "review-commit")
	binary := filepath.Join(t.TempDir(), "release-cli")
	build := exec.Command("go", "build", "-ldflags", flags, "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	command := exec.Command(binary, "--version")
	command.Env = append(os.Environ(), "ZSTACK_CONFIG="+filepath.Join(t.TempDir(), "config.yaml"))
	out, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "9.8.7-review, commit: review-commit") {
		t.Fatalf("err=%v output=%s", err, out)
	}
}
