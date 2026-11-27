// Package personnel memetakan RPC bertipe ke application service hubungan kerja HRM.
package personnel

import (
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
)

type Handler struct {
	// Kontrak privat wajib memiliki implementasi eksplisit untuk setiap RPC.
	pb.UnsafePersonnelServiceServer
	service *service.Service
}

// Register mengikat RPC hanya pada service dan registrar yang valid.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterPersonnelServiceServer(registrar, &Handler{service: svc})
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
func statusPointer(value *string) *entity.WorkerStatus {
	if value == nil {
		return nil
	}
	converted := entity.WorkerStatus(*value)
	return &converted
}
func statusString(value *entity.WorkerStatus) *string {
	if value == nil {
		return nil
	}
	converted := string(*value)
	return &converted
}
func decimalPointer(value *string) *entity.Decimal {
	if value == nil {
		return nil
	}
	converted := entity.Decimal(*value)
	return &converted
}
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return field[[]byte](selected, nil)
	}
	return field(selected, &value)
}
