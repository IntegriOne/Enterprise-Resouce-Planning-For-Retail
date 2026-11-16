// Package interceptor menjaga batas autentikasi, deadline dan error transport HRM.
package interceptor

import (
	"context"
	"log/slog"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Unary menolak proof ambigu dan menghapus detail error sebelum respons keluar.
func Unary(secret []byte, logger *slog.Logger) grpc.UnaryServerInterceptor {
	key := append([]byte(nil), secret...)
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (result any, err error) {
		started := time.Now()
		method := "unknown"
		if info != nil {
			method = info.FullMethod
		}
		defer func() {
			if recover() != nil {
				result = nil
				err = status.Error(codes.Internal, "internal server error")
			}
			err = response.TransportError(err)
			level := slog.LevelInfo
			if code := status.Code(err); code == codes.Internal || code == codes.Unavailable {
				level = slog.LevelError
			}
			logger.Log(ctx, level, "HRM RPC completed", "method", method,
				"code", status.Code(err).String(), "duration", time.Since(started))
		}()
		if next == nil || info == nil {
			return nil, status.Error(codes.Internal, "invalid handler configuration")
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if method != "/grpc.health.v1.Health/Check" {
			if len(key) < 32 || ambiguousIdentity(ctx) {
				return nil, status.Error(codes.Unauthenticated, "verified caller context required")
			}
			caller, present, proofErr := authcontext.ExtractFromIncomingGRPCContext(ctx, key, time.Now())
			if proofErr != nil || !present || !caller.Authenticated {
				return nil, status.Error(codes.Unauthenticated, "verified caller context required")
			}
			ctx = authcontext.WithCaller(ctx, caller)
		}
		result, err = next(ctx, req)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return result, nil
	}
}

func ambiguousIdentity(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return true
	}
	for _, key := range []string{
		authcontext.MetaPrincipalID, authcontext.MetaSessionID, authcontext.MetaUserKind,
		authcontext.MetaAuthEpoch, authcontext.MetaPermRevision, authcontext.MetaGatewayProof,
		authcontext.MetaRequestID, authcontext.MetaClientIP, authcontext.MetaUserAgent,
	} {
		if len(md.Get(key)) > 1 {
			return true
		}
	}
	return false
}
