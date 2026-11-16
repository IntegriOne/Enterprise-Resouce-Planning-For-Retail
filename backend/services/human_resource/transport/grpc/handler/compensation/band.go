package compensation

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/compensation/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/compensation"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func band(value *entity.CompensationBand) *pb.CompensationBand {
	return &pb.CompensationBand{Id: value.ID, BandCode: value.BandCode, BandVersion: value.BandVersion, JobLevelCode: value.JobLevelCode,
		JurisdictionCode: value.JurisdictionCode, CurrencyCode: value.CurrencyCode, PayFrequency: value.PayFrequency, MinimumAmount: string(value.MinimumAmount),
		MidpointAmount: string(value.MidpointAmount), MaximumAmount: string(value.MaximumAmount), ValidFrom: value.ValidFrom.Format("2006-01-02"), ValidTo: dateString(value.ValidTo), ApprovalReferenceId: value.ApprovalReferenceID}
}
func bandResponse(value *entity.CompensationBand, err error) (*pb.CompensationBandResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CompensationBandResponse{Item: band(value)}, nil
}
func bands(page operation.Page[entity.CompensationBand], err error) (*pb.CompensationBandsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.CompensationBand, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, band(&page.Items[index]))
	}
	return &pb.CompensationBandsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateCompensationBand(ctx context.Context, request *pb.CreateCompensationBandRequest) (*pb.CompensationBandResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetValidFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return bandResponse(h.service.CreateCompensationBand(ctx, service.CreateCompensationBandInput{BandCode: in.GetBandCode(), BandVersion: in.GetBandVersion(), JobLevelCode: in.GetJobLevelCode(),
		JurisdictionCode: in.GetJurisdictionCode(), CurrencyCode: in.GetCurrencyCode(), PayFrequency: in.GetPayFrequency(), MinimumAmount: entity.Decimal(in.GetMinimumAmount()),
		MidpointAmount: entity.Decimal(in.GetMidpointAmount()), MaximumAmount: entity.Decimal(in.GetMaximumAmount()), ValidFrom: from, ValidTo: to, ApprovalReferenceID: in.ApprovalReferenceId}))
}
func (h *Handler) UpdateCompensationBand(ctx context.Context, request *pb.UpdateCompensationBandRequest) (*pb.CompensationBandResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "band_code", "band_version", "job_level_code", "jurisdiction_code", "currency_code", "pay_frequency", "minimum_amount", "midpoint_amount", "maximum_amount", "valid_from", "valid_to", "approval_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["valid_from"], in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["valid_to"], in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return bandResponse(h.service.UpdateCompensationBand(ctx, request.Id, service.UpdateCompensationBandInput{BandCode: field(mask["band_code"], in.BandCode), BandVersion: field(mask["band_version"], in.BandVersion),
		JobLevelCode: field(mask["job_level_code"], in.JobLevelCode), JurisdictionCode: field(mask["jurisdiction_code"], in.JurisdictionCode), CurrencyCode: field(mask["currency_code"], in.CurrencyCode),
		PayFrequency: field(mask["pay_frequency"], in.PayFrequency), MinimumAmount: field(mask["minimum_amount"], decimalPointer(in.MinimumAmount)), MidpointAmount: field(mask["midpoint_amount"], decimalPointer(in.MidpointAmount)),
		MaximumAmount: field(mask["maximum_amount"], decimalPointer(in.MaximumAmount)), ValidFrom: from, ValidTo: to, ApprovalReferenceID: field(mask["approval_reference_id"], in.ApprovalReferenceId)}))
}
func (h *Handler) DeleteCompensationBand(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteCompensationBand(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetCompensationBandByID(ctx context.Context, request *common.IDRequest) (*pb.CompensationBandResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bandResponse(h.service.GetCompensationBandByID(ctx, request.Id))
}
func (h *Handler) GetAllCompensationBands(ctx context.Context, request *common.PageRequest) (*pb.CompensationBandsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bands(h.service.GetAllCompensationBands(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationBandByBandCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationBandsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bands(h.service.FindCompensationBandByBandCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationBandByJobLevelCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationBandsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bands(h.service.FindCompensationBandByJobLevelCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationBandByJurisdictionCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationBandsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bands(h.service.FindCompensationBandByJurisdictionCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationBandByCurrencyCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationBandsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bands(h.service.FindCompensationBandByCurrencyCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationBandByApprovalReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationBandsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return bands(h.service.FindCompensationBandByApprovalReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
