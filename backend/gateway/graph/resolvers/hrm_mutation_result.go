package resolvers

import (
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

func hrmMutationErrors(err error) []*model.MutationError {
	if err == nil {
		return []*model.MutationError{}
	}
	category := model.ErrorTypeEnumInternalError
	message := "internal server error"
	if gatewayError, ok := middleware.AsGatewayError(err); ok {
		category = model.ErrorTypeEnum(middleware.ErrorTypeForCode(gatewayError.Code))
		if safe := strings.TrimSpace(gatewayError.Message); safe != "" {
			message = safe
		}
	}
	return []*model.MutationError{{Type: category, Message: message}}
}
