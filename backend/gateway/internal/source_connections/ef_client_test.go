package source_connections

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubIAMAuthServer struct {
	iamv1.UnimplementedAuthServiceServer
	calls atomic.Int32
}

func (s *stubIAMAuthServer) CountUnusedRecoveryCodes(ctx context.Context, req *iamv1.CountUnusedRecoveryCodesRequest) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
	s.calls.Add(1)
	return &iamv1.CountUnusedRecoveryCodesResponse{Count: 4}, nil
}

type stubEFAddressServer struct {
	efv1.UnimplementedAddressServiceServer
	handler func(context.Context, *efv1.GetAddressByIdRequest) (*efv1.GetAddressByIdResponse, error)
}

func (s *stubEFAddressServer) GetAddressById(ctx context.Context, req *efv1.GetAddressByIdRequest) (*efv1.GetAddressByIdResponse, error) {
	if s.handler != nil {
		return s.handler(ctx, req)
	}
	return &efv1.GetAddressByIdResponse{
		Address: &efv1.Address{Id: req.GetId(), City: "Jakarta"},
	}, nil
}

func callerTestCtx() context.Context {
	return authcontext.WithCaller(context.Background(), authcontext.CallerContext{
		RequestID:     "req-test-1",
		PrincipalID:   "00000000-0000-0000-0000-000000000001",
		SessionID:     "00000000-0000-0000-0000-000000000002",
		Authenticated: true,
	})
}

func TestNewRegistry_EFDisabledAndEmptyTargetDoNotBreakIAM(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		efEnabled bool
		efTarget  string
		wantInit  bool
	}{
		{name: "ef disabled", efEnabled: false, efTarget: "", wantInit: false},
		{name: "ef enabled with empty target fails isolated", efEnabled: true, efTarget: "  ", wantInit: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reg, err := NewRegistry(Config{
				IAMTarget:  "127.0.0.1:9001",
				IAMTimeout: 2 * time.Second,
				EFEnabled:  tc.efEnabled,
				EFTarget:   tc.efTarget,
				EFTimeout:  2 * time.Second,
				HMACSecret: []byte("01234567890123456789012345678901"),
			})
			if err != nil {
				t.Fatalf("NewRegistry() should not fail when EF is disabled/empty: %v", err)
			}
			defer func() {
				if closeErr := reg.Close(); closeErr != nil {
					t.Errorf("reg.Close() error: %v", closeErr)
				}
			}()

			if reg.IAM.Auth == nil || reg.IAM.Users == nil {
				t.Fatalf("expected IAM clients to be initialized")
			}
			if (reg.EFInitError != nil) != tc.wantInit {
				t.Fatalf("EFInitError=%v wantInit=%v", reg.EFInitError, tc.wantInit)
			}

			_, callErr := InvokeEF(context.Background(), reg.EnterpriseFoundation, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
				return c.Addresses.GetAddressById(ctx, &efv1.GetAddressByIdRequest{Id: "1"})
			})
			gwErr, ok := middleware.AsGatewayError(callErr)
			if !ok || gwErr.Code != middleware.CodeUnavailable {
				t.Fatalf("expected GatewayError UNAVAILABLE, got %v", callErr)
			}
		})
	}
}

