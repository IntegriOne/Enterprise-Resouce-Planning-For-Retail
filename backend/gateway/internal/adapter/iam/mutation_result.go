package iam

import (
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

// ToMutationErrors mengonversi error Gateway/adapter/gRPC menjadi daftar MutationError agar frontend menerima tipe enum dan pesan error secara konsisten.
func ToMutationErrors(err error) []*model.MutationError {
	if err == nil {
		return noMutationErrors()
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

func noMutationErrors() []*model.MutationError {
	return []*model.MutationError{}
}

func mapGatewayCodeToErrorType(code middleware.ErrorCode) model.ErrorTypeEnum {
	candidate := model.ErrorTypeEnum(middleware.ErrorTypeForCode(code))
	if candidate.IsValid() {
		return candidate
	}
	return model.ErrorTypeEnumInternalError
}

func ptrBool(v bool) *bool {
	return &v
}

func ptrInt(v int) *int {
	return &v
}
