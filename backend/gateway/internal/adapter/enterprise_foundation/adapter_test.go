package enterprise_foundation

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	testUUID1     = "11111111-1111-4111-8111-111111111111"
	testUUID2     = "22222222-2222-4222-8222-222222222222"
	testUUID3     = "33333333-3333-4333-8333-333333333333"
	testUUID4     = "44444444-4444-4444-8444-444444444444"
	testUUID5     = "55555555-5555-4555-8555-555555555555"
	testUUID6     = "66666666-6666-4666-8666-666666666666"
	testUUID7     = "77777777-7777-4777-8777-777777777777"
	testUUID8     = "88888888-8888-4888-8888-888888888888"
	testUUIDUser  = "99999999-9999-4999-8999-999999999999"
	testHMACKey   = "01234567890123456789012345678901"
	testSessionID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

// fakeSessionStore menyediakan implementasi in-memory minimal SessionStorePort untuk pengujian unit adapter.
type fakeSessionStore struct{}

func (f *fakeSessionStore) ReadSessionContext(context.Context, string, time.Time) (*authsession.SessionValidationContext, error) {
	return nil, authsession.ErrSessionNotFound
}

func (f *fakeSessionStore) TouchActivity(context.Context, string, authsession.SessionTokenRecord, time.Duration, time.Time) error {
	return nil
}

func (f *fakeSessionStore) SetPreAuth(context.Context, authsession.PreAuthRecord, time.Duration) error {
	return nil
}

func (f *fakeSessionStore) GetPreAuth(context.Context, string, time.Time) (*authsession.PreAuthRecord, error) {
	return nil, authsession.ErrPreAuthNotFound
}

func (f *fakeSessionStore) DeletePreAuth(context.Context, string) error {
	return nil
}

func (f *fakeSessionStore) CheckRateLimit(context.Context, string, string, int64, time.Duration) (int64, error) {
	return 1, nil
}

func (f *fakeSessionStore) GetGate(context.Context) (*authsession.SecurityGateRecord, error) {
	return &authsession.SecurityGateRecord{Open: true}, nil
}

// fakeSnapshotRefresher mengembalikan snapshot deterministik untuk kebutuhan Authorizer pengujian.
type fakeSnapshotRefresher struct{}

func (f *fakeSnapshotRefresher) RefreshUserAccessSnapshot(_ context.Context, userID string, _ authcontext.CallerContext) (*authz.AccessSnapshot, error) {
	now := time.Now()
	return &authz.AccessSnapshot{
		PrincipalID:    userID,
		Status:         authz.StatusActive,
		FreshUntilUnix: now.Add(time.Minute).Unix(),
	}, nil
}

// fakeSensitiveChecker mengembalikan nil kecuali mode penolakan diaktifkan.
type fakeSensitiveChecker struct {
	err error
}

func (f *fakeSensitiveChecker) VerifySensitiveSession(context.Context, string, string, authz.ScopeGrant, authcontext.CallerContext) error {
	if f.err != nil {
		return f.err
	}
	return middleware.ErrForbidden("sensitive verification failed", nil)
}

func newTestAuthorizer(t *testing.T, sensitiveErr error) *middleware.Authorizer {
	t.Helper()
	loader := middleware.NewSnapshotLoader(&fakeSnapshotRefresher{}, time.Second, 30*time.Second)
	mgr, err := middleware.NewAuthManager(&fakeSessionStore{}, loader, middleware.SessionConfig{})
	if err != nil {
		t.Fatalf("NewAuthManager() failed: %v", err)
	}
	return middleware.NewAuthorizer(mgr, &fakeSensitiveChecker{err: sensitiveErr}, 10*time.Minute)
}

