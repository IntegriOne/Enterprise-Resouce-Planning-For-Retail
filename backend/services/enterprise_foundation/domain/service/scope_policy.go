package service

import (
	"fmt"
	"strings"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// ScopeTargetKind mengidentifikasi tipe objek scope milik Enterprise Foundation yang divalidasi oleh IAM dan modul bisnis lain.
type ScopeTargetKind string

const (
	// ScopeTargetLegalEntity memvalidasi target scope bertipe legal_entity.
	ScopeTargetLegalEntity ScopeTargetKind = "legal_entity"
	// ScopeTargetOrganizationUnit memvalidasi target scope bertipe organization_unit.
	ScopeTargetOrganizationUnit ScopeTargetKind = "organization_unit"
	// ScopeTargetFacility memvalidasi target scope bertipe facility yang berasal dari locations.
	ScopeTargetFacility ScopeTargetKind = "facility"
	// ScopeTargetSupplierParty memvalidasi target scope bertipe supplier_party (mitra aktif berperan supplier).
	ScopeTargetSupplierParty ScopeTargetKind = "supplier_party"
	// ScopeTargetVendorParty memvalidasi target scope bertipe vendor_party (mitra aktif berperan vendor).
	ScopeTargetVendorParty ScopeTargetKind = "vendor_party"
)

// String mengembalikan representasi teks ScopeTargetKind.
func (k ScopeTargetKind) String() string {
	return string(k)
}

// IsValid memeriksa apakah ScopeTargetKind didukung oleh Enterprise Foundation.
func (k ScopeTargetKind) IsValid() bool {
	switch k {
	case ScopeTargetLegalEntity,
		ScopeTargetOrganizationUnit,
		ScopeTargetFacility,
		ScopeTargetSupplierParty,
		ScopeTargetVendorParty:
		return true
	default:
		return false
	}
}

// ParseScopeTargetKind menormalkan dan memvalidasi tipe target scope milik Enterprise Foundation.
func ParseScopeTargetKind(raw string) (ScopeTargetKind, error) {
	k := ScopeTargetKind(strings.ToLower(strings.TrimSpace(raw)))
	if !k.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidScopeTarget, raw)
	}
	return k, nil
}

// ScopeTargetDecision merangkum hasil evaluasi kelayakan objek target scope secara fail-closed.
type ScopeTargetDecision struct {
	Eligible      bool
	ScopeKind     ScopeTargetKind
	TargetID      string
	Code          string
	LegalEntityID string
}

// ScopePolicy mengevaluasi kelayakan target scope dan referensi master data lintas modul secara fail-closed.
type ScopePolicy struct{}

// NewScopePolicy membentuk evaluator kelayakan target scope Enterprise Foundation.
func NewScopePolicy() ScopePolicy {
	return ScopePolicy{}
}

// EvaluateLegalEntity memeriksa kelayakan badan hukum untuk transaksi bisnis dan scope IAM legal_entity.
func (ScopePolicy) EvaluateLegalEntity(le *entity.LegalEntity) (ScopeTargetDecision, error) {
	if err := le.CanBeUsedInTransactions(); err != nil {
		return ScopeTargetDecision{}, err
	}
	return ScopeTargetDecision{
		Eligible:      true,
		ScopeKind:     ScopeTargetLegalEntity,
		TargetID:      le.ID,
		Code:          le.EntityCode,
		LegalEntityID: le.ID,
	}, nil
}

// EvaluateOrganizationUnit memeriksa kelayakan unit organisasi sekaligus memastikan badan hukum pemiliknya juga aktif.
func (ScopePolicy) EvaluateOrganizationUnit(unit *entity.OrganizationUnit, owner *entity.LegalEntity) (ScopeTargetDecision, error) {
	if err := unit.CanOperate(); err != nil {
		return ScopeTargetDecision{}, err
	}
	if err := owner.CanBeUsedInTransactions(); err != nil {
		return ScopeTargetDecision{}, err
	}
	if !strings.EqualFold(entity.NormalizeUUID(unit.LegalEntityID), entity.NormalizeUUID(owner.ID)) {
		return ScopeTargetDecision{}, eferrors.ErrScopeTargetNotEligible
	}
	return ScopeTargetDecision{
		Eligible:      true,
		ScopeKind:     ScopeTargetOrganizationUnit,
		TargetID:      unit.ID,
		Code:          unit.UnitCode,
		LegalEntityID: owner.ID,
	}, nil
}

// EvaluateFacility memeriksa kelayakan lokasi fisik (kantor, gudang, toko, cabang) sekaligus badan hukum pemiliknya.
func (ScopePolicy) EvaluateFacility(loc *entity.Location, owner *entity.LegalEntity) (ScopeTargetDecision, error) {
	if err := loc.CanOperate(); err != nil {
		return ScopeTargetDecision{}, err
	}
	if err := owner.CanBeUsedInTransactions(); err != nil {
		return ScopeTargetDecision{}, err
	}
	if !strings.EqualFold(entity.NormalizeUUID(loc.LegalEntityID), entity.NormalizeUUID(owner.ID)) {
		return ScopeTargetDecision{}, eferrors.ErrScopeTargetNotEligible
	}
	return ScopeTargetDecision{
		Eligible:      true,
		ScopeKind:     ScopeTargetFacility,
		TargetID:      loc.ID,
		Code:          loc.LocationCode,
		LegalEntityID: owner.ID,
	}, nil
}

// EvaluateWarehouse memverifikasi secara fail-closed bahwa fasilitas target berstatus aktif dan berjenis gudang.
func (p ScopePolicy) EvaluateWarehouse(loc *entity.Location, owner *entity.LegalEntity) (ScopeTargetDecision, error) {
	decision, err := p.EvaluateFacility(loc, owner)
	if err != nil {
		return ScopeTargetDecision{}, err
	}
	if loc.Kind != entity.LocationKindWarehouse {
		return ScopeTargetDecision{}, eferrors.ErrScopeTargetNotEligible
	}
	return decision, nil
}

// EvaluatePartyRole memverifikasi bahwa mitra aktif memiliki peran bisnis spesifik yang aktif untuk transaksi atau scope IAM.
func (ScopePolicy) EvaluatePartyRole(
	party *entity.Party,
	roles []*entity.PartyRole,
	requiredRole entity.PartyRoleKind,
) (ScopeTargetDecision, error) {
	if !requiredRole.IsValid() {
		return ScopeTargetDecision{}, eferrors.ErrInvalidPartyRoleKind
	}
	if err := party.IsEligibleForTransactions(); err != nil {
		return ScopeTargetDecision{}, err
	}

	var matchedRole *entity.PartyRole
	for _, r := range roles {
		if r != nil && strings.EqualFold(entity.NormalizeUUID(r.PartyID), entity.NormalizeUUID(party.ID)) && r.Role == requiredRole {
			matchedRole = r
			break
		}
	}
	if matchedRole == nil {
		return ScopeTargetDecision{}, eferrors.ErrPartyRoleNotFound
	}
	if !matchedRole.IsActive {
		return ScopeTargetDecision{}, eferrors.ErrPartyRoleInactive
	}

	scopeKind := scopeKindFromPartyRole(requiredRole)
	return ScopeTargetDecision{
		Eligible:  true,
		ScopeKind: scopeKind,
		TargetID:  party.ID,
		Code:      party.PartnerCode,
	}, nil
}

func scopeKindFromPartyRole(role entity.PartyRoleKind) ScopeTargetKind {
	if mapped, ok := role.IAMScopeType(); ok {
		return ScopeTargetKind(mapped)
	}
	return ScopeTargetKind(role.String())
}
