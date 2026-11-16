// Package engagement menyediakan pemetaan transport survei dan hasil agregat HRM.
package engagement

import (
	"context"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/engagement/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/engagement"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedEngagementServiceServer
	service *service.Service
}

// Register memasang handler survei yang memakai policy application service.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterEngagementServiceServer(registrar, &server{service: svc})
	return nil
}
func timeField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(value)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return operation.Field[[]byte]{Set: selected}
	}
	return operation.Field[[]byte]{Set: selected, Value: &value}
}
func decimalField(selected bool, value *string) operation.Field[entity.Decimal] {
	if value == nil {
		return operation.Field[entity.Decimal]{Set: selected}
	}
	parsed := entity.Decimal(*value)
	return operation.Field[entity.Decimal]{Set: selected, Value: &parsed}
}
func survey(value *entity.EngagementSurvey) *pb.EngagementSurvey {
	if value == nil {
		return nil
	}
	return &pb.EngagementSurvey{
		Id: value.ID, SurveyCode: value.SurveyCode, SurveyVersion: value.SurveyVersion, LegalEntityReferenceId: value.LegalEntityReferenceID,
		TopicCode: value.TopicCode, OpensAt: timestamppb.New(value.OpensAt), ClosesAt: timestamppb.New(value.ClosesAt),
		QuestionnaireDocumentVersionReferenceId: value.QuestionnaireDocumentVersionReferenceID, QuestionnaireDigest: value.QuestionnaireDigest,
		MinimumReportingPopulation: value.MinimumReportingPopulation,
	}
}
func result(value *entity.EngagementSurveyResult) *pb.EngagementSurveyResult {
	if value == nil {
		return nil
	}
	return &pb.EngagementSurveyResult{
		Id: value.ID, SurveyId: value.SurveyID, OrganizationUnitReferenceId: value.OrganizationUnitReferenceID, QuestionCode: value.QuestionCode,
		ResultVersion: value.ResultVersion, RespondentCount: value.RespondentCount, ScoreValue: string(value.ScoreValue),
		ScaleMinimum: string(value.ScaleMinimum), ScaleMaximum: string(value.ScaleMaximum), SourceDigest: value.SourceDigest, CalculatedAt: timestamppb.New(value.CalculatedAt),
	}
}
func surveys(page operation.Page[entity.EngagementSurvey]) *pb.EngagementSurveysResponse {
	items := make([]*pb.EngagementSurvey, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, survey(&page.Items[index]))
	}
	return &pb.EngagementSurveysResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}
func results(page operation.Page[entity.EngagementSurveyResult]) *pb.EngagementSurveyResultsResponse {
	items := make([]*pb.EngagementSurveyResult, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, result(&page.Items[index]))
	}
	return &pb.EngagementSurveyResultsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}

