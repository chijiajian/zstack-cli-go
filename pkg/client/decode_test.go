package client

import (
	"encoding/json"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
	"testing"
	"time"
)

func TestDecodeZStackDatesWithoutChangingStringsOrNumbers(t *testing.T) {
	var result struct {
		Items *[]view.VolumeInventoryView `json:"inventories"`
	}
	data := []byte(`{"inventories":[{"uuid":"id","name":"Jul 2, 2026 10:46:38 AM","createDate":"Jul 2, 2026 10:46:38 AM","lastOpDate":"2026-07-02T02:46:38Z","size":9007199254740993}]}`)
	if err := decodeResponse(data, &result); err != nil {
		t.Fatal(err)
	}
	item := (*result.Items)[0]
	if item.Size != 9007199254740993 || item.Name != "Jul 2, 2026 10:46:38 AM" || item.CreateDate.UTC().Format(time.RFC3339) != "2026-07-02T02:46:38Z" {
		t.Fatalf("incorrect decoding: %+v", item)
	}
	var raw map[string]any
	if err := decodeResponse(data, &raw); err != nil {
		t.Fatal(err)
	}
	row := raw["inventories"].([]any)[0].(map[string]any)
	if row["size"] != json.Number("9007199254740993") || row["createDate"] != "Jul 2, 2026 10:46:38 AM" {
		t.Fatal("untyped detail altered")
	}
	var invalid struct {
		Valid *bool `json:"valid"`
	}
	if err := decodeResponse([]byte(`{"valid":"true"}`), &invalid); err == nil {
		t.Fatal("accepted invalid boolean")
	}
}

func TestDecodeCompatibleTimeShapes(t *testing.T) {
	var result struct {
		CreateDate *time.Time            `json:"createDate"`
		History    map[string]*time.Time `json:"history"`
		Missing    *time.Time            `json:"missing"`
	}
	data := []byte(`{"CreateDate":"Jul 2, 2026 10:46:38 AM","history":{"last":"Jul 2, 2026 10:46:38 AM","empty":null},"missing":null}`)
	if err := decodeResponse(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.CreateDate == nil || result.History["last"] == nil || !result.CreateDate.Equal(*result.History["last"]) || result.History["empty"] != nil || result.Missing != nil {
		t.Fatalf("unexpected time shapes: %+v", result)
	}
}
