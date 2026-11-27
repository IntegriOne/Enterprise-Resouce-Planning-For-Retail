package source_connections

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultEFTimeout          = defaultUpstreamTimeout
	defaultEFFailureThreshold = defaultCircuitFailureThreshold
	defaultEFOpenDuration     = defaultCircuitOpenDuration
)

var (
	// ErrEFDisabled menandakan integrasi gRPC Enterprise Foundation dinonaktifkan lewat konfigurasi.
	ErrEFDisabled = errors.New("enterprise foundation integration is disabled")
	// ErrEFCircuitOpen menandakan pemutus sirkuit Enterprise Foundation sedang terbuka akibat kegagalan beruntun.
	ErrEFCircuitOpen = errors.New("enterprise foundation circuit breaker is open")
	// ErrEFUnavailable menandakan klien gRPC Enterprise Foundation belum siap atau sedang tidak tersedia.
	ErrEFUnavailable = errors.New("enterprise foundation client is unavailable")
)

// EnterpriseFoundationClientSet mengumpulkan seluruh stub klien gRPC Enterprise Foundation beserta proteksi pemutus sirkuit.
type EnterpriseFoundationClientSet struct {
	Addresses         efv1.AddressServiceClient
	LegalEntities     efv1.LegalEntityServiceClient
	Locations         efv1.LocationServiceClient
	OrganizationUnits efv1.OrganizationUnitServiceClient
	Parties           efv1.PartyServiceClient

	enabled bool
	timeout time.Duration
	initErr error
	breaker *circuitBreaker
}

// NewEnterpriseFoundationClientSet membentuk kumpulan stub gRPC Enterprise Foundation dari ClientConn siap pakai.
func NewEnterpriseFoundationClientSet(conn grpc.ClientConnInterface, timeout time.Duration) *EnterpriseFoundationClientSet {
	return newEnterpriseFoundationClientSetWithBreaker(
		conn,
		timeout,
		newCircuitBreaker(defaultEFFailureThreshold, defaultEFOpenDuration, ErrEFCircuitOpen, time.Now),
	)
}

func newEnterpriseFoundationClientSetWithBreaker(conn grpc.ClientConnInterface, timeout time.Duration, breaker *circuitBreaker) *EnterpriseFoundationClientSet {
	if timeout <= 0 {
		timeout = defaultEFTimeout
	}
	if breaker == nil {
		breaker = newCircuitBreaker(defaultEFFailureThreshold, defaultEFOpenDuration, ErrEFCircuitOpen, time.Now)
	}
	if conn == nil {
		return &EnterpriseFoundationClientSet{
			enabled: true,
			timeout: timeout,
			initErr: ErrEFUnavailable,
			breaker: breaker,
		}
	}
	return &EnterpriseFoundationClientSet{
		Addresses:         efv1.NewAddressServiceClient(conn),
		LegalEntities:     efv1.NewLegalEntityServiceClient(conn),
		Locations:         efv1.NewLocationServiceClient(conn),
		OrganizationUnits: efv1.NewOrganizationUnitServiceClient(conn),
		Parties:           efv1.NewPartyServiceClient(conn),
		enabled:           true,
		timeout:           timeout,
		breaker:           breaker,
	}
}

func newDisabledEnterpriseFoundationClientSet(timeout time.Duration, initErr error, enabled bool) *EnterpriseFoundationClientSet {
	if timeout <= 0 {
		timeout = defaultEFTimeout
	}
	return &EnterpriseFoundationClientSet{
		enabled: enabled,
		timeout: timeout,
		initErr: initErr,
		breaker: newCircuitBreaker(defaultEFFailureThreshold, defaultEFOpenDuration, ErrEFCircuitOpen, time.Now),
	}
}

// initEFConnection menyiapkan ClientConn non-blocking untuk Enterprise Foundation tanpa menggagalkan startup Gateway maupun IAM.
func initEFConnection(cfg Config, secretCopy []byte) (*grpc.ClientConn, *EnterpriseFoundationClientSet, time.Duration, error) {
	timeout := cfg.EFTimeout
	if timeout <= 0 {
		timeout = defaultEFTimeout
	}
	if !cfg.EFEnabled {
		return nil, newDisabledEnterpriseFoundationClientSet(timeout, ErrEFDisabled, false), timeout, nil
	}

	target := strings.TrimSpace(cfg.EFTarget)
	if target == "" {
		initErr := errors.New("enterprise foundation gRPC target is empty")
		return nil, newDisabledEnterpriseFoundationClientSet(timeout, initErr, true), timeout, initErr
	}

	conn, err := newResilientGRPCConn(target, secretCopy)
	if err != nil {
		return nil, newDisabledEnterpriseFoundationClientSet(timeout, err, true), timeout, err
	}

	return conn, NewEnterpriseFoundationClientSet(conn, timeout), timeout, nil
}

// Ready melaporkan apakah seluruh stub klien gRPC Enterprise Foundation aktif dan siap dipanggil.
func (c *EnterpriseFoundationClientSet) Ready() bool {
	return c != nil &&
		c.enabled &&
		c.initErr == nil &&
		c.Addresses != nil &&
		c.LegalEntities != nil &&
		c.Locations != nil &&
		c.OrganizationUnits != nil &&
		c.Parties != nil
}

