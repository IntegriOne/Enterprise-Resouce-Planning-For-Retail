package human_resource

import (
	"context"
	"crypto/sha256"
	"math"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// EmploymentContracts membaca metadata kontrak pekerja dengan izin sensitif.
func (a *Adapter) EmploymentContracts(ctx context.Context, rawWorkerID string, limit, offset int) (*Page[model.EmploymentContract], error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.employment_contracts.read")
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawWorkerID, workerGIDPrefix))
	if err != nil || id == uuid.Nil || !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid employment contracts request", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	workerID := id.String()
	response, err := a.personnel.FindEmploymentContractByWorkerID(rpcCtx, &common.StringFinderRequest{
		Value: &workerID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "employment contract")
	}
	items, page := response.GetItems(), response.GetPage()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource employment contract page", nil)
	}
	result := &Page[model.EmploymentContract]{
		Items: make([]*model.EmploymentContract, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, item := range items {
		if item.GetWorkerId() != workerID {
			return nil, middleware.ErrUnavailable("invalid human resource employment contract worker", nil)
		}
		contract, err := toModelEmploymentContract(item)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, contract)
	}
	return result, nil
}

// EmploymentContract membaca metadata satu kontrak kerja berdasarkan ID.
func (a *Adapter) EmploymentContract(ctx context.Context, rawID string) (*model.EmploymentContract, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.employment_contracts.read")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "employment-contract")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employment contract ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.GetEmploymentContractByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "employment contract")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource employment contract response", nil)
	}
	return toModelEmploymentContract(item)
}

func toModelEmploymentContract(item *personnel.EmploymentContract) (*model.EmploymentContract, error) {
	contractID, contractErr := uuid.Parse(item.GetId())
	workerID, workerErr := uuid.Parse(item.GetWorkerId())
	entityID, entityErr := uuid.Parse(item.GetEmployingLegalEntityReferenceId())
	if workerErr != nil || workerID == uuid.Nil || contractErr != nil || contractID == uuid.Nil || entityErr != nil || entityID == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource employment contract reference", nil)
	}
	contract := &model.EmploymentContract{
		ID: nodeID("employment-contract", item.GetId()), RawID: item.GetId(), WorkerID: workerGIDPrefix + workerID.String(), ContractNumber: item.GetContractNumber(),
		ContractType: item.GetContractType(), ContractStatus: item.GetContractStatus(),
		EmployingLegalEntityReferenceID: item.GetEmployingLegalEntityReferenceId(),
		EffectiveFrom:                   item.GetEffectiveFrom(), EffectiveTo: item.EffectiveTo,
	}
	if stamp := item.GetSignedAt(); stamp != nil {
		if stamp.CheckValid() != nil {
			return nil, middleware.ErrUnavailable("invalid human resource employment contract timestamp", nil)
		}
		value := stamp.AsTime()
		contract.SignedAt = &value
	}
	return contract, nil
}

