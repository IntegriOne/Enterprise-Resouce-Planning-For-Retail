package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	talent "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelTalentFramework(row *talent.TalentFramework) (*model.TalentFramework, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-framework response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-framework ID", nil)
	}
	out := &model.TalentFramework{ID: nodeID("talent-framework", row.GetId()), RawID: row.GetId()}
	out.FrameworkCode = row.GetFrameworkCode()
	out.FrameworkVersion = int(row.GetFrameworkVersion())
	out.FrameworkName = row.GetFrameworkName()
	out.RubricDocumentVersionReferenceID = row.GetRubricDocumentVersionReferenceId()
	out.RubricDigest = pbBase64StringValue(row.GetRubricDigest())
	out.EffectiveFrom = pbTimePtr(row.GetEffectiveFrom())
	out.EffectiveTo = pbTimePtr(row.GetEffectiveTo())
	out.ApprovedByPrincipalID = row.GetApprovedByPrincipalId()
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	return out, nil
}

// TalentFramework membaca detail talent-framework berdasarkan ID.
func (a *Adapter) TalentFramework(ctx context.Context, rawID string) (*model.TalentFramework, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_frameworks", "human_resource.talent_frameworks.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-framework", "talent-framework")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentFrameworkByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-framework")
	}
	return toModelTalentFramework(resp.GetValue())
}

// TalentFrameworks mendaftar dan memfilter talent-framework dengan paginasi.
func (a *Adapter) TalentFrameworks(ctx context.Context, limit, offset int, f_framework_code *string, f_rubric_document_version_reference_id *string, f_approved_by_principal_id *string) (*Page[model.TalentFramework], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_frameworks", "human_resource.talent_frameworks.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-framework pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_framework_code != nil, f_rubric_document_version_reference_id != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentFrameworksResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_framework_code, "framework_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentFrameworkByFrameworkCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_rubric_document_version_reference_id, "rubric_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentFrameworkByRubricDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentFrameworkByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentFrameworks(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-framework")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-framework", toModelTalentFramework)
}

// CreateTalentFramework membuat talent-framework baru melalui HRM.
func (a *Adapter) CreateTalentFramework(ctx context.Context, input model.CreateTalentFrameworkInput) (*model.TalentFramework, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_frameworks", "create", "human_resource.talent_frameworks.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentFrameworkInput{}
	patch.FrameworkCode = input.FrameworkCode
	patch.FrameworkVersion = optionalInt32Ptr(input.FrameworkVersion)
	patch.FrameworkName = input.FrameworkName
	patch.RubricDocumentVersionReferenceId = input.RubricDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.RubricDigest, "rubric_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.RubricDigest = b
	}
	patch.EffectiveFrom = optionalTimestamp(input.EffectiveFrom)
	patch.EffectiveTo = optionalTimestamp(input.EffectiveTo)
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	resp, err := a.talent.CreateTalentFramework(rpcCtx, &talent.CreateTalentFrameworkRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-framework")
	}
	return toModelTalentFramework(resp.GetValue())
}

// UpdateTalentFramework memperbarui talent-framework berdasarkan FieldMask.
func (a *Adapter) UpdateTalentFramework(ctx context.Context, rawID string, input model.UpdateTalentFrameworkInput) (*model.TalentFramework, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_frameworks", "update", "human_resource.talent_frameworks.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-framework", "talent-framework")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentFrameworkInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "framework_code", input.FrameworkCode, &patch.FrameworkCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "framework_version", input.FrameworkVersion, &patch.FrameworkVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "framework_name", input.FrameworkName, &patch.FrameworkName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rubric_document_version_reference_id", input.RubricDocumentVersionReferenceID, &patch.RubricDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "rubric_digest", input.RubricDigest, &patch.RubricDigest, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-framework update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentFramework(rpcCtx, &talent.UpdateTalentFrameworkRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-framework")
	}
	return toModelTalentFramework(resp.GetValue())
}

// DeleteTalentFramework menghapus talent-framework berdasarkan ID.
func (a *Adapter) DeleteTalentFramework(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_frameworks", "human_resource.talent_frameworks.delete", "talent-framework", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentFramework(c, req)
	})
}

func toModelTalentDimension(row *talent.TalentDimension) (*model.TalentDimension, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-dimension response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-dimension ID", nil)
	}
	out := &model.TalentDimension{ID: nodeID("talent-dimension", row.GetId()), RawID: row.GetId()}
	out.FrameworkID = row.GetFrameworkId()
	out.DimensionCode = row.GetDimensionCode()
	out.DimensionName = row.GetDimensionName()
	out.Definition = row.GetDefinition()
	if doc := row.GetRatingAnchors(); doc != nil {
		out.RatingAnchors = pbJSONString(doc.GetUtf8())
	}
	out.WeightPercent = row.GetWeightPercent()
	out.IsRequired = row.GetIsRequired()
	return out, nil
}

// TalentDimension membaca detail talent-dimension berdasarkan ID.
func (a *Adapter) TalentDimension(ctx context.Context, rawID string) (*model.TalentDimension, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_dimensions", "human_resource.talent_dimensions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-dimension", "talent-dimension")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentDimensionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-dimension")
	}
	return toModelTalentDimension(resp.GetValue())
}

