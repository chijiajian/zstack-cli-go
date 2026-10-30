package client

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExplicitLoginCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := LoginAccountContext(ctx, strings.TrimPrefix(server.URL, "http://"), "admin", "secret")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
}

func TestExplicitLoginRejectsMissingSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"inventory":{}}`)) }))
	defer server.Close()
	_, err := LoginAccountContext(context.Background(), strings.TrimPrefix(server.URL, "http://"), "admin", "secret")
	if err == nil || !strings.Contains(err.Error(), "missing session UUID") {
		t.Fatalf("expected missing UUID, got %v", err)
	}
}

func TestExplicitLoginAcceptsServerDisplayDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"inventory":{"uuid":"session","accountUuid":"account","userUuid":"user","expiredDate":"Oct 30, 2026 1:36:12 AM","createDate":"Oct 29, 2026 11:36:12 PM"}}`))
	}))
	defer server.Close()
	session, err := LoginAccountContext(context.Background(), strings.TrimPrefix(server.URL, "http://"), "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if session.Uuid != "session" || session.AccountUuid != "account" || session.UserUuid != "user" {
		t.Fatal("session identity lost")
	}
}

func TestAccountLoginRequestOmitsUnsetSDKFields(t *testing.T) {
	data, err := json.Marshal(accountLoginRequest("admin", "secret"))
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	params := body["logInByAccount"]
	if len(body) != 1 || len(params) != 2 || params["accountName"] != "admin" || params["password"] != fmt.Sprintf("%x", sha512.Sum512([]byte("secret"))) {
		t.Fatal("login body contains unexpected fields or credentials")
	}
}

func TestRenewalFailuresDoNotExposeSecrets(t *testing.T) {
	for _, failValidation := range []bool{false, true} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && !failValidation {
				w.Write([]byte(`{"valid":false}`))
				return
			}
			w.WriteHeader(500)
			w.Write([]byte(`{"error":{"details":"private-session private-password"}}`))
		}))
		cfg := &config.ZStackConfig{CurrentContext: "test", Contexts: map[string]config.Context{"test": {Endpoint: strings.TrimPrefix(server.URL, "http://"), Username: "admin", Password: "private-password", SessionUUID: "private-session"}}}
		config.SetContextOverride("")
		_, err := newAuthenticatedHTTPClient(context.Background(), cfg)
		server.Close()
		if err == nil || strings.Contains(err.Error(), "private-") {
			t.Fatalf("unsafe authentication error: %v", err)
		}
	}
}
