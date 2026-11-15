package authsession

import "github.com/integrione/erp-retail/backend/shared/authz"

// SessionTokenRecord merepresentasikan proyeksi inti sesi di Redis (hanya ditulis oleh iam_writer).
type SessionTokenRecord struct {
	SessionID        string `json:"session_id"`
	UserID           string `json:"user_id"`
	TokenHashHex     string `json:"token_hash_hex"`
	CSRFDigestHex    string `json:"csrf_digest_hex"`
	CSRFKeyVersion   int    `json:"csrf_key_version"`
	UserKind         string `json:"user_kind"`
	Generation       int64  `json:"generation"`
	AuthEpoch        int64  `json:"auth_epoch"`
	IssuedAtUnix     int64  `json:"issued_at_unix"`
	AbsExpiresAtUnix int64  `json:"abs_expires_at_unix"`
	MFAAuthenticated bool   `json:"mfa_authenticated"`
	Revoked          bool   `json:"revoked"`
}

// SessionActivityRecord merepresentasikan jendela idle timeout sesi (dapat diperpanjang Gateway lewat Lua XX).
type SessionActivityRecord struct {
	SessionID       string `json:"session_id"`
	UserID          string `json:"user_id"`
	Generation      int64  `json:"generation"`
	AuthEpoch       int64  `json:"auth_epoch"`
	LastSeenUnix    int64  `json:"last_seen_unix"`
	IdleExpiresUnix int64  `json:"idle_expires_unix"`
}

// UserSecurityRecord menyimpan status keamanan monotonik user untuk invalidasi instan O(1).
type UserSecurityRecord struct {
	UserID          string `json:"user_id"`
	Status          string `json:"status"`
	LockedUntilUnix int64  `json:"locked_until_unix"`
	AuthEpoch       int64  `json:"auth_epoch"`
	PermRevision    int64  `json:"perm_revision"`
	RecoveryEpoch   int64  `json:"recovery_epoch"`
	UpdatedAtUnix   int64  `json:"updated_at_unix"`
}

// SecurityGateRecord menyatakan apakah proyeksi Redis IAM sudah dipulihkan penuh dan aman dipercaya Gateway.
type SecurityGateRecord struct {
	Open          bool   `json:"open"`
	RecoveryEpoch int64  `json:"recovery_epoch"`
	UpdatedAtUnix int64  `json:"updated_at_unix"`
	Source        string `json:"source"`
}

// PreAuthRecord menyimpan pasangan handle __Host-integrione_session pra-login dan digest CSRF SHA-256.
type PreAuthRecord struct {
	TokenHashHex  string `json:"token_hash_hex"`
	CSRFDigestHex string `json:"csrf_digest_hex"`
	CreatedAtUnix int64  `json:"created_at_unix"`
	ExpiresAtUnix int64  `json:"expires_at_unix"`
}

// SessionValidationContext merangkum hasil pembacaan fail-closed seluruh state Redis untuk satu permintaan.
type SessionValidationContext struct {
	Gate     *SecurityGateRecord
	Token    *SessionTokenRecord
	Activity *SessionActivityRecord
	Security *UserSecurityRecord
	Snapshot *authz.AccessSnapshot
}
