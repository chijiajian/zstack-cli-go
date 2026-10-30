package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRecoverResizeSDKContracts(t *testing.T) {
	for _, tc := range []struct{ kind, action, volumeType, key, path string }{
		{"instance", "recover", "", "recoverVmInstance", "vm-instances/id1/actions"},
		{"volume", "recover", "Data", "recoverDataVolume", "volumes/id1/actions"},
		{"volume", "resize", "Data", "resizeDataVolume", "volumes/data/resize/id1/actions"},
		{"volume", "resize", "Root", "resizeRootVolume", "volumes/resize/id1/actions"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			for _, fail := range []bool{false, true} {
				t.Run(fmt.Sprint(fail), func(t *testing.T) {
					var writes atomic.Int32
					cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
						if r.Method == "GET" {
							if r.URL.Query().Get("q") == "uuid=demo" {
								inventory(w)
								return
							}
							status := "Ready"
							if tc.action == "recover" {
								status = "Deleted"
							}
							inventory(w, map[string]any{"uuid": "id1", "name": "demo", "type": tc.volumeType, "status": status, "state": "Destroyed", "size": 1 << 30, "vmInstanceUuid": "vm1"})
							return
						}
						writes.Add(1)
						if r.Method != "PUT" || r.URL.Path != "/zstack/v1/"+tc.path {
							t.Errorf("wrong action %s %s", r.Method, r.URL)
						}
						var body map[string]map[string]any
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Error(err)
						}
						params, ok := body[tc.key]
						if !ok {
							t.Errorf("missing action: %+v", body)
						}
						if tc.action == "resize" && params["size"] != float64(2<<30) {
							t.Errorf("wrong size: %+v", body)
						}
						if tc.volumeType == "Root" && params["vmInstanceUuid"] != "vm1" {
							t.Errorf("missing VM: %+v", body)
						}
						if fail {
							http.Error(w, `{"error":{"code":"operation.failed","description":"rejected"}}`, 500)
							return
						}
						fmt.Fprint(w, `{"inventory":{"uuid":"id1","name":"demo","size":2147483648}}`)
					})
					args := []string{tc.kind, tc.action, "demo", "-o", "json"}
					if tc.action == "resize" {
						args = append(args, "--size", "2G")
					}
					out, stderr, err := invokeCLI(t, cfg, "", args...)
					if writes.Load() != 1 || (err != nil) != fail {
						t.Fatalf("writes=%d err=%v out=%s stderr=%s", writes.Load(), err, out, stderr)
					}
					if fail {
						if out != "" || stderr == "" {
							t.Fatalf("out=%s stderr=%s", out, stderr)
						}
					} else if !json.Valid([]byte(out)) || !strings.Contains(out, "id1") {
						t.Fatal(out)
					}
				})
			}
		})
	}
}

func TestRecoverResizeRejectsBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name, kind, action, state, status, typ, size string
		ambiguous                                    bool
	}{
		{"running-instance", "instance", "recover", "Running", "", "", "", false},
		{"live-volume", "volume", "recover", "", "Ready", "Data", "", false},
		{"root-recovery", "volume", "recover", "", "Deleted", "Root", "", false},
		{"shrink", "volume", "resize", "", "Ready", "Data", "1G", false},
		{"equal", "volume", "resize", "", "Ready", "Data", "2G", false},
		{"deleted", "volume", "resize", "", "Deleted", "Data", "3G", false},
		{"unknown-type", "volume", "resize", "", "Ready", "Other", "3G", false},
		{"ambiguous", "volume", "resize", "", "Ready", "Data", "3G", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					writes.Add(1)
					http.Error(w, "unexpected mutation", 500)
					return
				}
				if r.URL.Query().Get("q") == "uuid=demo" {
					inventory(w)
					return
				}
				item := map[string]any{"uuid": "id1", "name": "demo", "state": tc.state, "status": tc.status, "type": tc.typ, "size": 2 << 30}
				if tc.ambiguous {
					inventory(w, item, map[string]any{"uuid": "id2", "name": "demo"})
					return
				}
				inventory(w, item)
			})
			args := []string{tc.kind, tc.action, "demo"}
			if tc.size != "" {
				args = append(args, "--size", tc.size)
			}
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if err == nil || writes.Load() != 0 || out != "" {
				t.Fatalf("err=%v writes=%d out=%s stderr=%s", err, writes.Load(), out, stderr)
			}
		})
	}
}

func TestRecoverResizeOfflineValidationAndDryRun(t *testing.T) {
	var requests atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", 500)
	})
	for _, args := range [][]string{
		{"instance", "recover", "demo", "--dry-run", "-o", "json"},
		{"volume", "recover", "demo", "--dry-run", "-o", "json"},
		{"volume", "resize", "demo", "--size", "20G", "--dry-run", "-o", "json"},
	} {
		out, stderr, err := invokeCLI(t, cfg, "", args...)
		if err != nil || !json.Valid([]byte(out)) {
			t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
		}
	}
	for _, size := range []string{"", "0", "-1G", "1.5G", "8388608T", "16777217T", "banana"} {
		args := []string{"volume", "resize", "demo", "--dry-run", "--size", size}
		out, stderr, err := invokeCLI(t, cfg, "", args...)
		if err == nil || out != "" {
			t.Fatalf("size=%s err=%v out=%s stderr=%s", size, err, out, stderr)
		}
	}
	if requests.Load() != 0 {
		t.Fatalf("offline paths made %d requests", requests.Load())
	}
}
