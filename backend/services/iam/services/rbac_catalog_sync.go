package services

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
	"github.com/integrione/erp-retail/backend/shared/rbac"
)

// Kunci advisory transaksi PostgreSQL khusus sinkronisasi manifest katalog RBAC (terpisah dari iamMigrationLockID).
const iamRBACCatalogSyncLockID int64 = 718291824

type catalogSyncStats struct {
	PermissionsTotal         int
	PermissionsCreated       int
	PermissionsUpdated       int
	SensitivityUpgraded      int
	RolesTotal               int
	RolesCreated             int
	SystemRolesSynced        int
	TemplateRolesSeededOnce  int
	RolePermissionsAssigned  int
	RolePermissionsRevoked   int
	RolesSecurityCacheBumped int
}

type catalogSyncState struct {
	permByCode               map[string]*entity.Permission
	roleByCode               map[string]*entity.Role
	upgradedSensitivePermIDs map[string]struct{}
	stats                    catalogSyncStats
}

// syncEmbeddedRBACCatalogs memuat manifest deklaratif yang disematkan dan menyinkronkannya ke database IAM.
func syncEmbeddedRBACCatalogs(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	logger *slog.Logger,
) (map[string]*entity.Permission, map[string]*entity.Role, error) {
	catalogs, err := rbac.LoadEmbeddedCatalogs()
	if err != nil {
		return nil, nil, iamerrors.Wrap(err, iamerrors.KindInvalidInput, iamerrors.Code("iam.bootstrap.invalid_rbac_catalog"), "embedded rbac catalogs failed validation")
	}
	return syncRBACCatalogSet(ctx, repos, reg, catalogs, logger)
}

func syncRBACCatalogSet(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	catalogs *rbac.CatalogSet,
	logger *slog.Logger,
) (map[string]*entity.Permission, map[string]*entity.Role, error) {
	if err := validateCatalogSyncInputs(repos, reg, catalogs); err != nil {
		return nil, nil, err
	}

	// Tolak sejak awal bila ada template peran yang memakai redirect path di luar allowlist aktif IAM.
	for _, spec := range catalogs.Roles {
		if err := reg.Roles.ValidateBootstrapRedirectPath(spec.DefaultRedirectPath); err != nil {
			return nil, nil, iamerrors.Wrap(err, iamerrors.KindInvalidInput, iamerrors.Code("iam.bootstrap.role_redirect_not_allowed"), "catalog role default_redirect_path failed allowlist check")
		}
	}

	var state catalogSyncState
	err := repos.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := pgstore.AcquireAdvisoryXactLock(txCtx, iamRBACCatalogSyncLockID); err != nil {
			return err
		}

		syncedState, err := syncCatalogPermissionsInTx(txCtx, repos.Permissions, reg.Permissions, catalogs)
		if err != nil {
			return err
		}
		if err := syncCatalogRolesInTx(txCtx, repos, reg.Roles, catalogs, &syncedState); err != nil {
			return err
		}
		state = syncedState
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if logger != nil {
		logger.Info(
			"iam rbac catalog sync completed",
			slog.Int("manifests", len(catalogs.Manifests)),
			slog.Int("permissions_total", state.stats.PermissionsTotal),
			slog.Int("permissions_created", state.stats.PermissionsCreated),
			slog.Int("permissions_updated", state.stats.PermissionsUpdated),
			slog.Int("sensitivity_upgraded", state.stats.SensitivityUpgraded),
			slog.Int("roles_total", state.stats.RolesTotal),
			slog.Int("roles_created", state.stats.RolesCreated),
			slog.Int("system_roles_synced", state.stats.SystemRolesSynced),
			slog.Int("template_roles_preserved", state.stats.TemplateRolesSeededOnce),
			slog.Int("role_permissions_assigned", state.stats.RolePermissionsAssigned),
			slog.Int("role_permissions_revoked", state.stats.RolePermissionsRevoked),
			slog.Int("roles_security_bumped", state.stats.RolesSecurityCacheBumped),
		)
	}
	return state.permByCode, state.roleByCode, nil
}

func validateCatalogSyncInputs(repos *iamrepo.Registry, reg *Registry, catalogs *rbac.CatalogSet) error {
	if repos == nil || repos.Transactor == nil || repos.Permissions == nil || repos.Roles == nil || repos.RolePermissions == nil ||
		reg == nil || reg.Permissions == nil || reg.Roles == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.bootstrap.nil_dependencies"), "bootstrap dependencies are required")
	}
	if catalogs == nil {
		return iamerrors.NewValidation("catalogs", iamerrors.Code("iam.bootstrap.nil_rbac_catalog"), "rbac catalog set must not be nil")
	}
	return nil
}

