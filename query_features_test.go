package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestNetworkQueriesAndVolumeAlias(t *testing.T) {
	cases := []struct{ name, path, filter, parent, parentPath string }{
		{"security-groups", "security-groups", "", "", ""},
		{"security-group-rules", "security-groups/rules", "security-group", "securityGroupUuid", "security-groups"},
		{"port-forwarding-rules", "port-forwarding", "", "", ""},
		{"load-balancers", "load-balancers", "", "", ""},
		{"load-balancer-listeners", "load-balancers/listeners", "load-balancer", "loadBalancerUuid", "load-balancers"},
		{"load-balancer-server-groups", "load-balancers/servergroups", "load-balancer", "loadBalancerUuid", "load-balancers"},
		{"volumes", "volumes", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reads atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("unexpected mutation %s", r.Method)
				}
				if tc.parentPath != "" && r.URL.Path == "/zstack/v1/"+tc.parentPath {
					if r.URL.Query().Get("q") == "uuid=parent" {
						inventory(w)
						return
					}
					inventory(w, map[string]any{"uuid": "parent-id", "name": "parent"})
					return
				}
				reads.Add(1)
				if r.URL.Path != "/zstack/v1/"+tc.path {
					t.Errorf("wrong path %s", r.URL.Path)
				}
				if tc.filter != "" && !strings.Contains(strings.Join(r.URL.Query()["q"], " "), tc.parent+"=parent-id") {
					t.Errorf("missing resolved filter: %s", r.URL)
				}
				inventory(w, map[string]any{"uuid": "item1", "name": "demo"})
			})
			args := []string{"get", tc.name, "-o", "json", "--fields", "uuid"}
			if tc.filter != "" {
				args = append(args, "--"+tc.filter, "parent")
			}
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			var rows []map[string]any
			if err != nil || json.Unmarshal([]byte(out), &rows) != nil || len(rows) != 1 || len(rows[0]) != 1 || rows[0]["uuid"] != "item1" || reads.Load() != 1 {
				t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
			}
		})
	}
}

func TestQueryOutputContracts(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"empty-json", []string{"get", "hosts", "-o", "json"}},
		{"empty-yaml", []string{"get", "eips", "-o", "yaml"}},
		{"count", []string{"get", "zones", "--count", "-o", "json"}},
		{"count-table", []string{"get", "zones", "--count"}},
		{"metadata", []string{"get", "zones", "--reply-with-count", "-o", "json", "--fields", "uuid"}},
		{"legacy-pagination", []string{"get", "instances", "--pagination", "-o", "json"}},
		{"all-pages", []string{"get", "zones", "--all-pages", "--limit", "2", "-q", "state=Enabled", "-o", "json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				switch tc.name {
				case "empty-json", "empty-yaml":
					inventory(w)
				case "count", "count-table":
					if r.URL.Query().Get("count") != "true" {
						t.Error("missing count")
					}
					fmt.Fprint(w, `{"total":3}`)
				case "all-pages":
					start, _ := strconv.Atoi(r.URL.Query().Get("start"))
					if r.URL.Query().Get("q") != "state=Enabled" || r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("sort") != "+uuid" {
						t.Errorf("bad page parameters: %s", r.URL)
					}
					objects := []map[string]any{}
					for i := start; i < start+2 && i < 3; i++ {
						objects = append(objects, map[string]any{"uuid": fmt.Sprint(i), "name": fmt.Sprint(i)})
					}
					json.NewEncoder(w).Encode(map[string]any{"inventories": objects, "total": 3})
				default:
					fmt.Fprint(w, `{"inventories":[{"uuid":"id1","name":"demo"}],"total":3}`)
				}
			})
			out, stderr, err := invokeCLI(t, cfg, "", tc.args...)
			if err != nil {
				t.Fatalf("%v %s", err, stderr)
			}
			switch tc.name {
			case "empty-json":
				if strings.TrimSpace(out) != "[]" {
					t.Fatalf("not empty JSON list: %s", out)
				}
			case "empty-yaml":
				var rows []any
				if yaml.Unmarshal([]byte(out), &rows) != nil || len(rows) != 0 || strings.TrimSpace(out) != "[]" {
					t.Fatalf("not empty YAML list: %s", out)
				}
			case "count", "count-table":
				if strings.TrimSpace(out) != "3" || calls.Load() != 1 {
					t.Fatalf("count: %s calls=%d", out, calls.Load())
				}
			case "all-pages":
				var rows []any
				if json.Unmarshal([]byte(out), &rows) != nil || len(rows) != 3 || calls.Load() != 2 {
					t.Fatalf("pages=%d output=%s", calls.Load(), out)
				}
			default:
				var result struct {
					Items []map[string]any `json:"items"`
					Total int              `json:"total"`
				}
				if json.Unmarshal([]byte(out), &result) != nil || result.Total != 3 || len(result.Items) != 1 {
					t.Fatalf("invalid envelope %s", out)
				}
				if tc.name == "metadata" && len(result.Items[0]) != 1 {
					t.Fatalf("projection %s", out)
				}
			}
		})
	}
}

