// Package handler mengimplementasikan registrasi handler gRPC Enterprise Foundation
// serta pemetaan error domain ke kode status gRPC secara aman.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

const efErrorDomain = "enterprise_foundation"

// ErrorMapper menerjemahkan error domain dan error sistem menjadi gRPC status.Error
// terstruktur tanpa membocorkan detail internal atau basis data ke pemanggil.
type ErrorMapper struct {
	logger *slog.Logger
}

// NewErrorMapper membuat instance ErrorMapper baru dengan logger terstruktur opsional.
func NewErrorMapper(logger *slog.Logger) *ErrorMapper {
	return &ErrorMapper{logger: logger}
}

// ToStatus mengonversi error menjadi gRPC status.Error lengkap dengan metadata ErrorInfo.
func (m *ErrorMapper) ToStatus(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	}

	if domErr, ok := errors.AsType[*eferrors.DomainError](err); ok {
		return m.domainErrorToStatus(err, domErr)
	}

	// Jangan membocorkan pesan error non-domain internal ke pemanggil.
	if m != nil && m.logger != nil {
		m.logger.Error("unhandled internal error in Enterprise Foundation gRPC handler",
			slog.String("error", err.Error()),
		)
	}
	return buildStatusWithDetails(codes.Internal, "internal error", efErrorDomain, "enterprise_foundation.internal_error", "")
}

func (m *ErrorMapper) domainErrorToStatus(rootErr error, domErr *eferrors.DomainError) error {
	grpcCode := mapKindToGRPCCode(domErr.Kind)
	reason := strings.TrimSpace(string(domErr.Code))
	if reason == "" {
		reason = "enterprise_foundation.error"
	}

	msg := sanitizedDomainMessage(domErr, grpcCode)
	if (grpcCode == codes.Internal || grpcCode == codes.Unavailable) && m != nil && m.logger != nil {
		m.logger.Error("infrastructure or internal Enterprise Foundation error",
			slog.String("code", reason),
			slog.String("grpc_code", grpcCode.String()),
			slog.String("error", rootErr.Error()),
		)
	}

	field := strings.TrimSpace(domErr.Field)
	return buildStatusWithDetails(grpcCode, msg, efErrorDomain, reason, field)
}

func mapKindToGRPCCode(kind eferrors.Kind) codes.Code {
	switch kind {
	case eferrors.KindInvalidInput:
		return codes.InvalidArgument
	case eferrors.KindNotFound:
		return codes.NotFound
	case eferrors.KindAlreadyExists:
		return codes.AlreadyExists
	case eferrors.KindUnauthorized:
		return codes.Unauthenticated
	case eferrors.KindForbidden:
		return codes.PermissionDenied
	case eferrors.KindConflict:
		return codes.FailedPrecondition
	case eferrors.KindRateLimited:
		return codes.ResourceExhausted
	case eferrors.KindUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

func sanitizedDomainMessage(domErr *eferrors.DomainError, grpcCode codes.Code) string {
	if grpcCode == codes.Internal {
		return "internal error"
	}
	msg := strings.TrimSpace(domErr.Message)
	if msg != "" {
		return msg
	}
	if grpcCode == codes.Unavailable {
		return "service temporarily unavailable"
	}
	if strings.TrimSpace(string(domErr.Code)) != "" {
		return string(domErr.Code)
	}
	return domErr.Kind.String()
}

func buildStatusWithDetails(grpcCode codes.Code, msg, domainName, reason, field string) error {
	st := status.New(grpcCode, msg)
	detailed, err := attachDetails(st, reason, domainName, field, msg)
	if err != nil {
		return st.Err()
	}
	return detailed.Err()
}

func attachDetails(st *status.Status, reason, domainName, field, msg string) (*status.Status, error) {
	info := &errdetails.ErrorInfo{
		Reason: reason,
		Domain: domainName,
	}
	if field == "" {
		return st.WithDetails(info)
	}
	badReq := &errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       field,
				Description: msg,
			},
		},
	}
	return st.WithDetails(info, badReq)
}
