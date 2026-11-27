package handler_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorMapper_ToStatus(t *testing.T) {
	t.Parallel()

	mapper := handler.NewErrorMapper(nil)

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		if got := mapper.ToStatus(nil); got != nil {
			t.Fatalf("expected nil status for nil error, got %v", got)
		}
	})

	t.Run("context errors map directly", func(t *testing.T) {
		t.Parallel()
		if got := status.Code(mapper.ToStatus(context.Canceled)); got != codes.Canceled {
			t.Fatalf("expected Canceled, got %v", got)
		}
		if got := status.Code(mapper.ToStatus(context.DeadlineExceeded)); got != codes.DeadlineExceeded {
			t.Fatalf("expected DeadlineExceeded, got %v", got)
		}
	})

	t.Run("domain kinds map to expected gRPC codes and details", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			err      error
			wantCode codes.Code
			wantMsg  string
			hasField bool
		}{
			{
				name:     "validation with field",
				err:      eferrors.NewValidation("entity_code", "enterprise_foundation.legal_entity.invalid_code", "entity code is required"),
				wantCode: codes.InvalidArgument,
				wantMsg:  "entity code is required",
				hasField: true,
			},
			{
				name:     "not found",
				err:      eferrors.ErrPartyNotFound,
				wantCode: codes.NotFound,
			},
			{
				name:     "already exists",
				err:      eferrors.Wrap(eferrors.ErrAlreadyExists, eferrors.KindAlreadyExists, "enterprise_foundation.address.duplicate", "address already exists"),
				wantCode: codes.AlreadyExists,
			},
			{
				name:     "unauthorized",
				err:      eferrors.ErrUnauthorized,
				wantCode: codes.Unauthenticated,
			},
			{
				name:     "forbidden",
				err:      eferrors.ErrForbidden,
				wantCode: codes.PermissionDenied,
			},
			{
				name:     "conflict",
				err:      eferrors.ErrConflict,
				wantCode: codes.FailedPrecondition,
			},
			{
				name:     "rate limited",
				err:      eferrors.ErrRateLimited,
				wantCode: codes.ResourceExhausted,
			},
			{
				name:     "unavailable hides internal cause",
				err:      eferrors.Wrap(errors.New("postgres secret dsn timeout"), eferrors.KindUnavailable, "enterprise_foundation.db.down", ""),
				wantCode: codes.Unavailable,
				wantMsg:  "service temporarily unavailable",
			},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				stErr := mapper.ToStatus(tc.err)
				st, ok := status.FromError(stErr)
				if !ok {
					t.Fatalf("expected gRPC status error, got %T", stErr)
				}
				if st.Code() != tc.wantCode {
					t.Fatalf("got code %v, want %v (msg=%s)", st.Code(), tc.wantCode, st.Message())
				}
				if tc.wantMsg != "" && st.Message() != tc.wantMsg {
					t.Fatalf("got message %q, want %q", st.Message(), tc.wantMsg)
				}
				if tc.wantCode == codes.Unavailable && strings.Contains(st.Message(), "secret") {
					t.Fatalf("internal detail leaked in unavailable message: %q", st.Message())
				}
				if tc.hasField {
					foundBadRequest := false
					for _, d := range st.Details() {
						if br, ok := d.(*errdetails.BadRequest); ok && len(br.GetFieldViolations()) > 0 {
							foundBadRequest = true
						}
					}
					if !foundBadRequest {
						t.Fatal("expected BadRequest field violation detail")
					}
				}
			})
		}
	})

	t.Run("raw non-domain error redacted as Internal", func(t *testing.T) {
		t.Parallel()
		stErr := mapper.ToStatus(errors.New("pq: password authentication failed for user secret_role"))
		st := status.Convert(stErr)
		if st.Code() != codes.Internal {
			t.Fatalf("got %v, want Internal", st.Code())
		}
		if strings.Contains(st.Message(), "password") || strings.Contains(st.Message(), "secret_role") {
			t.Fatalf("internal error message leaked: %q", st.Message())
		}
	})
}
