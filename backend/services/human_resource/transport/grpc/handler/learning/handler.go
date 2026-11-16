// Package learning memetakan RPC pembelajaran ke service HRM pemilik data.
package learning

import (
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/learning"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Handler struct {
	pb.UnimplementedLearningServiceServer
	service *service.Service
}

// Register mengikat service dan registrar yang sudah tervalidasi.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterLearningServiceServer(registrar, &Handler{service: svc})
	return nil
}
func field[T any](selected bool, value *T) operation.Field[T] {
	return operation.Field[T]{Set: selected, Value: value}
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
	formatted := value.Format("2006-01-02")
	return &formatted
}
func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
func timestampField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	converted, err := response.OptionalTimestamp(value)
	return field(selected, converted), err
}
func decimalPointer(value *string) *entity.Decimal {
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
func statusPointer(value *string) *entity.LearningStatus {
	if value == nil {
		return nil
	}
	converted := entity.LearningStatus(*value)
	return &converted
}
func level(value int32) (int16, error) {
	if value < -32768 || value > 32767 {
		return 0, hrerrors.ErrInvalidInput
	}
	return int16(value), nil
}
func levelField(selected bool, value *int32) (operation.Field[int16], error) {
	if !selected || value == nil {
		return field[int16](selected, nil), nil
	}
	converted, err := level(*value)
	if err != nil {
		return operation.Field[int16]{}, err
	}
	return field(selected, &converted), nil
}
