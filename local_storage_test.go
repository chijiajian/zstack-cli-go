package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestLocalStoragePreviewAndValidation(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) { t.Errorf("preview contacted API: %s", r.URL) })
	out, stderr, err := invokeCLI(t, cfg, "", "primary-storage", "add-local", "local", "--url", "/data/local", "--zone", "zone", "--dry-run", "-o", "json")
	var body struct {
		Params struct{ Name, Url, ZoneUUID string }
	}
	if err != nil || json.Unmarshal([]byte(out), &body) != nil || body.Params.Url != "/data/local" || body.Params.ZoneUUID != "zone" {
		t.Fatalf("err=%v out=%s stderr=%s", err, out, stderr)
	}
	for _, args := range [][]string{
		{"local", "--url", "relative", "--zone", "zone"},
		{"local", "--url", "/data/local"},
		{" ", "--url", "/data/local", "--zone", "zone"},
	} {
		out, stderr, err = invokeCLI(t, cfg, "", append([]string{"primary-storage", "add-local"}, append(args, "--dry-run")...)...)
		if err == nil || out != "" {
			t.Fatalf("invalid input accepted: %v %s %s", err, out, stderr)
		}
	}
}

func TestLocalStorageTypedSDKRequest(t *testing.T) {
	writes := 0
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			inventory(w, map[string]any{"uuid": "zone-id", "name": "zone"})
			return
		}
		writes++
		if r.Method != http.MethodPost || r.URL.Path != "/zstack/v1/primary-storage/local-storage" {
			t.Errorf("unexpected mutation %s %s", r.Method, r.URL)
		}
		var request struct {
			Params     struct{ Name, Url, ZoneUUID string }
			SystemTags []string
		}
		if json.NewDecoder(r.Body).Decode(&request) != nil || request.Params.ZoneUUID != "zone-id" || request.Params.Url != "/data/local" || len(request.SystemTags) != 1 {
			t.Errorf("wrong request: %+v", request)
		}
		json.NewEncoder(w).Encode(map[string]any{"inventory": map[string]any{"uuid": "storage-id", "name": "local"}})
	})
	out, stderr, err := invokeCLI(t, cfg, "", "primary-storage", "add-local", "local", "--url", "/data/local", "--zone", "zone", "--system-tag", "tag", "-o", "json")
	if err != nil || writes != 1 || !json.Valid([]byte(out)) {
		t.Fatalf("err=%v writes=%d out=%s stderr=%s", err, writes, out, stderr)
	}
}
