package authsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authz"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrRedisUnavailable menandakan kegagalan transport atau koneksi Redis (fail-closed).
	ErrRedisUnavailable = errors.New("security redis unavailable")
	// ErrSecurityGateClosed menandakan gerbang keamanan IAM di Redis belum terbuka atau sedang sinkronisasi ulang.
	ErrSecurityGateClosed = errors.New("security gate is closed")
	// ErrSessionNotFound menandakan sesi tidak ditemukan di Redis.
	ErrSessionNotFound = errors.New("session not found in redis")
	// ErrSessionRevoked menandakan sesi sudah dicabut atau ditimpa auth_epoch yang lebih tinggi.
	ErrSessionRevoked = errors.New("session revoked")
	// ErrSessionExpired menandakan batas waktu absolut atau idle sesi sudah terlewati.
	ErrSessionExpired = errors.New("session expired")
	// ErrSessionNotActive menandakan TouchActivity gagal karena sesi sudah dihapus atau berubah generasi.
	ErrSessionNotActive = errors.New("session activity record missing or stale")
	// ErrUserSecurityBlocked menandakan user tidak berstatus aktif atau sedang terkunci.
	ErrUserSecurityBlocked = errors.New("user security state is blocked")
	// ErrPreAuthNotFound menandakan token pra-autentikasi tidak ditemukan atau kedaluwarsa.
	ErrPreAuthNotFound = errors.New("pre-auth token not found")
	// ErrRateLimitExceeded menandakan permintaan melebihi ambang rate limit sementara.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// Store membungkus operasi Redis terstandar untuk IAM (writer) dan Gateway (runtime reader + activity/gw writer).
type Store struct {
	client *redis.Client
	ks     Keyspace
}

// NewStore membentuk Store dengan validasi klien dan keyspace.
func NewStore(client *redis.Client, ks Keyspace) (*Store, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	return &Store{
		client: client,
		ks:     ks,
	}, nil
}

// Keyspace mengembalikan keyspace environment milik store ini.
func (s *Store) Keyspace() Keyspace {
	return s.ks
}

// SetGate memperbarui status gerbang keamanan IAM secara monotonik menurut recovery_epoch.
func (s *Store) SetGate(ctx context.Context, open bool, recoveryEpoch int64, now time.Time, source string) (int64, error) {
	if now.IsZero() {
		now = time.Now()
	}
	openFlag := "0"
	if open {
		openFlag = "1"
	}
	res, err := scriptSetGate.Run(
		ctx,
		s.client,
		[]string{s.ks.GateKey(), s.ks.RecoveryEpochKey()},
		openFlag,
		recoveryEpoch,
		now.Unix(),
		strings.TrimSpace(source),
	).Int64()
	if err != nil {
		return 0, wrapRedisErr(err)
	}
	return res, nil
}

// GetGate membaca status gerbang keamanan IAM terkini.
func (s *Store) GetGate(ctx context.Context) (*SecurityGateRecord, error) {
	raw, err := s.client.Get(ctx, s.ks.GateKey()).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSecurityGateClosed
		}
		return nil, wrapRedisErr(err)
	}
	var gate SecurityGateRecord
	if err := json.Unmarshal(raw, &gate); err != nil {
		return nil, ErrSecurityGateClosed
	}
	return &gate, nil
}

// WriteSession memproyeksikan pasangan token sesi, jendela idle aktivitas, dan state keamanan user ke Redis secara atomik.
func (s *Store) WriteSession(
	ctx context.Context,
	tok SessionTokenRecord,
	act SessionActivityRecord,
	sec UserSecurityRecord,
	now time.Time,
) error {
	if now.IsZero() {
		now = time.Now()
	}
	absTTL := maxTTLSeconds(tok.AbsExpiresAtUnix, now.Unix())
	idleTTL := maxTTLSeconds(act.IdleExpiresUnix, now.Unix())
	if absTTL <= 0 || idleTTL <= 0 {
		return ErrSessionExpired
	}
	if idleTTL > absTTL {
		idleTTL = absTTL
	}

	tokJSON, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	actJSON, err := json.Marshal(act)
	if err != nil {
		return err
	}

	err = scriptWriteSessionPair.Run(
		ctx,
		s.client,
		[]string{
			s.ks.SessionTokenKey(tok.TokenHashHex),
			s.ks.SessionActivityKey(tok.TokenHashHex),
			s.ks.UserSecurityKey(tok.UserID),
		},
		string(tokJSON),
		absTTL,
		string(actJSON),
		idleTTL,
		sec.UserID,
		sec.Status,
		sec.LockedUntilUnix,
		sec.AuthEpoch,
		sec.PermRevision,
		sec.RecoveryEpoch,
		now.Unix(),
	).Err()
	return wrapRedisErr(err)
}

