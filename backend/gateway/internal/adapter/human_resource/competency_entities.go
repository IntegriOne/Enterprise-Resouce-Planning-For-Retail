package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	competency "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelWorkerQualification(row *competency.WorkerQualification) (*model.WorkerQualification, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-qualification response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-qualification ID", nil)
	}
	out := &model.WorkerQualification{ID: nodeID("worker-qualification", row.GetId()), RawID: row.GetId()}
	out.Title = row.GetTitle()
	out.ObtainedOn = row.ObtainedOn
	out.VerifiedByPrincipalID = row.VerifiedByPrincipalId
	return out, nil
}

// WorkerQualification membaca detail worker-qualification berdasarkan ID.
func (a *Adapter) WorkerQualification(ctx context.Context, rawID string) (*model.WorkerQualification, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_qualifications", "human_resource.worker_qualifications.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-qualification", "worker-qualification")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.competency.GetWorkerQualificationByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "worker-qualification")
	}
	return toModelWorkerQualification(resp.GetItem())
}

// WorkerQualifications mendaftar dan memfilter worker-qualification dengan paginasi.
func (a *Adapter) WorkerQualifications(ctx context.Context, limit, offset int, f_worker_id *string, f_qualification_type *string, f_qualification_code *string, f_issuer_party_reference_id *string, f_credential_reference_id *string, f_verification_status *string, f_verified_by_principal_id *string) (*Page[model.WorkerQualification], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_qualifications", "human_resource.worker_qualifications.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker-qualification pagination", nil)
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_qualification_type != nil, f_qualification_code != nil, f_issuer_party_reference_id != nil, f_credential_reference_id != nil, f_verification_status != nil, f_verified_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *competency.WorkerQualificationsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_qualification_type, "qualification_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByQualificationType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_qualification_code, "qualification_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByQualificationCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_issuer_party_reference_id, "issuer_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByIssuerPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_credential_reference_id, "credential_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByCredentialReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_verification_status, "verification_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByVerificationStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_verified_by_principal_id, "verified_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerQualificationByVerifiedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.competency.GetAllWorkerQualifications(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker-qualification")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker-qualification", toModelWorkerQualification)
}

// CreateWorkerQualification membuat worker-qualification baru melalui HRM.
func (a *Adapter) CreateWorkerQualification(ctx context.Context, input model.CreateWorkerQualificationInput) (*model.WorkerQualification, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_qualifications", "create", "human_resource.worker_qualifications.create")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.WorkerQualificationInput{}
	patch.WorkerId = input.WorkerID
	patch.Title = input.Title
	patch.ObtainedOn = input.ObtainedOn
	patch.VerifiedByPrincipalId = input.VerifiedByPrincipalID
	resp, err := a.competency.CreateWorkerQualification(rpcCtx, &competency.CreateWorkerQualificationRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-qualification")
	}
	return toModelWorkerQualification(resp.GetItem())
}

// UpdateWorkerQualification memperbarui worker-qualification berdasarkan FieldMask.
func (a *Adapter) UpdateWorkerQualification(ctx context.Context, rawID string, input model.UpdateWorkerQualificationInput) (*model.WorkerQualification, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_qualifications", "update", "human_resource.worker_qualifications.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-qualification", "worker-qualification")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.WorkerQualificationInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "title", input.Title, &patch.Title, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "obtained_on", input.ObtainedOn, &patch.ObtainedOn, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "verified_by_principal_id", input.VerifiedByPrincipalID, &patch.VerifiedByPrincipalId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("worker-qualification update is empty", nil)
	}
	resp, err := a.competency.UpdateWorkerQualification(rpcCtx, &competency.UpdateWorkerQualificationRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-qualification")
	}
	return toModelWorkerQualification(resp.GetItem())
}

// DeleteWorkerQualification menghapus worker-qualification berdasarkan ID.
func (a *Adapter) DeleteWorkerQualification(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_worker_qualifications", "human_resource.worker_qualifications.delete", "worker-qualification", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.competency.DeleteWorkerQualification(c, req)
	})
}

func toModelCompetency(row *competency.Competency) (*model.Competency, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource competency response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource competency ID", nil)
	}
	out := &model.Competency{ID: nodeID("competency", row.GetId()), RawID: row.GetId()}
	if doc := row.GetProficiencyScale(); doc != nil {
		out.ProficiencyScale = pbJSONString(doc.GetUtf8())
	}
	return out, nil
}

