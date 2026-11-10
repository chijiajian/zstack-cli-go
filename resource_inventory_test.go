package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestResourceInventoryNumericProjection(t *testing.T) {
	const capacity = int64(9007199254740993)
	cases := []struct {
		command, endpoint, reported, exact, unreported string
	}{
		{"hosts", "hosts", "totalCpuCapacity", "totalCpuHz", "availableMemoryBytes"},
		{"primary-storages", "primary-storage", "totalCapacity", "totalCapacityBytes", "availablePhysicalCapacityBytes"},
		{"image-storages", "backup-storage", "totalCapacity", "totalCapacityBytes", "availableCapacityBytes"},
		{"images", "images", "size", "sizeBytes", "actualSizeBytes"},
		{"instance-offerings", "instance-offerings", "memorySize", "memorySizeBytes", "cpuCount"},
		{"disk-offerings", "disk-offerings", "diskSize", "diskSizeBytes", ""},
		{"l2-networks", "l2-networks", "virtualNetworkId", "virtualNetworkId", ""},
	}
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			value, expected := capacity, "9007199254740993"
			if tc.command == "l2-networks" {
				value, expected = 0, "0"
			}
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/zstack/v1/"+tc.endpoint {
					t.Errorf("unexpected inventory request: %s %s", r.Method, r.URL)
				}
				inventory(w, map[string]any{"uuid": "item-a", tc.reported: value}, map[string]any{"uuid": "item-b"})
			})
			fields := "uuid," + tc.exact
			if tc.unreported != "" {
				fields += "," + tc.unreported
			}
			out, stderr, err := invokeCLI(t, cfg, "", "get", tc.command, "--fields", fields, "-o", "json")
			var rows []map[string]json.RawMessage
			if err != nil || json.Unmarshal([]byte(out), &rows) != nil || len(rows) != 2 {
				t.Fatalf("bad result: %v %s %s", err, out, stderr)
			}
			if string(rows[0][tc.exact]) != expected || string(rows[1][tc.exact]) != "null" ||
				tc.unreported != "" && string(rows[0][tc.unreported]) != "null" {
				t.Fatalf("numeric projection lost precision or null: %s", out)
			}
		})
	}
}

func TestDescribeInstanceUUIDOnlyDoesNotFallBackToName(t *testing.T) {
	var nameQueries atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		switch {
		case r.URL.Path == "/zstack/v1/vm-instances" && r.URL.Query().Get("q") == "uuid=fixture-name":
			inventory(w)
		case r.URL.Path == "/zstack/v1/vm-instances" && r.URL.Query().Get("q") == "name=fixture-name":
			nameQueries.Add(1)
			inventory(w, map[string]any{"uuid": "fixture-id", "name": "fixture-name"})
		case r.URL.Path == "/zstack/v1/vm-instances/fixture-id":
			inventory(w, map[string]any{"uuid": "fixture-id", "name": "fixture-name"})
		default:
			t.Errorf("unexpected query: %s", r.URL)
			inventory(w)
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "describe", "instance", "fixture-name", "--uuid-only", "-o", "json")
	if err == nil || out != "" || nameQueries.Load() != 0 {
		t.Fatalf("UUID lookup fell back to name: %v, stdout=%s, stderr=%s", err, out, stderr)
	}
	out, stderr, err = invokeCLI(t, cfg, "", "describe", "instance", "fixture-name", "-o", "json")
	if err != nil || !strings.Contains(out, `"uuid": "fixture-id"`) || nameQueries.Load() != 1 {
		t.Fatalf("normal name lookup changed: %v, stdout=%s, stderr=%s", err, out, stderr)
	}
}
