package group_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/group"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type passthroughTx struct{}

func (t passthroughTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type mockGroupStore struct {
	mu         sync.Mutex
	groups     map[string]*entity.Group
	members    map[string]map[string]*entity.GroupMember
	groupRoles map[string]*entity.GroupRole
	roles      map[string]*entity.Role
	users      map[string]*entity.User
}

type mockSecurityRuntime struct {
	domainrepo.SecurityRuntimeRepository
	store *mockGroupStore
}

func (m mockSecurityRuntime) Insert(context.Context, *entity.OutboxEvent) error      { return nil }
func (m mockSecurityRuntime) MarkEffective(context.Context, string, time.Time) error { return nil }

func (m mockSecurityRuntime) BumpPermRevisionForGroup(_ context.Context, groupID string) ([]string, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	var userIDs []string
	for userID := range m.store.members[groupID] {
		userIDs = append(userIDs, userID)
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

func (m mockSecurityRuntime) ResolveGroupRoleGroupID(_ context.Context, assignmentID string) (string, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	assignment, ok := m.store.groupRoles[assignmentID]
	if !ok {
		return "", iamerrors.ErrGroupRoleNotFound
	}
	return assignment.GroupID, nil
}

func newMockGroupStore() *mockGroupStore {
	return &mockGroupStore{
		groups:     make(map[string]*entity.Group),
		members:    make(map[string]map[string]*entity.GroupMember),
		groupRoles: make(map[string]*entity.GroupRole),
		roles:      make(map[string]*entity.Role),
		users:      make(map[string]*entity.User),
	}
}

func (m *mockGroupStore) SelectByID(_ context.Context, id string) (*entity.Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[id]
	if !ok {
		return nil, iamerrors.ErrGroupNotFound
	}
	cp := *g
	return &cp, nil
}

func (m *mockGroupStore) SelectByCode(_ context.Context, code string) (*entity.Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := entity.NormalizeCode(code)
	for _, g := range m.groups {
		if g.Code == norm {
			cp := *g
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrGroupNotFound
}

func (m *mockGroupStore) SelectList(_ context.Context, filter domainrepo.GroupFilter) (domainrepo.PageResult[*entity.Group], error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := filter.Normalize()
	var items []*entity.Group
	for _, g := range m.groups {
		if norm.ActiveOnly && !g.IsActive {
			continue
		}
		cp := *g
		items = append(items, &cp)
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm.Page), nil
}

func (m *mockGroupStore) Insert(_ context.Context, g *entity.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.groups {
		if existing.Code == g.Code {
			return iamerrors.ErrGroupAlreadyExists
		}
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}

func (m *mockGroupStore) Update(_ context.Context, g *entity.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[g.ID]; !ok {
		return iamerrors.ErrGroupNotFound
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}

func (m *mockGroupStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[id]; !ok {
		return iamerrors.ErrGroupNotFound
	}
	delete(m.groups, id)
	return nil
}

type mockMemberRepo struct{ store *mockGroupStore }

func (mr mockMemberRepo) SelectUsersByGroupID(_ context.Context, groupID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.User], error) {
	mr.store.mu.Lock()
	defer mr.store.mu.Unlock()
	norm := page.Normalize()
	var items []*entity.User
	for uid := range mr.store.members[groupID] {
		if u, ok := mr.store.users[uid]; ok {
			cp := *u
			items = append(items, &cp)
		}
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm), nil
}

func (mr mockMemberRepo) SelectGroupsByUserID(_ context.Context, userID string) ([]*entity.Group, error) {
	mr.store.mu.Lock()
	defer mr.store.mu.Unlock()
	var out []*entity.Group
	for gid, members := range mr.store.members {
		if _, ok := members[userID]; ok {
			if g, ok := mr.store.groups[gid]; ok {
				cp := *g
				out = append(out, &cp)
			}
		}
	}
	return out, nil
}

func (mr mockMemberRepo) CheckExists(_ context.Context, groupID, userID string) (bool, error) {
	mr.store.mu.Lock()
	defer mr.store.mu.Unlock()
	_, ok := mr.store.members[groupID][userID]
	return ok, nil
}

func (mr mockMemberRepo) Insert(_ context.Context, mem *entity.GroupMember) error {
	mr.store.mu.Lock()
	defer mr.store.mu.Unlock()
	if mr.store.members[mem.GroupID] == nil {
		mr.store.members[mem.GroupID] = make(map[string]*entity.GroupMember)
	}
	cp := *mem
	mr.store.members[mem.GroupID][mem.UserID] = &cp
	return nil
}

func (mr mockMemberRepo) Delete(_ context.Context, groupID, userID string) error {
	mr.store.mu.Lock()
	defer mr.store.mu.Unlock()
	if _, ok := mr.store.members[groupID][userID]; !ok {
		return iamerrors.ErrGroupMemberNotFound
	}
	delete(mr.store.members[groupID], userID)
	return nil
}

type mockGroupRoleRepo struct{ store *mockGroupStore }

func (gr mockGroupRoleRepo) SelectByGroupID(_ context.Context, groupID string, includeExpired bool, now time.Time) ([]*entity.GroupRole, error) {
	gr.store.mu.Lock()
	defer gr.store.mu.Unlock()
	var out []*entity.GroupRole
	for _, item := range gr.store.groupRoles {
		if item.GroupID != groupID {
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

func (gr mockGroupRoleRepo) CheckExists(_ context.Context, groupID, roleID string, scope entity.Scope) (bool, error) {
	gr.store.mu.Lock()
	defer gr.store.mu.Unlock()
	for _, item := range gr.store.groupRoles {
		if item.GroupID == groupID && item.RoleID == roleID && item.Scope.Type == scope.Type {
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

func (gr mockGroupRoleRepo) Insert(_ context.Context, a *entity.GroupRole) error {
	gr.store.mu.Lock()
	defer gr.store.mu.Unlock()
	cp := *a
	gr.store.groupRoles[a.ID] = &cp
	return nil
}

func (gr mockGroupRoleRepo) DeleteByID(_ context.Context, id string) error {
	gr.store.mu.Lock()
	defer gr.store.mu.Unlock()
	if _, ok := gr.store.groupRoles[id]; !ok {
		return iamerrors.ErrGroupRoleNotFound
	}
	delete(gr.store.groupRoles, id)
	return nil
}

type mockRoleRepo struct{ store *mockGroupStore }

func (r mockRoleRepo) SelectByID(_ context.Context, id string) (*entity.Role, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	item, ok := r.store.roles[id]
	if !ok {
		return nil, iamerrors.ErrRoleNotFound
	}
	cp := *item
	return &cp, nil
}
func (r mockRoleRepo) SelectByCode(_ context.Context, _ string) (*entity.Role, error) {
	return nil, iamerrors.ErrRoleNotFound
}
func (r mockRoleRepo) SelectList(_ context.Context, _ domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	return domainrepo.PageResult[*entity.Role]{}, nil
}
func (r mockRoleRepo) SelectEffectiveByUserID(_ context.Context, _ string, _ time.Time) ([]entity.Role, error) {
	return nil, nil
}
func (r mockRoleRepo) Insert(_ context.Context, _ *entity.Role) error { return nil }
func (r mockRoleRepo) Update(_ context.Context, _ *entity.Role) error { return nil }
func (r mockRoleRepo) Delete(_ context.Context, _ string) error       { return nil }

type mockUserRepo struct{ store *mockGroupStore }

func (u mockUserRepo) SelectByID(_ context.Context, id string) (*entity.User, error) {
	u.store.mu.Lock()
	defer u.store.mu.Unlock()
	item, ok := u.store.users[id]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	cp := *item
	return &cp, nil
}
func (u mockUserRepo) SelectList(_ context.Context, _ domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	return domainrepo.PageResult[*entity.User]{}, nil
}
func (u mockUserRepo) SelectAccessSnapshotByUserID(_ context.Context, _ string, _ time.Time) (*entity.AccessSnapshot, error) {
	return nil, iamerrors.ErrUserNotFound
}
func (u mockUserRepo) Insert(_ context.Context, _ *entity.User) error { return nil }
func (u mockUserRepo) InsertWithIdentities(_ context.Context, _ *entity.User, _ []*entity.UserIdentity) error {
	return nil
}
func (u mockUserRepo) Update(_ context.Context, _ *entity.User) error                 { return nil }
func (u mockUserRepo) UpdateLastLogin(_ context.Context, _ string, _ time.Time) error { return nil }
func (u mockUserRepo) SoftDelete(_ context.Context, _ string, _ time.Time) error      { return nil }

func TestGroupService_LifecycleMembershipAndRoles(t *testing.T) {
	t.Parallel()

	now := time.Now()
	idGen := crypto.NewUUIDGenerator()
	store := newMockGroupStore()

	svc, err := group.NewService(group.Dependencies{
		Transactor:      passthroughTx{},
		Groups:          store,
		GroupMembers:    mockMemberRepo{store: store},
		GroupRoles:      mockGroupRoleRepo{store: store},
		Roles:           mockRoleRepo{store: store},
		Users:           mockUserRepo{store: store},
		SecurityRuntime: mockSecurityRuntime{store: store},
		OutboxCoord: redisstore.NewCoordinator(
			mockSecurityRuntime{store: store}, nil, idGen, fixedClock{now: now}, nil,
		),
		Clock: fixedClock{now: now},
		IDGen: idGen,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	userID, _ := idGen.NewUUID()
	u, _ := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Warehouse Lead", now)
	store.users[userID] = u

	roleID, _ := idGen.NewUUID()
	r, _ := entity.NewRole(roleID, "wms_lead", "WMS Lead", "Leads warehouse", "wms", "/warehouse", 15, false, now)
	store.roles[roleID] = r

	grp, err := svc.Create(context.Background(), group.CreateInput{
		Code:        "WMS_Operators",
		Name:        "Warehouse Operators",
		Description: "Default warehouse team",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if grp.Code != "wms_operators" {
		t.Fatalf("expected normalized code wms_operators, got %q", grp.Code)
	}

	if _, err := svc.Create(context.Background(), group.CreateInput{
		Code: "wms_operators",
		Name: "Dup",
	}); !errors.Is(err, iamerrors.ErrGroupAlreadyExists) {
		t.Fatalf("expected ErrGroupAlreadyExists, got %v", err)
	}

	// Tambah anggota dan peran berscope facility ke grup.
	actorCtx := authcontext.WithCaller(context.Background(), authcontext.CallerContext{PrincipalID: userID, Authenticated: true})
	member, err := svc.AddMember(actorCtx, grp.ID, userID)
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}
	if member.AddedBy == nil || *member.AddedBy != userID {
		t.Fatalf("AddMember() added_by = %v, want %s", member.AddedBy, userID)
	}
	if _, err := svc.AddMember(context.Background(), grp.ID, userID); !errors.Is(err, iamerrors.ErrGroupMemberExists) {
		t.Fatalf("expected ErrGroupMemberExists, got %v", err)
	}

	facID, _ := idGen.NewUUID()
	if _, err := svc.AssignRole(actorCtx, group.AssignGroupRoleInput{
		GroupID: grp.ID, RoleID: roleID, ScopeType: entity.ScopeTypeFacility, ScopeID: &facID,
	}); err == nil {
		t.Fatal("external facility scope must fail without owner verification")
	}
	gr, err := svc.AssignRole(actorCtx, group.AssignGroupRoleInput{
		GroupID:   grp.ID,
		RoleID:    roleID,
		ScopeType: entity.ScopeTypeGlobal,
		GrantedBy: &facID,
	})
	if err != nil {
		t.Fatalf("AssignRole() error = %v", err)
	}
	if gr.GrantedBy == nil || *gr.GrantedBy != userID {
		t.Fatalf("AssignRole() granted_by = %v, want caller %s", gr.GrantedBy, userID)
	}

	roles, err := svc.ListGroupRoles(context.Background(), grp.ID, false)
	if err != nil || len(roles) != 1 || roles[0].ID != gr.ID {
		t.Fatalf("ListGroupRoles() = (%+v, %v)", roles, err)
	}

	members, err := svc.ListMembers(context.Background(), grp.ID, domainrepo.PageRequest{})
	if err != nil || len(members.Items) != 1 || members.Items[0].ID != userID {
		t.Fatalf("ListMembers() = (%+v, %v)", members, err)
	}

	// Nonaktifkan grup -> anggota baru & peran baru ditolak fail-closed.
	if _, err := svc.Deactivate(context.Background(), grp.ID); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	u2ID, _ := idGen.NewUUID()
	store.users[u2ID], _ = entity.NewUser(u2ID, entity.UserKindEmployee, entity.UserStatusActive, "U2", now)
	if _, err := svc.AddMember(context.Background(), grp.ID, u2ID); !errors.Is(err, iamerrors.ErrGroupInactive) {
		t.Fatalf("expected ErrGroupInactive on AddMember, got %v", err)
	}
	if _, err := svc.AssignRole(context.Background(), group.AssignGroupRoleInput{
		GroupID:   grp.ID,
		RoleID:    roleID,
		ScopeType: entity.ScopeTypeGlobal,
	}); !errors.Is(err, iamerrors.ErrGroupInactive) {
		t.Fatalf("expected ErrGroupInactive on AssignRole, got %v", err)
	}

	if err := svc.RevokeRole(context.Background(), gr.ID); err != nil {
		t.Fatalf("RevokeRole() error = %v", err)
	}
	if err := svc.RemoveMember(context.Background(), grp.ID, userID); err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
	if err := svc.Delete(context.Background(), grp.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestGroupService_RejectsMissingSecurityRuntime(t *testing.T) {
	t.Parallel()
	store := newMockGroupStore()
	_, err := group.NewService(group.Dependencies{
		Transactor:   passthroughTx{},
		Groups:       store,
		GroupMembers: mockMemberRepo{store: store},
		GroupRoles:   mockGroupRoleRepo{store: store},
		Roles:        mockRoleRepo{store: store},
		Users:        mockUserRepo{store: store},
		IDGen:        crypto.NewUUIDGenerator(),
	})
	if err == nil {
		t.Fatal("expected error without security runtime repository")
	}
	_, err = group.NewService(group.Dependencies{
		Transactor:      passthroughTx{},
		Groups:          store,
		GroupMembers:    mockMemberRepo{store: store},
		GroupRoles:      mockGroupRoleRepo{store: store},
		Roles:           mockRoleRepo{store: store},
		Users:           mockUserRepo{store: store},
		SecurityRuntime: mockSecurityRuntime{store: store},
		IDGen:           crypto.NewUUIDGenerator(),
	})
	if err == nil {
		t.Fatal("expected error without outbox coordinator")
	}
}
