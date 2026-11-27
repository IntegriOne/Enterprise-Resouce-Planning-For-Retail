package scopeverifier

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type stubLegalEntityClient struct {
	efv1.LegalEntityServiceClient
	decision *efv1.ScopeTargetDecision
	err      error
	lastCtx  context.Context
	lastReq  *efv1.ValidateLegalEntityScopeTargetRequest
}

func (s *stubLegalEntityClient) ValidateLegalEntityScopeTarget(ctx context.Context, in *efv1.ValidateLegalEntityScopeTargetRequest, _ ...grpc.CallOption) (*efv1.ValidateLegalEntityScopeTargetResponse, error) {
	s.lastCtx = ctx
	s.lastReq = in
	if s.err != nil {
		return nil, s.err
	}
	return &efv1.ValidateLegalEntityScopeTargetResponse{Decision: s.decision}, nil
}

type stubOrgUnitClient struct {
	efv1.OrganizationUnitServiceClient
	decision *efv1.ScopeTargetDecision
	err      error
	lastReq  *efv1.ValidateOrganizationUnitScopeTargetRequest
}

func (s *stubOrgUnitClient) ValidateOrganizationUnitScopeTarget(_ context.Context, in *efv1.ValidateOrganizationUnitScopeTargetRequest, _ ...grpc.CallOption) (*efv1.ValidateOrganizationUnitScopeTargetResponse, error) {
	s.lastReq = in
	if s.err != nil {
		return nil, s.err
	}
	return &efv1.ValidateOrganizationUnitScopeTargetResponse{Decision: s.decision}, nil
}

type stubLocationClient struct {
	efv1.LocationServiceClient
	decision *efv1.ScopeTargetDecision
	err      error
	lastReq  *efv1.ValidateFacilityScopeRequest
}

func (s *stubLocationClient) ValidateFacilityScope(_ context.Context, in *efv1.ValidateFacilityScopeRequest, _ ...grpc.CallOption) (*efv1.ValidateFacilityScopeResponse, error) {
	s.lastReq = in
	if s.err != nil {
		return nil, s.err
	}
	return &efv1.ValidateFacilityScopeResponse{Decision: s.decision}, nil
}

type stubPartyClient struct {
	efv1.PartyServiceClient
	decision *efv1.ScopeTargetDecision
	err      error
	lastReq  *efv1.ValidatePartyRoleScopeRequest
}

func (s *stubPartyClient) ValidatePartyRoleScope(_ context.Context, in *efv1.ValidatePartyRoleScopeRequest, _ ...grpc.CallOption) (*efv1.ValidatePartyRoleScopeResponse, error) {
	s.lastReq = in
	if s.err != nil {
		return nil, s.err
	}
	return &efv1.ValidatePartyRoleScopeResponse{Decision: s.decision}, nil
}

