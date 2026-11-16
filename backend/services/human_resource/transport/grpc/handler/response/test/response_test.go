package response_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestErrorClassificationAndRedaction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{"cancel", context.Canceled, codes.Canceled},
		{"deadline", context.DeadlineExceeded, codes.DeadlineExceeded},
		{"unauthenticated", hrerrors.ErrUnauthenticated, codes.Unauthenticated},
		{"forbidden", hrerrors.ErrForbidden, codes.PermissionDenied},
		{"invalid", hrerrors.Invalid("email", "private@example.test"), codes.InvalidArgument},
		{"missing", hrerrors.ErrNotFound, codes.NotFound},
		{"duplicate", hrerrors.ErrAlreadyExists, codes.AlreadyExists},
		{"unavailable", hrerrors.ErrUnavailable, codes.Unavailable},
		{"conflict", hrerrors.ErrConflict, codes.FailedPrecondition},
		{"workflow", hrerrors.ErrWorkflowRequired, codes.FailedPrecondition},
		{"reference", hrerrors.ErrReferenceConflict, codes.FailedPrecondition},
		{"constraint", hrerrors.ErrConstraintViolation, codes.FailedPrecondition},
		{"transaction", hrerrors.ErrTransactionRequired, codes.Internal},
		{"unexpected", errors.New("password=secret SQL error"), codes.Internal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := response.Error(fmt.Errorf("private database detail: %w", test.err))
			if status.Code(err) != test.code {
				t.Fatalf("got %v want %v", status.Code(err), test.code)
			}
			if status.Convert(err).Message() == test.err.Error() {
				t.Fatal("raw dependency error exposed")
			}
		})
	}
	if response.Error(nil) != nil {
		t.Fatal("nil error changed")
	}
}

func TestPatchMaskValidation(t *testing.T) {
	t.Parallel()
	for _, paths := range [][]string{nil, {"*"}, {"id"}, {"name", "name"}} {
		if _, err := response.ValidateMask(&fieldmaskpb.FieldMask{Paths: paths}, "name"); err == nil {
			t.Fatalf("invalid mask accepted: %v", paths)
		}
	}
	if _, err := response.ValidateMask(&fieldmaskpb.FieldMask{Paths: []string{"name"}}, "name"); err != nil {
		t.Fatal(err)
	}
}

func TestJSONNullPresence(t *testing.T) {
	t.Parallel()
	missing, err := response.JSON(nil)
	if err != nil || missing != nil {
		t.Fatalf("SQL NULL lost: %v %v", missing, err)
	}
	literal, err := response.JSON(&common.JsonDocument{Utf8: []byte("null")})
	if err != nil || string(literal) != "null" {
		t.Fatalf("JSON null lost: %s %v", literal, err)
	}
	if value, err := response.JSONValue(missing); err != nil || value != nil {
		t.Fatal("SQL NULL round trip failed")
	}
	if value, err := response.JSONValue(literal); err != nil || value == nil {
		t.Fatal("JSON null round trip failed")
	}
}

func TestJSONPreservesLargeNumbersAndRejectsInvalidBytes(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"number":9007199254740993,"decimal":0.12345678901234567890}`)
	decoded, err := response.JSON(&common.JsonDocument{Utf8: raw})
	if err != nil || string(decoded) != string(raw) {
		t.Fatalf("JSON precision lost: %s %v", decoded, err)
	}
	result, err := response.JSONValue(decoded)
	if err != nil || string(result.Utf8) != string(raw) {
		t.Fatalf("JSON round trip changed: %s %v", result.GetUtf8(), err)
	}
	if _, err := response.JSON(&common.JsonDocument{Utf8: []byte("not json")}); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}

func TestErrorDetailsDoNotExposeInput(t *testing.T) {
	t.Parallel()
	err := response.TransportError(response.Error(hrerrors.Invalid("contact_reference_id", "private@example.test")))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid request code: %v", err)
	}
	var info *errdetails.ErrorInfo
	var bad *errdetails.BadRequest
	for _, detail := range status.Convert(err).Details() {
		switch typed := detail.(type) {
		case *errdetails.ErrorInfo:
			info = typed
		case *errdetails.BadRequest:
			bad = typed
		}
	}
	if info == nil || info.Reason != "INVALID_INPUT" || bad == nil || bad.FieldViolations[0].Field != "contact_reference_id" {
		t.Fatalf("safe error details missing: %v", err)
	}
	if strings.Contains(err.Error(), "private@example.test") {
		t.Fatal("private validation detail exposed")
	}
}
