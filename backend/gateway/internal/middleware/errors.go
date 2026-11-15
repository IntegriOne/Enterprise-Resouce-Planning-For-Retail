// Package middleware mengimplementasikan verifikasi sesi Redis fail-closed, penyegaran snapshot singleflight, otorisasi RBAC/reauth, dan peta error Gateway.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode merepresentasikan kode ekstensi GraphQL terstandar yang aman dikonsumsi frontend IntegriOne.
type ErrorCode string

const (
	CodeUnauthenticated     ErrorCode = "UNAUTHENTICATED"
	CodeForbidden           ErrorCode = "FORBIDDEN"
	CodeUnavailable         ErrorCode = "UNAVAILABLE"
	CodeTooManyRequests     ErrorCode = "TOO_MANY_REQUESTS"
	CodeInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	CodeInvalidArgument     ErrorCode = "INVALID_ARGUMENT"
	CodeConflict            ErrorCode = "CONFLICT"
	CodeUnprocessableEntity ErrorCode = "UNPROCESSABLE_ENTITY"
	CodeNotFound            ErrorCode = "NOT_FOUND"
	CodeInternalError       ErrorCode = "INTERNAL_ERROR"
)

// GatewayError membungkus error aman untuk dikirim ke lapisan GraphQL ErrorPresenter tanpa membocorkan detail internal.
type GatewayError struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	Err        error
}

// Error mengembalikan pesan aman untuk pemanggil.
func (e *GatewayError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Unwrap mengembalikan penyebab internal asli untuk logging terstruktur server.
func (e *GatewayError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewGatewayError membentuk error Gateway terklasifikasi beserta kode status HTTP yang sesuai.
func NewGatewayError(code ErrorCode, message string, cause error) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatusForCode(code),
		Err:        cause,
	}
}

// AsGatewayError mengekstrak *GatewayError dari rantai error bila tersedia.
func AsGatewayError(err error) (*GatewayError, bool) {
	if err == nil {
		return nil, false
	}
	var gwErr *GatewayError
	if errors.As(err, &gwErr) {
		return gwErr, true
	}
	return nil, false
}

// ErrUnauthenticated membentuk error UNAUTHENTICATED terstandar.
func ErrUnauthenticated(msg string, cause error) error {
	if msg == "" {
		msg = "authentication required"
	}
	return NewGatewayError(CodeUnauthenticated, msg, cause)
}

// ErrForbidden membentuk error FORBIDDEN terstandar.
func ErrForbidden(msg string, cause error) error {
	if msg == "" {
		msg = "forbidden"
	}
	return NewGatewayError(CodeForbidden, msg, cause)
}

// ErrUnavailable membentuk error UNAVAILABLE fail-closed terstandar.
func ErrUnavailable(msg string, cause error) error {
	if msg == "" {
		msg = "service temporarily unavailable"
	}
	return NewGatewayError(CodeUnavailable, msg, cause)
}

// ErrTooManyRequests membentuk error TOO_MANY_REQUESTS terstandar.
func ErrTooManyRequests(msg string, cause error) error {
	if msg == "" {
		msg = "too many requests, please try again later"
	}
	return NewGatewayError(CodeTooManyRequests, msg, cause)
}

// ErrInvalidCredentials membentuk error INVALID_CREDENTIALS terstandar.
func ErrInvalidCredentials(msg string, cause error) error {
	if msg == "" {
		msg = "invalid credentials"
	}
	return NewGatewayError(CodeInvalidCredentials, msg, cause)
}

// ErrInvalidArgument membentuk error INVALID_ARGUMENT terstandar.
func ErrInvalidArgument(msg string, cause error) error {
	if msg == "" {
		msg = "invalid input"
	}
	return NewGatewayError(CodeInvalidArgument, msg, cause)
}

// ErrConflict membentuk error CONFLICT terstandar.
func ErrConflict(msg string, cause error) error {
	if msg == "" {
		msg = "resource state conflict"
	}
	return NewGatewayError(CodeConflict, msg, cause)
}

// ErrNotFound membentuk error NOT_FOUND terstandar.
func ErrNotFound(msg string, cause error) error {
	if msg == "" {
		msg = "resource not found"
	}
	return NewGatewayError(CodeNotFound, msg, cause)
}

func httpStatusForCode(code ErrorCode) int {
	switch code {
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeInvalidCredentials:
		return http.StatusUnauthorized
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeConflict:
		return http.StatusConflict
	case CodeUnprocessableEntity:
		return http.StatusUnprocessableEntity
	case CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// ErrUnprocessableEntity membentuk error UNPROCESSABLE_ENTITY untuk aturan bisnis yang tidak dapat diproses.
func ErrUnprocessableEntity(msg string, cause error) error {
	if msg == "" {
		msg = "request could not be processed"
	}
	return NewGatewayError(CodeUnprocessableEntity, msg, cause)
}

// ErrorTypeForCode memetakan kode error internal Gateway ke nama enum ErrorTypeEnum GraphQL.
func ErrorTypeForCode(code ErrorCode) string {
	switch code {
	case CodeInvalidArgument:
		return "ARGUMENT"
	case CodeUnauthenticated, CodeInvalidCredentials:
		return "AUTHENTICATION"
	case CodeForbidden:
		return "AUTHORIZATION"
	case CodeConflict:
		return "CONFLICT"
	case CodeNotFound:
		return "NOT_FOUND"
	case CodeTooManyRequests:
		return "THROTTLE"
	case CodeUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case CodeUnavailable:
		return "UNAVAILABLE"
	default:
		return "INTERNAL_ERROR"
	}
}

// FormatSafeCause membantu logging aman di sisi server tanpa mengubah pesan ke klien.
func FormatSafeCause(op string, err error) string {
	if err == nil {
		return op
	}
	return fmt.Sprintf("%s: %v", op, err)
}
