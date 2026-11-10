package client

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	sdk "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
)

// Resolve once per client, without changing the process or system time zone.
func accessKeyTimezone() (*time.Location, error) {
	name := os.Getenv("ZSTACK_API_TIMEZONE")
	if name == "" {
		return time.Local, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("invalid ZSTACK_API_TIMEZONE %q: use an IANA time zone such as Europe/Berlin or Asia/Shanghai: %w", name, err)
	}
	return loc, nil
}

func signAccessKeyRequest(req *http.Request, key, secret, path string, now time.Time, loc *time.Location) {
	if loc == nil {
		loc = time.Local
	}
	now = now.In(loc)
	// Preserve SDK compatibility for numeric zone names used in these regions.
	name, offset := now.Zone()
	if strings.HasPrefix(name, "+") || strings.HasPrefix(name, "-") {
		if offset == 7*3600 {
			now = now.In(time.FixedZone("ICT", offset))
		}
		if offset == 6*3600+30*60 {
			now = now.In(time.FixedZone("MMT", offset))
		}
	}
	date := now.Format(time.RFC1123)
	if loc == time.UTC {
		date = now.Format(http.TimeFormat)
	}
	mac := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(mac, "%s\n%s\n%s", req.Method, date, path)
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", "ZStack "+key+":"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}

func accessKeySDKConfig(cfg *sdk.ZSConfig, key, secret string, loc *time.Location) *sdk.ZSConfig {
	// SDK v0.0.10 exposes no signer/transport hook. Its proxy callback runs
	// before sending each request, including async polls. Re-sign here to keep
	// Date and Authorization consistent with our cancellable HTTP client.
	// Returning nil preserves the SDK's existing direct connection policy.
	return cfg.AccessKey(key, secret).ProxyFunc(func(req *http.Request) (*url.URL, error) {
		if strings.HasPrefix(req.Header.Get("Authorization"), "ZStack "+key+":") && strings.HasPrefix(req.URL.Path, "/zstack/") {
			signAccessKeyRequest(req, key, secret, strings.TrimPrefix(req.URL.EscapedPath(), "/zstack"), time.Now(), loc)
		}
		return nil, nil
	})
}
