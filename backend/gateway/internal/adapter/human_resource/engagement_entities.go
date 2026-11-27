package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	engagement "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/engagement/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelEngagementSurvey(row *engagement.EngagementSurvey) (*model.EngagementSurvey, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource engagement-survey response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource engagement-survey ID", nil)
	}
	out := &model.EngagementSurvey{ID: nodeID("engagement-survey", row.GetId()), RawID: row.GetId()}
	out.SurveyCode = row.GetSurveyCode()
	out.SurveyVersion = int(row.GetSurveyVersion())
	out.LegalEntityReferenceID = row.GetLegalEntityReferenceId()
	out.TopicCode = row.GetTopicCode()
	out.OpensAt = pbTimePtr(row.GetOpensAt())
	out.ClosesAt = pbTimePtr(row.GetClosesAt())
	out.QuestionnaireDocumentVersionReferenceID = row.GetQuestionnaireDocumentVersionReferenceId()
	out.QuestionnaireDigest = pbBase64StringValue(row.GetQuestionnaireDigest())
	out.MinimumReportingPopulation = int(row.GetMinimumReportingPopulation())
	return out, nil
}

// EngagementSurvey membaca detail engagement-survey berdasarkan ID.
func (a *Adapter) EngagementSurvey(ctx context.Context, rawID string) (*model.EngagementSurvey, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_engagement_surveys", "human_resource.engagement_surveys.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "engagement-survey", "engagement-survey")
	if err != nil {
		return nil, err
	}
	if a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.engagement.GetEngagementSurveyByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "engagement-survey")
	}
	return toModelEngagementSurvey(resp.GetValue())
}

// EngagementSurveys mendaftar dan memfilter engagement-survey dengan paginasi.
func (a *Adapter) EngagementSurveys(ctx context.Context, limit, offset int, f_survey_code *string, f_legal_entity_reference_id *string, f_topic_code *string, f_questionnaire_document_version_reference_id *string) (*Page[model.EngagementSurvey], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_engagement_surveys", "human_resource.engagement_surveys.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid engagement-survey pagination", nil)
	}
	if a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_survey_code != nil, f_legal_entity_reference_id != nil, f_topic_code != nil, f_questionnaire_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *engagement.EngagementSurveysResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_survey_code, "survey_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyBySurveyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_legal_entity_reference_id, "legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_topic_code, "topic_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyByTopicCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_questionnaire_document_version_reference_id, "questionnaire_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyByQuestionnaireDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.engagement.GetAllEngagementSurveys(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "engagement-survey")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "engagement-survey", toModelEngagementSurvey)
}

// CreateEngagementSurvey membuat engagement-survey baru melalui HRM.
func (a *Adapter) CreateEngagementSurvey(ctx context.Context, input model.CreateEngagementSurveyInput) (*model.EngagementSurvey, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_engagement_surveys", "create", "human_resource.engagement_surveys.create")
	if err != nil {
		return nil, err
	}
	if a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &engagement.EngagementSurveyInput{}
	patch.SurveyCode = input.SurveyCode
	patch.SurveyVersion = optionalInt32Ptr(input.SurveyVersion)
	patch.LegalEntityReferenceId = input.LegalEntityReferenceID
	patch.TopicCode = input.TopicCode
	patch.OpensAt = optionalTimestamp(input.OpensAt)
	patch.ClosesAt = optionalTimestamp(input.ClosesAt)
	patch.QuestionnaireDocumentVersionReferenceId = input.QuestionnaireDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.QuestionnaireDigest, "questionnaire_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.QuestionnaireDigest = b
	}
	patch.MinimumReportingPopulation = optionalInt32Ptr(input.MinimumReportingPopulation)
	resp, err := a.engagement.CreateEngagementSurvey(rpcCtx, &engagement.CreateEngagementSurveyRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "engagement-survey")
	}
	return toModelEngagementSurvey(resp.GetValue())
}

