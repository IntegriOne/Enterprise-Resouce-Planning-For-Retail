// Package authcontext menyediakan propagasi identitas pemanggil tepercaya lintas gateway dan gRPC internal.
package authcontext

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
)

const (
	// MetaPrincipalID menyimpan UUID principal terautentikasi dari gateway.
	MetaPrincipalID = "x-ig-principal-id"
	// MetaSessionID menyimpan UUID sesi aktif dari gateway.
	MetaSessionID = "x-ig-session-id"
	// MetaUserKind menyimpan kategori utama user (internal, external, customer, system).
	MetaUserKind = "x-ig-user-kind"
	// MetaAuthEpoch menyimpan epoch autentikasi monotonik user.
	MetaAuthEpoch = "x-ig-auth-epoch"
	// MetaPermRevision menyimpan revisi hak akses monotonik user.
	MetaPermRevision = "x-ig-perm-revision"
	// MetaRequestID menyimpan korelasi request ID end-to-end.
	MetaRequestID = "x-request-id"
	// MetaClientIP menyimpan alamat IP klien asal setelah melalui proxy tepercaya.
	MetaClientIP = "x-forwarded-for"
	// MetaUserAgent menyimpan User-Agent klien.
	MetaUserAgent = "x-user-agent"
	// MetaGatewayProof menyimpan tanda tangan HMAC-SHA256 berwaktu untuk membuktikan header berasal dari gateway tepercaya.
	MetaGatewayProof = "x-ig-gw-proof"

	maxProofSkewSeconds = 60
)

var (
	// ErrInvalidGatewayProof menandakan header identitas internal tidak memiliki tanda tangan gateway yang valid.
	ErrInvalidGatewayProof = errors.New("invalid internal gateway proof")
	// ErrExpiredGatewayProof menandakan timestamp tanda tangan gateway melewati jendela toleransi.
	ErrExpiredGatewayProof = errors.New("expired internal gateway proof")
)

type callerContextKey struct{}

// CallerContext merepresentasikan konteks identitas dan metadata permintaan yang terverifikasi di gateway.
type CallerContext struct {
	PrincipalID   string
	SessionID     string
	UserKind      string
	AuthEpoch     int64
	PermRevision  int64
	RequestID     string
	ClientIP      string
	UserAgent     string
	Authenticated bool
}

// WithCaller menyisipkan CallerContext ke dalam context.Context standar.
func WithCaller(ctx context.Context, caller CallerContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, callerContextKey{}, caller)
}

// FromContext mengambil CallerContext dari context.Context bila tersedia.
func FromContext(ctx context.Context) (CallerContext, bool) {
	if ctx == nil {
		return CallerContext{}, false
	}
	caller, ok := ctx.Value(callerContextKey{}).(CallerContext)
	if !ok {
		return CallerContext{}, false
	}
	return caller, true
}

// AppendToOutgoingGRPCContext menulis header metadata internal gRPC (beserta proof HMAC bila secret diberikan).
func AppendToOutgoingGRPCContext(ctx context.Context, caller CallerContext, hmacSecret []byte, now time.Time) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now()
	}

	pairs := make([]string, 0, 18)
	if caller.RequestID != "" {
		pairs = append(pairs, MetaRequestID, sanitizeHeaderValue(caller.RequestID))
	}
	if caller.ClientIP != "" {
		pairs = append(pairs, MetaClientIP, sanitizeHeaderValue(caller.ClientIP))
	}
	if caller.UserAgent != "" {
		pairs = append(pairs, MetaUserAgent, sanitizeHeaderValue(caller.UserAgent))
	}
	if caller.Authenticated && caller.PrincipalID != "" {
		pairs = append(pairs,
			MetaPrincipalID, sanitizeHeaderValue(caller.PrincipalID),
			MetaSessionID, sanitizeHeaderValue(caller.SessionID),
			MetaUserKind, sanitizeHeaderValue(caller.UserKind),
			MetaAuthEpoch, strconv.FormatInt(caller.AuthEpoch, 10),
			MetaPermRevision, strconv.FormatInt(caller.PermRevision, 10),
		)
	}
	if len(hmacSecret) > 0 {
		pairs = append(pairs, MetaGatewayProof, SignGatewayProof(caller, hmacSecret, now))
	}
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

