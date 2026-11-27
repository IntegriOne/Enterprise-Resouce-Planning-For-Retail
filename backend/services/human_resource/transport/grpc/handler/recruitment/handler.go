// Package recruitment memetakan kontrak ATS bertipe ke service HRM.
package recruitment

import (
	"context"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	// Kontrak privat wajib memiliki implementasi eksplisit untuk setiap RPC.
	pb.UnsafeRecruitmentServiceServer
	service *service.Service
}

// Register mengikat RPC hanya pada dependency service yang tersedia.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterRecruitmentServiceServer(registrar, &Handler{service: svc})
	return nil
}

func field[T any](selected bool, value *T) operation.Field[T] {
	return operation.Field[T]{Set: selected, Value: value}
}

func enumPointer[T ~string](value *string) *T {
	if value == nil {
		return nil
	}
	converted := T(*value)
	return &converted
}

func stringPointer[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	converted := string(*value)
	return &converted
}

func decimalPointer(value *string) *entity.Decimal {
	return enumPointer[entity.Decimal](value)
}

func optionalDate(value *string) (*time.Time, error) {
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
	formatted := value.Format("2006-01-02")
	return &formatted
}

func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
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
	parsed, err := optionalDate(value)
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

func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return field[[]byte](selected, nil)
	}
	return field(selected, &value)
}

func pageInfo[T any](page operation.Page[T]) *common.PageInfo {
	return &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}
}

func deleteRecord(ctx context.Context, request *common.IDRequest, remove func(context.Context, string) error) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := remove(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
