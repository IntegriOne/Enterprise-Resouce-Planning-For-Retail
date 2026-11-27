package eferrors

// Sentinel error domain untuk entitas badan hukum (legal_entities).
var (
	// ErrLegalEntityNotFound menandakan badan hukum tidak ditemukan.
	ErrLegalEntityNotFound = newSentinel(KindNotFound, "enterprise_foundation.legal_entity.not_found", "legal entity not found")
	// ErrLegalEntityDeleted menandakan badan hukum sudah dihapus secara lunak.
	ErrLegalEntityDeleted = newSentinel(KindConflict, "enterprise_foundation.legal_entity.deleted", "legal entity has been deleted")
	// ErrLegalEntityInactive menandakan badan hukum tidak berstatus aktif untuk transaksi atau scope.
	ErrLegalEntityInactive = newSentinel(KindForbidden, "enterprise_foundation.legal_entity.inactive", "legal entity is not active")
	// ErrLegalEntityCodeExists menandakan kode badan hukum sudah terdaftar pada entitas aktif lain.
	ErrLegalEntityCodeExists = newSentinel(KindAlreadyExists, "enterprise_foundation.legal_entity.code_exists", "legal entity code already exists")
	// ErrLegalEntityTaxIDExists menandakan NPWP/Tax ID badan hukum sudah terdaftar pada entitas aktif lain.
	ErrLegalEntityTaxIDExists = newSentinel(KindAlreadyExists, "enterprise_foundation.legal_entity.tax_id_exists", "legal entity tax id already exists")
	// ErrPrimaryLegalEntityExists menandakan perusahaan induk operasional utama sudah ditetapkan dan tidak boleh ganda atau dihapus sembarangan.
	ErrPrimaryLegalEntityExists = newSentinel(KindConflict, "enterprise_foundation.legal_entity.primary_conflict", "primary operating legal entity conflict")
	// ErrInvalidLegalEntityCode menandakan kode badan hukum tidak memenuhi format atau batas panjang.
	ErrInvalidLegalEntityCode = newSentinel(KindInvalidInput, "enterprise_foundation.legal_entity.invalid_code", "invalid legal entity code")
	// ErrInvalidLegalEntityName menandakan nama resmi atau dagang badan hukum tidak sah.
	ErrInvalidLegalEntityName = newSentinel(KindInvalidInput, "enterprise_foundation.legal_entity.invalid_name", "invalid legal entity name")
	// ErrInvalidLegalEntityStatus menandakan status siklus hidup badan hukum tidak dikenali oleh skema.
	ErrInvalidLegalEntityStatus = newSentinel(KindInvalidInput, "enterprise_foundation.legal_entity.invalid_status", "invalid legal entity status")
	// ErrLegalEntitySelfParent menandakan badan hukum menunjuk dirinya sendiri sebagai induk.
	ErrLegalEntitySelfParent = newSentinel(KindInvalidInput, "enterprise_foundation.legal_entity.self_parent", "legal entity cannot reference itself as parent")
	// ErrLegalEntityCycleDetected menandakan perubahan induk badan hukum membentuk siklus hierarki.
	ErrLegalEntityCycleDetected = newSentinel(KindConflict, "enterprise_foundation.legal_entity.cycle_detected", "legal entity hierarchy cycle detected")
	// ErrInvalidFiscalMonth menandakan bulan awal tahun buku tidak berada pada rentang 1 sampai 12.
	ErrInvalidFiscalMonth = newSentinel(KindInvalidInput, "enterprise_foundation.legal_entity.invalid_fiscal_month", "fiscal year start month must be between 1 and 12")
	// ErrInvalidEntityDates menandakan tanggal pembubaran mendahului tanggal pendirian badan hukum.
	ErrInvalidEntityDates = newSentinel(KindInvalidInput, "enterprise_foundation.legal_entity.invalid_dates", "dissolution date cannot precede incorporation date")
	// ErrLegalEntityHasDependents menandakan badan hukum masih memiliki lokasi atau unit organisasi aktif.
	ErrLegalEntityHasDependents = newSentinel(KindConflict, "enterprise_foundation.legal_entity.has_dependents", "legal entity still has dependent locations or organization units")
)
