package main

import (
	"net/http"
	"sync/atomic"
	"testing"
)

func TestLegacyAndSnapshotConfirmation(t *testing.T) {
	for _, args := range [][]string{{"delete", "image", "demo"}, {"delete", "instance", "demo"}, {"expunge", "image", "demo"}, {"expunge", "instance", "demo"}, {"snapshot", "revert", "demo"}} {
		t.Run(args[0]+"/"+args[1], func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					writes.Add(1)
					http.Error(w, "mock mutation failure", 500)
					return
				}
				state, status := "Running", "Ready"
				if args[0] == "expunge" {
					state, status = "Destroyed", "Deleted"
				}
				inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": state, "status": status})
			})
			out, stderr, err := invokeCLI(t, cfg, "no\n", args...)
			if err != nil || writes.Load() != 0 {
				t.Fatalf("decline mutated: %v %s %s", err, out, stderr)
			}
			out, stderr, err = invokeCLI(t, cfg, "", append(args, "--yes")...)
			if err == nil || writes.Load() != 1 {
				t.Fatalf("--yes/HTTP error: writes=%d err=%v out=%s stderr=%s", writes.Load(), err, out, stderr)
			}
		})
	}
}