// Competency membaca detail competency berdasarkan ID.
func (a *Adapter) Competency(ctx context.Context, rawID string) (*model.Competency, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_competencies", "human_resource.competencies.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "competency", "competency")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.competency.GetCompetencyByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "competency")
	}
	return toModelCompetency(resp.GetItem())
}

// Competencies mendaftar dan memfilter competency dengan paginasi.
func (a *Adapter) Competencies(ctx context.Context, limit, offset int, f_competency_code *string, f_category_code *string) (*Page[model.Competency], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_competencies", "human_resource.competencies.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid competency pagination", nil)
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_competency_code != nil, f_category_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *competency.CompetenciesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_competency_code, "competency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindCompetencyByCompetencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_category_code, "category_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindCompetencyByCategoryCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.competency.GetAllCompetencies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "competency")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "competency", toModelCompetency)
}

// CreateCompetency membuat competency baru melalui HRM.
func (a *Adapter) CreateCompetency(ctx context.Context, input model.CreateCompetencyInput) (*model.Competency, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_competencies", "create", "human_resource.competencies.create")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.CompetencyInput{}
	patch.CompetencyCode = input.CompetencyCode
	patch.CategoryCode = input.CategoryCode
	resp, err := a.competency.CreateCompetency(rpcCtx, &competency.CreateCompetencyRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "competency")
	}
	return toModelCompetency(resp.GetItem())
}

// UpdateCompetency memperbarui competency berdasarkan FieldMask.
func (a *Adapter) UpdateCompetency(ctx context.Context, rawID string, input model.UpdateCompetencyInput) (*model.Competency, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_competencies", "update", "human_resource.competencies.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "competency", "competency")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.CompetencyInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "competency_code", input.CompetencyCode, &patch.CompetencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "category_code", input.CategoryCode, &patch.CategoryCode, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("competency update is empty", nil)
	}
	resp, err := a.competency.UpdateCompetency(rpcCtx, &competency.UpdateCompetencyRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "competency")
	}
	return toModelCompetency(resp.GetItem())
}

// DeleteCompetency menghapus competency berdasarkan ID.
func (a *Adapter) DeleteCompetency(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_competencies", "human_resource.competencies.delete", "competency", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.competency.DeleteCompetency(c, req)
	})
}

func toModelJobCompetency(row *competency.JobCompetency) (*model.JobCompetency, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource job-competency response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource job-competency ID", nil)
	}
	out := &model.JobCompetency{ID: nodeID("job-competency", row.GetId()), RawID: row.GetId()}
	return out, nil
}

// JobCompetency membaca detail job-competency berdasarkan ID.
func (a *Adapter) JobCompetency(ctx context.Context, rawID string) (*model.JobCompetency, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_job_competencies", "human_resource.job_competencies.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job-competency", "job-competency")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.competency.GetJobCompetencyByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "job-competency")
	}
	return toModelJobCompetency(resp.GetItem())
}

// JobCompetencies mendaftar dan memfilter job-competency dengan paginasi.
func (a *Adapter) JobCompetencies(ctx context.Context, limit, offset int, f_job_id *string, f_competency_id *string) (*Page[model.JobCompetency], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_job_competencies", "human_resource.job_competencies.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid job-competency pagination", nil)
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_job_id != nil, f_competency_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *competency.JobCompetenciesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_job_id, "job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindJobCompetencyByJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_competency_id, "competency_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindJobCompetencyByCompetencyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.competency.GetAllJobCompetencies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "job-competency")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "job-competency", toModelJobCompetency)
}

// CreateJobCompetency membuat job-competency baru melalui HRM.
func (a *Adapter) CreateJobCompetency(ctx context.Context, input model.CreateJobCompetencyInput) (*model.JobCompetency, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_competencies", "create", "human_resource.job_competencies.create")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.JobCompetencyInput{}
	patch.JobId = input.JobID
	resp, err := a.competency.CreateJobCompetency(rpcCtx, &competency.CreateJobCompetencyRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-competency")
	}
	return toModelJobCompetency(resp.GetItem())
}

// UpdateJobCompetency memperbarui job-competency berdasarkan FieldMask.
func (a *Adapter) UpdateJobCompetency(ctx context.Context, rawID string, input model.UpdateJobCompetencyInput) (*model.JobCompetency, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_competencies", "update", "human_resource.job_competencies.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job-competency", "job-competency")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.JobCompetencyInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "job_id", input.JobID, &patch.JobId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("job-competency update is empty", nil)
	}
	resp, err := a.competency.UpdateJobCompetency(rpcCtx, &competency.UpdateJobCompetencyRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-competency")
	}
	return toModelJobCompetency(resp.GetItem())
}