func newTestSessionContext(
	t *testing.T,
	mfaEnabled bool,
	mfaAuthenticated bool,
	issuedAt time.Time,
	grants []authz.PermissionGrant,
) context.Context {
	t.Helper()
	now := time.Now()
	sess := &middleware.ResolvedSession{
		RawToken:  "raw-session-token",
		TokenHash: "hash-session-token",
		Token: &authsession.SessionTokenRecord{
			SessionID:        testSessionID,
			UserID:           testUUIDUser,
			UserKind:         "staff",
			IssuedAtUnix:     issuedAt.Unix(),
			AbsExpiresAtUnix: now.Add(time.Hour).Unix(),
			MFAAuthenticated: mfaAuthenticated,
		},
		Security: &authsession.UserSecurityRecord{
			UserID: testUUIDUser,
			Status: authz.StatusActive,
		},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    testUUIDUser,
			UserKind:       "staff",
			Status:         authz.StatusActive,
			MFAEnabled:     mfaEnabled,
			Permissions:    grants,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(5 * time.Minute).Unix(),
		},
		Caller: authcontext.CallerContext{
			PrincipalID:   testUUIDUser,
			SessionID:     testSessionID,
			UserKind:      "staff",
			RequestID:     "req-ef-test",
			ClientIP:      "127.0.0.1",
			UserAgent:     "gateway-unit-test",
			Authenticated: true,
		},
	}
	return middleware.WithResolvedSession(context.Background(), sess)
}

func globalEFGrants(codes ...string) []authz.PermissionGrant {
	grants := make([]authz.PermissionGrant, 0, len(codes))
	for _, code := range codes {
		grants = append(grants, authz.PermissionGrant{
			Code:  code,
			Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
		})
	}
	return grants
}

func allGlobalEFGrants() []authz.PermissionGrant {
	return globalEFGrants(
		PermAddressesRead, PermAddressesWrite,
		PermLegalEntitiesRead, PermLegalEntitiesWrite,
		PermLocationsRead, PermLocationsWrite,
		PermOrganizationUnitsRead, PermOrganizationUnitsWrite,
		PermPartiesRead, PermPartiesWrite,
	)
}

// efMockSuite menyediakan server gRPC IAM & Enterprise Foundation tiruan di port lokal.
type efMockSuite struct {
	mu                sync.Mutex
	calls             atomic.Int32
	lastCaller        authcontext.CallerContext
	hasCaller         bool
	forcedErr         error
	addressErr        error
	iamCalls          atomic.Int32
	addresses         map[string]*efv1.Address
	legalEntities     map[string]*efv1.LegalEntity
	locations         map[string]*efv1.Location
	orgUnits          map[string]*efv1.OrganizationUnit
	parties           map[string]*efv1.Party
	partyRoles        map[string]*efv1.PartyRole
	partyContacts     map[string]*efv1.PartyContact
	partyAddressLinks map[string]*efv1.PartyAddress
	efAddr            string
	efLis             net.Listener
	efSrv             *grpc.Server
	iamLis            net.Listener
	iamSrv            *grpc.Server
	hmacSecret        []byte
}

