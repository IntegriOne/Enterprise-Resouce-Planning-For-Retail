package handler_test

import (
	"context"
	"slices"
	"testing"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

func TestRegisterAll_FailClosedGuards(t *testing.T) {
	t.Parallel()

	if err := handler.RegisterAll(nil, handler.Dependencies{Services: &services.Registry{}}); err == nil {
		t.Fatal("expected nil registrar to fail")
	}
	if err := handler.RegisterAll(grpc.NewServer(), handler.Dependencies{}); err == nil {
		t.Fatal("expected nil services registry to fail")
	}
	if err := handler.RegisterAll(grpc.NewServer(), handler.Dependencies{Services: &services.Registry{}}); err == nil {
		t.Fatal("expected incomplete services registry to fail validation")
	}

	if _, err := handler.NewAddressHandler(nil, nil); err == nil {
		t.Fatal("expected nil address service error")
	}
	if _, err := handler.NewLegalEntityHandler(nil, nil); err == nil {
		t.Fatal("expected nil legal entity service error")
	}
	if _, err := handler.NewLocationHandler(nil, nil); err == nil {
		t.Fatal("expected nil location service error")
	}
	if _, err := handler.NewOrganizationUnitHandler(nil, nil); err == nil {
		t.Fatal("expected nil organization unit service error")
	}
	if _, err := handler.NewPartyHandler(nil, nil); err == nil {
		t.Fatal("expected nil party service error")
	}
}

func TestServerRegistration_HealthAndReflection(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)

	healthClient := healthpb.NewHealthClient(h.conn)
	resp, err := healthClient.Check(context.Background(), &healthpb.HealthCheckRequest{
		Service: "integrione-enterprise-foundation",
	})
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("unexpected health status: %v", resp.GetStatus())
	}

	// Pastikan seluruh 5 kontrak service domain terdaftar pada gRPC Server Reflection.
	reflClient := reflectionpb.NewServerReflectionClient(h.conn)
	stream, err := reflClient.ServerReflectionInfo(context.Background())
	if err != nil {
		t.Fatalf("open reflection stream: %v", err)
	}
	if err := stream.Send(&reflectionpb.ServerReflectionRequest{
		MessageRequest: &reflectionpb.ServerReflectionRequest_ListServices{
			ListServices: "",
		},
	}); err != nil {
		t.Fatalf("send list_services: %v", err)
	}
	reflResp, err := stream.Recv()
	if err != nil {
		t.Fatalf("recv list_services: %v", err)
	}
	_ = stream.CloseSend()

	listed := make([]string, 0)
	for _, svc := range reflResp.GetListServicesResponse().GetService() {
		listed = append(listed, svc.GetName())
	}

	required := []string{
		"enterprise_foundation.v1.AddressService",
		"enterprise_foundation.v1.LegalEntityService",
		"enterprise_foundation.v1.LocationService",
		"enterprise_foundation.v1.OrganizationUnitService",
		"enterprise_foundation.v1.PartyService",
	}
	for _, name := range required {
		if !slices.Contains(listed, name) {
			t.Fatalf("missing reflected service %q in %v", name, listed)
		}
	}

	// Pastikan pemanggilan domain tanpa konteks HMAC internal Gateway ditolak fail-closed.
	if _, err := h.addresses.GetAllAddresses(context.Background(), nil); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated without HMAC metadata, got %v (%v)", status.Code(err), err)
	}
}

func TestHandlers_NilRequestGuards(t *testing.T) {
	t.Parallel()

	addrRepo := newMemAddressRepo()
	leRepo := newMemLegalEntityRepo()
	locRepo := newMemLocationRepo()
	ouRepo := newMemOrgUnitRepo()
	partyRepo := newMemPartyRepo()
	roleRepo := newMemPartyRoleRepo()
	contactRepo := newMemPartyContactRepo()
	paRepo := newMemPartyAddressRepo()
	svcReg, err := services.NewRegistry(&repository.Registry{
		Transactor:        noopTransactor{},
		Addresses:         addrRepo,
		LegalEntities:     leRepo,
		Locations:         locRepo,
		OrganizationUnits: ouRepo,
		Parties:           partyRepo,
		PartyRoles:        roleRepo,
		PartyContacts:     contactRepo,
		PartyAddresses:    paRepo,
		Outbox:            &memOutboxRepo{},
	}, services.Options{IDGen: &seqUUIDGen{}})
	if err != nil {
		t.Fatalf("services.NewRegistry failed: %v", err)
	}

	errMapper := handler.NewErrorMapper(nil)
	addrH, _ := handler.NewAddressHandler(svcReg.Addresses, errMapper)
	leH, _ := handler.NewLegalEntityHandler(svcReg.LegalEntities, errMapper)
	locH, _ := handler.NewLocationHandler(svcReg.Locations, errMapper)
	ouH, _ := handler.NewOrganizationUnitHandler(svcReg.OrganizationUnits, errMapper)
	partyH, _ := handler.NewPartyHandler(svcReg.Parties, errMapper)

	ctx := context.Background()
	checks := []struct {
		name string
		err  error
	}{
		{"Address.GetAddressById", func() error { _, err := addrH.GetAddressById(ctx, nil); return err }()},
		{"Address.CreateAddress", func() error { _, err := addrH.CreateAddress(ctx, nil); return err }()},
		{"LegalEntity.GetAll", func() error { _, err := leH.GetAllLegalEntities(ctx, nil); return err }()},
		{"LegalEntity.Create", func() error { _, err := leH.CreateLegalEntity(ctx, nil); return err }()},
		{"Location.GetAll", func() error { _, err := locH.GetAllLocations(ctx, nil); return err }()},
		{"Location.ValidateFacility", func() error { _, err := locH.ValidateFacilityScope(ctx, nil); return err }()},
		{"OrgUnit.GetAll", func() error { _, err := ouH.GetAllOrganizationUnits(ctx, nil); return err }()},
		{"OrgUnit.AssignHead", func() error { _, err := ouH.AssignOrganizationUnitHeadPerson(ctx, nil); return err }()},
		{"Party.Create", func() error { _, err := partyH.CreateParty(ctx, nil); return err }()},
		{"PartyRole.Assign", func() error { _, err := partyH.AssignPartyRole(ctx, nil); return err }()},
		{"PartyContact.Add", func() error { _, err := partyH.AddPartyContact(ctx, nil); return err }()},
		{"PartyAddress.Link", func() error { _, err := partyH.LinkPartyAddress(ctx, nil); return err }()},
	}
	for _, tc := range checks {
		if status.Code(tc.err) != codes.InvalidArgument {
			t.Fatalf("%s nil request: got %v, want InvalidArgument", tc.name, status.Code(tc.err))
		}
	}
}
