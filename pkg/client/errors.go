package client

import (
	"errors"

	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/util/httputils"
)

// NormalizeAPIError keeps SDK-specific error knowledge in the client adapter.
// Domain/configuration code can use the error contract without importing the SDK.
func NormalizeAPIError(err error) error {
	var typed *clierr.Error
	var httpErr *clierr.HTTPError
	if errors.As(err, &typed) || errors.As(err, &httpErr) {
		return err
	}
	var sdkErr *httputils.JSONClientError
	if errors.As(err, &sdkErr) {
		return &clierr.HTTPError{Status: sdkErr.Code, Message: err.Error()}
	}
	return err
}