func syncCatalogPermissionsInTx(
	txCtx context.Context,
	repo domainrepo.PermissionRepository,
	permSvc *permission.Service,
	catalogs *rbac.CatalogSet,
) (catalogSyncState, error) {
	state := catalogSyncState{
		permByCode:               make(map[string]*entity.Permission, len(catalogs.Permissions)),
		roleByCode:               make(map[string]*entity.Role, len(catalogs.Roles)),
		upgradedSensitivePermIDs: make(map[string]struct{}),
		stats: catalogSyncStats{
			PermissionsTotal: len(catalogs.Permissions),
			RolesTotal:       len(catalogs.Roles),
		},
	}

	for _, spec := range catalogs.Permissions {
		existing, err := repo.SelectByCode(txCtx, spec.Code)
		if err != nil {
			if !errors.Is(err, iamerrors.ErrPermissionNotFound) {
				return catalogSyncState{}, err
			}
			created, createErr := permSvc.Create(txCtx, permissionCreateInputFromSpec(spec))
			if createErr != nil {
				return catalogSyncState{}, createErr
			}
			state.permByCode[created.Code] = created
			state.stats.PermissionsCreated++
			continue
		}

		// Pelemahan tingkat sensitivitas dilarang keras (fail-closed) agar perlindungan reauth tidak turun diam-diam.
		if existing.IsSensitive && !spec.IsSensitive {
			return catalogSyncState{}, iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.bootstrap.permission_sensitivity_downgrade_forbidden"), "bootstrap cannot downgrade permission sensitivity")
		}

		sensitivityUpgraded := !existing.IsSensitive && spec.IsSensitive
		needsUpdate := existing.Resource != spec.Resource ||
			existing.Action != spec.Action ||
			strings.TrimSpace(existing.Description) != spec.Description ||
			sensitivityUpgraded
		if needsUpdate {
			updated := *existing
			updated.Resource = spec.Resource
			updated.Action = spec.Action
			updated.Description = spec.Description
			updated.IsSensitive = spec.IsSensitive
			if updateErr := repo.Update(txCtx, &updated); updateErr != nil {
				return catalogSyncState{}, updateErr
			}
			existing = &updated
			state.stats.PermissionsUpdated++
		}
		if sensitivityUpgraded {
			state.upgradedSensitivePermIDs[existing.ID] = struct{}{}
			state.stats.SensitivityUpgraded++
		}
		state.permByCode[existing.Code] = existing
	}

	return state, nil
}

func syncCatalogRolesInTx(
	txCtx context.Context,
	repos *iamrepo.Registry,
	roleSvc *role.Service,
	catalogs *rbac.CatalogSet,
	state *catalogSyncState,
) error {
	bumpRoleIDsSet, err := collectRolesHoldingUpgradedSensitivity(txCtx, repos.Roles, repos.Permissions, state.upgradedSensitivePermIDs)
	if err != nil {
		return err
	}

	for _, spec := range catalogs.Roles {
		existing, err := repos.Roles.SelectByCode(txCtx, spec.Code)
		if err != nil {
			if !errors.Is(err, iamerrors.ErrRoleNotFound) {
				return err
			}
			created, assigned, createErr := createRoleFromSpec(txCtx, roleSvc, spec, state.permByCode)
			if createErr != nil {
				return createErr
			}
			state.roleByCode[created.Code] = created
			state.stats.RolesCreated++
			state.stats.RolePermissionsAssigned += assigned
			continue
		}

		if !spec.IsSystem {
			if existing.IsSystem {
				return iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.bootstrap.system_role_downgrade_forbidden"), "bootstrap cannot downgrade a system role to mutable template")
			}
			// Template peran operasional (is_system = false) bersifat seed-once agar perubahan admin via /identity tidak tertimpa saat deploy.
			state.roleByCode[existing.Code] = existing
			state.stats.TemplateRolesSeededOnce++
			continue
		}

		if !existing.IsSystem {
			return iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.bootstrap.non_system_role_conflict"), "existing non-system role conflicts with system role manifest")
		}

		synced, didMutate, err := syncExistingSystemRole(txCtx, repos.Permissions, roleSvc, existing, spec, state.permByCode, &state.stats)
		if err != nil {
			return err
		}
		if didMutate {
			// Perubahan izin atau metadata peran sudah langsung memicu bump revisi keamanan di service role.
			delete(bumpRoleIDsSet, synced.ID)
		}
		state.roleByCode[synced.Code] = synced
		state.stats.SystemRolesSynced++
	}

	if len(bumpRoleIDsSet) > 0 {
		bumpRoleIDs := make([]string, 0, len(bumpRoleIDsSet))
		for rid := range bumpRoleIDsSet {
			bumpRoleIDs = append(bumpRoleIDs, rid)
		}
		slices.Sort(bumpRoleIDs)
		if err := roleSvc.BootstrapBumpRolesSecurity(txCtx, bumpRoleIDs); err != nil {
			return err
		}
		state.stats.RolesSecurityCacheBumped = len(bumpRoleIDs)
	}
	return nil
}

