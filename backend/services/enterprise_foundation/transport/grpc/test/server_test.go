package grpc_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	transportgrpc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

const testBufSize = 1024 * 1024

type echoService interface {
	Ping(ctx context.Context, req *emptypb.Empty) (*wrapperspb.StringValue, error)
	Panic(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error)
}

type fakeEchoService struct{}

func (f *fakeEchoService) Ping(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	caller, ok := authcontext.FromContext(ctx)
	if !ok {
		return wrapperspb.String("anonymous"), nil
	}
	return wrapperspb.String(caller.PrincipalID + ":" + caller.UserKind), nil
}

func (f *fakeEchoService) Panic(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	panic("simulated panic")
}

func echoPingHandler(srv any, ctx context.Context, dec func(any) error, interceptor stdgrpc.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	impl := srv.(echoService)
	if interceptor == nil {
		return impl.Ping(ctx, in)
	}
	info := &stdgrpc.UnaryServerInfo{Server: srv, FullMethod: "/test.v1.Echo/Ping"}
	return interceptor(ctx, in, info, func(ctx context.Context, req any) (any, error) {
		return impl.Ping(ctx, req.(*emptypb.Empty))
	})
}

func echoPanicHandler(srv any, ctx context.Context, dec func(any) error, interceptor stdgrpc.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	impl := srv.(echoService)
	if interceptor == nil {
		return impl.Panic(ctx, in)
	}
	info := &stdgrpc.UnaryServerInfo{Server: srv, FullMethod: "/test.v1.Echo/Panic"}
	return interceptor(ctx, in, info, func(ctx context.Context, req any) (any, error) {
		return impl.Panic(ctx, req.(*emptypb.Empty))
	})
}

var echoServiceDesc = stdgrpc.ServiceDesc{
	ServiceName: "test.v1.Echo",
	HandlerType: (*echoService)(nil),
	Methods: []stdgrpc.MethodDesc{
		{MethodName: "Ping", Handler: echoPingHandler},
		{MethodName: "Panic", Handler: echoPanicHandler},
	},
}

func startTestServer(t *testing.T, hmacSecret []byte) (*transportgrpc.Server, *stdgrpc.ClientConn) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := transportgrpc.NewServer(transportgrpc.ServerConfig{
		ServiceName:               "integrione-enterprise-foundation",
		Address:                   "127.0.0.1:0",
		EnableReflection:          true,
		InternalGatewayHMACSecret: hmacSecret,
		ShutdownTimeout:           3 * time.Second,
		Logger:                    logger,
	}, handler.Dependencies{Logger: logger})
	if err != nil {
		t.Fatalf("NewServer() error: %v", err)
	}

	srv.GRPCServer().RegisterService(&echoServiceDesc, &fakeEchoService{})

	lis := bufconn.Listen(testBufSize)
	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- srv.Serve(lis)
	}()

	conn, err := stdgrpc.NewClient("passthrough:///bufnet",
		stdgrpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		stdgrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = lis.Close()
		_ = srv.Stop(context.Background())
		t.Fatalf("grpc.NewClient() error: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Stop(stopCtx); err != nil {
			t.Errorf("srv.Stop() error: %v", err)
		}
		_ = lis.Close()
		if err := <-serveErrCh; err != nil {
			t.Errorf("srv.Serve() error: %v", err)
		}
	})

	return srv, conn
}

