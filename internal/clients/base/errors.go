package base

import (
	"fmt"
	"io"
	"net/http"

	"github.com/cenkalti/backoff/v5"

	"github.com/rudderlabs/rudder-go-kit/bytesize"
)

// ErrUnsupportedOperation is returned when an operation is not supported for the current identity type.
var ErrUnsupportedOperation = fmt.Errorf("operation not supported for this client")

// UnexpectedStatusCodeError is returned when the control plane returns a non-200 status code. It includes the status code and first bytes of the response body for debugging purposes.
type UnexpectedStatusCodeError struct {
	StatusCode int
	Body       []byte
}

func (e *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d, body: %s", e.StatusCode, string(e.Body))
}

// NewUnexpectedStatusCodeError creates a new UnexpectedStatusCodeError from the given HTTP response.
// It reads the response body and includes it in the error for debugging purposes.
func NewUnexpectedStatusCodeError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, bytesize.MB))
	return &UnexpectedStatusCodeError{
		StatusCode: res.StatusCode,
		Body:       body,
	}
}

// PermanentError is an alias for backoff.PermanentError to avoid exposing the entire backoff package in the public API.
type PermanenentError = backoff.PermanentError
