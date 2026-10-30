package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestTextQueryProjectsFieldsInsteadOfZeroingThem(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		inventory(w, map[string]any{"uuid": "id1", "name": "unselected", "description": "unselected-description"})
	})
	out, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "--fields", "uuid", "-o", "text")
	if err != nil || strings.TrimSpace(out) != "[map[uuid:id1]]" {
		t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
	}
}

func TestTextProjectionKeepsTotalOnStderr(t *testing.T) {
	for _, empty := range []bool{false, true} {
		t.Run(strconv.FormatBool(empty), func(t *testing.T) {
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				items := []map[string]any{}
				if !empty {
					items = append(items, map[string]any{"uuid": "id1", "name": "unselected"})
				}
				json.NewEncoder(w).Encode(map[string]any{"inventories": items, "total": len(items)})
			})
			out, stderr, err := invokeCLI(t, cfg, "", "get", "zones", "--fields", "uuid", "--reply-with-count", "-o", "text")
			want, total := "[map[uuid:id1]]", "Total: 1"
			if empty {
				want, total = "[]", "Total: 0"
			}
			if err != nil || strings.TrimSpace(out) != want || strings.TrimSpace(stderr) != total {
				t.Fatalf("err=%v out=%q stderr=%q", err, out, stderr)
			}
		})
	}
}
