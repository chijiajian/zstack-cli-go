package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDiskQueryByteFields(t *testing.T) {
	const capacity = int64(9007199254740993)
	const occupied = int64(1610612737)
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/zstack/v1/volumes" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		inventory(w,
			map[string]any{"uuid": "root", "name": "root-disk", "type": "Root", "vmInstanceUuid": "vm-a", "size": capacity, "actualSize": occupied},
			map[string]any{"uuid": "zero", "name": "empty-disk", "type": "Data", "size": 0, "actualSize": 0},
			map[string]any{"uuid": "unknown", "name": "detached-disk", "type": "Data", "lastVmInstanceUuid": "vm-a", "actualSize": nil},
		)
	})
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			out, stderr, err := invokeCLI(t, cfg, "", "get", "disks", "-o", format)
			if err != nil {
				t.Fatalf("query failed: %v, stderr=%s", err, stderr)
			}
			var disks []struct {
				Type            string `json:"type" yaml:"type"`
				VM              string `json:"vmInstanceUuid" yaml:"vmInstanceUuid"`
				Attached        string `json:"attached" yaml:"attached"`
				Size            string `json:"size" yaml:"size"`
				ActualSize      string `json:"actualSize" yaml:"actualSize"`
				SizeBytes       *int64 `json:"sizeBytes" yaml:"sizeBytes"`
				ActualSizeBytes *int64 `json:"actualSizeBytes" yaml:"actualSizeBytes"`
			}
			unmarshal := json.Unmarshal
			nullSize, nullActual := `"sizeBytes": null`, `"actualSizeBytes": null`
			if format == "yaml" {
				unmarshal = yaml.Unmarshal
				nullSize, nullActual = "sizeBytes: null", "actualSizeBytes: null"
			}
			if err := unmarshal([]byte(out), &disks); err != nil || len(disks) != 3 {
				t.Fatalf("invalid %s output: %v\n%s", format, err, out)
			}
			root, zero, unknown := disks[0], disks[1], disks[2]
			if root.SizeBytes == nil || *root.SizeBytes != capacity || root.ActualSizeBytes == nil || *root.ActualSizeBytes != occupied ||
				root.Type != "Root" || root.VM != "vm-a" || root.Attached != "Yes" || root.Size != "8192.00 TB" || root.ActualSize != "1.50 GB" {
				t.Fatalf("root disk fields changed: %s", out)
			}
			if zero.SizeBytes == nil || *zero.SizeBytes != 0 || zero.ActualSizeBytes == nil || *zero.ActualSizeBytes != 0 {
				t.Fatalf("reported zero was lost: %s", out)
			}
			if unknown.SizeBytes != nil || unknown.ActualSizeBytes != nil || !strings.Contains(out, nullSize) || !strings.Contains(out, nullActual) ||
				unknown.Type != "Data" || unknown.VM != "" || unknown.Attached != "No" || unknown.Size != zero.Size || unknown.ActualSize != zero.ActualSize {
				t.Fatalf("missing bytes or detached disk behavior changed: %s", out)
			}
		})
	}
	out, stderr, err := invokeCLI(t, cfg, "", "get", "disks")
	if err != nil || strings.Contains(out, "SIZE BYTES") || !strings.Contains(out, "ACTUAL SIZE") || !strings.Contains(out, "1.50 GB") {
		t.Fatalf("default table changed: %v, stdout=%s, stderr=%s", err, out, stderr)
	}
	out, stderr, err = invokeCLI(t, cfg, "", "get", "disks", "--fields", "uuid,sizeBytes,actualSizeBytes")
	if err != nil || !strings.Contains(out, "SIZE BYTES") || !strings.Contains(out, "ACTUAL SIZE BYTES") ||
		!strings.Contains(out, "9007199254740993") || !strings.Contains(out, "1610612737") {
		t.Fatalf("explicit byte columns missing: %v, stdout=%s, stderr=%s", err, out, stderr)
	}
}

func TestVolumeAliasByteProjectionAndPaging(t *testing.T) {
	var pages atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/zstack/v1/volumes" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("q") == "uuid=missing" {
			inventory(w)
			return
		}
		pages.Add(1)
		if q.Get("q") != "vmInstanceUuid=vm-a" || q.Get("limit") != "1" || q.Get("sort") != "+uuid" || q.Get("replyWithCount") != "true" {
			t.Errorf("unexpected volume query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		switch q.Get("start") {
		case "", "0":
			_, _ = w.Write([]byte(`{"inventories":[{"uuid":"root","size":9007199254740993,"actualSize":0}],"total":2}`))
		case "1":
			_, _ = w.Write([]byte(`{"inventories":[{"uuid":"data","size":1610612737}],"total":2}`))
		default:
			t.Errorf("unexpected page: %s", r.URL.RawQuery)
			_, _ = w.Write([]byte(`{"inventories":[],"total":2}`))
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "get", "volumes", "-q", "vmInstanceUuid=vm-a", "--all-pages", "--limit", "1", "--reply-with-count", "--fields", "uuid,sizeBytes,actualSizeBytes", "-o", "json")
	var result struct {
		Items []map[string]json.RawMessage `json:"items"`
		Total int                          `json:"total"`
	}
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || result.Total != 2 || len(result.Items) != 2 || pages.Load() != 2 {
		t.Fatalf("unexpected page results: %v, stdout=%s, stderr=%s", err, out, stderr)
	}
	for i, want := range [][3]string{{`"root"`, "9007199254740993", "0"}, {`"data"`, "1610612737", "null"}} {
		item := result.Items[i]
		if len(item) != 3 || string(item["uuid"]) != want[0] || string(item["sizeBytes"]) != want[1] || string(item["actualSizeBytes"]) != want[2] {
			t.Fatalf("unexpected projection: %s", out)
		}
	}
	out, stderr, err = invokeCLI(t, cfg, "", "get", "volumes", "-q", "uuid=missing", "--fields", "sizeBytes,actualSizeBytes", "-o", "json")
	if err != nil || strings.TrimSpace(out) != "[]" {
		t.Fatalf("empty list contract changed: %v, stdout=%s, stderr=%s", err, out, stderr)
	}
}
