package role

import (
	"context"
	"errors"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// Service mengorkestrasi manajemen peran RBAC, ikatan izin, serta penugasan peran berscope ke pengguna.
type Service struct {
	tx              domainrepo.Transactor
	roles           domainrepo.RoleRepository
	rolePermissions domainrepo.RolePermissionRepository
	permissions     domainrepo.PermissionRepository
	users           domainrepo.UserRepository
	userRoles       domainrepo.UserRoleRepository
	securityRuntime domainrepo.SecurityRuntimeRepository
	clock           domainservice.Clock
	idGen           domainservice.IDGenerator
	redirect        domainservice.RedirectResolver
	outboxCoord     *redisstore.Coordinator
	projector       domainservice.RuntimeSecurityProjector
	scopeVerifier   domainservice.ScopeOwnerVerifier
}

// NewService membentuk Service peran dengan validasi fail-closed seluruh port ketergantungan.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Transactor == nil {
		return nil, iamerrors.NewValidation("transactor", iamerrors.Code("iam.service.nil_transactor"), "transactor is required")
	}
	if deps.Roles == nil {
		return nil, iamerrors.NewValidation("roles", iamerrors.Code("iam.service.nil_role_repo"), "role repository is required")
	}
	if deps.RolePermissions == nil {
		return nil, iamerrors.NewValidation("role_permissions", iamerrors.Code("iam.service.nil_role_permission_repo"), "role permission repository is required")
	}
	if deps.Permissions == nil {
		return nil, iamerrors.NewValidation("permissions", iamerrors.Code("iam.service.nil_permission_repo"), "permission repository is required")
	}
	if deps.Users == nil {
		return nil, iamerrors.NewValidation("users", iamerrors.Code("iam.service.nil_user_repo"), "user repository is required")
	}
	if deps.UserRoles == nil {
		return nil, iamerrors.NewValidation("user_roles", iamerrors.Code("iam.service.nil_user_role_repo"), "user role repository is required")
	}
	if deps.SecurityRuntime == nil {
		return nil, iamerrors.NewValidation("security_runtime", iamerrors.Code("iam.service.nil_security_runtime_repo"), "security runtime repository is required")
	}
	if deps.OutboxCoord == nil {
		return nil, iamerrors.NewValidation("outbox_coordinator", iamerrors.Code("iam.service.nil_outbox_coordinator"), "outbox coordinator is required")
	}
	if deps.IDGen == nil {
		return nil, iamerrors.NewValidation("id_gen", iamerrors.Code("iam.service.nil_id_generator"), "id generator is required")
	}
	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	return &Service{
		tx:              deps.Transactor,
		roles:           deps.Roles,
		rolePermissions: deps.RolePermissions,
		permissions:     deps.Permissions,
		users:           deps.Users,
		userRoles:       deps.UserRoles,
		securityRuntime: deps.SecurityRuntime,
		clock:           clock,
		idGen:           deps.IDGen,
		redirect:        deps.Redirect,
		outboxCoord:     deps.OutboxCoord,
		projector:       deps.Projector,
		scopeVerifier:   deps.ScopeVerifier,
	}, nil
}

// GetByID mengambil detail peran berdasarkan UUID-nya.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", id); err != nil {
		return nil, err
	}
	return s.roles.SelectByID(ctx, id)
}

// GetByCode mengambil detail peran berdasarkan kode unik bernormalisasi.
func (s *Service) GetByCode(ctx context.Context, code string) (*entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normalized, err := entity.ValidateEntityCode(code, entity.MaxRoleCodeLength, iamerrors.ErrInvalidRoleCode)
	if err != nil {
		return nil, err
	}
	return s.roles.SelectByCode(ctx, normalized)
}

// List mengambil daftar peran berdasarkan filter domain dan status aktif.
func (s *Service) List(ctx context.Context, filter domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Role]{}, err
	}
	return s.roles.SelectList(ctx, filter.Normalize())
}

