package interceptor_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/interceptor"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func signedContext(ctx context.Context, key []byte, now time.Time) context.Context {
	caller := authcontext.CallerContext{
		PrincipalID: "worker-principal", SessionID: "session", UserKind: "internal",
		AuthEpoch: 1, PermRevision: 2, Authenticated: true,
	}
	out := authcontext.AppendToOutgoingGRPCContext(ctx, caller, key, now)
	md, _ := metadata.FromOutgoingContext(out)
	return metadata.NewIncomingContext(ctx, md)
}

func TestUnaryAuthenticationAndSafeErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		code   codes.Code
		called bool
	}{
		{"success", codes.OK, true},
		{"missing proof", codes.Unauthenticated, false},
		{"wrong proof", codes.Unauthenticated, false},
		{"expired proof", codes.Unauthenticated, false},
		{"duplicate identity", codes.Unauthenticated, false},
		{"canceled", codes.Canceled, false},
		{"deadline", codes.DeadlineExceeded, false},
		{"dependency", codes.Unavailable, true},
		{"wrapped status", codes.PermissionDenied, true},
		{"panic", codes.Internal, true},
		{"short key", codes.Unauthenticated, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			key := []byte(strings.Repeat("k", 32))
			ctx := signedContext(t.Context(), key, time.Now())
			switch tc.name {
			case "missing proof":
				ctx = t.Context()
			case "wrong proof":
				ctx = signedContext(t.Context(), []byte(strings.Repeat("x", 32)), time.Now())
			case "expired proof":
				ctx = signedContext(t.Context(), key, time.Now().Add(-2*time.Minute))
			case "duplicate identity":
				md, _ := metadata.FromIncomingContext(ctx)
				md.Append(authcontext.MetaPrincipalID, "other")
				ctx = metadata.NewIncomingContext(ctx, md)
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "deadline":
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
			case "short key":
				key = []byte("short")
			}
			var logs bytes.Buffer
			guard := interceptor.Unary(key, slog.New(slog.NewJSONHandler(&logs, nil)))
			called := false
			value, err := guard(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/human_resource.workforceplanning.v1.WorkforcePlanningService/GetJob"}, func(ctx context.Context, req any) (any, error) {
				called = true
				caller, ok := authcontext.FromContext(ctx)
				if !ok || caller.PrincipalID != "worker-principal" {
					t.Error("verified identity missing")
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > 30*time.Second {
					t.Error("request deadline not bounded")
				}
				switch tc.name {
				case "dependency":
					return nil, fmt.Errorf("private-password-and-sql: %w", hrerrors.ErrUnavailable)
				case "wrapped status":
					return nil, fmt.Errorf("private-password-and-sql: %w", status.Error(codes.PermissionDenied, "private-payroll"))
				case "panic":
					panic("private-payroll")
				}
				return "ok", nil
			})
			if status.Code(err) != tc.code || called != tc.called {
				t.Fatalf("code=%s called=%v; want=%s/%v", status.Code(err), called, tc.code, tc.called)
			}
			if tc.code != codes.OK && value != nil {
				t.Fatal("failed request returned data")
			}
			if strings.Contains(logs.String(), "private-") || strings.Contains(status.Convert(err).Message(), "private-") {
				t.Fatal("private error details escaped boundary")
			}
		})
	}
}

func TestUnaryHealthAndCancellationDuringHandler(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	guard := interceptor.Unary([]byte(strings.Repeat("k", 32)), slog.New(slog.NewJSONHandler(&logs, nil)))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	value, err := guard(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}, func(context.Context, any) (any, error) {
		cancel()
		return "should not escape", nil
	})
	if status.Code(err) != codes.Canceled || value != nil {
		t.Fatalf("cancellation during handler: %v %v", value, err)
	}
}
