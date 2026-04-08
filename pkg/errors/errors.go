// Package errors provides domain-specific error types for NixGuard.
// Follows Google API error model with error codes and details.
package errors

import (
	"fmt"
	"net/http"
)

// Code represents an error category.
type Code int

const (
	CodeOK                 Code = 0
	CodeInvalidArgument    Code = 1
	CodeNotFound           Code = 2
	CodeAlreadyExists      Code = 3
	CodePermissionDenied   Code = 4
	CodeUnauthenticated    Code = 5
	CodeInternal           Code = 6
	CodeUnavailable        Code = 7
	CodeConflict           Code = 8
	CodeResourceExhausted  Code = 9
	CodeFailedPrecondition Code = 10
)

// Error is the standard NixGuard error type.
type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus maps error codes to HTTP status codes.
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodePermissionDenied:
		return http.StatusForbidden
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeConflict:
		return http.StatusConflict
	case CodeResourceExhausted:
		return http.StatusTooManyRequests
	case CodeFailedPrecondition:
		return http.StatusPreconditionFailed
	default:
		return http.StatusInternalServerError
	}
}

// Constructor helpers
func InvalidArgument(msg string) *Error   { return &Error{Code: CodeInvalidArgument, Message: msg} }
func NotFound(msg string) *Error          { return &Error{Code: CodeNotFound, Message: msg} }
func AlreadyExists(msg string) *Error     { return &Error{Code: CodeAlreadyExists, Message: msg} }
func PermissionDenied(msg string) *Error  { return &Error{Code: CodePermissionDenied, Message: msg} }
func Unauthenticated(msg string) *Error   { return &Error{Code: CodeUnauthenticated, Message: msg} }
func Internal(msg string) *Error          { return &Error{Code: CodeInternal, Message: msg} }
func Unavailable(msg string) *Error       { return &Error{Code: CodeUnavailable, Message: msg} }

func Wrap(err error, code Code, msg string) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}
