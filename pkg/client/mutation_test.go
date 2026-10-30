package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/config"
)

func mutationTestClient(t *testing.T, handler http.HandlerFunc) *MutationClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	reader, err := newHTTPClient(config.Context{Endpoint: strings.TrimPrefix(server.URL, "http://"), SessionUUID: "test-session"})
	if err != nil {
		t.Fatal(err)
	}
	return &MutationClient{reader: reader, pollInterval: time.Millisecond}
}
func TestMutationPollsWithoutRepeatingWrite(t *testing.T) {
	var writes, polls atomic.Int32
	cli := mutationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "OAuth test-session" {
			t.Error("missing auth")
		}
		if r.Method == "POST" {
			writes.Add(1)
			w.WriteHeader(202)
			fmt.Fprint(w, `{"location":"/zstack/v1/api-jobs/job-1"}`)
			return
		}
		if r.Method != "GET" || r.URL.Path != "/zstack/v1/api-jobs/job-1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		if polls.Add(1) == 1 {
			w.WriteHeader(202)
			return
		}
		fmt.Fprint(w, `{"inventory":{"uuid":"created"}}`)
	})
	var result struct {
		UUID string `json:"uuid"`
	}
	err := cli.Mutate(context.Background(), "POST", "v1/accounts", map[string]any{}, &result)
	if err != nil || result.UUID != "created" || writes.Load() != 1 || polls.Load() != 2 {
		t.Fatalf("%v %+v %d %d", err, result, writes.Load(), polls.Load())
	}
}
func TestMutationRejectsUnsafeCallbacks(t *testing.T) {
	for _, location := range []string{"http://elsewhere/zstack/v1/api-jobs/id", "/zstack/v1/accounts", "/zstack/v1/api-jobs/id?token=secret", "/zstack/v1/api-jobs/%2Faccounts", "/zstack/v1/api-jobs/"} {
		if _, err := apiJobPath("http://server:8080/zstack/", location); err == nil {
			t.Errorf("accepted %s", location)
		}
	}
}
func TestMutationCancellationAndRedaction(t *testing.T) {
	var writes atomic.Int32
	cli := mutationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			writes.Add(1)
			w.WriteHeader(202)
			fmt.Fprint(w, `{"location":"/zstack/v1/api-jobs/job"}`)
			return
		}
		w.WriteHeader(202)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if err := cli.Mutate(ctx, "POST", "v1/accounts", nil, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("%v", err)
	}
	if writes.Load() != 1 {
		t.Fatal("write retried")
	}
	for _, status := range []int{200, 500} {
		cli := mutationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			fmt.Fprint(w, `{"success":false,"error":{"code":"FAIL","details":"secret"}}`)
		})
		err := cli.Mutate(context.Background(), "POST", "v1/accounts", nil, nil)
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("failure ignored or secrets exposed")
		}
	}
}

func TestMutationDeadlineDuringResponseBodyKeepsTimeout(t *testing.T) {
	cli := mutationTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	err := cli.Mutate(ctx, "POST", "v1/accounts", map[string]any{}, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lost deadline: %v", err)
	}
}
