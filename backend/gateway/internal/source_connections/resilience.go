package source_connections

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const (
	defaultUpstreamTimeout         = 5 * time.Second
	defaultCircuitFailureThreshold = 5
	defaultCircuitOpenDuration     = 3 * time.Second
	maxSafeUpstreamErrorMsgLen     = 240
)

var commonBlockedErrorFragments = []string{
	"rpc error:",
	"connection refused",
	"connection reset",
	"transport:",
	"dial tcp",
	"no such host",
	"lookup ",
	"grpc:",
	"tls:",
	"x509:",
	"broken pipe",
	"unexpected eof",
	"context deadline exceeded",
	"sqlstate",
	"pq:",
	"gorm",
	"127.0.0.",
	"panic:",
	"stack trace",
}

type circuitState int

const (
	circuitClosed circuitState = iota
	circuitOpen
	circuitHalfOpen
)

// circuitBreaker memutus panggilan sementara saat layanan downstream mengalami gangguan agar Gateway dan modul lain tetap responsif.
type circuitBreaker struct {
	mu               sync.Mutex
	state            circuitState
	failures         int
	threshold        int
	openDuration     time.Duration
	openUntil        time.Time
	halfOpenInFlight bool
	openErr          error
	nowFn            func() time.Time
}

func newCircuitBreaker(threshold int, openDuration time.Duration, openErr error, nowFn func() time.Time) *circuitBreaker {
	if threshold <= 0 {
		threshold = defaultCircuitFailureThreshold
	}
	if openDuration <= 0 {
		openDuration = defaultCircuitOpenDuration
	}
	if openErr == nil {
		openErr = errors.New("upstream circuit breaker is open")
	}
	if nowFn == nil {
		nowFn = time.Now
	}
	return &circuitBreaker{
		state:        circuitClosed,
		threshold:    threshold,
		openDuration: openDuration,
		openErr:      openErr,
		nowFn:        nowFn,
	}
}

func (cb *circuitBreaker) beforeCall() (bool, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := cb.nowFn()
	switch cb.state {
	case circuitClosed:
		return false, nil
	case circuitOpen:
		if now.Before(cb.openUntil) {
			return false, cb.openErr
		}
		cb.state = circuitHalfOpen
		cb.halfOpenInFlight = true
		return true, nil
	case circuitHalfOpen:
		if cb.halfOpenInFlight {
			return false, cb.openErr
		}
		cb.halfOpenInFlight = true
		return true, nil
	default:
		cb.state = circuitClosed
		return false, nil
	}
}

func (cb *circuitBreaker) afterCall(isProbe bool, upstreamDown bool, callerCanceled bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if isProbe {
		cb.halfOpenInFlight = false
		if callerCanceled && !upstreamDown {
			// Pembatalan lokal oleh klien saat probe berlangsung tidak membuktikan upstream sehat; pertahankan state half-open/open.
			cb.state = circuitOpen
			return
		}
		if !upstreamDown {
			cb.failures = 0
			cb.state = circuitClosed
			return
		}
		cb.state = circuitOpen
		cb.openUntil = cb.nowFn().Add(cb.openDuration)
		return
	}

	if !upstreamDown {
		if !callerCanceled {
			cb.failures = 0
		}
		return
	}

	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = circuitOpen
		cb.openUntil = cb.nowFn().Add(cb.openDuration)
	}
}

// newResilientGRPCConn membentuk ClientConn gRPC non-blocking dengan WaitForReady(false), backoff cepat 250ms-2s, keepalive, dan interceptor HMAC internal.
func newResilientGRPCConn(target string, hmacSecret []byte) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(internalContextClientInterceptor(hmacSecret)),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(false)),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  250 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   2 * time.Second,
			},
			MinConnectTimeout: 2 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
}

func withBoundedTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = defaultUpstreamTimeout
	}
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 && remaining <= timeout {
			return ctx, func() {}
		}
	}
	return context.WithTimeout(ctx, timeout)
}

func isLocalCallerCancellation(parentCtx context.Context, err error) bool {
	if err == nil || parentCtx == nil || parentCtx.Err() == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.Canceled
}

func isUpstreamUnavailableFailure(parentCtx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if isLocalCallerCancellation(parentCtx, err) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded:
		// Jangan memicu circuit breaker global IAM akibat domain eksternal verifikasi scope belum terdaftar/aktif.
		if extractErrorReason(st) == "iam.scope.owner_verification_unavailable" {
			return false
		}
		return true
	case codes.Canceled:
		// Pembatalan transport tanpa pembatalan dari pemanggil menandakan koneksi upstream terputus di tengah eksekusi.
		return parentCtx == nil || parentCtx.Err() == nil
	default:
		return false
	}
}

func extractErrorReason(st *status.Status) string {
	if st == nil {
		return ""
	}
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			return strings.TrimSpace(info.GetReason())
		}
	}
	return ""
}

func extractErrorField(st *status.Status) string {
	if st == nil {
		return ""
	}
	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok {
			for _, fv := range badReq.GetFieldViolations() {
				if field := strings.TrimSpace(fv.GetField()); field != "" {
					return strings.ToLower(field)
				}
			}
		}
	}
	return ""
}

func sanitizeUpstreamMessage(msg, fallback string, extraBlocked ...string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" || len(trimmed) > maxSafeUpstreamErrorMsgLen {
		return fallback
	}
	lower := strings.ToLower(trimmed)
	for _, fragment := range commonBlockedErrorFragments {
		if strings.Contains(lower, fragment) {
			return fallback
		}
	}
	for _, fragment := range extraBlocked {
		if fragment != "" && strings.Contains(lower, fragment) {
			return fallback
		}
	}
	return trimmed
}