func newEFMockSuite(t *testing.T) (*efMockSuite, *source_connections.Registry) {
	t.Helper()
	nowTS := timestamppb.New(time.Date(2026, 11, 16, 1, 0, 0, 0, time.UTC))
	labelHQ := "Kantor Pusat"
	cityJkt := "Jakarta"
	countryID := "ID"
	trading := "IntegriOne"
	note := "Primary supplier role"
	email := "vendor@integrione.local"

	s := &efMockSuite{
		hmacSecret: []byte(testHMACKey),
		addresses: map[string]*efv1.Address{
			testUUID1: {
				Id:            testUUID1,
				Label:         &labelHQ,
				AddressLine_1: "Jl. Sudirman Kav. 1",
				City:          cityJkt,
				CountryCode:   countryID,
				CreatedAt:     nowTS,
				UpdatedAt:     nowTS,
			},
		},
		legalEntities: map[string]*efv1.LegalEntity{
			testUUID2: {
				Id:                       testUUID2,
				EntityCode:               "LE-001",
				LegalName:                "PT IntegriOne Retail Nusantara",
				TradingName:              &trading,
				IsPrimaryOperatingEntity: true,
				Status:                   efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE,
				JurisdictionCountryCode:  "ID",
				RegisteredAddressId:      &[]string{testUUID1}[0],
				FiscalYearStartMonth:     1,
				CreatedAt:                nowTS,
				UpdatedAt:                nowTS,
			},
		},
		locations: map[string]*efv1.Location{
			testUUID3: {
				Id:            testUUID3,
				LegalEntityId: testUUID2,
				AddressId:     &[]string{testUUID1}[0],
				LocationCode:  "LOC-JKT-01",
				Name:          "Flagship Jakarta",
				Kind:          efv1.LocationKind_LOCATION_KIND_HEADQUARTERS,
				Timezone:      "Asia/Jakarta",
				IsActive:      true,
				CreatedAt:     nowTS,
				UpdatedAt:     nowTS,
			},
		},
		orgUnits: map[string]*efv1.OrganizationUnit{
			testUUID4: {
				Id:                testUUID4,
				LegalEntityId:     testUUID2,
				PrimaryLocationId: &[]string{testUUID3}[0],
				UnitCode:          "ORG-FIN",
				Name:              "Finance Division",
				Kind:              efv1.OrgUnitKind_ORG_UNIT_KIND_DIVISION,
				IsActive:          true,
				CreatedAt:         nowTS,
				UpdatedAt:         nowTS,
			},
		},
		parties: map[string]*efv1.Party{
			testUUID5: {
				Id:                testUUID5,
				PartnerCode:       "SUP-001",
				Kind:              efv1.PartyKind_PARTY_KIND_ORGANIZATION,
				DisplayName:       "PT Pemasok Utama",
				Email:             &email,
				PreferredLanguage: "id",
				Status:            efv1.PartyStatus_PARTY_STATUS_ACTIVE,
				CreatedAt:         nowTS,
				UpdatedAt:         nowTS,
			},
		},
		partyRoles: map[string]*efv1.PartyRole{
			testUUID6: {
				Id:            testUUID6,
				PartyId:       testUUID5,
				Role:          efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER,
				IsActive:      true,
				ReferenceNote: &note,
				CreatedAt:     nowTS,
				UpdatedAt:     nowTS,
			},
		},
		partyContacts: map[string]*efv1.PartyContact{
			testUUID7: {
				Id:                 testUUID7,
				PartyId:            testUUID5,
				IamUserId:          &[]string{testUUIDUser}[0],
				FullName:           "Budi Santoso",
				ContactRole:        "Account Manager",
				Email:              &email,
				PortalAccessStatus: efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_ACTIVE,
				IsPrimary:          true,
				IsActive:           true,
				CreatedAt:          nowTS,
				UpdatedAt:          nowTS,
			},
		},
		partyAddressLinks: map[string]*efv1.PartyAddress{
			testUUID8: {
				Id:        testUUID8,
				PartyId:   testUUID5,
				AddressId: testUUID1,
				Usage:     efv1.AddressUsageKind_ADDRESS_USAGE_KIND_REGISTERED_OFFICE,
				IsPrimary: true,
				CreatedAt: nowTS,
				UpdatedAt: nowTS,
			},
		},
	}

	iamLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen IAM: %v", err)
	}
	s.iamLis = iamLis
	s.iamSrv = grpc.NewServer()
	iamv1.RegisterAuthServiceServer(s.iamSrv, &mockIAMAuthServer{suite: s})
	go func() { _ = s.iamSrv.Serve(iamLis) }()

	s.startEFServer(t, "")

	reg, err := source_connections.NewRegistry(source_connections.Config{
		IAMTarget:  iamLis.Addr().String(),
		IAMTimeout: time.Second,
		EFEnabled:  true,
		EFTarget:   s.efAddr,
		EFTimeout:  500 * time.Millisecond,
		HMACSecret: s.hmacSecret,
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	t.Cleanup(func() {
		_ = reg.Close()
		s.stopEFServer()
		s.iamSrv.Stop()
		_ = s.iamLis.Close()
	})
	return s, reg
}

func (s *efMockSuite) startEFServer(t *testing.T, addr string) {
	t.Helper()
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen EF: %v", err)
	}
	srv := grpc.NewServer()
	efv1.RegisterAddressServiceServer(srv, &mockAddressServer{suite: s})
	efv1.RegisterLegalEntityServiceServer(srv, &mockLegalEntityServer{suite: s})
	efv1.RegisterLocationServiceServer(srv, &mockLocationServer{suite: s})
	efv1.RegisterOrganizationUnitServiceServer(srv, &mockOrgUnitServer{suite: s})
	efv1.RegisterPartyServiceServer(srv, &mockPartyServer{suite: s})

	s.mu.Lock()
	s.efLis = lis
	s.efAddr = lis.Addr().String()
	s.efSrv = srv
	s.mu.Unlock()

	go func() { _ = srv.Serve(lis) }()
}

