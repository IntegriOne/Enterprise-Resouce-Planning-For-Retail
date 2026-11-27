package repository_test

import (
	"math"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
)

// Pastikan seluruh kontrak antarmuka repositori ekspor domain tersedia dan terkompilasi.
var _ = []any{
	(*repository.Transactor)(nil),
	(*repository.AddressRepository)(nil),
	(*repository.LegalEntityRepository)(nil),
	(*repository.LocationRepository)(nil),
	(*repository.OrganizationUnitRepository)(nil),
	(*repository.PartyRepository)(nil),
	(*repository.PartyRoleRepository)(nil),
	(*repository.PartyContactRepository)(nil),
	(*repository.PartyAddressRepository)(nil),
	(*repository.OutboxRepository)(nil),
}

func strPtr(s string) *string {
	return &s
}

func TestPageRequest_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   repository.PageRequest
		want repository.PageRequest
	}{
		{
			name: "zero values clamp to default limit and zero offset",
			in:   repository.PageRequest{Limit: 0, Offset: -5},
			want: repository.PageRequest{Limit: repository.DefaultPageLimit, Offset: 0},
		},
		{
			name: "negative limit clamps to default",
			in:   repository.PageRequest{Limit: -10, Offset: 10},
			want: repository.PageRequest{Limit: repository.DefaultPageLimit, Offset: 10},
		},
		{
			name: "limit above max clamps to max",
			in:   repository.PageRequest{Limit: 999, Offset: 40},
			want: repository.PageRequest{Limit: repository.MaxPageLimit, Offset: 40},
		},
		{
			name: "valid bounds preserved",
			in:   repository.PageRequest{Limit: 50, Offset: 20},
			want: repository.PageRequest{Limit: 50, Offset: 20},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.in.Normalize(); got != tc.want {
				t.Errorf("Normalize() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestNewPageResult(t *testing.T) {
	t.Parallel()

	t.Run("nil items converts to empty non-nil slice", func(t *testing.T) {
		t.Parallel()
		res := repository.NewPageResult[string](nil, 0, repository.PageRequest{Limit: 0, Offset: -1})
		if res.Items == nil || len(res.Items) != 0 {
			t.Fatalf("expected non-nil empty slice, got %#v", res.Items)
		}
		if res.Limit != repository.DefaultPageLimit || res.Offset != 0 || res.HasMore {
			t.Errorf("unexpected metadata: %+v", res)
		}
	})

	t.Run("has_more reflects remaining rows", func(t *testing.T) {
		t.Parallel()
		res := repository.NewPageResult([]string{"a", "b"}, 5, repository.PageRequest{Limit: 2, Offset: 0})
		if !res.HasMore {
			t.Errorf("expected HasMore=true when offset+len < total")
		}
		lastPage := repository.NewPageResult([]string{"e"}, 5, repository.PageRequest{Limit: 2, Offset: 4})
		if lastPage.HasMore {
			t.Errorf("expected HasMore=false on final page")
		}
		extremeOffset := repository.NewPageResult([]string{"x"}, 100, repository.PageRequest{Limit: 10, Offset: math.MaxInt})
		if extremeOffset.HasMore {
			t.Errorf("expected HasMore=false on extreme offset without int64 overflow")
		}
	})
}

func TestFilters_Normalize(t *testing.T) {
	t.Parallel()

	upperUUID := "AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE"
	lowerUUID := strings.ToLower(upperUUID)

	t.Run("AddressFilter", func(t *testing.T) {
		t.Parallel()
		got := (repository.AddressFilter{
			Query:       "  gudang pusat ",
			CountryCode: " id ",
			City:        " Bandung ",
			Province:    " Jawa Barat ",
			PostalCode:  " 40111 ",
			Page:        repository.PageRequest{Limit: 500, Offset: -3},
		}).Normalize()

		if got.Query != "gudang pusat" || got.CountryCode != "ID" || got.City != "Bandung" || got.Province != "Jawa Barat" || got.PostalCode != "40111" {
			t.Errorf("unexpected normalized AddressFilter: %+v", got)
		}
		if got.Page.Limit != repository.MaxPageLimit || got.Page.Offset != 0 {
			t.Errorf("unexpected normalized Page: %+v", got.Page)
		}
	})

	t.Run("LegalEntityFilter", func(t *testing.T) {
		t.Parallel()
		got := (repository.LegalEntityFilter{
			Query:               " PT Maju ",
			ParentLegalEntityID: strPtr(" " + upperUUID + " "),
		}).Normalize()
		if got.Query != "PT Maju" || got.ParentLegalEntityID == nil || *got.ParentLegalEntityID != lowerUUID {
			t.Errorf("unexpected normalized LegalEntityFilter: %+v", got)
		}

		blankParent := (repository.LegalEntityFilter{
			ParentLegalEntityID: strPtr("   "),
		}).Normalize()
		if blankParent.ParentLegalEntityID != nil {
			t.Errorf("blank ParentLegalEntityID should normalize to nil, got %v", blankParent.ParentLegalEntityID)
		}
	})

	t.Run("LocationFilter", func(t *testing.T) {
		t.Parallel()
		got := (repository.LocationFilter{
			Query:            " wh ",
			LegalEntityID:    " " + upperUUID + " ",
			ParentLocationID: strPtr(" "),
			AddressID:        strPtr(" " + upperUUID + " "),
		}).Normalize()
		if got.Query != "wh" || got.LegalEntityID != lowerUUID || got.ParentLocationID != nil || got.AddressID == nil || *got.AddressID != lowerUUID {
			t.Errorf("unexpected normalized LocationFilter: %+v", got)
		}
	})

	t.Run("OrganizationUnitFilter", func(t *testing.T) {
		t.Parallel()
		got := (repository.OrganizationUnitFilter{
			Query:             " finance ",
			LegalEntityID:     upperUUID,
			ParentUnitID:      strPtr(" " + upperUUID + " "),
			PrimaryLocationID: strPtr("  "),
		}).Normalize()
		if got.Query != "finance" || got.LegalEntityID != lowerUUID {
			t.Errorf("unexpected OrgUnit query/entity: %+v", got)
		}
		if got.ParentUnitID == nil || *got.ParentUnitID != lowerUUID || got.PrimaryLocationID != nil {
			t.Errorf("unexpected OrgUnit pointer filters: %+v", got)
		}
	})

	t.Run("PartyFilter", func(t *testing.T) {
		t.Parallel()
		role := entity.PartyRoleKindSupplier
		got := (repository.PartyFilter{
			Query: "  kain ",
			Role:  &role,
			Page:  repository.PageRequest{Limit: -1},
		}).Normalize()
		if got.Query != "kain" || got.Page.Limit != repository.DefaultPageLimit {
			t.Errorf("unexpected normalized PartyFilter: %+v", got)
		}
	})

	t.Run("PartyContactFilter", func(t *testing.T) {
		t.Parallel()
		got := (repository.PartyContactFilter{
			PartyID:     " " + upperUUID + " ",
			Query:       " Budi ",
			ContactRole: "  PORTAL_ADMIN ",
		}).Normalize()
		if got.PartyID != lowerUUID || got.Query != "Budi" || got.ContactRole != "portal_admin" {
			t.Errorf("unexpected normalized PartyContactFilter: %+v", got)
		}
	})

	t.Run("PartyRoleFilter", func(t *testing.T) {
		t.Parallel()
		role := entity.PartyRoleKindVendor
		got := (repository.PartyRoleFilter{
			PartyID:    " " + upperUUID + " ",
			Role:       &role,
			ActiveOnly: true,
			Query:      "  utama ",
			Page:       repository.PageRequest{Limit: 250, Offset: -2},
		}).Normalize()
		if got.PartyID != lowerUUID || got.Query != "utama" || got.Page.Limit != repository.MaxPageLimit || got.Page.Offset != 0 {
			t.Errorf("unexpected normalized PartyRoleFilter: %+v", got)
		}
	})

	t.Run("PartyAddressFilter", func(t *testing.T) {
		t.Parallel()
		usage := entity.AddressUsageKindShipping
		prim := true
		got := (repository.PartyAddressFilter{
			PartyID:     " " + upperUUID + " ",
			AddressID:   " " + upperUUID + " ",
			Usage:       &usage,
			PrimaryOnly: &prim,
			Page:        repository.PageRequest{Limit: 0, Offset: 10},
		}).Normalize()
		if got.PartyID != lowerUUID || got.AddressID != lowerUUID || got.Page.Limit != repository.DefaultPageLimit || got.Page.Offset != 10 {
			t.Errorf("unexpected normalized PartyAddressFilter: %+v", got)
		}
	})
}
