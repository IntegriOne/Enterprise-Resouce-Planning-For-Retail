package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/totp"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	"github.com/integrione/erp-retail/backend/services/iam/services/group"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	iamgrpc "github.com/integrione/erp-retail/backend/services/iam/transport/grpc"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

const bufSize = 1024 * 1024
const testGatewayHMACSecret = "test-gateway-proof-secret-with-32-bytes"

func withAuthenticatedCaller(ctx context.Context, userID string, sessionIDs ...string) context.Context {
	sessionID := "00000000-0000-4000-8000-000000000001"
	if len(sessionIDs) > 0 {
		sessionID = sessionIDs[0]
	}
	return authcontext.AppendToOutgoingGRPCContext(ctx, authcontext.CallerContext{
		PrincipalID:   userID,
		SessionID:     sessionID,
		Authenticated: true,
	}, []byte(testGatewayHMACSecret), time.Now())
}

// memStore menyimpan state repositori IAM di memori untuk pengujian gRPC bufconn.
type memStore struct {
	mu              sync.RWMutex
	permissions     map[string]*entity.Permission
	roles           map[string]*entity.Role
	rolePermissions map[string]*entity.RolePermission
	userRoles       map[string]*entity.UserRole
	groups          map[string]*entity.Group
	groupMembers    map[string]*entity.GroupMember
	groupRoles      map[string]*entity.GroupRole
	users           map[string]*entity.User
	identities      map[string]*entity.UserIdentity
	credentials     map[string]*entity.UserCredential
	totpFactors     map[string]*entity.UserTOTPFactor
	recoveryCodes   map[string]*entity.UserRecoveryCode
	tokens          map[string]*entity.UserToken
	sessions        map[string]*entity.Session
}

type memSecurityRuntime struct {
	store *memStore
}

func newMemCoordinator(store *memStore, idGen domainservice.IDGenerator, clock domainservice.Clock) *redisstore.Coordinator {
	return redisstore.NewCoordinator(memSecurityRuntime{store: store}, nil, idGen, clock, nil)
}

func (m memSecurityRuntime) GetRecoveryEpoch(context.Context) (int64, error)       { return 1, nil }
func (m memSecurityRuntime) IncrementRecoveryEpoch(context.Context) (int64, error) { return 2, nil }
func (m memSecurityRuntime) Insert(context.Context, *entity.OutboxEvent) error     { return nil }
func (m memSecurityRuntime) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (m memSecurityRuntime) MarkEffective(context.Context, string, time.Time) error { return nil }
func (m memSecurityRuntime) MarkRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (m memSecurityRuntime) MarkFailed(context.Context, string, string) error { return nil }
func (m memSecurityRuntime) CountPending(context.Context) (int64, error)      { return 0, nil }
func (m memSecurityRuntime) CountFailed(context.Context) (int64, error)       { return 0, nil }
func (m memSecurityRuntime) SelectActiveSessionDigestsByUserID(context.Context, string) ([][]byte, error) {
	return nil, nil
}
func (m memSecurityRuntime) SelectActiveSessionsForRecovery(context.Context, time.Time, int) ([]*entity.Session, error) {
	return nil, nil
}
func (m memSecurityRuntime) UpdateSessionSecurityMetadata(context.Context, string, int64, int, []byte) error {
	return nil
}

func (m memSecurityRuntime) BumpPermRevisionForRole(_ context.Context, roleID string) ([]string, error) {
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()
	var userIDs []string
	for _, assignment := range m.store.userRoles {
		if assignment.RoleID == roleID {
			userIDs = append(userIDs, assignment.UserID)
		}
	}
	return userIDs, nil
}

func (m memSecurityRuntime) BumpPermRevisionForGroup(_ context.Context, groupID string) ([]string, error) {
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()
	var userIDs []string
	for _, member := range m.store.groupMembers {
		if member.GroupID == groupID {
			userIDs = append(userIDs, member.UserID)
		}
	}
	return userIDs, nil
}

func (m memSecurityRuntime) IncrementUserSecurityCounters(_ context.Context, userID string, _, _ bool) (int64, int64, error) {
	m.store.mu.Lock()
	defer m.store.mu.Unlock()
	user, ok := m.store.users[userID]
	if !ok {
		return 0, 0, iamerrors.ErrUserNotFound
	}
	user.PermRevision++
	return user.AuthEpoch, user.PermRevision, nil
}

