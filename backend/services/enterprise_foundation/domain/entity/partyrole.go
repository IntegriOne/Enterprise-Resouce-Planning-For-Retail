package entity

import (
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// PartyRole merepresentasikan peran bisnis resmi mitra pada tabel party_roles (seperti supplier, vendor, customer_b2b, carrier).
type PartyRole struct {
	ID            string
	PartyID       string
	Role          PartyRoleKind
	IsActive      bool
	ReferenceNote *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewPartyRole membentuk entitas peran bisnis aktif untuk satu mitra eksternal.
func NewPartyRole(id, partyID string, role PartyRoleKind, referenceNote *string, now time.Time) (*PartyRole, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("party_id", partyID); err != nil {
		return nil, err
	}
	if !role.IsValid() {
		return nil, eferrors.ErrInvalidPartyRoleKind
	}
	note, err := ValidateOptionalText("reference_note", referenceNote, MaxReferenceNoteLength, eferrors.ErrInvalidInput)
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &PartyRole{
		ID:            NormalizeUUID(id),
		PartyID:       NormalizeUUID(partyID),
		Role:          role,
		IsActive:      true,
		ReferenceNote: note,
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}, nil
}

// IsEligibleForScope memeriksa apakah peran mitra cocok dengan jenis peran target dan berstatus aktif.
func (r *PartyRole) IsEligibleForScope(expected PartyRoleKind) bool {
	return r != nil && r.IsActive && r.Role == expected
}

// UpdateReferenceNote memperbarui catatan singkat konteks peran bisnis mitra.
func (r *PartyRole) UpdateReferenceNote(note *string, now time.Time) error {
	if r == nil {
		return eferrors.ErrPartyRoleNotFound
	}
	validated, err := ValidateOptionalText("reference_note", note, MaxReferenceNoteLength, eferrors.ErrInvalidInput)
	if err != nil {
		return err
	}
	r.ReferenceNote = validated
	r.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Activate mengaktifkan kembali peran bisnis mitra.
func (r *PartyRole) Activate(now time.Time) error {
	if r == nil {
		return eferrors.ErrPartyRoleNotFound
	}
	r.IsActive = true
	r.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Deactivate menonaktifkan peran bisnis mitra tanpa menghapus riwayat perannya.
func (r *PartyRole) Deactivate(now time.Time) error {
	if r == nil {
		return eferrors.ErrPartyRoleNotFound
	}
	r.IsActive = false
	r.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Clone mengembalikan salinan mendalam PartyRole.
func (r *PartyRole) Clone() *PartyRole {
	if r == nil {
		return nil
	}
	return &PartyRole{
		ID:            r.ID,
		PartyID:       r.PartyID,
		Role:          r.Role,
		IsActive:      r.IsActive,
		ReferenceNote: cloneStringPtr(r.ReferenceNote),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
