package human_resource

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	learning "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelLearningCourse(row *learning.LearningCourse) (*model.LearningCourse, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource learning-course response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource learning-course ID", nil)
	}
	out := &model.LearningCourse{ID: nodeID("learning-course", row.GetId()), RawID: row.GetId()}
	out.CourseCode = row.GetCourseCode()
	out.CourseVersion = row.GetCourseVersion()
	out.CourseName = row.GetCourseName()
	out.CourseType = row.GetCourseType()
	out.ProviderPartyReferenceID = row.ProviderPartyReferenceId
	out.DeliveryMode = row.GetDeliveryMode()
	out.DurationHours = row.DurationHours
	out.CredentialAwarded = row.GetCredentialAwarded()
	out.ExpiryMonths = pbIntPtr(row.ExpiryMonths)
	out.ValidFrom = row.ValidFrom
	out.ValidTo = row.ValidTo
	out.IsActive = row.GetIsActive()
	return out, nil
}

// LearningCourse membaca detail learning-course berdasarkan ID.
func (a *Adapter) LearningCourse(ctx context.Context, rawID string) (*model.LearningCourse, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_learning_courses", "human_resource.learning_courses.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-course", "learning-course")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.learning.GetLearningCourseByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "learning-course")
	}
	return toModelLearningCourse(resp.GetItem())
}

// LearningCourses mendaftar dan memfilter learning-course dengan paginasi.
func (a *Adapter) LearningCourses(ctx context.Context, limit, offset int, f_course_code *string, f_course_type *string, f_provider_party_reference_id *string) (*Page[model.LearningCourse], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_learning_courses", "human_resource.learning_courses.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid learning-course pagination", nil)
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_course_code != nil, f_course_type != nil, f_provider_party_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *learning.LearningCoursesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_course_code, "course_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningCourseByCourseCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_course_type, "course_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningCourseByCourseType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_provider_party_reference_id, "provider_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningCourseByProviderPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.learning.GetAllLearningCourses(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "learning-course")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "learning-course", toModelLearningCourse)
}

// CreateLearningCourse membuat learning-course baru melalui HRM.
func (a *Adapter) CreateLearningCourse(ctx context.Context, input model.CreateLearningCourseInput) (*model.LearningCourse, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_courses", "create", "human_resource.learning_courses.create")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &learning.LearningCourseInput{}
	patch.CourseCode = input.CourseCode
	patch.CourseVersion = input.CourseVersion
	patch.CourseName = input.CourseName
	patch.CourseType = input.CourseType
	patch.ProviderPartyReferenceId = input.ProviderPartyReferenceID
	patch.DeliveryMode = input.DeliveryMode
	patch.DurationHours = input.DurationHours
	patch.CredentialAwarded = input.CredentialAwarded
	patch.ExpiryMonths = optionalInt32Ptr(input.ExpiryMonths)
	patch.ValidFrom = input.ValidFrom
	patch.ValidTo = input.ValidTo
	patch.IsActive = input.IsActive
	resp, err := a.learning.CreateLearningCourse(rpcCtx, &learning.CreateLearningCourseRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-course")
	}
	return toModelLearningCourse(resp.GetItem())
}

// UpdateLearningCourse memperbarui learning-course berdasarkan FieldMask.
func (a *Adapter) UpdateLearningCourse(ctx context.Context, rawID string, input model.UpdateLearningCourseInput) (*model.LearningCourse, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_courses", "update", "human_resource.learning_courses.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-course", "learning-course")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &learning.LearningCourseInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "course_code", input.CourseCode, &patch.CourseCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "course_version", input.CourseVersion, &patch.CourseVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "course_name", input.CourseName, &patch.CourseName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "course_type", input.CourseType, &patch.CourseType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "provider_party_reference_id", input.ProviderPartyReferenceID, &patch.ProviderPartyReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "delivery_mode", input.DeliveryMode, &patch.DeliveryMode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "duration_hours", input.DurationHours, &patch.DurationHours, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "credential_awarded", input.CredentialAwarded, &patch.CredentialAwarded, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "expiry_months", input.ExpiryMonths, &patch.ExpiryMonths, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("learning-course update is empty", nil)
	}
	resp, err := a.learning.UpdateLearningCourse(rpcCtx, &learning.UpdateLearningCourseRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-course")
	}
	return toModelLearningCourse(resp.GetItem())
}

