package eferrors

// Sentinel error domain untuk entitas unit organisasi internal (organization_units).
var (
	// ErrOrgUnitNotFound menandakan unit organisasi tidak ditemukan.
	ErrOrgUnitNotFound = newSentinel(KindNotFound, "enterprise_foundation.organization_unit.not_found", "organization unit not found")
	// ErrOrgUnitDeleted menandakan unit organisasi sudah dihapus secara lunak.
	ErrOrgUnitDeleted = newSentinel(KindConflict, "enterprise_foundation.organization_unit.deleted", "organization unit has been deleted")
	// ErrOrgUnitInactive menandakan unit organisasi berada dalam status nonaktif.
	ErrOrgUnitInactive = newSentinel(KindForbidden, "enterprise_foundation.organization_unit.inactive", "organization unit is inactive")
	// ErrOrgUnitCodeExists menandakan kode unit organisasi sudah terdaftar dalam badan hukum yang sama.
	ErrOrgUnitCodeExists = newSentinel(KindAlreadyExists, "enterprise_foundation.organization_unit.code_exists", "organization unit code already exists in legal entity")
	// ErrInvalidOrgUnitCode menandakan kode unit organisasi tidak memenuhi format atau batas panjang.
	ErrInvalidOrgUnitCode = newSentinel(KindInvalidInput, "enterprise_foundation.organization_unit.invalid_code", "invalid organization unit code")
	// ErrInvalidOrgUnitName menandakan nama unit organisasi kosong atau melebihi batas panjang.
	ErrInvalidOrgUnitName = newSentinel(KindInvalidInput, "enterprise_foundation.organization_unit.invalid_name", "invalid organization unit name")
	// ErrInvalidOrgUnitKind menandakan jenis unit organisasi tidak dikenali oleh skema.
	ErrInvalidOrgUnitKind = newSentinel(KindInvalidInput, "enterprise_foundation.organization_unit.invalid_kind", "invalid organization unit kind")
	// ErrOrgUnitSelfParent menandakan unit organisasi menunjuk dirinya sendiri sebagai unit induk.
	ErrOrgUnitSelfParent = newSentinel(KindInvalidInput, "enterprise_foundation.organization_unit.self_parent", "organization unit cannot reference itself as parent")
	// ErrOrgUnitParentMismatch menandakan unit induk berada pada badan hukum yang berbeda.
	ErrOrgUnitParentMismatch = newSentinel(KindConflict, "enterprise_foundation.organization_unit.parent_entity_mismatch", "parent organization unit must belong to the same legal entity")
	// ErrOrgUnitLocationMismatch menandakan lokasi utama unit berada pada badan hukum yang berbeda.
	ErrOrgUnitLocationMismatch = newSentinel(KindConflict, "enterprise_foundation.organization_unit.location_entity_mismatch", "primary location must belong to the same legal entity")
	// ErrOrgUnitCycleDetected menandakan hierarki unit organisasi membentuk siklus.
	ErrOrgUnitCycleDetected = newSentinel(KindConflict, "enterprise_foundation.organization_unit.cycle_detected", "organization unit hierarchy cycle detected")
)