func TestServer_HealthAndInterceptors(t *testing.T) {
	t.Parallel()

	hmacSecret := []byte(strings.Repeat("k", 32))
	_, conn := startTestServer(t, hmacSecret)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("healthcheck serves without gateway HMAC headers", func(t *testing.T) {
		healthClient := healthpb.NewHealthClient(conn)
		for _, target := range []string{"", "integrione-enterprise-foundation"} {
			resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: target})
			if err != nil {
				t.Fatalf("Health.Check(%q) error: %v", target, err)
			}
			if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
				t.Fatalf("Health.Check(%q) status=%v, expected SERVING", target, resp.GetStatus())
			}
		}
	})

	t.Run("unauthenticated business RPC rejected fail-closed", func(t *testing.T) {
		out := new(wrapperspb.StringValue)
		err := conn.Invoke(ctx, "/test.v1.Echo/Ping", &emptypb.Empty{}, out)
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("got code=%v (%v), expected Unauthenticated", status.Code(err), err)
		}
	})

	t.Run("tampered gateway HMAC proof rejected", func(t *testing.T) {
		badCtx := authcontext.AppendToOutgoingGRPCContext(ctx, authcontext.CallerContext{
			PrincipalID:   "00000000-0000-4000-8000-000000000010",
			SessionID:     "00000000-0000-4000-8000-000000000020",
			UserKind:      "internal",
			RequestID:     "req-bad",
			Authenticated: true,
		}, []byte(strings.Repeat("x", 32)), time.Now())

		out := new(wrapperspb.StringValue)
		err := conn.Invoke(badCtx, "/test.v1.Echo/Ping", &emptypb.Empty{}, out)
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("got code=%v (%v), expected Unauthenticated", status.Code(err), err)
		}
	})

	t.Run("valid gateway HMAC proof propagates caller context", func(t *testing.T) {
		signedCtx := authcontext.AppendToOutgoingGRPCContext(ctx, authcontext.CallerContext{
			PrincipalID:   "00000000-0000-4000-8000-000000000010",
			SessionID:     "00000000-0000-4000-8000-000000000020",
			UserKind:      "internal",
			RequestID:     "req-ok",
			Authenticated: true,
		}, hmacSecret, time.Now())

		out := new(wrapperspb.StringValue)
		if err := conn.Invoke(signedCtx, "/test.v1.Echo/Ping", &emptypb.Empty{}, out); err != nil {
			t.Fatalf("Invoke(/test.v1.Echo/Ping) unexpected error: %v", err)
		}
		if out.GetValue() != "00000000-0000-4000-8000-000000000010:internal" {
			t.Fatalf("got caller output=%q, expected principal:internal", out.GetValue())
		}
	})

	t.Run("panic recovered as Internal without crashing", func(t *testing.T) {
		signedCtx := authcontext.AppendToOutgoingGRPCContext(ctx, authcontext.CallerContext{
			PrincipalID:   "00000000-0000-4000-8000-000000000010",
			SessionID:     "00000000-0000-4000-8000-000000000020",
			UserKind:      "internal",
			RequestID:     "req-panic",
			Authenticated: true,
		}, hmacSecret, time.Now())

		out := new(emptypb.Empty)
		err := conn.Invoke(signedCtx, "/test.v1.Echo/Panic", &emptypb.Empty{}, out)
		if status.Code(err) != codes.Internal {
			t.Fatalf("got code=%v (%v), expected Internal", status.Code(err), err)
		}
	})
}

func TestErrorMapper_ToStatus(t *testing.T) {
	t.Parallel()

	mapper := handler.NewErrorMapper(slog.New(slog.NewTextHandler(io.Discard, nil)))

	t.Run("maps validation error with field details", func(t *testing.T) {
		err := eferrors.NewValidation("code", "enterprise_foundation.invalid_code", "code must not be empty")
		grpcErr := mapper.ToStatus(err)
		st := status.Convert(grpcErr)
		if st.Code() != codes.InvalidArgument {
			t.Fatalf("got code=%v, expected InvalidArgument", st.Code())
		}

		var hasErrorInfo, hasBadRequest bool
		for _, detail := range st.Details() {
			switch d := detail.(type) {
			case *errdetails.ErrorInfo:
				hasErrorInfo = d.GetDomain() == "enterprise_foundation" && d.GetReason() == "enterprise_foundation.invalid_code"
			case *errdetails.BadRequest:
				hasBadRequest = len(d.GetFieldViolations()) == 1 && d.GetFieldViolations()[0].GetField() == "code"
			}
		}
		if !hasErrorInfo || !hasBadRequest {
			t.Fatalf("expected ErrorInfo and BadRequest details, got: %+v", st.Details())
		}
	})

	t.Run("masks unhandled internal error details", func(t *testing.T) {
		grpcErr := mapper.ToStatus(errors.New("pgx connection refused at 10.0.0.5:5432"))
		st := status.Convert(grpcErr)
		if st.Code() != codes.Internal {
			t.Fatalf("got code=%v, expected Internal", st.Code())
		}
		if st.Message() != "internal error" {
			t.Fatalf("got message=%q, expected masked 'internal error'", st.Message())
		}
	})
}
