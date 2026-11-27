package eferrors

// Sentinel error domain untuk outbox integrasi, validasi format umum, serta evaluasi scope lintas modul.
var (
	// ErrOutboxEventNotFound menandakan entri event outbox tidak ditemukan.
	ErrOutboxEventNotFound = newSentinel(KindNotFound, "enterprise_foundation.outbox.not_found", "outbox event not found")
	// ErrInvalidOutboxStatus menandakan status pengiriman outbox tidak dikenali oleh skema.
	ErrInvalidOutboxStatus = newSentinel(KindInvalidInput, "enterprise_foundation.outbox.invalid_status", "invalid outbox status")
	// ErrInvalidOutboxPayload menandakan payload JSON event outbox atau atribut JSONB tidak valid.
	ErrInvalidOutboxPayload = newSentinel(KindInvalidInput, "enterprise_foundation.outbox.invalid_payload", "invalid json payload")

	// ErrInvalidEmail menandakan format alamat email tidak sesuai standar.
	ErrInvalidEmail = newSentinel(KindInvalidInput, "enterprise_foundation.invalid_email", "invalid email format")
	// ErrInvalidPhone menandakan format nomor telepon tidak valid atau melebihi batas panjang.
	ErrInvalidPhone = newSentinel(KindInvalidInput, "enterprise_foundation.invalid_phone", "invalid phone format")
	// ErrInvalidURL menandakan URL situs tidak menggunakan skema http/https yang valid atau melebihi batas panjang.
	ErrInvalidURL = newSentinel(KindInvalidInput, "enterprise_foundation.invalid_url", "invalid website url format")

	// ErrInvalidScopeTarget menandakan tipe atau ID target scope tidak valid untuk Enterprise Foundation.
	ErrInvalidScopeTarget = newSentinel(KindInvalidInput, "enterprise_foundation.scope.invalid_target", "invalid scope target")
	// ErrScopeTargetNotEligible menandakan target scope ditemukan tetapi tidak aktif atau tidak memenuhi syarat bisnis.
	ErrScopeTargetNotEligible = newSentinel(KindForbidden, "enterprise_foundation.scope.not_eligible", "scope target is not eligible")
)
