package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

func TestWaitAccessKeySignatureAndResourceMapping(t *testing.T) {
	for resource, path := range map[string]string{"instance": "vm-instances", "image": "images", "volume": "volumes", "long-job": "longjobs"} {
		t.Run(resource, func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/zstack/v1/"+path || r.URL.Query().Get("q") != "uuid=id1" {
					t.Errorf("unexpected URL %s", r.URL)
				}
				if _, err := http.ParseTime(r.Header.Get("Date")); err != nil {
					t.Error(err)
				}
				mac := hmac.New(sha1.New, []byte("secret"))
				mac.Write([]byte("GET\n" + r.Header.Get("Date") + "\n" + strings.TrimPrefix(r.URL.Path, "/zstack")))
				expected := "ZStack key:" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
				if r.Header.Get("Authorization") != expected {
					t.Error("invalid access key signature")
				}
				json.NewEncoder(w).Encode(map[string]any{"inventories": []map[string]string{{"uuid": "id1", "name": "demo", "state": "Running", "status": "Ready"}}})
			}))
			defer server.Close()
			cfg := &config.ZStackConfig{CurrentContext: "test", Contexts: map[string]config.Context{"test": {Endpoint: strings.TrimPrefix(server.URL, "http://"), AuthType: config.AuthTypeAccessKey, AccessKeyID: "key", AccessKeySecret: "secret"}}}
			cli, err := newAuthenticatedHTTPClient(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			resolved, err := cli.Resolve(context.Background(), resource, "id1", ResolveUUID)
			if err != nil || resolved.UUID != "id1" || resolved.State != "Running" || calls != 1 {
				t.Fatalf("resource=%+v calls=%d err=%v", resolved, calls, err)
			}
		})
	}
}

func TestWaitCancellationInterruptsBodyRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer server.Close()
	cli := &ReadClient{endpoint: server.URL + "/", credentials: config.Context{AuthType: config.AuthTypeAccessKey, AccessKeyID: "fake", AccessKeySecret: "fake"}, httpClient: server.Client()}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := cli.Resolve(ctx, "instance", "demo", ResolveAuto)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatalf("elapsed=%s err=%v", time.Since(start), err)
	}
}

func TestReadRejectsTrailingJSON(t *testing.T) {
	for _, body := range []string{`{"inventories":[]} {"inventories":[{"uuid":"extra"}]}`, `{"inventories":[]} garbage`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
		cli := &ReadClient{endpoint: server.URL + "/", credentials: config.Context{AuthType: config.AuthTypeAccessKey, AccessKeyID: "fake", AccessKeySecret: "fake"}, httpClient: server.Client()}
		var result struct {
			Inventories []map[string]any `json:"inventories"`
		}
		err := cli.Read(context.Background(), "v1/zones", nil, &result)
		server.Close()
		if err == nil {
			t.Fatalf("accepted trailing content: %s", body)
		}
	}
}

func TestResolverRejectsInventoriesWithoutIdentity(t *testing.T) {
	for _, body := range []string{`{"inventories":[null]}`, `{"inventories":[{}]}`, `{"inventories":[{"uuid":" "}]}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
			defer server.Close()
			cli := &ReadClient{endpoint: server.URL + "/", httpClient: server.Client()}
			result, err := cli.Resolve(context.Background(), "instance", "target", ResolveAuto)
			if err == nil || result.UUID != "" {
				t.Fatalf("malformed resource accepted: %+v, %v", result, err)
			}
		})
	}
}
