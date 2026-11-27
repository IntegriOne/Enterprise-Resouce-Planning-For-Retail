package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/address"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/legalentity"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/location"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/organizationunit"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/outbox"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/party"
)

func TestRegistryAndConstructorsFailClosed(t *testing.T) {
	t.Parallel()

	t.Run("New fails closed on nil db", func(t *testing.T) {
		t.Parallel()
		reg, err := repository.New(nil)
		if reg != nil || eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Fatalf("expected nil registry and KindUnavailable, got (%v, %v)", reg, err)
		}
	})

	t.Run("subpackage constructors fail closed on nil db", func(t *testing.T) {
		t.Parallel()
		if _, err := address.NewRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("address.NewRepository(nil) = %v", err)
		}
		if _, err := legalentity.NewRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("legalentity.NewRepository(nil) = %v", err)
		}
		if _, err := location.NewRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("location.NewRepository(nil) = %v", err)
		}
		if _, err := organizationunit.NewRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("organizationunit.NewRepository(nil) = %v", err)
		}
		if _, err := party.NewPartyRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("party.NewPartyRepository(nil) = %v", err)
		}
		if _, err := party.NewPartyRoleRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("party.NewPartyRoleRepository(nil) = %v", err)
		}
		if _, err := party.NewPartyContactRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("party.NewPartyContactRepository(nil) = %v", err)
		}
		if _, err := party.NewPartyAddressRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("party.NewPartyAddressRepository(nil) = %v", err)
		}
		if _, err := outbox.NewRepository(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("outbox.NewRepository(nil) = %v", err)
		}
	})
}

