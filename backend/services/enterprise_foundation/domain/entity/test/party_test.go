package entity_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestNewParty(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	verifiedAt := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		id      string
		params  entity.PartyParams
		wantErr error
		check   func(t *testing.T, p *entity.Party)
	}{
		{
			name: "valid supplier and vendor partner with defaults",
			id:   strings.ToUpper(testUUID1),
			params: entity.PartyParams{
				PartnerCode: " pt-kain-01 ",
				Kind:        entity.PartyKindOrganization,
				DisplayName: " PT Kain Nusantara ",
				LegalName:   strPtr(" PT Kain Nusantara Tbk "),
				Email:       strPtr(" Finance@KainNusantara.CO.ID "),
				Phone:       strPtr(" +62 812-3456-7890 "),
				WebsiteURL:  strPtr(" https://kainnusantara.example.com "),
				TaxIDNumber: strPtr(" 01.234.567.8-901.000 "),
				IsPKP:       true,
				Status:      entity.PartyStatusPendingVerification,
				Notes:       strPtr(" pemasok kain & vendor makloon "),
			},
			check: func(t *testing.T, p *entity.Party) {
				t.Helper()
				if p.ID != testUUID1 {
					t.Errorf("ID = %q, want lowercase %q", p.ID, testUUID1)
				}
				if p.PartnerCode != "pt-kain-01" {
					t.Errorf("PartnerCode = %q, want %q", p.PartnerCode, "pt-kain-01")
				}
				if p.PreferredLanguage != entity.DefaultPreferredLanguage {
					t.Errorf("PreferredLanguage = %q, want %q", p.PreferredLanguage, entity.DefaultPreferredLanguage)
				}
				if p.Email == nil || *p.Email != "finance@kainnusantara.co.id" {
					t.Errorf("Email = %v, want normalized lowercase", p.Email)
				}
				if p.Status != entity.PartyStatusPendingVerification {
					t.Errorf("Status = %q, want %q", p.Status, entity.PartyStatusPendingVerification)
				}
			},
		},
		{
			name: "valid tax exempt party with reason",
			id:   testUUID2,
			params: entity.PartyParams{
				PartnerCode:        "umkm-jahit",
				Kind:               entity.PartyKindIndividual,
				DisplayName:        "Penjahit Makmur",
				IsTaxExempt:        true,
				TaxExemptionReason: strPtr("Peredaran bruto UMKM sesuai PP 55/2022"),
				TaxVerifiedAt:      &verifiedAt,
				PreferredLanguage:  "ID",
				Status:             entity.PartyStatusActive,
			},
			check: func(t *testing.T, p *entity.Party) {
				t.Helper()
				if !p.IsTaxExempt || p.TaxExemptionReason == nil {
					t.Fatalf("expected IsTaxExempt=true and TaxExemptionReason set")
				}
				if p.PreferredLanguage != "id" {
					t.Errorf("PreferredLanguage = %q, want %q", p.PreferredLanguage, "id")
				}
				if err := p.IsEligibleForTransactions(); err != nil {
					t.Errorf("IsEligibleForTransactions() unexpected error: %v", err)
				}
			},
		},
		{
			name: "invalid id",
			id:   "bad-uuid",
			params: entity.PartyParams{
				PartnerCode: "v-1",
				Kind:        entity.PartyKindOrganization,
				DisplayName: "Vendor A",
				Status:      entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrInvalidInput,
		},
		{
			name: "invalid partner code",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode: "bad code!",
				Kind:        entity.PartyKindOrganization,
				DisplayName: "Vendor A",
				Status:      entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrInvalidPartyCode,
		},
		{
			name: "invalid status",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode: "v-1",
				Kind:        entity.PartyKindOrganization,
				DisplayName: "Vendor A",
				Status:      entity.PartyStatus("deleted"),
			},
			wantErr: eferrors.ErrInvalidPartyStatus,
		},
		{
			name: "invalid kind",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode: "v-1",
				Kind:        entity.PartyKind("cooperative"),
				DisplayName: "Vendor A",
				Status:      entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrInvalidPartyKind,
		},
		{
			name: "empty display name",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode: "v-1",
				Kind:        entity.PartyKindOrganization,
				DisplayName: "   ",
				Status:      entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrInvalidPartyName,
		},
		{
			name: "invalid email",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode: "v-1",
				Kind:        entity.PartyKindOrganization,
				DisplayName: "Vendor A",
				Email:       strPtr("not-an-email"),
				Status:      entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrInvalidEmail,
		},
		{
			name: "invalid website scheme",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode: "v-1",
				Kind:        entity.PartyKindOrganization,
				DisplayName: "Vendor A",
				WebsiteURL:  strPtr("ftp://example.com"),
				Status:      entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrInvalidURL,
		},
		{
			name: "tax exempt without reason violates check constraint",
			id:   testUUID1,
			params: entity.PartyParams{
				PartnerCode:        "v-1",
				Kind:               entity.PartyKindOrganization,
				DisplayName:        "Vendor A",
				IsTaxExempt:        true,
				TaxExemptionReason: strPtr("   "),
				Status:             entity.PartyStatusActive,
			},
			wantErr: eferrors.ErrTaxExemptionReasonRequired,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := entity.NewParty(tc.id, tc.params, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("NewParty() error = %v, want errors.Is(_, %v)", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewParty() unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

func TestParty_LifecycleAndUpdates(t *testing.T) {
	t.Parallel()

	t0 := time.Date(2026, 11, 15, 8, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	t2 := t1.Add(time.Hour)

	p, err := entity.NewParty(testUUID1, entity.PartyParams{
		PartnerCode: "sup-benang",
		Kind:        entity.PartyKindOrganization,
		DisplayName: "CV Benang Prima",
		Status:      entity.PartyStatusPendingVerification,
	}, t0)
	if err != nil {
		t.Fatalf("setup NewParty failed: %v", err)
	}

	if err := p.IsEligibleForTransactions(); !errors.Is(err, eferrors.ErrPartyInactive) {
		t.Fatalf("pending party IsEligibleForTransactions() = %v, want %v", err, eferrors.ErrPartyInactive)
	}

	if err := p.VerifyAndActivate(t1); err != nil {
		t.Fatalf("VerifyAndActivate() failed: %v", err)
	}
	if err := p.IsEligibleForTransactions(); err != nil {
		t.Fatalf("active party IsEligibleForTransactions() failed: %v", err)
	}

	if err := p.UpdateProfile(entity.PartyProfileParams{
		Kind:              entity.PartyKindIndividual,
		DisplayName:       "CV Benang Prima Updated",
		PreferredLanguage: "en",
		Email:             strPtr("billing@benang.co.id"),
	}, t1); err != nil {
		t.Fatalf("UpdateProfile() failed: %v", err)
	}
	if p.PreferredLanguage != "en" || p.DisplayName != "CV Benang Prima Updated" {
		t.Errorf("UpdateProfile did not persist changes: %+v", p)
	}

	if err := p.UpdateTaxProfile(entity.PartyTaxParams{
		TaxIDNumber:        strPtr("02.345.678.9-012.000"),
		TaxRegisteredName:  strPtr("CV Benang Prima"),
		IsPKP:              true,
		IsTaxExempt:        false,
		TaxExemptionReason: nil,
		TaxVerifiedAt:      &t1,
	}, t1); err != nil {
		t.Fatalf("UpdateTaxProfile() failed: %v", err)
	}
	if !p.IsPKP || p.TaxVerifiedAt == nil {
		t.Errorf("UpdateTaxProfile did not persist PKP/TaxVerifiedAt")
	}

	clone := p.Clone()
	if clone == nil || clone == p {
		t.Fatalf("Clone() did not return distinct copy")
	}
	*clone.Email = "mutated@example.com"
	if *p.Email == "mutated@example.com" {
		t.Fatalf("Clone() shared Email pointer")
	}

	if err := p.Suspend(t1); err != nil {
		t.Fatalf("Suspend() failed: %v", err)
	}
	if err := p.IsEligibleForTransactions(); !errors.Is(err, eferrors.ErrPartyInactive) {
		t.Errorf("suspended party IsEligibleForTransactions() = %v, want ErrPartyInactive", err)
	}

	if err := p.Deactivate(t1); err != nil {
		t.Fatalf("Deactivate() failed: %v", err)
	}
	if p.Status != entity.PartyStatusInactive {
		t.Errorf("Status = %q, want inactive", p.Status)
	}

	if err := p.SoftDelete(t2); err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}
	if !p.IsDeleted() || p.Status != entity.PartyStatusInactive {
		t.Errorf("expected deleted=true and status=inactive after SoftDelete")
	}
	if err := p.IsEligibleForTransactions(); !errors.Is(err, eferrors.ErrPartyDeleted) {
		t.Errorf("deleted party IsEligibleForTransactions() = %v, want ErrPartyDeleted", err)
	}
	if err := p.UpdateProfile(entity.PartyProfileParams{Kind: entity.PartyKindOrganization, DisplayName: "X"}, t2); !errors.Is(err, eferrors.ErrPartyDeleted) {
		t.Errorf("UpdateProfile on deleted party = %v, want ErrPartyDeleted", err)
	}
	if err := p.UpdateTaxProfile(entity.PartyTaxParams{}, t2); !errors.Is(err, eferrors.ErrPartyDeleted) {
		t.Errorf("UpdateTaxProfile on deleted party = %v, want ErrPartyDeleted", err)
	}
	if err := p.VerifyAndActivate(t2); !errors.Is(err, eferrors.ErrPartyDeleted) {
		t.Errorf("VerifyAndActivate on deleted party = %v, want ErrPartyDeleted", err)
	}
	if err := p.SoftDelete(t2); !errors.Is(err, eferrors.ErrPartyDeleted) {
		t.Errorf("second SoftDelete = %v, want ErrPartyDeleted", err)
	}

	var nilParty *entity.Party
	if nilParty.Clone() != nil {
		t.Errorf("nil Party.Clone() should return nil")
	}
	if err := nilParty.IsEligibleForTransactions(); !errors.Is(err, eferrors.ErrPartyNotFound) {
		t.Errorf("nil Party.IsEligibleForTransactions() = %v, want ErrPartyNotFound", err)
	}
}
