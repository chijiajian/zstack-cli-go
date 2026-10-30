package common

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

type failedOutput struct{ err error }

func (w failedOutput) Write([]byte) (int, error) { return 0, w.err }

func TestQueryTotalOutputFailure(t *testing.T) {
	cmd := &cobra.Command{}
	AddQueryFlags(cmd)
	cmd.Flags().Set("output", "text")
	cmd.Flags().Set("reply-with-count", "true")
	failure := errors.New("output unavailable")
	cmd.SetErr(failedOutput{failure})
	total := 0
	if err := PrintQuery(cmd, []struct{}{}, &total); !errors.Is(err, failure) {
		t.Fatalf("expected output failure, got %v", err)
	}
}