func (s *efMockSuite) stopEFServer() {
	s.mu.Lock()
	srv := s.efSrv
	lis := s.efLis
	s.efSrv = nil
	s.efLis = nil
	s.mu.Unlock()
	if srv != nil {
		srv.Stop()
	}
	if lis != nil {
		_ = lis.Close()
	}
}

func (s *efMockSuite) beforeRPC(ctx context.Context) error {
	s.calls.Add(1)
	caller, ok, _ := authcontext.ExtractFromIncomingGRPCContext(ctx, s.hmacSecret, time.Now())
	s.mu.Lock()
	s.lastCaller = caller
	s.hasCaller = ok
	forced := s.forcedErr
	s.mu.Unlock()
	return forced
}

type mockIAMAuthServer struct {
	iamv1.UnimplementedAuthServiceServer
	suite *efMockSuite
}

func (m *mockIAMAuthServer) CountUnusedRecoveryCodes(context.Context, *iamv1.CountUnusedRecoveryCodesRequest) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
	m.suite.iamCalls.Add(1)
	return &iamv1.CountUnusedRecoveryCodesResponse{Count: 3}, nil
}

type mockAddressServer struct {
	efv1.UnimplementedAddressServiceServer
	suite *efMockSuite
}

func (m *mockAddressServer) GetAddressById(ctx context.Context, req *efv1.GetAddressByIdRequest) (*efv1.GetAddressByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	m.suite.mu.Lock()
	defer m.suite.mu.Unlock()
	if m.suite.addressErr != nil {
		return nil, m.suite.addressErr
	}
	return &efv1.GetAddressByIdResponse{Address: m.suite.addresses[req.GetId()]}, nil
}

func (m *mockAddressServer) ListAddresses(ctx context.Context, req *efv1.ListAddressesRequest) (*efv1.ListAddressesResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	page := req.GetPage()
	return &efv1.ListAddressesResponse{
		Items: []*efv1.Address{m.suite.addresses[testUUID1]},
		Page: &efv1.PageInfo{
			Total:   3,
			Limit:   page.GetLimit(),
			Offset:  page.GetOffset(),
			HasMore: page.GetOffset()+1 < 3,
		},
	}, nil
}

func (m *mockAddressServer) FindOrCreateAddress(ctx context.Context, req *efv1.FindOrCreateAddressRequest) (*efv1.FindOrCreateAddressResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	created := true
	return &efv1.FindOrCreateAddressResponse{
		Address: &efv1.Address{
			Id:            testUUID1,
			AddressLine_1: req.GetAddressLine_1(),
			City:          req.GetCity(),
			CountryCode:   req.GetCountryCode(),
		},
		Created: created,
	}, nil
}

func (m *mockAddressServer) CreateAddress(ctx context.Context, req *efv1.CreateAddressRequest) (*efv1.CreateAddressResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.CreateAddressResponse{
		Address: &efv1.Address{
			Id:            testUUID1,
			AddressLine_1: req.GetAddressLine_1(),
			City:          req.GetCity(),
			CountryCode:   req.GetCountryCode(),
		},
	}, nil
}

func (m *mockAddressServer) DeleteAddress(ctx context.Context, req *efv1.DeleteAddressRequest) (*efv1.DeleteAddressResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.DeleteAddressResponse{Deleted: true}, nil
}

