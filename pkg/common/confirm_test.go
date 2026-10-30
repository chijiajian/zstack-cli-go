package common

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type failedConfirmationReader struct{}

func (failedConfirmationReader) Read([]byte) (int, error) { return 0, errors.New("input unavailable") }

func TestConfirmationRequiresExplicitAnswerAndWorkingIO(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  bool
	}{{"yes\n", true}, {"Y", true}, {"no\n", false}, {"", false}} {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("yes", false, "")
		cmd.SetIn(strings.NewReader(tc.input))
		cmd.SetOut(io.Discard)
		got, err := Confirm(cmd, "confirm: ")
		if err != nil || got != tc.want {
			t.Fatalf("input=%q got=%v err=%v", tc.input, got, err)
		}
	}
	cmd := &cobra.Command{}
	cmd.Flags().Bool("yes", false, "")
	cmd.SetIn(failedConfirmationReader{})
	cmd.SetOut(io.Discard)
	if ok, err := Confirm(cmd, "confirm: "); ok || err == nil {
		t.Fatalf("read error accepted: %v %v", ok, err)
	}
	cmd.SetIn(strings.NewReader("yes\n"))
	cmd.SetOut(failedOutput{errors.New("closed")})
	if ok, err := Confirm(cmd, "confirm: "); ok || err == nil {
		t.Fatalf("write error accepted: %v %v", ok, err)
	}
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.Flags().Set("yes", "true")
	if ok, err := Confirm(cmd, "confirm: "); !ok || err != nil || output.Len() != 0 {
		t.Fatalf("explicit yes: %v %v %q", ok, err, output.String())
	}
}
