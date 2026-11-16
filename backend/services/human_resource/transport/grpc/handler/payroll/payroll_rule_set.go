package payroll

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func payrollRuleSet(v *entity.PayrollRuleSet) *pb.PayrollRuleSet {
	return &pb.PayrollRuleSet{
		Id: v.ID, RuleCode: v.RuleCode, RuleVersion: v.RuleVersion, JurisdictionCode: v.JurisdictionCode,
		EffectiveFrom: v.EffectiveFrom.Format("2006-01-02"), EffectiveTo: dateString(v.EffectiveTo), ArtifactDocumentVersionReferenceId: v.ArtifactDocumentVersionReferenceID,
		ArtifactDigest: append([]byte(nil), v.ArtifactDigest...), LegalBasisDocumentVersionReferenceId: v.LegalBasisDocumentVersionReferenceID,
		ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamppb.New(v.ApprovedAt),
	}
}
func payrollRuleSets(page operation.Page[entity.PayrollRuleSet], err error) (*pb.PayrollRuleSetsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollRuleSet, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollRuleSet(&page.Items[i]))
	}
	return &pb.PayrollRuleSetsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollRuleSet(ctx context.Context, r *pb.CreatePayrollRuleSetRequest) (*pb.PayrollRuleSetResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.Timestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreatePayrollRuleSet(ctx, service.CreatePayrollRuleSetInput{
		RuleCode: in.GetRuleCode(), RuleVersion: in.GetRuleVersion(), JurisdictionCode: in.GetJurisdictionCode(), EffectiveFrom: from, EffectiveTo: to,
		ArtifactDocumentVersionReferenceID: in.GetArtifactDocumentVersionReferenceId(), ArtifactDigest: append([]byte(nil), in.ArtifactDigest...),
		LegalBasisDocumentVersionReferenceID: in.GetLegalBasisDocumentVersionReferenceId(), ApprovedByPrincipalID: in.GetApprovedByPrincipalId(), ApprovedAt: approved,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRuleSetResponse{Item: payrollRuleSet(v)}, nil
}
func (h *Handler) UpdatePayrollRuleSet(ctx context.Context, r *pb.UpdatePayrollRuleSetRequest) (*pb.PayrollRuleSetResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "rule_code", "rule_version", "jurisdiction_code", "effective_from", "effective_to", "artifact_document_version_reference_id", "artifact_digest", "legal_basis_document_version_reference_id", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	from, err := dateField(m["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(m["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(m["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdatePayrollRuleSet(ctx, r.Id, service.UpdatePayrollRuleSetInput{
		RuleCode: field(m["rule_code"], in.RuleCode), RuleVersion: field(m["rule_version"], in.RuleVersion), JurisdictionCode: field(m["jurisdiction_code"], in.JurisdictionCode),
		EffectiveFrom: from, EffectiveTo: to, ArtifactDocumentVersionReferenceID: field(m["artifact_document_version_reference_id"], in.ArtifactDocumentVersionReferenceId),
		ArtifactDigest: bytesField(m["artifact_digest"], in.ArtifactDigest), LegalBasisDocumentVersionReferenceID: field(m["legal_basis_document_version_reference_id"], in.LegalBasisDocumentVersionReferenceId),
		ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId), ApprovedAt: approved,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRuleSetResponse{Item: payrollRuleSet(v)}, nil
}
func (h *Handler) DeletePayrollRuleSet(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollRuleSet(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollRuleSetByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollRuleSetResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollRuleSetByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRuleSetResponse{Item: payrollRuleSet(v)}, nil
}
func (h *Handler) GetAllPayrollRuleSets(ctx context.Context, r *common.PageRequest) (*pb.PayrollRuleSetsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuleSets(h.service.GetAllPayrollRuleSets(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRuleSetByRuleCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRuleSetsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuleSets(h.service.FindPayrollRuleSetByRuleCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRuleSetByJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRuleSetsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuleSets(h.service.FindPayrollRuleSetByJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRuleSetByArtifactDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRuleSetsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuleSets(h.service.FindPayrollRuleSetByArtifactDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRuleSetByLegalBasisDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRuleSetsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuleSets(h.service.FindPayrollRuleSetByLegalBasisDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRuleSetByApprovedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRuleSetsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuleSets(h.service.FindPayrollRuleSetByApprovedByPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