// Create membentuk peran baru dengan validasi kode unik dan allowlist jalur redirect.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := s.validateRedirectPath(in.DefaultRedirectPath); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.role.id_generation_failed"), "failed to generate role id")
	}

	r, err := entity.NewRole(
		id,
		in.Code,
		in.Name,
		in.Description,
		in.ManagementDomain,
		strings.TrimSpace(in.DefaultRedirectPath),
		in.RedirectPriority,
		in.IsSystem,
		s.clock.Now(),
	)
	if err != nil {
		return nil, err
	}

	existing, err := s.roles.SelectByCode(ctx, r.Code)
	if err == nil && existing != nil {
		return nil, iamerrors.WithField(iamerrors.ErrRoleAlreadyExists, "code", "role code already exists")
	}
	if err != nil && !errors.Is(err, iamerrors.ErrRoleNotFound) {
		return nil, err
	}

	if err := s.roles.Insert(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// UpdateDetails memperbarui metadata tampilan dan konfigurasi redirect peran yang ada.
func (s *Service) UpdateDetails(ctx context.Context, in UpdateInput) (*entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", in.ID); err != nil {
		return nil, err
	}
	if err := s.validateRedirectPath(in.DefaultRedirectPath); err != nil {
		return nil, err
	}

	var updated *entity.Role
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, in.ID)
		if err != nil {
			return err
		}
		if err := r.UpdateDetails(
			in.Name,
			in.Description,
			in.ManagementDomain,
			strings.TrimSpace(in.DefaultRedirectPath),
			in.RedirectPriority,
			s.clock.Now(),
		); err != nil {
			return err
		}
		if err := s.roles.Update(txCtx, r); err != nil {
			return err
		}
		if err := s.bumpAndRecordRoleSecurityInTx(txCtx, r.ID); err != nil {
			return err
		}
		updated = r
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Activate mengaktifkan peran agar kembali ikut dalam evaluasi hak akses.
func (s *Service) Activate(ctx context.Context, id string) (*entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", id); err != nil {
		return nil, err
	}

	var updated *entity.Role
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, id)
		if err != nil {
			return err
		}
		r.Activate(s.clock.Now())
		if err := s.roles.Update(txCtx, r); err != nil {
			return err
		}
		if err := s.bumpAndRecordRoleSecurityInTx(txCtx, r.ID); err != nil {
			return err
		}
		updated = r
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Deactivate menonaktifkan peran non-sistem sambil melindungi peran bawaan sistem.
func (s *Service) Deactivate(ctx context.Context, id string) (*entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", id); err != nil {
		return nil, err
	}

	var updated *entity.Role
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, id)
		if err != nil {
			return err
		}
		if err := r.Deactivate(s.clock.Now()); err != nil {
			return err
		}
		if err := s.roles.Update(txCtx, r); err != nil {
			return err
		}
		if err := s.bumpAndRecordRoleSecurityInTx(txCtx, r.ID); err != nil {
			return err
		}
		updated = r
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete menghapus peran non-sistem sambil menolak penghapusan peran bawaan sistem.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("role_id", id); err != nil {
		return err
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, id)
		if err != nil {
			return err
		}
		if r.IsSystem {
			return iamerrors.ErrSystemRoleImmutable
		}
		if err := s.bumpAndRecordRoleSecurityInTx(txCtx, id); err != nil {
			return err
		}
		return s.roles.Delete(txCtx, id)
	})
}

// AssignPermission mengikat izin atomik ke peran non-sistem secara transaksional.
func (s *Service) AssignPermission(ctx context.Context, roleID, permissionID string) (*entity.RolePermission, error) {
	return s.assignPermission(ctx, roleID, permissionID, false)
}

// BootstrapAssignPermission mengikat izin bawaan saat seed/bootstrap sistem termasuk pada peran sistem.
func (s *Service) BootstrapAssignPermission(ctx context.Context, roleID, permissionID string) (*entity.RolePermission, error) {
	return s.assignPermission(ctx, roleID, permissionID, true)
}

