package services

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/shared/rbac"
)

type syncTestStore struct {
	mu              sync.RWMutex
	permissions     map[string]*entity.Permission
	roles           map[string]*entity.Role
	rolePermissions map[string]*entity.RolePermission
	bumpedRoleIDs   []string
	outboxEvents    []*entity.OutboxEvent
}

func newSyncTestStore() *syncTestStore {
	return &syncTestStore{
		permissions:     make(map[string]*entity.Permission),
		roles:           make(map[string]*entity.Role),
		rolePermissions: make(map[string]*entity.RolePermission),
	}
}

type syncTestTx struct{}

func (syncTestTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type syncTestUserRepo struct{ domainrepo.UserRepository }
type syncTestIdentityRepo struct {
	domainrepo.UserIdentityRepository
}
type syncTestCredRepo struct {
	domainrepo.UserCredentialRepository
}
type syncTestTOTPRepo struct {
	domainrepo.UserTOTPFactorRepository
}
type syncTestRecoveryRepo struct {
	domainrepo.UserRecoveryCodeRepository
}
type syncTestTokenRepo struct{ domainrepo.UserTokenRepository }
type syncTestSessionRepo struct{ domainrepo.SessionRepository }
type syncTestUserRoleRepo struct{ domainrepo.UserRoleRepository }
type syncTestGroupRepo struct{ domainrepo.GroupRepository }
type syncTestGroupMemberRepo struct {
	domainrepo.GroupMemberRepository
}
type syncTestGroupRoleRepo struct{ domainrepo.GroupRoleRepository }

type syncTestPermRepo struct{ s *syncTestStore }

func (r *syncTestPermRepo) SelectByID(_ context.Context, id string) (*entity.Permission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.permissions[id]
	if !ok {
		return nil, iamerrors.ErrPermissionNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *syncTestPermRepo) SelectByCode(_ context.Context, code string) (*entity.Permission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.permissions {
		if p.Code == code {
			cp := *p
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrPermissionNotFound
}

func (r *syncTestPermRepo) SelectList(_ context.Context, filter domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]*entity.Permission, 0, len(r.s.permissions))
	for _, p := range r.s.permissions {
		if filter.Resource != "" && p.Resource != filter.Resource {
			continue
		}
		cp := *p
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.Permission) int { return strings.Compare(a.Code, b.Code) })
	return domainrepo.NewPageResult(out, int64(len(out)), filter.Page), nil
}

func (r *syncTestPermRepo) SelectByRoleID(_ context.Context, roleID string) ([]*entity.Permission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.Permission
	for _, rp := range r.s.rolePermissions {
		if rp.RoleID == roleID {
			if p, ok := r.s.permissions[rp.PermissionID]; ok {
				cp := *p
				out = append(out, &cp)
			}
		}
	}
	slices.SortFunc(out, func(a, b *entity.Permission) int { return strings.Compare(a.Code, b.Code) })
	return out, nil
}

func (r *syncTestPermRepo) Insert(_ context.Context, p *entity.Permission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *p
	r.s.permissions[p.ID] = &cp
	return nil
}

func (r *syncTestPermRepo) Update(_ context.Context, p *entity.Permission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.permissions[p.ID]; !ok {
		return iamerrors.ErrPermissionNotFound
	}
	cp := *p
	r.s.permissions[p.ID] = &cp
	return nil
}

func (r *syncTestPermRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.permissions[id]; !ok {
		return iamerrors.ErrPermissionNotFound
	}
	delete(r.s.permissions, id)
	return nil
}

type syncTestRoleRepo struct{ s *syncTestStore }

func (r *syncTestRoleRepo) SelectByID(_ context.Context, id string) (*entity.Role, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	roleObj, ok := r.s.roles[id]
	if !ok {
		return nil, iamerrors.ErrRoleNotFound
	}
	cp := *roleObj
	return &cp, nil
}

func (r *syncTestRoleRepo) SelectByCode(_ context.Context, code string) (*entity.Role, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, roleObj := range r.s.roles {
		if roleObj.Code == code {
			cp := *roleObj
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrRoleNotFound
}

func (r *syncTestRoleRepo) SelectList(_ context.Context, filter domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := make([]*entity.Role, 0, len(r.s.roles))
	for _, roleObj := range r.s.roles {
		cp := *roleObj
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.Role) int { return strings.Compare(a.Code, b.Code) })
	return domainrepo.NewPageResult(out, int64(len(out)), filter.Page), nil
}

func (r *syncTestRoleRepo) SelectEffectiveByUserID(context.Context, string, time.Time) ([]entity.Role, error) {
	return nil, nil
}

func (r *syncTestRoleRepo) Insert(_ context.Context, roleObj *entity.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *roleObj
	r.s.roles[roleObj.ID] = &cp
	return nil
}

func (r *syncTestRoleRepo) Update(_ context.Context, roleObj *entity.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.roles[roleObj.ID]; !ok {
		return iamerrors.ErrRoleNotFound
	}
	cp := *roleObj
	r.s.roles[roleObj.ID] = &cp
	return nil
}

func (r *syncTestRoleRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.roles[id]; !ok {
		return iamerrors.ErrRoleNotFound
	}
	delete(r.s.roles, id)
	return nil
}

type syncTestRolePermRepo struct{ s *syncTestStore }

func (r *syncTestRolePermRepo) CheckExists(_ context.Context, roleID, permID string) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	_, ok := r.s.rolePermissions[roleID+"|"+permID]
	return ok, nil
}

func (r *syncTestRolePermRepo) Insert(_ context.Context, rp *entity.RolePermission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *rp
	r.s.rolePermissions[rp.RoleID+"|"+rp.PermissionID] = &cp
	return nil
}

func (r *syncTestRolePermRepo) Delete(_ context.Context, roleID, permID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	key := roleID + "|" + permID
	if _, ok := r.s.rolePermissions[key]; !ok {
		return iamerrors.ErrRolePermissionNotFound
	}
	delete(r.s.rolePermissions, key)
	return nil
}

type syncTestSecurityRuntime struct {
	domainrepo.SecurityRuntimeRepository
	s *syncTestStore
}

func (m syncTestSecurityRuntime) GetRecoveryEpoch(context.Context) (int64, error) { return 1, nil }
func (m syncTestSecurityRuntime) IncrementRecoveryEpoch(context.Context) (int64, error) {
	return 2, nil
}
func (m syncTestSecurityRuntime) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (m syncTestSecurityRuntime) MarkEffective(context.Context, string, time.Time) error { return nil }
func (m syncTestSecurityRuntime) MarkRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (m syncTestSecurityRuntime) MarkFailed(context.Context, string, string) error { return nil }
func (m syncTestSecurityRuntime) CountPending(context.Context) (int64, error)      { return 0, nil }
func (m syncTestSecurityRuntime) CountFailed(context.Context) (int64, error)       { return 0, nil }
func (m syncTestSecurityRuntime) IncrementUserSecurityCounters(context.Context, string, bool, bool) (int64, int64, error) {
	return 1, 1, nil
}

func (m syncTestSecurityRuntime) BumpPermRevisionForRole(_ context.Context, roleID string) ([]string, error) {
	m.s.mu.Lock()
	defer m.s.mu.Unlock()
	m.s.bumpedRoleIDs = append(m.s.bumpedRoleIDs, roleID)
	return nil, nil
}

func (m syncTestSecurityRuntime) Insert(_ context.Context, ev *entity.OutboxEvent) error {
	m.s.mu.Lock()
	defer m.s.mu.Unlock()
	if ev != nil {
		cp := *ev
		m.s.outboxEvents = append(m.s.outboxEvents, &cp)
	}
	return nil
}

func newSyncTestHarness(t *testing.T, allowedRedirects ...string) (*syncTestStore, *iamrepo.Registry, *Registry) {
	t.Helper()
	store := newSyncTestStore()
	repos := &iamrepo.Registry{
		Transactor:        syncTestTx{},
		Users:             syncTestUserRepo{},
		UserIdentities:    syncTestIdentityRepo{},
		UserCredentials:   syncTestCredRepo{},
		UserTOTPFactors:   syncTestTOTPRepo{},
		UserRecoveryCodes: syncTestRecoveryRepo{},
		UserTokens:        syncTestTokenRepo{},
		Sessions:          syncTestSessionRepo{},
		Roles:             &syncTestRoleRepo{store},
		RolePermissions:   &syncTestRolePermRepo{store},
		UserRoles:         syncTestUserRoleRepo{},
		Permissions:       &syncTestPermRepo{store},
		Groups:            syncTestGroupRepo{},
		GroupMembers:      syncTestGroupMemberRepo{},
		GroupRoles:        syncTestGroupRoleRepo{},
		SecurityRuntime:   syncTestSecurityRuntime{s: store},
	}
	reg, err := NewRegistry(repos, SecurityOptions{
		AllowEphemeralDevKey:  true,
		AllowEphemeralCSRFKey: true,
		AllowedRedirectPaths:  allowedRedirects,
		IDGen:                 crypto.NewUUIDGenerator(),
	})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	return store, repos, reg
}

func TestSyncEmbeddedRBACCatalogs_Idempotent(t *testing.T) {
	t.Parallel()

	store, repos, reg := newSyncTestHarness(t)
	ctx := context.Background()

	perms1, roles1, err := syncEmbeddedRBACCatalogs(ctx, repos, reg, nil)
	if err != nil {
		t.Fatalf("syncEmbeddedRBACCatalogs(1) error = %v", err)
	}
	if len(perms1) != 105 {
		t.Fatalf("synced permission count = %d, want 105", len(perms1))
	}
	if len(roles1) != 12 {
		t.Fatalf("synced role count = %d, want 12", len(roles1))
	}

	store.mu.Lock()
	store.bumpedRoleIDs = nil
	store.outboxEvents = nil
	store.mu.Unlock()

	perms2, roles2, err := syncEmbeddedRBACCatalogs(ctx, repos, reg, nil)
	if err != nil {
		t.Fatalf("syncEmbeddedRBACCatalogs(2) error = %v", err)
	}
	if len(perms2) != 105 || len(roles2) != 12 {
		t.Fatalf("second sync counts = (%d, %d), want (105, 12)", len(perms2), len(roles2))
	}
	if len(store.bumpedRoleIDs) != 0 || len(store.outboxEvents) != 0 {
		t.Fatalf("idempotent re-sync should trigger zero security bumps, got bumps=%v outbox=%d", store.bumpedRoleIDs, len(store.outboxEvents))
	}
}

func TestSyncRBACCatalogSet_SensitivityAndNoAutoDelete(t *testing.T) {
	t.Parallel()

	store, repos, reg := newSyncTestHarness(t, "/warehouse")
	ctx := context.Background()

	v1, err := rbac.ValidateCatalogSet([]rbac.CatalogManifest{
		{
			SchemaVersion:    1,
			ManagementDomain: "warehouse",
			Permissions: []rbac.PermissionSpec{
				{Code: "warehouse.bins.read", Resource: "warehouse.bins", Action: "read", Description: "Read bins v1", IsSensitive: false},
				{Code: "warehouse.legacy.read", Resource: "warehouse.legacy", Action: "read", Description: "Legacy bin read", IsSensitive: false},
			},
			Roles: []rbac.RoleSpec{
				{
					Code:                "warehouse_operator",
					Name:                "Warehouse Operator",
					Description:         "Operates bins",
					ManagementDomain:    "warehouse",
					DefaultRedirectPath: "/warehouse",
					RedirectPriority:    30,
					IsSystem:            false,
					PermissionCodes:     []string{"warehouse.bins.read"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidateCatalogSet(v1) error = %v", err)
	}
	if _, _, err := syncRBACCatalogSet(ctx, repos, reg, v1, nil); err != nil {
		t.Fatalf("syncRBACCatalogSet(v1) error = %v", err)
	}

	opRole, err := repos.Roles.SelectByCode(ctx, "warehouse_operator")
	if err != nil {
		t.Fatalf("SelectByCode(warehouse_operator) error = %v", err)
	}

	store.mu.Lock()
	store.bumpedRoleIDs = nil
	store.outboxEvents = nil
	store.mu.Unlock()

	// V2 memperketat is_sensitive false -> true, memperbarui deskripsi, dan menghapus entri legacy dari manifest.
	v2, err := rbac.ValidateCatalogSet([]rbac.CatalogManifest{
		{
			SchemaVersion:    1,
			ManagementDomain: "warehouse",
			Permissions: []rbac.PermissionSpec{
				{Code: "warehouse.bins.read", Resource: "warehouse.bins", Action: "read", Description: "Read bins v2", IsSensitive: true},
			},
			Roles: v1.Manifests[0].Roles,
		},
	})
	if err != nil {
		t.Fatalf("ValidateCatalogSet(v2) error = %v", err)
	}
	if _, _, err := syncRBACCatalogSet(ctx, repos, reg, v2, nil); err != nil {
		t.Fatalf("syncRBACCatalogSet(v2) error = %v", err)
	}

	upgraded, err := repos.Permissions.SelectByCode(ctx, "warehouse.bins.read")
	if err != nil || !upgraded.IsSensitive || upgraded.Description != "Read bins v2" {
		t.Fatalf("updated permission = (%+v, %v), want IsSensitive=true and Description='Read bins v2'", upgraded, err)
	}
	if !slices.Contains(store.bumpedRoleIDs, opRole.ID) {
		t.Fatalf("expected BumpPermRevisionForRole for holding role %s on sensitivity upgrade, got %v", opRole.ID, store.bumpedRoleIDs)
	}
	// Izin legacy yang dihapus dari manifest wajib TIDAK terhapus otomatis dari DB.
	if _, err := repos.Permissions.SelectByCode(ctx, "warehouse.legacy.read"); err != nil {
		t.Fatalf("legacy permission must never be auto-deleted on manifest removal, got %v", err)
	}

	// Penurunan sensitivitas true -> false wajib gagal tertutup.
	if _, _, err := syncRBACCatalogSet(ctx, repos, reg, v1, nil); err == nil {
		t.Fatal("expected sensitivity downgrade true -> false to fail closed")
	}
}

func TestSyncRBACCatalogSet_SystemReconcileVsSeedOnceAndRejectsUnlistedRedirect(t *testing.T) {
	t.Parallel()

	store, repos, reg := newSyncTestHarness(t, "/warehouse")
	ctx := context.Background()

	v1, err := rbac.ValidateCatalogSet([]rbac.CatalogManifest{
		{
			SchemaVersion:    1,
			ManagementDomain: "warehouse",
			Permissions: []rbac.PermissionSpec{
				{Code: "warehouse.bins.read", Resource: "warehouse.bins", Action: "read", Description: "Read bins"},
				{Code: "warehouse.bins.write", Resource: "warehouse.bins", Action: "write", Description: "Write bins", IsSensitive: true},
			},
			Roles: []rbac.RoleSpec{
				{
					Code:                "warehouse_automation",
					Name:                "Warehouse Automation",
					Description:         "Automation v1",
					ManagementDomain:    "warehouse",
					DefaultRedirectPath: "",
					RedirectPriority:    999,
					IsSystem:            true,
					PermissionCodes:     []string{"warehouse.bins.read"},
				},
				{
					Code:                "warehouse_supervisor",
					Name:                "Warehouse Supervisor",
					Description:         "Supervisor v1",
					ManagementDomain:    "warehouse",
					DefaultRedirectPath: "/warehouse",
					RedirectPriority:    30,
					IsSystem:            false,
					PermissionCodes:     []string{"warehouse.bins.read"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidateCatalogSet(v1) error = %v", err)
	}
	if _, _, err := syncRBACCatalogSet(ctx, repos, reg, v1, nil); err != nil {
		t.Fatalf("syncRBACCatalogSet(v1) error = %v", err)
	}

	// Simulasikan kustomisasi admin perusahaan lewat /identity pada template non-sistem (is_system=false).
	supRole, err := repos.Roles.SelectByCode(ctx, "warehouse_supervisor")
	if err != nil {
		t.Fatal(err)
	}
	if err := supRole.UpdateDetails("Company Custom Supervisor", "Edited by company admin", "warehouse", "/warehouse", 15, time.Now()); err != nil {
		t.Fatalf("UpdateDetails() error = %v", err)
	}
	if err := repos.Roles.Update(ctx, supRole); err != nil {
		t.Fatal(err)
	}
	writePerm, err := repos.Permissions.SelectByCode(ctx, "warehouse.bins.write")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Roles.AssignPermission(ctx, supRole.ID, writePerm.ID); err != nil {
		t.Fatalf("AssignPermission() error = %v", err)
	}

	// Manifest V2 memperketat warehouse_automation (is_system=true) agar hanya pegang write,
	// sementara warehouse_supervisor tetap terdaftar di manifest dengan nama v2 yang tidak boleh menimpa DB.
	v2, err := rbac.ValidateCatalogSet([]rbac.CatalogManifest{
		{
			SchemaVersion:    1,
			ManagementDomain: "warehouse",
			Permissions:      v1.Manifests[0].Permissions,
			Roles: []rbac.RoleSpec{
				{
					Code:                "warehouse_automation",
					Name:                "Warehouse Automation v2",
					Description:         "Automation v2",
					ManagementDomain:    "warehouse",
					DefaultRedirectPath: "",
					RedirectPriority:    999,
					IsSystem:            true,
					PermissionCodes:     []string{"warehouse.bins.write"},
				},
				{
					Code:                "warehouse_supervisor",
					Name:                "Warehouse Supervisor v2",
					Description:         "Should not overwrite existing admin customization",
					ManagementDomain:    "warehouse",
					DefaultRedirectPath: "/warehouse",
					RedirectPriority:    50,
					IsSystem:            false,
					PermissionCodes:     []string{"warehouse.bins.read"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidateCatalogSet(v2) error = %v", err)
	}
	if _, _, err := syncRBACCatalogSet(ctx, repos, reg, v2, nil); err != nil {
		t.Fatalf("syncRBACCatalogSet(v2) error = %v", err)
	}

	// Verifikasi seed-once mempertahankan seluruh kustomisasi peran non-sistem.
	supAfter, err := repos.Roles.SelectByCode(ctx, "warehouse_supervisor")
	if err != nil {
		t.Fatal(err)
	}
	if supAfter.Name != "Company Custom Supervisor" || supAfter.RedirectPriority != 15 {
		t.Fatalf("seed-once role was overwritten: got %+v", supAfter)
	}
	supPerms, err := repos.Permissions.SelectByRoleID(ctx, supAfter.ID)
	if err != nil || len(supPerms) != 2 {
		t.Fatalf("seed-once role permissions were overwritten: got (%d, %v), want 2", len(supPerms), err)
	}

	// Verifikasi system role tersinkron penuh (metadata berubah, read dicabut, write terpasang).
	autoAfter, err := repos.Roles.SelectByCode(ctx, "warehouse_automation")
	if err != nil {
		t.Fatal(err)
	}
	if autoAfter.Name != "Warehouse Automation v2" || autoAfter.Description != "Automation v2" {
		t.Fatalf("system role metadata mismatch: %+v", autoAfter)
	}
	autoPerms, err := repos.Permissions.SelectByRoleID(ctx, autoAfter.ID)
	if err != nil || len(autoPerms) != 1 || autoPerms[0].Code != "warehouse.bins.write" {
		t.Fatalf("system role permissions mismatch: got (%+v, %v), want [warehouse.bins.write]", autoPerms, err)
	}

	// Tolak manifest dengan redirect path di luar allowlist IAM.
	badRedirect, err := rbac.ValidateCatalogSet([]rbac.CatalogManifest{
		{
			SchemaVersion:    1,
			ManagementDomain: "warehouse",
			Permissions:      v1.Manifests[0].Permissions,
			Roles: []rbac.RoleSpec{
				{
					Code:                "warehouse_rogue",
					Name:                "Warehouse Rogue",
					Description:         "Unregistered route",
					ManagementDomain:    "warehouse",
					DefaultRedirectPath: "/unregistered_route",
					RedirectPriority:    40,
					IsSystem:            false,
					PermissionCodes:     []string{"warehouse.bins.read"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ValidateCatalogSet(badRedirect) error = %v", err)
	}
	if _, _, err := syncRBACCatalogSet(ctx, repos, reg, badRedirect, nil); err == nil {
		t.Fatal("expected syncRBACCatalogSet to reject role with redirect outside RedirectResolver allowlist")
	}
	_ = store
}