// RevokeSessionTokens menghapus daftar sesi dari Redis berdasarkan digest token (dan memperbarui state keamanan user bila diberikan).
func (s *Store) RevokeSessionTokens(
	ctx context.Context,
	tokenHashes []string,
	updatedSec *UserSecurityRecord,
	invalidateAuthz bool,
	now time.Time,
) error {
	if now.IsZero() {
		now = time.Now()
	}

	keys := make([]string, 0, len(tokenHashes)*2)
	for _, h := range tokenHashes {
		clean := strings.ToLower(strings.TrimSpace(h))
		if clean == "" {
			continue
		}
		keys = append(keys, s.ks.SessionTokenKey(clean), s.ks.SessionActivityKey(clean))
	}

	if updatedSec != nil {
		if _, err := s.BumpUserSecurity(ctx, *updatedSec, invalidateAuthz, now); err != nil {
			return err
		}
	}
	if len(keys) == 0 {
		return nil
	}
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return wrapRedisErr(err)
	}
	return nil
}

// BumpUserSecurity memutakhirkan user:sec secara monotonik (dan opsional menghapus snapshot lama).
func (s *Store) BumpUserSecurity(
	ctx context.Context,
	sec UserSecurityRecord,
	invalidateAuthz bool,
	now time.Time,
) ([]int64, error) {
	if now.IsZero() {
		now = time.Now()
	}
	invFlag := "0"
	if invalidateAuthz {
		invFlag = "1"
	}
	res, err := scriptBumpUserSecurity.Run(
		ctx,
		s.client,
		[]string{s.ks.UserSecurityKey(sec.UserID), s.ks.AuthzSnapshotKey(sec.UserID)},
		sec.UserID,
		sec.Status,
		sec.LockedUntilUnix,
		sec.AuthEpoch,
		sec.PermRevision,
		sec.RecoveryEpoch,
		now.Unix(),
		invFlag,
	).Int64Slice()
	if err != nil {
		return nil, wrapRedisErr(err)
	}
	return res, nil
}

// ProjectAccessSnapshot menyimpan AccessSnapshot beserta UserSecurityRecord terkini oleh iam_writer.
func (s *Store) ProjectAccessSnapshot(
	ctx context.Context,
	snap *authz.AccessSnapshot,
	ttl time.Duration,
	now time.Time,
) error {
	if snap == nil || snap.PrincipalID == "" {
		return errors.New("access snapshot with principal_id is required")
	}
	if now.IsZero() {
		now = time.Now()
	}
	if ttl <= 0 {
		ttl = authz.DefaultSnapshotMaxAge + 5*time.Second
	}

	rawSnap, err := authz.MarshalSnapshot(snap)
	if err != nil {
		return err
	}

	sec := UserSecurityRecord{
		UserID:        snap.PrincipalID,
		Status:        snap.Status,
		AuthEpoch:     snap.AuthEpoch,
		PermRevision:  snap.PermRevision,
		RecoveryEpoch: snap.RecoveryEpoch,
		UpdatedAtUnix: now.Unix(),
	}
	if _, err := s.BumpUserSecurity(ctx, sec, false, now); err != nil {
		return err
	}
	if err := s.client.Set(ctx, s.ks.AuthzSnapshotKey(snap.PrincipalID), rawSnap, ttl).Err(); err != nil {
		return wrapRedisErr(err)
	}
	return nil
}