// TalentDimensions mendaftar dan memfilter talent-dimension dengan paginasi.
func (a *Adapter) TalentDimensions(ctx context.Context, limit, offset int, f_framework_id *string, f_dimension_code *string) (*Page[model.TalentDimension], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_dimensions", "human_resource.talent_dimensions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-dimension pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_framework_id != nil, f_dimension_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentDimensionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_framework_id, "framework_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentDimensionByFrameworkID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_dimension_code, "dimension_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentDimensionByDimensionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentDimensions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-dimension")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-dimension", toModelTalentDimension)
}

// CreateTalentDimension membuat talent-dimension baru melalui HRM.
func (a *Adapter) CreateTalentDimension(ctx context.Context, input model.CreateTalentDimensionInput) (*model.TalentDimension, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_dimensions", "create", "human_resource.talent_dimensions.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentDimensionInput{}
	patch.FrameworkId = input.FrameworkID
	patch.DimensionCode = input.DimensionCode
	patch.DimensionName = input.DimensionName
	patch.Definition = input.Definition
	if b, jErr := optionalJSONBytes(input.RatingAnchors, "rating_anchors"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.RatingAnchors = &common.JsonDocument{Utf8: b}
	}
	patch.WeightPercent = input.WeightPercent
	patch.IsRequired = input.IsRequired
	resp, err := a.talent.CreateTalentDimension(rpcCtx, &talent.CreateTalentDimensionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-dimension")
	}
	return toModelTalentDimension(resp.GetValue())
}

// UpdateTalentDimension memperbarui talent-dimension berdasarkan FieldMask.
func (a *Adapter) UpdateTalentDimension(ctx context.Context, rawID string, input model.UpdateTalentDimensionInput) (*model.TalentDimension, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_dimensions", "update", "human_resource.talent_dimensions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-dimension", "talent-dimension")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentDimensionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "framework_id", input.FrameworkID, &patch.FrameworkId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "dimension_code", input.DimensionCode, &patch.DimensionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "dimension_name", input.DimensionName, &patch.DimensionName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "definition", input.Definition, &patch.Definition, true); err != nil {
		return nil, err
	}
	if val, set := input.RatingAnchors.ValueOK(); set {
		if val == nil {
			patch.RatingAnchors = nil
		} else if b, jErr := requiredJSONBytes(*val, "rating_anchors"); jErr != nil {
			return nil, jErr
		} else {
			patch.RatingAnchors = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "rating_anchors")
	}
	if err := applyStringMask(mask, "weight_percent", input.WeightPercent, &patch.WeightPercent, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_required", input.IsRequired, &patch.IsRequired, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-dimension update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentDimension(rpcCtx, &talent.UpdateTalentDimensionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-dimension")
	}
	return toModelTalentDimension(resp.GetValue())
}

// DeleteTalentDimension menghapus talent-dimension berdasarkan ID.
func (a *Adapter) DeleteTalentDimension(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_dimensions", "human_resource.talent_dimensions.delete", "talent-dimension", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentDimension(c, req)
	})
}

func toModelTalentReviewCycle(row *talent.TalentReviewCycle) (*model.TalentReviewCycle, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-review-cycle response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-review-cycle ID", nil)
	}
	out := &model.TalentReviewCycle{ID: nodeID("talent-review-cycle", row.GetId()), RawID: row.GetId()}
	out.CycleCode = row.GetCycleCode()
	out.FrameworkID = row.GetFrameworkId()
	out.LegalEntityReferenceID = row.GetLegalEntityReferenceId()
	out.OrganizationUnitReferenceID = row.OrganizationUnitReferenceId
	out.PeriodStart = pbTimePtr(row.GetPeriodStart())
	out.PeriodEnd = pbTimePtr(row.GetPeriodEnd())
	out.Status = row.GetStatus()
	out.OwnerPrincipalID = row.GetOwnerPrincipalId()
	out.ClosedAt = pbTimePtr(row.GetClosedAt())
	return out, nil
}

// TalentReviewCycle membaca detail talent-review-cycle berdasarkan ID.
func (a *Adapter) TalentReviewCycle(ctx context.Context, rawID string) (*model.TalentReviewCycle, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_review_cycles", "human_resource.talent_review_cycles.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-review-cycle", "talent-review-cycle")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentReviewCycleByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-review-cycle")
	}
	return toModelTalentReviewCycle(resp.GetValue())
}

// TalentReviewCycles mendaftar dan memfilter talent-review-cycle dengan paginasi.
func (a *Adapter) TalentReviewCycles(ctx context.Context, limit, offset int, f_cycle_code *string, f_framework_id *string, f_legal_entity_reference_id *string, f_organization_unit_reference_id *string, f_status *string, f_owner_principal_id *string) (*Page[model.TalentReviewCycle], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_review_cycles", "human_resource.talent_review_cycles.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-review-cycle pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_cycle_code != nil, f_framework_id != nil, f_legal_entity_reference_id != nil, f_organization_unit_reference_id != nil, f_status != nil, f_owner_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentReviewCyclesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_cycle_code, "cycle_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentReviewCycleByCycleCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_framework_id, "framework_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentReviewCycleByFrameworkID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_legal_entity_reference_id, "legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentReviewCycleByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_organization_unit_reference_id, "organization_unit_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentReviewCycleByOrganizationUnitReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentReviewCycleByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_owner_principal_id, "owner_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentReviewCycleByOwnerPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentReviewCycles(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-review-cycle")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-review-cycle", toModelTalentReviewCycle)
}

