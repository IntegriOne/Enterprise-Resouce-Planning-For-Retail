package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
)

// BootstrapPrincipalsAndRBAC menjamin katalog permission, template peran bawaan, akun otomasi non-interaktif per management, dan akun human superadmin perdana terbentuk idempoten saat startup.
func BootstrapPrincipalsAndRBAC(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	cfg config.BootstrapConfig,
	logger *slog.Logger,
) error {
	if repos == nil || reg == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.bootstrap.nil_dependencies"), "bootstrap dependencies are required")
	}

	_, roleByCode, err := syncEmbeddedRBACCatalogs(ctx, repos, reg, logger)
	if err != nil {
		return err
	}

	// Jalankan migrasi & isolasi akun otomasi system terlebih dahulu sebelum akun human superadmin perdana dibuat.
	if err := ensureAutomationPrincipals(ctx, repos, reg, roleByCode, cfg.AutomationPrincipals, logger); err != nil {
		return err
	}

	if err := ensureBaselineSelfRolesForExistingUsers(ctx, repos, reg, roleByCode, logger); err != nil {
		return err
	}

	return ensureSuperadminPrincipal(ctx, repos, reg, roleByCode, cfg.Superadmin, logger)
}

func ensureSuperadminPrincipal(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	roleByCode map[string]*entity.Role,
	cfg config.SuperadminBootstrapConfig,
	logger *slog.Logger,
) error {
	if strings.TrimSpace(cfg.Password) == "" || (strings.TrimSpace(cfg.Email) == "" && strings.TrimSpace(cfg.Username) == "") {
		return nil
	}

	roleCode := entity.NormalizeCode(cfg.RoleCode)
	if roleCode == "" {
		roleCode = "superadmin"
	}
	if !entity.IsSuperAdminRoleCode(roleCode) {
		return iamerrors.WithField(iamerrors.ErrRoleUserKindMismatch, "role_code", "initial superadmin bootstrap must use superadmin role")
	}
	targetRole, ok := roleByCode[roleCode]
	if !ok {
		return iamerrors.WithField(iamerrors.ErrRoleNotFound, "role_code", "bootstrap superadmin role does not exist")
	}

	userKind := entity.UserKind(strings.ToLower(strings.TrimSpace(cfg.UserKind)))
	if userKind == "" {
		userKind = entity.UserKindEmployee
	}
	if userKind != entity.UserKindEmployee {
		return iamerrors.WithField(iamerrors.ErrInvalidUserKind, "user_kind", "superadmin principal must be a human employee account")
	}

	userID, err := ensureHumanSuperadminUser(ctx, repos, cfg)
	if err != nil {
		return err
	}

	globalScope := entity.Scope{Type: entity.ScopeTypeGlobal}
	exists, err := repos.UserRoles.CheckExists(ctx, userID, targetRole.ID, globalScope)
	if err != nil {
		return err
	}
	if !exists {
		if _, err := reg.Roles.AssignToUser(ctx, roleAssignmentInput(userID, targetRole.ID)); err != nil && !errors.Is(err, iamerrors.ErrUserRoleExists) {
			return err
		}
	}

	// Pastikan human superadmin juga memegang peran dasar self-service untuk mengelola profil dan MFA pribadinya.
	if selfRoleCode := entity.BaselineSelfRoleCodeForUserKind(entity.UserKindEmployee); selfRoleCode != "" {
		if selfRole := roleByCode[selfRoleCode]; selfRole != nil && selfRole.IsActive {
			uid := userID
			selfScope := entity.Scope{Type: entity.ScopeTypeSelf, ID: &uid}
			hasSelf, err := repos.UserRoles.CheckExists(ctx, userID, selfRole.ID, selfScope)
			if err != nil {
				return err
			}
			if !hasSelf {
				if _, err := reg.Roles.AssignToUser(ctx, selfRoleAssignmentInput(userID, selfRole.ID)); err != nil && !errors.Is(err, iamerrors.ErrUserRoleExists) {
					return err
				}
			}
		}
	}

	if reg.Sessions != nil {
		if _, err := reg.Sessions.RefreshAccessSnapshot(ctx, userID); err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Info("iam human superadmin bootstrap verified idempotently",
			slog.String("superadmin_user_id", userID),
			slog.String("superadmin_role", roleCode),
			slog.String("user_kind", userKind.String()),
		)
	}
	return nil
}