func (m memSecurityRuntime) ResolveUserRoleUserID(_ context.Context, assignmentID string) (string, error) {
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()
	assignment, ok := m.store.userRoles[assignmentID]
	if !ok {
		return "", iamerrors.ErrUserRoleNotFound
	}
	return assignment.UserID, nil
}

func (m memSecurityRuntime) ResolveGroupRoleGroupID(_ context.Context, assignmentID string) (string, error) {
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()
	assignment, ok := m.store.groupRoles[assignmentID]
	if !ok {
		return "", iamerrors.ErrGroupRoleNotFound
	}
	return assignment.GroupID, nil
}

func newMemStore() *memStore {
	return &memStore{
		permissions:     make(map[string]*entity.Permission),
		roles:           make(map[string]*entity.Role),
		rolePermissions: make(map[string]*entity.RolePermission),
		userRoles:       make(map[string]*entity.UserRole),
		groups:          make(map[string]*entity.Group),
		groupMembers:    make(map[string]*entity.GroupMember),
		groupRoles:      make(map[string]*entity.GroupRole),
		users:           make(map[string]*entity.User),
		identities:      make(map[string]*entity.UserIdentity),
		credentials:     make(map[string]*entity.UserCredential),
		totpFactors:     make(map[string]*entity.UserTOTPFactor),
		recoveryCodes:   make(map[string]*entity.UserRecoveryCode),
		tokens:          make(map[string]*entity.UserToken),
		sessions:        make(map[string]*entity.Session),
	}
}

type bufconnEnv struct {
	conn     *grpc.ClientConn
	store    *memStore
	registry *services.Registry
	totpProv *totp.RFC6238Provider
}

func startBufconnServer(t *testing.T) *bufconnEnv {
	t.Helper()

	store := newMemStore()
	tx := &memTransactor{}
	clock := domainservice.SystemClock{}
	idGen := crypto.NewUUIDGenerator()
	hasher := crypto.NewDefaultArgon2idHasher()
	tokenGen := crypto.NewOpaqueTokenGenerator()
	cipher, err := crypto.NewEphemeralAES256Cipher()
	if err != nil {
		t.Fatalf("NewEphemeralAES256Cipher() error = %v", err)
	}
	totpProv := totp.NewDefaultProvider()
	redirect := domainservice.NewDefaultRedirectResolver()
	sessPolicy := domainservice.NewDefaultSessionPolicy()
	evaluator := domainservice.NewAccessEvaluator(sessPolicy)

	permSvc, err := permission.NewService(&memPermRepo{store}, clock, idGen)
	if err != nil {
		t.Fatalf("permission.NewService() error = %v", err)
	}
	roleSvc, err := role.NewService(role.Dependencies{
		Transactor:      tx,
		Roles:           &memRoleRepo{store},
		RolePermissions: &memRolePermRepo{store},
		Permissions:     &memPermRepo{store},
		Users:           &memUserRepo{store},
		UserRoles:       &memUserRoleRepo{store},
		SecurityRuntime: memSecurityRuntime{store: store},
		OutboxCoord:     newMemCoordinator(store, idGen, clock),
		Clock:           clock,
		IDGen:           idGen,
		Redirect:        redirect,
	})
	if err != nil {
		t.Fatalf("role.NewService() error = %v", err)
	}
	groupSvc, err := group.NewService(group.Dependencies{
		Transactor:      tx,
		Groups:          &memGroupRepo{store},
		GroupMembers:    &memGroupMemberRepo{store},
		GroupRoles:      &memGroupRoleRepo{store},
		Roles:           &memRoleRepo{store},
		Users:           &memUserRepo{store},
		SecurityRuntime: memSecurityRuntime{store: store},
		OutboxCoord:     newMemCoordinator(store, idGen, clock),
		Clock:           clock,
		IDGen:           idGen,
	})
	if err != nil {
		t.Fatalf("group.NewService() error = %v", err)
	}
	sessSvc, err := session.NewService(session.Dependencies{
		Sessions:        &memSessionRepo{store},
		Users:           &memUserRepo{store},
		Clock:           clock,
		IDGen:           idGen,
		Tokens:          tokenGen,
		SessionPolicy:   sessPolicy,
		AccessEvaluator: evaluator,
		Redirect:        redirect,
	})
	if err != nil {
		t.Fatalf("session.NewService() error = %v", err)
	}
	userSvc, err := user.NewService(user.Dependencies{
		Transactor:     tx,
		Users:          &memUserRepo{store},
		Identities:     &memIdentityRepo{store},
		Credentials:    &memCredentialRepo{store},
		TOTPFactors:    &memTOTPRepo{store},
		RecoveryCodes:  &memRecoveryCodeRepo{store},
		Tokens:         &memTokenRepo{store},
		Roles:          &memRoleRepo{store},
		UserRoles:      &memUserRoleRepo{store},
		Sessions:       sessSvc,
		Clock:          clock,
		IDGen:          idGen,
		PasswordHasher: hasher,
		PasswordPolicy: domainservice.NewDefaultPasswordPolicy(),
		TokenGen:       tokenGen,
		SecretCipher:   cipher,
		TOTPProvider:   totpProv,
		Redirect:       redirect,
	})
	if err != nil {
		t.Fatalf("user.NewService() error = %v", err)
	}

	reg := &services.Registry{
		Permissions: permSvc,
		Roles:       roleSvc,
		Groups:      groupSvc,
		Sessions:    sessSvc,
		Users:       userSvc,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		AppEnv:                    "test",
		ServiceName:               "integrione-iam",
		ShutdownTimeout:           5 * time.Second,
		InternalGatewayHMACSecret: testGatewayHMACSecret,
	}
	srv := iamgrpc.NewServer(cfg, logger)
	if err := handler.RegisterAll(srv.GRPCServer(), reg, logger); err != nil {
		t.Fatalf("handler.RegisterAll() error = %v", err)
	}

	lis := bufconn.Listen(bufSize)
	go func() {
		_ = srv.GRPCServer().Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		srv.GRPCServer().Stop()
		t.Fatalf("grpc.NewClient(bufnet) error = %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop(context.Background())
	})

	return &bufconnEnv{
		conn:     conn,
		store:    store,
		registry: reg,
		totpProv: totpProv,
	}
}

