// Package authz menyediakan model AccessSnapshot bersama dan evaluator otorisasi fail-closed di gateway dan layanan.
package authz

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	// DefaultSnapshotMaxAge menetapkan batas kesegaran maksimum AccessSnapshot di cache (30 detik).
	DefaultSnapshotMaxAge = 30 * time.Second
	// ScopeGlobal menandakan cakupan akses berlaku global di seluruh perusahaan.
	ScopeGlobal = "global"
	// ScopeSelf menandakan cakupan akses terbatas hanya pada akun milik principal sendiri.
	ScopeSelf = "self"
	// StatusActive menandakan status user aktif dan dapat diotorisasi.
	StatusActive = "active"

	// Kode izin kanonik domain IAM untuk akses admin dan operasi mandiri (self-service).
	PermDashboardView      = "iam.dashboard.view"
	PermUsersRead          = "iam.users.read"
	PermUsersWrite         = "iam.users.write"
	PermUsersReadSelf      = "iam.users.readself"
	PermUsersUpdateSelf    = "iam.users.updateself"
	PermUsersDeleteSelf    = "iam.users.deleteself"
	PermRolesRead          = "iam.roles.read"
	PermRolesWrite         = "iam.roles.write"
	PermGroupsRead         = "iam.groups.read"
	PermGroupsWrite        = "iam.groups.write"
	PermPermissionsRead    = "iam.permissions.read"
	PermSessionsRead       = "iam.sessions.read"
	PermSessionsReadSelf   = "iam.sessions.readself"
	PermSessionsRevoke     = "iam.sessions.revoke"
	PermSessionsDeleteSelf = "iam.sessions.deleteself"
	PermPayslipsReadSelf   = "human_resource.payslips.read_self"
)

// IsSelfServicePermissionCode memeriksa apakah kode izin merupakan aksi mandiri (*self) yang terikat kepemilikan principal sendiri.
func IsSelfServicePermissionCode(permCode string) bool {
	norm := normalizeCode(permCode)
	return norm == PermPayslipsReadSelf ||
		strings.HasSuffix(norm, ".readself") ||
		strings.HasSuffix(norm, ".updateself") ||
		strings.HasSuffix(norm, ".deleteself")
}

// ScopeGrant merepresentasikan batasan lingkup izin efektif di Redis snapshot.
type ScopeGrant struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

// Matches memeriksa apakah ScopeGrant yang dimiliki memenuhi target scope permintaan.
func (s ScopeGrant) Matches(required ScopeGrant) bool {
	grantType := normalizeCode(s.Type)
	reqType := normalizeCode(required.Type)
	if grantType == "" || reqType == "" {
		return false
	}
	if grantType == ScopeGlobal {
		return s.ID == ""
	}
	if reqType == ScopeGlobal {
		return false
	}
	return grantType == reqType &&
		strings.TrimSpace(s.ID) != "" &&
		strings.EqualFold(strings.TrimSpace(s.ID), strings.TrimSpace(required.ID))
}

