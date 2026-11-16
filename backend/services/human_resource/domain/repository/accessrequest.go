package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// AccessRequestRepository adalah port persistensi hr_access_requests.
type AccessRequestRepository interface {
	Repository[entity.AccessRequest, string]
	InsertAccessRequests(context.Context, *entity.AccessRequest) error
	UpdateAccessRequests(context.Context, *entity.AccessRequest) error
	FindAccessRequestsByUUID(context.Context, string) (*entity.AccessRequest, error)
	FindAccessRequestsByUUIDForUpdate(context.Context, string) (*entity.AccessRequest, error)
	DeleteAccessRequestsByUUID(context.Context, string) error
	GetAllAccessRequests(context.Context, int, int) ([]entity.AccessRequest, error)
	CountAccessRequests(context.Context) (int64, error)
	SelectAccessRequestsWhere(context.Context, []Filter, int, int) ([]entity.AccessRequest, error)
	CountAccessRequestsWhere(context.Context, []Filter) (int64, error)
	FindAccessRequestsByWorkerID(context.Context, string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByLifecycleEventID(context.Context, int64, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByRequestedScopeType(context.Context, string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByRequestedScopeReferenceID(context.Context, string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByApprovedCatalogReferenceID(context.Context, *string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByIAMRequestReferenceID(context.Context, *string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByIAMDecisionEventReferenceID(context.Context, *string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByStatus(context.Context, string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByRequestedByPrincipalID(context.Context, string, int, int) ([]entity.AccessRequest, error)
	FindAccessRequestsByIdempotencyKey(context.Context, string, int, int) ([]entity.AccessRequest, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error)
	FindByLifecycleEventID(ctx context.Context, value int64, limit, offset int) ([]entity.AccessRequest, error)
	FindByRequestedScopeType(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error)
	FindByRequestedScopeReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error)
	FindByApprovedCatalogReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error)
	FindByIAMRequestReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error)
	FindByIAMDecisionEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error)
	FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error)
}