func TestEFVerifier_VerifyScope(t *testing.T) {
	t.Parallel()
	targetID := "11111111-1111-4111-8111-111111111111"

	tests := []struct {
		name     string
		verifier *EFVerifier
		scope    entity.Scope
		wantErr  error
		wantKind iamerrors.Kind
	}{
		{
			name: "legal entity eligible",
			verifier: NewEFVerifierWithClients(EFClients{
				LegalEntities: &stubLegalEntityClient{
					decision: &efv1.ScopeTargetDecision{
						Eligible:  true,
						ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY,
						TargetId:  targetID,
					},
				},
			}, 2*time.Second, []byte("secret-key-for-scope-verifier-test-32b")),
			scope: entity.Scope{Type: entity.ScopeTypeLegalEntity, ID: &targetID},
		},
		{
			name: "organization unit eligible",
			verifier: NewEFVerifierWithClients(EFClients{
				OrganizationUnits: &stubOrgUnitClient{
					decision: &efv1.ScopeTargetDecision{
						Eligible:  true,
						ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_ORGANIZATION_UNIT,
						TargetId:  targetID,
					},
				},
			}, 2*time.Second, nil),
			scope: entity.Scope{Type: entity.ScopeTypeOrganizationUnit, ID: &targetID},
		},
		{
			name: "facility eligible",
			verifier: NewEFVerifierWithClients(EFClients{
				Locations: &stubLocationClient{
					decision: &efv1.ScopeTargetDecision{
						Eligible:  true,
						ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY,
						TargetId:  targetID,
					},
				},
			}, 2*time.Second, nil),
			scope: entity.Scope{Type: entity.ScopeTypeFacility, ID: &targetID},
		},
		{
			name: "supplier party eligible",
			verifier: NewEFVerifierWithClients(EFClients{
				Parties: &stubPartyClient{
					decision: &efv1.ScopeTargetDecision{
						Eligible:  true,
						ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY,
						TargetId:  targetID,
					},
				},
			}, 2*time.Second, nil),
			scope: entity.Scope{Type: entity.ScopeTypeSupplierParty, ID: &targetID},
		},
		{
			name: "vendor party eligible",
			verifier: NewEFVerifierWithClients(EFClients{
				Parties: &stubPartyClient{
					decision: &efv1.ScopeTargetDecision{
						Eligible:  true,
						ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_VENDOR_PARTY,
						TargetId:  targetID,
					},
				},
			}, 2*time.Second, nil),
			scope: entity.Scope{Type: entity.ScopeTypeVendorParty, ID: &targetID},
		},
		{
			name: "target not found maps fail-closed to ErrScopeTargetNotFound",
			verifier: NewEFVerifierWithClients(EFClients{
				LegalEntities: &stubLegalEntityClient{
					err: status.Error(codes.NotFound, "legal entity not found"),
				},
			}, 2*time.Second, nil),
			scope:    entity.Scope{Type: entity.ScopeTypeLegalEntity, ID: &targetID},
			wantErr:  iamerrors.ErrScopeTargetNotFound,
			wantKind: iamerrors.KindNotFound,
		},
		{
			name: "ineligible or inactive target (FailedPrecondition) maps to ErrScopeTargetNotEligible",
			verifier: NewEFVerifierWithClients(EFClients{
				Parties: &stubPartyClient{
					err: status.Error(codes.FailedPrecondition, "party is suspended"),
				},
			}, 2*time.Second, nil),
			scope:    entity.Scope{Type: entity.ScopeTypeSupplierParty, ID: &targetID},
			wantErr:  iamerrors.ErrScopeTargetNotEligible,
			wantKind: iamerrors.KindConflict,
		},
		{
			name: "ineligible or inactive target (PermissionDenied from EF) maps to ErrScopeTargetNotEligible",
			verifier: NewEFVerifierWithClients(EFClients{
				LegalEntities: &stubLegalEntityClient{
					err: status.Error(codes.PermissionDenied, "legal entity is not active"),
				},
			}, 2*time.Second, nil),
			scope:    entity.Scope{Type: entity.ScopeTypeLegalEntity, ID: &targetID},
			wantErr:  iamerrors.ErrScopeTargetNotEligible,
			wantKind: iamerrors.KindConflict,
		},
		{
			name: "mismatched decision kind rejects target",
			verifier: NewEFVerifierWithClients(EFClients{
				Parties: &stubPartyClient{
					decision: &efv1.ScopeTargetDecision{
						Eligible:  true,
						ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY,
						TargetId:  targetID,
					},
				},
			}, 2*time.Second, nil),
			scope:    entity.Scope{Type: entity.ScopeTypeVendorParty, ID: &targetID},
			wantErr:  iamerrors.ErrScopeTargetNotEligible,
			wantKind: iamerrors.KindConflict,
		},
		{
			name: "unavailable EF service preserves owner verification unavailable code",
			verifier: NewEFVerifierWithClients(EFClients{
				Locations: &stubLocationClient{
					err: status.Error(codes.Unavailable, "connection refused"),
				},
			}, 2*time.Second, nil),
			scope:    entity.Scope{Type: entity.ScopeTypeFacility, ID: &targetID},
			wantErr:  iamerrors.ErrScopeOwnerVerificationUnavailable,
			wantKind: iamerrors.KindUnavailable,
		},
		{
			name:     "nil stub fails closed as unavailable",
			verifier: NewEFVerifierWithClients(EFClients{}, 2*time.Second, nil),
			scope:    entity.Scope{Type: entity.ScopeTypeOrganizationUnit, ID: &targetID},
			wantErr:  iamerrors.ErrScopeOwnerVerificationUnavailable,
			wantKind: iamerrors.KindUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.verifier.VerifyScope(context.Background(), tt.scope)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("VerifyScope() unexpected error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("VerifyScope() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantKind != iamerrors.KindUnknown && iamerrors.KindOf(err) != tt.wantKind {
				t.Fatalf("VerifyScope() kind = %s, want %s", iamerrors.KindOf(err), tt.wantKind)
			}
		})
	}
}

