// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime       operation.Dependencies
	AccessRequest operation.Repository[entity.AccessRequest, string]
}

// Service mengoordinasikan operasi domain accessrequest.
type Service struct {
	accessRequest *operation.Resource[entity.AccessRequest, string, *entity.AccessRequest]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.accessRequest, err = operation.NewResource[entity.AccessRequest, string, *entity.AccessRequest](deps.AccessRequest, deps.Runtime, specAccessRequest())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateAccessRequest menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateAccessRequest(ctx context.Context, in CreateAccessRequestInput) (*entity.AccessRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.accessRequest.Create(ctx, func() (*entity.AccessRequest, error) { return createAccessRequest(in) })
}

// UpdateAccessRequest menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateAccessRequest(ctx context.Context, key string, in UpdateAccessRequestInput) (*entity.AccessRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.accessRequest.Update(ctx, key, func(value *entity.AccessRequest) error { return patchAccessRequest(value, in) })
}

// DeleteAccessRequest menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteAccessRequest(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.accessRequest.Delete(ctx, key)
}

// GetAccessRequestByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAccessRequestByID(ctx context.Context, key string) (*entity.AccessRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.accessRequest.Get(ctx, key)
}

// GetAllAccessRequests menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllAccessRequests(ctx context.Context, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, nil, limit, offset)
}

// FindAccessRequestByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByLifecycleEventID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByLifecycleEventID(ctx context.Context, value int64, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "lifecycle_event_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByRequestedScopeType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByRequestedScopeType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "requested_scope_type", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByRequestedScopeReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByRequestedScopeReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "requested_scope_reference_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByApprovedCatalogReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByApprovedCatalogReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "approved_catalog_reference_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByIAMRequestReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByIAMRequestReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "iam_request_reference_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByIAMDecisionEventReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByIAMDecisionEventReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "iam_decision_event_reference_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByRequestedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "requested_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindAccessRequestByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindAccessRequestByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.AccessRequest], error) {
	if s == nil {
		return operation.Page[entity.AccessRequest]{}, hrerrors.ErrUnavailable
	}
	return s.accessRequest.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}
