package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestQueriesWithoutUUIDUseIDForPagination(t *testing.T) {
	for _, resource := range []string{"global-configs", "audits", "events"} {
		for _, explicit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/explicit=%v", resource, explicit), func(t *testing.T) {
				var calls atomic.Int32
				cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					want := "+id"
					if explicit {
						want = "-id"
					}
					if r.URL.Query().Get("sort") != want || r.URL.Query().Get("q") != "id>0" {
						t.Errorf("wrong query: %s", r.URL)
					}
					start, _ := strconv.Atoi(r.URL.Query().Get("start"))
					id := start + 1
					if explicit {
						id = 3 - start
					}
					json.NewEncoder(w).Encode(map[string]any{"total": 3, "inventories": []map[string]any{{"id": id, "name": fmt.Sprint(id), "category": "test"}}})
				})
				args := []string{"get", resource, "--all-pages", "--limit", "1", "-q", "id>0", "--reply-with-count", "-o", "json"}
				if explicit {
					args = append(args, "--sort", "-id")
				}
				out, stderr, err := invokeCLI(t, cfg, "", args...)
				var doc struct {
					Items []map[string]any `json:"items"`
					Total int              `json:"total"`
				}
				if err != nil || json.Unmarshal([]byte(out), &doc) != nil || doc.Total != 3 || len(doc.Items) != 3 || calls.Load() != 3 {
					t.Fatalf("err=%v calls=%d out=%s stderr=%s", err, calls.Load(), out, stderr)
				}
				for i, item := range doc.Items {
					want := i + 1
					if explicit {
						want = 3 - i
					}
					if resource == "global-configs" {
						if item["name"] != strconv.Itoa(want) {
							t.Fatalf("page order: item %d = %v, want name %d", i, item, want)
						}
					} else if item["id"] != float64(want) {
						t.Fatalf("page order: item %d = %v, want id %d", i, item, want)
					}
				}
			})
		}
	}
}

func TestUnknownQueryFieldsFailInEveryFormat(t *testing.T) {
	for _, format := range []string{"table", "text", "json", "yaml"} {
		for _, empty := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/empty=%v", format, empty), func(t *testing.T) {
				cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
					if empty {
						inventory(w)
					} else {
						inventory(w, map[string]any{"uuid": "id1", "name": "demo"})
					}
				})
				out, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "--fields", "notAField", "-o", format)
				if err == nil || out != "" || !strings.Contains(stderr, "unknown output field") {
					t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
				}
			})
		}
	}
}

func TestScriptRecordPositionalFilterUsesRecordName(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "recordName=run" {
			t.Errorf("wrong filter: %s", r.URL)
		}
		inventory(w, map[string]any{"uuid": "id1", "recordName": "run"})
	})
	out, stderr, err := invokeCLI(t, cfg, "", "get", "vm-script-records", "run", "-o", "json")
	if err != nil || !json.Valid([]byte(out)) {
		t.Fatalf("%v %s %s", err, out, stderr)
	}
}
