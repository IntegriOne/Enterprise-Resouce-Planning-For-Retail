// Package employeerelations menerjemahkan RPC bertipe ke service hubungan karyawan HRM.
package employeerelations

import (
	"context"
	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/employeerelations/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/employeerelations"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type server struct {
	pb.UnimplementedEmployeeRelationsServiceServer
	service *service.Service
}

// Register memasang service hubungan karyawan pada server pemilik.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterEmployeeRelationsServiceServer(registrar, &server{service: svc})
	return nil
}
func timestamp(v *time.Time) *timestamppb.Timestamp {
	if v == nil {
		return nil
	}
	return timestamppb.New(*v)
}
func timeField(set bool, v *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !set {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(v)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func bytesField(set bool, v []byte) operation.Field[[]byte] {
	if v == nil {
		return operation.Field[[]byte]{Set: set}
	}
	return operation.Field[[]byte]{Set: set, Value: &v}
}
func caseStatus(v *string) *entity.CaseStatus {
	if v == nil {
		return nil
	}
	status := entity.CaseStatus(*v)
	return &status
}
func relationCase(v *entity.EmployeeRelationsCase) *pb.EmployeeRelationsCase {
	if v == nil {
		return nil
	}
	return &pb.EmployeeRelationsCase{Id: v.ID, CaseNumber: v.CaseNumber, CaseType: v.CaseType, Status: string(v.Status),
		SubjectWorkerId: v.SubjectWorkerID, ReportingPartyReferenceId: v.ReportingPartyReferenceID,
		AssignedInvestigatorPrincipalId: v.AssignedInvestigatorPrincipalID, OpenedAt: timestamppb.New(v.OpenedAt),
		DueAt: timestamp(v.DueAt), ResolvedAt: timestamp(v.ResolvedAt), OutcomeCode: v.OutcomeCode,
		RestrictedSummaryReferenceId: v.RestrictedSummaryReferenceID, ConfidentialityClass: v.ConfidentialityClass,
		LegalHoldReferenceId: v.LegalHoldReferenceID, IdempotencyKey: v.IdempotencyKey}
}
func caseResult(v *entity.EmployeeRelationsCase, err error) (*pb.EmployeeRelationsCaseResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EmployeeRelationsCaseResponse{Value: relationCase(v)}, nil
}
func cases(p operation.Page[entity.EmployeeRelationsCase], err error) (*pb.EmployeeRelationsCasesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.EmployeeRelationsCase, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, relationCase(&p.Items[i]))
	}
	return &pb.EmployeeRelationsCasesResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func relationEvent(v *entity.EmployeeRelationsEvent) *pb.EmployeeRelationsEvent {
	if v == nil {
		return nil
	}
	return &pb.EmployeeRelationsEvent{Id: v.ID, CaseId: v.CaseID, EventType: v.EventType,
		OccurredAt: timestamppb.New(v.OccurredAt), ActorPrincipalId: v.ActorPrincipalID, EventDigest: v.EventDigest, EventReferenceId: v.EventReferenceID}
}
func eventResult(v *entity.EmployeeRelationsEvent, err error) (*pb.EmployeeRelationsEventResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EmployeeRelationsEventResponse{Value: relationEvent(v)}, nil
}
func events(p operation.Page[entity.EmployeeRelationsEvent], err error) (*pb.EmployeeRelationsEventsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.EmployeeRelationsEvent, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, relationEvent(&p.Items[i]))
	}
	return &pb.EmployeeRelationsEventsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func agreement(v *entity.CollectiveAgreement) *pb.CollectiveAgreement {
	if v == nil {
		return nil
	}
	return &pb.CollectiveAgreement{Id: v.ID, AgreementCode: v.AgreementCode,
		AgreementVersion: v.AgreementVersion, LegalEntityReferenceId: v.LegalEntityReferenceID,
		RepresentativePartyReferenceId: v.RepresentativePartyReferenceID, JurisdictionCode: v.JurisdictionCode,
		EffectiveFrom: timestamppb.New(v.EffectiveFrom), EffectiveTo: timestamp(v.EffectiveTo), SignedAt: timestamppb.New(v.SignedAt),
		DocumentVersionReferenceId: v.DocumentVersionReferenceID, ContentDigest: v.ContentDigest}
}
func agreementResult(v *entity.CollectiveAgreement, err error) (*pb.CollectiveAgreementResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CollectiveAgreementResponse{Value: agreement(v)}, nil
}
func agreements(p operation.Page[entity.CollectiveAgreement], err error) (*pb.CollectiveAgreementsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.CollectiveAgreement, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, agreement(&p.Items[i]))
	}
	return &pb.CollectiveAgreementsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateEmployeeRelationsCase(ctx context.Context, r *pb.CreateEmployeeRelationsCaseRequest) (*pb.EmployeeRelationsCaseResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	due, err := response.OptionalTimestamp(in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	resolved, err := response.OptionalTimestamp(in.ResolvedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return caseResult(s.service.CreateEmployeeRelationsCase(ctx, service.CreateEmployeeRelationsCaseInput{
		CaseNumber: in.GetCaseNumber(), CaseType: in.GetCaseType(), Status: caseStatus(in.Status), SubjectWorkerID: in.SubjectWorkerId,
		ReportingPartyReferenceID: in.ReportingPartyReferenceId, AssignedInvestigatorPrincipalID: in.AssignedInvestigatorPrincipalId,
		DueAt: due, ResolvedAt: resolved, OutcomeCode: in.OutcomeCode, RestrictedSummaryReferenceID: in.GetRestrictedSummaryReferenceId(),
		ConfidentialityClass: in.ConfidentialityClass, LegalHoldReferenceID: in.LegalHoldReferenceId, IdempotencyKey: in.GetIdempotencyKey()}))
}
func (s *server) UpdateEmployeeRelationsCase(ctx context.Context, r *pb.UpdateEmployeeRelationsCaseRequest) (*pb.EmployeeRelationsCaseResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "case_number", "case_type", "status", "subject_worker_id", "reporting_party_reference_id",
		"assigned_investigator_principal_id", "due_at", "resolved_at", "outcome_code", "restricted_summary_reference_id", "confidentiality_class", "legal_hold_reference_id", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	due, err := timeField(f["due_at"], in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	resolved, err := timeField(f["resolved_at"], in.ResolvedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return caseResult(s.service.UpdateEmployeeRelationsCase(ctx, r.Id, service.UpdateEmployeeRelationsCaseInput{
		CaseNumber: operation.Field[string]{Set: f["case_number"], Value: in.CaseNumber}, CaseType: operation.Field[string]{Set: f["case_type"], Value: in.CaseType},
		Status:                          operation.Field[entity.CaseStatus]{Set: f["status"], Value: caseStatus(in.Status)},
		SubjectWorkerID:                 operation.Field[string]{Set: f["subject_worker_id"], Value: in.SubjectWorkerId},
		ReportingPartyReferenceID:       operation.Field[string]{Set: f["reporting_party_reference_id"], Value: in.ReportingPartyReferenceId},
		AssignedInvestigatorPrincipalID: operation.Field[string]{Set: f["assigned_investigator_principal_id"], Value: in.AssignedInvestigatorPrincipalId},
		DueAt:                           due, ResolvedAt: resolved, OutcomeCode: operation.Field[string]{Set: f["outcome_code"], Value: in.OutcomeCode},
		RestrictedSummaryReferenceID: operation.Field[string]{Set: f["restricted_summary_reference_id"], Value: in.RestrictedSummaryReferenceId},
		ConfidentialityClass:         operation.Field[string]{Set: f["confidentiality_class"], Value: in.ConfidentialityClass},
		LegalHoldReferenceID:         operation.Field[string]{Set: f["legal_hold_reference_id"], Value: in.LegalHoldReferenceId},
		IdempotencyKey:               operation.Field[string]{Set: f["idempotency_key"], Value: in.IdempotencyKey}}))
}
func (s *server) DeleteEmployeeRelationsCase(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteEmployeeRelationsCase(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetEmployeeRelationsCaseByID(ctx context.Context, r *commonv1.IDRequest) (*pb.EmployeeRelationsCaseResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return caseResult(s.service.GetEmployeeRelationsCaseByID(ctx, r.Id))
}
func (s *server) GetAllEmployeeRelationsCases(ctx context.Context, r *commonv1.PageRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.GetAllEmployeeRelationsCases(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByCaseNumber(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByCaseNumber(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByCaseType(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByCaseType(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByStatus(ctx, entity.CaseStatus(r.GetValue()), int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseBySubjectWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseBySubjectWorkerID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByReportingPartyReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByReportingPartyReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByAssignedInvestigatorPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByAssignedInvestigatorPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByOutcomeCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByOutcomeCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByRestrictedSummaryReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByRestrictedSummaryReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByLegalHoldReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByLegalHoldReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsCaseByIdempotencyKey(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsCasesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cases(s.service.FindEmployeeRelationsCaseByIdempotencyKey(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) CreateEmployeeRelationsEvent(ctx context.Context, r *pb.CreateEmployeeRelationsEventRequest) (*pb.EmployeeRelationsEventResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	return eventResult(s.service.CreateEmployeeRelationsEvent(ctx, service.CreateEmployeeRelationsEventInput{
		CaseID: in.GetCaseId(), EventType: in.GetEventType(), EventDigest: in.EventDigest, EventReferenceID: in.EventReferenceId}))
}
func (s *server) UpdateEmployeeRelationsEvent(ctx context.Context, r *pb.UpdateEmployeeRelationsEventRequest) (*pb.EmployeeRelationsEventResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "case_id", "event_type", "event_digest", "event_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	return eventResult(s.service.UpdateEmployeeRelationsEvent(ctx, r.Id, service.UpdateEmployeeRelationsEventInput{
		CaseID: operation.Field[string]{Set: f["case_id"], Value: in.CaseId}, EventType: operation.Field[string]{Set: f["event_type"], Value: in.EventType},
		EventDigest: bytesField(f["event_digest"], in.EventDigest), EventReferenceID: operation.Field[string]{Set: f["event_reference_id"], Value: in.EventReferenceId}}))
}
func (s *server) DeleteEmployeeRelationsEvent(ctx context.Context, r *commonv1.IdentityIDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteEmployeeRelationsEvent(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetEmployeeRelationsEventByID(ctx context.Context, r *commonv1.IdentityIDRequest) (*pb.EmployeeRelationsEventResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return eventResult(s.service.GetEmployeeRelationsEventByID(ctx, r.Id))
}
func (s *server) GetAllEmployeeRelationsEvents(ctx context.Context, r *commonv1.PageRequest) (*pb.EmployeeRelationsEventsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.GetAllEmployeeRelationsEvents(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsEventByCaseID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsEventsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindEmployeeRelationsEventByCaseID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsEventByEventType(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsEventsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindEmployeeRelationsEventByEventType(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsEventByActorPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsEventsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindEmployeeRelationsEventByActorPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindEmployeeRelationsEventByEventReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.EmployeeRelationsEventsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindEmployeeRelationsEventByEventReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) CreateCollectiveAgreement(ctx context.Context, r *pb.CreateCollectiveAgreementRequest) (*pb.CollectiveAgreementResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	from, err := response.Timestamp(in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := response.OptionalTimestamp(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	signed, err := response.Timestamp(in.SignedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return agreementResult(s.service.CreateCollectiveAgreement(ctx, service.CreateCollectiveAgreementInput{
		AgreementCode: in.GetAgreementCode(), AgreementVersion: in.GetAgreementVersion(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		RepresentativePartyReferenceID: in.GetRepresentativePartyReferenceId(), JurisdictionCode: in.GetJurisdictionCode(),
		EffectiveFrom: from, EffectiveTo: to, SignedAt: signed, DocumentVersionReferenceID: in.GetDocumentVersionReferenceId(), ContentDigest: in.ContentDigest}))
}
func (s *server) UpdateCollectiveAgreement(ctx context.Context, r *pb.UpdateCollectiveAgreementRequest) (*pb.CollectiveAgreementResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "agreement_code", "agreement_version", "legal_entity_reference_id", "representative_party_reference_id",
		"jurisdiction_code", "effective_from", "effective_to", "signed_at", "document_version_reference_id", "content_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	from, err := timeField(f["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := timeField(f["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	signed, err := timeField(f["signed_at"], in.SignedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return agreementResult(s.service.UpdateCollectiveAgreement(ctx, r.Id, service.UpdateCollectiveAgreementInput{
		AgreementCode:                  operation.Field[string]{Set: f["agreement_code"], Value: in.AgreementCode},
		AgreementVersion:               operation.Field[int32]{Set: f["agreement_version"], Value: in.AgreementVersion},
		LegalEntityReferenceID:         operation.Field[string]{Set: f["legal_entity_reference_id"], Value: in.LegalEntityReferenceId},
		RepresentativePartyReferenceID: operation.Field[string]{Set: f["representative_party_reference_id"], Value: in.RepresentativePartyReferenceId},
		JurisdictionCode:               operation.Field[string]{Set: f["jurisdiction_code"], Value: in.JurisdictionCode}, EffectiveFrom: from, EffectiveTo: to, SignedAt: signed,
		DocumentVersionReferenceID: operation.Field[string]{Set: f["document_version_reference_id"], Value: in.DocumentVersionReferenceId}, ContentDigest: bytesField(f["content_digest"], in.ContentDigest)}))
}
func (s *server) DeleteCollectiveAgreement(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteCollectiveAgreement(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetCollectiveAgreementByID(ctx context.Context, r *commonv1.IDRequest) (*pb.CollectiveAgreementResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreementResult(s.service.GetCollectiveAgreementByID(ctx, r.Id))
}
func (s *server) GetAllCollectiveAgreements(ctx context.Context, r *commonv1.PageRequest) (*pb.CollectiveAgreementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreements(s.service.GetAllCollectiveAgreements(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindCollectiveAgreementByAgreementCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CollectiveAgreementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreements(s.service.FindCollectiveAgreementByAgreementCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCollectiveAgreementByLegalEntityReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CollectiveAgreementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreements(s.service.FindCollectiveAgreementByLegalEntityReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCollectiveAgreementByRepresentativePartyReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CollectiveAgreementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreements(s.service.FindCollectiveAgreementByRepresentativePartyReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCollectiveAgreementByJurisdictionCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CollectiveAgreementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreements(s.service.FindCollectiveAgreementByJurisdictionCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCollectiveAgreementByDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CollectiveAgreementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return agreements(s.service.FindCollectiveAgreementByDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