// ReadSessionContext membaca seluruh state keamanan sesi dari Redis secara fail-closed dalam dua MGET ringkas.
func (s *Store) ReadSessionContext(ctx context.Context, tokenHashHex string, now time.Time) (*SessionValidationContext, error) {
	if now.IsZero() {
		now = time.Now()
	}
	cleanHash := strings.ToLower(strings.TrimSpace(tokenHashHex))
	if cleanHash == "" {
		return nil, ErrSessionNotFound
	}

	firstBatch, err := s.client.MGet(
		ctx,
		s.ks.GateKey(),
		s.ks.SessionTokenKey(cleanHash),
		s.ks.SessionActivityKey(cleanHash),
	).Result()
	if err != nil {
		return nil, wrapRedisErr(err)
	}
	if len(firstBatch) != 3 {
		return nil, ErrRedisUnavailable
	}

	gate, err := decodeGate(firstBatch[0])
	if err != nil {
		return nil, err
	}
	if !gate.Open {
		return nil, ErrSecurityGateClosed
	}

	tok, err := decodeRecord[SessionTokenRecord](firstBatch[1], ErrSessionNotFound)
	if err != nil {
		return nil, err
	}
	if tok.Revoked {
		return nil, ErrSessionRevoked
	}
	nowUnix := now.Unix()
	if tok.AbsExpiresAtUnix <= nowUnix {
		return nil, ErrSessionExpired
	}

	act, err := decodeRecord[SessionActivityRecord](firstBatch[2], ErrSessionExpired)
	if err != nil {
		return nil, err
	}
	if act.IdleExpiresUnix <= nowUnix {
		return nil, ErrSessionExpired
	}
	if act.UserID != tok.UserID || act.Generation != tok.Generation || act.AuthEpoch != tok.AuthEpoch {
		return nil, ErrSessionRevoked
	}

	secondBatch, err := s.client.MGet(
		ctx,
		s.ks.UserSecurityKey(tok.UserID),
		s.ks.AuthzSnapshotKey(tok.UserID),
	).Result()
	if err != nil {
		return nil, wrapRedisErr(err)
	}
	if len(secondBatch) != 2 {
		return nil, ErrRedisUnavailable
	}

	var sec *UserSecurityRecord
	if secondBatch[0] != nil {
		sec, err = decodeRecord[UserSecurityRecord](secondBatch[0], ErrUserSecurityBlocked)
		if err != nil {
			return nil, ErrUserSecurityBlocked
		}
		if strings.ToLower(strings.TrimSpace(sec.Status)) != authz.StatusActive {
			return nil, ErrUserSecurityBlocked
		}
		if sec.LockedUntilUnix > nowUnix {
			return nil, ErrUserSecurityBlocked
		}
		if tok.AuthEpoch < sec.AuthEpoch {
			return nil, ErrSessionRevoked
		}
	}

	var snap *authz.AccessSnapshot
	if secondBatch[1] != nil {
		if rawStr, ok := asString(secondBatch[1]); ok {
			if parsed, parseErr := authz.UnmarshalSnapshot([]byte(rawStr)); parseErr == nil {
				snap = parsed
			}
		}
	}

	return &SessionValidationContext{
		Gate:     gate,
		Token:    tok,
		Activity: act,
		Security: sec,
		Snapshot: snap,
	}, nil
}

// TouchActivity memperpanjang idle expiry sesi melalui Lua XX yang hanya memodifikasi kunci sess:act:* milik Gateway.
func (s *Store) TouchActivity(
	ctx context.Context,
	tokenHashHex string,
	tok SessionTokenRecord,
	idleTimeout time.Duration,
	now time.Time,
) error {
	if now.IsZero() {
		now = time.Now()
	}
	if idleTimeout <= 0 {
		return ErrSessionExpired
	}
	nowUnix := now.Unix()
	absRemaining := tok.AbsExpiresAtUnix - nowUnix
	if absRemaining <= 0 {
		return ErrSessionExpired
	}

	newIdleExp := now.Add(idleTimeout).Unix()
	if newIdleExp > tok.AbsExpiresAtUnix {
		newIdleExp = tok.AbsExpiresAtUnix
	}
	ttlSec := newIdleExp - nowUnix
	if ttlSec <= 0 {
		ttlSec = 1
	}

	res, err := scriptTouchActivity.Run(
		ctx,
		s.client,
		[]string{s.ks.SessionActivityKey(tokenHashHex)},
		tok.UserID,
		strconv.FormatInt(tok.Generation, 10),
		strconv.FormatInt(tok.AuthEpoch, 10),
		strconv.FormatInt(nowUnix, 10),
		strconv.FormatInt(newIdleExp, 10),
		strconv.FormatInt(ttlSec, 10),
	).Int64()
	if err != nil {
		return wrapRedisErr(err)
	}
	if res == 0 {
		return ErrSessionNotActive
	}
	return nil
}

