package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
)

type bootstrapPermissionSpec struct {
	Code        string
	Domain      string
	Description string
	IsSensitive bool
}

type bootstrapRoleSpec struct {
	Code                string
	Name                string
	Description         string
	Domain              string
	DefaultRedirectPath string
	RedirectPriority    int
	PermissionCodes     []string
}

var defaultBootstrapPermissions = []bootstrapPermissionSpec{
	{Code: "iam.dashboard.view", Domain: "iam", Description: "View IntegriOne Identity dashboard", IsSensitive: false},
	{Code: "iam.users.read", Domain: "iam", Description: "Read IAM user principals and identities", IsSensitive: false},
	{Code: "iam.users.readself", Domain: "iam", Description: "Read own IAM user profile, identities, and assigned roles", IsSensitive: false},
	{Code: "iam.users.write", Domain: "iam", Description: "Create, update, suspend, or disable IAM users", IsSensitive: true},
	{Code: "iam.users.updateself", Domain: "iam", Description: "Update own IAM profile, email, username, password, and MFA factors", IsSensitive: false},
	{Code: "iam.users.deleteself", Domain: "iam", Description: "Delete own unverified pending email address", IsSensitive: false},
	{Code: "iam.roles.read", Domain: "iam", Description: "Read RBAC roles and assignments", IsSensitive: false},
	{Code: "iam.roles.write", Domain: "iam", Description: "Create, edit, or assign RBAC roles", IsSensitive: true},
	{Code: "iam.groups.read", Domain: "iam", Description: "Read IAM groups and memberships", IsSensitive: false},
	{Code: "iam.groups.write", Domain: "iam", Description: "Manage IAM groups and role bindings", IsSensitive: true},
	{Code: "iam.permissions.read", Domain: "iam", Description: "Read IAM permission catalog", IsSensitive: false},
	{Code: "iam.sessions.read", Domain: "iam", Description: "Inspect active IAM sessions", IsSensitive: false},
	{Code: "iam.sessions.readself", Domain: "iam", Description: "Inspect own active IAM sessions", IsSensitive: false},
	{Code: "iam.sessions.revoke", Domain: "iam", Description: "Revoke active IAM user sessions", IsSensitive: true},
	{Code: "iam.sessions.deleteself", Domain: "iam", Description: "Revoke own active IAM sessions", IsSensitive: true},
}

var defaultBootstrapRoles = []bootstrapRoleSpec{
	{
		Code:                "superadmin",
		Name:                "Super Administrator",
		Description:         "IntegriOne full platform and IAM governance super administrator",
		Domain:              "iam",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    1,
		PermissionCodes: []string{
			"iam.dashboard.view",
			"iam.users.read", "iam.users.readself", "iam.users.write", "iam.users.updateself", "iam.users.deleteself",
			"iam.roles.read", "iam.roles.write",
			"iam.groups.read", "iam.groups.write",
			"iam.permissions.read",
			"iam.sessions.read", "iam.sessions.readself", "iam.sessions.revoke", "iam.sessions.deleteself",
		},
	},
	{
		Code:                "iam_operator",
		Name:                "IAM Operator",
		Description:         "Operates user onboarding, group assignments, and session revocations",
		Domain:              "iam",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    10,
		PermissionCodes: []string{
			"iam.dashboard.view",
			"iam.users.read", "iam.users.readself", "iam.users.write", "iam.users.updateself", "iam.users.deleteself",
			"iam.roles.read",
			"iam.groups.read", "iam.groups.write",
			"iam.permissions.read",
			"iam.sessions.read", "iam.sessions.readself", "iam.sessions.revoke", "iam.sessions.deleteself",
		},
	},
	{
		Code:                "iam_readonly",
		Name:                "IAM Auditor Read-Only",
		Description:         "Read-only visibility into IAM principals, roles, and sessions",
		Domain:              "iam",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    20,
		PermissionCodes: []string{
			"iam.dashboard.view",
			"iam.users.read", "iam.users.readself",
			"iam.roles.read",
			"iam.groups.read",
			"iam.permissions.read",
			"iam.sessions.read", "iam.sessions.readself",
		},
	},
	{
		Code:                "customer",
		Name:                "Web Commerce Customer",
		Description:         "Default authenticated storefront customer role",
		Domain:              "commerce",
		DefaultRedirectPath: "/",
		RedirectPriority:    100,
		PermissionCodes: []string{
			"iam.users.readself",
			"iam.users.updateself",
			"iam.users.deleteself",
			"iam.sessions.readself",
			"iam.sessions.deleteself",
		},
	},
	{
		Code:                "iam_selfservice",
		Name:                "IAM Account Self-Service",
		Description:         "Default self-scoped profile, security, and session controls for internal and partner accounts",
		Domain:              "iam",
		DefaultRedirectPath: "",
		RedirectPriority:    999,
		PermissionCodes: []string{
			"iam.users.readself",
			"iam.users.updateself",
			"iam.users.deleteself",
			"iam.sessions.readself",
			"iam.sessions.deleteself",
		},
	},
}

