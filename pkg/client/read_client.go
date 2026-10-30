package client

// The SDK v0.0.10 HTTP transport drops request contexts. Keep bounded
// authentication and reads on net/http until the SDK supports cancellation.
import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

// ReadClient provides cancellable authentication and resource reads.
type ReadClient struct {
	endpoint    string
	credentials config.Context
	httpClient  *http.Client
}

// newHTTPClient builds transport without validating or renewing a session.
func newHTTPClient(credentials config.Context) (*ReadClient, error) {
	if _, err := NewSDKConfig(credentials.Endpoint); err != nil {
		return nil, err
	}
	endpoint := credentials.Endpoint
	if !strings.Contains(endpoint, ":") {
		endpoint += ":8080"
	}
	return &ReadClient{endpoint: "http://" + endpoint + "/zstack/", credentials: credentials,
		httpClient: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func newAuthenticatedHTTPClient(ctx context.Context, cfg *config.ZStackConfig) (*ReadClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	selectedName := config.SelectedContextName(cfg)
	credentials, err := config.GetCurrentContext(cfg)
	if err != nil {
		return nil, err
	}
	cli, err := newHTTPClient(*credentials)
	if err != nil {
		return nil, err
	}
	switch credentials.EffectiveAuthType() {
	case config.AuthTypeAccessKey:
		if credentials.AccessKeyID == "" || credentials.AccessKeySecret == "" {
			return nil, clierr.Wrap(clierr.Auth, fmt.Errorf("access key credentials missing; please login"))
		}
	case config.AuthTypeAccount:
		if credentials.SessionUUID != "" {
			var validation struct {
				Valid *bool `json:"valid"`
			}
			if err := cli.request(ctx, http.MethodGet, "v1/accounts/sessions/"+url.PathEscape(credentials.SessionUUID)+"/valid", nil, nil, &validation); err != nil {
				return nil, safeAuthenticationError(ctx, "validate saved session", err)
			}
			if validation.Valid == nil {
				return nil, fmt.Errorf("validate saved session: response missing valid")
			}
			if *validation.Valid {
				return cli, nil
			}
		}
		if credentials.Username == "" || credentials.Password == "" {
			return nil, clierr.Wrap(clierr.Auth, fmt.Errorf("session missing or expired and no saved password; please run 'zstack-cli login' again"))
		}
		body := accountLoginRequest(credentials.Username, credentials.Password)
		var result struct {
			Inventory struct {
				UUID string `json:"uuid"`
			} `json:"inventory"`
		}
		if err := cli.request(ctx, http.MethodPut, "v1/accounts/login", nil, body, &result); err != nil {
			return nil, safeAuthenticationError(ctx, "renew session", err)
		}
		if result.Inventory.UUID == "" {
			return nil, fmt.Errorf("login response missing session UUID")
		}
		cli.credentials.SessionUUID = result.Inventory.UUID
		if err := config.UpdateConfig(ctx, func(latest *config.ZStackConfig) error {
			saved, ok := latest.Contexts[selectedName]
			expected := *credentials
			comparable := saved
			comparable.SessionUUID = expected.SessionUUID
			if !ok || comparable != expected {
				return fmt.Errorf("context %q changed during login; retry with the updated configuration", selectedName)
			}
			// Another refresh may have committed while this login was in flight.
			if saved.SessionUUID != expected.SessionUUID && saved.SessionUUID != "" {
				cli.credentials = saved
				return nil
			}
			latest.Contexts[selectedName] = cli.credentials
			return nil
		}); err != nil {
			return nil, fmt.Errorf("save session: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported auth type %q", credentials.AuthType)
	}
	return cli, nil
}

func NewReadClient(ctx context.Context) (*ReadClient, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return newAuthenticatedHTTPClient(ctx, cfg)
}

func (cli *ReadClient) send(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	endpoint := cli.endpoint + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cli.credentials.EffectiveAuthType() == config.AuthTypeAccessKey {
		date := time.Now().UTC().Format(http.TimeFormat)
		mac := hmac.New(sha1.New, []byte(cli.credentials.AccessKeySecret))
		fmt.Fprintf(mac, "%s\n%s\n/%s", method, date, path)
		req.Header.Set("Date", date)
		req.Header.Set("Authorization", "ZStack "+cli.credentials.AccessKeyID+":"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	} else if path != "v1/accounts/login" {
		req.Header.Set("Authorization", "OAuth "+cli.credentials.SessionUUID)
		req.Header.Set("x-session-id", cli.credentials.SessionUUID)
	}
	return cli.httpClient.Do(req)
}

func (cli *ReadClient) request(ctx context.Context, method, path string, query url.Values, body, result any) error {
	resp, err := cli.send(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Session revocation may return an empty 200/204; 404 is already logged out.
	if method == http.MethodDelete && result == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound) {
		return ctx.Err()
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.Unmarshal(data, &body)
		return &clierr.HTTPError{Status: resp.StatusCode, Code: body.Error.Code, Message: fmt.Sprintf("%s %s: HTTP %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(data)))}
	}
	var raw json.RawMessage
	if err := decodeJSONDocument(resp.Body, &raw); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	if err := decodeResponse(raw, result); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return ctx.Err()
}

func (cli *ReadClient) query(ctx context.Context, path string, params *param.QueryParam, result any) error {
	var response struct {
		Inventories json.RawMessage `json:"inventories"`
	}
	if err := cli.request(ctx, http.MethodGet, path, params.Values, nil, &response); err != nil {
		return err
	}
	if len(response.Inventories) == 0 {
		return fmt.Errorf("query %s response missing inventories", path)
	}
	return decodeResponse(response.Inventories, result)
}

func (cli *ReadClient) Resolve(ctx context.Context, resourceType, identifier string, strategy ResolveStrategy) (ResolvedResource, error) {
	matches, err := resolveResource(ctx, cli.query, resourceType, identifier, ResolveOptions{Strategy: strategy})
	if err != nil {
		return ResolvedResource{}, err
	}
	return matches[0], nil
}

func (cli *ReadClient) Read(ctx context.Context, path string, values url.Values, result any) error {
	return cli.request(ctx, http.MethodGet, path, values, nil, result)
}

// NewProbeClient uses only supplied credentials, without validating or renewing
// them. The exported surface supports reads only; intended for diagnostics.
func NewProbeClient(credentials config.Context) (*ReadClient, error) {
	return newHTTPClient(credentials)
}
