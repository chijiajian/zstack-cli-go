// Package clierr defines the process-level error contract independently of Cobra.
package clierr

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"regexp"
	"syscall"
)

type Kind string

const (
	General   Kind = "operation_failed"
	Usage     Kind = "invalid_arguments"
	Auth      Kind = "authentication_failed"
	NotFound  Kind = "not_found"
	Conflict  Kind = "conflict"
	Timeout   Kind = "timeout"
	Transport Kind = "connection_failed"
	Remote    Kind = "api_failed"
	Output    Kind = "output_failed"
	Canceled  Kind = "canceled"
)

type Error struct {
	Kind  Kind
	Cause error
}

func (e *Error) Error() string { return e.Cause.Error() }
func (e *Error) Unwrap() error { return e.Cause }
func Wrap(kind Kind, err error) error {
	if err == nil {
		return nil
	}
	return &Error{kind, err}
}

// HTTPError preserves human diagnostics while exposing HTTP status structurally.
type HTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

type Diagnostic struct {
	Code       Kind   `json:"code"`
	Message    string `json:"message"`
	ExitCode   int    `json:"exitCode"`
	HTTPStatus int    `json:"httpStatus,omitempty"`
	APICode    string `json:"apiCode,omitempty"`
}

func Describe(err error) Diagnostic {
	kind := General
	var typed *Error
	var httpErr *HTTPError
	var network net.Error
	var unsupportedType *json.UnsupportedTypeError
	var unsupportedValue *json.UnsupportedValueError
	switch {
	case errors.Is(err, context.Canceled):
		kind = Canceled
	case errors.Is(err, context.DeadlineExceeded):
		kind = Timeout
	case errors.As(err, &typed):
		kind = typed.Kind
	case errors.As(err, &httpErr):
		kind = fromHTTP(httpErr.Status)
		if httpErr.Code == "ID.1000" {
			kind = Auth
		}
	case errors.As(err, &network):
		kind = Transport
		if network.Timeout() {
			kind = Timeout
		}
	case errors.Is(err, os.ErrClosed), errors.As(err, &unsupportedType), errors.As(err, &unsupportedValue), errors.Is(err, io.ErrShortWrite), errors.Is(err, io.ErrClosedPipe), errors.Is(err, syscall.EPIPE):
		kind = Output
	}
	// Structured diagnostics intentionally omit raw SDK errors: they may contain
	// request bodies, signed headers, credentials, or server echoes of secrets.
	messages := map[Kind]Diagnostic{
		General:   {Code: General, Message: "Operation failed; inspect resource state before retrying a write.", ExitCode: 1},
		Usage:     {Code: Usage, Message: "Invalid command arguments; see command --help.", ExitCode: 2},
		Auth:      {Code: Auth, Message: "Authentication or permission check failed; verify the selected context.", ExitCode: 3},
		NotFound:  {Code: NotFound, Message: "The requested resource was not found.", ExitCode: 4},
		Conflict:  {Code: Conflict, Message: "The operation conflicts with current state; refresh and retry.", ExitCode: 5},
		Timeout:   {Code: Timeout, Message: "The operation timed out; a write may still have completed.", ExitCode: 6},
		Transport: {Code: Transport, Message: "Unable to communicate with the configured endpoint.", ExitCode: 7},
		Remote:    {Code: Remote, Message: "The API rejected the request or returned an invalid response.", ExitCode: 8},
		Output:    {Code: Output, Message: "Unable to write command output; a write may still have completed.", ExitCode: 9},
		Canceled:  {Code: Canceled, Message: "The operation was canceled; a write may still have completed.", ExitCode: 130},
	}
	result, ok := messages[kind]
	if !ok {
		return messages[General]
	}
	if errors.As(err, &httpErr) {
		result.HTTPStatus = httpErr.Status
		result.APICode = SafeAPICode(httpErr.Code)
	}
	return result
}

var apiCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,20}\.[0-9]{1,8}$`)

func SafeAPICode(code string) string {
	if apiCodePattern.MatchString(code) {
		return code
	}
	return ""
}

func fromHTTP(status int) Kind {
	switch status {
	case 400, 422:
		return Usage
	case 401, 403:
		return Auth
	case 404:
		return NotFound
	case 409:
		return Conflict
	case 408, 504:
		return Timeout
	default:
		return Remote
	}
}
