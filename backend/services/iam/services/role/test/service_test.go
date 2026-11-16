package role_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type passthroughTx struct {
	mu    sync.Mutex
	calls int
}

func (t *passthroughTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	t.mu.Lock()
	t.calls++
	t.mu.Unlock()
	return fn(ctx)
}

type mockRoleStore struct {
	mu       sync.Mutex
	roles    map[string]*entity.Role
	perms    map[string]*entity.Permission
	rolePerm map[string]map[string]*entity.RolePermission
	users    map[string]*entity.User
	userRole map[string]*entity.UserRole
}

type mockSecurityRuntime struct {
	domainrepo.SecurityRuntimeRepository
	store *mockRoleStore
}

func (m mockSecurityRuntime) Insert(context.Context, *entity.OutboxEvent) error      { return nil }
func (m mockSecurityRuntime) MarkEffective(context.Context, string, time.Time) error { return nil }

func newMockCoordinator(store *mockRoleStore, idGen domainservice.IDGenerator, clock domainservice.Clock) *redisstore.Coordinator {
	return redisstore.NewCoordinator(mockSecurityRuntime{store: store}, nil, idGen, clock, nil)
}

func (m mockSecurityRuntime) BumpPermRevisionForRole(_ context.Context, roleID string) ([]string, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	var userIDs []string
	for _, assignment := range m.store.userRole {
		if assignment.RoleID == roleID {
			userIDs = append(userIDs, assignment.UserID)
		}
	}
	return userIDs, nil
}

func (m mockSecurityRuntime) IncrementUserSecurityCounters(_ context.Context, userID string, _, _ bool) (int64, int64, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	if _, ok := m.store.users[userID]; !ok {
		return 0, 0, iamerrors.ErrUserNotFound
	}
	return 1, 1, nil
}

func (m mockSecurityRuntime) ResolveUserRoleUserID(_ context.Context, assignmentID string) (string, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	assignment, ok := m.store.userRole[assignmentID]
	if !ok {
		return "", iamerrors.ErrUserRoleNotFound
	}
	return assignment.UserID, nil
}

func newMockRoleStore() *mockRoleStore {
	return &mockRoleStore{
		roles:    make(map[string]*entity.Role),
		perms:    make(map[string]*entity.Permission),
		rolePerm: make(map[string]map[string]*entity.RolePermission),
		users:    make(map[string]*entity.User),
		userRole: make(map[string]*entity.UserRole),
	}
}