func ensureAutomationPrincipals(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	roleByCode map[string]*entity.Role,
	specs []config.AutomationBootstrapConfig,
	logger *slog.Logger,
) error {
	for _, spec := range specs {
		if strings.TrimSpace(spec.Username) == "" || strings.TrimSpace(spec.RoleCode) == "" {
			continue
		}
		if err := ensureSingleAutomationPrincipal(ctx, repos, reg, roleByCode, spec, logger); err != nil {
			return err
		}
	}
	return nil
}

func ensureSingleAutomationPrincipal(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	roleByCode map[string]*entity.Role,
	spec config.AutomationBootstrapConfig,
	logger *slog.Logger,
) error {
	normUsername, err := entity.ValidateUsername(spec.Username)
	if err != nil {
		return err
	}
	roleCode := entity.NormalizeCode(spec.RoleCode)
	if !entity.IsAutomationRoleCode(roleCode) {
		return iamerrors.WithField(iamerrors.ErrRoleUserKindMismatch, "role_code", "automation principal must use a management *_automation role")
	}
	targetRole, ok := roleByCode[roleCode]
	if !ok || !targetRole.IsActive {
		return iamerrors.WithField(iamerrors.ErrRoleNotFound, "role_code", "automation role does not exist")
	}

	displayName := strings.TrimSpace(spec.DisplayName)
	if displayName == "" {
		displayName = "Automation System (" + normUsername + ")"
	}

	uid := strings.ToLower(strings.TrimSpace(spec.UserUID))
	if uid != "" {
		if err := entity.ValidateUUID("automation_user_uid", uid); err != nil {
			return err
		}
	}

	existingUser, err := findExistingAutomationUser(ctx, repos, uid, normUsername, spec.LegacyAliases)
	if err != nil {
		return err
	}

	now := domainservice.SystemClock{}.Now()
	var userID string
	var migratedLegacy bool

	if existingUser != nil {
		userID = existingUser.ID
		migratedLegacy = existingUser.Kind != entity.UserKindSystem || existingUser.DisplayName != displayName
		if err := repos.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if existingUser.Status != entity.UserStatusActive && !existingUser.IsDeleted() {
				if actErr := existingUser.Activate(now); actErr != nil {
					return actErr
				}
			}
			if recErr := existingUser.ReconcileBootstrapIdentity(entity.UserKindSystem, displayName, now); recErr != nil {
				return recErr
			}
			if updErr := repos.Users.Update(txCtx, existingUser); updErr != nil {
				return updErr
			}
			if idErr := reconcileAutomationIdentitiesInTx(txCtx, repos, userID, normUsername, now); idErr != nil {
				return idErr
			}
			if credErr := stripInteractiveCredentialsInTx(txCtx, repos, userID); credErr != nil {
				return credErr
			}
			if roleErr := stripIncompatibleSystemRolesInTx(txCtx, repos, userID, targetRole.ID, now); roleErr != nil {
				return roleErr
			}
			if repos.SecurityRuntime != nil {
				if _, _, secErr := repos.SecurityRuntime.IncrementUserSecurityCounters(txCtx, userID, true, true); secErr != nil {
					return secErr
				}
			}
			return nil
		}); err != nil {
			return err
		}
		if reg.Sessions != nil {
			if _, revErr := reg.Sessions.RevokeAllUserSessions(ctx, userID, entity.SessionEndUserDisabled); revErr != nil {
				return revErr
			}
		}
	} else {
		if uid == "" {
			genID, genErr := crypto.NewUUIDGenerator().NewUUID()
			if genErr != nil {
				return genErr
			}
			uid = genID
		}
		u, err := entity.NewUser(uid, entity.UserKindSystem, entity.UserStatusActive, displayName, now)
		if err != nil {
			return err
		}
		identID, err := crypto.NewUUIDGenerator().NewUUID()
		if err != nil {
			return err
		}
		userIdent, err := entity.NewUserIdentity(identID, u.ID, entity.IdentityKindUsername, normUsername, true, now)
		if err != nil {
			return err
		}
		userIdent.MarkVerified(now)

		txErr := repos.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			return repos.Users.InsertWithIdentities(txCtx, u, []*entity.UserIdentity{userIdent})
		})
		if txErr != nil {
			// Tangani race kondisi startup antar replika dengan lookup ulang identitas yang baru terbentuk.
			if retryUser, findErr := findExistingAutomationUser(ctx, repos, uid, normUsername, spec.LegacyAliases); findErr == nil && retryUser != nil {
				userID = retryUser.ID
			} else {
				return txErr
			}
		} else {
			userID = u.ID
		}
	}

	globalScope := entity.Scope{Type: entity.ScopeTypeGlobal}
	hasTargetRole, err := repos.UserRoles.CheckExists(ctx, userID, targetRole.ID, globalScope)
	if err != nil {
		return err
	}
	if !hasTargetRole {
		if _, err := reg.Roles.AssignToUser(ctx, roleAssignmentInput(userID, targetRole.ID)); err != nil && !errors.Is(err, iamerrors.ErrUserRoleExists) {
			return err
		}
	}

	if reg.Sessions != nil {
		if _, err := reg.Sessions.RefreshAccessSnapshot(ctx, userID); err != nil {
			return err
		}
	}

	if logger != nil {
		logger.Info("iam management automation principal verified",
			slog.String("automation_user_id", userID),
			slog.String("username", normUsername),
			slog.String("management_domain", entity.NormalizeCode(spec.ManagementDomain)),
			slog.String("role_code", roleCode),
			slog.Bool("reconciled_existing", migratedLegacy),
		)
	}
	return nil
}