func TestRegistry_EFOutageDoesNotAffectIAMAndRecoversAutomatically(t *testing.T) {
	iamLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen IAM: %v", err)
	}
	iamSrv := grpc.NewServer()
	iamStub := &stubIAMAuthServer{}
	iamv1.RegisterAuthServiceServer(iamSrv, iamStub)
	go func() { _ = iamSrv.Serve(iamLis) }()
	defer iamSrv.Stop()

	efLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen EF: %v", err)
	}
	efAddr := efLis.Addr().String()
	efSrv := grpc.NewServer()
	efStub := &stubEFAddressServer{}
	efv1.RegisterAddressServiceServer(efSrv, efStub)
	go func() { _ = efSrv.Serve(efLis) }()

	reg, err := NewRegistry(Config{
		IAMTarget:  iamLis.Addr().String(),
		IAMTimeout: time.Second,
		EFEnabled:  true,
		EFTarget:   efAddr,
		EFTimeout:  300 * time.Millisecond,
		HMACSecret: []byte("01234567890123456789012345678901"),
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	defer func() { _ = reg.Close() }()

	ctx := callerTestCtx()

	// 1. Both IAM and EF succeed initially.
	iamRes, err := reg.IAM.Auth.CountUnusedRecoveryCodes(ctx, &iamv1.CountUnusedRecoveryCodesRequest{UserId: "u1"})
	if err != nil || iamRes.GetCount() != 4 {
		t.Fatalf("IAM call failed before EF outage: res=%v err=%v", iamRes, err)
	}
	res, err := InvokeEF(ctx, reg.EnterpriseFoundation, func(rpcCtx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		return c.Addresses.GetAddressById(rpcCtx, &efv1.GetAddressByIdRequest{Id: "addr-1"})
	})
	if err != nil || res.GetAddress().GetCity() != "Jakarta" {
		t.Fatalf("EF GetAddressById failed initially: res=%v err=%v", res, err)
	}

	// 2. Kill EF completely; EF fails cleanly without leaking internal address, while IAM still works!
	efSrv.Stop()
	_ = efLis.Close()

	_, efErr := InvokeEF(ctx, reg.EnterpriseFoundation, func(rpcCtx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		return c.Addresses.GetAddressById(rpcCtx, &efv1.GetAddressByIdRequest{Id: "addr-1"})
	})
	gwErr, ok := middleware.AsGatewayError(efErr)
	if !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE during EF outage, got %v", efErr)
	}
	if strings.Contains(gwErr.Message, "127.0.0.1") || strings.Contains(gwErr.Message, "connection refused") {
		t.Fatalf("sanitized message leaked internal detail: %q", gwErr.Message)
	}
	if _, err := reg.IAM.Auth.CountUnusedRecoveryCodes(ctx, &iamv1.CountUnusedRecoveryCodesRequest{UserId: "u1"}); err != nil {
		t.Fatalf("IAM MUST continue working during EF outage, got %v", err)
	}

	// 3. Restart EF on same address -> Registry recovers automatically without recreating client!
	efLis2, err := net.Listen("tcp", efAddr)
	if err != nil {
		t.Fatalf("re-listen EF on %s: %v", efAddr, err)
	}
	efSrv2 := grpc.NewServer()
	efv1.RegisterAddressServiceServer(efSrv2, &stubEFAddressServer{})
	go func() { _ = efSrv2.Serve(efLis2) }()
	defer efSrv2.Stop()

	deadline := time.Now().Add(4 * time.Second)
	var recovered bool
	for time.Now().Before(deadline) {
		res, err = InvokeEF(ctx, reg.EnterpriseFoundation, func(rpcCtx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
			return c.Addresses.GetAddressById(rpcCtx, &efv1.GetAddressByIdRequest{Id: "addr-2"})
		})
		if err == nil && res.GetAddress().GetCity() == "Jakarta" {
			recovered = true
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if !recovered {
		t.Fatalf("expected EF client to recover automatically after server restart")
	}
}

func TestInvokeEF_CircuitBreakerAndPanicNilSafety(t *testing.T) {
	t.Parallel()

	now := time.Unix(1700000000, 0)
	breaker := newCircuitBreaker(2, 5*time.Second, ErrEFCircuitOpen, func() time.Time { return now })

	clients := &EnterpriseFoundationClientSet{
		Addresses:         &stubEFAddressClient{err: status.Error(codes.Unavailable, "dial tcp 127.0.0.1:9002: connect: connection refused")},
		LegalEntities:     efv1.NewLegalEntityServiceClient(nil),
		Locations:         efv1.NewLocationServiceClient(nil),
		OrganizationUnits: efv1.NewOrganizationUnitServiceClient(nil),
		Parties:           efv1.NewPartyServiceClient(nil),
		enabled:           true,
		timeout:           100 * time.Millisecond,
		breaker:           breaker,
	}

	// Business errors MUST NOT trip circuit breaker.
	clients.Addresses = &stubEFAddressClient{err: status.Error(codes.NotFound, "address not found")}
	for i := range 5 {
		_, err := InvokeEF(context.Background(), clients, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
			return c.Addresses.GetAddressById(ctx, &efv1.GetAddressByIdRequest{Id: "missing"})
		})
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code != middleware.CodeNotFound {
			t.Fatalf("iter %d: expected NOT_FOUND, got %v", i, err)
		}
	}

	// Consecutive Unavailable calls trip circuit breaker after threshold=2.
	clients.Addresses = &stubEFAddressClient{err: status.Error(codes.Unavailable, "connection refused")}
	for range 2 {
		_, _ = InvokeEF(context.Background(), clients, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
			return c.Addresses.GetAddressById(ctx, &efv1.GetAddressByIdRequest{Id: "1"})
		})
	}

	var calledWhenOpen atomic.Bool
	_, openErr := InvokeEF(context.Background(), clients, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		calledWhenOpen.Store(true)
		return &efv1.GetAddressByIdResponse{}, nil
	})
	if calledWhenOpen.Load() {
		t.Fatalf("callback must not execute while circuit breaker is open")
	}
	if !errors.Is(openErr, ErrEFCircuitOpen) {
		t.Fatalf("expected ErrEFCircuitOpen in wrapped cause, got %v", openErr)
	}

	// Advance mock clock -> half-open probe succeeds -> breaker closes.
	now = now.Add(6 * time.Second)
	res, err := InvokeEF(context.Background(), clients, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		return &efv1.GetAddressByIdResponse{Address: &efv1.Address{Id: "ok"}}, nil
	})
	if err != nil || res.GetAddress().GetId() != "ok" {
		t.Fatalf("half-open probe failed: res=%v err=%v", res, err)
	}

	// Panic inside RPC callback recovers safely.
	_, panicErr := InvokeEF(context.Background(), clients, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		panic("boom")
	})
	gwErr, ok := middleware.AsGatewayError(panicErr)
	if !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE after panic recovery, got %v", panicErr)
	}

	// Nil response returned safely as internal error instead of nil dereference downstream.
	_, nilErr := InvokeEF(context.Background(), clients, func(ctx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		return nil, nil
	})
	gwErr, ok = middleware.AsGatewayError(nilErr)
	if !ok || gwErr.Code != middleware.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR on nil response, got %v", nilErr)
	}
}

