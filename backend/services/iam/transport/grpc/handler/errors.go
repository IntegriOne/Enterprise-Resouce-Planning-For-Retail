// Package handler mengimplementasikan handler gRPC IAM, pemetaan proto-domain,
// serta konversi error domain IAM ke kode status gRPC secara aman.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const iamErrorDomain = "iam"

// ErrorMapper menerjemahkan error domain dan sistem menjadi error status gRPC
// tanpa membocorkan detail internal atau pesan basis data ke pemanggil.
type ErrorMapper struct {
	logger *slog.Logger
}

// NewErrorMapper membuat instance ErrorMapper baru dengan logger terstruktur opsional.
func NewErrorMapper(logger *slog.Logger) *ErrorMapper {
	return &ErrorMapper{logger: logger}
}

// ToStatus mengonversi error menjadi gRPC status.Error dengan detail terstruktur.
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

	if domErr, ok := errors.AsType[*iamerrors.DomainError](err); ok {
		return m.domainErrorToStatus(err, domErr)
	}

	// Jangan pernah membocorkan error non-domain ke pemanggil eksternal.
	if m != nil && m.logger != nil {
		m.logger.Error("unhandled internal error in IAM gRPC handler", slog.String("error", err.Error()))
	}
	return buildStatusWithDetails(codes.Internal, "internal error", iamErrorDomain, "iam.internal_error", "")
}

func (m *ErrorMapper) domainErrorToStatus(rootErr error, domErr *iamerrors.DomainError) error {
	grpcCode := mapKindToGRPCCode(domErr.Kind)
	reason := strings.TrimSpace(string(domErr.Code))
	if reason == "" {
		reason = "iam.error"
	}

	msg := sanitizedDomainMessage(domErr, grpcCode)
	if (grpcCode == codes.Internal || grpcCode == codes.Unavailable) && m != nil && m.logger != nil {
		m.logger.Error("infrastructure or internal IAM error",
			slog.String("code", reason),
			slog.String("grpc_code", grpcCode.String()),
			slog.String("error", rootErr.Error()),
		)
	}

	field := strings.TrimSpace(domErr.Field)
	return buildStatusWithDetails(grpcCode, msg, iamErrorDomain, reason, field)
}

func mapKindToGRPCCode(kind iamerrors.Kind) codes.Code {
	switch kind {
	case iamerrors.KindInvalidInput:
		return codes.InvalidArgument
	case iamerrors.KindNotFound:
		return codes.NotFound
	case iamerrors.KindAlreadyExists:
		return codes.AlreadyExists
	case iamerrors.KindUnauthorized:
		return codes.Unauthenticated
	case iamerrors.KindForbidden:
		return codes.PermissionDenied
	case iamerrors.KindConflict:
		return codes.FailedPrecondition
	case iamerrors.KindRateLimited:
		return codes.ResourceExhausted
	case iamerrors.KindUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

func sanitizedDomainMessage(domErr *iamerrors.DomainError, grpcCode codes.Code) string {
	// Untuk error Internal/Unknown jangan pernah mengambil isi domErr.Err mentah.
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