// CreateTalentReviewCycle membuat talent-review-cycle baru melalui HRM.
func (a *Adapter) CreateTalentReviewCycle(ctx context.Context, input model.CreateTalentReviewCycleInput) (*model.TalentReviewCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_review_cycles", "create", "human_resource.talent_review_cycles.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentReviewCycleInput{}
	patch.CycleCode = input.CycleCode
	patch.FrameworkId = input.FrameworkID
	patch.LegalEntityReferenceId = input.LegalEntityReferenceID
	patch.OrganizationUnitReferenceId = input.OrganizationUnitReferenceID
	patch.PeriodStart = optionalTimestamp(input.PeriodStart)
	patch.PeriodEnd = optionalTimestamp(input.PeriodEnd)
	patch.Status = input.Status
	patch.OwnerPrincipalId = input.OwnerPrincipalID
	patch.ClosedAt = optionalTimestamp(input.ClosedAt)
	resp, err := a.talent.CreateTalentReviewCycle(rpcCtx, &talent.CreateTalentReviewCycleRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-review-cycle")
	}
	return toModelTalentReviewCycle(resp.GetValue())
}

// UpdateTalentReviewCycle memperbarui talent-review-cycle berdasarkan FieldMask.
func (a *Adapter) UpdateTalentReviewCycle(ctx context.Context, rawID string, input model.UpdateTalentReviewCycleInput) (*model.TalentReviewCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_review_cycles", "update", "human_resource.talent_review_cycles.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-review-cycle", "talent-review-cycle")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentReviewCycleInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "cycle_code", input.CycleCode, &patch.CycleCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "framework_id", input.FrameworkID, &patch.FrameworkId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_entity_reference_id", input.LegalEntityReferenceID, &patch.LegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "organization_unit_reference_id", input.OrganizationUnitReferenceID, &patch.OrganizationUnitReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "period_start", input.PeriodStart, &patch.PeriodStart, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "period_end", input.PeriodEnd, &patch.PeriodEnd, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "owner_principal_id", input.OwnerPrincipalID, &patch.OwnerPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "closed_at", input.ClosedAt, &patch.ClosedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-review-cycle update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentReviewCycle(rpcCtx, &talent.UpdateTalentReviewCycleRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-review-cycle")
	}
	return toModelTalentReviewCycle(resp.GetValue())
}

// DeleteTalentReviewCycle menghapus talent-review-cycle berdasarkan ID.
func (a *Adapter) DeleteTalentReviewCycle(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_review_cycles", "human_resource.talent_review_cycles.delete", "talent-review-cycle", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentReviewCycle(c, req)
	})
}

func toModelTalentAssessment(row *talent.TalentAssessment) (*model.TalentAssessment, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-assessment response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-assessment ID", nil)
	}
	out := &model.TalentAssessment{ID: nodeID("talent-assessment", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.TalentReviewCycleID = row.GetTalentReviewCycleId()
	out.FrameworkID = row.GetFrameworkId()
	out.AssessmentVersion = int(row.GetAssessmentVersion())
	out.PerformanceReviewID = row.PerformanceReviewId
	out.Status = row.GetStatus()
	out.AssessedByPrincipalID = row.GetAssessedByPrincipalId()
	out.AssessedAt = pbTimePtr(row.GetAssessedAt())
	out.EvidenceDocumentVersionReferenceID = row.GetEvidenceDocumentVersionReferenceId()
	out.EvidenceDigest = pbBase64StringValue(row.GetEvidenceDigest())
	out.SegmentCode = row.SegmentCode
	out.FinalizedByPrincipalID = row.FinalizedByPrincipalId
	out.FinalizedAt = pbTimePtr(row.GetFinalizedAt())
	return out, nil
}

// TalentAssessment membaca detail talent-assessment berdasarkan ID.
func (a *Adapter) TalentAssessment(ctx context.Context, rawID string) (*model.TalentAssessment, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_assessments", "human_resource.talent_assessments.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment", "talent-assessment")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentAssessmentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-assessment")
	}
	return toModelTalentAssessment(resp.GetValue())
}

// TalentAssessments mendaftar dan memfilter talent-assessment dengan paginasi.
func (a *Adapter) TalentAssessments(ctx context.Context, limit, offset int, f_worker_id *string, f_talent_review_cycle_id *string, f_framework_id *string, f_performance_review_id *string, f_status *string, f_assessed_by_principal_id *string, f_evidence_document_version_reference_id *string, f_segment_code *string, f_finalized_by_principal_id *string) (*Page[model.TalentAssessment], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_assessments", "human_resource.talent_assessments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-assessment pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_talent_review_cycle_id != nil, f_framework_id != nil, f_performance_review_id != nil, f_status != nil, f_assessed_by_principal_id != nil, f_evidence_document_version_reference_id != nil, f_segment_code != nil, f_finalized_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentAssessmentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_talent_review_cycle_id, "talent_review_cycle_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByTalentReviewCycleID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_framework_id, "framework_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByFrameworkID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_performance_review_id, "performance_review_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByPerformanceReviewID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_assessed_by_principal_id, "assessed_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByAssessedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_segment_code, "segment_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentBySegmentCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_finalized_by_principal_id, "finalized_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentByFinalizedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentAssessments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-assessment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-assessment", toModelTalentAssessment)
}

