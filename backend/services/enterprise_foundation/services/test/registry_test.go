package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/address"
)

type stubTransactor struct{}

func (stubTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type stubAddressRepo struct {
	inserted *entity.Address
}

func (r *stubAddressRepo) SelectByID(context.Context, string) (*entity.Address, error) {
	return nil, eferrors.ErrAddressNotFound
}
func (r *stubAddressRepo) SelectByIDs(context.Context, []string) ([]*entity.Address, error) {
	return nil, nil
}
func (r *stubAddressRepo) SelectList(context.Context, domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	if r.inserted != nil {
		return domainrepo.PageResult[*entity.Address]{Items: []*entity.Address{r.inserted.Clone()}, Total: 1}, nil
	}
	return domainrepo.PageResult[*entity.Address]{}, nil
}
func (r *stubAddressRepo) CheckReferenced(context.Context, string) (bool, error) { return false, nil }
func (r *stubAddressRepo) Insert(_ context.Context, addr *entity.Address) error {
	r.inserted = addr.Clone()
	return nil
}
func (r *stubAddressRepo) Update(context.Context, *entity.Address) error { return nil }
func (r *stubAddressRepo) DeleteByID(context.Context, string) error      { return nil }

type stubLegalEntityRepo struct{}

func (stubLegalEntityRepo) SelectByID(context.Context, string) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func (stubLegalEntityRepo) SelectByIDs(context.Context, []string) ([]*entity.LegalEntity, error) {
	return nil, nil
}
func (stubLegalEntityRepo) SelectByCode(context.Context, string) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func (stubLegalEntityRepo) SelectByTaxIDNumber(context.Context, string) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func (stubLegalEntityRepo) SelectByNIB(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	return domainrepo.PageResult[*entity.LegalEntity]{}, nil
}
func (stubLegalEntityRepo) SelectPrimaryOperating(context.Context) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func (stubLegalEntityRepo) SelectList(context.Context, domainrepo.LegalEntityFilter) (domainrepo.PageResult[*entity.LegalEntity], error) {
	return domainrepo.PageResult[*entity.LegalEntity]{}, nil
}
func (stubLegalEntityRepo) SelectParentID(context.Context, string) (*string, error) { return nil, nil }
func (stubLegalEntityRepo) SelectChildrenByParentID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	return domainrepo.PageResult[*entity.LegalEntity]{}, nil
}
func (stubLegalEntityRepo) CountActiveDependents(context.Context, string) (int64, error) {
	return 0, nil
}
func (stubLegalEntityRepo) CheckCodeExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (stubLegalEntityRepo) CheckTaxIDExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (stubLegalEntityRepo) Insert(context.Context, *entity.LegalEntity) error { return nil }
func (stubLegalEntityRepo) Update(context.Context, *entity.LegalEntity) error { return nil }
func (stubLegalEntityRepo) UpdateClearPrimaryOperatingExcept(context.Context, string, time.Time) error {
	return nil
}
func (stubLegalEntityRepo) SoftDeleteByID(context.Context, string, time.Time) error { return nil }

type stubLocationRepo struct{}