// UpdateEngagementSurvey memperbarui engagement-survey berdasarkan FieldMask.
func (a *Adapter) UpdateEngagementSurvey(ctx context.Context, rawID string, input model.UpdateEngagementSurveyInput) (*model.EngagementSurvey, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_engagement_surveys", "update", "human_resource.engagement_surveys.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "engagement-survey", "engagement-survey")
	if err != nil {
		return nil, err
	}
	if a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &engagement.EngagementSurveyInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "survey_code", input.SurveyCode, &patch.SurveyCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "survey_version", input.SurveyVersion, &patch.SurveyVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_entity_reference_id", input.LegalEntityReferenceID, &patch.LegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "topic_code", input.TopicCode, &patch.TopicCode, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "opens_at", input.OpensAt, &patch.OpensAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "closes_at", input.ClosesAt, &patch.ClosesAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "questionnaire_document_version_reference_id", input.QuestionnaireDocumentVersionReferenceID, &patch.QuestionnaireDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "questionnaire_digest", input.QuestionnaireDigest, &patch.QuestionnaireDigest, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "minimum_reporting_population", input.MinimumReportingPopulation, &patch.MinimumReportingPopulation, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("engagement-survey update is empty", nil)
	}
	resp, err := a.engagement.UpdateEngagementSurvey(rpcCtx, &engagement.UpdateEngagementSurveyRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "engagement-survey")
	}
	return toModelEngagementSurvey(resp.GetValue())
}

// DeleteEngagementSurvey menghapus engagement-survey berdasarkan ID.
func (a *Adapter) DeleteEngagementSurvey(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_engagement_surveys", "human_resource.engagement_surveys.delete", "engagement-survey", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.engagement.DeleteEngagementSurvey(c, req)
	})
}

func toModelEngagementSurveyResult(row *engagement.EngagementSurveyResult) (*model.EngagementSurveyResult, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource engagement-survey-result response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource engagement-survey-result ID", nil)
	}
	out := &model.EngagementSurveyResult{ID: nodeID("engagement-survey-result", row.GetId()), RawID: row.GetId()}
	out.SurveyID = row.GetSurveyId()
	out.OrganizationUnitReferenceID = row.OrganizationUnitReferenceId
	out.QuestionCode = row.GetQuestionCode()
	out.ResultVersion = int(row.GetResultVersion())
	out.RespondentCount = int(row.GetRespondentCount())
	out.ScoreValue = row.GetScoreValue()
	out.ScaleMinimum = row.GetScaleMinimum()
	out.ScaleMaximum = row.GetScaleMaximum()
	out.SourceDigest = pbBase64StringValue(row.GetSourceDigest())
	out.CalculatedAt = pbTimePtr(row.GetCalculatedAt())
	return out, nil
}

// EngagementSurveyResult membaca detail engagement-survey-result berdasarkan ID.
func (a *Adapter) EngagementSurveyResult(ctx context.Context, rawID string) (*model.EngagementSurveyResult, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_engagement_survey_results", "human_resource.engagement_survey_results.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "engagement-survey-result", "engagement-survey-result")
	if err != nil {
		return nil, err
	}
	if a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.engagement.GetEngagementSurveyResultByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "engagement-survey-result")
	}
	return toModelEngagementSurveyResult(resp.GetValue())
}

// EngagementSurveyResults mendaftar dan memfilter engagement-survey-result dengan paginasi.
func (a *Adapter) EngagementSurveyResults(ctx context.Context, limit, offset int, f_survey_id *string, f_organization_unit_reference_id *string, f_question_code *string) (*Page[model.EngagementSurveyResult], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_engagement_survey_results", "human_resource.engagement_survey_results.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid engagement-survey-result pagination", nil)
	}
	if a.engagement == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_survey_id != nil, f_organization_unit_reference_id != nil, f_question_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *engagement.EngagementSurveyResultsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_survey_id, "survey_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyResultBySurveyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_organization_unit_reference_id, "organization_unit_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyResultByOrganizationUnitReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_question_code, "question_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.engagement.FindEngagementSurveyResultByQuestionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.engagement.GetAllEngagementSurveyResults(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "engagement-survey-result")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "engagement-survey-result", toModelEngagementSurveyResult)
}
