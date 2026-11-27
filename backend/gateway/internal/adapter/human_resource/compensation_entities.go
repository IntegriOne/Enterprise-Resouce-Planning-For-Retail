package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	compensation "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/compensation/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelCompensationBand(row *compensation.CompensationBand) (*model.CompensationBand, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource compensation-band response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource compensation-band ID", nil)
	}
	out := &model.CompensationBand{ID: nodeID("compensation-band", row.GetId()), RawID: row.GetId()}
	out.BandCode = row.GetBandCode()
	out.BandVersion = int(row.GetBandVersion())
	out.JobLevelCode = row.GetJobLevelCode()
	out.JurisdictionCode = row.GetJurisdictionCode()
	out.CurrencyCode = row.GetCurrencyCode()
	out.PayFrequency = row.GetPayFrequency()
	out.MinimumAmount = row.GetMinimumAmount()
	out.MidpointAmount = row.GetMidpointAmount()
	out.MaximumAmount = row.GetMaximumAmount()
	out.ValidFrom = row.GetValidFrom()
	out.ValidTo = row.ValidTo
	out.ApprovalReferenceID = row.ApprovalReferenceId
	return out, nil
}

// CompensationBand membaca detail compensation-band berdasarkan ID.
func (a *Adapter) CompensationBand(ctx context.Context, rawID string) (*model.CompensationBand, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_compensation_bands", "human_resource.compensation_bands.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "compensation-band", "compensation-band")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.compensation.GetCompensationBandByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "compensation-band")
	}
	return toModelCompensationBand(resp.GetItem())
}

// CompensationBands mendaftar dan memfilter compensation-band dengan paginasi.
func (a *Adapter) CompensationBands(ctx context.Context, limit, offset int, f_band_code *string, f_job_level_code *string, f_jurisdiction_code *string, f_currency_code *string, f_approval_reference_id *string) (*Page[model.CompensationBand], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_compensation_bands", "human_resource.compensation_bands.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid compensation-band pagination", nil)
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_band_code != nil, f_job_level_code != nil, f_jurisdiction_code != nil, f_currency_code != nil, f_approval_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *compensation.CompensationBandsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_band_code, "band_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationBandByBandCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_job_level_code, "job_level_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationBandByJobLevelCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_jurisdiction_code, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationBandByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationBandByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_approval_reference_id, "approval_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationBandByApprovalReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.compensation.GetAllCompensationBands(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "compensation-band")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "compensation-band", toModelCompensationBand)
}

// CreateCompensationBand membuat compensation-band baru melalui HRM.
func (a *Adapter) CreateCompensationBand(ctx context.Context, input model.CreateCompensationBandInput) (*model.CompensationBand, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_compensation_bands", "create", "human_resource.compensation_bands.create")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &compensation.CompensationBandInput{}
	patch.BandCode = input.BandCode
	patch.BandVersion = optionalInt32Ptr(input.BandVersion)
	patch.JobLevelCode = input.JobLevelCode
	patch.JurisdictionCode = input.JurisdictionCode
	patch.CurrencyCode = input.CurrencyCode
	patch.PayFrequency = input.PayFrequency
	patch.MinimumAmount = input.MinimumAmount
	patch.MidpointAmount = input.MidpointAmount
	patch.MaximumAmount = input.MaximumAmount
	patch.ValidFrom = input.ValidFrom
	patch.ValidTo = input.ValidTo
	patch.ApprovalReferenceId = input.ApprovalReferenceID
	resp, err := a.compensation.CreateCompensationBand(rpcCtx, &compensation.CreateCompensationBandRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "compensation-band")
	}
	return toModelCompensationBand(resp.GetItem())
}

// UpdateCompensationBand memperbarui compensation-band berdasarkan FieldMask.
func (a *Adapter) UpdateCompensationBand(ctx context.Context, rawID string, input model.UpdateCompensationBandInput) (*model.CompensationBand, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_compensation_bands", "update", "human_resource.compensation_bands.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "compensation-band", "compensation-band")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &compensation.CompensationBandInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "band_code", input.BandCode, &patch.BandCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "band_version", input.BandVersion, &patch.BandVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "job_level_code", input.JobLevelCode, &patch.JobLevelCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "pay_frequency", input.PayFrequency, &patch.PayFrequency, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "minimum_amount", input.MinimumAmount, &patch.MinimumAmount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "midpoint_amount", input.MidpointAmount, &patch.MidpointAmount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "maximum_amount", input.MaximumAmount, &patch.MaximumAmount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approval_reference_id", input.ApprovalReferenceID, &patch.ApprovalReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("compensation-band update is empty", nil)
	}
	resp, err := a.compensation.UpdateCompensationBand(rpcCtx, &compensation.UpdateCompensationBandRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "compensation-band")
	}
	return toModelCompensationBand(resp.GetItem())
}

// DeleteCompensationBand menghapus compensation-band berdasarkan ID.
func (a *Adapter) DeleteCompensationBand(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_compensation_bands", "human_resource.compensation_bands.delete", "compensation-band", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.compensation.DeleteCompensationBand(c, req)
	})
}