type memTransactor struct{}

func (t *memTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func paginateSlice[T any](items []T, page domainrepo.PageRequest) domainrepo.PageResult[T] {
	p := page.Normalize()
	total := int64(len(items))
	start := min(p.Offset, len(items))
	end := min(start+p.Limit, len(items))
	return domainrepo.NewPageResult(items[start:end], total, p)
}

// --- PermissionRepository ---

type memPermRepo struct{ s *memStore }

func (r *memPermRepo) SelectByID(_ context.Context, id string) (*entity.Permission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.permissions[id]
	if !ok {
		return nil, iamerrors.ErrPermissionNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *memPermRepo) SelectByCode(_ context.Context, code string) (*entity.Permission, error) {
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

func (r *memPermRepo) SelectList(_ context.Context, filter domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.Permission
	for _, p := range r.s.permissions {
		if filter.Query != "" && !strings.Contains(strings.ToLower(p.Code+p.Description), strings.ToLower(filter.Query)) {
			continue
		}
		if filter.Resource != "" && p.Resource != filter.Resource {
			continue
		}
		if filter.IsSensitive != nil && p.IsSensitive != *filter.IsSensitive {
			continue
		}
		cp := *p
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.Permission) int { return strings.Compare(a.Code, b.Code) })
	return paginateSlice(out, filter.Page), nil
}

func (r *memPermRepo) SelectByRoleID(_ context.Context, roleID string) ([]*entity.Permission, error) {
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

func (r *memPermRepo) Insert(_ context.Context, p *entity.Permission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *p
	r.s.permissions[p.ID] = &cp
	return nil
}

func (r *memPermRepo) Update(_ context.Context, p *entity.Permission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.permissions[p.ID]; !ok {
		return iamerrors.ErrPermissionNotFound
	}
	cp := *p
	r.s.permissions[p.ID] = &cp
	return nil
}

func (r *memPermRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.permissions[id]; !ok {
		return iamerrors.ErrPermissionNotFound
	}
	delete(r.s.permissions, id)
	return nil
}

// --- RoleRepository, RolePermissionRepository, UserRoleRepository ---

type memRoleRepo struct{ s *memStore }

func (r *memRoleRepo) SelectByID(_ context.Context, id string) (*entity.Role, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	role, ok := r.s.roles[id]
	if !ok {
		return nil, iamerrors.ErrRoleNotFound
	}
	cp := *role
	return &cp, nil
}

func (r *memRoleRepo) SelectByCode(_ context.Context, code string) (*entity.Role, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, role := range r.s.roles {
		if role.Code == code {
			cp := *role
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrRoleNotFound
}

func (r *memRoleRepo) SelectList(_ context.Context, filter domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.Role
	for _, role := range r.s.roles {
		if filter.Query != "" && !strings.Contains(strings.ToLower(role.Code+role.Name), strings.ToLower(filter.Query)) {
			continue
		}
		if filter.ManagementDomain != "" && role.ManagementDomain != filter.ManagementDomain {
			continue
		}
		if filter.ActiveOnly && !role.IsActive {
			continue
		}
		cp := *role
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.Role) int { return strings.Compare(a.Code, b.Code) })
	return paginateSlice(out, filter.Page), nil
}

func (r *memRoleRepo) SelectEffectiveByUserID(_ context.Context, userID string, now time.Time) ([]entity.Role, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	seen := make(map[string]struct{})
	var out []entity.Role
	for _, ur := range r.s.userRoles {
		if ur.UserID == userID && ur.IsEffective(now) {
			if role, ok := r.s.roles[ur.RoleID]; ok && role.IsActive {
				if _, dup := seen[role.ID]; !dup {
					seen[role.ID] = struct{}{}
					out = append(out, *role)
				}
			}
		}
	}
	userGroups := make(map[string]struct{})
	for _, gm := range r.s.groupMembers {
		if gm.UserID == userID {
			if g, ok := r.s.groups[gm.GroupID]; ok && g.IsActive {
				userGroups[g.ID] = struct{}{}
			}
		}
	}
	for _, gr := range r.s.groupRoles {
		if _, inGroup := userGroups[gr.GroupID]; inGroup && gr.IsEffective(now) {
			if role, ok := r.s.roles[gr.RoleID]; ok && role.IsActive {
				if _, dup := seen[role.ID]; !dup {
					seen[role.ID] = struct{}{}
					out = append(out, *role)
				}
			}
		}
	}
	slices.SortFunc(out, func(a, b entity.Role) int { return strings.Compare(a.Code, b.Code) })
	return out, nil
}

func (r *memRoleRepo) Insert(_ context.Context, role *entity.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *role
	r.s.roles[role.ID] = &cp
	return nil
}

func (r *memRoleRepo) Update(_ context.Context, role *entity.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.roles[role.ID]; !ok {
		return iamerrors.ErrRoleNotFound
	}
	cp := *role
	r.s.roles[role.ID] = &cp
	return nil
}

func (r *memRoleRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.roles[id]; !ok {
		return iamerrors.ErrRoleNotFound
	}
	delete(r.s.roles, id)
	return nil
}

type memRolePermRepo struct{ s *memStore }

func (r *memRolePermRepo) CheckExists(_ context.Context, roleID, permID string) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	_, ok := r.s.rolePermissions[roleID+"|"+permID]
	return ok, nil
}

func (r *memRolePermRepo) Insert(_ context.Context, rp *entity.RolePermission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *rp
	r.s.rolePermissions[rp.RoleID+"|"+rp.PermissionID] = &cp
	return nil
}

func (r *memRolePermRepo) Delete(_ context.Context, roleID, permID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	key := roleID + "|" + permID
	if _, ok := r.s.rolePermissions[key]; !ok {
		return iamerrors.ErrRolePermissionNotFound
	}
	delete(r.s.rolePermissions, key)
	return nil
}

type memUserRoleRepo struct{ s *memStore }

func (r *memUserRoleRepo) SelectByUserID(_ context.Context, userID string, includeExpired bool, now time.Time) ([]*entity.UserRole, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.UserRole
	for _, ur := range r.s.userRoles {
		if ur.UserID == userID && (includeExpired || ur.IsEffective(now)) {
			cp := *ur
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memUserRoleRepo) CheckExists(_ context.Context, userID, roleID string, scope entity.Scope) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, ur := range r.s.userRoles {
		if ur.UserID == userID && ur.RoleID == roleID && ur.Scope.Matches(scope) && scope.Matches(ur.Scope) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memUserRoleRepo) Insert(_ context.Context, ur *entity.UserRole) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *ur
	r.s.userRoles[ur.ID] = &cp
	return nil
}

func (r *memUserRoleRepo) DeleteByID(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.userRoles[id]; !ok {
		return iamerrors.ErrUserRoleNotFound
	}
	delete(r.s.userRoles, id)
	return nil
}

// --- GroupRepository, GroupMemberRepository, GroupRoleRepository ---

type memGroupRepo struct{ s *memStore }

func (r *memGroupRepo) SelectByID(_ context.Context, id string) (*entity.Group, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	g, ok := r.s.groups[id]
	if !ok {
		return nil, iamerrors.ErrGroupNotFound
	}
	cp := *g
	return &cp, nil
}

func (r *memGroupRepo) SelectByCode(_ context.Context, code string) (*entity.Group, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, g := range r.s.groups {
		if g.Code == code {
			cp := *g
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrGroupNotFound
}

func (r *memGroupRepo) SelectList(_ context.Context, filter domainrepo.GroupFilter) (domainrepo.PageResult[*entity.Group], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.Group
	for _, g := range r.s.groups {
		if filter.Query != "" && !strings.Contains(strings.ToLower(g.Code+g.Name), strings.ToLower(filter.Query)) {
			continue
		}
		if filter.ActiveOnly && !g.IsActive {
			continue
		}
		cp := *g
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.Group) int { return strings.Compare(a.Code, b.Code) })
	return paginateSlice(out, filter.Page), nil
}

func (r *memGroupRepo) Insert(_ context.Context, g *entity.Group) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *g
	r.s.groups[g.ID] = &cp
	return nil
}

func (r *memGroupRepo) Update(_ context.Context, g *entity.Group) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.groups[g.ID]; !ok {
		return iamerrors.ErrGroupNotFound
	}
	cp := *g
	r.s.groups[g.ID] = &cp
	return nil
}

func (r *memGroupRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.groups[id]; !ok {
		return iamerrors.ErrGroupNotFound
	}
	delete(r.s.groups, id)
	return nil
}

type memGroupMemberRepo struct{ s *memStore }

func (r *memGroupMemberRepo) SelectUsersByGroupID(_ context.Context, groupID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.User], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.User
	for _, gm := range r.s.groupMembers {
		if gm.GroupID == groupID {
			if u, ok := r.s.users[gm.UserID]; ok && !u.IsDeleted() {
				cp := *u
				out = append(out, &cp)
			}
		}
	}
	return paginateSlice(out, page), nil
}

func (r *memGroupMemberRepo) SelectGroupsByUserID(_ context.Context, userID string) ([]*entity.Group, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.Group
	for _, gm := range r.s.groupMembers {
		if gm.UserID == userID {
			if g, ok := r.s.groups[gm.GroupID]; ok {
				cp := *g
				out = append(out, &cp)
			}
		}
	}
	return out, nil
}

func (r *memGroupMemberRepo) CheckExists(_ context.Context, groupID, userID string) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	_, ok := r.s.groupMembers[groupID+"|"+userID]
	return ok, nil
}

func (r *memGroupMemberRepo) Insert(_ context.Context, gm *entity.GroupMember) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *gm
	r.s.groupMembers[gm.GroupID+"|"+gm.UserID] = &cp
	return nil
}

