package server

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

func buildErrorPresenter(logger *slog.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		if gwErr, ok := middleware.AsGatewayError(err); ok {
			if logger != nil && (gwErr.Code == middleware.CodeInternalError || gwErr.Code == middleware.CodeUnavailable) {
				logger.Warn("gateway request rejected",
					slog.String("code", string(gwErr.Code)),
					slog.String("message", gwErr.Message),
					slog.Any("cause", gwErr.Err),
				)
			}
			gqlErr := graphql.DefaultErrorPresenter(ctx, err)
			gqlErr.Message = gwErr.Message
			if gqlErr.Extensions == nil {
				gqlErr.Extensions = make(map[string]any, 2)
			}
			gqlErr.Extensions["code"] = string(gwErr.Code)
			gqlErr.Extensions["type"] = middleware.ErrorTypeForCode(gwErr.Code)
			return gqlErr
		}

		gqlErr := graphql.DefaultErrorPresenter(ctx, err)
		var parsedErr *gqlerror.Error
		if errors.As(err, &parsedErr) {
			if gqlErr.Extensions == nil {
				gqlErr.Extensions = make(map[string]any, 2)
			}
			if _, exists := gqlErr.Extensions["code"]; !exists {
				msgLower := strings.ToLower(gqlErr.Message)
				switch {
				case strings.Contains(msgLower, "operation complexity") || strings.Contains(msgLower, "complexity limit"):
					gqlErr.Extensions["code"] = "COMPLEXITY_LIMIT_EXCEEDED"
					gqlErr.Extensions["type"] = "ARGUMENT"
				case strings.Contains(msgLower, "introspection"):
					gqlErr.Extensions["code"] = string(middleware.CodeForbidden)
					gqlErr.Extensions["type"] = "AUTHORIZATION"
				default:
					if strings.Contains(msgLower, "json request body could not be decoded") ||
						strings.Contains(msgLower, "json body could not be decoded") ||
						strings.Contains(msgLower, "cannot unmarshal") ||
						strings.Contains(msgLower, "unexpected eof") {
						gqlErr.Message = "invalid JSON request body"
					}
					gqlErr.Extensions["code"] = "BAD_REQUEST"
					gqlErr.Extensions["type"] = "ARGUMENT"
				}
			}
			return gqlErr
		}

		// Jangan pernah membocorkan pesan error mentah yang tidak terklasifikasi ke klien.
		if logger != nil {
			logger.Error("unhandled error presented to graphql client", slog.String("error", err.Error()))
		}
		gqlErr.Message = "internal server error"
		gqlErr.Extensions = map[string]any{
			"code": string(middleware.CodeInternalError),
			"type": "INTERNAL_ERROR",
		}
		return gqlErr
	}
}

func buildRecoverFunc(logger *slog.Logger) graphql.RecoverFunc {
	return func(ctx context.Context, recovered any) error {
		if logger != nil {
			logger.Error("graphql resolver panic recovered",
				slog.Any("panic", recovered),
				slog.String("stack", string(debug.Stack())),
			)
		}
		return middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", nil)
	}
}