// DeleteLearningCourse menghapus learning-course berdasarkan ID.
func (a *Adapter) DeleteLearningCourse(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_learning_courses", "human_resource.learning_courses.delete", "learning-course", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.learning.DeleteLearningCourse(c, req)
	})
}

func toModelLearningEnrollment(row *learning.LearningEnrollment) (*model.LearningEnrollment, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource learning-enrollment response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource learning-enrollment ID", nil)
	}
	out := &model.LearningEnrollment{ID: nodeID("learning-enrollment", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.CourseID = row.GetCourseId()
	out.EnrollmentStatus = row.GetEnrollmentStatus()
	out.AssignedByPrincipalID = row.AssignedByPrincipalId
	out.AssignedAt = pbTimePtr(row.GetAssignedAt())
	out.DueAt = pbTimePtr(row.GetDueAt())
	out.StartedAt = pbTimePtr(row.GetStartedAt())
	out.CompletedAt = pbTimePtr(row.GetCompletedAt())
	out.Score = row.Score
	out.CredentialReferenceID = row.CredentialReferenceId
	out.FundingReferenceID = row.FundingReferenceId
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// LearningEnrollment membaca detail learning-enrollment berdasarkan ID.
func (a *Adapter) LearningEnrollment(ctx context.Context, rawID string) (*model.LearningEnrollment, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_learning_enrollments", "human_resource.learning_enrollments.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-enrollment", "learning-enrollment")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.learning.GetLearningEnrollmentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "learning-enrollment")
	}
	return toModelLearningEnrollment(resp.GetItem())
}

// LearningEnrollments mendaftar dan memfilter learning-enrollment dengan paginasi.
func (a *Adapter) LearningEnrollments(ctx context.Context, limit, offset int, f_worker_id *string, f_course_id *string, f_enrollment_status *string, f_assigned_by_principal_id *string, f_credential_reference_id *string, f_funding_reference_id *string, f_idempotency_key *string) (*Page[model.LearningEnrollment], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_learning_enrollments", "human_resource.learning_enrollments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid learning-enrollment pagination", nil)
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_course_id != nil, f_enrollment_status != nil, f_assigned_by_principal_id != nil, f_credential_reference_id != nil, f_funding_reference_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *learning.LearningEnrollmentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_course_id, "course_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByCourseID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_enrollment_status, "enrollment_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByEnrollmentStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_assigned_by_principal_id, "assigned_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByAssignedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_credential_reference_id, "credential_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByCredentialReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_funding_reference_id, "funding_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByFundingReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningEnrollmentByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.learning.GetAllLearningEnrollments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "learning-enrollment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "learning-enrollment", toModelLearningEnrollment)
}

// CreateLearningEnrollment membuat learning-enrollment baru melalui HRM.
func (a *Adapter) CreateLearningEnrollment(ctx context.Context, input model.CreateLearningEnrollmentInput) (*model.LearningEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_enrollments", "create", "human_resource.learning_enrollments.create")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &learning.LearningEnrollmentInput{}
	patch.WorkerId = input.WorkerID
	patch.CourseId = input.CourseID
	patch.EnrollmentStatus = input.EnrollmentStatus
	patch.AssignedByPrincipalId = input.AssignedByPrincipalID
	patch.DueAt = optionalTimestamp(input.DueAt)
	patch.StartedAt = optionalTimestamp(input.StartedAt)
	patch.CompletedAt = optionalTimestamp(input.CompletedAt)
	patch.Score = input.Score
	patch.CredentialReferenceId = input.CredentialReferenceID
	patch.FundingReferenceId = input.FundingReferenceID
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.learning.CreateLearningEnrollment(rpcCtx, &learning.CreateLearningEnrollmentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-enrollment")
	}
	return toModelLearningEnrollment(resp.GetItem())
}