// Enabled melaporkan apakah integrasi Enterprise Foundation diaktifkan oleh konfigurasi.
func (c *EnterpriseFoundationClientSet) Enabled() bool {
	return c != nil && c.enabled
}

// InitError mengembalikan kesalahan inisialisasi klien Enterprise Foundation bila ada.
func (c *EnterpriseFoundationClientSet) InitError() error {
	if c == nil {
		return ErrEFUnavailable
	}
	return c.initErr
}

// InvokeEF menjalankan satu panggilan unary Enterprise Foundation secara terisolasi lengkap dengan batas waktu, pemutus sirkuit, pemulihan panic, dan pemeriksaan respons nil.
func InvokeEF[Resp any](
	ctx context.Context,
	clients *EnterpriseFoundationClientSet,
	call func(context.Context, *EnterpriseFoundationClientSet) (*Resp, error),
) (resp *Resp, err error) {
	if clients == nil || !clients.Enabled() {
		return nil, middleware.ErrUnavailable("enterprise foundation service unavailable", ErrEFDisabled)
	}
	if !clients.Ready() {
		return nil, middleware.ErrUnavailable("enterprise foundation service unavailable", clients.InitError())
	}
	if call == nil {
		return nil, middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", errors.New("enterprise foundation rpc callback is nil"))
	}

	isProbe, cbErr := clients.breaker.beforeCall()
	if cbErr != nil {
		return nil, middleware.ErrUnavailable("enterprise foundation service unavailable", cbErr)
	}

	rpcCtx, cancel := withBoundedTimeout(ctx, clients.timeout)
	defer cancel()

	defer func() {
		if recovered := recover(); recovered != nil {
			clients.breaker.afterCall(isProbe, true, false)
			resp = nil
			err = middleware.ErrUnavailable(
				"enterprise foundation service unavailable",
				fmt.Errorf("recovered enterprise foundation panic: %v", recovered),
			)
		}
	}()

	res, rpcErr := call(rpcCtx, clients)
	clients.breaker.afterCall(isProbe, isUpstreamUnavailableFailure(ctx, rpcErr), isLocalCallerCancellation(ctx, rpcErr))
	if rpcErr != nil {
		return nil, MapEFGRPCError(rpcErr)
	}
	if res == nil {
		return nil, middleware.NewGatewayError(
			middleware.CodeInternalError,
			"internal server error",
			errors.New("enterprise foundation returned nil response"),
		)
	}
	return res, nil
}

// MapEFGRPCError menerjemahkan error gRPC Enterprise Foundation menjadi GatewayError aman tanpa membocorkan host internal maupun memutus sesi pengguna IAM.
func MapEFGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := middleware.AsGatewayError(err); ok {
		return err
	}
	if errors.Is(err, ErrEFDisabled) || errors.Is(err, ErrEFCircuitOpen) || errors.Is(err, ErrEFUnavailable) {
		return middleware.ErrUnavailable("enterprise foundation service unavailable", err)
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return middleware.ErrUnavailable("enterprise foundation request timed out", err)
	}

	st, ok := status.FromError(err)
	if !ok {
		return middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", err)
	}

	reason := ExtractEFErrorReason(err)
	if mapped := mapEFReasonError(reason, st.Message(), err); mapped != nil {
		return mapped
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return middleware.ErrInvalidArgument(sanitizeEFMessage(st.Message(), "invalid input"), err)
	case codes.NotFound:
		return middleware.ErrNotFound(sanitizeEFMessage(st.Message(), "resource not found"), err)
	case codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
		return middleware.ErrConflict(sanitizeEFMessage(st.Message(), "request conflicts with current state"), err)
	case codes.PermissionDenied:
		return middleware.ErrForbidden(sanitizeEFMessage(st.Message(), "forbidden"), err)
	case codes.ResourceExhausted:
		return middleware.ErrTooManyRequests("too many requests, please retry later", err)
	case codes.Unauthenticated:
		// Kegagalan HMAC/otentikasi internal antarlapisan gRPC tidak boleh mengakhiri sesi browser pengguna di Gateway.
		return middleware.ErrUnavailable("enterprise foundation service unavailable", err)
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
		return middleware.ErrUnavailable("enterprise foundation service unavailable", err)
	default:
		return middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", err)
	}
}

func mapEFReasonError(reason, rawMsg string, cause error) error {
	switch strings.TrimSpace(reason) {
	case "enterprise_foundation.service_unavailable":
		return middleware.ErrUnavailable("enterprise foundation service unavailable", cause)
	case "enterprise_foundation.invalid_input":
		return middleware.ErrInvalidArgument(sanitizeEFMessage(rawMsg, "invalid input"), cause)
	default:
		return nil
	}
}

// ExtractEFErrorReason mengekstrak kode alasan ErrorInfo dari error status gRPC Enterprise Foundation.
func ExtractEFErrorReason(err error) string {
	st, ok := status.FromError(err)
	if !ok {
		return ""
	}
	return extractErrorReason(st)
}

// ExtractEFErrorField mengekstrak nama field pertama pada BadRequest detail dari status gRPC Enterprise Foundation.
func ExtractEFErrorField(err error) string {
	st, ok := status.FromError(err)
	if !ok {
		return ""
	}
	return extractErrorField(st)
}

func sanitizeEFMessage(msg, fallback string) string {
	return sanitizeUpstreamMessage(msg, fallback, "enterprise-foundation:", ":9002")
}
