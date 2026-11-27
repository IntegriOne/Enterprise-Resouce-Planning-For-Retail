package enterprisefoundation_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	efclient "github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/integration/enterprise_foundation"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type foundationStubServer struct {
	efv1.UnimplementedLegalEntityServiceServer
	efv1.UnimplementedOrganizationUnitServiceServer
	efv1.UnimplementedLocationServiceServer
	efv1.UnimplementedPartyServiceServer

	legalEntities map[string]*efv1.LegalEntity
	orgUnits      map[string]*efv1.OrganizationUnit
	locations     map[string]*efv1.Location
	parties       map[string]*efv1.Party
	contacts      map[string]*efv1.PartyContact
}

func (s *foundationStubServer) GetLegalEntityById(_ context.Context, req *efv1.GetLegalEntityByIdRequest) (*efv1.GetLegalEntityByIdResponse, error) {
	item, ok := s.legalEntities[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "legal entity not found")
	}
	return &efv1.GetLegalEntityByIdResponse{LegalEntity: item}, nil
}

func (s *foundationStubServer) GetOrganizationUnitById(_ context.Context, req *efv1.GetOrganizationUnitByIdRequest) (*efv1.GetOrganizationUnitByIdResponse, error) {
	item, ok := s.orgUnits[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "organization unit not found")
	}
	return &efv1.GetOrganizationUnitByIdResponse{OrganizationUnit: item}, nil
}

func (s *foundationStubServer) GetLocationById(_ context.Context, req *efv1.GetLocationByIdRequest) (*efv1.GetLocationByIdResponse, error) {
	item, ok := s.locations[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "location not found")
	}
	return &efv1.GetLocationByIdResponse{Location: item}, nil
}

func (s *foundationStubServer) GetPartyById(_ context.Context, req *efv1.GetPartyByIdRequest) (*efv1.GetPartyByIdResponse, error) {
	item, ok := s.parties[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "party not found")
	}
	return &efv1.GetPartyByIdResponse{Party: item}, nil
}

func (s *foundationStubServer) GetPartyContactById(_ context.Context, req *efv1.GetPartyContactByIdRequest) (*efv1.GetPartyContactByIdResponse, error) {
	contact := s.contacts[req.GetContactId()]
	if contact == nil {
		return nil, status.Error(codes.NotFound, "party contact not found")
	}
	return &efv1.GetPartyContactByIdResponse{Contact: contact}, nil
}

func hmacVerifierInterceptor(secret []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		if len(md.Get(authcontext.MetaGatewayProof)) != 1 || len(md.Get(authcontext.MetaRequestID)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "expected one proof and one request ID")
		}
		caller, ok, err := authcontext.ExtractFromIncomingGRPCContext(ctx, secret, time.Now())
		if err != nil || !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid internal caller context")
		}
		return handler(authcontext.WithCaller(ctx, caller), req)
	}
}

