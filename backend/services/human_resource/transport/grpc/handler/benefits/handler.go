// Package benefits memetakan RPC benefit ke service HRM pemilik data.
package benefits

import (
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/benefits"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Handler struct {
	pb.UnimplementedBenefitsServiceServer
	service *service.Service
}

// Register mengikat service dan registrar yang sudah tervalidasi.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterBenefitsServiceServer(registrar, &Handler{service: svc})
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