// UpdateLearningEnrollment memperbarui learning-enrollment berdasarkan FieldMask.
func (a *Adapter) UpdateLearningEnrollment(ctx context.Context, rawID string, input model.UpdateLearningEnrollmentInput) (*model.LearningEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_enrollments", "update", "human_resource.learning_enrollments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "learning-enrollment", "learning-enrollment")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &learning.LearningEnrollmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "course_id", input.CourseID, &patch.CourseId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "enrollment_status", input.EnrollmentStatus, &patch.EnrollmentStatus, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assigned_by_principal_id", input.AssignedByPrincipalID, &patch.AssignedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_at", input.DueAt, &patch.DueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "started_at", input.StartedAt, &patch.StartedAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "completed_at", input.CompletedAt, &patch.CompletedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "score", input.Score, &patch.Score, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "credential_reference_id", input.CredentialReferenceID, &patch.CredentialReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "funding_reference_id", input.FundingReferenceID, &patch.FundingReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("learning-enrollment update is empty", nil)
	}
	resp, err := a.learning.UpdateLearningEnrollment(rpcCtx, &learning.UpdateLearningEnrollmentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-enrollment")
	}
	return toModelLearningEnrollment(resp.GetItem())
}

// DeleteLearningEnrollment menghapus learning-enrollment berdasarkan ID.
func (a *Adapter) DeleteLearningEnrollment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_learning_enrollments", "human_resource.learning_enrollments.delete", "learning-enrollment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.learning.DeleteLearningEnrollment(c, req)
	})
}

func toModelLearningCourseCompetency(row *learning.LearningCourseCompetency) (*model.LearningCourseCompetency, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource learning-course-competency response", nil)
	}
	if row.GetCourseId() == "" || row.GetCompetencyId() == "" {
		return nil, middleware.ErrUnavailable("invalid human resource learning-course-competency key", nil)
	}
	rawID := row.GetCourseId() + ":" + row.GetCompetencyId()
	out := &model.LearningCourseCompetency{ID: nodeID("learning-course-competency", rawID), RawID: rawID}
	out.CourseID = row.GetCourseId()
	out.CompetencyID = row.GetCompetencyId()
	out.TargetLevel = int(row.GetTargetLevel())
	return out, nil
}

// LearningCourseCompetency membaca detail learning-course-competency berdasarkan ID.
func (a *Adapter) LearningCourseCompetency(ctx context.Context, rawID string) (*model.LearningCourseCompetency, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_learning_course_competencies", "human_resource.learning_course_competencies.read")
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(rawNodeID(strings.TrimSpace(rawID), "learning-course-competency"), ":", 2)
	if len(parts) != 2 {
		return nil, middleware.ErrInvalidArgument("invalid learning-course-competency composite id", nil)
	}
	p1, err := normalizeRequiredUUID(parts[0], "course_id", "learning-course-competency")
	if err != nil {
		return nil, err
	}
	p2, err := normalizeRequiredUUID(parts[1], "competency_id", "learning-course-competency")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.learning.GetLearningCourseCompetencyByID(rpcCtx, &learning.LearningCourseCompetencyKey{CourseId: p1, CompetencyId: p2})
	if err != nil {
		return nil, mapReadError(err, "learning-course-competency")
	}
	return toModelLearningCourseCompetency(resp.GetItem())
}

