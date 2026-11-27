// Package address mengimplementasikan logika bisnis pengelolaan master alamat geografis Enterprise Foundation.
package address

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

// Service mengelola operasi CRUD, pencarian, dedup prapembuatan, dan pemeriksaan keterpakaian alamat master.
type Service struct {
	repo  domainrepo.AddressRepository
	clock domainservice.Clock
	idGen domainservice.IDGenerator
}

// NewService membentuk Service master alamat dengan validasi fail-closed atas repositori dan generator ID.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Addresses == nil {
		return nil, eferrors.NewValidation("addresses", eferrors.Code("enterprise_foundation.service.nil_address_repo"), "address repository is required")
	}
	if deps.IDGen == nil {
		return nil, eferrors.NewValidation("id_gen", eferrors.Code("enterprise_foundation.service.nil_id_generator"), "id generator is required")
	}
	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	return &Service{
		repo:  deps.Addresses,
		clock: clock,
		idGen: deps.IDGen,
	}, nil
}

// GetByID mengambil detail satu alamat berdasarkan UUID-nya.
func (s *Service) GetByID(ctx context.Context, id string) (*entity.Address, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("address_id", id); err != nil {
		return nil, err
	}
	return s.repo.SelectByID(ctx, entity.NormalizeUUID(id))
}

// GetByIDs mengambil beberapa alamat secara massal berdasarkan daftar UUID.
func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]*entity.Address, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	normIDs, err := servicekit.ValidateOptionalUUIDList("address_ids", ids)
	if err != nil {
		return nil, err
	}
	if len(normIDs) == 0 {
		return []*entity.Address{}, nil
	}
	return s.repo.SelectByIDs(ctx, normIDs)
}

// List mengambil halaman daftar alamat berdasarkan filter kota, provinsi, kode pos, negara, atau kata kunci.
func (s *Service) List(ctx context.Context, filter domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Address]{}, err
	}
	return s.repo.SelectList(ctx, filter.Normalize())
}

// GetAll mengambil daftar alamat master secara berhalaman.
func (s *Service) GetAll(ctx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Address], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Address]{}, err
	}
	return s.repo.SelectList(ctx, domainrepo.AddressFilter{Page: page.Normalize()})
}

// IsReferenced memeriksa apakah alamat sedang dirujuk oleh badan hukum, lokasi fisik, atau tautan alamat mitra.
func (s *Service) IsReferenced(ctx context.Context, id string) (bool, error) {
	if err := s.ensureReady(); err != nil {
		return false, err
	}
	if err := entity.ValidateUUID("address_id", id); err != nil {
		return false, err
	}
	normID := entity.NormalizeUUID(id)
	if _, err := s.repo.SelectByID(ctx, normID); err != nil {
		return false, err
	}
	return s.repo.CheckReferenced(ctx, normID)
}

// Create membentuk alamat baru setelah memvalidasi field skema dan memastikan tidak ada alamat identik yang sudah tersimpan.
func (s *Service) Create(ctx context.Context, in CreateInput) (*entity.Address, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.address.id_generation_failed"), "failed to generate address id")
	}

	addr, err := entity.NewAddress(id, toAddressParams(in), s.clock.Now())
	if err != nil {
		return nil, err
	}

	dup, err := s.findDuplicate(ctx, addr, "")
	if err != nil {
		return nil, err
	}
	if dup != nil {
		return nil, eferrors.WithField(eferrors.ErrAlreadyExists, "address_line_1", "identical address already exists")
	}

	if err := s.repo.Insert(ctx, addr); err != nil {
		return nil, err
	}
	return addr, nil
}

// FindOrCreate mengembalikan alamat identik yang sudah ada bila ditemukan atau membuat baris alamat baru secara aman.
func (s *Service) FindOrCreate(ctx context.Context, in CreateInput) (*entity.Address, bool, error) {
	if err := s.ensureReady(); err != nil {
		return nil, false, err
	}

	id, err := s.idGen.NewUUID()
	if err != nil {
		return nil, false, eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.address.id_generation_failed"), "failed to generate address id")
	}

	candidate, err := entity.NewAddress(id, toAddressParams(in), s.clock.Now())
	if err != nil {
		return nil, false, err
	}

	existing, err := s.findDuplicate(ctx, candidate, "")
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	if err := s.repo.Insert(ctx, candidate); err != nil {
		return nil, false, err
	}
	return candidate, true, nil
}