// CreateTalentAssessment membuat talent-assessment baru melalui HRM.
func (a *Adapter) CreateTalentAssessment(ctx context.Context, input model.CreateTalentAssessmentInput) (*model.TalentAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessments", "create", "human_resource.talent_assessments.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentAssessmentInput{}
	patch.WorkerId = input.WorkerID
	patch.TalentReviewCycleId = input.TalentReviewCycleID
	patch.FrameworkId = input.FrameworkID
	patch.AssessmentVersion = optionalInt32Ptr(input.AssessmentVersion)
	patch.PerformanceReviewId = input.PerformanceReviewID
	patch.Status = input.Status
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.EvidenceDigest, "evidence_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EvidenceDigest = b
	}
	patch.SegmentCode = input.SegmentCode
	patch.FinalizedByPrincipalId = input.FinalizedByPrincipalID
	patch.FinalizedAt = optionalTimestamp(input.FinalizedAt)
	resp, err := a.talent.CreateTalentAssessment(rpcCtx, &talent.CreateTalentAssessmentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment")
	}
	return toModelTalentAssessment(resp.GetValue())
}

// UpdateTalentAssessment memperbarui talent-assessment berdasarkan FieldMask.
func (a *Adapter) UpdateTalentAssessment(ctx context.Context, rawID string, input model.UpdateTalentAssessmentInput) (*model.TalentAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessments", "update", "human_resource.talent_assessments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment", "talent-assessment")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentAssessmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "talent_review_cycle_id", input.TalentReviewCycleID, &patch.TalentReviewCycleId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "framework_id", input.FrameworkID, &patch.FrameworkId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "assessment_version", input.AssessmentVersion, &patch.AssessmentVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "performance_review_id", input.PerformanceReviewID, &patch.PerformanceReviewId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "evidence_digest", input.EvidenceDigest, &patch.EvidenceDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "segment_code", input.SegmentCode, &patch.SegmentCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "finalized_by_principal_id", input.FinalizedByPrincipalID, &patch.FinalizedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "finalized_at", input.FinalizedAt, &patch.FinalizedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-assessment update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentAssessment(rpcCtx, &talent.UpdateTalentAssessmentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment")
	}
	return toModelTalentAssessment(resp.GetValue())
}

// DeleteTalentAssessment menghapus talent-assessment berdasarkan ID.
func (a *Adapter) DeleteTalentAssessment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_assessments", "human_resource.talent_assessments.delete", "talent-assessment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentAssessment(c, req)
	})
}

func toModelTalentAssessmentScore(row *talent.TalentAssessmentScore) (*model.TalentAssessmentScore, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-assessment-score response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-assessment-score ID", nil)
	}
	out := &model.TalentAssessmentScore{ID: nodeID("talent-assessment-score", row.GetId()), RawID: row.GetId()}
	out.AssessmentID = row.GetAssessmentId()
	out.FrameworkID = row.GetFrameworkId()
	out.DimensionID = row.GetDimensionId()
	out.Score = int(row.GetScore())
	out.EvidenceDocumentVersionReferenceID = row.GetEvidenceDocumentVersionReferenceId()
	out.EvidenceDigest = pbBase64StringValue(row.GetEvidenceDigest())
	return out, nil
}

// TalentAssessmentScore membaca detail talent-assessment-score berdasarkan ID.
func (a *Adapter) TalentAssessmentScore(ctx context.Context, rawID string) (*model.TalentAssessmentScore, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_assessment_scores", "human_resource.talent_assessment_scores.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment-score", "talent-assessment-score")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentAssessmentScoreByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-assessment-score")
	}
	return toModelTalentAssessmentScore(resp.GetValue())
}

// TalentAssessmentScores mendaftar dan memfilter talent-assessment-score dengan paginasi.
func (a *Adapter) TalentAssessmentScores(ctx context.Context, limit, offset int, f_assessment_id *string, f_framework_id *string, f_dimension_id *string) (*Page[model.TalentAssessmentScore], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_assessment_scores", "human_resource.talent_assessment_scores.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-assessment-score pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_assessment_id != nil, f_framework_id != nil, f_dimension_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentAssessmentScoresResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_assessment_id, "assessment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentScoreByAssessmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_framework_id, "framework_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentScoreByFrameworkID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_dimension_id, "dimension_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentAssessmentScoreByDimensionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentAssessmentScores(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-assessment-score")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-assessment-score", toModelTalentAssessmentScore)
}

// CreateTalentAssessmentScore membuat talent-assessment-score baru melalui HRM.
func (a *Adapter) CreateTalentAssessmentScore(ctx context.Context, input model.CreateTalentAssessmentScoreInput) (*model.TalentAssessmentScore, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessment_scores", "create", "human_resource.talent_assessment_scores.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentAssessmentScoreInput{}
	patch.AssessmentId = input.AssessmentID
	patch.FrameworkId = input.FrameworkID
	patch.DimensionId = input.DimensionID
	patch.Score = optionalInt32Ptr(input.Score)
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.EvidenceDigest, "evidence_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EvidenceDigest = b
	}
	resp, err := a.talent.CreateTalentAssessmentScore(rpcCtx, &talent.CreateTalentAssessmentScoreRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment-score")
	}
	return toModelTalentAssessmentScore(resp.GetValue())
}

