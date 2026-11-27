// Package scopeverifier mengimplementasikan verifikasi objek target scope eksternal melalui layanan pemiliknya.
package scopeverifier

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const (
	defaultEFTimeout        = 5 * time.Second
	defaultScopeCallerReqID = "iam-scope-verifier"
	callerDeadlineHeadroom  = 500 * time.Millisecond
)

// Config mengatur koneksi klien gRPC Enterprise Foundation untuk validasi target scope IAM.
type Config struct {
	Target     string
	Timeout    time.Duration
	HMACSecret []byte
}

// EFClients mengumpulkan stub gRPC layanan Enterprise Foundation yang memvalidasi kelayakan target scope.
type EFClients struct {
	LegalEntities     efv1.LegalEntityServiceClient
	OrganizationUnits efv1.OrganizationUnitServiceClient
	Locations         efv1.LocationServiceClient
	Parties           efv1.PartyServiceClient
}

// EFVerifier memverifikasi kelayakan objek scope eksternal melalui gRPC Enterprise Foundation secara fail-closed.
type EFVerifier struct {
	conn       *grpc.ClientConn
	clients    EFClients
	timeout    time.Duration
	hmacSecret []byte
}

var _ domainservice.ScopeOwnerVerifier = (*EFVerifier)(nil)

// NewEFVerifier membentuk verifier target scope berbasis klien gRPC non-blocking ke Enterprise Foundation.
func NewEFVerifier(cfg Config) (*EFVerifier, error) {
	target := strings.TrimSpace(cfg.Target)
	if target == "" {
		return nil, errors.New("enterprise foundation gRPC target is required")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultEFTimeout
	}
	secret := bytes.Clone(cfg.HMACSecret)

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(outgoingInternalContextInterceptor(secret)),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(false)),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  250 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   2 * time.Second,
			},
			MinConnectTimeout: 2 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                40 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create enterprise foundation gRPC client: %w", err)
	}

	return &EFVerifier{
		conn: conn,
		clients: EFClients{
			LegalEntities:     efv1.NewLegalEntityServiceClient(conn),
			OrganizationUnits: efv1.NewOrganizationUnitServiceClient(conn),
			Locations:         efv1.NewLocationServiceClient(conn),
			Parties:           efv1.NewPartyServiceClient(conn),
		},
		timeout:    timeout,
		hmacSecret: secret,
	}, nil
}

// NewEFVerifierWithClients membentuk verifier dari kumpulan stub klien gRPC untuk kebutuhan pengujian terisolasi.
func NewEFVerifierWithClients(clients EFClients, timeout time.Duration, hmacSecret []byte) *EFVerifier {
	if timeout <= 0 {
		timeout = defaultEFTimeout
	}
	return &EFVerifier{
		clients:    clients,
		timeout:    timeout,
		hmacSecret: bytes.Clone(hmacSecret),
	}
}

// Close menutup koneksi gRPC ke Enterprise Foundation bila ada.
func (v *EFVerifier) Close() error {
	if v == nil || v.conn == nil {
		return nil
	}
	return v.conn.Close()
}

// VerifyScope memvalidasi bahwa target scope eksternal terdaftar dan aktif/layak di Enterprise Foundation.
func (v *EFVerifier) VerifyScope(ctx context.Context, scope entity.Scope) error {
	if !scope.IsValid() {
		return iamerrors.ErrInvalidScope
	}
	if scope.Type == entity.ScopeTypeGlobal || scope.Type == entity.ScopeTypeSelf {
		return nil
	}
	if v == nil || scope.ID == nil {
		return iamerrors.ErrScopeOwnerVerificationUnavailable
	}
	targetID := strings.TrimSpace(*scope.ID)
	if targetID == "" {
		return iamerrors.ErrInvalidScope
	}

	rpcCtx, cancel := withBoundedTimeout(ctx, v.timeout)
	defer cancel()

	switch scope.Type {
	case entity.ScopeTypeLegalEntity:
		if v.clients.LegalEntities == nil {
			return iamerrors.ErrScopeOwnerVerificationUnavailable
		}
		resp, err := v.clients.LegalEntities.ValidateLegalEntityScopeTarget(rpcCtx, &efv1.ValidateLegalEntityScopeTargetRequest{Id: targetID})
		var decision *efv1.ScopeTargetDecision
		if resp != nil {
			decision = resp.GetDecision()
		}
		return validateDecision(ctx, decision, err, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY)

	case entity.ScopeTypeOrganizationUnit:
		if v.clients.OrganizationUnits == nil {
			return iamerrors.ErrScopeOwnerVerificationUnavailable
		}
		resp, err := v.clients.OrganizationUnits.ValidateOrganizationUnitScopeTarget(rpcCtx, &efv1.ValidateOrganizationUnitScopeTargetRequest{Id: targetID})
		var decision *efv1.ScopeTargetDecision
		if resp != nil {
			decision = resp.GetDecision()
		}
		return validateDecision(ctx, decision, err, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_ORGANIZATION_UNIT)

	case entity.ScopeTypeFacility:
		if v.clients.Locations == nil {
			return iamerrors.ErrScopeOwnerVerificationUnavailable
		}
		resp, err := v.clients.Locations.ValidateFacilityScope(rpcCtx, &efv1.ValidateFacilityScopeRequest{Id: targetID})
		var decision *efv1.ScopeTargetDecision
		if resp != nil {
			decision = resp.GetDecision()
		}
		return validateDecision(ctx, decision, err, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY)

	case entity.ScopeTypeSupplierParty:
		return v.verifyPartyRoleScope(ctx, rpcCtx, targetID, efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY)

	case entity.ScopeTypeVendorParty:
		return v.verifyPartyRoleScope(ctx, rpcCtx, targetID, efv1.PartyRoleKind_PARTY_ROLE_KIND_VENDOR, efv1.ScopeTargetKind_SCOPE_TARGET_KIND_VENDOR_PARTY)

	default:
		return iamerrors.ErrInvalidScope
	}
}