// Update memperbarui detail alamat yang ada sambil mencegah duplikasi terhadap baris alamat lain.
func (s *Service) Update(ctx context.Context, in UpdateInput) (*entity.Address, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("address_id", in.ID); err != nil {
		return nil, err
	}

	normID := entity.NormalizeUUID(in.ID)
	addr, err := s.repo.SelectByID(ctx, normID)
	if err != nil {
		return nil, err
	}

	if err := addr.Update(entity.AddressParams{
		Label:            in.Label,
		AddressLine1:     in.AddressLine1,
		AddressLine2:     in.AddressLine2,
		City:             in.City,
		Province:         in.Province,
		PostalCode:       in.PostalCode,
		CountryCode:      in.CountryCode,
		FormattedAddress: in.FormattedAddress,
		Latitude:         in.Latitude,
		Longitude:        in.Longitude,
	}, s.clock.Now()); err != nil {
		return nil, err
	}

	dup, err := s.findDuplicate(ctx, addr, addr.ID)
	if err != nil {
		return nil, err
	}
	if dup != nil {
		return nil, eferrors.WithField(eferrors.ErrAlreadyExists, "address_line_1", "identical address already exists")
	}

	if err := s.repo.Update(ctx, addr); err != nil {
		return nil, err
	}
	return addr, nil
}

// Delete menghapus alamat master hanya bila alamat tersebut tidak lagi dirujuk oleh tabel lain.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("address_id", id); err != nil {
		return err
	}

	normID := entity.NormalizeUUID(id)
	if _, err := s.repo.SelectByID(ctx, normID); err != nil {
		return err
	}
	referenced, err := s.repo.CheckReferenced(ctx, normID)
	if err != nil {
		return err
	}
	if referenced {
		return eferrors.ErrAddressInUse
	}
	return s.repo.DeleteByID(ctx, normID)
}

func (s *Service) ensureReady() error {
	if s == nil || s.repo == nil || s.idGen == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.service.not_initialized"), "address service is not initialized")
	}
	return nil
}

func (s *Service) findDuplicate(ctx context.Context, candidate *entity.Address, excludeID string) (*entity.Address, error) {
	if candidate == nil {
		return nil, nil
	}
	filter := domainrepo.AddressFilter{
		Query:       candidate.AddressLine1,
		CountryCode: candidate.CountryCode,
		City:        candidate.City,
		Province:    derefString(candidate.Province),
		PostalCode:  derefString(candidate.PostalCode),
	}
	matches, err := servicekit.CollectAllPages(ctx, func(pageCtx context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Address], error) {
		f := filter
		f.Page = page
		return s.repo.SelectList(pageCtx, f.Normalize())
	})
	if err != nil {
		return nil, err
	}
	normExclude := entity.NormalizeUUID(excludeID)
	for _, item := range matches {
		if item == nil {
			continue
		}
		if normExclude != "" && entity.NormalizeUUID(item.ID) == normExclude {
			continue
		}
		if sameNormalizedAddress(item, candidate) {
			return item, nil
		}
	}
	return nil, nil
}

func toAddressParams(in CreateInput) entity.AddressParams {
	return entity.AddressParams{
		Label:            in.Label,
		AddressLine1:     in.AddressLine1,
		AddressLine2:     in.AddressLine2,
		City:             in.City,
		Province:         in.Province,
		PostalCode:       in.PostalCode,
		CountryCode:      in.CountryCode,
		FormattedAddress: in.FormattedAddress,
		Latitude:         in.Latitude,
		Longitude:        in.Longitude,
	}
}

func sameNormalizedAddress(a, b *entity.Address) bool {
	return strings.EqualFold(a.CountryCode, b.CountryCode) &&
		strings.EqualFold(a.City, b.City) &&
		strings.EqualFold(a.AddressLine1, b.AddressLine1) &&
		equalOptionalFold(a.AddressLine2, b.AddressLine2) &&
		equalOptionalFold(a.Province, b.Province) &&
		equalOptionalFold(a.PostalCode, b.PostalCode) &&
		equalOptionalFold(a.Label, b.Label)
}

func equalOptionalFold(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*a), strings.TrimSpace(*b))
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