func (s *Service) assignPermission(ctx context.Context, roleID, permissionID string, allowSystem bool) (*entity.RolePermission, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", roleID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("permission_id", permissionID); err != nil {
		return nil, err
	}

	var created *entity.RolePermission
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, roleID)
		if err != nil {
			return err
		}
		if r.IsSystem && !allowSystem {
			return iamerrors.ErrSystemRoleImmutable
		}
		if _, err := s.permissions.SelectByID(txCtx, permissionID); err != nil {
			return err
		}
		exists, err := s.rolePermissions.CheckExists(txCtx, roleID, permissionID)
		if err != nil {
			return err
		}
		if exists {
			return iamerrors.ErrRolePermissionExists
		}

		rel, err := entity.NewRolePermission(roleID, permissionID, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.rolePermissions.Insert(txCtx, rel); err != nil {
			return err
		}
		if err := s.bumpAndRecordRoleSecurityInTx(txCtx, roleID); err != nil {
			return err
		}
		created = rel
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// RevokePermission melepaskan izin atomik dari peran non-sistem.
func (s *Service) RevokePermission(ctx context.Context, roleID, permissionID string) error {
	return s.revokePermission(ctx, roleID, permissionID, false)
}

// BootstrapRevokePermission melepaskan izin usang dari peran bawaan saat sinkronisasi katalog RBAC termasuk pada peran sistem.
func (s *Service) BootstrapRevokePermission(ctx context.Context, roleID, permissionID string) error {
	return s.revokePermission(ctx, roleID, permissionID, true)
}

func (s *Service) revokePermission(ctx context.Context, roleID, permissionID string, allowSystem bool) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("role_id", roleID); err != nil {
		return err
	}
	if err := entity.ValidateUUID("permission_id", permissionID); err != nil {
		return err
	}
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, roleID)
		if err != nil {
			return err
		}
		if r.IsSystem && !allowSystem {
			return iamerrors.ErrSystemRoleImmutable
		}
		if err := s.rolePermissions.Delete(txCtx, roleID, permissionID); err != nil {
			return err
		}
		return s.bumpAndRecordRoleSecurityInTx(txCtx, roleID)
	})
}

// BootstrapSyncSystemRoleDetails menyinkronkan metadata peran sistem terhadap manifest resmi dan memicu invalidasi cache bila berubah.
func (s *Service) BootstrapSyncSystemRoleDetails(ctx context.Context, roleID string, in UpdateInput) (*entity.Role, bool, error) {
	if err := s.ensureReady(); err != nil {
		return nil, false, err
	}
	if err := entity.ValidateUUID("role_id", roleID); err != nil {
		return nil, false, err
	}
	if err := s.validateRedirectPath(in.DefaultRedirectPath); err != nil {
		return nil, false, err
	}

	var (
		updated *entity.Role
		changed bool
	)
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		r, err := s.roles.SelectByID(txCtx, roleID)
		if err != nil {
			return err
		}
		didChange, err := r.ApplyBootstrapSystemMetadata(
			in.Name,
			in.Description,
			in.ManagementDomain,
			strings.TrimSpace(in.DefaultRedirectPath),
			in.RedirectPriority,
			s.clock.Now(),
		)
		if err != nil {
			return err
		}
		if !didChange {
			updated = r
			changed = false
			return nil
		}
		if err := s.roles.Update(txCtx, r); err != nil {
			return err
		}
		if err := s.bumpAndRecordRoleSecurityInTx(txCtx, r.ID); err != nil {
			return err
		}
		updated = r
		changed = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return updated, changed, nil
}

// BootstrapBumpRolesSecurity menaikkan revisi keamanan bagi seluruh pemegang peran yang sensitivitas izinnya diperketat saat bootstrap.
func (s *Service) BootstrapBumpRolesSecurity(ctx context.Context, roleIDs []string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(roleIDs))
	unique := make([]string, 0, len(roleIDs))
	for _, rawID := range roleIDs {
		id := strings.TrimSpace(rawID)
		if err := entity.ValidateUUID("role_id", id); err != nil {
			return err
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return nil
	}
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		for _, rid := range unique {
			if err := s.bumpAndRecordRoleSecurityInTx(txCtx, rid); err != nil {
				return err
			}
		}
		return nil
	})
}

// ValidateBootstrapRedirectPath memvalidasi path redirect bawaan peran terhadap allowlist rute aktif IAM.
func (s *Service) ValidateBootstrapRedirectPath(raw string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.validateRedirectPath(raw)
}

