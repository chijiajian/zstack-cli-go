package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestExactDeviceGuards(t *testing.T) {
	var mutex sync.Mutex
	var writes []string
	iso := "image-iso"
	nicPresent := true
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		path := strings.TrimPrefix(r.URL.Path, "/zstack/v1/")
		query := r.URL.Query()
		if r.Method == http.MethodGet {
			switch path {
			case "vm-instances":
				inventory(w, map[string]any{"uuid": strings.TrimPrefix(query.Get("q"), "uuid="), "state": "Stopped"})
			case "vm-instances/nics":
				if query.Get("q") == "uuid=nic-exact" && nicPresent {
					inventory(w, map[string]any{"uuid": "nic-exact", "vmInstanceUuid": "vm-target", "l3NetworkUuid": "l3-target"})
				} else {
					inventory(w)
				}
			case "vm-instances/cdroms":
				inventory(w, map[string]any{"uuid": "drive-a", "vmInstanceUuid": "vm-target", "isoUuid": iso})
			case "volumes":
				if query.Get("q") == "uuid=root-a" {
					inventory(w, map[string]any{"uuid": "root-a", "type": "Root", "status": "Ready", "size": 1 << 30, "vmInstanceUuid": "vm-target"})
				} else {
					inventory(w, map[string]any{"uuid": "disk-a", "type": "Data", "status": "Ready", "size": 1 << 30, "vmInstanceUuid": "vm-target"})
				}
			case "volume-snapshots":
				if query.Get("q") == "parentUuid=snap-a" {
					inventory(w, map[string]any{"uuid": "child", "volumeUuid": "disk-a", "parentUuid": "snap-a"})
				} else {
					inventory(w, map[string]any{"uuid": "snap-a", "volumeUuid": "disk-a", "status": "Ready"})
				}
			default:
				t.Errorf("unexpected read %s", path)
				http.NotFound(w, r)
			}
			return
		}
		writes = append(writes, r.Method+" "+path)
		switch path {
		case "vm-instances/nics/nic-exact":
			if r.Method != http.MethodDelete || query.Get("deleteMode") != "Permissive" {
				t.Errorf("unexpected NIC detach request %s %s", r.Method, r.URL)
			}
			nicPresent = false
		case "vm-instances/vm-target/iso":
			if r.Method != http.MethodDelete || query.Get("isoUuid") != "image-iso" {
				t.Errorf("unexpected ISO eject request %s %s", r.Method, r.URL)
			}
			iso = ""
		default:
			t.Errorf("unexpected write %s %s", r.Method, r.URL)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	check := func(command []string, operation, uuid, status, failure string) {
		t.Helper()
		flags := []string{"--uuid-only", "--yes", "-o", "json", "--error-format", "json"}
		out, stderr, err := invokeCLI(t, cfg, "", append(command, flags...)...)
		if failure != "" {
			if err == nil || !strings.Contains(stderr, `"code":"`+failure+`"`) {
				t.Fatalf("%v: expected %s, got %v %s", command, failure, err, stderr)
			}
			return
		}
		var result struct {
			Operation string
			UUID      string `json:"uuid"`
			Status    string
		}
		if err != nil || json.Unmarshal([]byte(out), &result) != nil || result.Operation != operation ||
			result.UUID != uuid || result.Status != status {
			t.Fatalf("%v: unexpected result %v %s %s", command, err, out, stderr)
		}
	}

	check([]string{"delete", "volume", "root-a"}, "", "", "", "conflict")
	check([]string{"volume", "detach", "disk-a", "--instance", "vm-other"}, "", "", "", "conflict")
	check([]string{"volume", "resize", "disk-a", "--size", "1G"}, "", "", "", "conflict")
	check([]string{"snapshot", "revert", "snap-a", "--volume", "disk-other"}, "", "", "", "conflict")
	check([]string{"delete", "snapshot", "snap-a", "--volume", "disk-a"}, "", "", "", "conflict")
	check([]string{"instance", "detach-l3", "vm-target", "--nic", "nic-exact", "--l3-network", "l3-other"}, "", "", "", "conflict")
	check([]string{"instance", "detach-l3", "vm-other", "--nic", "nic-exact"}, "", "", "", "conflict")
	check([]string{"instance", "detach-l3", "vm-target", "--nic", "nic-exact", "--dry-run"}, "detach-nic", "", "skipped", "")
	check([]string{"instance", "detach-l3", "vm-target", "--nic", "nic-exact"}, "detach-nic", "nic-exact", "succeeded", "")
	check([]string{"instance", "detach-iso", "vm-target", "--cdrom", "drive-a"}, "detach-iso", "drive-a", "succeeded", "")
	check([]string{"instance", "detach-iso", "vm-target", "--cdrom", "drive-a"}, "detach-iso", "drive-a", "skipped", "")

	mutex.Lock()
	defer mutex.Unlock()
	if !slices.Equal(writes, []string{"DELETE vm-instances/nics/nic-exact", "DELETE vm-instances/vm-target/iso"}) {
		t.Fatalf("unexpected writes: %v", writes)
	}
}

func TestLegacyISOEjectUsesISOEndpoint(t *testing.T) {
	const vm = "ab45398daffb4adba041953e6e0890b7"
	const iso = "d47e1bbf83114f75929abcf38a404b4c"
	var mutex sync.Mutex
	var deletes []string
	machine := map[string]any{
		"uuid": vm, "name": "test-vm", "state": "Stopped",
		"vmCdRoms": []map[string]any{{"uuid": "drive-a", "isoUuid": iso}},
	}
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		path := strings.TrimPrefix(r.URL.Path, "/zstack/v1/")
		switch {
		case r.Method == http.MethodGet && (path == "vm-instances" || path == "vm-instances/"+vm):
			inventory(w, machine)
		case r.Method == http.MethodGet && (path == "images" || path == "images/"+iso):
			inventory(w, map[string]any{"uuid": iso, "name": "test-iso", "mediaType": "ISO"})
		case r.Method == http.MethodDelete:
			deletes = append(deletes, path)
			if path != "vm-instances/"+vm+"/iso" || r.URL.Query().Get("isoUuid") != iso {
				http.Error(w, "wrong endpoint", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected API request %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "instance", "detach-iso", vm, "--iso", iso, "--yes")
	if err != nil || !strings.Contains(out, "detached iso/") {
		t.Fatalf("legacy eject failed: %v %s %s", err, out, stderr)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if !slices.Equal(deletes, []string{"vm-instances/" + vm + "/iso"}) {
		t.Fatalf("unexpected DELETE paths: %v", deletes)
	}
}

func TestExactCreateVolumeVerifiesRequestedProperties(t *testing.T) {
	const size int64 = 2 << 30
	for _, tc := range []struct {
		name, offering, storage string
		flags                   []string
		size                    int64
		fails                   bool
	}{
		{"explicit size", "", "", []string{"--size", "2G"}, size, false},
		{"wrong size", "", "", []string{"--size", "2G"}, 1 << 30, true},
		{"missing size", "", "", []string{"--size", "2G"}, 0, true},
		{"offering and storage", "offering-a", "storage-a", []string{"--disk-offering", "offering-a", "--primary-storage", "storage-a"}, size, false},
		{"wrong offering size", "offering-a", "storage-a", []string{"--disk-offering", "offering-a", "--primary-storage", "storage-a"}, 1 << 30, true},
		{"wrong offering", "offering-b", "storage-a", []string{"--disk-offering", "offering-a", "--primary-storage", "storage-a"}, size, true},
		{"missing offering", "", "storage-a", []string{"--disk-offering", "offering-a", "--primary-storage", "storage-a"}, size, true},
		{"wrong storage", "offering-a", "storage-b", []string{"--disk-offering", "offering-a", "--primary-storage", "storage-a"}, size, true},
		{"missing storage", "", "", []string{"--size", "2G", "--primary-storage", "storage-a"}, size, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.Method + " " + strings.TrimPrefix(r.URL.Path, "/zstack/v1/") {
				case "GET disk-offerings":
					inventory(w, map[string]any{"uuid": "offering-a", "diskSize": size})
				case "GET primary-storage":
					inventory(w, map[string]any{"uuid": "storage-a"})
				case "POST volumes/data":
					writes.Add(1)
					w.WriteHeader(http.StatusNoContent)
				case "GET volumes":
					if writes.Load() == 0 {
						inventory(w)
					} else {
						inventory(w, map[string]any{"uuid": "disk-a", "name": "new-disk", "type": "Data",
							"status": "NotInstantiated", "size": tc.size,
							"diskOfferingUuid": tc.offering, "primaryStorageUuid": tc.storage})
					}
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
					http.NotFound(w, r)
				}
			})
			args := []string{"create", "volume", "new-disk", "--resource-uuid", "disk-a", "--uuid-only", "--yes", "-o", "json", "--error-format", "json"}
			out, stderr, err := invokeCLI(t, cfg, "", append(args, tc.flags...)...)
			assertDeviceReadback(t, out, stderr, err, tc.fails, writes.Load())
		})
	}
}

func TestExactNICAttachVerifiesMACAndStaticIP(t *testing.T) {
	const mac = "02:00:5E:10:00:AB"
	for _, tc := range []struct {
		name, readMAC, readIP, secondaryIP, secondaryL3 string
		flags                                           []string
		fails                                           bool
	}{
		{name: "defaults allowed"},
		{name: "MAC case normalized", readMAC: "02:00:5e:10:00:ab", flags: []string{"--mac", mac}},
		{name: "wrong MAC", readMAC: "02:00:5e:10:00:ac", flags: []string{"--mac", mac}, fails: true},
		{name: "missing MAC", flags: []string{"--mac", mac}, fails: true},
		{name: "IPv4 matches", readIP: "192.0.2.11", flags: []string{"--static-ip", "192.0.2.11"}},
		{name: "wrong IP", readIP: "192.0.2.12", flags: []string{"--static-ip", "192.0.2.11"}, fails: true},
		{name: "missing IP", flags: []string{"--static-ip", "192.0.2.11"}, fails: true},
		{name: "IPv6 normalized in usedIps", readIP: "192.0.2.11", secondaryIP: "2001:db8::12", secondaryL3: "l3-a", flags: []string{"--static-ip", "2001:0db8:0:0::12"}},
		{name: "IPv6 from another L3 ignored", secondaryIP: "2001:db8::12", secondaryL3: "l3-b", flags: []string{"--static-ip", "2001:db8::12"}, fails: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var writes atomic.Int32
			cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.Method + " " + strings.TrimPrefix(r.URL.Path, "/zstack/v1/") {
				case "GET vm-instances":
					inventory(w, map[string]any{"uuid": "vm-a", "state": "Stopped"})
				case "GET l3-networks":
					inventory(w, map[string]any{"uuid": "l3-a"})
				case "POST vm-instances/vm-a/l3-networks/l3-a":
					writes.Add(1)
					if slices.Contains(tc.flags, "--mac") {
						var body struct{ Params struct{ CustomMac string } }
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Params.CustomMac != strings.ToLower(mac) {
							t.Errorf("MAC was not sent in canonical form: %v", err)
						}
					}
					w.Write([]byte(`{"inventory":{"uuid":"vm-a","vmNics":[{"uuid":"nic-a","vmInstanceUuid":"vm-a","l3NetworkUuid":"l3-a","mac":"02:00:5e:10:00:ab","ip":"192.0.2.11"}]}}`))
				case "GET vm-instances/nics":
					if r.URL.Query().Get("q") == "uuid=nic-a" {
						inventory(w, map[string]any{"uuid": "nic-a", "vmInstanceUuid": "vm-a", "l3NetworkUuid": "l3-a", "mac": tc.readMAC, "ip": tc.readIP,
							"usedIps": []map[string]any{{"ip": tc.secondaryIP, "l3NetworkUuid": tc.secondaryL3}}})
					} else {
						inventory(w)
					}
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
					http.NotFound(w, r)
				}
			})
			args := []string{"instance", "attach-l3", "vm-a", "--l3-network", "l3-a", "--uuid-only", "--yes", "-o", "json", "--error-format", "json"}
			out, stderr, err := invokeCLI(t, cfg, "", append(args, tc.flags...)...)
			assertDeviceReadback(t, out, stderr, err, tc.fails, writes.Load())
		})
	}
}

func assertDeviceReadback(t *testing.T, out, stderr string, err error, fails bool, writes int32) {
	t.Helper()
	if writes != 1 {
		t.Fatalf("expected one write, got %d: %v %s", writes, err, stderr)
	}
	if fails {
		assertDiagnostic(t, stderr, err, "api_failed", 8)
		if out != "" {
			t.Fatalf("unverified write emitted success: %s", out)
		}
		return
	}
	if err != nil || !strings.Contains(out, `"status":"succeeded"`) {
		t.Fatalf("verified write failed: %v %s %s", err, out, stderr)
	}
}
