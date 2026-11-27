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
		meta, _ := middleware.RequestMetadataFromContext(ctx)
		if err == nil {
			if logger != nil {
				logger.Error("nil error presented to graphql client", requestIDAttr(meta.RequestID)...)
			}
			return internalServerGQLError()
		}

		if gwErr, ok := middleware.AsGatewayError(err); ok {
			if logger != nil {
				attrs := append([]any{
					slog.String("code", string(gwErr.Code)),
					slog.String("message", gwErr.Message),
				}, requestIDAttr(meta.RequestID)...)
				if gwErr.Err != nil {
					attrs = append(attrs, slog.Any("cause", gwErr.Err))
				}
				switch gwErr.Code {
				case middleware.CodeInternalError:
					logger.Error("gateway internal error", attrs...)
				case middleware.CodeUnavailable:
					logger.Warn("gateway upstream unavailable", attrs...)
				}
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
		if _, ok := errors.AsType[*gqlerror.Error](err); ok {
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
			attrs := append([]any{slog.String("error", err.Error())}, requestIDAttr(meta.RequestID)...)
			logger.Error("unhandled error presented to graphql client", attrs...)
		}
		gqlErr.Message = "internal server error"
		gqlErr.Extensions = map[string]any{
			"code": string(middleware.CodeInternalError),
			"type": "INTERNAL_ERROR",
		}
		return gqlErr
	}
}

func requestIDAttr(requestID string) []any {
	if strings.TrimSpace(requestID) == "" {
		return nil
	}
	return []any{slog.String("request_id", requestID)}
}

func internalServerGQLError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: "internal server error",
		Extensions: map[string]any{
			"code": string(middleware.CodeInternalError),
			"type": "INTERNAL_ERROR",
		},
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
