package enterprise_foundation

import (
	"slices"
	"testing"

	"github.com/integrione/erp-retail/backend/shared/rbac"
)

func TestEnterpriseFoundationPermissionsMatchManifestCatalog(t *testing.T) {
	t.Parallel()

	catalogs, err := rbac.LoadEmbeddedCatalogs()
	if err != nil {
		t.Fatalf("LoadEmbeddedCatalogs() error = %v", err)
	}

	gotAdapterCodes := []string{
		PermAddressesRead,
		PermAddressesWrite,
		PermLegalEntitiesRead,
		PermLegalEntitiesWrite,
		PermLocationsRead,
		PermLocationsWrite,
		PermOrganizationUnitsRead,
		PermOrganizationUnitsWrite,
		PermPartiesRead,
		PermPartiesWrite,
	}
	slices.Sort(gotAdapterCodes)

	wantCatalogCodes := catalogs.PermissionCodesForDomain("enterprise_foundation")
	if !slices.Equal(gotAdapterCodes, wantCatalogCodes) {
		t.Fatalf("Enterprise Foundation adapter permission codes drifted from shared/rbac manifest:\ngot  = %v\nwant = %v", gotAdapterCodes, wantCatalogCodes)
	}

	if roleSpec, ok := catalogs.RoleByCode["ef_automation"]; !ok {
		t.Fatal("ef_automation role missing from shared/rbac catalogs")
	} else {
		roleCodes := slices.Clone(roleSpec.PermissionCodes)
		slices.Sort(roleCodes)
		if !slices.Equal(roleCodes, wantCatalogCodes) {
			t.Fatalf("ef_automation permission_codes drifted from domain catalog:\ngot  = %v\nwant = %v", roleCodes, wantCatalogCodes)
		}
	}
}