func (stubLocationRepo) SelectByID(context.Context, string) (*entity.Location, error) {
	return nil, eferrors.ErrLocationNotFound
}
func (stubLocationRepo) SelectByIDs(context.Context, []string) ([]*entity.Location, error) {
	return nil, nil
}
func (stubLocationRepo) SelectByCode(context.Context, string) (*entity.Location, error) {
	return nil, eferrors.ErrLocationNotFound
}
func (stubLocationRepo) SelectByLegalEntityAndKind(context.Context, string, entity.LocationKind, bool, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (stubLocationRepo) SelectByAddressID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (stubLocationRepo) SelectList(context.Context, domainrepo.LocationFilter) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (stubLocationRepo) SelectParentID(context.Context, string) (*string, error) { return nil, nil }
func (stubLocationRepo) SelectChildrenByParentID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (stubLocationRepo) CountActiveDependents(context.Context, string) (int64, error) { return 0, nil }
func (stubLocationRepo) CheckCodeExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (stubLocationRepo) Insert(context.Context, *entity.Location) error          { return nil }
func (stubLocationRepo) Update(context.Context, *entity.Location) error          { return nil }
func (stubLocationRepo) SoftDeleteByID(context.Context, string, time.Time) error { return nil }

type stubOrgUnitRepo struct{}

func (stubOrgUnitRepo) SelectByID(context.Context, string) (*entity.OrganizationUnit, error) {
	return nil, eferrors.ErrOrgUnitNotFound
}
func (stubOrgUnitRepo) SelectByIDs(context.Context, []string) ([]*entity.OrganizationUnit, error) {
	return nil, nil
}
func (stubOrgUnitRepo) SelectByEntityAndCode(context.Context, string, string) (*entity.OrganizationUnit, error) {
	return nil, eferrors.ErrOrgUnitNotFound
}
func (stubOrgUnitRepo) SelectByHeadPersonID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	return domainrepo.PageResult[*entity.OrganizationUnit]{}, nil
}
func (stubOrgUnitRepo) SelectByPrimaryLocationID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	return domainrepo.PageResult[*entity.OrganizationUnit]{}, nil
}
func (stubOrgUnitRepo) SelectList(context.Context, domainrepo.OrganizationUnitFilter) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	return domainrepo.PageResult[*entity.OrganizationUnit]{}, nil
}
func (stubOrgUnitRepo) SelectParentID(context.Context, string) (*string, error) { return nil, nil }
func (stubOrgUnitRepo) SelectChildrenByParentID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	return domainrepo.PageResult[*entity.OrganizationUnit]{}, nil
}
func (stubOrgUnitRepo) CountActiveDependents(context.Context, string) (int64, error) { return 0, nil }
func (stubOrgUnitRepo) CheckCodeExistsInEntity(context.Context, string, string, string) (bool, error) {
	return false, nil
}
func (stubOrgUnitRepo) Insert(context.Context, *entity.OrganizationUnit) error  { return nil }
func (stubOrgUnitRepo) Update(context.Context, *entity.OrganizationUnit) error  { return nil }
func (stubOrgUnitRepo) SoftDeleteByID(context.Context, string, time.Time) error { return nil }

type stubPartyRepo struct{}

func (stubPartyRepo) SelectByID(context.Context, string) (*entity.Party, error) {
	return nil, eferrors.ErrPartyNotFound
}
func (stubPartyRepo) SelectByIDs(context.Context, []string) ([]*entity.Party, error) { return nil, nil }
func (stubPartyRepo) SelectByCode(context.Context, string) (*entity.Party, error) {
	return nil, eferrors.ErrPartyNotFound
}
func (stubPartyRepo) SelectByTaxIDNumber(context.Context, string) (*entity.Party, error) {
	return nil, eferrors.ErrPartyNotFound
}
func (stubPartyRepo) SelectByEmail(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Party], error) {
	return domainrepo.PageResult[*entity.Party]{}, nil
}
func (stubPartyRepo) SelectList(context.Context, domainrepo.PartyFilter) (domainrepo.PageResult[*entity.Party], error) {
	return domainrepo.PageResult[*entity.Party]{}, nil
}
func (stubPartyRepo) CheckCodeExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (stubPartyRepo) CheckTaxIDExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (stubPartyRepo) Insert(context.Context, *entity.Party) error             { return nil }
func (stubPartyRepo) Update(context.Context, *entity.Party) error             { return nil }
func (stubPartyRepo) SoftDeleteByID(context.Context, string, time.Time) error { return nil }

type stubPartyRoleRepo struct{}

func (stubPartyRoleRepo) SelectByID(context.Context, string) (*entity.PartyRole, error) {
	return nil, eferrors.ErrPartyRoleNotFound
}
func (stubPartyRoleRepo) SelectList(context.Context, domainrepo.PartyRoleFilter) (domainrepo.PageResult[*entity.PartyRole], error) {
	return domainrepo.PageResult[*entity.PartyRole]{}, nil
}
func (stubPartyRoleRepo) SelectByPartyID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	return domainrepo.PageResult[*entity.PartyRole]{}, nil
}
func (stubPartyRoleRepo) SelectByPartyAndRole(context.Context, string, entity.PartyRoleKind) (*entity.PartyRole, error) {
	return nil, eferrors.ErrPartyRoleNotFound
}
func (stubPartyRoleRepo) SelectByRoleKind(context.Context, entity.PartyRoleKind, bool, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	return domainrepo.PageResult[*entity.PartyRole]{}, nil
}
func (stubPartyRoleRepo) CheckExists(context.Context, string, entity.PartyRoleKind) (bool, error) {
	return false, nil
}
func (stubPartyRoleRepo) Insert(context.Context, *entity.PartyRole) error { return nil }
func (stubPartyRoleRepo) Update(context.Context, *entity.PartyRole) error { return nil }
func (stubPartyRoleRepo) DeleteByID(context.Context, string) error        { return nil }