// UpdateTalentAssessmentScore memperbarui talent-assessment-score berdasarkan FieldMask.
func (a *Adapter) UpdateTalentAssessmentScore(ctx context.Context, rawID string, input model.UpdateTalentAssessmentScoreInput) (*model.TalentAssessmentScore, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_assessment_scores", "update", "human_resource.talent_assessment_scores.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-assessment-score", "talent-assessment-score")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentAssessmentScoreInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "assessment_id", input.AssessmentID, &patch.AssessmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "framework_id", input.FrameworkID, &patch.FrameworkId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "dimension_id", input.DimensionID, &patch.DimensionId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "score", input.Score, &patch.Score, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "evidence_digest", input.EvidenceDigest, &patch.EvidenceDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-assessment-score update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentAssessmentScore(rpcCtx, &talent.UpdateTalentAssessmentScoreRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-assessment-score")
	}
	return toModelTalentAssessmentScore(resp.GetValue())
}

// DeleteTalentAssessmentScore menghapus talent-assessment-score berdasarkan ID.
func (a *Adapter) DeleteTalentAssessmentScore(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_assessment_scores", "human_resource.talent_assessment_scores.delete", "talent-assessment-score", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentAssessmentScore(c, req)
	})
}

func toModelTalentCalibrationSession(row *talent.TalentCalibrationSession) (*model.TalentCalibrationSession, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-calibration-session response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-calibration-session ID", nil)
	}
	out := &model.TalentCalibrationSession{ID: nodeID("talent-calibration-session", row.GetId()), RawID: row.GetId()}
	out.TalentReviewCycleID = row.GetTalentReviewCycleId()
	out.SessionCode = row.GetSessionCode()
	out.ChairPrincipalID = row.GetChairPrincipalId()
	out.ScheduledAt = pbTimePtr(row.GetScheduledAt())
	out.Status = row.GetStatus()
	out.ClosedAt = pbTimePtr(row.GetClosedAt())
	out.MinutesDocumentVersionReferenceID = row.MinutesDocumentVersionReferenceId
	return out, nil
}

// TalentCalibrationSession membaca detail talent-calibration-session berdasarkan ID.
func (a *Adapter) TalentCalibrationSession(ctx context.Context, rawID string) (*model.TalentCalibrationSession, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_calibration_sessions", "human_resource.talent_calibration_sessions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-calibration-session", "talent-calibration-session")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentCalibrationSessionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-calibration-session")
	}
	return toModelTalentCalibrationSession(resp.GetValue())
}

// TalentCalibrationSessions mendaftar dan memfilter talent-calibration-session dengan paginasi.
func (a *Adapter) TalentCalibrationSessions(ctx context.Context, limit, offset int, f_talent_review_cycle_id *string, f_session_code *string, f_chair_principal_id *string, f_status *string, f_minutes_document_version_reference_id *string) (*Page[model.TalentCalibrationSession], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_calibration_sessions", "human_resource.talent_calibration_sessions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-calibration-session pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_talent_review_cycle_id != nil, f_session_code != nil, f_chair_principal_id != nil, f_status != nil, f_minutes_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentCalibrationSessionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_talent_review_cycle_id, "talent_review_cycle_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationSessionByTalentReviewCycleID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_session_code, "session_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationSessionBySessionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_chair_principal_id, "chair_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationSessionByChairPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationSessionByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_minutes_document_version_reference_id, "minutes_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationSessionByMinutesDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentCalibrationSessions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-calibration-session")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-calibration-session", toModelTalentCalibrationSession)
}

// CreateTalentCalibrationSession membuat talent-calibration-session baru melalui HRM.
func (a *Adapter) CreateTalentCalibrationSession(ctx context.Context, input model.CreateTalentCalibrationSessionInput) (*model.TalentCalibrationSession, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_calibration_sessions", "create", "human_resource.talent_calibration_sessions.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentCalibrationSessionInput{}
	patch.TalentReviewCycleId = input.TalentReviewCycleID
	patch.SessionCode = input.SessionCode
	patch.ChairPrincipalId = input.ChairPrincipalID
	patch.ScheduledAt = optionalTimestamp(input.ScheduledAt)
	patch.Status = input.Status
	patch.ClosedAt = optionalTimestamp(input.ClosedAt)
	patch.MinutesDocumentVersionReferenceId = input.MinutesDocumentVersionReferenceID
	resp, err := a.talent.CreateTalentCalibrationSession(rpcCtx, &talent.CreateTalentCalibrationSessionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-calibration-session")
	}
	return toModelTalentCalibrationSession(resp.GetValue())
}

// UpdateTalentCalibrationSession memperbarui talent-calibration-session berdasarkan FieldMask.
func (a *Adapter) UpdateTalentCalibrationSession(ctx context.Context, rawID string, input model.UpdateTalentCalibrationSessionInput) (*model.TalentCalibrationSession, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_calibration_sessions", "update", "human_resource.talent_calibration_sessions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-calibration-session", "talent-calibration-session")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentCalibrationSessionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "talent_review_cycle_id", input.TalentReviewCycleID, &patch.TalentReviewCycleId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "session_code", input.SessionCode, &patch.SessionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "chair_principal_id", input.ChairPrincipalID, &patch.ChairPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "scheduled_at", input.ScheduledAt, &patch.ScheduledAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "closed_at", input.ClosedAt, &patch.ClosedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "minutes_document_version_reference_id", input.MinutesDocumentVersionReferenceID, &patch.MinutesDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-calibration-session update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentCalibrationSession(rpcCtx, &talent.UpdateTalentCalibrationSessionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-calibration-session")
	}
	return toModelTalentCalibrationSession(resp.GetValue())
}

