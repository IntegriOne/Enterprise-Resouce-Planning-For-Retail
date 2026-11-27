package interceptor_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/interceptor"
)

type stubServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *stubServerStream) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

func TestInternalContextInterceptorRejectsMissingProof(t *testing.T) {
	info := &grpc.UnaryServerInfo{FullMethod: "/iam.v1.UserService/GetUserById"}
	handler := func(context.Context, any) (any, error) {
		t.Fatal("handler must not run without internal proof")
		return nil, nil
	}
	for _, secret := range [][]byte{nil, []byte("0123456789abcdef0123456789abcdef")} {
		_, err := interceptor.InternalContextUnaryInterceptor(secret, true)(context.Background(), nil, info, handler)
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("missing proof error = %v, want Unauthenticated", err)
		}
	}
}

func TestRecoveryAndLoggingInterceptorsTolerateNilLoggerAndRecoverPanics(t *testing.T) {
	t.Parallel()

	info := &grpc.UnaryServerInfo{FullMethod: "/iam.v1.AuthService/AuthenticatePassword"}
	_, err := interceptor.RecoveryUnaryInterceptor(nil)(context.Background(), nil, info, func(context.Context, any) (any, error) {
		panic("boom-unary")
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("RecoveryUnaryInterceptor panic code = %v, want Internal", status.Code(err))
	}

	streamInfo := &grpc.StreamServerInfo{FullMethod: "/iam.v1.AuditService/StreamEvents"}
	err = interceptor.RecoveryStreamInterceptor(nil)(nil, &stubServerStream{}, streamInfo, func(any, grpc.ServerStream) error {
		panic("boom-stream")
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("RecoveryStreamInterceptor panic code = %v, want Internal", status.Code(err))
	}

	resp, err := interceptor.LoggingUnaryInterceptor(nil)(context.Background(), nil, info, func(context.Context, any) (any, error) {
		return "ok", nil
	})
	if err != nil || resp != "ok" {
		t.Fatalf("LoggingUnaryInterceptor(nil) = (%v, %v), want (ok, nil)", resp, err)
	}
}
