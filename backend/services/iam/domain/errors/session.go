package iamerrors

// Sentinel error domain untuk sesi autentikasi, CSRF, re-autentikasi, dan kesegaran snapshot akses.
var (
	// ErrSessionNotFound menandakan sesi autentikasi tidak ditemukan.
	ErrSessionNotFound = newSentinel(KindNotFound, "iam.session.not_found", "session not found")
	// ErrSessionExpired menandakan masa berlaku sesi telah habis.
	ErrSessionExpired = newSentinel(KindUnauthorized, "iam.session.expired", "session expired")
	// ErrSessionIdleExpired menandakan batas idle timeout 30 menit pada sesi telah terlampaui.
	ErrSessionIdleExpired = newSentinel(KindUnauthorized, "iam.session.idle_expired", "session idle timeout exceeded")
	// ErrSessionAbsoluteExpired menandakan batas absolute timeout 8 jam pada sesi telah terlampaui.
	ErrSessionAbsoluteExpired = newSentinel(KindUnauthorized, "iam.session.absolute_expired", "session absolute timeout exceeded")
	// ErrSessionRevoked menandakan sesi telah dicabut atau diakhiri sebelum kedaluwarsa alami.
	ErrSessionRevoked = newSentinel(KindUnauthorized, "iam.session.revoked", "session revoked")
	// ErrSessionAlreadyEnded menandakan sesi sudah berada dalam status berakhir.
	ErrSessionAlreadyEnded = newSentinel(KindConflict, "iam.session.already_ended", "session already ended")
	// ErrCSRFMismatch menandakan token CSRF permintaan tidak cocok dengan sesi aktif.
	ErrCSRFMismatch = newSentinel(KindForbidden, "iam.session.csrf_mismatch", "csrf token mismatch")
	// ErrSessionReauthRequired menandakan aksi sensitif membutuhkan autentikasi ulang atau verifikasi MFA terbaru.
	ErrSessionReauthRequired = newSentinel(KindForbidden, "iam.session.reauth_required", "recent authentication required for sensitive action")
	// ErrSnapshotStale menandakan snapshot otorisasi melewati batas kesegaran 30 detik dan harus ditolak fail-closed.
	ErrSnapshotStale = newSentinel(KindUnavailable, "iam.session.snapshot_stale", "access snapshot freshness exceeded")
	// ErrInvalidIPAddress menandakan alamat IP klien pada sesi tidak valid.
	ErrInvalidIPAddress = newSentinel(KindInvalidInput, "iam.session.invalid_ip_address", "invalid ip address")
	// ErrInvalidSessionEndReason menandakan alasan pengakhiran sesi tidak dikenali.
	ErrInvalidSessionEndReason = newSentinel(KindInvalidInput, "iam.session.invalid_end_reason", "invalid session end reason")
)