// CreateEmploymentContract mencatat kontrak kerja draf.
func (a *Adapter) CreateEmploymentContract(ctx context.Context, input model.CreateEmploymentContractInput) (*model.EmploymentContract, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.employment_contracts.create")
	if err != nil {
		return nil, err
	}
	worker, workerErr := uuid.Parse(strings.TrimPrefix(input.WorkerID, workerGIDPrefix))
	entity, entityErr := parseNodeID(input.EmployingLegalEntityReferenceID, "legal-entity")
	if workerErr != nil || worker == uuid.Nil || entityErr != nil || entity == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employment contract references", nil)
	}
	var notice *int32
	if input.NoticePeriodDays != nil {
		if *input.NoticePeriodDays < 0 || *input.NoticePeriodDays > math.MaxInt32 {
			return nil, middleware.ErrInvalidArgument("invalid notice period", nil)
		}
		days := int32(*input.NoticePeriodDays)
		notice = &days
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	workerID, entityID := worker.String(), entity.String()
	response, err := a.personnel.CreateEmploymentContract(rpcCtx, &personnel.CreateEmploymentContractRequest{
		Input: &personnel.EmploymentContractInput{
			WorkerId: &workerID, EmployingLegalEntityReferenceId: &entityID,
			ContractNumber: &input.ContractNumber, ContractType: &input.ContractType,
			EffectiveFrom: &input.EffectiveFrom, EffectiveTo: input.EffectiveTo,
			ProbationEndDate: input.ProbationEndDate, NoticePeriodDays: notice,
			WorkCountryCode: &input.WorkCountryCode, GoverningJurisdictionCode: &input.GoverningJurisdictionCode,
			CollectiveAgreementId: input.CollectiveAgreementID,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("employment contract conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "employment contract")
	}
	item := response.GetItem()
	if item.GetWorkerId() != workerID || item.GetEmployingLegalEntityReferenceId() != entityID ||
		item.GetContractStatus() != "draft" || item.GetSignedAt() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource draft contract response", nil)
	}
	return toModelEmploymentContract(item)
}

// RecordSignedEmploymentContract meminta HRM memverifikasi bukti tanda tangan DRM.
func (a *Adapter) RecordSignedEmploymentContract(ctx context.Context, rawID, rawVersionID string) (*model.EmploymentContract, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.employment_contracts.record_signature")
	if err != nil {
		return nil, err
	}
	id, idErr := parseNodeID(rawID, "employment-contract")
	versionID, versionErr := uuid.Parse(rawVersionID)
	if idErr != nil || id == uuid.Nil || versionErr != nil || versionID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employment contract signature references", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.RecordSignedEmploymentContract(rpcCtx, &personnel.RecordSignedEmploymentContractRequest{
		ContractId: id.String(), DocumentVersionId: versionID.String(),
	})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("employment contract does not meet signature requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "employment contract")
	}
	item := response.GetItem()
	documentID, documentErr := uuid.Parse(item.GetSignedDocumentReferenceId())
	signedAt := item.GetSignedAt()
	if item.GetId() != id.String() || item.GetContractStatus() != "signed" ||
		item.GetSignedDocumentVersionReferenceId() != versionID.String() ||
		documentErr != nil || documentID == uuid.Nil || len(item.GetSignedContentDigest()) != sha256.Size ||
		signedAt == nil || !signedAt.IsValid() || signedAt.AsTime().IsZero() {
		return nil, middleware.ErrUnavailable("invalid human resource signed contract response", nil)
	}
	return toModelEmploymentContract(item)
}

// DeleteEmploymentContract meminta HRM menghapus kontrak draf tanpa tanda tangan.
func (a *Adapter) DeleteEmploymentContract(ctx context.Context, rawID string) (*bool, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.employment_contracts.delete")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "employment-contract")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employment contract ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.DeleteEmploymentContract(rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("employment contract does not meet deletion requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "employment contract")
	}
	if response == nil {
		return nil, middleware.ErrUnavailable("invalid human resource contract deletion response", nil)
	}
	deleted := true
	return &deleted, nil
}

// UpdateEmploymentContract mengoreksi field yang dikirim pada kontrak draf.
func (a *Adapter) UpdateEmploymentContract(ctx context.Context, rawID string, input model.UpdateEmploymentContractInput) (*model.EmploymentContract, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.employment_contracts.update")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "employment-contract")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid employment contract ID", nil)
	}
	patch := &personnel.EmploymentContractInput{}
	mask := &fieldmaskpb.FieldMask{}
	for _, field := range []struct {
		path     string
		input    graphql.Omittable[*string]
		target   **string
		nullable bool
	}{
		{"contract_number", input.ContractNumber, &patch.ContractNumber, false},
		{"contract_type", input.ContractType, &patch.ContractType, false},
		{"effective_from", input.EffectiveFrom, &patch.EffectiveFrom, false},
		{"effective_to", input.EffectiveTo, &patch.EffectiveTo, true},
		{"probation_end_date", input.ProbationEndDate, &patch.ProbationEndDate, true},
		{"work_country_code", input.WorkCountryCode, &patch.WorkCountryCode, false},
		{"governing_jurisdiction_code", input.GoverningJurisdictionCode, &patch.GoverningJurisdictionCode, false},
		{"collective_agreement_id", input.CollectiveAgreementID, &patch.CollectiveAgreementId, true},
	} {
		value, set := field.input.ValueOK()
		if !set {
			continue
		}
		if value == nil && !field.nullable {
			return nil, middleware.ErrInvalidArgument("required employment contract field cannot be null", nil)
		}
		*field.target = value
		mask.Paths = append(mask.Paths, field.path)
	}
	if days, set := input.NoticePeriodDays.ValueOK(); set {
		if days != nil {
			if *days < 0 || *days > math.MaxInt32 {
				return nil, middleware.ErrInvalidArgument("invalid notice period", nil)
			}
			value := int32(*days)
			patch.NoticePeriodDays = &value
		}
		mask.Paths = append(mask.Paths, "notice_period_days")
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("employment contract update is empty", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.UpdateEmploymentContract(rpcCtx, &personnel.UpdateEmploymentContractRequest{
		Id: id.String(), Input: patch, UpdateMask: mask,
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("employment contract conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "employment contract")
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetContractStatus() != "draft" || item.GetSignedAt() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource draft contract response", nil)
	}
	return toModelEmploymentContract(item)
}
