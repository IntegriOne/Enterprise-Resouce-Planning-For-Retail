package authsession

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authz"
	"github.com/redis/go-redis/v9"
)

func TestKeyspaceAndHashing(t *testing.T) {
	t.Parallel()

	ks := NewKeyspace("Development:Local*1")
	if ks.Env() != "development_local_1" {
		t.Fatalf("unexpected sanitized env: %q", ks.Env())
	}

	h := HashOpaqueToken("sample-cookie-secret")
	if len(h) != 64 {
		t.Fatalf("expected 64-char sha256 hex digest, got %d", len(h))
	}
	if got, want := ks.SessionTokenKey(h), "ig:development_local_1:iam:sess:tok:"+h; got != want {
		t.Fatalf("SessionTokenKey() = %q, want %q", got, want)
	}
	if got, want := ks.SessionActivityKey(h), "ig:development_local_1:iam:sess:act:"+h; got != want {
		t.Fatalf("SessionActivityKey() = %q, want %q", got, want)
	}
	if got, want := ks.PreAuthKey(h), "ig:development_local_1:gw:preauth:"+h; got != want {
		t.Fatalf("PreAuthKey() = %q, want %q", got, want)
	}
}

func TestBuildTLSConfigRequiresCAWhenStrict(t *testing.T) {
	t.Parallel()

	if _, err := NewTLSClient(RedisConfig{
		Addr:     "127.0.0.1:6379",
		Username: "gw_runtime",
		Password: "secret",
	}); err == nil {
		t.Fatal("expected error when CACertPath is missing in strict TLS mode")
	}
}

func TestReadSessionContextValidationDefaults(t *testing.T) {
	t.Parallel()

	if authz.DefaultSnapshotMaxAge != 30*time.Second {
		t.Fatalf("expected 30s snapshot ceiling, got %s", authz.DefaultSnapshotMaxAge)
	}
}

type userSecurityReadHook struct {
	read redis.ProcessHook
}

func (h userSecurityReadHook) ProcessHook(redis.ProcessHook) redis.ProcessHook { return h.read }
func (h userSecurityReadHook) DialHook(next redis.DialHook) redis.DialHook     { return next }
func (h userSecurityReadHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestReadUserSecurityRejectsStaleProjection(t *testing.T) {
	t.Parallel()
	const userID = "00000000-0000-4000-8000-000000000001"
	const open = `{"open":true,"recovery_epoch":4}`
	const current = `{"user_id":"00000000-0000-4000-8000-000000000001","status":"disabled","recovery_epoch":4,"auth_epoch":5}`
	for _, tc := range []struct {
		name           string
		gate, security any
		want           error
	}{
		{"current epoch", open, current, nil},
		{"closed gate", `{"open":false,"recovery_epoch":4}`, current, ErrSecurityGateClosed},
		{"stale recovery epoch", `{"open":true,"recovery_epoch":5}`, current, ErrUserSecurityBlocked},
		{"missing projection", open, nil, ErrUserSecurityBlocked},
		{"wrong user", open, `{"user_id":"other","status":"disabled","recovery_epoch":4}`, ErrUserSecurityBlocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			keyspace := NewKeyspace("test")
			client := redis.NewClient(&redis.Options{Addr: "unused:6379"})
			t.Cleanup(func() { _ = client.Close() })
			client.AddHook(userSecurityReadHook{read: func(_ context.Context, cmd redis.Cmder) error {
				read, ok := cmd.(*redis.SliceCmd)
				args := cmd.Args()
				if !ok || len(args) != 3 || args[0] != "mget" || args[1] != keyspace.GateKey() || args[2] != keyspace.UserSecurityKey(userID) {
					return errors.New("expected one atomic gate and user security read")
				}
				read.SetVal([]any{tc.gate, tc.security})
				return nil
			}})
			store, err := NewStore(client, keyspace)
			if err != nil {
				t.Fatal(err)
			}
			record, err := store.ReadUserSecurity(t.Context(), userID)
			if !errors.Is(err, tc.want) {
				t.Fatalf("read security: %v, want %v", err, tc.want)
			}
			if tc.want == nil && (record == nil || record.AuthEpoch != 5 || record.Status != "disabled") {
				t.Fatalf("unexpected security record: %+v", record)
			}
			if tc.want != nil && record != nil {
				t.Fatal("invalid projection returned a security record")
			}
		})
	}
}