type stubPartyContactRepo struct{}

func (stubPartyContactRepo) SelectByID(context.Context, string) (*entity.PartyContact, error) {
	return nil, eferrors.ErrPartyContactNotFound
}
func (stubPartyContactRepo) SelectByPartyID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	return domainrepo.PageResult[*entity.PartyContact]{}, nil
}
func (stubPartyContactRepo) SelectPrimaryByPartyID(context.Context, string) (*entity.PartyContact, error) {
	return nil, eferrors.ErrPartyContactNotFound
}
func (stubPartyContactRepo) SelectByIAMUserID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	return domainrepo.PageResult[*entity.PartyContact]{}, nil
}
func (stubPartyContactRepo) SelectByPartyAndIAMUserID(context.Context, string, string) (*entity.PartyContact, error) {
	return nil, eferrors.ErrPartyContactNotFound
}
func (stubPartyContactRepo) SelectByEmail(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	return domainrepo.PageResult[*entity.PartyContact]{}, nil
}
func (stubPartyContactRepo) SelectList(context.Context, domainrepo.PartyContactFilter) (domainrepo.PageResult[*entity.PartyContact], error) {
	return domainrepo.PageResult[*entity.PartyContact]{}, nil
}
func (stubPartyContactRepo) CheckIAMUserExistsInParty(context.Context, string, string, string) (bool, error) {
	return false, nil
}
func (stubPartyContactRepo) UpdateClearPrimaryByPartyExcept(context.Context, string, string, time.Time) error {
	return nil
}
func (stubPartyContactRepo) Insert(context.Context, *entity.PartyContact) error { return nil }
func (stubPartyContactRepo) Update(context.Context, *entity.PartyContact) error { return nil }
func (stubPartyContactRepo) SoftDeleteByID(context.Context, string, time.Time) error {
	return nil
}

type stubPartyAddressRepo struct{}

func (stubPartyAddressRepo) SelectByID(context.Context, string) (*entity.PartyAddress, error) {
	return nil, eferrors.ErrPartyAddressNotFound
}
func (stubPartyAddressRepo) SelectList(context.Context, domainrepo.PartyAddressFilter) (domainrepo.PageResult[*entity.PartyAddress], error) {
	return domainrepo.PageResult[*entity.PartyAddress]{}, nil
}
func (stubPartyAddressRepo) SelectByPartyID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	return domainrepo.PageResult[*entity.PartyAddress]{}, nil
}
func (stubPartyAddressRepo) SelectByPartyAndUsage(context.Context, string, entity.AddressUsageKind, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	return domainrepo.PageResult[*entity.PartyAddress]{}, nil
}
func (stubPartyAddressRepo) SelectPrimaryByPartyAndUsage(context.Context, string, entity.AddressUsageKind) (*entity.PartyAddress, error) {
	return nil, eferrors.ErrPartyAddressNotFound
}
func (stubPartyAddressRepo) SelectByAddressID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	return domainrepo.PageResult[*entity.PartyAddress]{}, nil
}
func (stubPartyAddressRepo) SelectByPartyAddressUsage(context.Context, string, string, entity.AddressUsageKind) (*entity.PartyAddress, error) {
	return nil, eferrors.ErrPartyAddressNotFound
}
func (stubPartyAddressRepo) UpdateClearPrimaryByPartyUsageExcept(context.Context, string, entity.AddressUsageKind, string, time.Time) error {
	return nil
}
func (stubPartyAddressRepo) Insert(context.Context, *entity.PartyAddress) error { return nil }
func (stubPartyAddressRepo) Update(context.Context, *entity.PartyAddress) error { return nil }
func (stubPartyAddressRepo) DeleteByID(context.Context, string) error           { return nil }

type stubOutboxRepo struct{}

