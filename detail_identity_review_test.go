package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestDescribeRejectsMalformedIdentity(t *testing.T) {
	for _, kind := range []string{"instance", "volume", "job"} {
		for _, item := range []any{nil, map[string]any{}, map[string]any{"uuid": "other"}} {
			t.Run(kind+"/"+fmt.Sprint(item), func(t *testing.T) {
				cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.URL.Path, "/id1") {
						json.NewEncoder(w).Encode(map[string]any{"inventories": []any{item}})
						return
					}
					inventory(w, map[string]any{"uuid": "id1", "name": "demo"})
				})
				out, stderr, err := invokeCLI(t, cfg, "", "describe", kind, "id1", "-o", "json")
				if err == nil || out != "" {
					t.Fatalf("malformed detail accepted: err=%v out=%s stderr=%s", err, out, stderr)
				}
			})
		}
	}
}

func TestRecoveryRejectsMissingIdentityBeforeMutation(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected mutation: %s %s", r.Method, r.URL.Path)
		}
		inventory(w, map[string]any{"state": "Destroyed"})
	})
	out, stderr, err := invokeCLI(t, cfg, "", "instance", "recover", "target", "-o", "json")
	if err == nil || out != "" || !strings.Contains(stderr, "without UUID") {
		t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
	}
}
