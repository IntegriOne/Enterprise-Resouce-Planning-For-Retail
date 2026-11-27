package enterprise_foundation

import (
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

// NoMutationErrors mengembalikan slice kosong terinisialisasi untuk payload mutasi yang sukses.
func NoMutationErrors() []*model.MutationError {
	return []*model.MutationError{}
}

// ToMutationErrors mengonversi error Gateway/adapter/gRPC menjadi daftar MutationError aman untuk frontend.
func ToMutationErrors(err error) []*model.MutationError {
	if err == nil {
		return NoMutationErrors()
	}
	gwErr, ok := middleware.AsGatewayError(err)
	if !ok {
		gwErr = middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", err)
	}
	msg := strings.TrimSpace(gwErr.Message)
	if msg == "" {
		msg = "internal server error"
	}
	return []*model.MutationError{
		{
			Type:    mapGatewayCodeToErrorType(gwErr.Code),
			Message: msg,
		},
	}
}

func mapGatewayCodeToErrorType(code middleware.ErrorCode) model.ErrorTypeEnum {
	candidate := model.ErrorTypeEnum(middleware.ErrorTypeForCode(code))
	if candidate.IsValid() {
		return candidate
	}
	return model.ErrorTypeEnumInternalError
}
