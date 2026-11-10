package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
)

// MutationClient adds bounded writes to the same authentication/transport layer.
// It never retries the initial mutation. Only the returned API job is polled.
type MutationClient struct {
	reader       *ReadClient
	pollInterval time.Duration
}

func NewMutationClient(ctx context.Context) (*MutationClient, error) {
	reader, err := NewReadClient(ctx)
	if err != nil {
		return nil, err
	}
	return &MutationClient{reader: reader, pollInterval: time.Second}, nil
}
func (cli *MutationClient) Reader() *ReadClient { return cli.reader }

func (cli *MutationClient) Mutate(ctx context.Context, method, path string, body any, result any) error {
	return cli.mutate(ctx, method, path, nil, body, result)
}

func (cli *MutationClient) Delete(ctx context.Context, path, mode string) error {
	if mode != "Permissive" && mode != "Enforcing" {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("invalid deletion mode"))
	}
	return cli.mutate(ctx, http.MethodDelete, path, url.Values{"deleteMode": {mode}}, nil, nil)
}

func (cli *MutationClient) mutate(ctx context.Context, method, path string, query url.Values, body, result any) error {
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("unsupported mutation method"))
	}
	if !strings.HasPrefix(path, "v1/") || strings.ContainsAny(path, "?#") {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("invalid API path"))
	}
	var peer atomic.Value
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{GotConn: func(info httptrace.GotConnInfo) {
		peer.Store(info.Conn.RemoteAddr().String())
	}})
	pollPath := ""
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := cli.reader.send(ctx, method, path, query, body)
		if err != nil {
			return err
		}
		status := resp.StatusCode
		var envelope struct {
			Success   *bool           `json:"success"`
			Location  string          `json:"location"`
			Inventory json.RawMessage `json:"inventory"`
			Error     json.RawMessage `json:"error"`
		}
		if status != http.StatusNoContent {
			err = decodeJSONDocument(resp.Body, &envelope)
		}
		resp.Body.Close()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, io.EOF) && (status == http.StatusAccepted && pollPath != "" || status == http.StatusOK && result == nil) {
			err = nil
		}
		if err != nil {
			if status >= 400 {
				return &clierr.HTTPError{Status: status, Message: fmt.Sprintf("API mutation failed (HTTP %d)", status)}
			}
			return clierr.Wrap(clierr.Remote, fmt.Errorf("invalid mutation response (HTTP %d)", status))
		}
		if status < 200 || status >= 300 || envelope.Success != nil && !*envelope.Success || len(envelope.Error) > 0 && string(envelope.Error) != "null" {
			var api struct {
				Code string `json:"code"`
			}
			_ = json.Unmarshal(envelope.Error, &api)
			errorStatus := status
			if errorStatus < 400 {
				errorStatus = 500
			}
			// No response/request contents are exposed: identity operations may echo passwords.
			return &clierr.HTTPError{Status: errorStatus, Code: api.Code, Message: fmt.Sprintf("API mutation failed (HTTP %d, code %s); inspect resource state before retrying", status, clierr.SafeAPICode(api.Code))}
		}
		if envelope.Location != "" {
			connectedPeer, _ := peer.Load().(string)
			pollPath, err = apiJobPath(cli.reader.endpoint, envelope.Location, connectedPeer)
			if err != nil {
				return err
			}
		}
		if status == http.StatusAccepted || envelope.Location != "" {
			if pollPath == "" {
				return clierr.Wrap(clierr.Remote, fmt.Errorf("accepted mutation omitted job location"))
			}
			timer := time.NewTimer(cli.pollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			method, path, body = http.MethodGet, pollPath, nil
			query = nil
			continue
		}
		if result != nil {
			if len(envelope.Inventory) == 0 || string(envelope.Inventory) == "null" {
				return clierr.Wrap(clierr.Remote, fmt.Errorf("mutation response missing inventory"))
			}
			if err := decodeResponse(envelope.Inventory, result); err != nil {
				return clierr.Wrap(clierr.Remote, fmt.Errorf("invalid mutation inventory"))
			}
		}
		return ctx.Err()
	}
}

// A management node may advertise its peer IP when the endpoint was configured
// with a hostname. Return only the path so polling uses the configured endpoint.
func apiJobPath(endpoint, location, connectedPeer string) (string, error) {
	base, _ := url.Parse(endpoint)
	ref, err := url.Parse(location)
	if err != nil {
		return "", clierr.Wrap(clierr.Remote, fmt.Errorf("invalid API job location"))
	}
	target := base.ResolveReference(ref)
	prefix := base.Path + "v1/api-jobs/"
	if target.Scheme != base.Scheme || target.User != nil || target.RawQuery != "" || target.Fragment != "" || !strings.HasPrefix(target.Path, prefix) {
		return "", clierr.Wrap(clierr.Remote, fmt.Errorf("unsafe API job location"))
	}
	if target.Host != base.Host {
		peerIP, peerPort, _ := net.SplitHostPort(connectedPeer)
		callbackIP := net.ParseIP(target.Hostname())
		if callbackIP == nil || !callbackIP.Equal(net.ParseIP(peerIP)) || target.Port() != peerPort || target.Port() != base.Port() {
			return "", clierr.Wrap(clierr.Remote, fmt.Errorf("unsafe API job location"))
		}
	}
	job := strings.TrimPrefix(target.Path, prefix)
	if job == "" || strings.ContainsAny(job, "/\\") || job == "." || job == ".." {
		return "", clierr.Wrap(clierr.Remote, fmt.Errorf("invalid API job identifier"))
	}
	return "v1/api-jobs/" + url.PathEscape(job), nil
}
