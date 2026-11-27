package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// MaxRedirectPriority membatasi prioritas redirect pada rentang smallint PostgreSQL (0..32767).
const MaxRedirectPriority = 32767

// Role merepresentasikan entitas peran RBAC dan metadata rute dashboard bawaan pada tabel roles.
type Role struct {
	ID                  string
	Code                string
	Name                string
	Description         string
	ManagementDomain    string
	DefaultRedirectPath string
	RedirectPriority    int
	IsSystem            bool
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewRole membentuk peran baru sekaligus menegakkan kode unik, path redirect aman, dan prioritas non-negatif.
func NewRole(
	id, code, name, description, managementDomain, defaultRedirectPath string,
	redirectPriority int,
	isSystem bool,
	now time.Time,
) (*Role, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	normalizedCode, err := ValidateEntityCode(code, MaxRoleCodeLength, iamerrors.ErrInvalidRoleCode)
	if err != nil {
		return nil, err
	}
	trimmedName, trimmedDomain, trimmedRedirect, err := validateRoleMetadata(name, managementDomain, defaultRedirectPath, redirectPriority)
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &Role{
		ID:                  strings.TrimSpace(id),
		Code:                normalizedCode,
		Name:                trimmedName,
		Description:         strings.TrimSpace(description),
		ManagementDomain:    trimmedDomain,
		DefaultRedirectPath: trimmedRedirect,
		RedirectPriority:    redirectPriority,
		IsSystem:            isSystem,
		IsActive:            true,
		CreatedAt:           ts,
		UpdatedAt:           ts,
	}, nil
}

// UpdateDetails memperbarui metadata peran termasuk rute redirect bawaan dengan validasi yang sama.
func (r *Role) UpdateDetails(
	name, description, managementDomain, defaultRedirectPath string,
	redirectPriority int,
	now time.Time,
) error {
	if r == nil {
		return iamerrors.ErrRoleNotFound
	}
	if r.IsSystem {
		return iamerrors.ErrSystemRoleImmutable
	}
	trimmedName, trimmedDomain, trimmedRedirect, err := validateRoleMetadata(name, managementDomain, defaultRedirectPath, redirectPriority)
	if err != nil {
		return err
	}

	r.Name = trimmedName
	r.Description = strings.TrimSpace(description)
	r.ManagementDomain = trimmedDomain
	r.DefaultRedirectPath = trimmedRedirect
	r.RedirectPriority = redirectPriority
	r.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// ApplyBootstrapSystemMetadata menyinkronkan metadata peran sistem dari manifest katalog resmi saat proses bootstrap internal.
func (r *Role) ApplyBootstrapSystemMetadata(
	name, description, managementDomain, defaultRedirectPath string,
	redirectPriority int,
	now time.Time,
) (bool, error) {
	if r == nil {
		return false, iamerrors.ErrRoleNotFound
	}
	if !r.IsSystem {
		return false, iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.bootstrap.system_role_expected"), "bootstrap system metadata update requires a system role")
	}
	trimmedName, trimmedDomain, trimmedRedirect, err := validateRoleMetadata(name, managementDomain, defaultRedirectPath, redirectPriority)
	if err != nil {
		return false, err
	}
	trimmedDesc := strings.TrimSpace(description)
	changed := r.Name != trimmedName ||
		r.Description != trimmedDesc ||
		r.ManagementDomain != trimmedDomain ||
		r.DefaultRedirectPath != trimmedRedirect ||
		r.RedirectPriority != redirectPriority ||
		!r.IsActive
	if !changed {
		return false, nil
	}

	r.Name = trimmedName
	r.Description = trimmedDesc
	r.ManagementDomain = trimmedDomain
	r.DefaultRedirectPath = trimmedRedirect
	r.RedirectPriority = redirectPriority
	r.IsActive = true
	r.UpdatedAt = normalizeLocalTime(now)
	return true, nil
}

// IsSuperAdminRoleCode memeriksa apakah kode peran merupakan peran administrator super manusia.
func IsSuperAdminRoleCode(code string) bool {
	normalized := NormalizeCode(code)
	return normalized == "superadmin" || normalized == "super_admin"
}

// IsAutomationRoleCode memeriksa apakah kode peran dikhususkan bagi akun layanan otomasi non-interaktif (`*_automation`).
func IsAutomationRoleCode(code string) bool {
	normalized := NormalizeCode(code)
	return strings.HasSuffix(normalized, "_automation")
}

// IsRoleCompatibleWithUserKind memastikan peran superadmin hanya untuk pegawai manusia (employee) dan peran otomasi hanya untuk akun system.
func IsRoleCompatibleWithUserKind(roleCode string, kind UserKind) bool {
	if kind == UserKindSystem {
		return IsAutomationRoleCode(roleCode)
	}
	if IsAutomationRoleCode(roleCode) {
		return false
	}
	if IsSuperAdminRoleCode(roleCode) {
		return kind == UserKindEmployee
	}
	return kind.IsValid()
}

func validateRoleMetadata(name, managementDomain, defaultRedirectPath string, redirectPriority int) (string, string, string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || utf8.RuneCountInString(trimmedName) > MaxRoleNameLength {
		return "", "", "", iamerrors.NewValidation("name", iamerrors.Code("iam.role.invalid_name"), "role name must be non-empty and within length bounds")
	}
	trimmedDomain := NormalizeCode(managementDomain)
	if utf8.RuneCountInString(trimmedDomain) > MaxManagementDomainLength {
		return "", "", "", iamerrors.NewValidation("management_domain", iamerrors.Code("iam.role.invalid_domain"), "management domain exceeds length bounds")
	}
	trimmedRedirect := strings.TrimSpace(defaultRedirectPath)
	if trimmedRedirect != "" {
		if err := ValidateInternalRedirectPath(trimmedRedirect); err != nil {
			return "", "", "", err
		}
	}
	if redirectPriority < 0 || redirectPriority > MaxRedirectPriority {
		return "", "", "", iamerrors.ErrInvalidRedirectPriority
	}
	return trimmedName, trimmedDomain, trimmedRedirect, nil
}

// Activate mengaktifkan kembali peran agar hak aksesnya dapat dievaluasi.
func (r *Role) Activate(now time.Time) {
	if r == nil {
		return
	}
	r.IsActive = true
	r.UpdatedAt = normalizeLocalTime(now)
}

// Deactivate menonaktifkan peran non-sistem sambil melindungi peran bawaan sistem agar tidak dimatikan.
func (r *Role) Deactivate(now time.Time) error {
	if r == nil {
		return iamerrors.ErrRoleNotFound
	}
	if r.IsSystem {
		return iamerrors.ErrSystemRoleImmutable
	}
	r.IsActive = false
	r.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Permission merepresentasikan izin atomik berbasis resource dan action pada tabel permissions.
type Permission struct {
	ID          string
	Code        string
	Resource    string
	Action      string
	Description string
	IsSensitive bool
	CreatedAt   time.Time
}

// NewPermission membentuk izin atomik baru dengan normalisasi code, resource, dan action.
func NewPermission(id, code, resource, action, description string, isSensitive bool, now time.Time) (*Permission, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	normalizedCode, err := ValidatePermissionCode(code)
	if err != nil {
		return nil, err
	}
	normalizedResource := NormalizeCode(resource)
	if normalizedResource == "" || utf8.RuneCountInString(normalizedResource) > MaxPermissionResourceLength {
		return nil, iamerrors.NewValidation("resource", iamerrors.Code("iam.permission.invalid_resource"), "permission resource must be non-empty and within length bounds")
	}
	normalizedAction := NormalizeCode(action)
	if normalizedAction == "" || utf8.RuneCountInString(normalizedAction) > MaxPermissionActionLength {
		return nil, iamerrors.NewValidation("action", iamerrors.Code("iam.permission.invalid_action"), "permission action must be non-empty and within length bounds")
	}
	if !strings.Contains(normalizedResource, ".") || !codePattern.MatchString(normalizedResource) || !codePattern.MatchString(normalizedAction) || strings.Contains(normalizedAction, ".") || normalizedCode != normalizedResource+"."+normalizedAction {
		return nil, iamerrors.NewValidation("code", iamerrors.Code("iam.permission.inconsistent_code"), "permission code must match resource and action")
	}

	return &Permission{
		ID:          strings.TrimSpace(id),
		Code:        normalizedCode,
		Resource:    normalizedResource,
		Action:      normalizedAction,
		Description: strings.TrimSpace(description),
		IsSensitive: isSensitive,
		CreatedAt:   normalizeLocalTime(now),
	}, nil
}

// RolePermission merepresentasikan relasi many-to-many pada tabel role_permissions.
type RolePermission struct {
	RoleID       string
	PermissionID string
	CreatedAt    time.Time
}

// NewRolePermission membentuk ikatan izin ke peran dengan validasi kedua UUID.
func NewRolePermission(roleID, permissionID string, now time.Time) (*RolePermission, error) {
	if err := ValidateUUID("role_id", roleID); err != nil {
		return nil, err
	}
	if err := ValidateUUID("permission_id", permissionID); err != nil {
		return nil, err
	}
	return &RolePermission{
		RoleID:       strings.TrimSpace(roleID),
		PermissionID: strings.TrimSpace(permissionID),
		CreatedAt:    normalizeLocalTime(now),
	}, nil
}
