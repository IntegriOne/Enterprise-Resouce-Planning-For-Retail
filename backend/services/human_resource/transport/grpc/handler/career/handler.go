// Package career menerjemahkan RPC bertipe ke service pengembangan karier HRM.
package career

import (
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/career/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedCareerServiceServer
	service *service.Service
}

// Register memasang service karier pada registrar milik server HRM.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterCareerServiceServer(registrar, &server{service: svc})
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
func level(v *int32) (*int16, error) {
	if v == nil {
		return nil, nil
	}
	if *v < -32768 || *v > 32767 {
		return nil, hrerrors.ErrInvalidInput
	}
	result := int16(*v)
	return &result, nil
}
func levelValue(v *int16) *int32 {
	if v == nil {
		return nil
	}
	result := int32(*v)
	return &result
}