// DeleteTalentCalibrationSession menghapus talent-calibration-session berdasarkan ID.
func (a *Adapter) DeleteTalentCalibrationSession(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_calibration_sessions", "human_resource.talent_calibration_sessions.delete", "talent-calibration-session", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentCalibrationSession(c, req)
	})
}

func toModelTalentCalibrationDecision(row *talent.TalentCalibrationDecision) (*model.TalentCalibrationDecision, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-calibration-decision response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-calibration-decision ID", nil)
	}
	out := &model.TalentCalibrationDecision{ID: nodeID("talent-calibration-decision", row.GetId()), RawID: row.GetId()}
	out.SessionID = row.GetSessionId()
	out.TalentReviewCycleID = row.GetTalentReviewCycleId()
	out.AssessmentID = row.GetAssessmentId()
	out.DimensionID = row.GetDimensionId()
	out.OriginalScore = int(row.GetOriginalScore())
	out.CalibratedScore = int(row.GetCalibratedScore())
	out.RationaleDocumentVersionReferenceID = row.GetRationaleDocumentVersionReferenceId()
	out.RationaleDigest = pbBase64StringValue(row.GetRationaleDigest())
	out.DecidedByPrincipalID = row.GetDecidedByPrincipalId()
	out.DecidedAt = pbTimePtr(row.GetDecidedAt())
	return out, nil
}

// TalentCalibrationDecision membaca detail talent-calibration-decision berdasarkan ID.
func (a *Adapter) TalentCalibrationDecision(ctx context.Context, rawID string) (*model.TalentCalibrationDecision, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_calibration_decisions", "human_resource.talent_calibration_decisions.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-calibration-decision", "talent-calibration-decision")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentCalibrationDecisionByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-calibration-decision")
	}
	return toModelTalentCalibrationDecision(resp.GetValue())
}

// TalentCalibrationDecisions mendaftar dan memfilter talent-calibration-decision dengan paginasi.
func (a *Adapter) TalentCalibrationDecisions(ctx context.Context, limit, offset int, f_session_id *string, f_talent_review_cycle_id *string, f_assessment_id *string, f_dimension_id *string, f_rationale_document_version_reference_id *string, f_decided_by_principal_id *string) (*Page[model.TalentCalibrationDecision], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_calibration_decisions", "human_resource.talent_calibration_decisions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-calibration-decision pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_session_id != nil, f_talent_review_cycle_id != nil, f_assessment_id != nil, f_dimension_id != nil, f_rationale_document_version_reference_id != nil, f_decided_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentCalibrationDecisionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_session_id, "session_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationDecisionBySessionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_talent_review_cycle_id, "talent_review_cycle_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationDecisionByTalentReviewCycleID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_assessment_id, "assessment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationDecisionByAssessmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_dimension_id, "dimension_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationDecisionByDimensionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_rationale_document_version_reference_id, "rationale_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationDecisionByRationaleDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_decided_by_principal_id, "decided_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentCalibrationDecisionByDecidedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentCalibrationDecisions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-calibration-decision")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-calibration-decision", toModelTalentCalibrationDecision)
}

// CreateTalentCalibrationDecision membuat talent-calibration-decision baru melalui HRM.
func (a *Adapter) CreateTalentCalibrationDecision(ctx context.Context, input model.CreateTalentCalibrationDecisionInput) (*model.TalentCalibrationDecision, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_calibration_decisions", "create", "human_resource.talent_calibration_decisions.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentCalibrationDecisionInput{}
	patch.SessionId = input.SessionID
	patch.TalentReviewCycleId = input.TalentReviewCycleID
	patch.AssessmentId = input.AssessmentID
	patch.DimensionId = input.DimensionID
	patch.OriginalScore = optionalInt32Ptr(input.OriginalScore)
	patch.CalibratedScore = optionalInt32Ptr(input.CalibratedScore)
	patch.RationaleDocumentVersionReferenceId = input.RationaleDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.RationaleDigest, "rationale_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.RationaleDigest = b
	}
	resp, err := a.talent.CreateTalentCalibrationDecision(rpcCtx, &talent.CreateTalentCalibrationDecisionRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-calibration-decision")
	}
	return toModelTalentCalibrationDecision(resp.GetValue())
}

