// Package competency memetakan RPC kompetensi ke service HRM pemilik data.
package competency

import (
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Handler struct {
	pb.UnimplementedCompetencyServiceServer
	service *service.Service
}

// Register mengikat service dan registrar yang sudah tervalidasi.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterCompetencyServiceServer(registrar, &Handler{service: svc})
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
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return field[[]byte](selected, nil)
	}
	return field(selected, &value)
}