func (s *server) CreateEngagementSurvey(ctx context.Context, request *pb.CreateEngagementSurveyRequest) (*pb.EngagementSurveyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	opens, err := response.Timestamp(in.OpensAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closes, err := response.Timestamp(in.ClosesAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.CreateEngagementSurvey(ctx, service.CreateEngagementSurveyInput{
		SurveyCode: in.GetSurveyCode(), SurveyVersion: in.GetSurveyVersion(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		TopicCode: in.GetTopicCode(), OpensAt: opens, ClosesAt: closes, QuestionnaireDocumentVersionReferenceID: in.GetQuestionnaireDocumentVersionReferenceId(),
		QuestionnaireDigest: in.QuestionnaireDigest, MinimumReportingPopulation: in.GetMinimumReportingPopulation(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EngagementSurveyResponse{Value: survey(value)}, nil
}
func (s *server) UpdateEngagementSurvey(ctx context.Context, request *pb.UpdateEngagementSurveyRequest) (*pb.EngagementSurveyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"survey_code", "survey_version", "legal_entity_reference_id", "topic_code", "opens_at", "closes_at",
		"questionnaire_document_version_reference_id", "questionnaire_digest", "minimum_reporting_population",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	opens, err := timeField(selected["opens_at"], in.OpensAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closes, err := timeField(selected["closes_at"], in.ClosesAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.UpdateEngagementSurvey(ctx, request.Id, service.UpdateEngagementSurveyInput{
		SurveyCode: operation.Field[string]{Set: selected["survey_code"], Value: in.SurveyCode}, SurveyVersion: operation.Field[int32]{Set: selected["survey_version"], Value: in.SurveyVersion},
		LegalEntityReferenceID: operation.Field[string]{Set: selected["legal_entity_reference_id"], Value: in.LegalEntityReferenceId}, TopicCode: operation.Field[string]{Set: selected["topic_code"], Value: in.TopicCode},
		OpensAt: opens, ClosesAt: closes, QuestionnaireDocumentVersionReferenceID: operation.Field[string]{Set: selected["questionnaire_document_version_reference_id"], Value: in.QuestionnaireDocumentVersionReferenceId},
		QuestionnaireDigest: bytesField(selected["questionnaire_digest"], in.QuestionnaireDigest), MinimumReportingPopulation: operation.Field[int32]{Set: selected["minimum_reporting_population"], Value: in.MinimumReportingPopulation},
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EngagementSurveyResponse{Value: survey(value)}, nil
}
func (s *server) DeleteEngagementSurvey(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteEngagementSurvey(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetEngagementSurveyByID(ctx context.Context, request *commonv1.IDRequest) (*pb.EngagementSurveyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetEngagementSurveyByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EngagementSurveyResponse{Value: survey(value)}, nil
}
func (s *server) GetAllEngagementSurveys(ctx context.Context, request *commonv1.PageRequest) (*pb.EngagementSurveysResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllEngagementSurveys(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return surveys(page), nil
}
func (s *server) FindEngagementSurveyBySurveyCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveysResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyBySurveyCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return surveys(page), nil
}
func (s *server) FindEngagementSurveyByLegalEntityReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveysResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyByLegalEntityReferenceID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return surveys(page), nil
}
func (s *server) FindEngagementSurveyByTopicCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveysResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyByTopicCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return surveys(page), nil
}
func (s *server) FindEngagementSurveyByQuestionnaireDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveysResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyByQuestionnaireDocumentVersionReferenceID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return surveys(page), nil
}

func (s *server) CreateEngagementSurveyResult(ctx context.Context, request *pb.CreateEngagementSurveyResultRequest) (*pb.EngagementSurveyResultResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	value, err := s.service.CreateEngagementSurveyResult(ctx, service.CreateEngagementSurveyResultInput{
		SurveyID: in.GetSurveyId(), OrganizationUnitReferenceID: in.OrganizationUnitReferenceId, QuestionCode: in.GetQuestionCode(),
		ResultVersion: in.GetResultVersion(), RespondentCount: in.GetRespondentCount(), ScoreValue: entity.Decimal(in.GetScoreValue()),
		ScaleMinimum: entity.Decimal(in.GetScaleMinimum()), ScaleMaximum: entity.Decimal(in.GetScaleMaximum()), SourceDigest: in.SourceDigest,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EngagementSurveyResultResponse{Value: result(value)}, nil
}
func (s *server) UpdateEngagementSurveyResult(ctx context.Context, request *pb.UpdateEngagementSurveyResultRequest) (*pb.EngagementSurveyResultResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"survey_id", "organization_unit_reference_id", "question_code", "result_version", "respondent_count", "score_value", "scale_minimum", "scale_maximum", "source_digest",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	value, err := s.service.UpdateEngagementSurveyResult(ctx, request.Id, service.UpdateEngagementSurveyResultInput{
		SurveyID: operation.Field[string]{Set: selected["survey_id"], Value: in.SurveyId}, OrganizationUnitReferenceID: operation.Field[string]{Set: selected["organization_unit_reference_id"], Value: in.OrganizationUnitReferenceId},
		QuestionCode: operation.Field[string]{Set: selected["question_code"], Value: in.QuestionCode}, ResultVersion: operation.Field[int32]{Set: selected["result_version"], Value: in.ResultVersion},
		RespondentCount: operation.Field[int32]{Set: selected["respondent_count"], Value: in.RespondentCount}, ScoreValue: decimalField(selected["score_value"], in.ScoreValue),
		ScaleMinimum: decimalField(selected["scale_minimum"], in.ScaleMinimum), ScaleMaximum: decimalField(selected["scale_maximum"], in.ScaleMaximum), SourceDigest: bytesField(selected["source_digest"], in.SourceDigest),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EngagementSurveyResultResponse{Value: result(value)}, nil
}
func (s *server) DeleteEngagementSurveyResult(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteEngagementSurveyResult(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetEngagementSurveyResultByID(ctx context.Context, request *commonv1.IDRequest) (*pb.EngagementSurveyResultResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetEngagementSurveyResultByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EngagementSurveyResultResponse{Value: result(value)}, nil
}
func (s *server) GetAllEngagementSurveyResults(ctx context.Context, request *commonv1.PageRequest) (*pb.EngagementSurveyResultsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllEngagementSurveyResults(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return results(page), nil
}
func (s *server) FindEngagementSurveyResultBySurveyID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveyResultsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyResultBySurveyID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return results(page), nil
}
func (s *server) FindEngagementSurveyResultByOrganizationUnitReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveyResultsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyResultByOrganizationUnitReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return results(page), nil
}
func (s *server) FindEngagementSurveyResultByQuestionCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.EngagementSurveyResultsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindEngagementSurveyResultByQuestionCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return results(page), nil
}
