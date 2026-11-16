// Package onboarding memetakan RPC bertipe ke application service onboarding HRM.
package onboarding

import (
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/onboarding/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/onboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedOnboardingServiceServer
	service *service.Service
}

// Register hanya menerima service dan registrar yang valid.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if operation.RequireDependency(registrar) != nil || svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterOnboardingServiceServer(registrar, &Handler{service: svc})
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
