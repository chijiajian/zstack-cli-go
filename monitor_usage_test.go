package main

import (
	"encoding/json"
	"github.com/chijiajian/zstack-cli-go/pkg/monitor"
	"net/http"
	"testing"
	"time"
)

func TestMonitorUsageMatchesHostsAndTimestamp(t *testing.T) {
	now := time.Now().Unix() - 10
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected method %s", r.Method)
		}
		switch r.URL.Path {
		case "/zstack/v1/hosts/capacities/cpu-memory":
			if r.URL.Query().Get("all") != "true" {
				t.Error("missing all=true")
			}
			w.Write([]byte(`{"totalCpu":10,"availableCpu":4,"totalMemory":0,"availableMemory":0}`))
		case "/zstack/v1/primary-storage/capacities":
			w.Write([]byte(`{"totalCapacity":100,"availableCapacity":25,"totalPhysicalCapacity":200,"availablePhysicalCapacity":150}`))
		case "/zstack/v1/hosts":
			if r.URL.Query().Get("q") != "status=Connected" {
				t.Error("missing connected filter")
			}
			w.Write([]byte(`{"inventories":[{"uuid":"a"},{"uuid":"b"}]}`))
		case "/zstack/v1/zwatch/metrics":
			rows := []map[string]any{}
			add := func(host string, ts int64, value float64) {
				rows = append(rows, map[string]any{"labels": map[string]string{"HostUuid": host}, "time": ts, "value": value})
			}
			switch r.URL.Query().Get("metricName") {
			case "CPUAverageUsedUtilization":
				add("a", now-60, 10)
				add("a", now, 30)
				add("b", now, 95)
				add("offline", now, 99)
			case "MemoryUsedInPercent":
				add("a", now-60, 15)
				add("a", now, 40)
				add("b", now-60, 80)
				add("offline", now, 95)
			default:
				t.Error("unexpected metric")
			}
			json.NewEncoder(w).Encode(map[string]any{"data": rows})
		default:
			t.Errorf("unexpected request %s", r.URL)
			http.NotFound(w, r)
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "monitor", "usage", "-o", "json")
	if err != nil {
		t.Fatalf("%v %s", err, stderr)
	}
	var report monitor.Report
	if err = json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Capacity) != 4 || *report.Capacity[0].Percent != 60 || report.Capacity[1].Percent != nil || *report.Capacity[3].Percent != 25 {
		t.Fatalf("wrong capacity %+v", report.Capacity)
	}
	l := report.Load
	if l == nil || l.SampledHosts != 1 || l.ConnectedHosts != 2 || *l.CPU != 30 || *l.Memory != 40 || l.SampleTime.Unix() != now {
		t.Fatalf("wrong host load %+v", l)
	}
}
func TestMonitorMissingSamplesRemainUnknown(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/zstack/v1/hosts":
			w.Write([]byte(`{"inventories":[{"uuid":"a"}]}`))
		case "/zstack/v1/zwatch/metrics":
			w.Write([]byte(`{"data":[]}`))
		default:
			t.Errorf("unexpected request %s", r.URL)
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "monitor", "usage", "--section", "load", "-o", "json")
	var report monitor.Report
	if err != nil || json.Unmarshal([]byte(out), &report) != nil || report.Load == nil || report.Load.CPU != nil || report.Load.Memory != nil || report.Load.SampleTime != nil || report.Load.SampledHosts != 0 {
		t.Fatalf("%v %s %s", err, out, stderr)
	}
}
