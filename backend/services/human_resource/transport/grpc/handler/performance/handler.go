// Package performance memetakan RPC evaluasi HRM ke application service.
package performance

import (
	"context"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	pb.UnimplementedPerformanceServiceServer
	service *service.Service
}

// Register menolak registrasi tanpa service domain yang valid.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterPerformanceServiceServer(registrar, &Handler{service: svc})
	return nil
}

func field[T any](selected bool, value *T) operation.Field[T] {
	return operation.Field[T]{Set: selected, Value: value}
}
func enumPtr[T ~string](value *string) *T {
	if value == nil {
		return nil
	}
	converted := T(*value)
	return &converted
}
func decimalPtr(value *string) *entity.Decimal {
	if value == nil {
		return nil
	}
	converted := entity.Decimal(*value)
	return &converted
}
func decimalString(value *entity.Decimal) *string {
	if value == nil {
		return nil
	}
	converted := string(*value)
	return &converted
}
func datePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := response.Date(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func dateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	parsed := value.Format("2006-01-02")
	return &parsed
}
func timeField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(value)
	return field(selected, parsed), err
}
func dateField(selected bool, value *string) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := datePtr(value)
	return field(selected, parsed), err
}
func jsonField(selected bool, value *common.JsonDocument) (operation.Field[entity.JSON], error) {
	if !selected {
		return operation.Field[entity.JSON]{}, nil
	}
	if value == nil {
		return field[entity.JSON](true, nil), nil
	}
	parsed, err := response.JSON(value)
	return field(true, &parsed), err
}
func decimalField(selected bool, value *string) operation.Field[entity.Decimal] {
	return field(selected, decimalPtr(value))
}
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return field[[]byte](selected, nil)
	}
	return field(selected, &value)
}
func pageInfo[T any](page operation.Page[T]) *common.PageInfo {
	return &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}
}
func remove(ctx context.Context, req *common.IDRequest, fn func(context.Context, string) error) (*common.EmptyResponse, error) {
	if req == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := fn(ctx, req.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