func (r *memGroupMemberRepo) Delete(_ context.Context, groupID, userID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	key := groupID + "|" + userID
	if _, ok := r.s.groupMembers[key]; !ok {
		return iamerrors.ErrGroupMemberNotFound
	}
	delete(r.s.groupMembers, key)
	return nil
}

type memGroupRoleRepo struct{ s *memStore }

func (r *memGroupRoleRepo) SelectByGroupID(_ context.Context, groupID string, includeExpired bool, now time.Time) ([]*entity.GroupRole, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.GroupRole
	for _, gr := range r.s.groupRoles {
		if gr.GroupID == groupID && (includeExpired || gr.IsEffective(now)) {
			cp := *gr
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memGroupRoleRepo) CheckExists(_ context.Context, groupID, roleID string, scope entity.Scope) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, gr := range r.s.groupRoles {
		if gr.GroupID == groupID && gr.RoleID == roleID && gr.Scope.Matches(scope) && scope.Matches(gr.Scope) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memGroupRoleRepo) Insert(_ context.Context, gr *entity.GroupRole) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *gr
	r.s.groupRoles[gr.ID] = &cp
	return nil
}

func (r *memGroupRoleRepo) DeleteByID(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.groupRoles[id]; !ok {
		return iamerrors.ErrGroupRoleNotFound
	}
	delete(r.s.groupRoles, id)
	return nil
}