// SetPreAuth menyimpan atau memperbarui token pra-autentikasi milik Gateway di namespace ig:*:gw:preauth:*.
func (s *Store) SetPreAuth(ctx context.Context, rec PreAuthRecord, ttl time.Duration) error {
	if rec.TokenHashHex == "" || rec.CSRFDigestHex == "" {
		return errors.New("pre-auth record hashes are required")
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if err := s.client.Set(ctx, s.ks.PreAuthKey(rec.TokenHashHex), raw, ttl).Err(); err != nil {
		return wrapRedisErr(err)
	}
	return nil
}

// GetPreAuth mengambil record pra-autentikasi milik Gateway berdasarkan hash cookie.
func (s *Store) GetPreAuth(ctx context.Context, cookieHashHex string, now time.Time) (*PreAuthRecord, error) {
	if now.IsZero() {
		now = time.Now()
	}
	raw, err := s.client.Get(ctx, s.ks.PreAuthKey(cookieHashHex)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPreAuthNotFound
		}
		return nil, wrapRedisErr(err)
	}
	var rec PreAuthRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, ErrPreAuthNotFound
	}
	if rec.ExpiresAtUnix > 0 && rec.ExpiresAtUnix <= now.Unix() {
		return nil, ErrPreAuthNotFound
	}
	return &rec, nil
}

// DeletePreAuth menghapus token pra-autentikasi setelah login sukses merotasi cookie menjadi sesi aktif.
func (s *Store) DeletePreAuth(ctx context.Context, cookieHashHex string) error {
	if cookieHashHex == "" {
		return nil
	}
	if err := s.client.Del(ctx, s.ks.PreAuthKey(cookieHashHex)).Err(); err != nil {
		return wrapRedisErr(err)
	}
	return nil
}

// CheckRateLimit menegakkan pembatas laju atomik per scope dan subjek di namespace ig:*:gw:rl:*.
func (s *Store) CheckRateLimit(ctx context.Context, scope, subject string, limit int64, window time.Duration) (int64, error) {
	if limit <= 0 || window <= 0 {
		return 0, nil
	}
	windowSec := int64(math.Ceil(window.Seconds()))
	if windowSec <= 0 {
		windowSec = 1
	}
	subHash := HashOpaqueToken(subject)
	res, err := scriptRateLimit.Run(
		ctx,
		s.client,
		[]string{s.ks.RateLimitKey(scope, subHash)},
		windowSec,
	).Int64Slice()
	if err != nil {
		return 0, wrapRedisErr(err)
	}
	if len(res) == 0 {
		return 0, ErrRedisUnavailable
	}
	if res[0] > limit {
		return res[0], fmt.Errorf("%w: count=%d limit=%d", ErrRateLimitExceeded, res[0], limit)
	}
	return res[0], nil
}

// DeleteTestNamespace membersihkan seluruh kunci milik namespace uji (khusus test_*, bukan FLUSHALL).
func (s *Store) DeleteTestNamespace(ctx context.Context) error {
	if !strings.HasPrefix(s.ks.Env(), "test") {
		return fmt.Errorf("refusing namespace cleanup on non-test env %q", s.ks.Env())
	}
	pattern := s.ks.Prefix() + "*"
	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return wrapRedisErr(err)
		}
		if len(keys) > 0 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return wrapRedisErr(err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			return nil
		}
	}
}

func decodeGate(raw any) (*SecurityGateRecord, error) {
	if raw == nil {
		return nil, ErrSecurityGateClosed
	}
	s, ok := asString(raw)
	if !ok {
		return nil, ErrSecurityGateClosed
	}
	var gate SecurityGateRecord
	if err := json.Unmarshal([]byte(s), &gate); err != nil {
		return nil, ErrSecurityGateClosed
	}
	return &gate, nil
}

func decodeRecord[T any](raw any, notFoundErr error) (*T, error) {
	if raw == nil {
		return nil, notFoundErr
	}
	s, ok := asString(raw)
	if !ok {
		return nil, notFoundErr
	}
	var out T
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, notFoundErr
	}
	return &out, nil
}

func asString(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case []byte:
		return string(val), true
	default:
		return "", false
	}
}

func maxTTLSeconds(expiresUnix, nowUnix int64) int64 {
	diff := expiresUnix - nowUnix
	if diff <= 0 {
		return 0
	}
	return diff
}

func wrapRedisErr(err error) error {
	if err == nil {
		return nil
	}
	if redis.HasErrorPrefix(err, "NOPERM") {
		return fmt.Errorf("redis ACL denied operation: %w", err)
	}
	return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
}
