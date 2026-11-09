package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	sdk "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
)

func TestAccessKeyDateTimezones(t *testing.T) {
	for _, tc := range []struct{ zone, instant, want string }{
		{"Asia/Shanghai", "2026-11-09T20:05:06Z", "Tue, 10 Nov 2026 04:05:06 CST"},
		{"Europe/Berlin", "2026-11-09T20:05:06Z", "Mon, 09 Nov 2026 21:05:06 CET"},
		{"Europe/Berlin", "2026-07-09T20:05:06Z", "Thu, 09 Jul 2026 22:05:06 CEST"},
		{"Asia/Singapore", "2026-11-09T20:05:06Z", "Tue, 10 Nov 2026 04:05:06 +08"},
		{"Asia/Kuala_Lumpur", "2026-11-09T20:05:06Z", "Tue, 10 Nov 2026 04:05:06 +08"},
		{"Asia/Dubai", "2026-11-09T20:05:06Z", "Tue, 10 Nov 2026 00:05:06 +04"},
		{"UTC", "2026-11-09T20:05:06Z", "Mon, 09 Nov 2026 20:05:06 GMT"},
	} {
		t.Run(tc.zone+tc.instant, func(t *testing.T) {
			t.Setenv("ZSTACK_API_TIMEZONE", tc.zone)
			loc, err := accessKeyTimezone()
			if err != nil {
				t.Fatal(err)
			}
			instant, _ := time.Parse(time.RFC3339, tc.instant)
			req, _ := http.NewRequest("GET", "http://example.invalid/zstack/v1/zones", nil)
			signAccessKeyRequest(req, "key", "secret", "/v1/zones", instant, loc)
			if got := req.Header.Get("Date"); got != tc.want {
				t.Fatalf("Date = %q, want %q", got, tc.want)
			}
			parsed, err := time.ParseInLocation(time.RFC1123, req.Header.Get("Date"), loc)
			if err != nil || !parsed.Equal(instant) {
				t.Fatalf("changed instant: %v, %v", parsed, err)
			}
		})
	}
}

func TestAccessKeyTimezoneConfiguration(t *testing.T) {
	t.Setenv("ZSTACK_API_TIMEZONE", "")
	loc, err := accessKeyTimezone()
	if err != nil || loc != time.Local {
		t.Fatalf("default = %v, %v", loc, err)
	}
	t.Setenv("ZSTACK_API_TIMEZONE", "not/a/timezone")
	if _, err := newHTTPClient(config.Context{Endpoint: "example.invalid", AuthType: config.AuthTypeAccessKey}); err == nil {
		t.Fatal("invalid zone accepted")
	}
	if _, err := newHTTPClient(config.Context{Endpoint: "example.invalid", AuthType: config.AuthTypeAccount}); err != nil {
		t.Fatalf("session affected: %v", err)
	}
}

// Model deployments requiring their configured zone and a matching signature.
// Exercise real HTTP requests from both CLI transport paths.
func TestAccessKeyTimezoneCompatibility(t *testing.T) {
	for _, zone := range []string{"Asia/Shanghai", "Asia/Singapore", "Asia/Kuala_Lumpur", "Europe/Berlin", "Asia/Dubai"} {
		t.Run(zone, func(t *testing.T) {
			t.Setenv("ZSTACK_API_TIMEZONE", zone)
			loc, err := accessKeyTimezone()
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				date := r.Header.Get("Date")
				parsed, err := time.ParseInLocation(time.RFC1123, date, loc)
				if err != nil || !strings.HasSuffix(date, " "+time.Now().In(loc).Format("MST")) || time.Since(parsed).Abs() > 5*time.Second {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"error":{"code":"requestTimeTooSkewed"}}`))
					return
				}
				mac := hmac.New(sha1.New, []byte("secret"))
				mac.Write([]byte(r.Method + "\n" + date + "\n" + strings.TrimPrefix(r.URL.EscapedPath(), "/zstack")))
				if r.Header.Get("Authorization") != "ZStack key:"+base64.StdEncoding.EncodeToString(mac.Sum(nil)) {
					t.Error("signature does not match the outgoing Date/method/path")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				json.NewEncoder(w).Encode(map[string]any{"inventories": []any{}})
			}))
			defer server.Close()
			endpoint := strings.TrimPrefix(server.URL, "http://")
			reader, err := newHTTPClient(config.Context{Endpoint: endpoint, AuthType: config.AuthTypeAccessKey, AccessKeyID: "key", AccessKeySecret: "secret"})
			if err != nil {
				t.Fatal(err)
			}
			// Confirm this fixture rejects the old GMT request.
			req, _ := http.NewRequest("GET", server.URL+"/zstack/v1/zones", nil)
			req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("GMT status: %d", resp.StatusCode)
			}
			for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
				var result map[string]any
				if err := reader.request(context.Background(), method, "v1/zones", nil, nil, &result); err != nil {
					t.Fatalf("HTTP %s: %v", method, err)
				}
			}
			cfg, err := NewSDKConfig(endpoint)
			if err != nil {
				t.Fatal(err)
			}
			cli := sdk.NewZSClient(accessKeySDKConfig(cfg, "key", "secret", reader.signingLocation))
			for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
				var result map[string]any
				switch method {
				case "GET":
					err = cli.GetWithRespKey("v1/zones", "", "", nil, &result)
				case "POST":
					err = cli.PostWithRespKey(context.Background(), "v1/zones", "", map[string]any{}, &result)
				case "PUT":
					err = cli.PutWithRespKey("v1/zones", "", "", map[string]any{}, &result)
				case "DELETE":
					err = cli.DeleteWithBody("v1/zones", map[string]any{}, &result)
				}
				if err != nil {
					t.Fatalf("SDK %s: %v", method, err)
				}
			}
			if calls != 9 {
				t.Fatalf("requests = %d, want 9 (no retries)", calls)
			}
		})
	}
}