func findExistingAutomationUser(
	ctx context.Context,
	repos *iamrepo.Registry,
	uid, normUsername string,
	legacyAliases []string,
) (*entity.User, error) {
	if uid != "" {
		u, err := repos.Users.SelectByID(ctx, uid)
		if err == nil && u != nil {
			return u, nil
		}
		if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
			return nil, err
		}
	}

	candidates := append([]string{normUsername}, legacyAliases...)
	for _, rawAlias := range candidates {
		alias := strings.TrimSpace(rawAlias)
		if alias == "" {
			continue
		}
		var ident *entity.UserIdentity
		var err error
		if strings.Contains(alias, "@") {
			var normEmail string
			normEmail, err = entity.ValidateEmail(alias)
			if err != nil {
				continue
			}
			ident, err = repos.UserIdentities.FindUserIdentityByEmailKind(ctx, normEmail)
		} else {
			var normUser string
			normUser, err = entity.ValidateUsername(alias)
			if err != nil {
				continue
			}
			ident, err = repos.UserIdentities.FindUserIdentityByUsernameKind(ctx, normUser)
		}
		if err == nil && ident != nil {
			return repos.Users.SelectByID(ctx, ident.UserID)
		}
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return nil, err
		}
	}
	return nil, nil
}

func reconcileAutomationIdentitiesInTx(
	txCtx context.Context,
	repos *iamrepo.Registry,
	userID, normUsername string,
	now time.Time,
) error {
	idents, err := repos.UserIdentities.SelectByUserID(txCtx, userID)
	if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
		return err
	}

	var currentUsername *entity.UserIdentity
	for _, item := range idents {
		if item == nil {
			continue
		}
		if item.Kind == entity.IdentityKindEmail {
			// Akun otomasi system tidak pernah memiliki email login interaktif.
			if delErr := repos.UserIdentities.Delete(txCtx, item.ID); delErr != nil && !errors.Is(delErr, iamerrors.ErrIdentityNotFound) {
				return delErr
			}
			continue
		}
		if item.Kind == entity.IdentityKindUsername {
			if currentUsername == nil {
				currentUsername = item
			} else if delErr := repos.UserIdentities.Delete(txCtx, item.ID); delErr != nil && !errors.Is(delErr, iamerrors.ErrIdentityNotFound) {
				return delErr
			}
		}
	}

	if currentUsername != nil {
		if currentUsername.Value != normUsername || !currentUsername.IsPrimary || !currentUsername.IsVerified {
			if err := currentUsername.UpdateUsernameValue(normUsername, now); err != nil {
				return err
			}
			currentUsername.SetPrimary(true, now)
			return repos.UserIdentities.Update(txCtx, currentUsername)
		}
		return nil
	}

	idGen := crypto.NewUUIDGenerator()
	newID, err := idGen.NewUUID()
	if err != nil {
		return err
	}
	newIdent, err := entity.NewUserIdentity(newID, userID, entity.IdentityKindUsername, normUsername, true, now)
	if err != nil {
		return err
	}
	newIdent.MarkVerified(now)
	return repos.UserIdentities.Insert(txCtx, newIdent)
}