func (v *EFVerifier) verifyPartyRoleScope(
	parentCtx context.Context,
	rpcCtx context.Context,
	partyID string,
	roleKind efv1.PartyRoleKind,
	expectedKind efv1.ScopeTargetKind,
) error {
	if v.clients.Parties == nil {
		return iamerrors.ErrScopeOwnerVerificationUnavailable
	}
	resp, err := v.clients.Parties.ValidatePartyRoleScope(rpcCtx, &efv1.ValidatePartyRoleScopeRequest{
		PartyId:      partyID,
		RequiredRole: roleKind,
	})
	var decision *efv1.ScopeTargetDecision
	if resp != nil {
		decision = resp.GetDecision()
	}
	return validateDecision(parentCtx, decision, err, expectedKind)
}

func validateDecision(
	parentCtx context.Context,
	decision *efv1.ScopeTargetDecision,
	rpcErr error,
	expectedKind efv1.ScopeTargetKind,
) error {
	if rpcErr != nil {
		return mapEFVerificationError(parentCtx, rpcErr)
	}
	if decision == nil || !decision.GetEligible() || decision.GetScopeKind() != expectedKind {
		return iamerrors.WithField(iamerrors.ErrScopeTargetNotEligible, "scope_id", "scope target is not active or eligible")
	}
	return nil
}

func mapEFVerificationError(parentCtx context.Context, err error) error {
	if errors.Is(err, context.Canceled) && parentCtx != nil && parentCtx.Err() != nil {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.scope.owner_verification_unavailable"), "scope owner verification is unavailable")
	}

	st, ok := status.FromError(err)
	if !ok {
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.scope.owner_verification_unavailable"), "scope owner verification is unavailable")
	}

	switch st.Code() {
	case codes.Canceled:
		if parentCtx != nil && parentCtx.Err() != nil {
			return context.Canceled
		}
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.scope.owner_verification_unavailable"), "scope owner verification is unavailable")
	case codes.NotFound:
		return iamerrors.WithField(iamerrors.ErrScopeTargetNotFound, "scope_id", "scope target not found")
	case codes.InvalidArgument:
		return iamerrors.NewValidation("scope_id", iamerrors.Code("iam.rbac.invalid_scope"), "invalid scope target identifier")
	case codes.FailedPrecondition, codes.PermissionDenied:
		return iamerrors.WithField(iamerrors.ErrScopeTargetNotEligible, "scope_id", "scope target is not active or eligible")
	default:
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.scope.owner_verification_unavailable"), "scope owner verification is unavailable")
	}
}

func outgoingInternalContextInterceptor(hmacSecret []byte) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		outCtx := injectOutgoingCaller(ctx, hmacSecret)
		return invoker(outCtx, method, req, reply, cc, opts...)
	}
}

func injectOutgoingCaller(ctx context.Context, hmacSecret []byte) context.Context {
	caller, ok := authcontext.FromContext(ctx)
	if !ok {
		caller = authcontext.CallerContext{}
	}
	if strings.TrimSpace(caller.RequestID) == "" {
		caller.RequestID = defaultScopeCallerReqID
	}
	return authcontext.AppendToOutgoingGRPCContext(ctx, caller, hmacSecret, time.Now())
}

func withBoundedTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = defaultEFTimeout
	}
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return context.WithTimeout(ctx, time.Nanosecond)
		}
		// Sisakan ruang waktu sebelum deadline pemanggil (mis. Gateway 5s) agar timeout EF tetap dipetakan sebagai Unavailable terisolasi, bukan Canceled.
		if remaining > callerDeadlineHeadroom {
			budget := min(timeout, remaining-callerDeadlineHeadroom)
			return context.WithTimeout(ctx, budget)
		}
		if remaining <= timeout {
			return ctx, func() {}
		}
	}
	return context.WithTimeout(ctx, timeout)
}
