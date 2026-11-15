package permission

import (
	"context"
	"errors"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

// Service mengelola katalog izin atomik RBAC IAM beserta validasi duplikasinya.
type Service struct {
	repo  domainrepo.PermissionRepository
	clock domainservice.Clock
	idGen domainservice.IDGenerator
}

// NewService membentuk Service izin dengan validasi ketergantungan wajib secara fail-closed.
func NewService(
	repo domainrepo.PermissionRepository,
	clock domainservice.Clock,
	idGen domainservice.IDGenerator,
) (*Service, error) {
	if repo == nil {
		return nil, iamerrors.NewValidation("permission_repo", iamerrors.Code("iam.service.nil_permission_repo"), "permission repository is required")
	}
	if idGen == nil {
		return nil, iamerrors.NewValidation("id_generator", iamerrors.Code("iam.service.nil_id_generator"), "id generator is required")
	}
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	return &Service{
		repo:  repo,
		clock: clock,
		idGen: idGen,
	}, nil
}

// GetByID mengambil satu izin berdasarkan UUID-nya.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.Permission, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("permission_id", id); err != nil {
		return nil, err
	}
	return s.repo.SelectByID(ctx, id)
}

// GetByCode mengambil satu izin berdasarkan kode bernormalisasi (misal "iam.users:read").
func (s *Service) GetByCode(ctx context.Context, code string) (*entity.Permission, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normalized, err := entity.ValidatePermissionCode(code)
	if err != nil {
		return nil, err
	}
	return s.repo.SelectByCode(ctx, normalized)
}

// List mengambil halaman daftar katalog izin berdasarkan filter pencarian dan sensitivitas.
func (s *Service) List(ctx context.Context, filter domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Permission]{}, err
	}
	return s.repo.SelectList(ctx, filter.Normalize())
}

// ListByRoleID mengambil seluruh izin yang terpasang langsung pada peran tertentu.
func (s *Service) ListByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("role_id", roleID); err != nil {
		return nil, err
	}
	return s.repo.SelectByRoleID(ctx, roleID)
}

// Create membuat izin atomik baru setelah memvalidasi format kode dan ketiadaan duplikasi.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.Permission, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.permission.id_generation_failed"), "failed to generate permission id")
	}

	perm, err := entity.NewPermission(id, in.Code, in.Resource, in.Action, in.Description, in.IsSensitive, s.clock.Now())
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.SelectByCode(ctx, perm.Code)
	if err == nil && existing != nil {
		return nil, iamerrors.WithField(iamerrors.ErrPermissionAlreadyExists, "code", "permission code already exists")
	}
	if err != nil && !errors.Is(err, iamerrors.ErrPermissionNotFound) {
		return nil, err
	}

	if err := s.repo.Insert(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// Delete menghapus izin atomik dari katalog berdasarkan UUID-nya.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("permission_id", id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) ensureReady() error {
	if s == nil || s.repo == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.permission.service_unavailable"), "permission service is not initialized")
	}
	return nil
}
