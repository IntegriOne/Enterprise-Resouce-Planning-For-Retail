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
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubIAMSessionServer struct {
	iamv1.UnimplementedSessionServiceServer
	calls atomic.Int32
	fn    func(context.Context, *iamv1.RefreshUserAccessSnapshotRequest) (*iamv1.RefreshUserAccessSnapshotResponse, error)
}

func (s *stubIAMSessionServer) RefreshUserAccessSnapshot(ctx context.Context, req *iamv1.RefreshUserAccessSnapshotRequest) (*iamv1.RefreshUserAccessSnapshotResponse, error) {
	s.calls.Add(1)
	if s.fn != nil {
		return s.fn(ctx, req)
	}
	return &iamv1.RefreshUserAccessSnapshotResponse{
		Snapshot: &iamv1.AccessSnapshot{
			UserId:     req.GetUserId(),
			UserStatus: iamv1.UserStatus_USER_STATUS_ACTIVE,
		},
	}, nil
}

type dummyGRPCConn struct{}

func (dummyGRPCConn) Invoke(context.Context, string, any, any, ...grpc.CallOption) error {
	return nil
}

func (dummyGRPCConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func isUnavailableGatewayErr(err error) bool {
	gwErr, ok := middleware.AsGatewayError(err)
	return ok && gwErr.Code == middleware.CodeUnavailable
}

func TestNewRegistry_IAMUnavailableDoesNotBreakStartupOrEF(t *testing.T) {
	t.Parallel()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	efv1.RegisterAddressServiceServer(srv, &stubEFAddressServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)

	reg, err := NewRegistry(Config{
		IAMTarget:  "",
		IAMTimeout: 2 * time.Second,
		EFEnabled:  true,
		EFTarget:   lis.Addr().String(),
		EFTimeout:  2 * time.Second,
		HMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	})
	if err != nil {
		t.Fatalf("expected non-fatal NewRegistry when IAM target missing, got %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })

	if reg.IAMInitError == nil || reg.IAM.InitError() == nil {
		t.Fatal("expected IAMInitError to be populated when IAM target is empty")
	}
	if reg.IAM.Ready() {
		t.Fatal("expected degraded IAMClientSet.Ready() == false")
	}
	if !reg.IAM.Managed() {
		t.Fatal("expected degraded IAMClientSet from NewRegistry to report Managed() == true")
	}
	if reg.EFInitError != nil || !reg.EnterpriseFoundation.Ready() {
		t.Fatalf("expected EF client to remain ready when IAM fails init: efErr=%v ready=%v", reg.EFInitError, reg.EnterpriseFoundation.Ready())
	}

	_, iamErr := InvokeIAM(callerTestCtx(), reg.IAM, reg.IAM.Auth != nil, time.Second, func(rpcCtx context.Context, c IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		return c.Auth.CountUnusedRecoveryCodes(rpcCtx, &iamv1.CountUnusedRecoveryCodesRequest{})
	})
	gwErr, ok := middleware.AsGatewayError(iamErr)
	if !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE GatewayError from degraded IAM, got %v", iamErr)
	}

	addrResp, efErr := InvokeEF(callerTestCtx(), reg.EnterpriseFoundation, func(rpcCtx context.Context, c *EnterpriseFoundationClientSet) (*efv1.GetAddressByIdResponse, error) {
		return c.Addresses.GetAddressById(rpcCtx, &efv1.GetAddressByIdRequest{Id: "addr-1"})
	})
	if efErr != nil || addrResp.GetAddress().GetCity() != "Jakarta" {
		t.Fatalf("expected EF call to succeed while IAM unavailable, got resp=%v err=%v", addrResp, efErr)
	}
}

func TestInvokeIAM_CircuitBreakerPanicAndNilSafety(t *testing.T) {
	t.Parallel()

	ctx := callerTestCtx()

	// 1. Nil stub dan respons nil memetakan ke UNAVAILABLE tanpa panic.
	_, errNilStub := InvokeIAM(ctx, IAMClientSet{}, false, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		return nil, nil
	})
	if !isUnavailableGatewayErr(errNilStub) {
		t.Fatalf("expected UNAVAILABLE on nil IAM stub, got %v", errNilStub)
	}

	_, errNilResp := InvokeIAM(ctx, IAMClientSet{}, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		return nil, nil
	})
	gwNilResp, ok := middleware.AsGatewayError(errNilResp)
	if !ok || gwNilResp.Code != middleware.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR on nil IAM response, got %v", errNilResp)
	}

	// 2. Panic di dalam callback IAM dipulihkan aman menjadi UNAVAILABLE.
	_, errPanic := InvokeIAM(ctx, IAMClientSet{}, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		panic("boom iam")
	})
	if !isUnavailableGatewayErr(errPanic) {
		t.Fatalf("expected UNAVAILABLE after panic recovery, got %v", errPanic)
	}

	// 3. Circuit breaker trip setelah 5 kegagalan upstream dan pulih otomatis via probe half-open.
	now := time.Unix(1_700_000_000, 0)
	cb := newCircuitBreaker(5, 3*time.Second, ErrIAMCircuitOpen, func() time.Time { return now })
	clients := newIAMClientSetWithBreaker(dummyGRPCConn{}, time.Second, cb)

	var backendCalls int
	// Error bisnis tidak boleh memicu circuit breaker.
	for i := 0; i < 6; i++ {
		_, errBiz := InvokeIAM(ctx, clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
			backendCalls++
			return nil, status.Error(codes.NotFound, "user not found")
		})
		gwErr, ok := middleware.AsGatewayError(errBiz)
		if !ok || gwErr.Code != middleware.CodeNotFound {
			t.Fatalf("expected NOT_FOUND for business error, got %v", errBiz)
		}
	}
	// 4 kegagalan Unavailable lalu kegagalan ke-5 (codes.Canceled dari peer) membuka sirkuit.
	for i := 0; i < 4; i++ {
		_, _ = InvokeIAM(ctx, clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
			backendCalls++
			return nil, status.Error(codes.Unavailable, "dial tcp 172.18.0.5:9001: connection refused")
		})
	}
	_, _ = InvokeIAM(ctx, clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		backendCalls++
		return nil, status.Error(codes.Canceled, "grpc: the client connection is closing")
	})

	callsBeforeOpen := backendCalls
	_, errOpen := InvokeIAM(ctx, clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		backendCalls++
		return &iamv1.CountUnusedRecoveryCodesResponse{Count: 1}, nil
	})
	if !isUnavailableGatewayErr(errOpen) || !errors.Is(errOpen, ErrIAMCircuitOpen) {
		t.Fatalf("expected ErrIAMCircuitOpen, got %v", errOpen)
	}
	if backendCalls != callsBeforeOpen {
		t.Fatalf("expected fail-fast while IAM circuit open, got %d -> %d calls", callsBeforeOpen, backendCalls)
	}

	// Setelah jeda 3 detik, satu probe half-open menutup sirkuit kembali saat sukses.
	now = now.Add(4 * time.Second)
	respOK, errProbe := InvokeIAM(ctx, clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		backendCalls++
		return &iamv1.CountUnusedRecoveryCodesResponse{Count: 7}, nil
	})
	if errProbe != nil || respOK.GetCount() != 7 || backendCalls != callsBeforeOpen+1 {
		t.Fatalf("expected half-open probe recovery, got resp=%v err=%v calls=%d", respOK, errProbe, backendCalls)
	}
}

