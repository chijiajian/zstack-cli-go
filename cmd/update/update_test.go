package update

import (
	"testing"

	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

func TestPreserveCurrentNameFillsEmptyName(t *testing.T) {
	description := "updated"
	params := param.UpdateZoneParam{
		Params: param.UpdateZoneParamDetail{
			Description: &description,
		},
	}

	preserveCurrentName(&params, "zone-01")

	if params.Params.Name != "zone-01" {
		t.Fatalf("expected current name to be preserved, got %q", params.Params.Name)
	}
}

func TestPreserveCurrentNameKeepsExplicitName(t *testing.T) {
	params := param.UpdateZoneParam{
		Params: param.UpdateZoneParamDetail{
			Name: "new-zone",
		},
	}

	preserveCurrentName(&params, "zone-01")

	if params.Params.Name != "new-zone" {
		t.Fatalf("expected explicit name to be kept, got %q", params.Params.Name)
	}
}
