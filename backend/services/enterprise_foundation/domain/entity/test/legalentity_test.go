package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestLegalEntity_LifecycleAndInvariants(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.Local)
	incDate := time.Date(2024, 5, 10, 23, 0, 0, 0, time.FixedZone("WIB", 7*3600))

	le, err := entity.NewLegalEntity(testUUID1, entity.LegalEntityParams{
		EntityCode:               " PT-IO-01 ",
		LegalName:                " PT IntegriOne Retail Nusantara ",
		TradingName:              strPtr(" IntegriOne "),
		LegalForm:                strPtr("PT"),
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		JurisdictionCountryCode:  "",
		TaxIDNumber:              strPtr(" 01.234.567.8-901.000 "),
		IsPKP:                    true,
		RegisteredAddressID:      strPtr(testUUID2),
		FiscalYearStartMonth:     1,
		IncorporationDate:        &incDate,
	}, now)
	if err != nil {
		t.Fatalf("NewLegalEntity() error = %v", err)
	}
	if le.EntityCode != "pt-io-01" || le.JurisdictionCountryCode != "ID" || !le.IsEligibleScope() {
		t.Fatalf("unexpected LegalEntity fields: %+v", le)
	}
	if err := le.CanBeUsedInTransactions(); err != nil {
		t.Fatalf("CanBeUsedInTransactions() error = %v", err)
	}

	// Entitas utama operasional tidak boleh langsung disuspend atau dihapus sebelum dilepas status utamanya.
	if err := le.Suspend(now.Add(time.Minute)); !errors.Is(err, eferrors.ErrPrimaryLegalEntityExists) {
		t.Fatalf("Suspend() primary error = %v, want ErrPrimaryLegalEntityExists", err)
	}
	if err := le.SoftDelete(now.Add(time.Minute)); !errors.Is(err, eferrors.ErrPrimaryLegalEntityExists) {
		t.Fatalf("SoftDelete() primary error = %v, want ErrPrimaryLegalEntityExists", err)
	}

	if err := le.SetPrimaryOperatingEntity(false, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("SetPrimaryOperatingEntity(false) error = %v", err)
	}
	if err := le.Suspend(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("Suspend() error = %v", err)
	}
	if le.IsEligibleScope() {
		t.Fatal("suspended legal entity must not be eligible scope")
	}
	if err := le.CanBeUsedInTransactions(); !errors.Is(err, eferrors.ErrLegalEntityInactive) {
		t.Fatalf("CanBeUsedInTransactions() suspended = %v, want ErrLegalEntityInactive", err)
	}

	badDissolution := time.Date(2024, 5, 9, 0, 0, 0, 0, time.UTC)
	if err := le.Dissolve(&badDissolution, now.Add(4*time.Minute)); !errors.Is(err, eferrors.ErrInvalidEntityDates) {
		t.Fatalf("Dissolve() early date error = %v, want ErrInvalidEntityDates", err)
	}

	goodDissolution := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	if err := le.Dissolve(&goodDissolution, now.Add(5*time.Minute)); err != nil {
		t.Fatalf("Dissolve() error = %v", err)
	}
	if err := le.Activate(now.Add(6 * time.Minute)); !errors.Is(err, eferrors.ErrConflict) {
		t.Fatalf("Activate() after dissolved error = %v, want ErrConflict", err)
	}

	if err := le.SoftDelete(now.Add(7 * time.Minute)); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	if err := le.CanBeUsedInTransactions(); !errors.Is(err, eferrors.ErrLegalEntityDeleted) {
		t.Fatalf("CanBeUsedInTransactions() deleted error = %v, want ErrLegalEntityDeleted", err)
	}
}

func TestNewLegalEntity_ValidationErrors(t *testing.T) {
	t.Parallel()

	now := time.Now()
	inc := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	diss := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		params  entity.LegalEntityParams
		wantErr error
	}{
		{
			name: "self parent",
			params: entity.LegalEntityParams{
				ParentLegalEntityID:  strPtr(testUUID1),
				EntityCode:           "pt-io",
				LegalName:            "PT IO",
				Status:               entity.LegalEntityStatusActive,
				FiscalYearStartMonth: 1,
			},
			wantErr: eferrors.ErrLegalEntitySelfParent,
		},
		{
			name: "invalid fiscal month",
			params: entity.LegalEntityParams{
				EntityCode:           "pt-io",
				LegalName:            "PT IO",
				Status:               entity.LegalEntityStatusActive,
				FiscalYearStartMonth: 13,
			},
			wantErr: eferrors.ErrInvalidFiscalMonth,
		},
		{
			name: "dissolution before incorporation",
			params: entity.LegalEntityParams{
				EntityCode:           "pt-io",
				LegalName:            "PT IO",
				Status:               entity.LegalEntityStatusActive,
				FiscalYearStartMonth: 1,
				IncorporationDate:    &inc,
				DissolutionDate:      &diss,
			},
			wantErr: eferrors.ErrInvalidEntityDates,
		},
		{
			name: "primary non active",
			params: entity.LegalEntityParams{
				EntityCode:               "pt-io",
				LegalName:                "PT IO",
				IsPrimaryOperatingEntity: true,
				Status:                   entity.LegalEntityStatusPending,
				FiscalYearStartMonth:     1,
			},
			wantErr: eferrors.ErrConflict,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := entity.NewLegalEntity(testUUID1, tc.params, now); !errors.Is(err, tc.wantErr) {
				t.Fatalf("NewLegalEntity() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
