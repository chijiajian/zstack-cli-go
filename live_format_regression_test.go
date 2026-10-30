package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestReadCommandsAcceptZStackTimestampFormat(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		inventory(w, map[string]any{"uuid": "id1", "name": "demo", "state": "Succeeded", "createDate": "Jul 2, 2026 10:46:38 AM", "lastOpDate": "Oct 25, 2026 2:50:49 PM"})
	})
	for _, args := range [][]string{
		{"get", "zones", "-o", "json"},
		{"describe", "job", "id1", "--fields", "uuid,createDate", "-o", "json"},
		{"wait", "job", "id1", "--timeout", "1s", "-o", "json"},
	} {
		out, stderr, err := invokeCLI(t, cfg, "", args...)
		if err != nil || !json.Valid([]byte(out)) {
			t.Fatalf("%v: err=%v out=%s stderr=%s", args, err, out, stderr)
		}
	}
}

func TestVolumeDetailQueriesPrimaryStorageByUUID(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/zstack/v1/volumes", "/zstack/v1/volumes/id1":
			inventory(w, map[string]any{"uuid": "id1", "name": "volume", "primaryStorageUuid": "ps1"})
		case "/zstack/v1/primary-storage":
			if r.URL.Query().Get("q") != "uuid=ps1" {
				t.Errorf("incorrect filter: %s", r.URL)
			}
			inventory(w, map[string]any{"uuid": "ps1", "name": "storage"})
		default:
			t.Errorf("unsupported detail endpoint: %s", r.URL)
			http.Error(w, "method not allowed", 405)
		}
	})
	out, stderr, err := invokeCLI(t, cfg, "", "describe", "volume", "id1", "-o", "json")
	var doc struct {
		Related map[string]map[string]any `json:"related"`
	}
	if err != nil || stderr != "" || json.Unmarshal([]byte(out), &doc) != nil || doc.Related["primaryStorage"]["uuid"] != "ps1" {
		t.Fatalf("%v %s %s", err, out, stderr)
	}
}