// UpdateTalentCalibrationDecision memperbarui talent-calibration-decision berdasarkan FieldMask.
func (a *Adapter) UpdateTalentCalibrationDecision(ctx context.Context, rawID string, input model.UpdateTalentCalibrationDecisionInput) (*model.TalentCalibrationDecision, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_calibration_decisions", "update", "human_resource.talent_calibration_decisions.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-calibration-decision", "talent-calibration-decision")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentCalibrationDecisionInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "session_id", input.SessionID, &patch.SessionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "talent_review_cycle_id", input.TalentReviewCycleID, &patch.TalentReviewCycleId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assessment_id", input.AssessmentID, &patch.AssessmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "dimension_id", input.DimensionID, &patch.DimensionId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "original_score", input.OriginalScore, &patch.OriginalScore, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "calibrated_score", input.CalibratedScore, &patch.CalibratedScore, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rationale_document_version_reference_id", input.RationaleDocumentVersionReferenceID, &patch.RationaleDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "rationale_digest", input.RationaleDigest, &patch.RationaleDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-calibration-decision update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentCalibrationDecision(rpcCtx, &talent.UpdateTalentCalibrationDecisionRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-calibration-decision")
	}
	return toModelTalentCalibrationDecision(resp.GetValue())
}

// DeleteTalentCalibrationDecision menghapus talent-calibration-decision berdasarkan ID.
func (a *Adapter) DeleteTalentCalibrationDecision(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_calibration_decisions", "human_resource.talent_calibration_decisions.delete", "talent-calibration-decision", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentCalibrationDecision(c, req)
	})
}

func toModelTalentPool(row *talent.TalentPool) (*model.TalentPool, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-pool response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-pool ID", nil)
	}
	out := &model.TalentPool{ID: nodeID("talent-pool", row.GetId()), RawID: row.GetId()}
	out.PoolCode = row.GetPoolCode()
	out.PoolVersion = int(row.GetPoolVersion())
	out.PoolName = row.GetPoolName()
	out.LegalEntityReferenceID = row.GetLegalEntityReferenceId()
	out.TargetJobID = row.TargetJobId
	out.EligibilityDocumentVersionReferenceID = row.GetEligibilityDocumentVersionReferenceId()
	out.EligibilityDigest = pbBase64StringValue(row.GetEligibilityDigest())
	out.ValidFrom = pbTimePtr(row.GetValidFrom())
	out.ValidTo = pbTimePtr(row.GetValidTo())
	out.OwnerPrincipalID = row.GetOwnerPrincipalId()
	return out, nil
}

// TalentPool membaca detail talent-pool berdasarkan ID.
func (a *Adapter) TalentPool(ctx context.Context, rawID string) (*model.TalentPool, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_pools", "human_resource.talent_pools.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool", "talent-pool")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentPoolByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-pool")
	}
	return toModelTalentPool(resp.GetValue())
}

// TalentPools mendaftar dan memfilter talent-pool dengan paginasi.
func (a *Adapter) TalentPools(ctx context.Context, limit, offset int, f_pool_code *string, f_legal_entity_reference_id *string, f_target_job_id *string, f_eligibility_document_version_reference_id *string, f_owner_principal_id *string) (*Page[model.TalentPool], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_pools", "human_resource.talent_pools.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-pool pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_pool_code != nil, f_legal_entity_reference_id != nil, f_target_job_id != nil, f_eligibility_document_version_reference_id != nil, f_owner_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentPoolsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_pool_code, "pool_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolByPoolCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_legal_entity_reference_id, "legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_target_job_id, "target_job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolByTargetJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_eligibility_document_version_reference_id, "eligibility_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolByEligibilityDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_owner_principal_id, "owner_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolByOwnerPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentPools(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-pool")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-pool", toModelTalentPool)
}

// CreateTalentPool membuat talent-pool baru melalui HRM.
func (a *Adapter) CreateTalentPool(ctx context.Context, input model.CreateTalentPoolInput) (*model.TalentPool, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pools", "create", "human_resource.talent_pools.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentPoolInput{}
	patch.PoolCode = input.PoolCode
	patch.PoolVersion = optionalInt32Ptr(input.PoolVersion)
	patch.PoolName = input.PoolName
	patch.LegalEntityReferenceId = input.LegalEntityReferenceID
	patch.TargetJobId = input.TargetJobID
	patch.EligibilityDocumentVersionReferenceId = input.EligibilityDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.EligibilityDigest, "eligibility_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EligibilityDigest = b
	}
	patch.ValidFrom = optionalTimestamp(input.ValidFrom)
	patch.ValidTo = optionalTimestamp(input.ValidTo)
	patch.OwnerPrincipalId = input.OwnerPrincipalID
	resp, err := a.talent.CreateTalentPool(rpcCtx, &talent.CreateTalentPoolRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool")
	}
	return toModelTalentPool(resp.GetValue())
}

// UpdateTalentPool memperbarui talent-pool berdasarkan FieldMask.
func (a *Adapter) UpdateTalentPool(ctx context.Context, rawID string, input model.UpdateTalentPoolInput) (*model.TalentPool, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pools", "update", "human_resource.talent_pools.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool", "talent-pool")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentPoolInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "pool_code", input.PoolCode, &patch.PoolCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "pool_version", input.PoolVersion, &patch.PoolVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "pool_name", input.PoolName, &patch.PoolName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_entity_reference_id", input.LegalEntityReferenceID, &patch.LegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "target_job_id", input.TargetJobID, &patch.TargetJobId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "eligibility_document_version_reference_id", input.EligibilityDocumentVersionReferenceID, &patch.EligibilityDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "eligibility_digest", input.EligibilityDigest, &patch.EligibilityDigest, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "owner_principal_id", input.OwnerPrincipalID, &patch.OwnerPrincipalId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-pool update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentPool(rpcCtx, &talent.UpdateTalentPoolRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool")
	}
	return toModelTalentPool(resp.GetValue())
}

