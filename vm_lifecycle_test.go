package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestExactVMLifecycle(t *testing.T) {
	const id = "vm-target"
	var mutex sync.Mutex
	state, policy, writes := "Stopped", "Delay", 0
	rebooted := false
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		path := strings.TrimPrefix(r.URL.Path, "/zstack/v1/")
		if r.Method == http.MethodGet {
			switch path {
			case "vm-instances":
				q := r.URL.Query().Get("q")
				if q == "uuid="+id {
					inventory(w, map[string]any{"uuid": id, "name": "vm", "state": state})
				} else if q == "uuid=wrong-identity" {
					inventory(w, map[string]any{"uuid": id, "name": "vm", "state": state})
				} else {
					if !strings.HasPrefix(q, "uuid=") {
						t.Errorf("unexpected name fallback: %s", q)
					}
					inventory(w)
				}
			case "global-configurations":
				if q := r.URL.Query()["q"]; !slices.Equal(q, []string{"category=vm", "name=deletionPolicy"}) {
					t.Errorf("wrong configuration query: %s", q)
				}
				inventory(w, map[string]any{"category": "vm", "name": "deletionPolicy", "value": policy})
			default:
				t.Errorf("unexpected GET %s", path)
				http.NotFound(w, r)
			}
			return
		}
		writes++
		if r.Method == http.MethodDelete && path == "vm-instances/"+id {
			if r.URL.Query().Get("deleteMode") != "Permissive" || state != "Stopped" {
				t.Errorf("unexpected delete: %s %s", r.URL, state)
			}
			state = "Destroyed"
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var body map[string]map[string]any
		if r.Method != http.MethodPut || path != "vm-instances/"+id+"/actions" || json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 1 {
			t.Errorf("unexpected write: %s %s %v", r.Method, r.URL, body)
			http.Error(w, "bad request", 400)
			return
		}
		for operation, detail := range body {
			switch operation {
			case "updateVmInstance":
				if !reflect.DeepEqual(detail, map[string]any{"description": ""}) {
					t.Errorf("unexpected update fields: %v", detail)
				}
			case "pauseVmInstance":
				state = "Paused"
			case "resumeVmInstance":
				state = "Running"
			case "rebootVmInstance":
				rebooted = true
			case "stopVmInstance":
				if !reflect.DeepEqual(detail, map[string]any{"type": "grace", "stopHA": "true"}) {
					t.Errorf("unexpected stop fields: %v", detail)
				}
				state = "Stopped"
			case "recoverVmInstance":
				state = "Stopped"
			default:
				t.Errorf("unexpected action: %s", operation)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"inventory": map[string]any{"uuid": id, "name": "vm", "state": state}})
	})
	check := func(verb string, args []string, status, code string) {
		t.Helper()
		group, sub := "instance", verb
		if verb == "delete" {
			group, sub = "delete", "instance"
		}
		base := []string{group, sub, id, "--uuid-only", "--yes", "-o", "json", "--error-format", "json"}
		out, stderr, err := invokeCLI(t, cfg, "", append(base, args...)...)
		if code != "" && (err == nil || !strings.Contains(stderr, `"code":"`+code+`"`)) || code == "" && err != nil {
			t.Fatalf("%s: unexpected exit: %v %s %s", verb, err, out, stderr)
		}
		if status != "" {
			var response struct {
				Operation string
				Items     []struct{ UUID, Status string }
			}
			if json.Unmarshal([]byte(out), &response) != nil || response.Operation != verb || len(response.Items) != 1 ||
				response.Items[0].UUID != id || response.Items[0].Status != status {
				t.Fatalf("%s: invalid result %s", verb, out)
			}
		}
	}
	check("update", []string{"--description="}, "succeeded", "")
	check("update", []string{"--state", "Running"}, "", "invalid_arguments")
	check("delete", []string{"--dry-run"}, "skipped", "")
	mutex.Lock()
	policy = "Direct"
	mutex.Unlock()
	check("delete", nil, "failed", "conflict")
	mutex.Lock()
	policy, state = "Delay", "Running"
	mutex.Unlock()
	check("update", []string{"--cpu", "2"}, "failed", "conflict")
	check("pause", nil, "succeeded", "")
	check("pause", nil, "skipped", "")
	check("resume", nil, "succeeded", "")
	check("resume", nil, "skipped", "")
	check("restart", nil, "succeeded", "")
	check("stop", []string{"--stop-ha"}, "succeeded", "")
	check("delete", nil, "succeeded", "")
	out, stderr, err := invokeCLI(t, cfg, "", "wait", "instance", id, "--uuid-only", "--for", "Destroyed", "-o", "json")
	if err != nil || !strings.Contains(out, `"state":"Destroyed"`) {
		t.Fatalf("Destroyed wait: %v %s %s", err, out, stderr)
	}
	_, stderr, err = invokeCLI(t, cfg, "", "wait", "instance", id, "--uuid-only", "--for", "Running", "--error-format", "json")
	if err == nil || !strings.Contains(stderr, `"code":"conflict"`) {
		t.Fatalf("terminal wait: %v %s", err, stderr)
	}
	check("recover", nil, "succeeded", "")
	_, stderr, err = invokeCLI(t, cfg, "", "instance", "recover", "wrong-identity", "--uuid-only", "--yes", "-o", "json", "--error-format", "json")
	if err == nil || !strings.Contains(stderr, `"code":"api_failed"`) {
		t.Fatalf("identity mismatch: %v %s", err, stderr)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if writes != 7 || !rebooted || state != "Stopped" {
		t.Fatalf("writes=%d rebooted=%v state=%s", writes, rebooted, state)
	}
}