func TestMapEFGRPCError_AndExtractors(t *testing.T) {
	t.Parallel()

	badSt, _ := status.New(codes.InvalidArgument, "city is required").WithDetails(
		&errdetails.ErrorInfo{Reason: "enterprise_foundation.invalid_input", Domain: "enterprise_foundation"},
		&errdetails.BadRequest{FieldViolations: []*errdetails.BadRequest_FieldViolation{{Field: "City", Description: "required"}}},
	)

	if field := ExtractEFErrorField(badSt.Err()); field != "city" {
		t.Fatalf("expected field=city, got %q", field)
	}
	if reason := ExtractEFErrorReason(badSt.Err()); reason != "enterprise_foundation.invalid_input" {
		t.Fatalf("expected reason=enterprise_foundation.invalid_input, got %q", reason)
	}

	cases := []struct {
		name     string
		err      error
		wantCode middleware.ErrorCode
	}{
		{name: "invalid argument", err: badSt.Err(), wantCode: middleware.CodeInvalidArgument},
		{name: "not found", err: status.Error(codes.NotFound, "not found"), wantCode: middleware.CodeNotFound},
		{name: "already exists", err: status.Error(codes.AlreadyExists, "duplicate code"), wantCode: middleware.CodeConflict},
		{name: "failed precondition", err: status.Error(codes.FailedPrecondition, "cannot deactivate primary"), wantCode: middleware.CodeConflict},
		{name: "permission denied", err: status.Error(codes.PermissionDenied, "denied"), wantCode: middleware.CodeForbidden},
		{name: "rate limited", err: status.Error(codes.ResourceExhausted, "limit"), wantCode: middleware.CodeTooManyRequests},
		{name: "internal unauthenticated maps to unavailable", err: status.Error(codes.Unauthenticated, "bad internal hmac"), wantCode: middleware.CodeUnavailable},
		{name: "unavailable", err: status.Error(codes.Unavailable, "dial tcp 127.0.0.1:9002: connection refused"), wantCode: middleware.CodeUnavailable},
		{name: "internal", err: status.Error(codes.Internal, "pq: syntax error"), wantCode: middleware.CodeInternalError},
		{name: "deadline exceeded", err: context.DeadlineExceeded, wantCode: middleware.CodeUnavailable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mapped := MapEFGRPCError(tc.err)
			gwErr, ok := middleware.AsGatewayError(mapped)
			if !ok {
				t.Fatalf("expected GatewayError, got %T (%v)", mapped, mapped)
			}
			if gwErr.Code != tc.wantCode {
				t.Fatalf("code=%s want=%s msg=%q", gwErr.Code, tc.wantCode, gwErr.Message)
			}
			if strings.Contains(gwErr.Message, "pq:") || strings.Contains(gwErr.Message, "127.0.0.1") {
				t.Fatalf("message leaked internal detail: %q", gwErr.Message)
			}
		})
	}
}

type stubEFAddressClient struct {
	efv1.AddressServiceClient
	err error
}

func (s *stubEFAddressClient) GetAddressById(ctx context.Context, in *efv1.GetAddressByIdRequest, opts ...grpc.CallOption) (*efv1.GetAddressByIdResponse, error) {
	return nil, s.err
}
