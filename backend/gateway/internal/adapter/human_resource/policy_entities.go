package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	policy "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/policy/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelPolicy(row *policy.Policy) (*model.Policy, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource policy response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource policy ID", nil)
	}
	out := &model.Policy{ID: nodeID("policy", row.GetId()), RawID: row.GetId()}
	out.PolicyCode = row.GetPolicyCode()
	out.PolicyVersion = int(row.GetPolicyVersion())
	out.PolicyName = row.GetPolicyName()
	out.CategoryCode = row.GetCategoryCode()
	out.LegalEntityReferenceID = row.GetLegalEntityReferenceId()
	out.EffectiveFrom = pbTimePtr(row.GetEffectiveFrom())
	out.EffectiveTo = pbTimePtr(row.GetEffectiveTo())
	out.DocumentVersionReferenceID = row.GetDocumentVersionReferenceId()
	out.ContentDigest = pbBase64StringValue(row.GetContentDigest())
	out.ApprovedByPrincipalID = row.GetApprovedByPrincipalId()
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	return out, nil
}

// Policy membaca detail policy berdasarkan ID.
func (a *Adapter) Policy(ctx context.Context, rawID string) (*model.Policy, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_policies", "human_resource.policies.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "policy", "policy")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.policy.GetPolicyByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "policy")
	}
	return toModelPolicy(resp.GetValue())
}

// Policies mendaftar dan memfilter policy dengan paginasi.
func (a *Adapter) Policies(ctx context.Context, limit, offset int, f_policy_code *string, f_category_code *string, f_legal_entity_reference_id *string, f_document_version_reference_id *string, f_approved_by_principal_id *string) (*Page[model.Policy], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_policies", "human_resource.policies.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid policy pagination", nil)
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_policy_code != nil, f_category_code != nil, f_legal_entity_reference_id != nil, f_document_version_reference_id != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *policy.PoliciesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_policy_code, "policy_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyByPolicyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_category_code, "category_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyByCategoryCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_legal_entity_reference_id, "legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_document_version_reference_id, "document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyByDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.policy.GetAllPolicies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "policy")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "policy", toModelPolicy)
}

// CreatePolicy membuat policy baru melalui HRM.
func (a *Adapter) CreatePolicy(ctx context.Context, input model.CreatePolicyInput) (*model.Policy, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_policies", "create", "human_resource.policies.create")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &policy.PolicyInput{}
	patch.PolicyCode = input.PolicyCode
	patch.PolicyVersion = optionalInt32Ptr(input.PolicyVersion)
	patch.PolicyName = input.PolicyName
	patch.CategoryCode = input.CategoryCode
	patch.LegalEntityReferenceId = input.LegalEntityReferenceID
	patch.EffectiveFrom = optionalTimestamp(input.EffectiveFrom)
	patch.EffectiveTo = optionalTimestamp(input.EffectiveTo)
	patch.DocumentVersionReferenceId = input.DocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.ContentDigest, "content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.ContentDigest = b
	}
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	resp, err := a.policy.CreatePolicy(rpcCtx, &policy.CreatePolicyRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "policy")
	}
	return toModelPolicy(resp.GetValue())
}

// UpdatePolicy memperbarui policy berdasarkan FieldMask.
func (a *Adapter) UpdatePolicy(ctx context.Context, rawID string, input model.UpdatePolicyInput) (*model.Policy, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_policies", "update", "human_resource.policies.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "policy", "policy")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &policy.PolicyInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "policy_code", input.PolicyCode, &patch.PolicyCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "policy_version", input.PolicyVersion, &patch.PolicyVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "policy_name", input.PolicyName, &patch.PolicyName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "category_code", input.CategoryCode, &patch.CategoryCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_entity_reference_id", input.LegalEntityReferenceID, &patch.LegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "document_version_reference_id", input.DocumentVersionReferenceID, &patch.DocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "content_digest", input.ContentDigest, &patch.ContentDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("policy update is empty", nil)
	}
	resp, err := a.policy.UpdatePolicy(rpcCtx, &policy.UpdatePolicyRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "policy")
	}
	return toModelPolicy(resp.GetValue())
}

// DeletePolicy menghapus policy berdasarkan ID.
func (a *Adapter) DeletePolicy(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_policies", "human_resource.policies.delete", "policy", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.policy.DeletePolicy(c, req)
	})
}

