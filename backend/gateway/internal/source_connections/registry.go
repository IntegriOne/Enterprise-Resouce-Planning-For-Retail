// Package source_connections mengelola koneksi klien gRPC internal Gateway beserta propagasi konteks HMAC aman ke masing-masing manajemen.
package source_connections

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type iamClientState struct {
	timeout time.Duration
	initErr error
	breaker *circuitBreaker
}

// IAMClientSet menyatukan klien gRPC layanan IAM yang dapat di-mock pada pengujian unit Gateway.
type IAMClientSet struct {
	Auth                 iamv1.AuthServiceClient
	Users                iamv1.UserServiceClient
	Sessions             iamv1.SessionServiceClient
	Roles                iamv1.RoleServiceClient
	Permission           iamv1.PermissionServiceClient
	Groups               iamv1.GroupServiceClient
	EmployeeProvisioning iamv1.EmployeeProvisioningServiceClient

	state *iamClientState
}

// Registry mengelola seluruh koneksi downstream internal Gateway (Zona 2).
type Registry struct {
	iamConn              *grpc.ClientConn
	IAM                  IAMClientSet
	IAMTimeout           time.Duration
	IAMInitError         error
	efConn               *grpc.ClientConn
	EnterpriseFoundation *EnterpriseFoundationClientSet
	EFTimeout            time.Duration
	EFEnabled            bool
	EFInitError          error
	HumanResource        *grpc.ClientConn
	HRMInitError         error
}

// Config menyimpan parameter target dan batas waktu koneksi gRPC downstream.
type Config struct {
	IAMTarget      string
	IAMTimeout     time.Duration
	EFEnabled      bool
	EFTarget       string
	EFTimeout      time.Duration
	HMACSecret     []byte
	HRMEnabled     bool
	HRM            HumanResourceConnectionConfig
	HRMConfigError error
}

// NewRegistry membuka koneksi gRPC internal terkonfigurasi keepalive, backoff cepat, dan interceptor HMAC tanpa memblokir startup saat upstream tidak aktif.
func NewRegistry(cfg Config) (*Registry, error) {
	secretCopy := make([]byte, len(cfg.HMACSecret))
	copy(secretCopy, cfg.HMACSecret)

	iamConn, iamClients, iamTimeout, iamInitErr := initIAMConnection(cfg, secretCopy)
	efConn, efClients, efTimeout, efInitErr := initEFConnection(cfg, secretCopy)

	var hrmConn *grpc.ClientConn
	hrmInitErr := cfg.HRMConfigError
	if cfg.HRMEnabled && hrmInitErr == nil {
		hrmConn, hrmInitErr = NewHumanResourceConnection(cfg.HRM, secretCopy)
	}

	return &Registry{
		iamConn:              iamConn,
		IAM:                  iamClients,
		IAMTimeout:           iamTimeout,
		IAMInitError:         iamInitErr,
		efConn:               efConn,
		EnterpriseFoundation: efClients,
		EFTimeout:            efTimeout,
		EFEnabled:            cfg.EFEnabled,
		EFInitError:          efInitErr,
		HumanResource:        hrmConn,
		HRMInitError:         hrmInitErr,
	}, nil
}

// Close menutup seluruh koneksi gRPC internal Gateway saat shutdown.
func (r *Registry) Close() error {
	if r == nil {
		return nil
	}
	var errs []error
	if r.iamConn != nil {
		if err := r.iamConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.efConn != nil {
		if err := r.efConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.HumanResource != nil {
		if err := r.HumanResource.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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
