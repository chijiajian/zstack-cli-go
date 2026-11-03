// Package monitor reads capacity counters and host utilization samples from ZStack.
package monitor

import (
	"fmt"
	"time"
)

type Reader interface {
	GetWithRespKey(path, uuid, key string, params, result interface{}) error
}

type Capacity struct {
	Metric    string   `json:"metric" yaml:"metric"`
	Unit      string   `json:"unit" yaml:"unit"`
	Total     int64    `json:"total" yaml:"total"`
	Used      int64    `json:"used" yaml:"used"`
	Available int64    `json:"available" yaml:"available"`
	Percent   *float64 `json:"percent" yaml:"percent"`
}
type Load struct {
	SampleTime     *time.Time `json:"sampleTime" yaml:"sampleTime"`
	ConnectedHosts int        `json:"connectedHosts" yaml:"connectedHosts"`
	SampledHosts   int        `json:"sampledHosts" yaml:"sampledHosts"`
	CPU            *float64   `json:"cpuPercent" yaml:"cpuPercent"`
	Memory         *float64   `json:"memoryPercent" yaml:"memoryPercent"`
}
type Report struct {
	Capacity []Capacity `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	Load     *Load      `json:"load,omitempty" yaml:"load,omitempty"`
}

func ReadCapacity(r Reader) ([]Capacity, error) {
	all := struct {
		All bool `json:"all"`
	}{true}
	var compute struct{ TotalCpu, AvailableCpu, TotalMemory, AvailableMemory *int64 }
	var storage struct{ TotalCapacity, AvailableCapacity, TotalPhysicalCapacity, AvailablePhysicalCapacity *int64 }
	if err := r.GetWithRespKey("v1/hosts/capacities/cpu-memory", "", "", &all, &compute); err != nil {
		return nil, err
	}
	if err := r.GetWithRespKey("v1/primary-storage/capacities", "", "", &all, &storage); err != nil {
		return nil, err
	}
	var rows []Capacity
	for _, c := range []struct {
		name, unit   string
		total, avail *int64
	}{
		{"cpu", "vCPU", compute.TotalCpu, compute.AvailableCpu}, {"memory", "bytes", compute.TotalMemory, compute.AvailableMemory},
		{"primaryStorageLogical", "bytes", storage.TotalCapacity, storage.AvailableCapacity}, {"primaryStoragePhysical", "bytes", storage.TotalPhysicalCapacity, storage.AvailablePhysicalCapacity},
	} {
		if c.total == nil || c.avail == nil || *c.avail < 0 || *c.avail > *c.total {
			return nil, fmt.Errorf("invalid %s capacity counters", c.name)
		}
		used := *c.total - *c.avail
		row := Capacity{Metric: c.name, Unit: c.unit, Total: *c.total, Available: *c.avail, Used: used}
		if *c.total > 0 {
			v := float64(used) * 100 / float64(*c.total)
			row.Percent = &v
		}
		rows = append(rows, row)
	}
	return rows, nil
}

type sample struct {
	Time   int64    `json:"time"`
	Value  *float64 `json:"value"`
	Labels struct {
		HostUUID string `json:"HostUuid"`
	} `json:"labels"`
}
type sampleKey struct {
	ts   int64
	host string
}

func ReadLoad(r Reader, now time.Time) (*Load, error) {
	hosts := map[string]bool{}
	for start := 0; ; {
		var page struct {
			Inventories *[]struct {
				UUID string `json:"uuid"`
			} `json:"inventories"`
		}
		q := struct {
			Q      string `json:"q"`
			Fields string `json:"fields"`
			Sort   string `json:"sort"`
			Start  int    `json:"start"`
			Limit  int    `json:"limit"`
		}{"status=Connected", "uuid", "+uuid", start, 100}
		if err := r.GetWithRespKey("v1/hosts", "", "", &q, &page); err != nil {
			return nil, err
		}
		if page.Inventories == nil {
			return nil, fmt.Errorf("missing host inventory")
		}
		for _, host := range *page.Inventories {
			if host.UUID == "" || hosts[host.UUID] {
				return nil, fmt.Errorf("invalid or duplicate host identity during pagination")
			}
			hosts[host.UUID] = true
		}
		start += len(*page.Inventories)
		if len(*page.Inventories) < q.Limit {
			break
		}
	}
	load := &Load{ConnectedHosts: len(hosts)}
	if len(hosts) == 0 {
		return load, nil
	}
	values := make([]map[sampleKey]float64, 2)
	for i, name := range []string{"CPUAverageUsedUtilization", "MemoryUsedInPercent"} {
		q := struct {
			Namespace string `json:"namespace"`
			Metric    string `json:"metricName"`
			Start     int64  `json:"startTime"`
			End       int64  `json:"endTime"`
			Period    int    `json:"period"`
		}{"ZStack/Host", name, now.Add(-15 * time.Minute).Unix(), now.Unix(), 60}
		var resp struct {
			Data *[]sample `json:"data"`
		}
		if err := r.GetWithRespKey("v1/zwatch/metrics", "", "", &q, &resp); err != nil {
			return nil, err
		}
		if resp.Data == nil {
			return nil, fmt.Errorf("missing %s data", name)
		}
		values[i] = map[sampleKey]float64{}
		for _, s := range *resp.Data {
			if !hosts[s.Labels.HostUUID] || s.Value == nil || s.Time < q.Start || s.Time > q.End {
				continue
			}
			values[i][sampleKey{s.Time, s.Labels.HostUUID}] = *s.Value
		}
	}
	var latest int64
	for k := range values[0] {
		if _, ok := values[1][k]; ok && k.ts > latest {
			latest = k.ts
		}
	}
	if latest == 0 {
		return load, nil
	}
	var cpu, mem float64
	for k, v := range values[0] {
		if m, ok := values[1][k]; ok && k.ts == latest {
			load.SampledHosts++
			cpu += v
			mem += m
		}
	}
	cpu /= float64(load.SampledHosts)
	mem /= float64(load.SampledHosts)
	ts := time.Unix(latest, 0).UTC()
	load.CPU = &cpu
	load.Memory = &mem
	load.SampleTime = &ts
	return load, nil
}