func toModelCompensationRecord(row *compensation.CompensationRecord) (*model.CompensationRecord, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource compensation-record response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource compensation-record ID", nil)
	}
	out := &model.CompensationRecord{ID: nodeID("compensation-record", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.AssignmentID = row.GetAssignmentId()
	out.ComponentID = row.GetComponentId()
	out.ComponentCodeSnapshot = row.GetComponentCodeSnapshot()
	out.CompensationBandID = row.CompensationBandId
	out.CompensationType = row.GetCompensationType()
	out.Amount = row.GetAmount()
	out.CurrencyCode = row.GetCurrencyCode()
	out.PayFrequency = row.GetPayFrequency()
	out.EffectiveFrom = row.GetEffectiveFrom()
	out.EffectiveTo = row.EffectiveTo
	out.Status = row.GetStatus()
	out.ReasonCode = row.ReasonCode
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	out.SourceChangeReferenceID = row.SourceChangeReferenceId
	out.AccessClass = row.GetAccessClass()
	return out, nil
}

// CompensationRecord membaca detail compensation-record berdasarkan ID.
func (a *Adapter) CompensationRecord(ctx context.Context, rawID string) (*model.CompensationRecord, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_compensation_records", "human_resource.compensation_records.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "compensation-record", "compensation-record")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.compensation.GetCompensationRecordByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "compensation-record")
	}
	return toModelCompensationRecord(resp.GetItem())
}

// CompensationRecords mendaftar dan memfilter compensation-record dengan paginasi.
func (a *Adapter) CompensationRecords(ctx context.Context, limit, offset int, f_worker_id *string, f_assignment_id *string, f_component_id *string, f_compensation_band_id *string, f_compensation_type *string, f_currency_code *string, f_status *string, f_reason_code *string, f_approved_by_principal_id *string, f_source_change_reference_id *string) (*Page[model.CompensationRecord], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_compensation_records", "human_resource.compensation_records.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid compensation-record pagination", nil)
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_assignment_id != nil, f_component_id != nil, f_compensation_band_id != nil, f_compensation_type != nil, f_currency_code != nil, f_status != nil, f_reason_code != nil, f_approved_by_principal_id != nil, f_source_change_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *compensation.CompensationRecordsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_assignment_id, "assignment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByAssignmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_component_id, "component_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByComponentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_compensation_band_id, "compensation_band_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByCompensationBandID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_compensation_type, "compensation_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByCompensationType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_reason_code, "reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_source_change_reference_id, "source_change_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.compensation.FindCompensationRecordBySourceChangeReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.compensation.GetAllCompensationRecords(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "compensation-record")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "compensation-record", toModelCompensationRecord)
}

// CreateCompensationRecord membuat compensation-record baru melalui HRM.
func (a *Adapter) CreateCompensationRecord(ctx context.Context, input model.CreateCompensationRecordInput) (*model.CompensationRecord, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_compensation_records", "create", "human_resource.compensation_records.create")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &compensation.CompensationRecordInput{}
	patch.WorkerId = input.WorkerID
	patch.AssignmentId = input.AssignmentID
	patch.ComponentId = input.ComponentID
	patch.ComponentCodeSnapshot = input.ComponentCodeSnapshot
	patch.CompensationBandId = input.CompensationBandID
	patch.CompensationType = input.CompensationType
	patch.Amount = input.Amount
	patch.CurrencyCode = input.CurrencyCode
	patch.PayFrequency = input.PayFrequency
	patch.EffectiveFrom = input.EffectiveFrom
	patch.EffectiveTo = input.EffectiveTo
	patch.Status = input.Status
	patch.ReasonCode = input.ReasonCode
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	patch.SourceChangeReferenceId = input.SourceChangeReferenceID
	patch.AccessClass = input.AccessClass
	resp, err := a.compensation.CreateCompensationRecord(rpcCtx, &compensation.CreateCompensationRecordRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "compensation-record")
	}
	return toModelCompensationRecord(resp.GetItem())
}

// UpdateCompensationRecord memperbarui compensation-record berdasarkan FieldMask.
func (a *Adapter) UpdateCompensationRecord(ctx context.Context, rawID string, input model.UpdateCompensationRecordInput) (*model.CompensationRecord, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_compensation_records", "update", "human_resource.compensation_records.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "compensation-record", "compensation-record")
	if err != nil {
		return nil, err
	}
	if a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &compensation.CompensationRecordInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assignment_id", input.AssignmentID, &patch.AssignmentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "component_id", input.ComponentID, &patch.ComponentId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "component_code_snapshot", input.ComponentCodeSnapshot, &patch.ComponentCodeSnapshot, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "compensation_band_id", input.CompensationBandID, &patch.CompensationBandId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "compensation_type", input.CompensationType, &patch.CompensationType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "amount", input.Amount, &patch.Amount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "pay_frequency", input.PayFrequency, &patch.PayFrequency, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reason_code", input.ReasonCode, &patch.ReasonCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "source_change_reference_id", input.SourceChangeReferenceID, &patch.SourceChangeReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "access_class", input.AccessClass, &patch.AccessClass, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("compensation-record update is empty", nil)
	}
	resp, err := a.compensation.UpdateCompensationRecord(rpcCtx, &compensation.UpdateCompensationRecordRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "compensation-record")
	}
	return toModelCompensationRecord(resp.GetItem())
}

// DeleteCompensationRecord menghapus compensation-record berdasarkan ID.
func (a *Adapter) DeleteCompensationRecord(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.compensation == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_compensation_records", "human_resource.compensation_records.delete", "compensation-record", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.compensation.DeleteCompensationRecord(c, req)
	})
}