func TestExactVMCreate(t *testing.T) {
	const id = "11111111222233334444555555555555"
	var mutex sync.Mutex
	writes := 0
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		path := strings.TrimPrefix(r.URL.Path, "/zstack/v1/")
		switch {
		case r.Method == http.MethodPost && path == "vm-instances":
			writes++
			var body struct{ Params map[string]any }
			want := map[string]any{"name": "-example", "imageUuid": "image", "l3NetworkUuids": []any{"network"},
				"resourceUuid": id, "strategy": "CreateStopped", "cpuNum": float64(2), "memorySize": float64(1610612736)}
			if json.NewDecoder(r.Body).Decode(&body) != nil || !reflect.DeepEqual(body.Params, want) {
				t.Errorf("wrong create body: %v", body.Params)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"inventory": map[string]any{"uuid": id, "state": "Stopped"}})
		case r.Method == http.MethodGet && path == "images/image":
			inventory(w, map[string]any{"uuid": "image", "mediaType": "RootVolumeTemplate"})
		case r.Method == http.MethodGet && path == "l3-networks" && r.URL.Query().Get("q") == "uuid=network":
			inventory(w, map[string]any{"uuid": "network"})
		case r.Method == http.MethodGet && path == "vm-instances" && r.URL.Query().Get("q") == "uuid="+id:
			if writes == 0 {
				inventory(w)
			} else {
				inventory(w, map[string]any{"uuid": id, "state": "Stopped"})
			}
		default:
			t.Errorf("unexpected create request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	})
	args := []string{"--error-format", "json", "create", "instance", "--uuid-only", "--yes", "-o", "json",
		"--resource-uuid", id, "--strategy", "CreateStopped", "--cpu", "2", "--memory", "1536M", "--image", "image", "--l3-network", "network", "--", "-example"}
	out, stderr, err := invokeCLI(t, cfg, "", args...)
	if err != nil || !strings.Contains(out, `"status":"succeeded"`) {
		t.Fatalf("create: %v %s %s", err, out, stderr)
	}
	_, stderr, err = invokeCLI(t, cfg, "", args...)
	if err == nil || !strings.Contains(stderr, `"code":"conflict"`) {
		t.Fatalf("repeated UUID: %v %s", err, stderr)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if writes != 1 {
		t.Fatalf("create submitted %d times", writes)
	}
}
