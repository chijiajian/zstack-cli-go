// Package diagnostics performs read-only checks without refreshing credentials.
package diagnostics

import (
	"context"
	"fmt"
	"net/url"

	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

type Reader interface {
	Read(context.Context, string, url.Values, any) error
}
type Check struct {
	Name    string `json:"name" yaml:"name"`
	Status  string `json:"status" yaml:"status"`
	Code    string `json:"code,omitempty" yaml:"code,omitempty"`
	Message string `json:"message" yaml:"message"`
}
type Report struct {
	Context string  `json:"context" yaml:"context"`
	OK      bool    `json:"ok" yaml:"ok"`
	Checks  []Check `json:"checks" yaml:"checks"`
}

// Run never returns remote response bodies or credentials in its report.
func Run(ctx context.Context, name string, credentials config.Context, reader Reader, offline bool) (Report, error) {
	report := Report{Context: name, OK: true, Checks: []Check{}}
	var first error
	failed := func(name string, err error) {
		d := clierr.Describe(err)
		report.OK = false
		report.Checks = append(report.Checks, Check{name, "failed", string(d.Code), d.Message})
		if first == nil {
			first = clierr.Wrap(d.Code, fmt.Errorf("doctor check %s failed", name))
		}
	}
	passed := func(name, message string) {
		report.Checks = append(report.Checks, Check{Name: name, Status: "passed", Message: message})
	}
	if credentials.EffectiveAuthType() == config.AuthTypeAccessKey {
		if credentials.AccessKeyID == "" || credentials.AccessKeySecret == "" {
			failed("credentials", clierr.Wrap(clierr.Auth, fmt.Errorf("incomplete keys")))
		} else {
			passed("credentials", "Access key credentials are configured.")
		}
	} else if credentials.EffectiveAuthType() != config.AuthTypeAccount {
		failed("credentials", clierr.Wrap(clierr.Auth, fmt.Errorf("unsupported authentication type")))
	} else if credentials.SessionUUID == "" {
		failed("credentials", clierr.Wrap(clierr.Auth, fmt.Errorf("no saved session; login explicitly")))
	} else {
		passed("credentials", "A saved session is configured; no renewal will be attempted.")
	}
	if offline {
		report.Checks = append(report.Checks, Check{Name: "api", Status: "skipped", Message: "Offline mode; no network requests performed."})
		return report, first
	}
	if first != nil {
		return report, first
	}
	if credentials.EffectiveAuthType() == config.AuthTypeAccount {
		var result struct {
			Valid *bool `json:"valid"`
		}
		err := reader.Read(ctx, "v1/accounts/sessions/"+url.PathEscape(credentials.SessionUUID)+"/valid", nil, &result)
		if err == nil && (result.Valid == nil || !*result.Valid) {
			err = clierr.Wrap(clierr.Auth, fmt.Errorf("saved session is invalid"))
		}
		if err != nil {
			failed("session", err)
			return report, first
		}
		passed("session", "Saved session is valid.")
	}
	var result struct {
		Inventories *[]struct {
			UUID string `json:"uuid"`
		} `json:"inventories"`
	}
	err := reader.Read(ctx, "v1/zones", url.Values{"limit": {"1"}}, &result)
	if err == nil && result.Inventories == nil {
		err = clierr.Wrap(clierr.Remote, fmt.Errorf("query response missing inventories"))
	}
	if err != nil {
		failed("query", err)
	} else {
		passed("query", "API query response is compatible; zone query permission is available.")
	}
	return report, first
}