func toModelPolicyAcknowledgement(row *policy.PolicyAcknowledgement) (*model.PolicyAcknowledgement, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource policy-acknowledgement response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource policy-acknowledgement ID", nil)
	}
	out := &model.PolicyAcknowledgement{ID: nodeID("policy-acknowledgement", row.GetId()), RawID: row.GetId()}
	out.PolicyID = row.GetPolicyId()
	out.WorkerID = row.GetWorkerId()
	out.AssignedAt = pbTimePtr(row.GetAssignedAt())
	out.DueAt = pbTimePtr(row.GetDueAt())
	out.AcknowledgedAt = pbTimePtr(row.GetAcknowledgedAt())
	out.AcknowledgedByPrincipalID = row.AcknowledgedByPrincipalId
	out.EvidenceDocumentVersionReferenceID = row.EvidenceDocumentVersionReferenceId
	return out, nil
}

// PolicyAcknowledgement membaca detail policy-acknowledgement berdasarkan ID.
func (a *Adapter) PolicyAcknowledgement(ctx context.Context, rawID string) (*model.PolicyAcknowledgement, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_policy_acknowledgements", "human_resource.policy_acknowledgements.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "policy-acknowledgement", "policy-acknowledgement")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.policy.GetPolicyAcknowledgementByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "policy-acknowledgement")
	}
	return toModelPolicyAcknowledgement(resp.GetValue())
}

// PolicyAcknowledgements mendaftar dan memfilter policy-acknowledgement dengan paginasi.
func (a *Adapter) PolicyAcknowledgements(ctx context.Context, limit, offset int, f_policy_id *string, f_worker_id *string, f_acknowledged_by_principal_id *string, f_evidence_document_version_reference_id *string) (*Page[model.PolicyAcknowledgement], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_policy_acknowledgements", "human_resource.policy_acknowledgements.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid policy-acknowledgement pagination", nil)
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_policy_id != nil, f_worker_id != nil, f_acknowledged_by_principal_id != nil, f_evidence_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *policy.PolicyAcknowledgementsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_policy_id, "policy_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyAcknowledgementByPolicyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyAcknowledgementByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_acknowledged_by_principal_id, "acknowledged_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyAcknowledgementByAcknowledgedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.policy.FindPolicyAcknowledgementByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.policy.GetAllPolicyAcknowledgements(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "policy-acknowledgement")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "policy-acknowledgement", toModelPolicyAcknowledgement)
}

// CreatePolicyAcknowledgement membuat policy-acknowledgement baru melalui HRM.
func (a *Adapter) CreatePolicyAcknowledgement(ctx context.Context, input model.CreatePolicyAcknowledgementInput) (*model.PolicyAcknowledgement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_policy_acknowledgements", "create", "human_resource.policy_acknowledgements.create")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &policy.PolicyAcknowledgementInput{}
	patch.PolicyId = input.PolicyID
	patch.WorkerId = input.WorkerID
	patch.DueAt = optionalTimestamp(input.DueAt)
	patch.AcknowledgedAt = optionalTimestamp(input.AcknowledgedAt)
	patch.AcknowledgedByPrincipalId = input.AcknowledgedByPrincipalID
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	resp, err := a.policy.CreatePolicyAcknowledgement(rpcCtx, &policy.CreatePolicyAcknowledgementRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "policy-acknowledgement")
	}
	return toModelPolicyAcknowledgement(resp.GetValue())
}

// UpdatePolicyAcknowledgement memperbarui policy-acknowledgement berdasarkan FieldMask.
func (a *Adapter) UpdatePolicyAcknowledgement(ctx context.Context, rawID string, input model.UpdatePolicyAcknowledgementInput) (*model.PolicyAcknowledgement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_policy_acknowledgements", "update", "human_resource.policy_acknowledgements.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "policy-acknowledgement", "policy-acknowledgement")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &policy.PolicyAcknowledgementInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "policy_id", input.PolicyID, &patch.PolicyId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_at", input.DueAt, &patch.DueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "acknowledged_at", input.AcknowledgedAt, &patch.AcknowledgedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "acknowledged_by_principal_id", input.AcknowledgedByPrincipalID, &patch.AcknowledgedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("policy-acknowledgement update is empty", nil)
	}
	resp, err := a.policy.UpdatePolicyAcknowledgement(rpcCtx, &policy.UpdatePolicyAcknowledgementRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "policy-acknowledgement")
	}
	return toModelPolicyAcknowledgement(resp.GetValue())
}

// DeletePolicyAcknowledgement menghapus policy-acknowledgement berdasarkan ID.
func (a *Adapter) DeletePolicyAcknowledgement(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_policy_acknowledgements", "human_resource.policy_acknowledgements.delete", "policy-acknowledgement", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.policy.DeletePolicyAcknowledgement(c, req)
	})
}