func collectRolesHoldingUpgradedSensitivity(
	txCtx context.Context,
	roleRepo domainrepo.RoleRepository,
	permRepo domainrepo.PermissionRepository,
	upgradedPermIDs map[string]struct{},
) (map[string]struct{}, error) {
	affectedRoleIDs := make(map[string]struct{})
	if len(upgradedPermIDs) == 0 {
		return affectedRoleIDs, nil
	}

	offset := 0
	for {
		page, err := roleRepo.SelectList(txCtx, domainrepo.RoleFilter{
			Page: domainrepo.PageRequest{
				Limit:  domainrepo.MaxPageLimit,
				Offset: offset,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, r := range page.Items {
			if r == nil {
				continue
			}
			perms, err := permRepo.SelectByRoleID(txCtx, r.ID)
			if err != nil {
				return nil, err
			}
			for _, p := range perms {
				if p == nil {
					continue
				}
				if _, hit := upgradedPermIDs[p.ID]; hit {
					affectedRoleIDs[r.ID] = struct{}{}
					break
				}
			}
		}
		if !page.HasMore || len(page.Items) == 0 {
			break
		}
		offset += len(page.Items)
	}
	return affectedRoleIDs, nil
}

func createRoleFromSpec(
	txCtx context.Context,
	roleSvc *role.Service,
	spec rbac.RoleSpec,
	permByCode map[string]*entity.Permission,
) (*entity.Role, int, error) {
	created, err := roleSvc.Create(txCtx, roleCreateInputFromSpec(spec))
	if err != nil {
		return nil, 0, err
	}
	assigned := 0
	for _, permCode := range spec.PermissionCodes {
		perm, ok := permByCode[permCode]
		if !ok || perm == nil {
			return nil, 0, iamerrors.ErrPermissionNotFound
		}
		if _, err := roleSvc.BootstrapAssignPermission(txCtx, created.ID, perm.ID); err != nil && !errors.Is(err, iamerrors.ErrRolePermissionExists) {
			return nil, 0, err
		}
		assigned++
	}
	return created, assigned, nil
}

func syncExistingSystemRole(
	txCtx context.Context,
	permRepo domainrepo.PermissionRepository,
	roleSvc *role.Service,
	existing *entity.Role,
	spec rbac.RoleSpec,
	permByCode map[string]*entity.Permission,
	stats *catalogSyncStats,
) (*entity.Role, bool, error) {
	synced, metaChanged, err := roleSvc.BootstrapSyncSystemRoleDetails(txCtx, existing.ID, role.UpdateInput{
		ID:                  existing.ID,
		Name:                spec.Name,
		Description:         spec.Description,
		ManagementDomain:    spec.ManagementDomain,
		DefaultRedirectPath: spec.DefaultRedirectPath,
		RedirectPriority:    spec.RedirectPriority,
	})
	if err != nil {
		return nil, false, err
	}

	currentPerms, err := permRepo.SelectByRoleID(txCtx, synced.ID)
	if err != nil {
		return nil, false, err
	}
	currentByCode := make(map[string]*entity.Permission, len(currentPerms))
	for _, p := range currentPerms {
		if p != nil {
			currentByCode[p.Code] = p
		}
	}

	desiredSet := make(map[string]struct{}, len(spec.PermissionCodes))
	didMutate := metaChanged
	for _, permCode := range spec.PermissionCodes {
		desiredSet[permCode] = struct{}{}
		if _, alreadyAssigned := currentByCode[permCode]; alreadyAssigned {
			continue
		}
		perm, ok := permByCode[permCode]
		if !ok || perm == nil {
			return nil, false, iamerrors.ErrPermissionNotFound
		}
		if _, err := roleSvc.BootstrapAssignPermission(txCtx, synced.ID, perm.ID); err != nil && !errors.Is(err, iamerrors.ErrRolePermissionExists) {
			return nil, false, err
		}
		didMutate = true
		stats.RolePermissionsAssigned++
	}

	for _, existingPerm := range currentPerms {
		if existingPerm == nil {
			continue
		}
		if _, keep := desiredSet[existingPerm.Code]; keep {
			continue
		}
		if err := roleSvc.BootstrapRevokePermission(txCtx, synced.ID, existingPerm.ID); err != nil && !errors.Is(err, iamerrors.ErrRolePermissionNotFound) {
			return nil, false, err
		}
		didMutate = true
		stats.RolePermissionsRevoked++
	}

	return synced, didMutate, nil
}

func permissionCreateInputFromSpec(spec rbac.PermissionSpec) permission.CreateInput {
	return permission.CreateInput{
		Code:        spec.Code,
		Resource:    spec.Resource,
		Action:      spec.Action,
		Description: spec.Description,
		IsSensitive: spec.IsSensitive,
	}
}

func roleCreateInputFromSpec(spec rbac.RoleSpec) role.CreateInput {
	return role.CreateInput{
		Code:                spec.Code,
		Name:                spec.Name,
		Description:         spec.Description,
		ManagementDomain:    spec.ManagementDomain,
		DefaultRedirectPath: spec.DefaultRedirectPath,
		RedirectPriority:    spec.RedirectPriority,
		IsSystem:            spec.IsSystem,
	}
}