// BootstrapSystemPrincipalAndRBAC menjamin katalog permission IAM, template peran bawaan, dan akun system bootstrap terbentuk idempoten saat startup.
func BootstrapSystemPrincipalAndRBAC(
	ctx context.Context,
	repos *iamrepo.Registry,
	reg *Registry,
	cfg config.SystemBootstrapConfig,
	logger *slog.Logger,
) error {
	if repos == nil || reg == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.bootstrap.nil_dependencies"), "bootstrap dependencies are required")
	}

	permByCode, err := ensureBootstrapPermissions(ctx, repos.Permissions, reg)
	if err != nil {
		return err
	}
	roleByCode, err := ensureBootstrapRoles(ctx, repos.Roles, reg, permByCode)
	if err != nil {
		return err
	}
	if err := ensureBaselineSelfRolesForExistingUsers(ctx, repos, reg, roleByCode, logger); err != nil {
		return err
	}

	if strings.TrimSpace(cfg.Password) == "" || (strings.TrimSpace(cfg.Email) == "" && strings.TrimSpace(cfg.Username) == "") {
		return nil
	}

	userID, err := ensureBootstrapUser(ctx, repos, cfg)
	if err != nil {
		return err
	}

	roleCode := strings.ToLower(strings.TrimSpace(cfg.RoleCode))
	if roleCode == "" {
		roleCode = "superadmin"
	}
	targetRole, ok := roleByCode[roleCode]
	if !ok {
		return iamerrors.WithField(iamerrors.ErrRoleNotFound, "role_code", "bootstrap role does not exist")
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

	if reg.Sessions != nil {
		_, _ = reg.Sessions.RefreshAccessSnapshot(ctx, userID)
	}

	if logger != nil {
		logger.Info("iam system bootstrap verified idempotently",
			slog.String("system_user_id", userID),
			slog.String("system_role", roleCode),
		)
	}
	return nil
}

func ensureBootstrapPermissions(ctx context.Context, repo domainrepo.PermissionRepository, reg *Registry) (map[string]*entity.Permission, error) {
	permByCode := make(map[string]*entity.Permission, len(defaultBootstrapPermissions))
	for _, spec := range defaultBootstrapPermissions {
		p, err := reg.Permissions.GetByCode(ctx, spec.Code)
		if errors.Is(err, iamerrors.ErrPermissionNotFound) {
			created, createErr := reg.Permissions.Create(ctx, permissionCreateInput(spec))
			if errors.Is(createErr, iamerrors.ErrPermissionAlreadyExists) {
				p, err = reg.Permissions.GetByCode(ctx, spec.Code)
				if err != nil {
					return nil, err
				}
			} else if createErr != nil {
				return nil, createErr
			} else {
				p = created
			}
		} else if err != nil {
			return nil, err
		}
		expected := permissionCreateInput(spec)
		if p.IsSensitive != expected.IsSensitive {
			return nil, iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.bootstrap.permission_sensitivity_mismatch"), "bootstrap permission sensitivity does not match the approved catalog")
		}
		if p.Resource != expected.Resource || p.Action != expected.Action {
			corrected := *p
			corrected.Resource = expected.Resource
			corrected.Action = expected.Action
			if err := repo.Update(ctx, &corrected); err != nil {
				return nil, err
			}
			p = &corrected
		}
		permByCode[spec.Code] = p
	}
	return permByCode, nil
}

