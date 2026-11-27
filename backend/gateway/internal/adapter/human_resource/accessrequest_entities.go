package human_resource

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	accessrequest "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/accessrequest/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelAccessRequest(row *accessrequest.AccessRequest) (*model.AccessRequest, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource access-request response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource access-request ID", nil)
	}
	out := &model.AccessRequest{ID: nodeID("access-request", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.LifecycleEventID = strconv.FormatInt(row.GetLifecycleEventId(), 10)
	out.RequestKind = row.GetRequestKind()
	out.RequestedScopeType = row.GetRequestedScopeType()
	out.RequestedScopeReferenceID = row.GetRequestedScopeReferenceId()
	out.ApprovedCatalogReferenceID = row.ApprovedCatalogReferenceId
	out.IamRequestReferenceID = row.IamRequestReferenceId
	out.IamDecisionEventReferenceID = row.IamDecisionEventReferenceId
	out.Status = row.GetStatus()
	out.RequestedByPrincipalID = row.GetRequestedByPrincipalId()
	out.RequestedAt = pbTimePtr(row.GetRequestedAt())
	out.LastReconciledAt = pbTimePtr(row.GetLastReconciledAt())
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// AccessRequest membaca detail access-request berdasarkan ID.
func (a *Adapter) AccessRequest(ctx context.Context, rawID string) (*model.AccessRequest, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_access_requests", "human_resource.access_requests.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "access-request", "access-request")
	if err != nil {
		return nil, err
	}
	if a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.accessRequests.GetAccessRequestByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "access-request")
	}
	return toModelAccessRequest(resp.GetValue())
}

// AccessRequests mendaftar dan memfilter access-request dengan paginasi.
func (a *Adapter) AccessRequests(ctx context.Context, limit, offset int, f_worker_id *string, f_lifecycle_event_id *string, f_requested_scope_type *string, f_requested_scope_reference_id *string, f_approved_catalog_reference_id *string, f_iam_request_reference_id *string, f_iam_decision_event_reference_id *string, f_status *string, f_requested_by_principal_id *string, f_idempotency_key *string) (*Page[model.AccessRequest], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_access_requests", "human_resource.access_requests.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid access-request pagination", nil)
	}
	if a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_lifecycle_event_id != nil, f_requested_scope_type != nil, f_requested_scope_reference_id != nil, f_approved_catalog_reference_id != nil, f_iam_request_reference_id != nil, f_iam_decision_event_reference_id != nil, f_status != nil, f_requested_by_principal_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *accessrequest.AccessRequestsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, parseErr := strconv.ParseInt(strings.TrimSpace(*f_lifecycle_event_id), 10, 64)
		if parseErr != nil {
			return nil, middleware.ErrInvalidArgument("invalid lifecycle_event_id", nil)
		}
		resp, err = a.accessRequests.FindAccessRequestByLifecycleEventID(rpcCtx, &common.Int64FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_requested_scope_type, "requested_scope_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByRequestedScopeType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_requested_scope_reference_id, "requested_scope_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByRequestedScopeReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_approved_catalog_reference_id, "approved_catalog_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByApprovedCatalogReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_iam_request_reference_id, "iam_request_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByIAMRequestReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_iam_decision_event_reference_id, "iam_decision_event_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByIAMDecisionEventReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_requested_by_principal_id, "requested_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByRequestedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.accessRequests.FindAccessRequestByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.accessRequests.GetAllAccessRequests(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "access-request")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "access-request", toModelAccessRequest)
}

// CreateAccessRequest membuat access-request baru melalui HRM.
func (a *Adapter) CreateAccessRequest(ctx context.Context, input model.CreateAccessRequestInput) (*model.AccessRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_access_requests", "create", "human_resource.access_requests.create")
	if err != nil {
		return nil, err
	}
	if a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &accessrequest.AccessRequestInput{}
	patch.WorkerId = input.WorkerID
	if v, iErr := optionalInt64FromStr(input.LifecycleEventID, "lifecycle_event_id"); iErr != nil {
		return nil, iErr
	} else {
		patch.LifecycleEventId = v
	}
	patch.RequestKind = input.RequestKind
	patch.RequestedScopeType = input.RequestedScopeType
	patch.RequestedScopeReferenceId = input.RequestedScopeReferenceID
	patch.ApprovedCatalogReferenceId = input.ApprovedCatalogReferenceID
	patch.IamRequestReferenceId = input.IamRequestReferenceID
	patch.IamDecisionEventReferenceId = input.IamDecisionEventReferenceID
	patch.Status = input.Status
	patch.LastReconciledAt = optionalTimestamp(input.LastReconciledAt)
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.accessRequests.CreateAccessRequest(rpcCtx, &accessrequest.CreateAccessRequestRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "access-request")
	}
	return toModelAccessRequest(resp.GetValue())
}

// UpdateAccessRequest memperbarui access-request berdasarkan FieldMask.
func (a *Adapter) UpdateAccessRequest(ctx context.Context, rawID string, input model.UpdateAccessRequestInput) (*model.AccessRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_access_requests", "update", "human_resource.access_requests.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "access-request", "access-request")
	if err != nil {
		return nil, err
	}
	if a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &accessrequest.AccessRequestInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyInt64Mask(mask, "lifecycle_event_id", input.LifecycleEventID, &patch.LifecycleEventId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "request_kind", input.RequestKind, &patch.RequestKind, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "requested_scope_type", input.RequestedScopeType, &patch.RequestedScopeType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "requested_scope_reference_id", input.RequestedScopeReferenceID, &patch.RequestedScopeReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_catalog_reference_id", input.ApprovedCatalogReferenceID, &patch.ApprovedCatalogReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "iam_request_reference_id", input.IamRequestReferenceID, &patch.IamRequestReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "iam_decision_event_reference_id", input.IamDecisionEventReferenceID, &patch.IamDecisionEventReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "last_reconciled_at", input.LastReconciledAt, &patch.LastReconciledAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("access-request update is empty", nil)
	}
	resp, err := a.accessRequests.UpdateAccessRequest(rpcCtx, &accessrequest.UpdateAccessRequestRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "access-request")
	}
	return toModelAccessRequest(resp.GetValue())
}

// DeleteAccessRequest menghapus access-request berdasarkan ID.
func (a *Adapter) DeleteAccessRequest(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.accessRequests == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_access_requests", "human_resource.access_requests.delete", "access-request", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.accessRequests.DeleteAccessRequest(c, req)
	})
}
