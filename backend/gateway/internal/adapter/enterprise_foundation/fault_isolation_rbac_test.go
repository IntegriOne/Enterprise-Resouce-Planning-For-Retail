package enterprise_foundation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEFAdapter_FailClosedRBACAndScopes(t *testing.T) {
	suite, reg := newEFMockSuite(t)
	authzEng := newTestAuthorizer(t, nil)
	adp, err := NewAdapter(reg.EnterpriseFoundation, authzEng, 2*time.Second)
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}

	// 1. Permintaan tanpa sesi harus ditolak fail-closed sebelum menyentuh gRPC.
	_, unauthErr := adp.Address(context.Background(), EncodeGID(ResourceAddress, testUUID1))
	gwErr, ok := middleware.AsGatewayError(unauthErr)
	if !ok || gwErr.Code != middleware.CodeUnauthenticated {
		t.Fatalf("expected UNAUTHENTICATED for query without session, got %v", unauthErr)
	}
	mutUnauth := adp.CreateAddress(context.Background(), model.CreateAddressInput{
		AddressLine1: "Jl. Gatot Subroto",
		City:         "Jakarta",
		CountryCode:  "ID",
	})
	if mutUnauth.WasSuccessful || len(mutUnauth.Errors) != 1 || mutUnauth.Errors[0].Type != model.ErrorTypeEnumAuthentication {
		t.Fatalf("expected AUTHENTICATION MutationError without session, got %+v", mutUnauth)
	}
	if suite.calls.Load() != 0 {
		t.Fatalf("expected 0 EF gRPC calls on unauthenticated requests, got %d", suite.calls.Load())
	}

	// 2. User tanpa izin Enterprise Foundation ditolak AUTHORIZATION sebelum menyentuh gRPC.
	noPermCtx := newTestSessionContext(t, false, false, time.Now(), []authz.PermissionGrant{
		{Code: authz.PermDashboardView, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
	})
	_, forbErr := adp.Addresses(noPermCtx, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	gwErr, ok = middleware.AsGatewayError(forbErr)
	if !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected FORBIDDEN without EF permission, got %v", forbErr)
	}
	if suite.calls.Load() != 0 {
		t.Fatalf("expected 0 EF gRPC calls when permission missing, got %d", suite.calls.Load())
	}

	// 3. Uji cakupan ABAC (legal_entity, facility, organization_unit, vendor_party, supplier_party).
	scopedGrants := []authz.PermissionGrant{
		{Code: PermLegalEntitiesRead, Scope: authz.ScopeGrant{Type: scopeTypeLegalEntity, ID: testUUID2}},
		{Code: PermLocationsRead, Scope: authz.ScopeGrant{Type: scopeTypeFacility, ID: testUUID3}},
		{Code: PermOrganizationUnitsRead, Scope: authz.ScopeGrant{Type: scopeTypeOrganizationUnit, ID: testUUID4}},
		{Code: PermPartiesRead, Scope: authz.ScopeGrant{Type: scopeTypeSupplierParty, ID: testUUID5}},
	}
	scopedCtx := newTestSessionContext(t, false, false, time.Now(), scopedGrants)

	// Scope cocok -> berhasil & membawa CallerContext tervalidasi HMAC ke EF.
	if _, err := adp.LegalEntity(scopedCtx, EncodeGID(ResourceLegalEntity, testUUID2)); err != nil {
		t.Fatalf("matching legal_entity scope failed: %v", err)
	}
	if _, err := adp.Location(scopedCtx, EncodeGID(ResourceLocation, testUUID3)); err != nil {
		t.Fatalf("matching facility scope failed: %v", err)
	}
	if _, err := adp.OrganizationUnit(scopedCtx, EncodeGID(ResourceOrganizationUnit, testUUID4)); err != nil {
		t.Fatalf("matching organization_unit scope failed: %v", err)
	}
	if _, err := adp.Party(scopedCtx, EncodeGID(ResourceParty, testUUID5)); err != nil {
		t.Fatalf("matching supplier_party scope failed: %v", err)
	}

	suite.mu.Lock()
	gotCaller, gotOK := suite.lastCaller, suite.hasCaller
	suite.mu.Unlock()
	if !gotOK || !gotCaller.Authenticated || gotCaller.PrincipalID != testUUIDUser || gotCaller.SessionID != testSessionID {
		t.Fatalf("expected signed HMAC CallerContext on EF server, got ok=%v caller=%+v", gotOK, gotCaller)
	}

	// Scope tidak cocok -> ditolak AUTHORIZATION sebelum RPC.
	callsBeforeMismatch := suite.calls.Load()
	for _, mismatchCall := range []func() error{
		func() error {
			_, err := adp.LegalEntity(scopedCtx, EncodeGID(ResourceLegalEntity, testUUID1))
			return err
		},
		func() error {
			_, err := adp.Location(scopedCtx, EncodeGID(ResourceLocation, testUUID1))
			return err
		},
		func() error {
			_, err := adp.OrganizationUnit(scopedCtx, EncodeGID(ResourceOrganizationUnit, testUUID1))
			return err
		},
		func() error {
			_, err := adp.Party(scopedCtx, EncodeGID(ResourceParty, testUUID1))
			return err
		},
	} {
		err := mismatchCall()
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code != middleware.CodeForbidden {
			t.Fatalf("expected FORBIDDEN on scope mismatch, got %v", err)
		}
	}
	if suite.calls.Load() != callsBeforeMismatch {
		t.Fatalf("expected 0 extra EF gRPC calls on scope mismatch")
	}

	// 4. Operasi sensitif wajib re-autentikasi < 10 menit dan step-up MFA bila aktif.
	staleCtx := newTestSessionContext(t, false, false, time.Now().Add(-15*time.Minute), allGlobalEFGrants())
	callsBeforeSensitive := suite.calls.Load()

	delRes := adp.DeleteAddress(staleCtx, EncodeGID(ResourceAddress, testUUID1))
	if delRes.WasSuccessful || len(delRes.Errors) == 0 || delRes.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("expected AUTHORIZATION error on stale DeleteAddress, got %+v", delRes)
	}
	dissolveRes := adp.DissolveLegalEntity(staleCtx, EncodeGID(ResourceLegalEntity, testUUID2), nil)
	if dissolveRes.WasSuccessful || len(dissolveRes.Errors) == 0 || dissolveRes.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("expected AUTHORIZATION error on stale DissolveLegalEntity, got %+v", dissolveRes)
	}
	portalRes := adp.ActivatePartyPortalAccess(staleCtx, EncodeGID(ResourcePartyContact, testUUID7), EncodeGID(ResourceIAMUser, testUUIDUser))
	if portalRes.WasSuccessful || len(portalRes.Errors) == 0 || portalRes.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("expected AUTHORIZATION error on stale ActivatePartyPortalAccess, got %+v", portalRes)
	}
	if suite.calls.Load() != callsBeforeSensitive {
		t.Fatalf("sensitive operations must fail closed before reaching EF gRPC")
	}

	// Sesi baru dengan MFA terverifikasi diizinkan.
	freshMFACtx := newTestSessionContext(t, true, true, time.Now(), allGlobalEFGrants())
	if res := adp.DeleteAddress(freshMFACtx, EncodeGID(ResourceAddress, testUUID1)); !res.WasSuccessful {
		t.Fatalf("expected DeleteAddress to succeed on fresh MFA session, got %+v", res)
	}
}

