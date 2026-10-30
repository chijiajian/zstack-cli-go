package client

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"net/http"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

// LoginAccountContext avoids SDK authentication logs that expose request credentials.
func LoginAccountContext(ctx context.Context, endpoint, username, password string) (*view.SessionInventoryView, error) {
	cli, err := newHTTPClient(config.Context{Endpoint: endpoint})
	if err != nil {
		return nil, err
	}
	body := accountLoginRequest(username, password)
	var response struct {
		// Only decode identity fields needed to save the session. Server display
		// timestamps vary by locale and must not invalidate a successful login.
		Inventory struct {
			Uuid        string `json:"uuid"`
			AccountUuid string `json:"accountUuid"`
			UserUuid    string `json:"userUuid"`
		} `json:"inventory"`
	}
	if err := cli.request(ctx, http.MethodPut, "v1/accounts/login", nil, body, &response); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, clierr.Wrap(clierr.Describe(err).Code, fmt.Errorf("check endpoint connectivity and account credentials"))
	}
	if response.Inventory.Uuid == "" {
		return nil, fmt.Errorf("login response missing session UUID")
	}
	return &view.SessionInventoryView{Uuid: response.Inventory.Uuid, AccountUuid: response.Inventory.AccountUuid, UserUuid: response.Inventory.UserUuid}, nil
}

// accountLoginRequest omits unset SDK optional fields: an empty accountType is
// interpreted as a specified account type by some ZStack versions.
func accountLoginRequest(username, password string) any {
	var request struct {
		LoginByAccount struct {
			AccountName string `json:"accountName"`
			Password    string `json:"password"`
		} `json:"logInByAccount"`
	}
	request.LoginByAccount.AccountName = username
	request.LoginByAccount.Password = fmt.Sprintf("%x", sha512.Sum512([]byte(password)))
	return request
}

// Authentication errors cannot expose request credentials or saved session paths.
func safeAuthenticationError(ctx context.Context, stage string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	diagnostic := clierr.Describe(err)
	return clierr.Wrap(diagnostic.Code, fmt.Errorf("%s: %s", stage, diagnostic.Message))
}
