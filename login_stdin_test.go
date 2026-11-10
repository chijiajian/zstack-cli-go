package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

func TestLoginPasswordStdin(t *testing.T) {
	const password = " \t合成-password\n "
	digest := sha512.Sum512([]byte(password))
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var request map[string]map[string]string
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || r.Method != http.MethodPut || r.URL.Path != "/zstack/v1/accounts/login" {
			t.Error("unexpected login request")
		}
		account := request["logInByAccount"]
		if account["accountName"] != "test" || account["password"] != hex.EncodeToString(digest[:]) {
			t.Error("password or username did not arrive intact")
		}
		fmt.Fprint(w, `{"inventory":{"uuid":"test-session-id"}}`)
	}))
	defer server.Close()
	cfg := writeContexts(t, config.ZStackConfig{CurrentContext: "existing", Contexts: map[string]config.Context{"existing": {Endpoint: "existing"}}})
	args := []string{"login", strings.TrimPrefix(server.URL, "http://"), "--password-stdin", "--username", "test", "--context-name", "target", "--no-switch", "--error-format", "json"}
	out, stderr, err := invokeCLI(t, cfg, password, args...)
	if err != nil || stderr != "" || strings.Contains(out, "test-session-id") || strings.Contains(out, password) {
		t.Fatalf("stdin login failed or exposed credentials: %v", err)
	}
	saved := readContexts(t, cfg)
	if saved.CurrentContext != "existing" || saved.Contexts["target"].Password != "" || saved.Contexts["target"].SessionUUID != "test-session-id" {
		t.Fatal("stdin login did not save session only or changed default")
	}
	for _, tc := range []struct {
		name, input string
		flags       []string
	}{
		{"empty", "", nil}, {"oversize", strings.Repeat("a", 65537), nil}, {"invalid-utf8", "\xff", nil},
		{"password-flag", password, []string{"--password="}}, {"key-flag", password, []string{"--ak="}},
		{"save", password, []string{"--save-password=true"}}, {"missing-user", password, []string{"--username="}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, stderr, err := invokeCLI(t, cfg, tc.input, append(append([]string{}, args...), tc.flags...)...)
			assertDiagnostic(t, stderr, err, "invalid_arguments", 2)
			if out != "" || strings.Contains(stderr, password) || calls.Load() != 1 {
				t.Fatal("invalid input prompted, leaked, or made a request")
			}
		})
	}
}

func TestLoginAccessKeySecretStdin(t *testing.T) {
	const secret = " synthetic-key-secret\n "
	cfg := writeContexts(t, config.ZStackConfig{CurrentContext: "existing", Contexts: map[string]config.Context{"existing": {Endpoint: "existing"}}})
	args := []string{"login", "127.0.0.1:1", "--access-key-id", "test-key", "--access-key-secret-stdin", "--context-name", "target", "--no-switch", "--error-format", "json"}
	out, stderr, err := invokeCLI(t, cfg, secret, args...)
	if err != nil || stderr != "" || strings.Contains(out, strings.TrimSpace(secret)) {
		t.Fatalf("access key stdin failed or leaked the secret: %v", err)
	}
	saved := readContexts(t, cfg)
	key := saved.Contexts["target"]
	if saved.CurrentContext != "existing" || key.EffectiveAuthType() != config.AuthTypeAccessKey ||
		key.AccessKeyID != "test-key" || key.AccessKeySecret != secret || key.SessionUUID != "" || key.Password != "" {
		t.Fatal("access key stdin stored unexpected credentials or changed the default context")
	}
	for _, tc := range []struct {
		name, input string
		flags       []string
	}{
		{"empty", "", nil}, {"oversize", strings.Repeat("a", 65537), nil}, {"invalid-utf8", "\xff", nil},
		{"missing-key-id", secret, []string{"--access-key-id="}}, {"secret-flag", secret, []string{"--sk="}},
		{"password-mode", secret, []string{"--password-stdin"}}, {"username", secret, []string{"--username=test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, stderr, err := invokeCLI(t, cfg, tc.input, append(append([]string{}, args...), tc.flags...)...)
			assertDiagnostic(t, stderr, err, "invalid_arguments", 2)
			if out != "" || strings.Contains(stderr, strings.TrimSpace(secret)) {
				t.Fatal("invalid key input prompted or leaked")
			}
		})
	}
}