func stripInteractiveCredentialsInTx(txCtx context.Context, repos *iamrepo.Registry, userID string) error {
	if repos.UserCredentials != nil {
		if err := repos.UserCredentials.DeleteByUserID(txCtx, userID); err != nil && !errors.Is(err, iamerrors.ErrCredentialNotFound) {
			return err
		}
	}
	if repos.UserTOTPFactors != nil {
		if err := repos.UserTOTPFactors.DeleteByUserID(txCtx, userID); err != nil && !errors.Is(err, iamerrors.ErrTOTPNotFound) {
			return err
		}
	}
	if repos.UserRecoveryCodes != nil {
		if err := repos.UserRecoveryCodes.DeleteByUserID(txCtx, userID); err != nil && !errors.Is(err, iamerrors.ErrRecoveryCodeNotFound) {
			return err
		}
	}
	return nil
}

func stripIncompatibleSystemRolesInTx(
	txCtx context.Context,
	repos *iamrepo.Registry,
	userID, allowedRoleID string,
	now time.Time,
) error {
	if repos.UserRoles != nil {
		assignments, err := repos.UserRoles.SelectByUserID(txCtx, userID, true, now)
		if err != nil {
			return err
		}
		for _, ur := range assignments {
			if ur == nil {
				continue
			}
			if ur.RoleID == allowedRoleID && ur.Scope.Type == entity.ScopeTypeGlobal {
				continue
			}
			if err := repos.UserRoles.DeleteByID(txCtx, ur.ID); err != nil && !errors.Is(err, iamerrors.ErrUserRoleNotFound) {
				return err
			}
		}
	}
	if repos.GroupMembers != nil {
		groups, err := repos.GroupMembers.SelectGroupsByUserID(txCtx, userID)
		if err != nil {
			return err
		}
		for _, grp := range groups {
			if grp == nil {
				continue
			}
			incompatible := true
			if repos.GroupRoles != nil && repos.Roles != nil {
				grs, grErr := repos.GroupRoles.SelectByGroupID(txCtx, grp.ID, true, now)
				if grErr != nil {
					return grErr
				}
				incompatible = false
				for _, gr := range grs {
					if gr == nil {
						continue
					}
					roleObj, rErr := repos.Roles.SelectByID(txCtx, gr.RoleID)
					if rErr != nil {
						if errors.Is(rErr, iamerrors.ErrRoleNotFound) {
							continue
						}
						return rErr
					}
					if roleObj != nil && !entity.IsRoleCompatibleWithUserKind(roleObj.Code, entity.UserKindSystem) {
						incompatible = true
						break
					}
				}
			}
			if incompatible {
				if err := repos.GroupMembers.Delete(txCtx, grp.ID, userID); err != nil && !errors.Is(err, iamerrors.ErrGroupMemberNotFound) {
					return err
				}
			}
		}
	}
	return nil
}

