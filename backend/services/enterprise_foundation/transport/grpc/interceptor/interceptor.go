// Package interceptor menyediakan unary interceptor gRPC untuk pemulihan panic,
// verifikasi konteks internal Gateway berbasis HMAC, serta pencatatan log terstruktur.
package interceptor

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

const (
	grpcHealthServicePrefix     = "/grpc.health.v1."
	grpcReflectionServicePrefix = "/grpc.reflection."
)

// RecoveryUnaryInterceptor menangkap panic di handler unary agar server tidak crash
// dan mengembalikan status gRPC Internal tanpa membocorkan detail runtime ke pemanggil.
func RecoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				if logger != nil {
					logger.Error("panic recovered in Enterprise Foundation gRPC unary handler",
						slog.String("method", info.FullMethod),
						slog.Any("panic", rec),
					)
				}
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// InternalContextUnaryInterceptor mengekstrak dan memverifikasi metadata pemanggil tepercaya
// dari Gateway sebelum diteruskan ke handler domain (gagal tertutup bila bukti HMAC tidak sah).
func InternalContextUnaryInterceptor(hmacSecret []byte, logger *slog.Logger) grpc.UnaryServerInterceptor {
	secretCopy := append([]byte(nil), hmacSecret...)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isInfrastructureMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		caller, ok, err := authcontext.ExtractFromIncomingGRPCContext(ctx, secretCopy, time.Now())
		if err != nil {
			if logger != nil {
				logger.Warn("rejected gRPC request with invalid internal gateway proof",
					slog.String("method", info.FullMethod),
					slog.String("error", err.Error()),
				)
			}
			return nil, status.Error(codes.Unauthenticated, "invalid internal caller context")
		}
		if !ok && len(secretCopy) > 0 {
			if logger != nil {
				logger.Warn("rejected gRPC request missing internal caller context",
					slog.String("method", info.FullMethod),
				)
			}
			return nil, status.Error(codes.Unauthenticated, "missing internal caller context")
		}
		if ok {
			ctx = authcontext.WithCaller(ctx, caller)
		}

		return handler(ctx, req)
	}
}

// LoggingUnaryInterceptor mencatat ringkasan eksekusi setiap request gRPC unary secara terstruktur.
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startedAt := time.Now()
		resp, err := handler(ctx, req)
		if logger == nil {
			return resp, err
		}

		code := status.Code(err)
		attrs := []any{
			slog.String("method", info.FullMethod),
			slog.String("grpc_code", code.String()),
			slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
		}
		if caller, ok := authcontext.FromContext(ctx); ok {
			attrs = append(attrs,
				slog.String("request_id", caller.RequestID),
				slog.String("principal_id", caller.PrincipalID),
			)
		}

		if err != nil {
			if code == codes.Internal || code == codes.Unavailable || code == codes.Unknown {
				logger.Error("gRPC unary request failed", attrs...)
			} else {
				logger.Warn("gRPC unary request rejected", attrs...)
			}
			return resp, err
		}

		if !isInfrastructureMethod(info.FullMethod) {
			logger.Info("gRPC unary request completed", attrs...)
		}
		return resp, err
	}
}

func isInfrastructureMethod(fullMethod string) bool {
	return strings.HasPrefix(fullMethod, grpcHealthServicePrefix) ||
		strings.HasPrefix(fullMethod, grpcReflectionServicePrefix)
}
