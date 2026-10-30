package main

import (
	"crypto/sha512"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

func TestWriteOutputValidationPrecedesRequests(t *testing.T) {
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("invalid output contacted API: %s %s", r.Method, r.URL)
	})
	for _, args := range [][]string{
		{"create", "volume", "vol", "--size", "1G"},
		{"update", "volume", "vol", "--description", "updated"},
		{"primary-storage", "add-local", "local", "--zone", "zone", "--url", "/data/local"},
		{"delete", "volume", "vol", "--yes"},
		{"expunge", "image", "image", "--yes"},
	} {
		out, stderr, err := invokeCLI(t, cfg, "", append(args, "-o", "invalid")...)
		if err == nil || out != "" || !strings.Contains(stderr, "unsupported output format") {
			t.Fatalf("%v err=%v out=%s stderr=%s", args, err, out, stderr)
		}
	}
}

func TestLogoutCLIUsesSelectedContextAndNoLogin(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodDelete || r.URL.Path != "/zstack/v1/accounts/sessions/selected-token" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cfg := writeContexts(t, config.ZStackConfig{CurrentContext: "default", Contexts: map[string]config.Context{
		"default":  {Endpoint: "unchanged", SessionUUID: "keep"},
		"selected": {Endpoint: strings.TrimPrefix(server.URL, "http://"), Username: "admin", Password: "saved", SessionUUID: "selected-token"},
	}})
	out, stderr, err := invokeCLI(t, cfg, "", "logout", "--context", "selected", "--request-timeout", "2s")
	if err != nil || calls != 1 || !strings.Contains(out, "Logged out") || stderr != "" {
		t.Fatalf("err=%v calls=%d out=%s stderr=%s", err, calls, out, stderr)
	}
	saved := readContexts(t, cfg)
	if saved.CurrentContext != "default" || saved.Contexts["default"].SessionUUID != "keep" || saved.Contexts["selected"].SessionUUID != "" || saved.Contexts["selected"].Password != "" {
		t.Fatal("wrong logout configuration")
	}
}

func TestLoginFailureDoesNotExposeAuthenticationBody(t *testing.T) {
	const password = "private-login-password"
	digest := fmt.Sprintf("%x", sha512.Sum512([]byte(password)))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error":{"code":"AUTH.1000","details":%q}}`, "echo: "+password+" "+digest)
	}))
	defer server.Close()
	out, stderr, err := invokeCLI(t, filepath.Join(t.TempDir(), "config.yaml"), "", "login", strings.TrimPrefix(server.URL, "http://"), "--username", "admin", "--password", password, "--save-password=false")
	if err == nil || !strings.Contains(stderr, "login failed") {
		t.Fatalf("expected actionable login error: %v %s", err, stderr)
	}
	for _, secret := range []string{password, digest, "logInByAccount", "AUTH.1000"} {
		if strings.Contains(out+stderr, secret) {
			t.Fatal("authentication details leaked in output")
		}
	}
}
