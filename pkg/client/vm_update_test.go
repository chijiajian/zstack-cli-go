package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

func TestUpdateVMInstanceSendsOnlyRequestedFields(t *testing.T) {
	const uuid = "0123456789abcdef0123456789abcdef"
	memory := int64(8 << 30)
	emptyDescription := ""
	for _, tc := range []struct {
		name   string
		detail param.UpdateVmInstanceParamDetail
	}{
		{"memory only", param.UpdateVmInstanceParamDetail{MemorySize: &memory}},
		{"rename", param.UpdateVmInstanceParamDetail{Name: "new-name"}},
		{"clear description", param.UpdateVmInstanceParamDetail{Description: &emptyDescription}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := param.UpdateVmInstanceParam{Params: tc.detail}
			expected, err := json.Marshal(params)
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]map[string]json.RawMessage
			if err := json.Unmarshal(expected, &want); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.URL.Path != "/zstack/v1/vm-instances/"+uuid+"/actions" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				body, err := io.ReadAll(r.Body)
				var got map[string]map[string]json.RawMessage
				if err != nil || json.Unmarshal(body, &got) != nil || !reflect.DeepEqual(got, want) {
					t.Errorf("request body = %s, want %s", body, expected)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"inventory":{"uuid":"`+uuid+`","name":"example-vm"}}`)
			}))
			defer server.Close()
			cfg, err := NewSDKConfig(strings.TrimPrefix(server.URL, "http://"))
			if err != nil {
				t.Fatal(err)
			}
			result, err := UpdateVMInstance(t.Context(), sdk.NewZSClient(cfg.AccessKey("fake", "fake")), uuid, params)
			if err != nil {
				t.Fatal(err)
			}
			if result.UUID != uuid || result.Name != "example-vm" {
				t.Fatalf("unexpected update result: %+v", result)
			}
		})
	}
}
