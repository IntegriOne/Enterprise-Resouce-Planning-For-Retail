package interceptor

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// InternalContextUnaryInterceptor memvalidasi tanda tangan konteks caller internal dari Gateway bila kunci HMAC dikonfigurasi.
func InternalContextUnaryInterceptor(hmacSecret []byte, requireProof bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if info != nil && (strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") ||
			strings.HasPrefix(info.FullMethod, "/grpc.reflection.")) {
			return handler(ctx, req)
		}
		if len(hmacSecret) > 0 {
			caller, hasAny, err := authcontext.ExtractFromIncomingGRPCContext(ctx, hmacSecret, time.Now())
			if err != nil {
				return nil, status.Error(codes.PermissionDenied, "invalid internal caller context")
			}
			if requireProof && !hasAny {
				return nil, status.Error(codes.PermissionDenied, "missing internal caller context")
			}
			if hasAny {
				ctx = authcontext.WithCaller(ctx, caller)
			}
		}
		return handler(ctx, req)
	}
}

// RecoveryUnaryInterceptor menangkap panic pada handler gRPC agar server tidak crash.
func RecoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				method := ""
				if info != nil {
					method = info.FullMethod
				}
				logger.Error("grpc handler panic recovered",
					slog.String("method", method),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// LoggingUnaryInterceptor mencatat metode, durasi eksekusi, dan kode status setiap panggilan gRPC.
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		method := ""
		if info != nil {
			method = info.FullMethod
		}

		code := status.Code(err)
		if err != nil {
			logger.Warn("grpc request failed",
				slog.String("method", method),
				slog.String("code", code.String()),
				slog.Duration("duration", duration),
				slog.String("error", err.Error()),
			)
			return resp, err
		}

		logger.Info("grpc request completed",
			slog.String("method", method),
			slog.String("code", code.String()),
			slog.Duration("duration", duration),
		)
		return resp, nil
	}
}
