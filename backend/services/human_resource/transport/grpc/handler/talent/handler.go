// Package talent menerjemahkan RPC bertipe ke service pengelolaan talenta HRM.
package talent

import (
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type server struct {
	pb.UnimplementedTalentServiceServer
	service *service.Service
}

// Register memasang service talenta pada registrar milik server HRM.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterTalentServiceServer(registrar, &server{service: svc})
	return nil
}
func timestamp(v *time.Time) *timestamppb.Timestamp {
	if v == nil {
		return nil
	}
	return timestamppb.New(*v)
}
func timeField(set bool, v *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !set {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(v)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func bytesField(set bool, v []byte) operation.Field[[]byte] {
	if v == nil {
		return operation.Field[[]byte]{Set: set}
	}
	return operation.Field[[]byte]{Set: set, Value: &v}
}
func score(v *int32) (*int16, error) {
	if v == nil {
		return nil, nil
	}
	if *v < -32768 || *v > 32767 {
		return nil, hrerrors.ErrInvalidInput
	}
	result := int16(*v)
	return &result, nil
}
func decimal(v *string) *entity.Decimal {
	if v == nil {
		return nil
	}
	result := entity.Decimal(*v)
	return &result
}