func (m *mockAddressServer) IsAddressReferenced(ctx context.Context, _ *efv1.IsAddressReferencedRequest) (*efv1.IsAddressReferencedResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.IsAddressReferencedResponse{Referenced: true}, nil
}

type mockLegalEntityServer struct {
	efv1.UnimplementedLegalEntityServiceServer
	suite *efMockSuite
}

func (m *mockLegalEntityServer) GetLegalEntityById(ctx context.Context, req *efv1.GetLegalEntityByIdRequest) (*efv1.GetLegalEntityByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetLegalEntityByIdResponse{LegalEntity: m.suite.legalEntities[req.GetId()]}, nil
}

func (m *mockLegalEntityServer) ValidateLegalEntityScopeTarget(ctx context.Context, req *efv1.ValidateLegalEntityScopeTargetRequest) (*efv1.ValidateLegalEntityScopeTargetResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.ValidateLegalEntityScopeTargetResponse{
		Decision: &efv1.ScopeTargetDecision{
			Eligible:      true,
			ScopeKind:     efv1.ScopeTargetKind_SCOPE_TARGET_KIND_LEGAL_ENTITY,
			TargetId:      req.GetId(),
			Code:          "LE-001",
			LegalEntityId: req.GetId(),
		},
	}, nil
}

func (m *mockLegalEntityServer) DissolveLegalEntity(ctx context.Context, req *efv1.DissolveLegalEntityRequest) (*efv1.DissolveLegalEntityResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	le := m.suite.legalEntities[req.GetId()]
	return &efv1.DissolveLegalEntityResponse{LegalEntity: le}, nil
}

type mockLocationServer struct {
	efv1.UnimplementedLocationServiceServer
	suite *efMockSuite
}

func (m *mockLocationServer) GetLocationById(ctx context.Context, req *efv1.GetLocationByIdRequest) (*efv1.GetLocationByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetLocationByIdResponse{Location: m.suite.locations[req.GetId()]}, nil
}

func (m *mockLocationServer) ValidateWarehouseScope(ctx context.Context, req *efv1.ValidateWarehouseScopeRequest) (*efv1.ValidateWarehouseScopeResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.ValidateWarehouseScopeResponse{
		Decision: &efv1.ScopeTargetDecision{
			Eligible:      true,
			ScopeKind:     efv1.ScopeTargetKind_SCOPE_TARGET_KIND_FACILITY,
			TargetId:      req.GetId(),
			Code:          "WH-01",
			LegalEntityId: testUUID2,
		},
	}, nil
}

type mockOrgUnitServer struct {
	efv1.UnimplementedOrganizationUnitServiceServer
	suite *efMockSuite
}

func (m *mockOrgUnitServer) GetOrganizationUnitById(ctx context.Context, req *efv1.GetOrganizationUnitByIdRequest) (*efv1.GetOrganizationUnitByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetOrganizationUnitByIdResponse{OrganizationUnit: m.suite.orgUnits[req.GetId()]}, nil
}

type mockPartyServer struct {
	efv1.UnimplementedPartyServiceServer
	suite *efMockSuite
}

func (m *mockPartyServer) GetPartyById(ctx context.Context, req *efv1.GetPartyByIdRequest) (*efv1.GetPartyByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetPartyByIdResponse{Party: m.suite.parties[req.GetId()]}, nil
}

func (m *mockPartyServer) GetPartyRoleById(ctx context.Context, req *efv1.GetPartyRoleByIdRequest) (*efv1.GetPartyRoleByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetPartyRoleByIdResponse{Role: m.suite.partyRoles[req.GetRoleId()]}, nil
}

func (m *mockPartyServer) GetPartyContactById(ctx context.Context, req *efv1.GetPartyContactByIdRequest) (*efv1.GetPartyContactByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetPartyContactByIdResponse{Contact: m.suite.partyContacts[req.GetContactId()]}, nil
}

func (m *mockPartyServer) GetPartyAddressLinkById(ctx context.Context, req *efv1.GetPartyAddressLinkByIdRequest) (*efv1.GetPartyAddressLinkByIdResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.GetPartyAddressLinkByIdResponse{Link: m.suite.partyAddressLinks[req.GetLinkId()]}, nil
}