func TestEFVerifier_InjectsOutgoingHMACProof(t *testing.T) {
	t.Parallel()
	secret := []byte("0123456789abcdef0123456789abcdef")
	interceptor := outgoingInternalContextInterceptor(secret)
	callerCtx := authcontext.WithCaller(context.Background(), authcontext.CallerContext{
		PrincipalID:   "22222222-2222-4222-8222-222222222222",
		SessionID:     "33333333-3333-4333-8333-333333333333",
		UserKind:      "employee",
		RequestID:     "req-scope-check",
		Authenticated: true,
	})
	var capturedCtx context.Context
	err := interceptor(
		callerCtx,
		"/enterprise_foundation.v1.LegalEntityService/ValidateLegalEntityScopeTarget",
		&efv1.ValidateLegalEntityScopeTargetRequest{Id: "11111111-1111-4111-8111-111111111111"},
		&efv1.ValidateLegalEntityScopeTargetResponse{},
		nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			capturedCtx = ctx
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor() error = %v", err)
	}
	md, ok := metadata.FromOutgoingContext(capturedCtx)
	if !ok {
		t.Fatal("expected outgoing gRPC metadata on verifier context")
	}
	if vals := md.Get(authcontext.MetaGatewayProof); len(vals) != 1 {
		t.Fatalf("expected exactly 1 %s metadata header, got %d", authcontext.MetaGatewayProof, len(vals))
	}
}

type blockingLegalEntityClient struct {
	efv1.LegalEntityServiceClient
}

func (b *blockingLegalEntityClient) ValidateLegalEntityScopeTarget(
	ctx context.Context,
	_ *efv1.ValidateLegalEntityScopeTargetRequest,
	_ ...grpc.CallOption,
) (*efv1.ValidateLegalEntityScopeTargetResponse, error) {
	<-ctx.Done()
	return nil, status.FromContextError(ctx.Err()).Err()
}

// TestEFVerifier_CallerDeadlineHeadroomPreservesFailClosedError memastikan timeout EF selalu berakhir lebih cepat daripada deadline pemanggil.
func TestEFVerifier_CallerDeadlineHeadroomPreservesFailClosedError(t *testing.T) {
	t.Parallel()

	v := NewEFVerifierWithClients(
		EFClients{LegalEntities: &blockingLegalEntityClient{}},
		5*time.Second,
		[]byte("0123456789abcdef0123456789abcdef"),
	)
	parentCtx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()

	id := "11111111-1111-4111-8111-111111111111"
	err := v.VerifyScope(parentCtx, entity.Scope{Type: entity.ScopeTypeLegalEntity, ID: &id})
	if !errors.Is(err, iamerrors.ErrScopeOwnerVerificationUnavailable) {
		t.Fatalf("expected ErrScopeOwnerVerificationUnavailable, got %v", err)
	}
	if parentCtx.Err() != nil {
		t.Fatalf("expected caller deadline headroom to preserve active parent context, got parentErr=%v", parentCtx.Err())
	}
}

type roundtripEFServer struct {
	efv1.UnimplementedLegalEntityServiceServer
	efv1.UnimplementedOrganizationUnitServiceServer
	efv1.UnimplementedLocationServiceServer
	efv1.UnimplementedPartyServiceServer
	secret  []byte
	partyID string
}

func (s *roundtripEFServer) verifyInternalCaller(ctx context.Context) error {
	caller, ok, err := authcontext.ExtractFromIncomingGRPCContext(ctx, s.secret, time.Now())
	if err != nil || !ok || caller.RequestID == "" {
		return status.Error(codes.Unauthenticated, "invalid internal caller proof")
	}
	return nil
}

func (s *roundtripEFServer) ValidateLegalEntityScopeTarget(ctx context.Context, req *efv1.ValidateLegalEntityScopeTargetRequest) (*efv1.ValidateLegalEntityScopeTargetResponse, error) {
	if err := s.verifyInternalCaller(ctx); err != nil {
		return nil, err
	}
	return &efv1.ValidateLegalEntityScopeTargetResponse{
		Decision: &efv1.ScopeTargetDecision{
			Eligible:      true,
			ScopeKind:     efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY,
			TargetId:      req.GetId(),
			Code:          "pt-test",
			LegalEntityId: req.GetId(),
		},
	}, nil
}

