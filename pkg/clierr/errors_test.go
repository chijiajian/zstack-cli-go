package clierr

import (
	"context"
	"fmt"
	"io"
	"testing"
)

func TestWrappedErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code Kind
		exit int
	}{{context.Canceled, Canceled, 130}, {context.DeadlineExceeded, Timeout, 6}, {io.ErrShortWrite, Output, 9}, {Wrap(Conflict, fmt.Errorf("state")), Conflict, 5}, {&HTTPError{Status: 503, Code: "ID.1000", Message: "secret"}, Auth, 3}} {
		d := Describe(fmt.Errorf("operation: %w", tc.err))
		if d.Code != tc.code || d.ExitCode != tc.exit {
			t.Fatalf("%+v", d)
		}
	}
}

func TestAPICodeMetadataIsBoundedAndSanitized(t *testing.T) {
	diagnostic := Describe(&HTTPError{Status: 503, Code: "LICENSE.1001", Message: "raw secret"})
	if diagnostic.APICode != "LICENSE.1001" || diagnostic.HTTPStatus != 503 {
		t.Fatalf("%+v", diagnostic)
	}
	diagnostic = Describe(&HTTPError{Status: 500, Code: "password=secret", Message: "raw secret"})
	if diagnostic.APICode != "" {
		t.Fatal("untrusted API code leaked")
	}
}