func TestMapIAMGRPCError_SanitizesAndBridgePropagation(t *testing.T) {
	t.Parallel()

	// 1. Pesan internal jaringan/SQL harus disanitasi pada Unavailable & Internal.
	rawDialErr := status.Error(codes.Unavailable, "dial tcp 127.0.0.1:9001 iam:9001: connection refused")
	mappedDial := MapIAMGRPCError(rawDialErr)
	gwDial, ok := middleware.AsGatewayError(mappedDial)
	if !ok || gwDial.Code != middleware.CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE, got %v", mappedDial)
	}
	for _, leaked := range []string{"127.0.0.1", "iam:9001", "dial tcp", "connection refused"} {
		if strings.Contains(strings.ToLower(gwDial.Message), leaked) {
			t.Fatalf("leaked internal detail %q in message %q", leaked, gwDial.Message)
		}
	}

	rawSQLErr := status.Error(codes.Internal, "pq: sqlstate 42P01 relation iam_users failed")
	mappedSQL := MapIAMGRPCError(rawSQLErr)
	gwSQL, ok := middleware.AsGatewayError(mappedSQL)
	if !ok || gwSQL.Code != middleware.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %v", mappedSQL)
	}
	if strings.Contains(strings.ToLower(gwSQL.Message), "sqlstate") || strings.Contains(strings.ToLower(gwSQL.Message), "pq:") {
		t.Fatalf("leaked SQL detail in message %q", gwSQL.Message)
	}

	// 2. ExtractIAMErrorReason & ExtractIAMErrorField dipertahankan melalui pembungkusan GatewayError.
	stAlready, _ := status.New(codes.AlreadyExists, "email already registered").WithDetails(
		&errdetails.ErrorInfo{Reason: "iam.identity.already_exists"},
		&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{{Field: "email"}},
		},
	)
	mappedAlready := MapIAMGRPCError(stAlready.Err())
	if got := ExtractIAMErrorReason(mappedAlready); got != "iam.identity.already_exists" {
		t.Fatalf("expected preserved reason iam.identity.already_exists, got %q", got)
	}
	if got := ExtractIAMErrorField(mappedAlready); got != "email" {
		t.Fatalf("expected preserved field email, got %q", got)
	}

	// 3. IAMRuntimeBridge nil dan target mati memetakan ke UNAVAILABLE tanpa hang/panic.
	var nilBridge *IAMRuntimeBridge
	if _, err := nilBridge.RefreshUserAccessSnapshot(context.Background(), "u-1", authcontext.CallerContext{}); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected UNAVAILABLE from nil bridge snapshot refresh, got %v", err)
	}
	if err := nilBridge.VerifySensitiveSession(context.Background(), "tok", "iam.users.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, authcontext.CallerContext{}); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected UNAVAILABLE from nil bridge sensitive verify, got %v", err)
	}

	// Loopback server IAM yang dihentikan lalu dijalankan menguji konektivitas & bridge nyata.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	sessStub := &stubIAMSessionServer{}
	iamv1.RegisterSessionServiceServer(srv, sessStub)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)

	reg, err := NewRegistry(Config{
		IAMTarget:  lis.Addr().String(),
		IAMTimeout: time.Second,
		HMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })

	bridge := NewIAMRuntimeBridge(reg.IAM, time.Second)
	snap, err := bridge.RefreshUserAccessSnapshot(callerTestCtx(), "user-99", authcontext.CallerContext{RequestID: "req-99"})
	if err != nil || snap == nil || snap.PrincipalID != "user-99" {
		t.Fatalf("expected live bridge snapshot, got snap=%+v err=%v", snap, err)
	}
}

