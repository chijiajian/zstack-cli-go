package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestVMQueryMemoryBytes(t *testing.T) {
	const capacity = int64(1610612737)
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		inventory(w, map[string]any{"uuid": "a", "name": "vm", "memorySize": capacity, "cpuNum": 2}, map[string]any{"uuid": "b", "memorySize": 0, "cpuNum": 0}, map[string]any{"uuid": "c"})
	})
	for _, flags := range [][]string{nil, {"--fields", "uuid,cpuNum,memorySizeBytes"}} {
		out, stderr, err := invokeCLI(t, cfg, "", append([]string{"get", "instances", "-o", "json"}, flags...)...)
		var items []struct {
			MemorySize      string
			MemorySizeBytes *int64
			CPUNum          *int
		}
		if err != nil || json.Unmarshal([]byte(out), &items) != nil || len(items) != 3 {
			t.Fatalf("%v %s %s", err, out, stderr)
		}
		if items[0].MemorySizeBytes == nil || *items[0].MemorySizeBytes != capacity || items[1].MemorySizeBytes == nil || *items[1].MemorySizeBytes != 0 || items[2].MemorySizeBytes != nil || len(flags) == 0 && items[0].MemorySize == "" {
			t.Fatalf("wrong byte or legacy field: %s", out)
		}
		if items[0].CPUNum == nil || *items[0].CPUNum != 2 || items[1].CPUNum == nil || *items[1].CPUNum != 0 || items[2].CPUNum != nil {
			t.Fatalf("wrong reported CPU value: %s", out)
		}
	}
	out, stderr, err := invokeCLI(t, cfg, "", "get", "instances")
	if err != nil || strings.Contains(strings.ToLower(out), "memorysizebytes") {
		t.Fatalf("default table changed: %v %s %s", err, out, stderr)
	}
	out, stderr, err = invokeCLI(t, cfg, "", "get", "instances", "--fields", "memorySizeBytes")
	if err != nil || !strings.Contains(out, "1610612737") {
		t.Fatalf("opt-in byte column incorrect: %v %s %s", err, out, stderr)
	}
}