func (m *mockPartyServer) EvaluateRepresentativePortalAccess(ctx context.Context, req *efv1.EvaluateRepresentativePortalAccessRequest) (*efv1.EvaluateRepresentativePortalAccessResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.EvaluateRepresentativePortalAccessResponse{
		Evaluation: &efv1.PortalAccessEvaluation{
			Allowed:     true,
			PartyId:     testUUID5,
			ContactId:   testUUID7,
			IamUserId:   req.GetIamUserId(),
			ActiveRoles: []efv1.PartyRoleKind{efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER},
		},
	}, nil
}

func (m *mockPartyServer) ActivatePartyPortalAccess(ctx context.Context, req *efv1.ActivatePartyPortalAccessRequest) (*efv1.ActivatePartyPortalAccessResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	return &efv1.ActivatePartyPortalAccessResponse{Contact: m.suite.partyContacts[req.GetContactId()]}, nil
}

func (m *mockPartyServer) GetPartyRoleByPartyAndKind(ctx context.Context, req *efv1.GetPartyRoleByPartyAndKindRequest) (*efv1.GetPartyRoleByPartyAndKindResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	m.suite.mu.Lock()
	defer m.suite.mu.Unlock()
	for _, r := range m.suite.partyRoles {
		if r != nil && r.GetPartyId() == req.GetPartyId() && r.GetRole() == req.GetKind() {
			return &efv1.GetPartyRoleByPartyAndKindResponse{Role: r}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "party role not found")
}

func (m *mockPartyServer) AddPartyContact(ctx context.Context, req *efv1.AddPartyContactRequest) (*efv1.AddPartyContactResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	nowTS := timestamppb.New(time.Now())
	contact := &efv1.PartyContact{
		Id:                 "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		PartyId:            req.GetPartyId(),
		IamUserId:          req.IamUserId,
		FullName:           req.GetFullName(),
		JobTitle:           req.JobTitle,
		ContactRole:        req.GetContactRole(),
		Email:              req.Email,
		Phone:              req.Phone,
		PortalAccessStatus: req.GetPortalAccessStatus(),
		IsPrimary:          req.GetIsPrimary(),
		IsActive:           true,
		CreatedAt:          nowTS,
		UpdatedAt:          nowTS,
	}
	m.suite.mu.Lock()
	m.suite.partyContacts[contact.Id] = contact
	m.suite.mu.Unlock()
	return &efv1.AddPartyContactResponse{Contact: contact}, nil
}

func (m *mockPartyServer) RequestPartyPortalLink(ctx context.Context, req *efv1.RequestPartyPortalLinkRequest) (*efv1.RequestPartyPortalLinkResponse, error) {
	if err := m.suite.beforeRPC(ctx); err != nil {
		return nil, err
	}
	m.suite.mu.Lock()
	defer m.suite.mu.Unlock()
	contact := m.suite.partyContacts[req.GetContactId()]
	if contact == nil {
		return nil, status.Error(codes.NotFound, "party contact not found")
	}
	uid := req.GetIamUserId()
	contact.IamUserId = &uid
	contact.PortalAccessStatus = efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_PENDING_LINK
	return &efv1.RequestPartyPortalLinkResponse{Contact: contact}, nil
}

func TestGIDHelpers(t *testing.T) {
	t.Parallel()

	resources := []GIDResource{
		ResourceAddress,
		ResourceLegalEntity,
		ResourceLocation,
		ResourceOrganizationUnit,
		ResourceParty,
		ResourcePartyRole,
		ResourcePartyContact,
		ResourcePartyAddress,
	}

	for _, res := range resources {
		gid := EncodeGID(res, testUUID1)
		if !IsEFNodeID(gid) {
			t.Fatalf("IsEFNodeID(%q) = false, want true", gid)
		}
		gotRes, gotRaw, err := DecodeNodeGID(gid)
		if err != nil || gotRes != res || gotRaw != testUUID1 {
			t.Fatalf("DecodeNodeGID(%q) = (%q, %q, %v), want (%q, %q, nil)", gid, gotRes, gotRaw, err, res, testUUID1)
		}
		parsed, err := ParseTypedID(gid, res)
		if err != nil || parsed != testUUID1 {
			t.Fatalf("ParseTypedID(%q) = (%q, %v), want %q", gid, parsed, err, testUUID1)
		}
	}

	if IsEFNodeID(EncodeGID(ResourceIAMUser, testUUIDUser)) {
		t.Fatalf("IAM user GID should not be treated as EF node ID")
	}
	if got := EncodeOptionalGID(ResourceAddress, " "); got != nil {
		t.Fatalf("EncodeOptionalGID blank = %v, want nil", *got)
	}
	if _, err := ParseTypedID("not-a-gid", ResourceAddress); err == nil {
		t.Fatalf("expected error for non-GID value")
	}
	if _, err := ParseUUIDOrGID("invalid-uuid", ResourceLegalEntity); err == nil {
		t.Fatalf("expected error for non-UUID raw value")
	}
	parsedRaw, err := ParseUUIDOrGID(testUUID2, ResourceLegalEntity)
	if err != nil || parsedRaw != testUUID2 {
		t.Fatalf("ParseUUIDOrGID(uuid) = (%q, %v), want %q", parsedRaw, err, testUUID2)
	}
}

func TestRelayPaginationHelpers(t *testing.T) {
	t.Parallel()

	first2 := 2
	afterCur := EncodeCursor(0)
	win, err := ParseRelayPage(&afterCur, nil, &first2, nil)
	if err != nil {
		t.Fatalf("ParseRelayPage forward failed: %v", err)
	}
	if win.StartOffset != 1 || win.Limit != 2 || win.IsBackward {
		t.Fatalf("unexpected forward window: %+v", win)
	}

	neg := -1
	if _, err := ParseRelayPage(nil, nil, &neg, nil); err == nil {
		t.Fatalf("expected error for negative first")
	}
	if _, err := ParseRelayPage(nil, nil, &first2, &first2); err == nil {
		t.Fatalf("expected error when both first and last provided")
	}

	// Uji paginasi forward dan backward menggunakan FetchRelayPage.
	items, offset, pageInfo, total, err := FetchRelayPage(win, func(limit, reqOffset int32) ([]string, *efv1.PageInfo, error) {
		if limit != 2 || reqOffset != 1 {
			t.Fatalf("unexpected fetch args limit=%d offset=%d", limit, reqOffset)
		}
		return []string{"b", "c"}, &efv1.PageInfo{Total: 4, Limit: limit, Offset: reqOffset, HasMore: true}, nil
	})
	if err != nil {
		t.Fatalf("FetchRelayPage failed: %v", err)
	}
	if len(items) != 2 || offset != 1 || total != 4 || !pageInfo.HasPreviousPage || !pageInfo.HasNextPage {
		t.Fatalf("unexpected FetchRelayPage output: items=%v offset=%d total=%d pageInfo=%+v", items, offset, total, pageInfo)
	}

	zero := 0
	zeroWin, err := ParseRelayPage(nil, nil, &zero, nil)
	if err != nil {
		t.Fatalf("ParseRelayPage zero failed: %v", err)
	}
	zeroItems, _, zeroInfo, zeroTotal, err := FetchRelayPage(zeroWin, func(int32, int32) ([]string, *efv1.PageInfo, error) {
		return []string{"ignored"}, &efv1.PageInfo{Total: 2}, nil
	})
	if err != nil || len(zeroItems) != 0 || zeroTotal != 2 || !zeroInfo.HasNextPage {
		t.Fatalf("unexpected explicit zero page output: items=%v total=%d info=%+v err=%v", zeroItems, zeroTotal, zeroInfo, err)
	}
}

func TestToMutationErrors(t *testing.T) {
	t.Parallel()

	if errs := ToMutationErrors(nil); len(errs) != 0 {
		t.Fatalf("ToMutationErrors(nil) len=%d, want 0", len(errs))
	}
	errs := ToMutationErrors(middleware.ErrConflict("duplicate code", errors.New("db unique violation")))
	if len(errs) != 1 || errs[0].Type != model.ErrorTypeEnumConflict || errs[0].Message != "duplicate code" {
		t.Fatalf("unexpected conflict MutationError: %+v", errs)
	}
	unavail := ToMutationErrors(middleware.ErrUnavailable("ef down", nil))
	if len(unavail) != 1 || unavail[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("unexpected unavailable MutationError: %+v", unavail)
	}
}

func TestAdapter_NodeAndEntityMappers(t *testing.T) {
	suite, reg := newEFMockSuite(t)
	authzEng := newTestAuthorizer(t, nil)
	adp, err := NewAdapter(reg.EnterpriseFoundation, authzEng, 2*time.Second)
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}

	ctx := newTestSessionContext(t, false, false, time.Now(), allGlobalEFGrants())

	// Verifikasi semua 8 tipe Node GID EF dapat di-resolve oleh Adapter.Node.
	cases := []struct {
		res GIDResource
		id  string
	}{
		{ResourceAddress, testUUID1},
		{ResourceLegalEntity, testUUID2},
		{ResourceLocation, testUUID3},
		{ResourceOrganizationUnit, testUUID4},
		{ResourceParty, testUUID5},
		{ResourcePartyRole, testUUID6},
		{ResourcePartyContact, testUUID7},
		{ResourcePartyAddress, testUUID8},
	}

	for _, tc := range cases {
		node, err := adp.Node(ctx, EncodeGID(tc.res, tc.id))
		if err != nil {
			t.Fatalf("Node(%s) returned error: %v", tc.res, err)
		}
		if node == nil {
			t.Fatalf("Node(%s) returned nil", tc.res)
		}
	}

	// Verifikasi konversi timestamp WIB (+07:00) dan nested helpers.
	le, err := adp.LegalEntity(ctx, EncodeGID(ResourceLegalEntity, testUUID2))
	if err != nil {
		t.Fatalf("LegalEntity failed: %v", err)
	}
	if _, off := le.CreatedAt.Zone(); off != 7*3600 {
		t.Fatalf("expected WIB +07:00 zone offset, got %d", off)
	}
	regAddr, err := adp.ResolveLegalEntityRegisteredAddress(ctx, le)
	if err != nil || regAddr == nil || regAddr.RawID != testUUID1 {
		t.Fatalf("ResolveLegalEntityRegisteredAddress failed: addr=%+v err=%v", regAddr, err)
	}

	// Verifikasi FindOrCreateAddress & Scope/Portal helpers.
	createdPayload := adp.FindOrCreateAddress(ctx, model.FindOrCreateAddressInput{
		AddressLine1: "Jl. Thamrin 10",
		City:         "Jakarta",
		CountryCode:  "ID",
	})
	if !createdPayload.WasSuccessful || createdPayload.Created == nil || !*createdPayload.Created {
		t.Fatalf("FindOrCreateAddress payload unexpected: %+v", createdPayload)
	}
	decision, err := adp.ValidateWarehouseScope(ctx, EncodeGID(ResourceLocation, testUUID3))
	if err != nil || !decision.Eligible || decision.ScopeKind != model.ScopeTargetKindFacility {
		t.Fatalf("ValidateWarehouseScope unexpected: decision=%+v err=%v", decision, err)
	}
	eval, err := adp.EvaluateRepresentativePortalAccess(ctx, EncodeGID(ResourceParty, testUUID5), EncodeGID(ResourceIAMUser, testUUIDUser), model.PartyRoleKindSupplier)
	if err != nil || !eval.Allowed || len(eval.ActiveRoles) != 1 || eval.ActiveRoles[0] != model.PartyRoleKindSupplier {
		t.Fatalf("EvaluateRepresentativePortalAccess unexpected: eval=%+v err=%v", eval, err)
	}
	if suite.calls.Load() == 0 {
		t.Fatalf("expected real gRPC calls to be recorded")
	}
}