func ensureHumanSuperadminUser(ctx context.Context, repos *iamrepo.Registry, cfg config.SuperadminBootstrapConfig) (string, error) {
	var normEmail, normUsername string
	var forbiddenTerms []string
	if strings.TrimSpace(cfg.Email) != "" {
		val, err := entity.ValidateEmail(cfg.Email)
		if err != nil {
			return "", err
		}
		normEmail = val
		forbiddenTerms = append(forbiddenTerms, normEmail)
	}
	if strings.TrimSpace(cfg.Username) != "" {
		val, err := entity.ValidateUsername(cfg.Username)
		if err != nil {
			return "", err
		}
		normUsername = val
		forbiddenTerms = append(forbiddenTerms, normUsername)
	}

	displayName := strings.TrimSpace(cfg.DisplayName)
	if displayName == "" {
		displayName = "Super Administrator"
	}
	forbiddenTerms = append(forbiddenTerms, displayName)

	if err := domainservice.NewDefaultPasswordPolicy().Validate(cfg.Password, forbiddenTerms...); err != nil {
		return "", err
	}

	if existingUID, err := findBootstrapUserIDByIdentities(ctx, repos.UserIdentities, normEmail, normUsername); err != nil {
		return "", err
	} else if existingUID != "" {
		existingUser, err := repos.Users.SelectByID(ctx, existingUID)
		if err != nil {
			return "", err
		}
		if existingUser.Kind == entity.UserKindSystem {
			return "", iamerrors.ErrSystemUserInteractiveForbidden
		}
		if err := activateOrUpdateHumanSuperadmin(ctx, repos, existingUser, displayName, cfg.Password); err != nil {
			return "", err
		}
		return existingUID, nil
	}

	uid := strings.ToLower(strings.TrimSpace(cfg.UserUID))
	if uid != "" {
		if err := entity.ValidateUUID("superadmin_user_uid", uid); err != nil {
			return "", err
		}
		if existingByID, err := repos.Users.SelectByID(ctx, uid); err == nil && existingByID != nil {
			if existingByID.Kind == entity.UserKindSystem {
				return "", iamerrors.ErrSystemUserInteractiveForbidden
			}
			if err := activateOrUpdateHumanSuperadmin(ctx, repos, existingByID, displayName, cfg.Password); err != nil {
				return "", err
			}
			return existingByID.ID, nil
		} else if err != nil && !errors.Is(err, iamerrors.ErrUserNotFound) {
			return "", err
		}
	} else {
		genID, err := crypto.NewUUIDGenerator().NewUUID()
		if err != nil {
			return "", err
		}
		uid = genID
	}

	now := domainservice.SystemClock{}.Now()
	u, err := entity.NewUser(uid, entity.UserKindEmployee, entity.UserStatusActive, displayName, now)
	if err != nil {
		return "", err
	}

	idGen := crypto.NewUUIDGenerator()
	var idents []*entity.UserIdentity
	if normEmail != "" {
		eid, err := idGen.NewUUID()
		if err != nil {
			return "", err
		}
		emailIdent, err := entity.NewUserIdentity(eid, u.ID, entity.IdentityKindEmail, normEmail, true, now)
		if err != nil {
			return "", err
		}
		emailIdent.MarkVerified(now)
		idents = append(idents, emailIdent)
	}
	if normUsername != "" {
		unid, err := idGen.NewUUID()
		if err != nil {
			return "", err
		}
		userIdent, err := entity.NewUserIdentity(unid, u.ID, entity.IdentityKindUsername, normUsername, normEmail == "", now)
		if err != nil {
			return "", err
		}
		userIdent.MarkVerified(now)
		idents = append(idents, userIdent)
	}

	hash, err := crypto.NewDefaultArgon2idHasher().Hash(cfg.Password)
	if err != nil {
		return "", err
	}
	cred, err := entity.NewUserCredential(u.ID, hash, false, now)
	if err != nil {
		return "", err
	}

	txErr := repos.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := repos.Users.InsertWithIdentities(txCtx, u, idents); err != nil {
			return err
		}
		return repos.UserCredentials.Insert(txCtx, cred)
	})
	if txErr != nil {
		// Tangani balapan startup antar replika IAM secara aman.
		if existingUID, checkErr := findBootstrapUserIDByIdentities(ctx, repos.UserIdentities, normEmail, normUsername); checkErr == nil && existingUID != "" {
			return existingUID, nil
		}
		return "", txErr
	}
	return u.ID, nil
}

