package client

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	sdk "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
)

// NewSDKConfig accepts a hostname or host:port. Bare hosts use ZStack's default port.
func NewSDKConfig(endpoint string) (*sdk.ZSConfig, error) {
	if strings.TrimSpace(endpoint) == "" || strings.ContainsAny(endpoint, "/?#@ ") {
		return nil, fmt.Errorf("invalid endpoint %q: expected host or host:port", endpoint)
	}
	if strings.Contains(endpoint, ":") {
		host, portString, err := net.SplitHostPort(endpoint)
		if err != nil {
			return nil, fmt.Errorf("invalid endpoint %q: expected host or host:port", endpoint)
		}
		port, err := strconv.Atoi(portString)
		if err != nil || port < 1 || port > 65535 || host == "" {
			return nil, fmt.Errorf("invalid endpoint %q", endpoint)
		}
		if strings.Contains(host, ":") {
			host = "[" + host + "]"
		}
		return sdk.NewZSConfig(host, port, "zstack"), nil
	}
	return sdk.DefaultZSConfig(endpoint), nil
}
