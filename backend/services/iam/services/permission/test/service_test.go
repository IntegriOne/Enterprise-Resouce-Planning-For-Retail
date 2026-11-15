package permission_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

type mockPermissionRepo struct {
	mu      sync.Mutex
	byID    map[string]*entity.Permission
	byRole  map[string][]*entity.Permission
	listErr error
}

func newMockPermissionRepo() *mockPermissionRepo {
	return &mockPermissionRepo{
		byID:   make(map[string]*entity.Permission),
		byRole: make(map[string][]*entity.Permission),
	}
}

func (m *mockPermissionRepo) SelectByID(_ context.Context, id string) (*entity.Permission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.byID[id]
	if !ok {
		return nil, iamerrors.ErrPermissionNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockPermissionRepo) SelectByCode(_ context.Context, code string) (*entity.Permission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := entity.NormalizeCode(code)
	for _, p := range m.byID {
		if p.Code == norm {
			cp := *p
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrPermissionNotFound
}

func (m *mockPermissionRepo) SelectList(_ context.Context, filter domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listErr != nil {
		return domainrepo.PageResult[*entity.Permission]{}, m.listErr
	}
	norm := filter.Normalize()
	items := make([]*entity.Permission, 0, len(m.byID))
	for _, p := range m.byID {
		if norm.Resource != "" && p.Resource != entity.NormalizeCode(norm.Resource) {
			continue
		}
		cp := *p
		items = append(items, &cp)
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm.Page), nil
}

func (m *mockPermissionRepo) SelectByRoleID(_ context.Context, roleID string) ([]*entity.Permission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byRole[roleID], nil
}

func (m *mockPermissionRepo) Insert(_ context.Context, p *entity.Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.byID {
		if existing.Code == p.Code || (existing.Resource == p.Resource && existing.Action == p.Action) {
			return iamerrors.ErrPermissionAlreadyExists
		}
	}
	cp := *p
	m.byID[p.ID] = &cp
	return nil
}

func (m *mockPermissionRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[id]; !ok {
		return iamerrors.ErrPermissionNotFound
	}
	delete(m.byID, id)
	return nil
}

func TestPermissionService_CreateGetListDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := newMockPermissionRepo()
	svc, err := permission.NewService(repo, fixedClock{now: now}, crypto.NewUUIDGenerator())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := svc.Create(context.Background(), permission.CreateInput{
		Code:        " IAM.Users:Read ",
		Resource:    "IAM.Users",
		Action:      "Read",
		Description: "Read IAM user directory",
		IsSensitive: false,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Code != "iam.users:read" || created.Resource != "iam.users" || created.Action != "read" {
		t.Fatalf("unexpected normalized permission: %+v", created)
	}

	// Duplikasi kode izin harus ditolak sebelum penyisipan ganda.
	if _, err := svc.Create(context.Background(), permission.CreateInput{
		Code:     "iam.users:read",
		Resource: "iam.users",
		Action:   "read",
	}); !errors.Is(err, iamerrors.ErrPermissionAlreadyExists) {
		t.Fatalf("expected ErrPermissionAlreadyExists, got %v", err)
	}

	gotByID, err := svc.GetByID(context.Background(), created.ID)
	if err != nil || gotByID.Code != created.Code {
		t.Fatalf("GetByID() = (%+v, %v)", gotByID, err)
	}

	gotByCode, err := svc.GetByCode(context.Background(), "IAM.USERS:READ")
	if err != nil || gotByCode.ID != created.ID {
		t.Fatalf("GetByCode() = (%+v, %v)", gotByCode, err)
	}

	page, err := svc.List(context.Background(), domainrepo.PermissionFilter{Resource: "iam.users"})
	if err != nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("List() = (%+v, %v)", page, err)
	}

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := svc.GetByID(context.Background(), created.ID); !errors.Is(err, iamerrors.ErrPermissionNotFound) {
		t.Fatalf("expected ErrPermissionNotFound after delete, got %v", err)
	}
}

func TestPermissionService_FailClosedConstructorAndValidation(t *testing.T) {
	t.Parallel()

	if _, err := permission.NewService(nil, nil, nil); err == nil {
		t.Fatal("expected NewService(nil) error")
	}

	repo := newMockPermissionRepo()
	svc, err := permission.NewService(repo, nil, crypto.NewUUIDGenerator())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, err := svc.GetByID(context.Background(), "not-a-uuid"); err == nil {
		t.Fatal("expected invalid uuid error")
	}
	if _, err := svc.GetByCode(context.Background(), ""); !errors.Is(err, iamerrors.ErrInvalidPermissionCode) {
		t.Fatalf("expected ErrInvalidPermissionCode, got %v", err)
	}
}
