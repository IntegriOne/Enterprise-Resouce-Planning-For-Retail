package entity

import (
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// PartyAddress merepresentasikan relasi antara mitra eksternal (parties) dan master alamat (addresses) beserta peruntukannya pada tabel party_addresses.
type PartyAddress struct {
	ID        string
	PartyID   string
	AddressID string
	Usage     AddressUsageKind
	IsPrimary bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPartyAddress membentuk tautan alamat milik mitra dengan validasi UUID dan jenis peruntukan alamat.
func NewPartyAddress(id, partyID, addressID string, usage AddressUsageKind, isPrimary bool, now time.Time) (*PartyAddress, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("party_id", partyID); err != nil {
		return nil, err
	}
	if err := ValidateUUID("address_id", addressID); err != nil {
		return nil, err
	}
	if !usage.IsValid() {
		return nil, eferrors.ErrInvalidAddressUsage
	}

	ts := normalizeLocalTime(now)
	return &PartyAddress{
		ID:        NormalizeUUID(id),
		PartyID:   NormalizeUUID(partyID),
		AddressID: NormalizeUUID(addressID),
		Usage:     usage,
		IsPrimary: isPrimary,
		CreatedAt: ts,
		UpdatedAt: ts,
	}, nil
}

// UpdateUsage memperbarui jenis peruntukan dan penanda alamat utama untuk peruntukan tersebut.
func (pa *PartyAddress) UpdateUsage(usage AddressUsageKind, isPrimary bool, now time.Time) error {
	if pa == nil {
		return eferrors.ErrPartyAddressNotFound
	}
	if !usage.IsValid() {
		return eferrors.ErrInvalidAddressUsage
	}
	pa.Usage = usage
	pa.IsPrimary = isPrimary
	pa.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SetPrimary menandai atau melepas status alamat utama pada peruntukan terkait.
func (pa *PartyAddress) SetPrimary(primary bool, now time.Time) error {
	if pa == nil {
		return eferrors.ErrPartyAddressNotFound
	}
	pa.IsPrimary = primary
	pa.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Clone mengembalikan salinan PartyAddress.
func (pa *PartyAddress) Clone() *PartyAddress {
	if pa == nil {
		return nil
	}
	copyVal := *pa
	return &copyVal
}
