package eferrors

// Sentinel error domain untuk entitas lokasi fisik internal (locations).
var (
	// ErrLocationNotFound menandakan lokasi tidak ditemukan.
	ErrLocationNotFound = newSentinel(KindNotFound, "enterprise_foundation.location.not_found", "location not found")
	// ErrLocationDeleted menandakan lokasi sudah dihapus secara lunak.
	ErrLocationDeleted = newSentinel(KindConflict, "enterprise_foundation.location.deleted", "location has been deleted")
	// ErrLocationInactive menandakan lokasi berada dalam status nonaktif.
	ErrLocationInactive = newSentinel(KindForbidden, "enterprise_foundation.location.inactive", "location is inactive")
	// ErrLocationCodeExists menandakan kode lokasi sudah terdaftar pada lokasi aktif lain.
	ErrLocationCodeExists = newSentinel(KindAlreadyExists, "enterprise_foundation.location.code_exists", "location code already exists")
	// ErrInvalidLocationCode menandakan kode lokasi tidak memenuhi format atau batas panjang.
	ErrInvalidLocationCode = newSentinel(KindInvalidInput, "enterprise_foundation.location.invalid_code", "invalid location code")
	// ErrInvalidLocationName menandakan nama lokasi kosong atau melebihi batas panjang.
	ErrInvalidLocationName = newSentinel(KindInvalidInput, "enterprise_foundation.location.invalid_name", "invalid location name")
	// ErrInvalidLocationKind menandakan jenis lokasi tidak dikenali oleh skema.
	ErrInvalidLocationKind = newSentinel(KindInvalidInput, "enterprise_foundation.location.invalid_kind", "invalid location kind")
	// ErrInvalidTimezone menandakan zona waktu lokasi kosong, tidak valid, atau melebihi batas panjang.
	ErrInvalidTimezone = newSentinel(KindInvalidInput, "enterprise_foundation.location.invalid_timezone", "invalid timezone")
	// ErrLocationSelfParent menandakan lokasi menunjuk dirinya sendiri sebagai lokasi induk.
	ErrLocationSelfParent = newSentinel(KindInvalidInput, "enterprise_foundation.location.self_parent", "location cannot reference itself as parent")
	// ErrLocationParentMismatch menandakan lokasi induk berada pada badan hukum yang berbeda.
	ErrLocationParentMismatch = newSentinel(KindConflict, "enterprise_foundation.location.parent_entity_mismatch", "parent location must belong to the same legal entity")
	// ErrLocationCycleDetected menandakan hierarki lokasi membentuk siklus.
	ErrLocationCycleDetected = newSentinel(KindConflict, "enterprise_foundation.location.cycle_detected", "location hierarchy cycle detected")
)