// LearningCourseCompetencies mendaftar dan memfilter learning-course-competency dengan paginasi.
func (a *Adapter) LearningCourseCompetencies(ctx context.Context, limit, offset int, f_course_id *string, f_competency_id *string) (*Page[model.LearningCourseCompetency], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_learning_course_competencies", "human_resource.learning_course_competencies.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid learning-course-competency pagination", nil)
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_course_id != nil, f_competency_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *learning.LearningCourseCompetenciesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_course_id, "course_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningCourseCompetencyByCourseID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_competency_id, "competency_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.learning.FindLearningCourseCompetencyByCompetencyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.learning.GetAllLearningCourseCompetencies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "learning-course-competency")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "learning-course-competency", toModelLearningCourseCompetency)
}

// CreateLearningCourseCompetency membuat learning-course-competency baru melalui HRM.
func (a *Adapter) CreateLearningCourseCompetency(ctx context.Context, input model.CreateLearningCourseCompetencyInput) (*model.LearningCourseCompetency, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_course_competencies", "create", "human_resource.learning_course_competencies.create")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &learning.LearningCourseCompetencyInput{}
	patch.CourseId = input.CourseID
	patch.CompetencyId = input.CompetencyID
	patch.TargetLevel = optionalInt32Ptr(input.TargetLevel)
	resp, err := a.learning.CreateLearningCourseCompetency(rpcCtx, &learning.CreateLearningCourseCompetencyRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-course-competency")
	}
	return toModelLearningCourseCompetency(resp.GetItem())
}

// UpdateLearningCourseCompetency memperbarui learning-course-competency berdasarkan FieldMask.
func (a *Adapter) UpdateLearningCourseCompetency(ctx context.Context, rawID string, input model.UpdateLearningCourseCompetencyInput) (*model.LearningCourseCompetency, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_course_competencies", "update", "human_resource.learning_course_competencies.update")
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(rawNodeID(strings.TrimSpace(rawID), "learning-course-competency"), ":", 2)
	if len(parts) != 2 {
		return nil, middleware.ErrInvalidArgument("invalid learning-course-competency composite id", nil)
	}
	p1, err := normalizeRequiredUUID(parts[0], "course_id", "learning-course-competency")
	if err != nil {
		return nil, err
	}
	p2, err := normalizeRequiredUUID(parts[1], "competency_id", "learning-course-competency")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	mask := &fieldmaskpb.FieldMask{}
	var targetLevel *int32
	if err := applyInt32Mask(mask, "target_level", input.TargetLevel, &targetLevel, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("learning-course-competency update is empty", nil)
	}
	resp, err := a.learning.UpdateLearningCourseCompetency(rpcCtx, &learning.UpdateLearningCourseCompetencyRequest{Key: &learning.LearningCourseCompetencyKey{CourseId: p1, CompetencyId: p2}, TargetLevel: targetLevel, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "learning-course-competency")
	}
	return toModelLearningCourseCompetency(resp.GetItem())
}

// DeleteLearningCourseCompetency menghapus learning-course-competency berdasarkan ID.
func (a *Adapter) DeleteLearningCourseCompetency(ctx context.Context, rawID string) (*bool, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_learning_course_competencies", "delete", "human_resource.learning_course_competencies.delete")
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(rawNodeID(strings.TrimSpace(rawID), "learning-course-competency"), ":", 2)
	if len(parts) != 2 {
		return nil, middleware.ErrInvalidArgument("invalid learning-course-competency composite id", nil)
	}
	p1, err := normalizeRequiredUUID(parts[0], "course_id", "learning-course-competency")
	if err != nil {
		return nil, err
	}
	p2, err := normalizeRequiredUUID(parts[1], "competency_id", "learning-course-competency")
	if err != nil {
		return nil, err
	}
	if a.learning == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	if _, err := a.learning.DeleteLearningCourseCompetency(rpcCtx, &learning.LearningCourseCompetencyKey{CourseId: p1, CompetencyId: p2}); err != nil {
		return nil, mapHRMWriteError(err, "learning-course-competency")
	}
	ok := true
	return &ok, nil
}