func TestNilReceiversAndInputValidationGuards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now()
	const validUUID = "11111111-2222-4333-8444-555555555555"

	t.Run("nil repository receivers fail closed with unavailable", func(t *testing.T) {
		t.Parallel()
		var addrRepo *address.Repository
		var leRepo *legalentity.Repository
		var locRepo *location.Repository
		var ouRepo *organizationunit.Repository
		var partyRepo *party.PartyRepository
		var roleRepo *party.PartyRoleRepository
		var contactRepo *party.PartyContactRepository
		var partyAddrRepo *party.PartyAddressRepository
		var outboxRepo *outbox.Repository

		for _, err := range []error{
			func() error { _, err := addrRepo.SelectByID(ctx, validUUID); return err }(),
			func() error { _, err := leRepo.SelectPrimaryOperating(ctx); return err }(),
			func() error { _, err := locRepo.SelectByCode(ctx, "wh-01"); return err }(),
			func() error {
				_, err := ouRepo.SelectByHeadPersonID(ctx, validUUID, domainrepo.PageRequest{})
				return err
			}(),
			func() error { _, err := partyRepo.SelectByTaxIDNumber(ctx, "012345"); return err }(),
			func() error { _, err := roleRepo.SelectByPartyID(ctx, validUUID, domainrepo.PageRequest{}); return err }(),
			func() error { _, err := roleRepo.SelectList(ctx, domainrepo.PartyRoleFilter{}); return err }(),
			func() error {
				_, err := contactRepo.SelectByIAMUserID(ctx, validUUID, domainrepo.PageRequest{})
				return err
			}(),
			func() error {
				_, err := partyAddrRepo.SelectByPartyID(ctx, validUUID, domainrepo.PageRequest{})
				return err
			}(),
			func() error { _, err := partyAddrRepo.SelectList(ctx, domainrepo.PartyAddressFilter{}); return err }(),
			func() error { _, err := outboxRepo.CountPending(ctx); return err }(),
		} {
			if eferrors.KindOf(err) != eferrors.KindUnavailable {
				t.Errorf("expected KindUnavailable on nil receiver, got %v", err)
			}
		}
	})

	t.Run("zero-value repos validate inputs fail-closed before touching nil db", func(t *testing.T) {
		t.Parallel()
		addrRepo := &address.Repository{}
		if _, err := addrRepo.SelectByID(ctx, "bad-uuid"); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid uuid error, got %v", err)
		}
		if res, err := addrRepo.SelectByIDs(ctx, nil); err != nil || len(res) != 0 {
			t.Errorf("expected empty SelectByIDs without db error, got (%v, %v)", res, err)
		}
		badCountry := "INDONESIA"
		if _, err := addrRepo.SelectList(ctx, domainrepo.AddressFilter{CountryCode: badCountry}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid country filter error, got %v", err)
		}

		leRepo := &legalentity.Repository{}
		if _, err := leRepo.SelectByTaxIDNumber(ctx, "  "); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid empty tax id error, got %v", err)
		}
		if ok, err := leRepo.CheckTaxIDExists(ctx, "   ", ""); err != nil || ok {
			t.Errorf("expected blank tax id CheckTaxIDExists=(false,nil), got (%v, %v)", ok, err)
		}
		badLEStatus := entity.LegalEntityStatus("bogus")
		if _, err := leRepo.SelectList(ctx, domainrepo.LegalEntityFilter{Status: &badLEStatus}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid status error, got %v", err)
		}
		if err := leRepo.Insert(ctx, &entity.LegalEntity{ID: validUUID, ParentLegalEntityID: strPtr(validUUID)}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected self-parent or validation error, got %v", err)
		}

		locRepo := &location.Repository{}
		if _, err := locRepo.SelectByLegalEntityAndKind(ctx, validUUID, entity.LocationKind("bogus"), true, domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid location kind error, got %v", err)
		}
		if err := locRepo.Insert(ctx, &entity.Location{
			ID:                    validUUID,
			LegalEntityID:         validUUID,
			LocationCode:          "wh-1",
			Name:                  "Gudang",
			Kind:                  entity.LocationKindWarehouse,
			Timezone:              "Asia/Jakarta",
			OperationalAttributes: json.RawMessage("{bad"),
			CreatedAt:             now,
			UpdatedAt:             now,
		}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid jsonb attributes error, got %v", err)
		}

		ouRepo := &organizationunit.Repository{}
		if _, err := ouRepo.SelectByEntityAndCode(ctx, validUUID, "bad space code"); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid org unit code error, got %v", err)
		}

		partyRepo := &party.PartyRepository{}
		if _, err := partyRepo.SelectByEmail(ctx, "not-an-email", domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid email error, got %v", err)
		}

		roleRepo := &party.PartyRoleRepository{}
		if _, err := roleRepo.SelectByRoleKind(ctx, entity.PartyRoleKind("unknown"), true, domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid role kind error, got %v", err)
		}
		badRoleKind := entity.PartyRoleKind("unknown")
		if _, err := roleRepo.SelectList(ctx, domainrepo.PartyRoleFilter{Role: &badRoleKind}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid role filter error, got %v", err)
		}

		contactRepo := &party.PartyContactRepository{}
		if err := contactRepo.Insert(ctx, &entity.PartyContact{
			ID:                 validUUID,
			PartyID:            validUUID,
			FullName:           "Budi",
			ContactRole:        "general",
			PortalAccessStatus: entity.PortalAccessStatusActive,
			CreatedAt:          now,
			UpdatedAt:          now,
		}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected portal link invariant error when iam_user_id nil, got %v", err)
		}

		partyAddrRepo := &party.PartyAddressRepository{}
		if _, err := partyAddrRepo.SelectByPartyAndUsage(ctx, validUUID, entity.AddressUsageKind("bogus"), domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid usage error, got %v", err)
		}
		badUsage := entity.AddressUsageKind("bogus")
		if _, err := partyAddrRepo.SelectList(ctx, domainrepo.PartyAddressFilter{Usage: &badUsage}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid usage filter error, got %v", err)
		}

		outboxRepo := &outbox.Repository{}
		if _, err := outboxRepo.DeleteEffectiveBefore(ctx, time.Time{}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid zero cutoff error, got %v", err)
		}
		if _, err := outboxRepo.SelectByAggregate(ctx, "", validUUID, domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Errorf("expected invalid aggregate type error, got %v", err)
		}
	})
}

func strPtr(s string) *string {
	return &s
}