// RoleGrant merepresentasikan ringkasan peran efektif user untuk evaluasi prioritas dan rute default.
type RoleGrant struct {
	ID                  string `json:"id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	Priority            int32  `json:"priority"`
	DefaultRedirectPath string `json:"default_redirect_path,omitempty"`
}

// PermissionGrant merepresentasikan satu kode izin efektif beserta cakupan dan bendera sensitivitasnya.
type PermissionGrant struct {
	Code        string     `json:"code"`
	IsSensitive bool       `json:"is_sensitive,omitempty"`
	Scope       ScopeGrant `json:"scope"`
}

// AccessSnapshot menyimpan proyeksi otorisasi terpusat milik principal yang dipakai gateway tanpa hit DB per permintaan.
type AccessSnapshot struct {
	PrincipalID          string            `json:"principal_id"`
	UserKind             string            `json:"user_kind"`
	Status               string            `json:"status"`
	MFAEnabled           bool              `json:"mfa_enabled"`
	AuthEpoch            int64             `json:"auth_epoch"`
	PermRevision         int64             `json:"perm_revision"`
	RecoveryEpoch        int64             `json:"recovery_epoch"`
	Roles                []RoleGrant       `json:"roles"`
	Permissions          []PermissionGrant `json:"permissions"`
	CanAccessIdentity    bool              `json:"can_access_identity"`
	DefaultRedirectPath  string            `json:"default_redirect_path"`
	AllowedRedirectPaths []string          `json:"allowed_redirect_paths"`
	SyncedAtUnix         int64             `json:"synced_at_unix"`
	FreshUntilUnix       int64             `json:"fresh_until_unix"`
}

// MarshalSnapshot menyandikan AccessSnapshot menjadi payload JSON ringkas.
func MarshalSnapshot(snap *AccessSnapshot) ([]byte, error) {
	return json.Marshal(snap)
}

// UnmarshalSnapshot mengurai payload JSON menjadi AccessSnapshot.
func UnmarshalSnapshot(raw []byte) (*AccessSnapshot, error) {
	var snap AccessSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// IsFresh memverifikasi secara fail-closed bahwa snapshot masih aktif, tidak kedaluwarsa (>30s), dan sinkron.
func (s *AccessSnapshot) IsFresh(now time.Time) bool {
	if s == nil || s.PrincipalID == "" || normalizeCode(s.Status) != StatusActive {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}
	nowUnix := now.Unix()
	if s.SyncedAtUnix <= 0 || s.FreshUntilUnix <= 0 {
		return false
	}
	if nowUnix > s.FreshUntilUnix {
		return false
	}
	return s.FreshUntilUnix-s.SyncedAtUnix <= int64(DefaultSnapshotMaxAge.Seconds())+1
}

// HasRole memeriksa apakah user memiliki kode peran tertentu dalam snapshot aktif.
func (s *AccessSnapshot) HasRole(roleCode string) bool {
	if s == nil || normalizeCode(s.Status) != StatusActive {
		return false
	}
	target := normalizeCode(roleCode)
	if target == "" {
		return false
	}
	for _, r := range s.Roles {
		if normalizeCode(r.Code) == target {
			return true
		}
	}
	return false
}

// HasPermission memeriksa apakah user memiliki izin pada cakupan global.
func (s *AccessSnapshot) HasPermission(permCode string) (PermissionGrant, bool) {
	return s.HasScopedPermission(permCode, ScopeGrant{Type: ScopeGlobal})
}

// HasScopedPermission memeriksa apakah user memiliki izin yang memenuhi scope permintaan secara fail-closed.
func (s *AccessSnapshot) HasScopedPermission(permCode string, required ScopeGrant) (PermissionGrant, bool) {
	if s == nil || normalizeCode(s.Status) != StatusActive {
		return PermissionGrant{}, false
	}
	target := normalizeCode(permCode)
	if target == "" {
		return PermissionGrant{}, false
	}
	if target == PermPayslipsReadSelf && strings.TrimSpace(s.PrincipalID) == "" {
		return PermissionGrant{}, false
	}
	if IsSelfServicePermissionCode(target) {
		// Izin self-service tidak pernah boleh lolos pada evaluasi global atau scope self milik principal lain.
		if normalizeCode(required.Type) != ScopeSelf || strings.TrimSpace(required.ID) == "" {
			return PermissionGrant{}, false
		}
		if strings.TrimSpace(s.PrincipalID) != "" && !strings.EqualFold(strings.TrimSpace(required.ID), strings.TrimSpace(s.PrincipalID)) {
			return PermissionGrant{}, false
		}
	}
	for _, p := range s.Permissions {
		if target == PermPayslipsReadSelf && normalizeCode(p.Scope.Type) != ScopeSelf {
			continue
		}
		if normalizeCode(p.Code) == target && p.Scope.Matches(required) {
			return p, true
		}
	}
	return PermissionGrant{}, false
}

// HasAnyPermission memeriksa apakah minimal satu dari kode izin terpenuhi pada scope global.
func (s *AccessSnapshot) HasAnyPermission(permCodes ...string) bool {
	for _, code := range permCodes {
		if _, ok := s.HasPermission(code); ok {
			return true
		}
	}
	return false
}

// CanAccessWorkspaceIdentity memutuskan secara fail-closed apakah principal berhak masuk dashboard /identity.
func (s *AccessSnapshot) CanAccessWorkspaceIdentity() bool {
	if s == nil || normalizeCode(s.Status) != StatusActive {
		return false
	}
	return s.CanAccessIdentity
}

// ResolveDefaultLocation mengembalikan rute default teraman ("/" jika kosong).
func (s *AccessSnapshot) ResolveDefaultLocation() string {
	if s == nil || strings.TrimSpace(s.DefaultRedirectPath) == "" {
		return "/"
	}
	return strings.TrimSpace(s.DefaultRedirectPath)
}

func normalizeCode(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
