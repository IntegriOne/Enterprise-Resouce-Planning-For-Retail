// Package leave memetakan RPC cuti ke application service pemilik domain.
package leave

import (
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	pb.UnimplementedLeaveServiceServer
	service *service.Service
}

// Register mengikat service dengan dependency yang sudah divalidasi.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterLeaveServiceServer(registrar, &Handler{service: svc})
	return nil
}

func field[T any](selected bool, value *T) operation.Field[T] {
	return operation.Field[T]{Set: selected, Value: value}
}
func decimal(value *string) *entity.Decimal {
	if value == nil {
		return nil
	}
	result := entity.Decimal(*value)
	return &result
}
func decimalString(value *entity.Decimal) *string {
	if value == nil {
		return nil
	}
	result := string(*value)
	return &result
}
func optionalDate(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := response.Date(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func dateField(selected bool, raw *string) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	value, err := optionalDate(raw)
	return field(selected, value), err
}
func dateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	result := value.Format("2006-01-02")
	return &result
}
func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