func TestIAMScopeOwnerVerificationDoesNotTripBreakerAndErrorMappings(t *testing.T) {
	t.Parallel()

	stScope, _ := status.New(codes.Unavailable, "scope owner service unavailable").WithDetails(
		&errdetails.ErrorInfo{Reason: "iam.scope.owner_verification_unavailable"},
	)
	if isUpstreamUnavailableFailure(context.Background(), stScope.Err()) {
		t.Fatal("iam.scope.owner_verification_unavailable must not trip Gateway IAM circuit breaker")
	}
	if gwErr, ok := middleware.AsGatewayError(MapIAMGRPCError(stScope.Err())); !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected CodeUnavailable for iam.scope.owner_verification_unavailable, got %+v", gwErr)
	}

	stNotFound, _ := status.New(codes.NotFound, "session not found").WithDetails(
		&errdetails.ErrorInfo{Reason: "iam.session.not_found"},
	)
	if gwErr, ok := middleware.AsGatewayError(MapIAMGRPCError(stNotFound.Err())); !ok || gwErr.Code != middleware.CodeNotFound {
		t.Fatalf("expected CodeNotFound for iam.session.not_found, got %+v", gwErr)
	}

	stReauth, _ := status.New(codes.PermissionDenied, "sensitive reauth required").WithDetails(
		&errdetails.ErrorInfo{Reason: "iam.rbac.sensitive_reauth_required"},
	)
	if gwErr, ok := middleware.AsGatewayError(MapIAMGRPCError(stReauth.Err())); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected CodeForbidden for iam.rbac.sensitive_reauth_required, got %+v", gwErr)
	}

	stRoleInUse, _ := status.New(codes.FailedPrecondition, "role is in use").WithDetails(
		&errdetails.ErrorInfo{Reason: "iam.role.in_use"},
	)
	if gwErr, ok := middleware.AsGatewayError(MapIAMGRPCError(stRoleInUse.Err())); !ok || gwErr.Code != middleware.CodeConflict {
		t.Fatalf("expected CodeConflict for iam.role.in_use, got %+v", gwErr)
	}

	if gwErr, ok := middleware.AsGatewayError(MapIAMGRPCError(status.Error(codes.Aborted, "concurrent update"))); !ok || gwErr.Code != middleware.CodeConflict {
		t.Fatalf("expected CodeConflict for codes.Aborted, got %+v", gwErr)
	}

	for _, internalMsg := range []string{
		"internal caller verification is unavailable",
		"invalid internal caller context",
		"missing internal caller context",
	} {
		gwErr, ok := middleware.AsGatewayError(MapIAMGRPCError(status.Error(codes.Unauthenticated, internalMsg)))
		if !ok || gwErr.Code != middleware.CodeUnavailable {
			t.Fatalf("expected CodeUnavailable for internal S2S unauthenticated %q, got %+v", internalMsg, gwErr)
		}
	}
}

