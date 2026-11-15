// Package authsession menyatukan kontrak keyspace Redis, konfigurasi klien TLS/ACL, record runtime, dan skrip Lua atomik untuk IAM dan Gateway.
package authsession

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Keyspace membungkus prefiks kunci Redis terisolasi menurut environment (mis. development, test_<id>, production).
type Keyspace struct {
	env string
}

// NewKeyspace membangun Keyspace dengan normalisasi aman agar bebas karakter pemisah ':' yang merusak struktur key.
func NewKeyspace(env string) Keyspace {
	cleaned := strings.ToLower(strings.TrimSpace(env))
	cleaned = strings.ReplaceAll(cleaned, ":", "_")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")
	cleaned = strings.ReplaceAll(cleaned, "*", "_")
	if cleaned == "" {
		cleaned = "development"
	}
	return Keyspace{env: cleaned}
}

// Env mengembalikan nama environment ternormalisasi pada Keyspace ini.
func (k Keyspace) Env() string {
	if k.env == "" {
		return "development"
	}
	return k.env
}

// Prefix mengembalikan prefiks root seluruh kunci dalam namespace environment ini.
func (k Keyspace) Prefix() string {
	return fmt.Sprintf("ig:%s:", k.Env())
}

// SessionTokenKey mengembalikan kunci metadata inti sesi berdasarkan digest SHA-256 hex token cookie.
func (k Keyspace) SessionTokenKey(tokenHashHex string) string {
	return fmt.Sprintf("ig:%s:iam:sess:tok:%s", k.Env(), strings.ToLower(strings.TrimSpace(tokenHashHex)))
}

// SessionActivityKey mengembalikan kunci pelacak idle timeout sesi berdasarkan digest SHA-256 hex token cookie.
func (k Keyspace) SessionActivityKey(tokenHashHex string) string {
	return fmt.Sprintf("ig:%s:iam:sess:act:%s", k.Env(), strings.ToLower(strings.TrimSpace(tokenHashHex)))
}

// AuthzSnapshotKey mengembalikan kunci AccessSnapshot JSON milik principal tertentu.
func (k Keyspace) AuthzSnapshotKey(userID string) string {
	return fmt.Sprintf("ig:%s:iam:authz:%s", k.Env(), strings.TrimSpace(userID))
}

// UserSecurityKey mengembalikan kunci status keamanan monotonik (auth_epoch, perm_revision, status, locked_until) user.
func (k Keyspace) UserSecurityKey(userID string) string {
	return fmt.Sprintf("ig:%s:iam:user:sec:%s", k.Env(), strings.TrimSpace(userID))
}

// GateKey mengembalikan kunci gerbang keamanan global IAM yang wajib terbuka agar Gateway melayani otentikasi cache.
func (k Keyspace) GateKey() string {
	return fmt.Sprintf("ig:%s:iam:ctrl:gate", k.Env())
}

// RecoveryEpochKey mengembalikan kunci epoch pemulihan global IAM monotonik.
func (k Keyspace) RecoveryEpochKey() string {
	return fmt.Sprintf("ig:%s:iam:ctrl:recovery_epoch", k.Env())
}

// PreAuthKey mengembalikan kunci state pra-autentikasi (__Host-integrione_session sebelum login) milik Gateway.
func (k Keyspace) PreAuthKey(cookieHashHex string) string {
	return fmt.Sprintf("ig:%s:gw:preauth:%s", k.Env(), strings.ToLower(strings.TrimSpace(cookieHashHex)))
}

// RateLimitKey mengembalikan kunci pembatas laju (rate limit) milik Gateway.
func (k Keyspace) RateLimitKey(scope, subjectHashHex string) string {
	cleanScope := strings.ToLower(strings.TrimSpace(scope))
	cleanScope = strings.ReplaceAll(cleanScope, ":", "_")
	if cleanScope == "" {
		cleanScope = "default"
	}
	return fmt.Sprintf("ig:%s:gw:rl:%s:%s", k.Env(), cleanScope, strings.ToLower(strings.TrimSpace(subjectHashHex)))
}

// HashOpaqueToken menghitung digest SHA-256 hex huruf kecil atas token buram atau CSRF plaintext.
func HashOpaqueToken(plaintext string) string {
	trimmed := strings.TrimSpace(plaintext)
	if trimmed == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(sum[:])
}