func TestEnterpriseFoundationClientVerification(t *testing.T) {
	t.Parallel()
	secret := []byte(strings.Repeat("k", 32))
	leID := "00000000-0000-4000-8000-000000000101"
	otherLE := "00000000-0000-4000-8000-000000000102"
	unitID := "00000000-0000-4000-8000-000000000201"
	locID := "00000000-0000-4000-8000-000000000301"
	personID := "00000000-0000-4000-8000-000000000401"
	orgPartyID := "00000000-0000-4000-8000-000000000402"
	inactivePartyID := "00000000-0000-4000-8000-000000000403"
	deletedPartyID := "00000000-0000-4000-8000-000000000404"
	missingPartyID := "00000000-0000-4000-8000-000000000405"
	inactiveUnitID := "00000000-0000-4000-8000-000000000202"
	mismatchedID := "00000000-0000-4000-8000-000000000999"
	malformedDateID := "00000000-0000-4000-8000-000000000103"
	malformedEntityUnitID := "00000000-0000-4000-8000-000000000203"

	stub := &foundationStubServer{
		legalEntities: map[string]*efv1.LegalEntity{
			leID: {
				Id: leID, Status: efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE,
				IncorporationDate: timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			otherLE: {Id: otherLE, Status: efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_SUSPENDED},
			malformedDateID: {
				Id: malformedDateID, Status: efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE,
				IncorporationDate: &timestamppb.Timestamp{Nanos: -1},
			},
		},
		orgUnits: map[string]*efv1.OrganizationUnit{
			unitID:                {Id: unitID, LegalEntityId: leID, IsActive: true},
			inactiveUnitID:        {Id: inactiveUnitID, LegalEntityId: leID, IsActive: false},
			malformedEntityUnitID: {Id: malformedEntityUnitID, LegalEntityId: malformedDateID, IsActive: true},
		},
		locations: map[string]*efv1.Location{
			locID: {Id: locID, LegalEntityId: leID, IsActive: true},
		},
		parties: map[string]*efv1.Party{
			personID:        {Id: personID, Kind: efv1.PartyKind_PARTY_KIND_INDIVIDUAL, Status: efv1.PartyStatus_PARTY_STATUS_ACTIVE},
			orgPartyID:      {Id: orgPartyID, Kind: efv1.PartyKind_PARTY_KIND_ORGANIZATION, Status: efv1.PartyStatus_PARTY_STATUS_ACTIVE},
			inactivePartyID: {Id: inactivePartyID, Kind: efv1.PartyKind_PARTY_KIND_INDIVIDUAL, Status: efv1.PartyStatus_PARTY_STATUS_INACTIVE},
			deletedPartyID: {Id: deletedPartyID, Kind: efv1.PartyKind_PARTY_KIND_INDIVIDUAL, Status: efv1.PartyStatus_PARTY_STATUS_ACTIVE,
				DeletedAt: timestamppb.New(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))},
		},
	}
	stub.legalEntities[mismatchedID] = stub.legalEntities[leID]
	stub.orgUnits[mismatchedID] = stub.orgUnits[unitID]
	stub.locations[mismatchedID] = stub.locations[locID]
	stub.parties[mismatchedID] = stub.parties[personID]

	contactID := "00000000-0000-4000-8000-000000000501"
	inactiveContactID := "00000000-0000-4000-8000-000000000502"
	stub.contacts = map[string]*efv1.PartyContact{
		contactID:         {Id: contactID, PartyId: personID, IsActive: true},
		inactiveContactID: {Id: inactiveContactID, PartyId: personID},
		mismatchedID:      {Id: contactID, PartyId: personID, IsActive: true},
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	srv := grpc.NewServer(grpc.UnaryInterceptor(hmacVerifierInterceptor(secret)))
	efv1.RegisterLegalEntityServiceServer(srv, stub)
	efv1.RegisterOrganizationUnitServiceServer(srv, stub)
	efv1.RegisterLocationServiceServer(srv, stub)
	efv1.RegisterPartyServiceServer(srv, stub)
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(srv.Stop)

	conn, client, err := efclient.Dial(listener.Addr().String(), secret, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if err := client.VerifyParty(t.Context(), personID, true); err != nil {
		t.Fatalf("active individual party rejected: %v", err)
	}
	for _, id := range []string{orgPartyID, inactivePartyID, deletedPartyID, missingPartyID} {
		if err := client.VerifyParty(t.Context(), id, true); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Errorf("invalid individual party %s: %v", id, err)
		}
	}

	if err := client.VerifyPartyContact(t.Context(), contactID, personID); err != nil {
		t.Fatalf("active contact rejected: %v", err)
	}
	for _, refs := range [][2]string{{contactID, orgPartyID}, {inactiveContactID, personID}, {mismatchedID, personID}} {
		if err := client.VerifyPartyContact(t.Context(), refs[0], refs[1]); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Errorf("invalid party contact %v: %v", refs, err)
		}
	}

	validFrom := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		verify func() error
	}{
		{"legal entity", func() error { return client.VerifyLegalEntity(t.Context(), mismatchedID) }},
		{"organization unit", func() error { return client.VerifyOrganizationUnit(t.Context(), mismatchedID, leID) }},
		{"location", func() error { return client.VerifyLocation(t.Context(), mismatchedID, leID) }},
		{"party", func() error { return client.VerifyParty(t.Context(), mismatchedID, true) }},
	} {
		if err := tc.verify(); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Errorf("%s with mismatched response ID = %v", tc.name, err)
		}
	}

	if err := client.VerifyPositionReferences(t.Context(), security.PositionReferenceCheck{
		OrganizationUnitID: unitID, LegalEntityID: leID, WorkLocationID: &locID, ValidFrom: validFrom,
	}); err != nil {
		t.Fatalf("valid position references rejected: %v", err)
	}
	previousDay := validFrom.AddDate(0, 0, -1)
	if err := client.VerifyPositionReferences(t.Context(), security.PositionReferenceCheck{
		OrganizationUnitID: unitID, LegalEntityID: leID, ValidFrom: validFrom, ValidTo: &previousDay,
	}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inverted position period accepted: %v", err)
	}
	if err := client.VerifyPositionReferences(t.Context(), security.PositionReferenceCheck{
		OrganizationUnitID: malformedEntityUnitID, LegalEntityID: malformedDateID, ValidFrom: validFrom,
	}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("malformed incorporation timestamp accepted: %v", err)
	}
	if err := client.VerifyWorkerReferences(t.Context(), security.WorkerReferenceCheck{
		PersonPartyID: personID, LegalEntityID: leID,
	}); err != nil {
		t.Fatalf("valid worker references rejected: %v", err)
	}
	beforeIncorporation := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	if err := client.VerifyLegalEntityWindow(t.Context(), leID, beforeIncorporation, validFrom); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("payroll window before incorporation accepted: %v", err)
	}
	if err := client.VerifyLegalEntityWindow(t.Context(), leID, validFrom, validFrom.AddDate(0, 0, 29)); err != nil {
		t.Fatalf("legal entity payroll window rejected: %v", err)
	}
	for _, check := range []security.WorkerReferenceCheck{
		{PersonPartyID: personID, LegalEntityID: leID, HireDate: &beforeIncorporation},
		{PersonPartyID: personID, LegalEntityID: leID, ServiceStartDate: &beforeIncorporation},
	} {
		if err := client.VerifyWorkerReferences(t.Context(), check); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Fatalf("worker date before incorporation accepted: %v", err)
		}
	}
	if err := client.VerifyWorkerReferences(t.Context(), security.WorkerReferenceCheck{
		PersonPartyID: personID, LegalEntityID: leID, HireDate: &validFrom, ServiceStartDate: &validFrom,
	}); err != nil {
		t.Fatalf("worker dates within entity lifetime rejected: %v", err)
	}
	if err := client.VerifyPositionReferences(t.Context(), security.PositionReferenceCheck{
		OrganizationUnitID: unitID, LegalEntityID: leID, ValidFrom: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("position before incorporation accepted: %v", err)
	}
	if err := client.VerifyPositionReferences(t.Context(), security.PositionReferenceCheck{
		OrganizationUnitID: inactiveUnitID, LegalEntityID: leID, ValidFrom: validFrom,
	}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive organization unit accepted: %v", err)
	}
	if err := client.VerifyPositionReferences(t.Context(), security.PositionReferenceCheck{
		OrganizationUnitID: unitID, LegalEntityID: otherLE, ValidFrom: validFrom,
	}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("suspended legal entity accepted: %v", err)
	}
	if err := client.VerifyWorkerReferences(t.Context(), security.WorkerReferenceCheck{
		PersonPartyID: orgPartyID, LegalEntityID: leID,
	}); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("organization party accepted as worker person: %v", err)
	}

	disabled := efclient.NewDisabledClient()
	if err := disabled.VerifyWorkerReferences(t.Context(), security.WorkerReferenceCheck{
		PersonPartyID: personID, LegalEntityID: leID,
	}); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("disabled client must fail closed: %v", err)
	}

	badHMAC := efclient.NewClient(conn, []byte(strings.Repeat("x", 32)), 2*time.Second)
	if err := badHMAC.VerifyLegalEntity(t.Context(), leID); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("bad HMAC must map to ErrUnavailable: %v", err)
	}
}