func activateOrUpdateHumanSuperadmin(ctx context.Context, repos *iamrepo.Registry, existingUser *entity.User, displayName, rawPassword string) error {
	if existingUser == nil {
		return iamerrors.ErrUserNotFound
	}
	needsUpdate := false
	now := domainservice.SystemClock{}.Now()
	if existingUser.Status != entity.UserStatusActive && !existingUser.IsDeleted() {
		if err := existingUser.Activate(now); err != nil {
			return err
		}
		needsUpdate = true
	}
	if existingUser.Kind != entity.UserKindEmployee || existingUser.DisplayName != displayName {
		if err := existingUser.ReconcileBootstrapIdentity(entity.UserKindEmployee, displayName, now); err != nil {
			return err
		}
		needsUpdate = true
	}
	if needsUpdate {
		if err := repos.Users.Update(ctx, existingUser); err != nil {
			return err
		}
	}
	if repos.UserCredentials != nil {
		if _, err := repos.UserCredentials.SelectByUserID(ctx, existingUser.ID); errors.Is(err, iamerrors.ErrCredentialNotFound) {
			hash, hashErr := crypto.NewDefaultArgon2idHasher().Hash(rawPassword)
			if hashErr != nil {
				return hashErr
			}
			cred, credErr := entity.NewUserCredential(existingUser.ID, hash, false, now)
			if credErr != nil {
				return credErr
			}
			if insErr := repos.UserCredentials.Insert(ctx, cred); insErr != nil && !errors.Is(insErr, iamerrors.ErrCredentialAlreadyExists) {
				return insErr
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

// findBootstrapUserIDByIdentities mencari ID pengguna bootstrap yang sudah ada melalui lookup eksplisit email dan username.
func findBootstrapUserIDByIdentities(
	ctx context.Context,
	repo interface {
		FindUserIdentityByEmailKind(ctx context.Context, email string) (*entity.UserIdentity, error)
		FindUserIdentityByUsernameKind(ctx context.Context, username string) (*entity.UserIdentity, error)
	},
	normEmail, normUsername string,
) (string, error) {
	if normEmail != "" {
		ident, err := repo.FindUserIdentityByEmailKind(ctx, normEmail)
		if err == nil && ident != nil {
			return ident.UserID, nil
		}
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return "", err
		}
	}
	if normUsername != "" {
		ident, err := repo.FindUserIdentityByUsernameKind(ctx, normUsername)
		if err == nil && ident != nil {
			return ident.UserID, nil
		}
		if err != nil && !errors.Is(err, iamerrors.ErrIdentityNotFound) {
			return "", err
		}
	}
	return "", nil
}

func roleAssignmentInput(userID, roleID string) role.AssignUserRoleInput {
	return role.AssignUserRoleInput{
		UserID:    userID,
		RoleID:    roleID,
		ScopeType: entity.ScopeTypeGlobal,
	}
}

func selfRoleAssignmentInput(userID, roleID string) role.AssignUserRoleInput {
	uid := strings.TrimSpace(userID)
	return role.AssignUserRoleInput{
		UserID:    uid,
		RoleID:    roleID,
		ScopeType: entity.ScopeTypeSelf,
		ScopeID:   &uid,
	}
}

func ensureBaselineSelfRolesForExistingUsers(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	roleByCode map[string]*entity.Role,
	logger *slog.Logger,
) error {
	if repos == nil || repos.Users == nil || repos.UserRoles == nil || reg == nil || reg.Roles == nil {
		return nil
	}
	offset := 0
	assignedCount := 0
	for {
		page, err := repos.Users.SelectList(ctx, domainrepo.UserFilter{
			IncludeDeleted: false,
			Page: domainrepo.PageRequest{
				Limit:  domainrepo.MaxPageLimit,
				Offset: offset,
			},
		})
		if err != nil {
			return err
		}
		if len(page.Items) == 0 {
			break
		}
		for _, u := range page.Items {
			if u == nil || u.IsDeleted() {
				continue
			}
			roleCode := entity.BaselineSelfRoleCodeForUserKind(u.Kind)
			if roleCode == "" {
				continue
			}
			targetRole := roleByCode[roleCode]
			if targetRole == nil || !targetRole.IsActive {
				continue
			}
			uid := u.ID
			selfScope := entity.Scope{Type: entity.ScopeTypeSelf, ID: &uid}
			exists, err := repos.UserRoles.CheckExists(ctx, u.ID, targetRole.ID, selfScope)
			if err != nil {
				return err
			}
			if !exists {
				if _, assignErr := reg.Roles.AssignToUser(ctx, selfRoleAssignmentInput(u.ID, targetRole.ID)); assignErr != nil && !errors.Is(assignErr, iamerrors.ErrUserRoleExists) {
					return assignErr
				}
				assignedCount++
			}
		}
		if !page.HasMore {
			break
		}
		offset += len(page.Items)
	}
	if logger != nil && assignedCount > 0 {
		logger.Info("iam baseline self-service roles backfilled", slog.Int("assigned_count", assignedCount))
	}
	return nil
}
