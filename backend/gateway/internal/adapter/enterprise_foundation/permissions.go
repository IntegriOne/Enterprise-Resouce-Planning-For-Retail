package enterprise_foundation

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

const (
	// PermAddressesRead mengizinkan pembacaan data master alamat geografis.
	PermAddressesRead = "enterprise_foundation.addresses.read"
	// PermAddressesWrite mengizinkan pembuatan, pembaruan, dan penghapusan master alamat geografis.
	PermAddressesWrite = "enterprise_foundation.addresses.write"

	// PermLegalEntitiesRead mengizinkan pembacaan data badan hukum perusahaan dan validasi scope legal_entity.
	PermLegalEntitiesRead = "enterprise_foundation.legal_entities.read"
	// PermLegalEntitiesWrite mengizinkan pengelolaan siklus hidup badan hukum perusahaan.
	PermLegalEntitiesWrite = "enterprise_foundation.legal_entities.write"

	// PermLocationsRead mengizinkan pembacaan lokasi fisik internal dan validasi scope facility/warehouse.
	PermLocationsRead = "enterprise_foundation.locations.read"
	// PermLocationsWrite mengizinkan pengelolaan lokasi fisik operasional internal.
	PermLocationsWrite = "enterprise_foundation.locations.write"

	// PermOrganizationUnitsRead mengizinkan pembacaan unit organisasi internal dan validasi scope organization_unit.
	PermOrganizationUnitsRead = "enterprise_foundation.organization_units.read"
	// PermOrganizationUnitsWrite mengizinkan pengelolaan struktur dan kepala unit organisasi internal.
	PermOrganizationUnitsWrite = "enterprise_foundation.organization_units.write"

	// PermPartiesRead mengizinkan pembacaan mitra eksternal, peran bisnis, PIC portal IAM, dan relasi alamat.
	PermPartiesRead = "enterprise_foundation.parties.read"
	// PermPartiesWrite mengizinkan pengelolaan mitra eksternal, peran, PIC, akses portal IAM, dan relasi alamat.
	PermPartiesWrite = "enterprise_foundation.parties.write"
)

const (
	scopeTypeLegalEntity      = "legal_entity"
	scopeTypeOrganizationUnit = "organization_unit"
	scopeTypeFacility         = "facility"
	scopeTypeSupplierParty    = "supplier_party"
	scopeTypeVendorParty      = "vendor_party"
)

func (a *Adapter) requirePerm(ctx context.Context, permCode string, requireSensitive bool) (*middleware.ResolvedSession, error) {
	return a.requireAnyScope(ctx, permCode, requireSensitive, authz.ScopeGrant{Type: authz.ScopeGlobal})
}

func (a *Adapter) requireAnyScope(
	ctx context.Context,
	permCode string,
	requireSensitive bool,
	scopes ...authz.ScopeGrant,
) (*middleware.ResolvedSession, error) {
	if a == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("authorizer unavailable", nil)
	}
	if len(scopes) == 0 {
		scopes = []authz.ScopeGrant{{Type: authz.ScopeGlobal}}
	}

	var firstForbidden error
	for _, sc := range scopes {
		sess, err := a.authorizer.RequirePermission(ctx, permCode, sc, requireSensitive)
		if err == nil {
			return sess, nil
		}
		gwErr, ok := middleware.AsGatewayError(err)
		if !ok || gwErr.Code != middleware.CodeForbidden {
			return nil, err
		}
		// Jangan lanjutkan fallback ke scope lain bila kegagalan dipicu oleh wajib re-autentikasi sensitif/MFA.
		if !strings.HasPrefix(gwErr.Message, "missing required permission:") {
			return nil, err
		}
		if firstForbidden == nil {
			firstForbidden = err
		}
	}
	return nil, firstForbidden
}

func partyScopes(partyID string) []authz.ScopeGrant {
	if strings.TrimSpace(partyID) == "" {
		return []authz.ScopeGrant{{Type: authz.ScopeGlobal}}
	}
	return []authz.ScopeGrant{
		{Type: scopeTypeVendorParty, ID: partyID},
		{Type: scopeTypeSupplierParty, ID: partyID},
	}
}
