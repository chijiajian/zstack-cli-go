package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestIAMCommandsUseSDKWireContracts(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "password")
	policy := filepath.Join(dir, "statements.json")
	os.WriteFile(secret, []byte("plain-secret\n"), 0600)
	os.WriteFile(policy, []byte(`[{"name":"read","effect":"Allow","actions":["org.zstack.header.identity.APIQueryAccountMsg"]}]`), 0600)
	resources := []struct {
		kind, path, action string
		flags              []string
	}{
		{"account", "accounts", "updateAccount", []string{"--password-file", secret}},
		{"user", "accounts/users", "updateUser", []string{"--password-file", secret}},
		{"role", "identities/roles", "updateRole", nil},
		{"policy", "accounts/policies", "", []string{"--statements-file", policy}},
		{"project", "iam2/projects", "updateIAM2Project", nil},
		{"organization", "iam2/organizations", "updateIAM2Organization", []string{"--type", "Department"}},
	}
	for _, resource := range resources {
		for _, op := range []string{"create", "update", "delete"} {
			if op == "update" && resource.action == "" {
				continue
			}
			t.Run(op+"/"+resource.kind, func(t *testing.T) {
				var writes atomic.Int32
				cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
					base := "/zstack/v1/" + resource.path
					if r.Method == "GET" {
						if r.URL.Path != base {
							t.Errorf("unexpected resolver path %s", r.URL)
						}
						inventory(w, map[string]any{"uuid": "target", "name": "example"})
						return
					}
					writes.Add(1)
					if op == "delete" {
						if r.Method != "DELETE" || r.URL.Path != base+"/target" || r.URL.Query().Get("deleteMode") != "Permissive" {
							t.Errorf("delete contract: %s %s", r.Method, r.URL)
						}
						w.WriteHeader(204)
						return
					}
					expected := base
					method := "POST"
					key := "params"
					if op == "update" {
						method = "PUT"
						key = resource.action
						expected += "/target/actions"
						if resource.kind == "account" {
							expected = base + "/target"
						}
						if resource.kind == "user" {
							expected = base + "/actions"
						}
					}
					if r.Method != method || r.URL.Path != expected {
						t.Errorf("wire path: %s %s; want %s %s", r.Method, r.URL, method, expected)
					}
					var body map[string]map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					params, ok := body[key]
					if !ok {
						t.Errorf("missing SDK envelope %s", key)
					}
					if op == "create" && (resource.kind == "account" || resource.kind == "user") {
						password, _ := params["password"].(string)
						if len(password) != 128 || password == "plain-secret" {
							t.Error("password not hashed")
						}
					}
					if op == "update" && params["description"] != "" {
						t.Error("explicit empty description lost")
					}
					if op == "update" && resource.kind == "user" && params["uuid"] != "target" {
						t.Error("user action UUID missing")
					}
					w.Write([]byte(`{"inventory":{"uuid":"target","name":"example"}}`))
				})
				args := []string{op, resource.kind, "example", "-o", "json"}
				switch op {
				case "create":
					args = append(args, resource.flags...)
				case "update":
					args = append(args, "--description", "")
				case "delete":
					args = append(args, "--yes")
				}
				out, stderr, err := invokeCLI(t, cfg, "", args...)
				if err != nil || stderr != "" || !strings.Contains(out, "target") || writes.Load() != 1 {
					t.Fatalf("%v %s %s writes=%d", err, out, stderr, writes.Load())
				}
			})
		}
	}
}
func TestIAMPreviewAndValidationNeverContactAPI(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "password")
	os.WriteFile(secret, []byte("plain-secret"), 0600)
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) { t.Error("offline preview contacted server") })
	out, stderr, err := invokeCLI(t, cfg, "", "create", "account", "test", "--password-file", secret, "--dry-run", "-o", "json")
	if err != nil || !strings.Contains(out, "redacted") || strings.Contains(out, "plain-secret") {
		t.Fatalf("preview: %v %s %s", err, out, stderr)
	}
	for _, args := range [][]string{{"create", "account", "test"}, {"create", "organization", "test"}, {"update", "user", "test"}, {"create", "project", " "}, {"delete", "project", "test", "--delete-mode", "invalid"}, {"create", "policy", "test"}, {"create", "project", "test", "--resource-uuid", "invalid"}} {
		_, stderr, err := invokeCLI(t, cfg, "", append(args, "--error-format", "json")...)
		assertDiagnostic(t, stderr, err, "invalid_arguments", 2)
	}
}
func TestIAMDeleteRequiresConfirmation(t *testing.T) {
	var writes atomic.Int32
	cfg := mockConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes.Add(1)
		}
		inventory(w, map[string]any{"uuid": "target", "name": "example"})
	})
	_, _, err := invokeCLI(t, cfg, "no\n", "delete", "account", "example")
	if err != nil || writes.Load() != 0 {
		t.Fatal("declined identity deletion sent write")
	}
}
