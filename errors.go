package cpsdk

import (
	"github.com/cenkalti/backoff/v5"

	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
)

// ErrUnsupportedOperation is returned when an operation is not supported for the current identity type.
var ErrUnsupportedOperation = base.ErrUnsupportedOperation

// UnexpectedStatusCodeError is returned when the control plane returns a non-200 status code. It includes the status code and first bytes of the response body for debugging purposes.
type UnexpectedStatusCodeError = base.UnexpectedStatusCodeError

// PermanentError is an alias for backoff.PermanentError to avoid exposing the entire backoff package in the public API.
type PermanenentError = backoff.PermanentError