// ListPermissions mengembalikan seluruh izin yang terikat ke suatu peran.
func (s *Service) ListPermissions(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", roleID); err != nil {
		return nil, err
	}
	if _, err := s.roles.SelectByID(ctx, roleID); err != nil {
		return nil, err
	}
	return s.permissions.SelectByRoleID(ctx, roleID)
}

// AssignToUser menugaskan peran aktif berscope ke pengguna secara atomik dan fail-closed.
func (s *Service) AssignToUser(ctx context.Context, in AssignUserRoleInput) (*entity.UserRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", in.RoleID); err != nil {
		return nil, err
	}

	scope, err := entity.NewScope(in.ScopeType, in.ScopeID)
	if err != nil {
		return nil, err
	}
	if err := domainservice.VerifyScope(ctx, s.scopeVerifier, scope); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if in.ExpiresAt != nil && !in.ExpiresAt.After(now) {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "expires_at", "expires_at must be in the future")
	}
	assignmentID, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.user_role.id_generation_failed"), "failed to generate user role id")
	}

	var created *entity.UserRole
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.users.SelectByID(txCtx, in.UserID)
		if err != nil {
			return err
		}
		if u.IsDeleted() {
			return iamerrors.ErrUserDeleted
		}

		r, err := s.roles.SelectByID(txCtx, in.RoleID)
		if err != nil {
			return err
		}
		if !r.IsActive {
			return iamerrors.ErrRoleInactive
		}
		if !entity.IsRoleCompatibleWithUserKind(r.Code, u.Kind) {
			return iamerrors.WithField(iamerrors.ErrRoleUserKindMismatch, "role_id", "")
		}

		exists, err := s.userRoles.CheckExists(txCtx, in.UserID, in.RoleID, scope)
		if err != nil {
			return err
		}
		if exists {
			return iamerrors.ErrUserRoleExists
		}

		var grantedBy *string
		if caller, ok := authcontext.FromContext(txCtx); ok && caller.Authenticated && caller.PrincipalID != "" {
			if err := entity.ValidateUUID("principal_id", caller.PrincipalID); err != nil {
				return err
			}
			actor := caller.PrincipalID
			grantedBy = &actor
		}
		assignment, err := entity.NewUserRole(assignmentID, in.UserID, in.RoleID, scope, grantedBy, in.ExpiresAt, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.userRoles.Insert(txCtx, assignment); err != nil {
			return err
		}
		if err := s.bumpAndRecordUserPermissionsInTx(txCtx, in.UserID); err != nil {
			return err
		}
		created = assignment
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// RevokeUserRole mencabut penugasan peran langsung pengguna berdasarkan ID penugasannya.
func (s *Service) RevokeUserRole(ctx context.Context, assignmentID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("user_role_id", assignmentID); err != nil {
		return err
	}
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		userID, err := s.securityRuntime.ResolveUserRoleUserID(txCtx, assignmentID)
		if err != nil {
			return err
		}
		if err := s.userRoles.DeleteByID(txCtx, assignmentID); err != nil {
			return err
		}
		return s.bumpAndRecordUserPermissionsInTx(txCtx, userID)
	})
}

// ListUserRoles mengambil penugasan peran langsung yang dimiliki pengguna.
func (s *Service) ListUserRoles(ctx context.Context, userID string, includeExpired bool) ([]*entity.UserRole, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return s.userRoles.SelectByUserID(ctx, userID, includeExpired, s.clock.Now())
}

// ListEffectiveRoles mengambil seluruh peran efektif aktif baik dari penugasan langsung maupun keanggotaan grup.
func (s *Service) ListEffectiveRoles(ctx context.Context, userID string) ([]entity.Role, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return s.roles.SelectEffectiveByUserID(ctx, userID, s.clock.Now())
}

func (s *Service) validateRedirectPath(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return s.redirect.ValidateAllowedPath(trimmed)
}

func (s *Service) ensureReady() error {
	if s == nil || s.roles == nil || s.tx == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.role.service_unavailable"), "role service is not initialized")
	}
	return nil
}