func (s *roundtripEFServer) ValidateOrganizationUnitScopeTarget(ctx context.Context, req *efv1.ValidateOrganizationUnitScopeTargetRequest) (*efv1.ValidateOrganizationUnitScopeTargetResponse, error) {
	if err := s.verifyInternalCaller(ctx); err != nil {
		return nil, err
	}
	return &efv1.ValidateOrganizationUnitScopeTargetResponse{
		Decision: &efv1.ScopeTargetDecision{
			Eligible:  true,
			ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_ORGANIZATION_UNIT,
			TargetId:  req.GetId(),
		},
	}, nil
}

func (s *roundtripEFServer) ValidateFacilityScope(ctx context.Context, req *efv1.ValidateFacilityScopeRequest) (*efv1.ValidateFacilityScopeResponse, error) {
	if err := s.verifyInternalCaller(ctx); err != nil {
		return nil, err
	}
	return &efv1.ValidateFacilityScopeResponse{
		Decision: &efv1.ScopeTargetDecision{
			Eligible:  true,
			ScopeKind: efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY,
			TargetId:  req.GetId(),
		},
	}, nil
}

func (s *roundtripEFServer) ValidatePartyRoleScope(ctx context.Context, req *efv1.ValidatePartyRoleScopeRequest) (*efv1.ValidatePartyRoleScopeResponse, error) {
	if err := s.verifyInternalCaller(ctx); err != nil {
		return nil, err
	}
	kind := efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY
	if req.GetRequiredRole() == efv1.PartyRoleKind_PARTY_ROLE_KIND_VENDOR {
		kind = efv1.ScopeTargetKind_SCOPE_TARGET_KIND_VENDOR_PARTY
	}
	return &efv1.ValidatePartyRoleScopeResponse{
		Decision: &efv1.ScopeTargetDecision{
			Eligible:  req.GetPartyId() == s.partyID,
			ScopeKind: kind,
			TargetId:  req.GetPartyId(),
		},
	}, nil
}

// TestEFVerifier_BufconnRoundtripAndHMACVerification memverifikasi kontrak gRPC serta tanda tangan HMAC ke server EF untuk ke-5 scope eksternal.
func TestEFVerifier_BufconnRoundtripAndHMACVerification(t *testing.T) {
	t.Parallel()

	secret := []byte("0123456789abcdef0123456789abcdef")
	targetID := "11111111-1111-4111-8111-111111111111"
	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() { _ = lis.Close() })

	srv := grpc.NewServer()
	handler := &roundtripEFServer{secret: secret, partyID: targetID}
	efv1.RegisterLegalEntityServiceServer(srv, handler)
	efv1.RegisterOrganizationUnitServiceServer(srv, handler)
	efv1.RegisterLocationServiceServer(srv, handler)
	efv1.RegisterPartyServiceServer(srv, handler)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(outgoingInternalContextInterceptor(secret)),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient(bufnet) error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	v := &EFVerifier{
		conn: conn,
		clients: EFClients{
			LegalEntities:     efv1.NewLegalEntityServiceClient(conn),
			OrganizationUnits: efv1.NewOrganizationUnitServiceClient(conn),
			Locations:         efv1.NewLocationServiceClient(conn),
			Parties:           efv1.NewPartyServiceClient(conn),
		},
		timeout:    2 * time.Second,
		hmacSecret: secret,
	}

	for _, st := range []entity.ScopeType{
		entity.ScopeTypeLegalEntity,
		entity.ScopeTypeOrganizationUnit,
		entity.ScopeTypeFacility,
		entity.ScopeTypeSupplierParty,
		entity.ScopeTypeVendorParty,
	} {
		if err := v.VerifyScope(context.Background(), entity.Scope{Type: st, ID: &targetID}); err != nil {
			t.Fatalf("VerifyScope(%s) over bufconn HMAC error = %v", st, err)
		}
	}

	otherPartyID := "22222222-2222-4222-8222-222222222222"
	if err := v.VerifyScope(context.Background(), entity.Scope{Type: entity.ScopeTypeVendorParty, ID: &otherPartyID}); !errors.Is(err, iamerrors.ErrScopeTargetNotEligible) {
		t.Fatalf("expected ErrScopeTargetNotEligible on ineligible party over bufconn, got %v", err)
	}
}
