package interceptor_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/interceptor"
)

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
