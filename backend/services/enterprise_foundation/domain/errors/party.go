package eferrors

// Sentinel error domain untuk entitas mitra eksternal (parties), peran (party_roles), PIC portal (party_contacts), dan alamat mitra (party_addresses).
var (
	// ErrPartyNotFound menandakan entitas mitra eksternal tidak ditemukan.
	ErrPartyNotFound = newSentinel(KindNotFound, "enterprise_foundation.party.not_found", "party not found")
	// ErrPartyDeleted menandakan entitas mitra eksternal sudah dihapus secara lunak.
	ErrPartyDeleted = newSentinel(KindConflict, "enterprise_foundation.party.deleted", "party has been deleted")
	// ErrPartyInactive menandakan status mitra tidak aktif atau belum terverifikasi untuk bertransaksi.
	ErrPartyInactive = newSentinel(KindForbidden, "enterprise_foundation.party.inactive", "party is not active")
	// ErrPartyCodeExists menandakan kode partner mitra sudah terdaftar pada mitra aktif lain.
	ErrPartyCodeExists = newSentinel(KindAlreadyExists, "enterprise_foundation.party.code_exists", "party partner code already exists")
	// ErrPartyTaxIDExists menandakan NPWP/Tax ID mitra sudah terdaftar pada mitra aktif lain.
	ErrPartyTaxIDExists = newSentinel(KindAlreadyExists, "enterprise_foundation.party.tax_id_exists", "party tax id already exists")
	// ErrInvalidPartyCode menandakan kode partner mitra tidak memenuhi format atau batas panjang.
	ErrInvalidPartyCode = newSentinel(KindInvalidInput, "enterprise_foundation.party.invalid_code", "invalid party partner code")
	// ErrInvalidPartyName menandakan nama tampilan atau nama resmi mitra tidak sah.
	ErrInvalidPartyName = newSentinel(KindInvalidInput, "enterprise_foundation.party.invalid_name", "invalid party name")
	// ErrInvalidPartyKind menandakan tipe bentuk entitas mitra tidak dikenali oleh skema.
	ErrInvalidPartyKind = newSentinel(KindInvalidInput, "enterprise_foundation.party.invalid_kind", "invalid party kind")
	// ErrInvalidPartyStatus menandakan status siklus hidup mitra tidak dikenali oleh skema.
	ErrInvalidPartyStatus = newSentinel(KindInvalidInput, "enterprise_foundation.party.invalid_status", "invalid party status")
	// ErrTaxExemptionReasonRequired menandakan alasan pembebasan pajak wajib diisi saat is_tax_exempt bernilai true.
	ErrTaxExemptionReasonRequired = newSentinel(KindInvalidInput, "enterprise_foundation.party.tax_exemption_reason_required", "tax exemption reason is required when tax exempt is enabled")

	// ErrPartyRoleNotFound menandakan peran mitra tidak ditemukan.
	ErrPartyRoleNotFound = newSentinel(KindNotFound, "enterprise_foundation.party_role.not_found", "party role not found")
	// ErrPartyRoleExists menandakan peran yang sama sudah terpasang pada mitra tersebut.
	ErrPartyRoleExists = newSentinel(KindAlreadyExists, "enterprise_foundation.party_role.already_exists", "party role already exists")
	// ErrPartyRoleInactive menandakan peran mitra berstatus nonaktif.
	ErrPartyRoleInactive = newSentinel(KindForbidden, "enterprise_foundation.party_role.inactive", "party role is inactive")
	// ErrInvalidPartyRoleKind menandakan jenis peran bisnis mitra tidak dikenali oleh skema.
	ErrInvalidPartyRoleKind = newSentinel(KindInvalidInput, "enterprise_foundation.party_role.invalid_kind", "invalid party role kind")

	// ErrPartyContactNotFound menandakan kontak/PIC mitra tidak ditemukan.
	ErrPartyContactNotFound = newSentinel(KindNotFound, "enterprise_foundation.party_contact.not_found", "party contact not found")
	// ErrPartyContactDeleted menandakan kontak/PIC mitra sudah dihapus secara lunak.
	ErrPartyContactDeleted = newSentinel(KindConflict, "enterprise_foundation.party_contact.deleted", "party contact has been deleted")
	// ErrPartyContactInactive menandakan kontak/PIC mitra berada dalam status nonaktif.
	ErrPartyContactInactive = newSentinel(KindForbidden, "enterprise_foundation.party_contact.inactive", "party contact is inactive")
	// ErrPartyContactIAMUserExists menandakan principal IAM tersebut sudah tertaut ke kontak lain pada mitra yang sama.
	ErrPartyContactIAMUserExists = newSentinel(KindAlreadyExists, "enterprise_foundation.party_contact.iam_user_already_linked", "iam user is already linked to this party")
	// ErrContactChannelRequired menandakan kontak harus memiliki minimal salah satu antara email, phone, atau iam_user_id.
	ErrContactChannelRequired = newSentinel(KindInvalidInput, "enterprise_foundation.party_contact.channel_required", "party contact requires email, phone, or iam user id")
	// ErrInvalidPortalAccessStatus menandakan status akses portal PIC tidak dikenali oleh skema.
	ErrInvalidPortalAccessStatus = newSentinel(KindInvalidInput, "enterprise_foundation.party_contact.invalid_portal_access_status", "invalid portal access status")
	// ErrPortalLinkInvariantViolation menandakan pelanggaran invarian antara iam_user_id dan portal_access_status.
	ErrPortalLinkInvariantViolation = newSentinel(KindInvalidInput, "enterprise_foundation.party_contact.portal_link_invariant", "iam user id and portal access status must match portal link invariant")
	// ErrPortalAccessNotActive menandakan kontak belum memiliki tautan portal aktif ke IAM.
	ErrPortalAccessNotActive = newSentinel(KindForbidden, "enterprise_foundation.party_contact.portal_access_not_active", "party contact portal access is not active")

	// ErrPartyAddressNotFound menandakan relasi alamat milik mitra tidak ditemukan.
	ErrPartyAddressNotFound = newSentinel(KindNotFound, "enterprise_foundation.party_address.not_found", "party address not found")
	// ErrPartyAddressExists menandakan alamat dengan peruntukan yang sama sudah tertaut pada mitra tersebut.
	ErrPartyAddressExists = newSentinel(KindAlreadyExists, "enterprise_foundation.party_address.already_exists", "party address link already exists for usage")
	// ErrPrimaryPartyAddressConflict menandakan konflik alamat utama untuk peruntukan yang sama pada mitra.
	ErrPrimaryPartyAddressConflict = newSentinel(KindConflict, "enterprise_foundation.party_address.primary_conflict", "primary party address conflict for usage")
)
