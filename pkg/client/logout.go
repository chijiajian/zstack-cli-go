package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

// LogoutContext revokes only the selected saved session and clears its credentials.
// Network I/O stays outside the config transaction. A concurrent replacement wins.
func LogoutContext(ctx context.Context, localOnly bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	selected, err := config.GetCurrentContext(cfg)
	if err != nil {
		return err
	}
	name := config.SelectedContextName(cfg)
	expected := *selected
	if !localOnly && expected.EffectiveAuthType() == config.AuthTypeAccount && expected.SessionUUID != "" {
		cli, err := newHTTPClient(expected)
		if err != nil {
			return err
		}
		err = cli.request(ctx, http.MethodDelete, "v1/accounts/sessions/"+url.PathEscape(expected.SessionUUID), nil, nil, nil)
		ResetClient()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("logout failed; credentials retained: %s", strings.ReplaceAll(err.Error(), expected.SessionUUID, "<session>"))
		}
	}
	err = config.UpdateConfig(ctx, func(latest *config.ZStackConfig) error {
		if current, ok := latest.Contexts[name]; !ok || current != expected {
			return fmt.Errorf("context %q changed during logout; newer credentials retained", name)
		}
		cleared := expected
		cleared.SessionUUID = ""
		cleared.Password = ""
		cleared.AccessKeyID = ""
		cleared.AccessKeySecret = ""
		latest.Contexts[name] = cleared
		return nil
	})
	if err == nil {
		ResetClient()
	}
	return err
}