func (stubOutboxRepo) Insert(context.Context, *entity.OutboxEvent) error { return nil }
func (stubOutboxRepo) SelectByID(context.Context, string) (*entity.OutboxEvent, error) {
	return nil, eferrors.ErrOutboxEventNotFound
}
func (stubOutboxRepo) SelectByAggregate(context.Context, string, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return domainrepo.PageResult[*entity.OutboxEvent]{}, nil
}
func (stubOutboxRepo) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (stubOutboxRepo) UpdateEffective(context.Context, string, time.Time) error { return nil }
func (stubOutboxRepo) UpdateRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (stubOutboxRepo) UpdateFailed(context.Context, string, string) error { return nil }
func (stubOutboxRepo) CountPending(context.Context) (int64, error)        { return 0, nil }
func (stubOutboxRepo) CountFailed(context.Context) (int64, error)         { return 0, nil }
func (stubOutboxRepo) DeleteEffectiveBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func fullStubRepos() *repository.Registry {
	return &repository.Registry{
		Transactor:        stubTransactor{},
		Addresses:         &stubAddressRepo{},
		LegalEntities:     stubLegalEntityRepo{},
		Locations:         stubLocationRepo{},
		OrganizationUnits: stubOrgUnitRepo{},
		Parties:           stubPartyRepo{},
		PartyRoles:        stubPartyRoleRepo{},
		PartyContacts:     stubPartyContactRepo{},
		PartyAddresses:    stubPartyAddressRepo{},
		Outbox:            stubOutboxRepo{},
	}
}

func TestNewRegistryFailClosedAndDefaults(t *testing.T) {
	if _, err := services.NewRegistry(nil, services.Options{}); err == nil || eferrors.KindOf(err) != eferrors.KindInvalidInput {
		t.Fatalf("expected KindInvalidInput for nil repos, got %v", err)
	}

	var nilReg *services.Registry
	if err := nilReg.Validate(); eferrors.KindOf(err) != eferrors.KindUnavailable {
		t.Fatalf("expected KindUnavailable on nil registry Validate, got %v", err)
	}
	if err := (&services.Registry{}).Validate(); eferrors.KindOf(err) != eferrors.KindUnavailable {
		t.Fatalf("expected KindUnavailable on incomplete registry Validate, got %v", err)
	}

	missingCases := []struct {
		name  string
		mut   func(*repository.Registry)
		field string
	}{
		{"missing addresses", func(r *repository.Registry) { r.Addresses = nil }, "addresses"},
		{"missing legal_entities", func(r *repository.Registry) { r.LegalEntities = nil }, "legal_entities"},
		{"missing locations", func(r *repository.Registry) { r.Locations = nil }, "locations"},
		{"missing organization_units", func(r *repository.Registry) { r.OrganizationUnits = nil }, "organization_units"},
		{"missing parties", func(r *repository.Registry) { r.Parties = nil }, "parties"},
		{"missing outbox", func(r *repository.Registry) { r.Outbox = nil }, "outbox"},
	}
	for _, tc := range missingCases {
		t.Run(tc.name, func(t *testing.T) {
			repos := fullStubRepos()
			tc.mut(repos)
			_, err := services.NewRegistry(repos, services.Options{})
			if err == nil || eferrors.KindOf(err) != eferrors.KindInvalidInput || eferrors.FieldOf(err) != tc.field {
				t.Fatalf("expected invalid input for %s, got %v", tc.field, err)
			}
		})
	}

	reg, err := services.NewRegistry(fullStubRepos(), services.Options{})
	if err != nil {
		t.Fatalf("NewRegistry full stubs failed: %v", err)
	}
	if err := reg.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Default uuid.NewGenerator dan SystemClock langsung aktif saat Create melalui registry.
	addr, err := reg.Addresses.Create(context.Background(), address.CreateInput{
		AddressLine1: "Jl. Gatot Subroto No. 50",
		City:         "Jakarta Selatan",
		CountryCode:  "ID",
	})
	if err != nil {
		t.Fatalf("Addresses.Create via default registry failed: %v", err)
	}
	if err := entity.ValidateUUID("id", addr.ID); err != nil {
		t.Fatalf("expected valid default UUIDv4, got %q (%v)", addr.ID, err)
	}
}

func (r stubOrgUnitRepo) SelectByIDForUpdate(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return r.SelectByID(ctx, id)
}
