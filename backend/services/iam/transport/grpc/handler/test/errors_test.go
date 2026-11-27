package handler_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
)

func TestErrorMapper_ToStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		err             error
		wantNil         bool
		wantCode        codes.Code
		wantReason      string
		wantField       string
		wantMsgContains string
		forbidSubstr    string
	}{
		{
			name:    "nil error produces nil status",
			err:     nil,
			wantNil: true,
		},
		{
			name:            "context canceled maps to Canceled",
			err:             context.Canceled,
			wantCode:        codes.Canceled,
			wantMsgContains: "canceled",
		},
		{
			name:            "context deadline exceeded maps to DeadlineExceeded",
			err:             context.DeadlineExceeded,
			wantCode:        codes.DeadlineExceeded,
			wantMsgContains: "deadline",
		},
		{
			name:            "validation domain error attaches ErrorInfo and BadRequest field violation",
			err:             iamerrors.NewValidation("email", iamerrors.Code("iam.identity.invalid_email"), "email format is invalid"),
			wantCode:        codes.InvalidArgument,
			wantReason:      "iam.identity.invalid_email",
			wantField:       "email",
			wantMsgContains: "email format is invalid",
		},
		{
			name:       "not found maps to NotFound",
			err:        iamerrors.ErrUserNotFound,
			wantCode:   codes.NotFound,
			wantReason: string(iamerrors.ErrUserNotFound.Code),
		},
		{
			name:     "already exists maps to AlreadyExists",
			err:      iamerrors.New(iamerrors.KindAlreadyExists, iamerrors.Code("iam.role.duplicate"), "role code already exists"),
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "unauthorized maps to Unauthenticated",
			err:      iamerrors.New(iamerrors.KindUnauthorized, iamerrors.Code("iam.auth.invalid_credentials"), "invalid credentials"),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "forbidden maps to PermissionDenied",
			err:      iamerrors.New(iamerrors.KindForbidden, iamerrors.Code("iam.authz.denied"), "permission denied"),
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "conflict maps to FailedPrecondition",
			err:      iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.session.ended"), "session has already ended"),
			wantCode: codes.FailedPrecondition,
		},
		{
			name:       "assigned permission cannot be deleted",
			err:        iamerrors.ErrPermissionInUse,
			wantCode:   codes.FailedPrecondition,
			wantReason: string(iamerrors.ErrPermissionInUse.Code),
		},
		{
			name:            "email already verified maps to FailedPrecondition with ErrorInfo",
			err:             iamerrors.ErrEmailAlreadyVerified,
			wantCode:        codes.FailedPrecondition,
			wantReason:      string(iamerrors.ErrEmailAlreadyVerified.Code),
			wantMsgContains: "your email has been verified",
		},
		{
			name:     "rate limited maps to ResourceExhausted",
			err:      iamerrors.New(iamerrors.KindRateLimited, iamerrors.Code("iam.auth.locked"), "too many failed attempts"),
			wantCode: codes.ResourceExhausted,
		},
		{
			name:            "unavailable sanitizes wrapped database secret error",
			err:             iamerrors.Wrap(errors.New("dial tcp 10.0.0.5:5432 password=topsecret"), iamerrors.KindUnavailable, iamerrors.Code("iam.db.unavailable"), "database unavailable"),
			wantCode:        codes.Unavailable,
			wantReason:      "iam.db.unavailable",
			wantMsgContains: "database unavailable",
			forbidSubstr:    "topsecret",
		},
		{
			name:         "unknown plain error sanitizes internal details",
			err:          errors.New("pq: relation users violates constraint secret_token_idx"),
			wantCode:     codes.Internal,
			wantReason:   "iam.internal_error",
			forbidSubstr: "secret_token_idx",
		},
		{
			name:         "wrapped grpc status sanitizes internal details",
			err:          errors.Join(errors.New("secret_token_idx"), status.Error(codes.Unavailable, "dial tcp password=topsecret")),
			wantCode:     codes.Internal,
			wantReason:   "iam.internal_error",
			forbidSubstr: "topsecret",
		},
		{
			name:            "withfield empty message inherits sentinel message",
			err:             iamerrors.WithField(iamerrors.ErrInvalidUserKind, "kind", ""),
			wantCode:        codes.InvalidArgument,
			wantReason:      string(iamerrors.ErrInvalidUserKind.Code),
			wantField:       "kind",
			wantMsgContains: "invalid user kind",
			forbidSubstr:    "iam.user.invalid_kind",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			mapper := handler.NewErrorMapper(logger)

			got := mapper.ToStatus(tc.err)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil error, got %v", got)
				}
				return
			}

			st, ok := status.FromError(got)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", got)
			}
			if st.Code() != tc.wantCode {
				t.Fatalf("expected code %s, got %s (msg=%q)", tc.wantCode, st.Code(), st.Message())
			}
			if tc.wantMsgContains != "" && !strings.Contains(st.Message(), tc.wantMsgContains) {
				t.Fatalf("expected message containing %q, got %q", tc.wantMsgContains, st.Message())
			}
			if tc.forbidSubstr != "" && strings.Contains(st.Message(), tc.forbidSubstr) {
				t.Fatalf("gRPC status message leaked sensitive substring %q: %q", tc.forbidSubstr, st.Message())
			}

			if tc.wantReason != "" {
				var foundReason bool
				for _, d := range st.Details() {
					if info, ok := d.(*errdetails.ErrorInfo); ok && info.Reason == tc.wantReason {
						foundReason = true
					}
				}
				if !foundReason {
					t.Fatalf("expected ErrorInfo reason %q in status details %+v", tc.wantReason, st.Details())
				}
			}

			if tc.wantField != "" {
				var foundField bool
				for _, d := range st.Details() {
					if br, ok := d.(*errdetails.BadRequest); ok {
						for _, v := range br.GetFieldViolations() {
							if v.GetField() == tc.wantField {
								foundField = true
							}
						}
					}
				}
				if !foundField {
					t.Fatalf("expected BadRequest field violation %q in details %+v", tc.wantField, st.Details())
				}
			}
		})
	}
}