func ensureBootstrapRoles(
	ctx context.Context,
	roleRepo domainrepo.RoleRepository,
	reg *Registry,
	permByCode map[string]*entity.Permission,
) (map[string]*entity.Role, error) {
	roleByCode := make(map[string]*entity.Role, len(defaultBootstrapRoles))
	for _, spec := range defaultBootstrapRoles {
		r, err := reg.Roles.GetByCode(ctx, spec.Code)
		if errors.Is(err, iamerrors.ErrRoleNotFound) {
			created, createErr := reg.Roles.Create(ctx, roleCreateInput(spec))
			if errors.Is(createErr, iamerrors.ErrRoleAlreadyExists) {
				r, err = reg.Roles.GetByCode(ctx, spec.Code)
				if err != nil {
					return nil, err
				}
			} else if createErr != nil {
				return nil, createErr
			} else {
				r = created
			}
		} else if err != nil {
			return nil, err
		} else if roleRepo != nil && r.IsSystem {
			specRedirect := strings.TrimSpace(spec.DefaultRedirectPath)
			if strings.TrimSpace(r.DefaultRedirectPath) != specRedirect || r.RedirectPriority != spec.RedirectPriority {
				// Sinkronkan metadata redirect peran sistem bawaan jika DB lama sempat menyimpan nilai lama (mis. "/").
				r.DefaultRedirectPath = specRedirect
				r.RedirectPriority = spec.RedirectPriority
				r.UpdatedAt = domainservice.SystemClock{}.Now()
				if updateErr := roleRepo.Update(ctx, r); updateErr != nil {
					return nil, updateErr
				}
			}
		}

		for _, permCode := range spec.PermissionCodes {
			perm := permByCode[permCode]
			if perm == nil {
				continue
			}
			_, assignErr := reg.Roles.BootstrapAssignPermission(ctx, r.ID, perm.ID)
			if assignErr != nil && !errors.Is(assignErr, iamerrors.ErrRolePermissionExists) {
				return nil, assignErr
			}
		}
		roleByCode[spec.Code] = r
	}
	return roleByCode, nil
}

func ensureBootstrapUser(ctx context.Context, repos *iamrepo.Registry, cfg config.SystemBootstrapConfig) (string, error) {
	var normEmail, normUsername string
	if strings.TrimSpace(cfg.Email) != "" {
		val, err := entity.ValidateEmail(cfg.Email)
		if err != nil {
			return "", err
		}
		normEmail = val
	}
	if strings.TrimSpace(cfg.Username) != "" {
		val, err := entity.ValidateUsername(cfg.Username)
		if err != nil {
			return "", err
		}
		normUsername = val
	}

	if existingUID, err := findBootstrapUserIDByIdentities(ctx, repos.UserIdentities, normEmail, normUsername); err != nil {
		return "", err
	} else if existingUID != "" {
		existingUser, err := repos.Users.SelectByID(ctx, existingUID)
		if err != nil {
			return "", err
		}
		if existingUser.Status != entity.UserStatusActive && !existingUser.IsDeleted() {
			now := domainservice.SystemClock{}.Now()
			if err := existingUser.Activate(now); err == nil {
				_ = repos.Users.Update(ctx, existingUser)
			}
		}
		return existingUID, nil
	}

	uid := strings.ToLower(strings.TrimSpace(cfg.UserUID))
	if uid != "" {
		if err := entity.ValidateUUID("system_user_uid", uid); err != nil {
			return "", err
		}
		if existingByID, err := repos.Users.SelectByID(ctx, uid); err == nil && existingByID != nil {
			return existingByID.ID, nil
		}
	} else {
		genID, err := crypto.NewUUIDGenerator().NewUUID()
		if err != nil {
			return "", err
		}
		uid = genID
	}

	kind := entity.UserKind(strings.ToLower(strings.TrimSpace(cfg.UserKind)))
	if !kind.IsValid() {
		kind = entity.UserKindSystem
	}
	displayName := strings.TrimSpace(cfg.DisplayName)
	if displayName == "" {
		displayName = "IntegriOne System Administrator"
	}

	now := domainservice.SystemClock{}.Now()
	u, err := entity.NewUser(uid, kind, entity.UserStatusActive, displayName, now)
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

func permissionCreateInput(spec bootstrapPermissionSpec) permission.CreateInput {
	separator := strings.LastIndexByte(spec.Code, '.')
	resource := spec.Domain
	action := spec.Code
	if separator > 0 && separator < len(spec.Code)-1 {
		resource = spec.Code[:separator]
		action = spec.Code[separator+1:]
	}
	return permission.CreateInput{
		Code:        spec.Code,
		Resource:    resource,
		Action:      action,
		Description: spec.Description,
		IsSensitive: spec.IsSensitive,
	}
}

func roleCreateInput(spec bootstrapRoleSpec) role.CreateInput {
	return role.CreateInput{
		Code:                spec.Code,
		Name:                spec.Name,
		Description:         spec.Description,
		ManagementDomain:    spec.Domain,
		DefaultRedirectPath: spec.DefaultRedirectPath,
		RedirectPriority:    spec.RedirectPriority,
		IsSystem:            true,
	}
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