// DeleteTalentPool menghapus talent-pool berdasarkan ID.
func (a *Adapter) DeleteTalentPool(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_pools", "human_resource.talent_pools.delete", "talent-pool", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentPool(c, req)
	})
}

func toModelTalentPoolMembership(row *talent.TalentPoolMembership) (*model.TalentPoolMembership, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-pool-membership response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource talent-pool-membership ID", nil)
	}
	out := &model.TalentPoolMembership{ID: nodeID("talent-pool-membership", row.GetId()), RawID: row.GetId()}
	out.TalentPoolID = row.GetTalentPoolId()
	out.WorkerID = row.GetWorkerId()
	out.AssessmentID = row.GetAssessmentId()
	out.NominatedByPrincipalID = row.GetNominatedByPrincipalId()
	out.Status = row.GetStatus()
	out.ValidFrom = pbTimePtr(row.GetValidFrom())
	out.ValidTo = pbTimePtr(row.GetValidTo())
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	out.ExitReasonCode = row.ExitReasonCode
	return out, nil
}

// TalentPoolMembership membaca detail talent-pool-membership berdasarkan ID.
func (a *Adapter) TalentPoolMembership(ctx context.Context, rawID string) (*model.TalentPoolMembership, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_pool_memberships", "human_resource.talent_pool_memberships.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool-membership", "talent-pool-membership")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.talent.GetTalentPoolMembershipByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "talent-pool-membership")
	}
	return toModelTalentPoolMembership(resp.GetValue())
}

// TalentPoolMemberships mendaftar dan memfilter talent-pool-membership dengan paginasi.
func (a *Adapter) TalentPoolMemberships(ctx context.Context, limit, offset int, f_talent_pool_id *string, f_worker_id *string, f_assessment_id *string, f_nominated_by_principal_id *string, f_status *string, f_approved_by_principal_id *string, f_exit_reason_code *string) (*Page[model.TalentPoolMembership], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_talent_pool_memberships", "human_resource.talent_pool_memberships.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid talent-pool-membership pagination", nil)
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_talent_pool_id != nil, f_worker_id != nil, f_assessment_id != nil, f_nominated_by_principal_id != nil, f_status != nil, f_approved_by_principal_id != nil, f_exit_reason_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *talent.TalentPoolMembershipsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_talent_pool_id, "talent_pool_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByTalentPoolID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_assessment_id, "assessment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByAssessmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_nominated_by_principal_id, "nominated_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByNominatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_exit_reason_code, "exit_reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.talent.FindTalentPoolMembershipByExitReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.talent.GetAllTalentPoolMemberships(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "talent-pool-membership")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "talent-pool-membership", toModelTalentPoolMembership)
}

// CreateTalentPoolMembership membuat talent-pool-membership baru melalui HRM.
func (a *Adapter) CreateTalentPoolMembership(ctx context.Context, input model.CreateTalentPoolMembershipInput) (*model.TalentPoolMembership, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pool_memberships", "create", "human_resource.talent_pool_memberships.create")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentPoolMembershipInput{}
	patch.TalentPoolId = input.TalentPoolID
	patch.WorkerId = input.WorkerID
	patch.AssessmentId = input.AssessmentID
	patch.NominatedByPrincipalId = input.NominatedByPrincipalID
	patch.Status = input.Status
	patch.ValidFrom = optionalTimestamp(input.ValidFrom)
	patch.ValidTo = optionalTimestamp(input.ValidTo)
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	patch.ExitReasonCode = input.ExitReasonCode
	resp, err := a.talent.CreateTalentPoolMembership(rpcCtx, &talent.CreateTalentPoolMembershipRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool-membership")
	}
	return toModelTalentPoolMembership(resp.GetValue())
}

// UpdateTalentPoolMembership memperbarui talent-pool-membership berdasarkan FieldMask.
func (a *Adapter) UpdateTalentPoolMembership(ctx context.Context, rawID string, input model.UpdateTalentPoolMembershipInput) (*model.TalentPoolMembership, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_talent_pool_memberships", "update", "human_resource.talent_pool_memberships.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "talent-pool-membership", "talent-pool-membership")
	if err != nil {
		return nil, err
	}
	if a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &talent.TalentPoolMembershipInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "talent_pool_id", input.TalentPoolID, &patch.TalentPoolId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assessment_id", input.AssessmentID, &patch.AssessmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "nominated_by_principal_id", input.NominatedByPrincipalID, &patch.NominatedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "exit_reason_code", input.ExitReasonCode, &patch.ExitReasonCode, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("talent-pool-membership update is empty", nil)
	}
	resp, err := a.talent.UpdateTalentPoolMembership(rpcCtx, &talent.UpdateTalentPoolMembershipRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "talent-pool-membership")
	}
	return toModelTalentPoolMembership(resp.GetValue())
}

// DeleteTalentPoolMembership menghapus talent-pool-membership berdasarkan ID.
func (a *Adapter) DeleteTalentPoolMembership(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.talent == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_talent_pool_memberships", "human_resource.talent_pool_memberships.delete", "talent-pool-membership", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.talent.DeleteTalentPoolMembership(c, req)
	})
}
