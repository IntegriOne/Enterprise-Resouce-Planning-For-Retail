package handler

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services"
)

// RegisterAll mendaftarkan layanan gRPC IAM ke server secara fail-closed.
func RegisterAll(grpcSrv *grpc.Server, reg *services.Registry, logger *slog.Logger) error {
	if grpcSrv == nil {
		return fmt.Errorf("grpc server cannot be nil")
	}
	if reg == nil {
		return fmt.Errorf("iam service registry cannot be nil")
	}
	if err := reg.Validate(); err != nil {
		return fmt.Errorf("invalid iam service registry: %w", err)
	}

	errMapper := NewErrorMapper(logger)

	permH, err := NewPermissionHandler(reg.Permissions, errMapper)
	if err != nil {
		return err
	}
	roleH, err := NewRoleHandler(reg.Roles, errMapper)
	if err != nil {
		return err
	}
	groupH, err := NewGroupHandler(reg.Groups, errMapper)
	if err != nil {
		return err
	}
	sessH, err := NewSessionHandler(reg.Sessions, errMapper)
	if err != nil {
		return err
	}
	userH, err := NewUserHandler(reg.Users, errMapper)
	if err != nil {
		return err
	}
	authH, err := NewAuthHandler(reg.Users, errMapper)
	if err != nil {
		return err
	}

	iamv1.RegisterPermissionServiceServer(grpcSrv, permH)
	iamv1.RegisterRoleServiceServer(grpcSrv, roleH)
	iamv1.RegisterGroupServiceServer(grpcSrv, groupH)
	iamv1.RegisterSessionServiceServer(grpcSrv, sessH)
	iamv1.RegisterUserServiceServer(grpcSrv, userH)
	iamv1.RegisterAuthServiceServer(grpcSrv, authH)
	if reg.EmployeeProvisioning != nil {
		iamv1.RegisterEmployeeProvisioningServiceServer(grpcSrv, &EmployeeProvisioningHandler{service: reg.EmployeeProvisioning, errMapper: errMapper})
	}
	return nil
}

func nilRequestError(errMapper *ErrorMapper) error {
	err := iamerrors.New(iamerrors.KindInvalidInput, iamerrors.ErrInvalidInput.Code, "request cannot be nil")
	if errMapper == nil {
		return NewErrorMapper(nil).ToStatus(err)
	}
	return errMapper.ToStatus(err)
}