// DeleteJobCompetency menghapus job-competency berdasarkan ID.
func (a *Adapter) DeleteJobCompetency(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_job_competencies", "human_resource.job_competencies.delete", "job-competency", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.competency.DeleteJobCompetency(c, req)
	})
}

func toModelWorkerCompetencyAssessment(row *competency.WorkerCompetencyAssessment) (*model.WorkerCompetencyAssessment, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-competency-assessment response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-competency-assessment ID", nil)
	}
	out := &model.WorkerCompetencyAssessment{ID: nodeID("worker-competency-assessment", row.GetId()), RawID: row.GetId()}
	out.AssessedAt = pbTimePtr(row.GetAssessedAt())
	out.EvidenceDocumentVersionReferenceID = row.EvidenceDocumentVersionReferenceId
	return out, nil
}

// WorkerCompetencyAssessment membaca detail worker-competency-assessment berdasarkan ID.
func (a *Adapter) WorkerCompetencyAssessment(ctx context.Context, rawID string) (*model.WorkerCompetencyAssessment, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_competency_assessments", "human_resource.worker_competency_assessments.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-competency-assessment", "worker-competency-assessment")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.competency.GetWorkerCompetencyAssessmentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "worker-competency-assessment")
	}
	return toModelWorkerCompetencyAssessment(resp.GetItem())
}

// WorkerCompetencyAssessments mendaftar dan memfilter worker-competency-assessment dengan paginasi.
func (a *Adapter) WorkerCompetencyAssessments(ctx context.Context, limit, offset int, f_worker_id *string, f_competency_id *string, f_assessed_by_principal_id *string, f_evidence_document_version_reference_id *string, f_idempotency_key *string) (*Page[model.WorkerCompetencyAssessment], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_competency_assessments", "human_resource.worker_competency_assessments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker-competency-assessment pagination", nil)
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_competency_id != nil, f_assessed_by_principal_id != nil, f_evidence_document_version_reference_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *competency.WorkerCompetencyAssessmentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerCompetencyAssessmentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_competency_id, "competency_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerCompetencyAssessmentByCompetencyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_assessed_by_principal_id, "assessed_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerCompetencyAssessmentByAssessedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerCompetencyAssessmentByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.competency.FindWorkerCompetencyAssessmentByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.competency.GetAllWorkerCompetencyAssessments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker-competency-assessment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker-competency-assessment", toModelWorkerCompetencyAssessment)
}

// CreateWorkerCompetencyAssessment membuat worker-competency-assessment baru melalui HRM.
func (a *Adapter) CreateWorkerCompetencyAssessment(ctx context.Context, input model.CreateWorkerCompetencyAssessmentInput) (*model.WorkerCompetencyAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_competency_assessments", "create", "human_resource.worker_competency_assessments.create")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.WorkerCompetencyAssessmentInput{}
	patch.WorkerId = input.WorkerID
	patch.AssessedAt = optionalTimestamp(input.AssessedAt)
	if b, bErr := optionalBase64Bytes(input.EvidenceDigest, "evidence_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EvidenceDigest = b
	}
	resp, err := a.competency.CreateWorkerCompetencyAssessment(rpcCtx, &competency.CreateWorkerCompetencyAssessmentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-competency-assessment")
	}
	return toModelWorkerCompetencyAssessment(resp.GetItem())
}

// UpdateWorkerCompetencyAssessment memperbarui worker-competency-assessment berdasarkan FieldMask.
func (a *Adapter) UpdateWorkerCompetencyAssessment(ctx context.Context, rawID string, input model.UpdateWorkerCompetencyAssessmentInput) (*model.WorkerCompetencyAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_competency_assessments", "update", "human_resource.worker_competency_assessments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-competency-assessment", "worker-competency-assessment")
	if err != nil {
		return nil, err
	}
	if a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &competency.WorkerCompetencyAssessmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "assessed_at", input.AssessedAt, &patch.AssessedAt, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "evidence_digest", input.EvidenceDigest, &patch.EvidenceDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("worker-competency-assessment update is empty", nil)
	}
	resp, err := a.competency.UpdateWorkerCompetencyAssessment(rpcCtx, &competency.UpdateWorkerCompetencyAssessmentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-competency-assessment")
	}
	return toModelWorkerCompetencyAssessment(resp.GetItem())
}

// DeleteWorkerCompetencyAssessment menghapus worker-competency-assessment berdasarkan ID.
func (a *Adapter) DeleteWorkerCompetencyAssessment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.competency == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_worker_competency_assessments", "human_resource.worker_competency_assessments.delete", "worker-competency-assessment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.competency.DeleteWorkerCompetencyAssessment(c, req)
	})
}
