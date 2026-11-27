package entity_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestNewAddressAndUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 9, 0, 0, 0, time.Local)
	addr, err := entity.NewAddress(testUUID1, entity.AddressParams{
		Label:        strPtr(" Gudang Utama "),
		AddressLine1: " Jl. Raya Industri No. 12 ",
		City:         " Bekasi ",
		Province:     strPtr(" Jawa Barat "),
		PostalCode:   strPtr(" 17530 "),
		CountryCode:  "",
		Latitude:     float64Ptr(-6.2835),
		Longitude:    float64Ptr(107.1705),
	}, now)
	if err != nil {
		t.Fatalf("NewAddress() error = %v", err)
	}
	if addr.CountryCode != "ID" || addr.City != "Bekasi" || *addr.Label != "Gudang Utama" {
		t.Fatalf("unexpected address normalization: %+v", addr)
	}

	cloned := addr.Clone()
	*cloned.Label = "Changed"
	if *addr.Label != "Gudang Utama" {
		t.Fatal("Clone() mutated original address pointer")
	}

	if err := addr.Update(entity.AddressParams{
		AddressLine1: "Jl. Jenderal Sudirman Kav. 1",
		City:         "Jakarta Pusat",
		CountryCode:  "sg",
		Latitude:     float64Ptr(1.3521),
		Longitude:    float64Ptr(103.8198),
	}, now.Add(time.Minute)); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if addr.CountryCode != "SG" || addr.Label != nil {
		t.Fatalf("unexpected updated address: %+v", addr)
	}
}

func TestNewAddress_InvalidInputs(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nanVal := math.NaN()
	cases := []struct {
		name    string
		id      string
		params  entity.AddressParams
		wantErr error
	}{
		{
			name:    "invalid uuid",
			id:      "bad-id",
			params:  entity.AddressParams{AddressLine1: "Jl. Test", City: "Jakarta"},
			wantErr: eferrors.ErrInvalidInput,
		},
		{
			name:    "empty line 1",
			id:      testUUID1,
			params:  entity.AddressParams{AddressLine1: "   ", City: "Jakarta"},
			wantErr: eferrors.ErrInvalidAddressLine,
		},
		{
			name:    "empty city",
			id:      testUUID1,
			params:  entity.AddressParams{AddressLine1: "Jl. Test", City: ""},
			wantErr: eferrors.ErrInvalidCity,
		},
		{
			name:    "latitude out of range",
			id:      testUUID1,
			params:  entity.AddressParams{AddressLine1: "Jl. Test", City: "Jakarta", Latitude: float64Ptr(-91)},
			wantErr: eferrors.ErrInvalidCoordinates,
		},
		{
			name:    "latitude NaN",
			id:      testUUID1,
			params:  entity.AddressParams{AddressLine1: "Jl. Test", City: "Jakarta", Latitude: &nanVal},
			wantErr: eferrors.ErrInvalidCoordinates,
		},
		{
			name:    "longitude out of range",
			id:      testUUID1,
			params:  entity.AddressParams{AddressLine1: "Jl. Test", City: "Jakarta", Longitude: float64Ptr(180.5)},
			wantErr: eferrors.ErrInvalidCoordinates,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := entity.NewAddress(tc.id, tc.params, now); !errors.Is(err, tc.wantErr) {
				t.Fatalf("NewAddress() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