func TestInvokeIAM_HalfOpenProbeCallerCancelDoesNotResetBreaker(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	cb := newCircuitBreaker(1, 3*time.Second, ErrIAMCircuitOpen, func() time.Time { return now })
	clients := newIAMClientSetWithBreaker(dummyGRPCConn{}, time.Second, cb)

	// Picu 1 kegagalan upstream agar sirkuit terbuka.
	_, errFail := InvokeIAM(callerTestCtx(), clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		return nil, status.Error(codes.Unavailable, "dial tcp iam:9001: connection refused")
	})
	if !isUnavailableGatewayErr(errFail) {
		t.Fatalf("expected UNAVAILABLE to trip breaker, got %v", errFail)
	}

	// Masuk fase half-open lalu batalkan context pemanggil sebelum respons IAM diterima.
	now = now.Add(4 * time.Second)
	canceledCtx, cancel := context.WithCancel(callerTestCtx())
	cancel()
	_, errCanceled := InvokeIAM(canceledCtx, clients, true, time.Second, func(context.Context, IAMClientSet) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
		return nil, status.Error(codes.Canceled, "context canceled")
	})
	if !isUnavailableGatewayErr(errCanceled) {
		t.Fatalf("expected UNAVAILABLE on canceled half-open probe, got %v", errCanceled)
	}

	cb.mu.Lock()
	stateAfterCancel := cb.state
	halfOpenInFlight := cb.halfOpenInFlight
	cb.mu.Unlock()
	if stateAfterCancel != circuitOpen || halfOpenInFlight {
		t.Fatalf("expected breaker to stay open with halfOpenInFlight=false after caller-canceled probe, got state=%d inFlight=%v", stateAfterCancel, halfOpenInFlight)
	}
}