func TestDescribePreservesDetailWhenRelatedResourceDenied(t *testing.T) {
	for _, kind := range []string{"instance", "volume"} {
		t.Run(kind, func(t *testing.T) {
			path := "vm-instances"
			field, relatedPath := "hostUuid", "hosts"
			if kind == "volume" {
				path = "volumes"
				field, relatedPath = "primaryStorageUuid", "primary-storage"
			}
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/zstack/v1/"+relatedPath+"/") || r.URL.Path == "/zstack/v1/primary-storage" {
					if kind == "volume" && r.URL.Query().Get("q") != "uuid=related-id" {
						t.Errorf("missing exact UUID filter: %s", r.URL)
					}
					http.Error(w, `{"error":{"code":"permission.denied"}}`, 403)
					return
				}
				if r.URL.Path != "/zstack/v1/"+path && r.URL.Path != "/zstack/v1/"+path+"/id1" {
					t.Errorf("unexpected path %s", r.URL)
				}
				inventory(w, map[string]any{"uuid": "id1", "name": "demo", field: "related-id", "size": int64(9007199254740993), "state": "Running"})
			})
			out, stderr, err := invokeCLI(t, cfg, "", "describe", kind, "id1", "-o", "json")
			var doc struct {
				Resource map[string]json.RawMessage `json:"resource"`
				Warnings []string                   `json:"warnings"`
			}
			if err != nil || json.Unmarshal([]byte(out), &doc) != nil || len(doc.Warnings) != 1 || !strings.Contains(stderr, "permission.denied") || string(doc.Resource["size"]) != "9007199254740993" {
				t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
			}
		})
	}
}

func TestSDK010DetachContract(t *testing.T) {
	for _, instanceScoped := range []bool{false, true} {
		t.Run(fmt.Sprint(instanceScoped), func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					id := "volume-id"
					if strings.Contains(r.URL.Path, "vm-instances") {
						id = "vm-id"
					}
					inventory(w, map[string]any{"uuid": id, "name": "demo"})
					return
				}
				writes.Add(1)
				if r.Method != "DELETE" || r.URL.Path != "/zstack/v1/volumes/volume-id/vm-instances" {
					t.Errorf("wrong detach request %s %s", r.Method, r.URL)
				}
				if instanceScoped && r.URL.Query().Get("vmUuid") != "vm-id" {
					t.Errorf("missing VM scope %s", r.URL)
				}
				fmt.Fprint(w, `{"inventory":{"uuid":"volume-id","name":"demo"}}`)
			})
			args := []string{"volume", "detach", "volume-id"}
			if instanceScoped {
				args = []string{"instance", "detach-volume", "vm-id", "--volume", "volume-id"}
			}
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if err != nil || writes.Load() != 1 {
				t.Fatalf("err=%v out=%s stderr=%s writes=%d", err, out, stderr, writes.Load())
			}
		})
	}
}

func TestQueryFailuresDoNotEmitPartialResults(t *testing.T) {
	for _, scenario := range []string{"later-page", "deadline"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if scenario == "deadline" {
					select {
					case <-r.Context().Done():
					case <-time.After(3 * time.Second):
					}
					return
				}
				if r.URL.Query().Get("start") == "0" {
					fmt.Fprint(w, `{"inventories":[{"uuid":"first"}],"total":2}`)
					return
				}
				http.Error(w, "second page unavailable", http.StatusServiceUnavailable)
			})
			args := []string{"get", "security-groups", "--all-pages", "--limit", "1", "-o", "json"}
			if scenario == "deadline" {
				args = append(args, "--request-timeout", "500ms")
			}
			out, stderr, err := invokeCLI(t, cfg, "", args...)
			if err == nil || strings.TrimSpace(out) != "" {
				t.Fatalf("expected failure without partial output, err=%v stdout=%s stderr=%s", err, out, stderr)
			}
			want := "HTTP 503"
			if scenario == "deadline" {
				want = "context deadline exceeded"
			}
			if !strings.Contains(stderr, want) {
				t.Fatalf("expected %q: %s", want, stderr)
			}
		})
	}
}