// ExtractFromIncomingGRPCContext membaca metadata gRPC masuk dan memvalidasi tanda tangan gateway jika hmacSecret aktif.
func ExtractFromIncomingGRPCContext(ctx context.Context, hmacSecret []byte, now time.Time) (CallerContext, bool, error) {
	if ctx == nil {
		return CallerContext{}, false, nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return CallerContext{}, false, nil
	}

	var caller CallerContext
	caller.RequestID = firstMeta(md, MetaRequestID)
	rawClientIP := firstMeta(md, MetaClientIP)
	// Pertahankan nilai ClientIP mentah selama verifikasi HMAC karena penandatanganan Gateway memakai nilai sebelum normalisasi hop pertama.
	caller.ClientIP = rawClientIP
	caller.UserAgent = firstMeta(md, MetaUserAgent)
	caller.PrincipalID = firstMeta(md, MetaPrincipalID)
	caller.SessionID = firstMeta(md, MetaSessionID)
	caller.UserKind = firstMeta(md, MetaUserKind)
	rawAuthEpoch := firstMeta(md, MetaAuthEpoch)
	rawPermRevision := firstMeta(md, MetaPermRevision)
	caller.AuthEpoch, _ = strconv.ParseInt(rawAuthEpoch, 10, 64)
	caller.PermRevision, _ = strconv.ParseInt(rawPermRevision, 10, 64)
	caller.Authenticated = caller.PrincipalID != "" && caller.SessionID != ""

	hasIdentityMeta := caller.PrincipalID != "" || caller.SessionID != "" || caller.UserKind != "" || rawAuthEpoch != "" || rawPermRevision != ""
	hasMetadata := caller.Authenticated || hasIdentityMeta || caller.RequestID != "" || rawClientIP != "" || caller.UserAgent != ""
	proof := firstMeta(md, MetaGatewayProof)
	if len(hmacSecret) > 0 {
		if hasMetadata || proof != "" {
			if err := VerifyGatewayProof(proof, caller, hmacSecret, now); err != nil {
				return CallerContext{}, false, err
			}
		}
	}
	caller.ClientIP = firstClientIP(rawClientIP)

	hasAny := hasMetadata || proof != ""
	return caller, hasAny, nil
}

// SignGatewayProof membentuk string proof berformat "<unix_seconds>.<hmac_hex>".
func SignGatewayProof(caller CallerContext, hmacSecret []byte, now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	ts := now.Unix()
	payload := canonicalProofPayload(caller, ts)
	mac := hmac.New(sha256.New, hmacSecret)
	mac.Write([]byte(payload))
	return strconv.FormatInt(ts, 10) + "." + hex.EncodeToString(mac.Sum(nil))
}

// VerifyGatewayProof memeriksa validitas HMAC dan jendela replay proof gateway.
func VerifyGatewayProof(proof string, caller CallerContext, hmacSecret []byte, now time.Time) error {
	if len(hmacSecret) == 0 {
		return nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	dot := strings.IndexByte(proof, '.')
	if dot <= 0 || dot >= len(proof)-1 {
		return ErrInvalidGatewayProof
	}
	ts, err := strconv.ParseInt(proof[:dot], 10, 64)
	if err != nil {
		return ErrInvalidGatewayProof
	}
	diff := now.Unix() - ts
	if diff < 0 {
		diff = -diff
	}
	if diff > maxProofSkewSeconds {
		return fmt.Errorf("%w: skew=%ds", ErrExpiredGatewayProof, diff)
	}

	providedMAC, err := hex.DecodeString(proof[dot+1:])
	if err != nil || len(providedMAC) != sha256.Size {
		return ErrInvalidGatewayProof
	}
	expected := hmac.New(sha256.New, hmacSecret)
	expected.Write([]byte(canonicalProofPayload(caller, ts)))
	if !hmac.Equal(providedMAC, expected.Sum(nil)) {
		return ErrInvalidGatewayProof
	}
	return nil
}

func canonicalProofPayload(caller CallerContext, ts int64) string {
	return strings.Join([]string{
		strconv.FormatInt(ts, 10),
		sanitizeHeaderValue(caller.PrincipalID),
		sanitizeHeaderValue(caller.SessionID),
		sanitizeHeaderValue(caller.UserKind),
		strconv.FormatInt(caller.AuthEpoch, 10),
		strconv.FormatInt(caller.PermRevision, 10),
		sanitizeHeaderValue(caller.RequestID),
		sanitizeHeaderValue(caller.ClientIP),
	}, "|")
}

func firstMeta(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return strings.TrimSpace(vals[0])
}

func firstClientIP(raw string) string {
	if raw == "" {
		return ""
	}
	if idx := strings.IndexByte(raw, ','); idx >= 0 {
		return strings.TrimSpace(raw[:idx])
	}
	return strings.TrimSpace(raw)
}

func sanitizeHeaderValue(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "\r", "")
	v = strings.ReplaceAll(v, "\n", "")
	return v
}
