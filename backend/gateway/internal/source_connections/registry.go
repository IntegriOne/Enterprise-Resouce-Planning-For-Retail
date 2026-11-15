// Package source_connections mengelola koneksi klien gRPC internal Gateway beserta propagasi konteks HMAC aman ke masing-masing manajemen.
package source_connections

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// IAMClientSet menyatukan klien gRPC layanan IAM yang dapat di-mock pada pengujian unit Gateway.
type IAMClientSet struct {
	Auth       iamv1.AuthServiceClient
	Users      iamv1.UserServiceClient
	Sessions   iamv1.SessionServiceClient
	Roles      iamv1.RoleServiceClient
	Permission iamv1.PermissionServiceClient
	Groups     iamv1.GroupServiceClient
}

// Registry mengelola seluruh koneksi downstream internal Gateway (Zona 2).
type Registry struct {
	iamConn    *grpc.ClientConn
	IAM        IAMClientSet
	IAMTimeout time.Duration
}

// Config menyimpan parameter target dan batas waktu koneksi gRPC downstream.
type Config struct {
	IAMTarget  string
	IAMTimeout time.Duration
	HMACSecret []byte
}

// NewRegistry membuka koneksi gRPC internal terkonfigurasi keepalive dan interceptor tanda tangan HMAC authcontext.
func NewRegistry(cfg Config) (*Registry, error) {
	target := strings.TrimSpace(cfg.IAMTarget)
	if target == "" {
		return nil, errors.New("IAM gRPC target is required")
	}
	if cfg.IAMTimeout <= 0 {
		cfg.IAMTimeout = 5 * time.Second
	}

	secretCopy := make([]byte, len(cfg.HMACSecret))
	copy(secretCopy, cfg.HMACSecret)

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(internalContextClientInterceptor(secretCopy)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Registry{
		iamConn:    conn,
		IAMTimeout: cfg.IAMTimeout,
		IAM: IAMClientSet{
			Auth:       iamv1.NewAuthServiceClient(conn),
			Users:      iamv1.NewUserServiceClient(conn),
			Sessions:   iamv1.NewSessionServiceClient(conn),
			Roles:      iamv1.NewRoleServiceClient(conn),
			Permission: iamv1.NewPermissionServiceClient(conn),
			Groups:     iamv1.NewGroupServiceClient(conn),
		},
	}, nil
}

// Close menutup seluruh koneksi gRPC internal Gateway saat shutdown.
func (r *Registry) Close() error {
	if r == nil || r.iamConn == nil {
		return nil
	}
	return r.iamConn.Close()
}

func internalContextClientInterceptor(hmacSecret []byte) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		caller, ok := authcontext.FromContext(ctx)
		if !ok {
			caller = authcontext.CallerContext{}
		}
		if strings.TrimSpace(caller.RequestID) == "" {
			caller.RequestID = "gw-internal"
		}
		outCtx := authcontext.AppendToOutgoingGRPCContext(ctx, caller, hmacSecret, time.Now())
		return invoker(outCtx, method, req, reply, cc, opts...)
	}
}
