package iamerrors

// Sentinel error domain untuk pengguna, identitas login, kredensial, MFA TOTP, kode pemulihan, dan token.
var (
	// ErrUserNotFound menandakan akun pengguna tidak ditemukan.
	ErrUserNotFound = newSentinel(KindNotFound, "iam.user.not_found", "user not found")
	// ErrUserAlreadyExists menandakan pengguna dengan pengenal tersebut sudah terdaftar.
	ErrUserAlreadyExists = newSentinel(KindAlreadyExists, "iam.user.already_exists", "user already exists")
	// ErrUserInactive menandakan akun pengguna belum aktif atau tidak dapat digunakan untuk autentikasi.
	ErrUserInactive = newSentinel(KindForbidden, "iam.user.inactive", "user is not active")
	// ErrUserSuspended menandakan akun pengguna sedang ditangguhkan sementara.
	ErrUserSuspended = newSentinel(KindForbidden, "iam.user.suspended", "user account is suspended")
	// ErrUserDisabled menandakan akun pengguna dinonaktifkan secara permanen atau administratif.
	ErrUserDisabled = newSentinel(KindForbidden, "iam.user.disabled", "user account is disabled")
	// ErrUserDeleted menandakan akun pengguna sudah dihapus secara lunak.
	ErrUserDeleted = newSentinel(KindConflict, "iam.user.deleted", "user account has been deleted")

	// ErrInvalidDisplayName menandakan nama tampilan kosong atau melebihi batas panjang.
	ErrInvalidDisplayName = newSentinel(KindInvalidInput, "iam.user.invalid_display_name", "invalid display name")
	// ErrInvalidAvatarURL menandakan format atau panjang URL avatar tidak valid.
	ErrInvalidAvatarURL = newSentinel(KindInvalidInput, "iam.user.invalid_avatar_url", "invalid avatar url")
	// ErrInvalidLocale menandakan nilai locale melebihi batas atau tidak valid.
	ErrInvalidLocale = newSentinel(KindInvalidInput, "iam.user.invalid_locale", "invalid locale")
	// ErrInvalidTimezone menandakan nilai zona waktu tidak valid atau melebihi batas panjang.
	ErrInvalidTimezone = newSentinel(KindInvalidInput, "iam.user.invalid_timezone", "invalid timezone")
	// ErrInvalidUserKind menandakan tipe user_kind tidak dikenali oleh skema IAM.
	ErrInvalidUserKind = newSentinel(KindInvalidInput, "iam.user.invalid_kind", "invalid user kind")
	// ErrInvalidUserStatus menandakan status pengguna tidak dikenali oleh skema IAM.
	ErrInvalidUserStatus = newSentinel(KindInvalidInput, "iam.user.invalid_status", "invalid user status")

	// ErrIdentityNotFound menandakan pengenal login pengguna tidak ditemukan.
	ErrIdentityNotFound = newSentinel(KindNotFound, "iam.identity.not_found", "user identity not found")
	// ErrIdentityAlreadyExists menandakan identitas login dengan kind dan value tersebut sudah terdaftar.
	ErrIdentityAlreadyExists = newSentinel(KindAlreadyExists, "iam.identity.already_exists", "user identity already exists")
	// ErrIdentityNotVerified menandakan identitas login belum melewati proses verifikasi.
	ErrIdentityNotVerified = newSentinel(KindForbidden, "iam.identity.not_verified", "user identity is not verified")
	// ErrInvalidIdentityKind menandakan jenis identitas login tidak dikenali.
	ErrInvalidIdentityKind = newSentinel(KindInvalidInput, "iam.identity.invalid_kind", "invalid identity kind")
	// ErrInvalidIdentityValue menandakan nilai identitas login kosong atau melebihi panjang maksimum.
	ErrInvalidIdentityValue = newSentinel(KindInvalidInput, "iam.identity.invalid_value", "invalid identity value")
	// ErrInvalidEmail menandakan format alamat email tidak sesuai standar.
	ErrInvalidEmail = newSentinel(KindInvalidInput, "iam.identity.invalid_email", "invalid email format")
	// ErrInvalidPhone menandakan format nomor telepon tidak sesuai E.164 atau terlalu pendek/panjang.
	ErrInvalidPhone = newSentinel(KindInvalidInput, "iam.identity.invalid_phone", "invalid phone format")
	// ErrInvalidUsername menandakan format username tidak memenuhi aturan karakter atau panjang.
	ErrInvalidUsername = newSentinel(KindInvalidInput, "iam.identity.invalid_username", "invalid username format")

	// ErrCredentialNotFound menandakan kredensial kata sandi pengguna belum terdaftar.
	ErrCredentialNotFound = newSentinel(KindNotFound, "iam.credential.not_found", "user credential not found")
	// ErrCredentialAlreadyExists menandakan kredensial kata sandi pengguna sudah ada.
	ErrCredentialAlreadyExists = newSentinel(KindAlreadyExists, "iam.credential.already_exists", "user credential already exists")
	// ErrInvalidCredentials menandakan kombinasi identitas atau kata sandi tidak valid.
	ErrInvalidCredentials = newSentinel(KindUnauthorized, "iam.credential.invalid", "invalid credentials")
	// ErrAccountLocked menandakan akun terkunci sementara akibat kegagalan login berulang.
	ErrAccountLocked = newSentinel(KindRateLimited, "iam.credential.locked", "account temporarily locked due to failed login attempts")
	// ErrPasswordTooShort menandakan panjang kata sandi kurang dari batas minimum kebijakan.
	ErrPasswordTooShort = newSentinel(KindInvalidInput, "iam.credential.password_too_short", "password is too short")
	// ErrPasswordTooLong menandakan panjang kata sandi melebihi batas maksimum kebijakan.
	ErrPasswordTooLong = newSentinel(KindInvalidInput, "iam.credential.password_too_long", "password exceeds maximum length")
	// ErrPasswordTooWeak menandakan kata sandi tidak memenuhi persyaratan variasi karakter.
	ErrPasswordTooWeak = newSentinel(KindInvalidInput, "iam.credential.password_too_weak", "password does not meet complexity requirements")
	// ErrPasswordContainsIdentity menandakan kata sandi mengandung pengenal akun sehingga mudah ditebak.
	ErrPasswordContainsIdentity = newSentinel(KindInvalidInput, "iam.credential.password_contains_identity", "password must not contain identity terms")
	// ErrPasswordRotationRequired menandakan pengguna wajib mengganti kata sandi sebelum melanjutkan akses penuh.
	ErrPasswordRotationRequired = newSentinel(KindForbidden, "iam.credential.password_rotation_required", "password change required")

	// ErrTOTPNotFound menandakan faktor TOTP Google Authenticator pengguna belum dibuat.
	ErrTOTPNotFound = newSentinel(KindNotFound, "iam.totp.not_found", "totp factor not found")
	// ErrTOTPAlreadyExists menandakan pengguna sudah memiliki faktor TOTP terdaftar.
	ErrTOTPAlreadyExists = newSentinel(KindAlreadyExists, "iam.totp.already_exists", "totp factor already exists")
	// ErrTOTPNotConfirmed menandakan faktor TOTP belum dikonfirmasi kepemilikannya oleh pengguna.
	ErrTOTPNotConfirmed = newSentinel(KindConflict, "iam.totp.not_confirmed", "totp factor is not confirmed")
	// ErrTOTPAlreadyConfirmed menandakan faktor TOTP sudah terkonfirmasi sebelumnya.
	ErrTOTPAlreadyConfirmed = newSentinel(KindConflict, "iam.totp.already_confirmed", "totp factor already confirmed")
	// ErrInvalidTOTPCode menandakan kode OTP 6 digit tidak sah.
	ErrInvalidTOTPCode = newSentinel(KindUnauthorized, "iam.totp.invalid_code", "invalid totp verification code")
	// ErrTOTPReplayDetected menandakan time step TOTP yang sama dicoba digunakan ulang.
	ErrTOTPReplayDetected = newSentinel(KindConflict, "iam.totp.replay_detected", "totp step already used")

	// ErrRecoveryCodeNotFound menandakan kode pemulihan 2FA tidak ditemukan untuk pengguna.
	ErrRecoveryCodeNotFound = newSentinel(KindNotFound, "iam.recovery_code.not_found", "recovery code not found")
	// ErrRecoveryCodeAlreadyUsed menandakan kode pemulihan 2FA sudah pernah dipakai.
	ErrRecoveryCodeAlreadyUsed = newSentinel(KindConflict, "iam.recovery_code.already_used", "recovery code already used")
	// ErrInvalidRecoveryCode menandakan format atau nilai kode pemulihan tidak cocok.
	ErrInvalidRecoveryCode = newSentinel(KindUnauthorized, "iam.recovery_code.invalid", "invalid recovery code")

	// ErrTokenNotFound menandakan token sekali pakai tidak ditemukan.
	ErrTokenNotFound = newSentinel(KindNotFound, "iam.token.not_found", "token not found")
	// ErrTokenExpired menandakan token sekali pakai sudah melewati batas waktu kedaluwarsa.
	ErrTokenExpired = newSentinel(KindUnauthorized, "iam.token.expired", "token expired")
	// ErrTokenRevoked menandakan token sekali pakai telah dibatalkan sebelum digunakan.
	ErrTokenRevoked = newSentinel(KindUnauthorized, "iam.token.revoked", "token revoked")
	// ErrTokenAlreadyUsed menandakan token sekali pakai sudah digunakan sebelumnya.
	ErrTokenAlreadyUsed = newSentinel(KindConflict, "iam.token.already_used", "token already used")
	// ErrTokenInvalidPurpose menandakan tujuan token tidak sesuai dengan alur yang diminta.
	ErrTokenInvalidPurpose = newSentinel(KindInvalidInput, "iam.token.invalid_purpose", "invalid token purpose")
	// ErrInvalidTokenStatus menandakan status token tidak dikenali oleh skema IAM.
	ErrInvalidTokenStatus = newSentinel(KindInvalidInput, "iam.token.invalid_status", "invalid token status")
	// ErrInvalidTokenDigest menandakan digest token kosong atau tidak valid.
	ErrInvalidTokenDigest = newSentinel(KindInvalidInput, "iam.token.invalid_digest", "invalid token digest")
	// ErrPendingEmailNotFound menandakan pengguna tidak memiliki email tertunda yang sedang menunggu verifikasi.
	ErrPendingEmailNotFound = newSentinel(KindNotFound, "iam.identity.pending_email_not_found", "pending email identity not found")
	// ErrEmailAlreadyVerified menandakan tidak ada email tertunda karena alamat email pengguna sudah terverifikasi.
	ErrEmailAlreadyVerified = newSentinel(KindConflict, "iam.identity.email_already_verified", "your email has been verified")
	// ErrSameEmail menandakan alamat email baru sama dengan alamat email aktif saat ini.
	ErrSameEmail = newSentinel(KindConflict, "iam.identity.same_email", "new email must differ from current active email")
	// ErrSameUsername menandakan username baru sama dengan username aktif saat ini.
	ErrSameUsername = newSentinel(KindConflict, "iam.identity.same_username", "new username must differ from current username")
	// ErrIdentityLimitExceeded menandakan batas jumlah identitas per jenis (1 email aktif dan 1 username) sudah tercapai.
	ErrIdentityLimitExceeded = newSentinel(KindConflict, "iam.identity.limit_exceeded", "user identity limit exceeded for this kind")
)