func (m *mockRoleStore) SelectByID(_ context.Context, id string) (*entity.Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.roles[id]
	if !ok {
		return nil, iamerrors.ErrRoleNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockRoleStore) SelectByCode(_ context.Context, code string) (*entity.Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := entity.NormalizeCode(code)
	for _, r := range m.roles {
		if r.Code == norm {
			cp := *r
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrRoleNotFound
}

func (m *mockRoleStore) SelectList(_ context.Context, filter domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := filter.Normalize()
	items := make([]*entity.Role, 0, len(m.roles))
	for _, r := range m.roles {
		if norm.ActiveOnly && !r.IsActive {
			continue
		}
		cp := *r
		items = append(items, &cp)
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm.Page), nil
}

func (m *mockRoleStore) SelectEffectiveByUserID(_ context.Context, userID string, now time.Time) ([]entity.Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []entity.Role
	for _, ur := range m.userRole {
		if ur.UserID == userID && ur.IsEffective(now) {
			if r, ok := m.roles[ur.RoleID]; ok && r.IsActive {
				out = append(out, *r)
			}
		}
	}
	return out, nil
}

func (m *mockRoleStore) Insert(_ context.Context, r *entity.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.roles {
		if existing.Code == r.Code {
			return iamerrors.ErrRoleAlreadyExists
		}
	}
	cp := *r
	m.roles[r.ID] = &cp
	return nil
}

func (m *mockRoleStore) Update(_ context.Context, r *entity.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.roles[r.ID]; !ok {
		return iamerrors.ErrRoleNotFound
	}
	cp := *r
	m.roles[r.ID] = &cp
	return nil
}

func (m *mockRoleStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.roles[id]; !ok {
		return iamerrors.ErrRoleNotFound
	}
	delete(m.roles, id)
	return nil
}

type mockPermRepo struct{ store *mockRoleStore }

func (p mockPermRepo) SelectByID(_ context.Context, id string) (*entity.Permission, error) {
	p.store.mu.Lock()
	defer p.store.mu.Unlock()
	perm, ok := p.store.perms[id]
	if !ok {
		return nil, iamerrors.ErrPermissionNotFound
	}
	cp := *perm
	return &cp, nil
}
func (p mockPermRepo) SelectByCode(_ context.Context, code string) (*entity.Permission, error) {
	return nil, iamerrors.ErrPermissionNotFound
}
func (p mockPermRepo) SelectList(_ context.Context, _ domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	return domainrepo.PageResult[*entity.Permission]{}, nil
}
func (p mockPermRepo) SelectByRoleID(_ context.Context, roleID string) ([]*entity.Permission, error) {
	p.store.mu.Lock()
	defer p.store.mu.Unlock()
	var out []*entity.Permission
	for permID := range p.store.rolePerm[roleID] {
		if perm, ok := p.store.perms[permID]; ok {
			cp := *perm
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (p mockPermRepo) Insert(_ context.Context, perm *entity.Permission) error {
	p.store.mu.Lock()
	defer p.store.mu.Unlock()
	p.store.perms[perm.ID] = perm
	return nil
}
func (p mockPermRepo) Update(_ context.Context, perm *entity.Permission) error {
	p.store.mu.Lock()
	defer p.store.mu.Unlock()
	p.store.perms[perm.ID] = perm
	return nil
}
func (p mockPermRepo) Delete(_ context.Context, id string) error { return nil }

type mockRolePermRepo struct{ store *mockRoleStore }

func (rp mockRolePermRepo) CheckExists(_ context.Context, roleID, permID string) (bool, error) {
	rp.store.mu.Lock()
	defer rp.store.mu.Unlock()
	_, ok := rp.store.rolePerm[roleID][permID]
	return ok, nil
}
func (rp mockRolePermRepo) Insert(_ context.Context, rel *entity.RolePermission) error {
	rp.store.mu.Lock()
	defer rp.store.mu.Unlock()
	if rp.store.rolePerm[rel.RoleID] == nil {
		rp.store.rolePerm[rel.RoleID] = make(map[string]*entity.RolePermission)
	}
	cp := *rel
	rp.store.rolePerm[rel.RoleID][rel.PermissionID] = &cp
	return nil
}
func (rp mockRolePermRepo) Delete(_ context.Context, roleID, permID string) error {
	rp.store.mu.Lock()
	defer rp.store.mu.Unlock()
	if _, ok := rp.store.rolePerm[roleID][permID]; !ok {
		return iamerrors.ErrRolePermissionNotFound
	}
	delete(rp.store.rolePerm[roleID], permID)
	return nil
}

type mockUserRepo struct{ store *mockRoleStore }

func (u mockUserRepo) SelectByID(_ context.Context, id string) (*entity.User, error) {
	u.store.mu.Lock()
	defer u.store.mu.Unlock()
	usr, ok := u.store.users[id]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	cp := *usr
	return &cp, nil
}
func (u mockUserRepo) SelectList(_ context.Context, _ domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	return domainrepo.PageResult[*entity.User]{}, nil
}
func (u mockUserRepo) SelectAccessSnapshotByUserID(_ context.Context, _ string, _ time.Time) (*entity.AccessSnapshot, error) {
	return nil, iamerrors.ErrUserNotFound
}
func (u mockUserRepo) Insert(_ context.Context, usr *entity.User) error { return nil }
func (u mockUserRepo) InsertWithIdentities(_ context.Context, _ *entity.User, _ []*entity.UserIdentity) error {
	return nil
}
func (u mockUserRepo) Update(_ context.Context, _ *entity.User) error                 { return nil }
func (u mockUserRepo) UpdateLastLogin(_ context.Context, _ string, _ time.Time) error { return nil }
func (u mockUserRepo) SoftDelete(_ context.Context, _ string, _ time.Time) error      { return nil }

type mockUserRoleRepo struct{ store *mockRoleStore }

func (ur mockUserRoleRepo) SelectByUserID(_ context.Context, userID string, includeExpired bool, now time.Time) ([]*entity.UserRole, error) {
	ur.store.mu.Lock()
	defer ur.store.mu.Unlock()
	var out []*entity.UserRole
	for _, item := range ur.store.userRole {
		if item.UserID != userID {
			continue
		}
		if !includeExpired && item.IsExpired(now) {
			continue
		}
		cp := *item
		out = append(out, &cp)
	}
	return out, nil
}
func (ur mockUserRoleRepo) CheckExists(_ context.Context, userID, roleID string, scope entity.Scope) (bool, error) {
	ur.store.mu.Lock()
	defer ur.store.mu.Unlock()
	for _, item := range ur.store.userRole {
		if item.UserID == userID && item.RoleID == roleID && item.Scope.Type == scope.Type {
			if scope.ID == nil && item.Scope.ID == nil {
				return true, nil
			}
			if scope.ID != nil && item.Scope.ID != nil && *scope.ID == *item.Scope.ID {
				return true, nil
			}
		}
	}
	return false, nil
}
func (ur mockUserRoleRepo) Insert(_ context.Context, a *entity.UserRole) error {
	ur.store.mu.Lock()
	defer ur.store.mu.Unlock()
	cp := *a
	ur.store.userRole[a.ID] = &cp
	return nil
}
func (ur mockUserRoleRepo) DeleteByID(_ context.Context, id string) error {
	ur.store.mu.Lock()
	defer ur.store.mu.Unlock()
	if _, ok := ur.store.userRole[id]; !ok {
		return iamerrors.ErrUserRoleNotFound
	}
	delete(ur.store.userRole, id)
	return nil
}

func TestRoleService_LifecyclePermissionsAndUserAssignment(t *testing.T) {
	t.Parallel()

	now := time.Now()
	idGen := crypto.NewUUIDGenerator()
	store := newMockRoleStore()
	tx := &passthroughTx{}
	resolver, err := domainservice.NewRedirectResolver([]string{"/identity", "/dashboard"})
	if err != nil {
		t.Fatalf("NewRedirectResolver() error = %v", err)
	}

	svc, err := role.NewService(role.Dependencies{
		Transactor:      tx,
		Roles:           store,
		RolePermissions: mockRolePermRepo{store: store},
		Permissions:     mockPermRepo{store: store},
		Users:           mockUserRepo{store: store},
		UserRoles:       mockUserRoleRepo{store: store},
		SecurityRuntime: mockSecurityRuntime{store: store},
		OutboxCoord:     newMockCoordinator(store, idGen, fixedClock{now: now}),
		Clock:           fixedClock{now: now},
		IDGen:           idGen,
		Redirect:        resolver,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	// Buat izin & user dummy terlebih dahulu.
	permID, _ := idGen.NewUUID()
	perm, _ := entity.NewPermission(permID, "iam.roles.write", "iam.roles", "write", "Manage roles", true, now)
	store.perms[permID] = perm

	userID, _ := idGen.NewUUID()
	u, _ := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "IAM Admin", now)
	store.users[userID] = u

	// Reject redirect path di luar allowlist.
	if _, err := svc.Create(context.Background(), role.CreateInput{
		Code:                "bad_route",
		Name:                "Bad Route Role",
		DefaultRedirectPath: "/not-in-allowlist",
		RedirectPriority:    10,
	}); !errors.Is(err, iamerrors.ErrRedirectPathInvalid) {
		t.Fatalf("expected ErrRedirectPathInvalid for disallowed redirect route, got %v", err)
	}

	// Buat peran kustom valid.
	created, err := svc.Create(context.Background(), role.CreateInput{
		Code:                "iam_operator",
		Name:                "IAM Operator",
		Description:         "Operates IAM module",
		ManagementDomain:    "iam",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    20,
		IsSystem:            false,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Pasang izin ke peran dan verifikasi daftar izin.
	if _, err := svc.AssignPermission(context.Background(), created.ID, permID); err != nil {
		t.Fatalf("AssignPermission() error = %v", err)
	}
	if _, err := svc.AssignPermission(context.Background(), created.ID, permID); !errors.Is(err, iamerrors.ErrRolePermissionExists) {
		t.Fatalf("expected ErrRolePermissionExists on duplicate attach, got %v", err)
	}
	perms, err := svc.ListPermissions(context.Background(), created.ID)
	if err != nil || len(perms) != 1 || perms[0].ID != permID {
		t.Fatalf("ListPermissions() = (%+v, %v)", perms, err)
	}

	// Tugaskan peran ke pengguna secara global dan periksa effective roles.
	actorCtx := authcontext.WithCaller(context.Background(), authcontext.CallerContext{PrincipalID: userID, Authenticated: true})
	assignment, err := svc.AssignToUser(actorCtx, role.AssignUserRoleInput{
		UserID:    userID,
		RoleID:    created.ID,
		ScopeType: entity.ScopeTypeGlobal,
		GrantedBy: &permID,
	})
	if err != nil {
		t.Fatalf("AssignToUser() error = %v", err)
	}
	if assignment.GrantedBy == nil || *assignment.GrantedBy != userID {
		t.Fatalf("AssignToUser() granted_by = %v, want caller %s", assignment.GrantedBy, userID)
	}
	eff, err := svc.ListEffectiveRoles(context.Background(), userID)
	if err != nil || len(eff) != 1 || eff[0].ID != created.ID {
		t.Fatalf("ListEffectiveRoles() = (%+v, %v)", eff, err)
	}

	// Nonaktifkan peran -> pengguna tidak lagi mendapat peran aktif tersebut.
	if _, err := svc.Deactivate(context.Background(), created.ID); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	effAfter, err := svc.ListEffectiveRoles(context.Background(), userID)
	if err != nil || len(effAfter) != 0 {
		t.Fatalf("expected 0 effective roles after deactivate, got %+v err=%v", effAfter, err)
	}

	// Penugasan peran nonaktif ke user harus ditolak fail-closed.
	otherUserID, _ := idGen.NewUUID()
	u2, _ := entity.NewUser(otherUserID, entity.UserKindEmployee, entity.UserStatusActive, "Other", now)
	store.users[otherUserID] = u2
	if _, err := svc.AssignToUser(context.Background(), role.AssignUserRoleInput{
		UserID:    otherUserID,
		RoleID:    created.ID,
		ScopeType: entity.ScopeTypeGlobal,
	}); !errors.Is(err, iamerrors.ErrRoleInactive) {
		t.Fatalf("expected ErrRoleInactive, got %v", err)
	}

	// Cabut user role dan hapus peran non-sistem.
	if err := svc.RevokeUserRole(context.Background(), assignment.ID); err != nil {
		t.Fatalf("RevokeUserRole() error = %v", err)
	}
	if err := svc.RevokePermission(context.Background(), created.ID, permID); err != nil {
		t.Fatalf("RevokePermission() error = %v", err)
	}
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if tx.calls == 0 {
		t.Fatal("expected transactional writes to invoke Transactor")
	}
}

func TestRoleService_SystemRoleProtectedAgainstDeactivateAndDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	idGen := crypto.NewUUIDGenerator()
	store := newMockRoleStore()
	tx := &passthroughTx{}
	resolver := domainservice.NewDefaultRedirectResolver()

	svc, err := role.NewService(role.Dependencies{
		Transactor:      tx,
		Roles:           store,
		RolePermissions: mockRolePermRepo{store: store},
		Permissions:     mockPermRepo{store: store},
		Users:           mockUserRepo{store: store},
		UserRoles:       mockUserRoleRepo{store: store},
		SecurityRuntime: mockSecurityRuntime{store: store},
		OutboxCoord:     newMockCoordinator(store, idGen, fixedClock{now: now}),
		Clock:           fixedClock{now: now},
		IDGen:           idGen,
		Redirect:        resolver,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	sysRole, err := svc.Create(context.Background(), role.CreateInput{
		Code:                "super_admin",
		Name:                "Super Admin",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    100,
		IsSystem:            true,
	})
	if err != nil {
		t.Fatalf("Create(system) error = %v", err)
	}

	if _, err := svc.Deactivate(context.Background(), sysRole.ID); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on Deactivate, got %v", err)
	}
	if err := svc.Delete(context.Background(), sysRole.ID); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on Delete, got %v", err)
	}
}

func TestRoleService_SystemRoleImmutableMutationsAndBootstrap(t *testing.T) {
	t.Parallel()

	now := time.Now()
	idGen := crypto.NewUUIDGenerator()
	store := newMockRoleStore()
	tx := &passthroughTx{}
	resolver, err := domainservice.NewRedirectResolver([]string{"/identity"})
	if err != nil {
		t.Fatalf("NewRedirectResolver() error = %v", err)
	}

	svc, err := role.NewService(role.Dependencies{
		Transactor:      tx,
		Roles:           store,
		RolePermissions: mockRolePermRepo{store: store},
		Permissions:     mockPermRepo{store: store},
		Users:           mockUserRepo{store: store},
		UserRoles:       mockUserRoleRepo{store: store},
		SecurityRuntime: mockSecurityRuntime{store: store},
		OutboxCoord:     newMockCoordinator(store, idGen, fixedClock{now: now}),
		Clock:           fixedClock{now: now},
		IDGen:           idGen,
		Redirect:        resolver,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	permID, _ := idGen.NewUUID()
	perm, _ := entity.NewPermission(permID, "iam.users.read", "iam.users", "read", "Read users", false, now)
	store.perms[permID] = perm

	sysRole, err := svc.Create(context.Background(), role.CreateInput{
		Code:                "sys_immutable_guard",
		Name:                "System Guard",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    90,
		IsSystem:            true,
	})
	if err != nil {
		t.Fatalf("Create(system) error = %v", err)
	}

	if _, err := svc.UpdateDetails(context.Background(), role.UpdateInput{
		ID:                  sysRole.ID,
		Name:                "Mutated Name",
		Description:         "should fail",
		ManagementDomain:    "iam",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    10,
	}); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on UpdateDetails, got %v", err)
	}
	if _, err := svc.AssignPermission(context.Background(), sysRole.ID, permID); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on AssignPermission, got %v", err)
	}
	if _, err := svc.BootstrapAssignPermission(context.Background(), sysRole.ID, permID); err != nil {
		t.Fatalf("BootstrapAssignPermission() error = %v", err)
	}
	if err := svc.RevokePermission(context.Background(), sysRole.ID, permID); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on RevokePermission, got %v", err)
	}
	if _, err := svc.Deactivate(context.Background(), sysRole.ID); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on Deactivate, got %v", err)
	}
	if err := svc.Delete(context.Background(), sysRole.ID); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("expected ErrSystemRoleImmutable on Delete, got %v", err)
	}
}

func TestRoleService_RejectsMissingSecurityRuntime(t *testing.T) {
	t.Parallel()
	store := newMockRoleStore()
	_, err := role.NewService(role.Dependencies{
		Transactor:      &passthroughTx{},
		Roles:           store,
		RolePermissions: mockRolePermRepo{store: store},
		Permissions:     mockPermRepo{store: store},
		Users:           mockUserRepo{store: store},
		UserRoles:       mockUserRoleRepo{store: store},
		IDGen:           crypto.NewUUIDGenerator(),
	})
	if err == nil {
		t.Fatal("expected error without security runtime repository")
	}
	_, err = role.NewService(role.Dependencies{
		Transactor:      &passthroughTx{},
		Roles:           store,
		RolePermissions: mockRolePermRepo{store: store},
		Permissions:     mockPermRepo{store: store},
		Users:           mockUserRepo{store: store},
		UserRoles:       mockUserRoleRepo{store: store},
		SecurityRuntime: mockSecurityRuntime{store: store},
		IDGen:           crypto.NewUUIDGenerator(),
	})
	if err == nil {
		t.Fatal("expected error without outbox coordinator")
	}
}
