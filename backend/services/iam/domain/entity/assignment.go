package entity

import (
	"slices"
	"strings"
	"time"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// Scope merepresentasikan batasan lingkup akses yang menegakkan aturan global vs non-global berserta ID.
type Scope struct {
	Type ScopeType
	ID   *string
}

// GlobalScope mengembalikan value object Scope global tanpa ID lingkup.
func GlobalScope() Scope {
	return Scope{Type: ScopeTypeGlobal}
}

// NewScope membangun Scope tervalidasi sesuai constraint ck_user_roles_scope dan ck_group_roles_scope.
func NewScope(scopeType ScopeType, scopeID *string) (Scope, error) {
	if !scopeType.IsValid() {
		return Scope{}, iamerrors.ErrInvalidScope
	}
	if scopeType == ScopeTypeGlobal {
		if scopeID != nil && strings.TrimSpace(*scopeID) != "" {
			return Scope{}, iamerrors.ErrInvalidScope
		}
		return GlobalScope(), nil
	}
	if scopeID == nil || !IsValidUUID(*scopeID) {
		return Scope{}, iamerrors.ErrInvalidScope
	}
	normalized := strings.TrimSpace(*scopeID)
	return Scope{
		Type: scopeType,
		ID:   &normalized,
	}, nil
}

// IsValid memverifikasi apakah Scope memenuhi aturan skema tanpa modifikasi manual yang melanggar invarian.
func (s Scope) IsValid() bool {
	if !s.Type.IsValid() {
		return false
	}
	if s.Type == ScopeTypeGlobal {
		return s.ID == nil
	}
	return s.ID != nil && IsValidUUID(*s.ID)
}

// Clone menyalin pointer ID lingkup secara aman.
func (s Scope) Clone() Scope {
	if s.ID == nil {
		return Scope{Type: s.Type}
	}
	idCopy := *s.ID
	return Scope{Type: s.Type, ID: &idCopy}
}

// Matches mengevaluasi secara fail-closed apakah scope yang dimiliki memenuhi scope target yang diminta.
func (s Scope) Matches(required Scope) bool {
	if !s.IsValid() || !required.IsValid() {
		return false
	}
	if s.Type == ScopeTypeGlobal {
		return true
	}
	if required.Type == ScopeTypeGlobal {
		return false
	}
	return s.Type == required.Type && strings.EqualFold(strings.TrimSpace(*s.ID), strings.TrimSpace(*required.ID))
}

type roleAssignmentCore struct {
	ID        string
	RoleID    string
	Scope     Scope
	GrantedBy *string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

func newRoleAssignmentCore(
	id, subjectField, subjectID, roleID string,
	scope Scope,
	grantedBy *string,
	expiresAt *time.Time,
	now time.Time,
) (roleAssignmentCore, error) {
	if err := ValidateUUID("id", id); err != nil {
		return roleAssignmentCore{}, err
	}
	if err := ValidateUUID(subjectField, subjectID); err != nil {
		return roleAssignmentCore{}, err
	}
	if err := ValidateUUID("role_id", roleID); err != nil {
		return roleAssignmentCore{}, err
	}
	if !scope.IsValid() {
		return roleAssignmentCore{}, iamerrors.ErrInvalidScope
	}
	var grantedByCopy *string
	if grantedBy != nil {
		if err := ValidateUUID("granted_by", *grantedBy); err != nil {
			return roleAssignmentCore{}, err
		}
		g := strings.TrimSpace(*grantedBy)
		grantedByCopy = &g
	}
	createdLocal := normalizeLocalTime(now)
	var expiresCopy *time.Time
	if expiresAt != nil {
		expLocal := normalizeLocalTime(*expiresAt)
		if !expLocal.After(createdLocal) {
			return roleAssignmentCore{}, iamerrors.NewValidation("expires_at", iamerrors.Code("iam.assignment.invalid_expiry"), "expires_at must be strictly after created_at")
		}
		expiresCopy = &expLocal
	}
	return roleAssignmentCore{
		ID:        strings.TrimSpace(id),
		RoleID:    strings.TrimSpace(roleID),
		Scope:     scope.Clone(),
		GrantedBy: grantedByCopy,
		ExpiresAt: expiresCopy,
		CreatedAt: createdLocal,
	}, nil
}

// UserRole merepresentasikan penugasan peran langsung ke pengguna pada tabel user_roles.
type UserRole struct {
	ID        string
	UserID    string
	RoleID    string
	Scope     Scope
	GrantedBy *string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

// NewUserRole membentuk penugasan peran pengguna dengan validasi scope dan masa berlaku.
func NewUserRole(
	id, userID, roleID string,
	scope Scope,
	grantedBy *string,
	expiresAt *time.Time,
	now time.Time,
) (*UserRole, error) {
	core, err := newRoleAssignmentCore(id, "user_id", userID, roleID, scope, grantedBy, expiresAt, now)
	if err != nil {
		return nil, err
	}
	if core.Scope.Type == ScopeTypeSelf &&
		(core.Scope.ID == nil || !strings.EqualFold(strings.TrimSpace(*core.Scope.ID), strings.TrimSpace(userID))) {
		return nil, iamerrors.ErrInvalidScope
	}
	return &UserRole{
		ID:        core.ID,
		UserID:    strings.TrimSpace(userID),
		RoleID:    core.RoleID,
		Scope:     core.Scope,
		GrantedBy: core.GrantedBy,
		ExpiresAt: core.ExpiresAt,
		CreatedAt: core.CreatedAt,
	}, nil
}

// IsExpired melaporkan apakah penugasan peran pengguna sudah melewati batas waktu kedaluwarsa.
func (ur *UserRole) IsExpired(now time.Time) bool {
	if ur == nil || ur.ExpiresAt == nil {
		return false
	}
	return !normalizeLocalTime(now).Before(normalizeLocalTime(*ur.ExpiresAt))
}

// IsEffective melaporkan apakah penugasan peran pengguna masih berlaku pada waktu now.
func (ur *UserRole) IsEffective(now time.Time) bool {
	return ur != nil && !ur.IsExpired(now)
}

// GroupRole merepresentasikan penugasan peran melalui grup pada tabel group_roles.
type GroupRole struct {
	ID        string
	GroupID   string
	RoleID    string
	Scope     Scope
	GrantedBy *string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

// NewGroupRole membentuk penugasan peran grup dengan validasi scope dan masa berlaku.
func NewGroupRole(
	id, groupID, roleID string,
	scope Scope,
	grantedBy *string,
	expiresAt *time.Time,
	now time.Time,
) (*GroupRole, error) {
	if scope.Type == ScopeTypeSelf {
		return nil, iamerrors.ErrInvalidScope
	}
	core, err := newRoleAssignmentCore(id, "group_id", groupID, roleID, scope, grantedBy, expiresAt, now)
	if err != nil {
		return nil, err
	}
	return &GroupRole{
		ID:        core.ID,
		GroupID:   strings.TrimSpace(groupID),
		RoleID:    core.RoleID,
		Scope:     core.Scope,
		GrantedBy: core.GrantedBy,
		ExpiresAt: core.ExpiresAt,
		CreatedAt: core.CreatedAt,
	}, nil
}

// IsExpired melaporkan apakah penugasan peran grup sudah melewati batas waktu kedaluwarsa.
func (gr *GroupRole) IsExpired(now time.Time) bool {
	if gr == nil || gr.ExpiresAt == nil {
		return false
	}
	return !normalizeLocalTime(now).Before(normalizeLocalTime(*gr.ExpiresAt))
}

// IsEffective melaporkan apakah penugasan peran grup masih berlaku pada waktu now.
func (gr *GroupRole) IsEffective(now time.Time) bool {
	return gr != nil && !gr.IsExpired(now)
}

// ScopedPermission merepresentasikan satu izin efektif beserta cakupan dan status sensitivitasnya.
type ScopedPermission struct {
	PermissionCode string
	IsSensitive    bool
	Scope          Scope
}

// AccessSnapshot memuat seluruh hak akses efektif hasil resolusi peran langsung dan grup dalam bentuk snapshot terpusat.
type AccessSnapshot struct {
	UserID        string
	UserKind      UserKind
	UserStatus    UserStatus
	MFAEnabled    bool
	AuthEpoch     int64
	PermRevision  int64
	RecoveryEpoch int64
	Roles         []Role
	Permissions   []ScopedPermission
	CheckedAt     time.Time
}

// HasPermission memeriksa apakah snapshot memiliki izin bernormalisasi yang sesuai pada scope target.
func (s AccessSnapshot) HasPermission(code string, required Scope) (ScopedPermission, bool) {
	// Akun selain status ACTIVE (seperti PENDING, SUSPENDED, atau DISABLED) tidak diizinkan memakai izin apa pun meski role sudah terpasang.
	if s.UserStatus != UserStatusActive {
		return ScopedPermission{}, false
	}
	normalized := NormalizeCode(code)
	if normalized == "" || !required.IsValid() {
		return ScopedPermission{}, false
	}
	if IsSelfServicePermissionCode(normalized) {
		// Izin self-service hanya sah pada cakupan self milik pengguna terkait.
		if required.Type != ScopeTypeSelf || required.ID == nil || strings.TrimSpace(*required.ID) == "" {
			return ScopedPermission{}, false
		}
		if strings.TrimSpace(s.UserID) != "" && !strings.EqualFold(strings.TrimSpace(*required.ID), strings.TrimSpace(s.UserID)) {
			return ScopedPermission{}, false
		}
	}
	for _, perm := range s.Permissions {
		if normalized == authz.PermPayslipsReadSelf && perm.Scope.Type != ScopeTypeSelf {
			continue
		}
		if NormalizeCode(perm.PermissionCode) == normalized && perm.Scope.Matches(required) {
			return perm, true
		}
	}
	return ScopedPermission{}, false
}

// IsFresh memeriksa apakah snapshot masih berada di dalam jendela kesegaran 30 detik terhadap waktu sekarang.
func (s AccessSnapshot) IsFresh(maxAge time.Duration, now time.Time) bool {
	if s.CheckedAt.IsZero() || maxAge <= 0 {
		return false
	}
	nowLocal := normalizeLocalTime(now)
	checkedLocal := normalizeLocalTime(s.CheckedAt)
	if nowLocal.Before(checkedLocal) {
		return false
	}
	return nowLocal.Sub(checkedLocal) <= maxAge
}

// Clone membuat salinan dalam untuk menghindari data race atau mutasi slice antar pemanggil.
func (s AccessSnapshot) Clone() AccessSnapshot {
	clonedPerms := make([]ScopedPermission, len(s.Permissions))
	for i, p := range s.Permissions {
		clonedPerms[i] = ScopedPermission{
			PermissionCode: p.PermissionCode,
			IsSensitive:    p.IsSensitive,
			Scope:          p.Scope.Clone(),
		}
	}
	return AccessSnapshot{
		UserID:        s.UserID,
		UserKind:      s.UserKind,
		UserStatus:    s.UserStatus,
		MFAEnabled:    s.MFAEnabled,
		AuthEpoch:     s.AuthEpoch,
		PermRevision:  s.PermRevision,
		RecoveryEpoch: s.RecoveryEpoch,
		Roles:         slices.Clone(s.Roles),
		Permissions:   clonedPerms,
		CheckedAt:     normalizeLocalTime(s.CheckedAt),
	}
}
