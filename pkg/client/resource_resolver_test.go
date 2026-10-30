package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
)

func TestScalarHelpersRejectAmbiguityAndFuzzyMatches(t *testing.T) {
	helpers := map[string]func(*sdk.ZSClient, string) (string, error){"zone": GetZoneUUIDByName, "cluster": GetClusterUUIDByName, "host": GetHostUUIDByName, "backup-storage": GetBackupStorageUUIDByName, "image": GetImageUUIDByName, "instance-offering": GetInstanceOfferingUUIDByName, "l3-network": GetL3NetworkUUIDByName, "primary-storage": GetPrimaryStorageUUIDByName}
	for name, resolve := range helpers {
		t.Run(name, func(t *testing.T) {
			for _, scenario := range []string{"uuid", "exact", "ambiguous", "similar"} {
				t.Run(scenario, func(t *testing.T) {
					var queries []string
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						q := r.URL.Query().Get("q")
						queries = append(queries, q)
						objects := []map[string]string{}
						if (scenario == "uuid" && q == "uuid=target") || (scenario == "exact" && q == "name=target") {
							objects = append(objects, map[string]string{"uuid": "chosen", "name": "target", "state": "Enabled", "status": "Ready"})
						}
						if scenario == "ambiguous" && q == "name=target" {
							objects = append(objects, map[string]string{"uuid": "a"}, map[string]string{"uuid": "b"})
						}
						if strings.Contains(q, "~=") {
							objects = append(objects, map[string]string{"uuid": "wrong", "state": "Enabled", "status": "Ready"})
						}
						json.NewEncoder(w).Encode(map[string]any{"inventories": objects})
					}))
					defer server.Close()
					cfg, err := NewSDKConfig(strings.TrimPrefix(server.URL, "http://"))
					if err != nil {
						t.Fatal(err)
					}
					id, err := resolve(sdk.NewZSClient(cfg.AccessKey("fake", "fake")), "target")
					if scenario == "uuid" || scenario == "exact" {
						if err != nil || id != "chosen" {
							t.Fatalf("id=%q err=%v", id, err)
						}
					} else if err == nil {
						t.Fatalf("expected error, got %q", id)
					}
					if len(queries) == 0 || queries[0] != "uuid=target" || strings.Contains(strings.Join(queries, " "), "~=") {
						t.Fatalf("unsafe queries: %v", queries)
					}
				})
			}
		})
	}
}

func TestBulkOperationHelpersRejectMissingUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"inventories": []map[string]string{{"state": "Running", "status": "Ready"}}})
	}))
	defer server.Close()
	cfg, err := NewSDKConfig(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	cli := sdk.NewZSClient(cfg.AccessKey("fake", "fake"))
	checks := map[string]func() error{
		"ready images":        func() error { _, err := GetReadyImagesByNameOrUUID(cli, "target"); return err },
		"deleted images":      func() error { _, err := GetDeletedImagesByNameOrUUID(cli, "target"); return err },
		"active instances":    func() error { _, err := GetReadyVMsByNameOrUUID(cli, "target"); return err },
		"destroyed instances": func() error { _, err := GetDestroyedVMsByNameOrUUID(cli, "target"); return err },
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			if err := check(); err == nil {
				t.Fatal("accepted malformed identity")
			}
		})
	}
}

func TestResolverDiscardsResultAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	query := func(_ context.Context, _ string, _ *param.QueryParam, result any) error {
		cancel()
		return json.Unmarshal([]byte(`[{"uuid":"id1"}]`), result)
	}
	result, err := resolveResource(ctx, query, "instance", "id1", ResolveOptions{})
	if !errors.Is(err, context.Canceled) || len(result) != 0 {
		t.Fatalf("canceled result accepted: %+v %v", result, err)
	}
}
