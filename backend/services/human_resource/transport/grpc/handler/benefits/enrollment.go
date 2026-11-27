package benefits

import (
	"context"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/benefits"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func enrollment(value *entity.BenefitEnrollment) *pb.BenefitEnrollment {
	return &pb.BenefitEnrollment{
		Id: value.ID, WorkerId: value.WorkerID, BenefitPlanId: value.BenefitPlanID,
		EnrollmentStatus: value.EnrollmentStatus,
		EffectiveFrom:    value.EffectiveFrom.Format("2006-01-02"), EffectiveTo: dateString(value.EffectiveTo),
		EnrollmentReference: value.EnrollmentReference, ProviderMemberReference: value.ProviderMemberReference,
		EmployeeContributionAmount: decimalString(value.EmployeeContributionAmount),
		EmployerContributionAmount: decimalString(value.EmployerContributionAmount),
		CurrencyCode:               value.CurrencyCode, ElectedAt: timestamp(value.ElectedAt), VerifiedAt: timestamp(value.VerifiedAt),
	}
}
func enrollmentResponse(value *entity.BenefitEnrollment, err error) (*pb.BenefitEnrollmentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.BenefitEnrollmentResponse{Item: enrollment(value)}, nil
}
func enrollments(page operation.Page[entity.BenefitEnrollment], err error) (*pb.BenefitEnrollmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.BenefitEnrollment, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, enrollment(&page.Items[index]))
	}
	return &pb.BenefitEnrollmentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateBenefitEnrollment(ctx context.Context, request *pb.CreateBenefitEnrollmentRequest) (*pb.BenefitEnrollmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	elected, err := response.OptionalTimestamp(in.ElectedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	verified, err := response.OptionalTimestamp(in.VerifiedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return enrollmentResponse(h.service.CreateBenefitEnrollment(ctx, service.CreateBenefitEnrollmentInput{
		WorkerID: in.GetWorkerId(), BenefitPlanID: in.GetBenefitPlanId(), EnrollmentStatus: in.EnrollmentStatus,
		EffectiveFrom: from, EffectiveTo: to, EnrollmentReference: in.EnrollmentReference,
		ProviderMemberReference:    in.ProviderMemberReference,
		EmployeeContributionAmount: decimalPointer(in.EmployeeContributionAmount),
		EmployerContributionAmount: decimalPointer(in.EmployerContributionAmount),
		CurrencyCode:               in.CurrencyCode, ElectedAt: elected, VerifiedAt: verified,
	}))
}
func (h *Handler) UpdateBenefitEnrollment(ctx context.Context, request *pb.UpdateBenefitEnrollmentRequest) (*pb.BenefitEnrollmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask,
		"worker_id", "benefit_plan_id", "enrollment_status", "effective_from", "effective_to",
		"enrollment_reference", "provider_member_reference", "employee_contribution_amount",
		"employer_contribution_amount", "currency_code", "elected_at", "verified_at",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	elected, err := timestampField(mask["elected_at"], in.ElectedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	verified, err := timestampField(mask["verified_at"], in.VerifiedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return enrollmentResponse(h.service.UpdateBenefitEnrollment(ctx, request.Id, service.UpdateBenefitEnrollmentInput{
		WorkerID:         field(mask["worker_id"], in.WorkerId),
		BenefitPlanID:    field(mask["benefit_plan_id"], in.BenefitPlanId),
		EnrollmentStatus: field(mask["enrollment_status"], in.EnrollmentStatus),
		EffectiveFrom:    from, EffectiveTo: to,
		EnrollmentReference:        field(mask["enrollment_reference"], in.EnrollmentReference),
		ProviderMemberReference:    field(mask["provider_member_reference"], in.ProviderMemberReference),
		EmployeeContributionAmount: field(mask["employee_contribution_amount"], decimalPointer(in.EmployeeContributionAmount)),
		EmployerContributionAmount: field(mask["employer_contribution_amount"], decimalPointer(in.EmployerContributionAmount)),
		CurrencyCode:               field(mask["currency_code"], in.CurrencyCode), ElectedAt: elected, VerifiedAt: verified,
	}))
}
func (h *Handler) DeleteBenefitEnrollment(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteBenefitEnrollment(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetBenefitEnrollmentByID(ctx context.Context, request *common.IDRequest) (*pb.BenefitEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollmentResponse(h.service.GetBenefitEnrollmentByID(ctx, request.Id))
}
func (h *Handler) GetAllBenefitEnrollments(ctx context.Context, request *common.PageRequest) (*pb.BenefitEnrollmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.GetAllBenefitEnrollments(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitEnrollmentByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindBenefitEnrollmentByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitEnrollmentByBenefitPlanID(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindBenefitEnrollmentByBenefitPlanID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitEnrollmentByEnrollmentStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitEnrollmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindBenefitEnrollmentByEnrollmentStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitEnrollmentByCurrencyCode(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitEnrollmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollments(h.service.FindBenefitEnrollmentByCurrencyCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}

func (h *Handler) ApproveBenefitEnrollment(ctx context.Context, request *common.IDRequest) (*pb.BenefitEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return enrollmentResponse(h.service.ApproveBenefitEnrollment(ctx, request.Id))
}

func (h *Handler) EndBenefitEnrollment(ctx context.Context, request *pb.EndBenefitEnrollmentRequest) (*pb.BenefitEnrollmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	effectiveTo, err := response.Date(request.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return enrollmentResponse(h.service.EndBenefitEnrollment(ctx, request.Id, effectiveTo))
}