func TestEFAdapter_FaultIsolationAndRecovery(t *testing.T) {
	// 1. Adapter dengan client EF nil (misal EF dinonaktifkan) tidak menjatuhkan konstruksi atau panic.
	authzEng := newTestAuthorizer(t, nil)
	nilAdp, err := NewAdapter(nil, authzEng, 2*time.Second)
	if err != nil {
		t.Fatalf("NewAdapter(nil) should succeed for startup isolation, got %v", err)
	}
	freshCtx := newTestSessionContext(t, false, false, time.Now(), allGlobalEFGrants())

	if _, qErr := nilAdp.Address(freshCtx, EncodeGID(ResourceAddress, testUUID1)); qErr == nil {
		t.Fatalf("expected UNAVAILABLE query error when EF client is nil")
	} else if gwErr, ok := middleware.AsGatewayError(qErr); !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE, got %v", qErr)
	}
	nilMut := nilAdp.CreateAddress(freshCtx, model.CreateAddressInput{
		AddressLine1: "Jl. Thamrin",
		City:         "Jakarta",
		CountryCode:  "ID",
	})
	if nilMut.WasSuccessful || len(nilMut.Errors) != 1 || nilMut.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE mutation payload when EF client nil, got %+v", nilMut)
	}

	// 2. Uji downstream gRPC error mapping (HMAC Unauthenticated -> UNAVAILABLE, pesan SQL/host disanitasi).
	suite, reg := newEFMockSuite(t)
	adp, err := NewAdapter(reg.EnterpriseFoundation, authzEng, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}

	suite.mu.Lock()
	suite.addressErr = status.Error(codes.Unauthenticated, "invalid internal hmac signature from 127.0.0.1")
	suite.mu.Unlock()

	_, hmacErr := adp.Address(freshCtx, EncodeGID(ResourceAddress, testUUID1))
	gwErr, ok := middleware.AsGatewayError(hmacErr)
	if !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("downstream Unauthenticated must map to UNAVAILABLE (never log out browser user), got %v", hmacErr)
	}
	if strings.Contains(gwErr.Message, "127.0.0.1") || strings.Contains(gwErr.Message, "hmac") {
		t.Fatalf("downstream internal message leaked in query error: %q", gwErr.Message)
	}

	suite.mu.Lock()
	suite.addressErr = status.Error(codes.Internal, "pq: relation foundation_addresses does not exist at 10.0.0.5:5432")
	suite.mu.Unlock()

	_, internalErr := adp.Address(freshCtx, EncodeGID(ResourceAddress, testUUID1))
	gwErr, ok = middleware.AsGatewayError(internalErr)
	if !ok || gwErr.Code != middleware.CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %v", internalErr)
	}
	if strings.Contains(gwErr.Message, "foundation_addresses") || strings.Contains(gwErr.Message, "10.0.0.5") {
		t.Fatalf("SQL/host details leaked in client error message: %q", gwErr.Message)
	}

	suite.mu.Lock()
	suite.addressErr = nil
	suite.mu.Unlock()

	// 3. Hentikan server EF -> circuit breaker terbuka & hanya EF yang gagal, IAM tetap berjalan 100%.
	efAddr := suite.efAddr
	suite.stopEFServer()

	for i := 0; i < 6; i++ {
		_, outErr := adp.Address(freshCtx, EncodeGID(ResourceAddress, testUUID1))
		gwErr, ok := middleware.AsGatewayError(outErr)
		if !ok || gwErr.Code != middleware.CodeUnavailable {
			t.Fatalf("attempt %d: expected UNAVAILABLE during outage, got %v", i+1, outErr)
		}
		if strings.Contains(gwErr.Message, "127.0.0.1") || strings.Contains(gwErr.Message, "connection refused") {
			t.Fatalf("internal host/port detail leaked during outage: %q", gwErr.Message)
		}
	}

	mutOutage := adp.CreateAddress(freshCtx, model.CreateAddressInput{
		AddressLine1: "Jl. Asia Afrika",
		City:         "Bandung",
		CountryCode:  "ID",
	})
	if mutOutage.WasSuccessful || len(mutOutage.Errors) == 0 || mutOutage.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE MutationResult when circuit open, got %+v", mutOutage)
	}

	// IAM gRPC tetap 100% sehat selama EF mati total.
	if iamRes, iamErr := reg.IAM.Auth.CountUnusedRecoveryCodes(freshCtx, &iamv1.CountUnusedRecoveryCodesRequest{UserId: testUUIDUser}); iamErr != nil || iamRes.GetCount() != 3 {
		t.Fatalf("IAM must stay completely unaffected during EF outage: res=%v err=%v", iamRes, iamErr)
	}

	// 4. Nyalakan ulang server EF & tunggu cooldown circuit breaker (3s) -> pulih otomatis tanpa restart Gateway.
	suite.startEFServer(t, efAddr)
	time.Sleep(3100 * time.Millisecond)

	recoveredAddr, recErr := adp.Address(freshCtx, EncodeGID(ResourceAddress, testUUID1))
	if recErr != nil || recoveredAddr == nil || recoveredAddr.City != "Jakarta" {
		t.Fatalf("expected EF adapter to auto-recover after EF server restart, addr=%+v err=%v", recoveredAddr, recErr)
	}
}
