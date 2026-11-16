// Package response memetakan batas transport HRM tanpa membocorkan detail persistensi.
package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Error mengembalikan kode stabil tanpa SQL, kredensial, atau nilai data pribadi.
func Error(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	case errors.Is(err, hrerrors.ErrUnauthenticated):
		return classified(codes.Unauthenticated, "valid session required", "UNAUTHENTICATED", "")
	case errors.Is(err, hrerrors.ErrForbidden):
		return classified(codes.PermissionDenied, "access denied", "FORBIDDEN", "")
	case errors.Is(err, hrerrors.ErrInvalidInput):
		var validation *hrerrors.ValidationError
		if errors.As(err, &validation) {
			return classified(codes.InvalidArgument, "invalid request", "INVALID_INPUT", validation.Field)
		}
		return classified(codes.InvalidArgument, "invalid request", "INVALID_INPUT", "")
	case errors.Is(err, hrerrors.ErrNotFound):
		return classified(codes.NotFound, "record not found", "NOT_FOUND", "")
	case errors.Is(err, hrerrors.ErrAlreadyExists):
		return classified(codes.AlreadyExists, "record already exists", "ALREADY_EXISTS", "")
	case errors.Is(err, hrerrors.ErrUnavailable):
		return classified(codes.Unavailable, "required dependency unavailable", "UNAVAILABLE", "")
	case errors.Is(err, hrerrors.ErrConflict):
		return classified(codes.FailedPrecondition, "operation precondition not met", "CONFLICT", "")
	case errors.Is(err, hrerrors.ErrWorkflowRequired),
		errors.Is(err, hrerrors.ErrReferenceConflict),
		errors.Is(err, hrerrors.ErrConstraintViolation):
		return classified(codes.FailedPrecondition, "operation precondition not met", "FAILED_PRECONDITION", "")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

var safeField = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func classified(code codes.Code, message, reason, field string) error {
	s := status.New(code, message)
	info := &errdetails.ErrorInfo{Reason: reason, Domain: "human_resource"}
	var withDetails *status.Status
	var err error
	if field != "" && safeField.MatchString(field) {
		withDetails, err = s.WithDetails(info, &errdetails.BadRequest{FieldViolations: []*errdetails.BadRequest_FieldViolation{{Field: field, Description: "invalid field"}}})
	} else {
		withDetails, err = s.WithDetails(info)
	}
	if err != nil {
		return s.Err()
	}
	return withDetails.Err()
}

// TransportError menghapus detail status termasuk pesan dari error yang dibungkus.
func TransportError(err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok {
		return Error(err)
	}
	message := "internal error"
	reason := ""
	field := ""
	for _, detail := range s.Details() {
		if bad, ok := detail.(*errdetails.BadRequest); ok && len(bad.FieldViolations) == 1 {
			field = bad.FieldViolations[0].Field
		}
	}
	switch s.Code() {
	case codes.Canceled:
		message = "request canceled"
	case codes.DeadlineExceeded:
		message = "request deadline exceeded"
	case codes.Unauthenticated:
		message = "valid session required"
		reason = "UNAUTHENTICATED"
	case codes.PermissionDenied:
		message = "access denied"
		reason = "FORBIDDEN"
	case codes.InvalidArgument:
		message = "invalid request"
		reason = "INVALID_INPUT"
	case codes.NotFound:
		message = "record not found"
		reason = "NOT_FOUND"
	case codes.AlreadyExists:
		message = "record already exists"
		reason = "ALREADY_EXISTS"
	case codes.Unavailable:
		message = "required dependency unavailable"
		reason = "UNAVAILABLE"
	case codes.Aborted:
		message = "concurrent or conflicting operation"
	case codes.FailedPrecondition:
		message = "operation precondition not met"
		reason = "FAILED_PRECONDITION"
	case codes.ResourceExhausted:
		message = "request capacity exceeded"
	case codes.Unimplemented:
		message = "operation unavailable"
	default:
		return status.Error(codes.Internal, message)
	}
	if reason != "" {
		return classified(s.Code(), message, reason, field)
	}
	return status.Error(s.Code(), message)
}

// Date membaca tanggal kalender ISO tanpa normalisasi tanggal tidak valid.
func Date(raw string) (time.Time, error) {
	value, err := time.Parse("2006-01-02", raw)
	if err != nil || value.Format("2006-01-02") != raw {
		return time.Time{}, hrerrors.ErrInvalidInput
	}
	return value, nil
}

// OptionalTimestamp mempertahankan NULL terpisah dari waktu bernilai nol.
func OptionalTimestamp(value *timestamppb.Timestamp) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := Timestamp(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// JSON mempertahankan perbedaan SQL NULL dan nilai JSON null.
func JSON(value *common.JsonDocument) (entity.JSON, error) {
	if value == nil {
		return nil, nil
	}
	if !json.Valid(value.Utf8) {
		return nil, hrerrors.ErrInvalidInput
	}
	return entity.JSON(append([]byte(nil), value.Utf8...)), nil
}

// JSONValue memetakan hanya kolom JSON dinamis yang didefinisikan domain.
func JSONValue(value entity.JSON) (*common.JsonDocument, error) {
	if value == nil {
		return nil, nil
	}
	if !json.Valid(value) {
		return nil, fmt.Errorf("invalid stored JSON document")
	}
	return &common.JsonDocument{Utf8: append([]byte(nil), value...)}, nil
}

// ValidateMask menolak mask kosong, duplikat, wildcard, atau kolom yang tidak dapat diperbarui.
func ValidateMask(mask *fieldmaskpb.FieldMask, allowed ...string) (map[string]bool, error) {
	if mask == nil || len(mask.Paths) == 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	valid := make(map[string]bool, len(allowed))
	for _, path := range allowed {
		valid[path] = true
	}
	selected := make(map[string]bool, len(mask.Paths))
	for _, path := range mask.Paths {
		if !valid[path] || selected[path] {
			return nil, hrerrors.ErrInvalidInput
		}
		selected[path] = true
	}
	return selected, nil
}

// Timestamp memvalidasi waktu wajib sebelum pemetaan ke tipe domain.
func Timestamp(value *timestamppb.Timestamp) (time.Time, error) {
	if value == nil {
		return time.Time{}, hrerrors.ErrInvalidInput
	}
	if err := value.CheckValid(); err != nil {
		return time.Time{}, hrerrors.ErrInvalidInput
	}
	return value.AsTime(), nil
}
