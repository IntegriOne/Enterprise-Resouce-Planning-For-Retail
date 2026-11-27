//go:build integration

package test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// workerSessionReader memakai Redis disposable bila alamatnya diberikan secara eksplisit.
func workerSessionReader(t *testing.T, ctx context.Context, projection workerSessionProjection) security.SessionReader {
	t.Helper()
	addr := os.Getenv("HRM_TEST_DISPOSABLE_REDIS_ADDR")
	if addr == "" {
		return projection
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("integration test requires a loopback Redis endpoint")
	}
	client, err := authsession.NewTLSClient(authsession.RedisConfig{
		Addr: addr, Username: "default", Password: os.Getenv("HRM_TEST_DISPOSABLE_REDIS_PASSWORD"),
		CACertPath: os.Getenv("HRM_TEST_DISPOSABLE_REDIS_CA"), ServerName: "localhost",
	})
	if err != nil {
		t.Fatal("configure disposable Redis TLS client:", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error("close Redis test client:", err)
		}
	})
	id, err := (operation.UUIDGenerator{}).NewID()
	if err != nil {
		t.Fatal(err)
	}
	store, err := authsession.NewStore(client, authsession.NewKeyspace("test_hrm_"+id))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := store.DeleteTestNamespace(cleanupCtx); err != nil {
			t.Error("clean Redis test namespace:", err)
		}
	})
	now := time.Now()
	state, err := projection.ReadSessionContext(ctx, scopedWorkerTokenHash, now)
	if err != nil {
		t.Fatal(err)
	}
	state.Token.TokenHashHex = scopedWorkerTokenHash
	state.Token.Generation = 1
	state.Token.AuthEpoch = 1
	state.Token.IssuedAtUnix = now.Unix()
	state.Token.AbsExpiresAtUnix = now.Add(time.Hour).Unix()
	state.Security.Status = authz.StatusActive
	activity := authsession.SessionActivityRecord{
		UserID: scopedWorkerPrincipal, SessionID: scopedWorkerSession,
		Generation: 1, AuthEpoch: 1, LastSeenUnix: now.Unix(), IdleExpiresUnix: now.Add(time.Minute).Unix(),
	}
	if _, err := store.SetGate(ctx, true, 1, now, "hrm_integration_test"); err != nil {
		t.Fatal("prepare Redis test gate:", err)
	}
	if err := store.WriteSession(ctx, *state.Token, activity, *state.Security, now); err != nil {
		t.Fatal("prepare Redis test session:", err)
	}
	if err := store.ProjectAccessSnapshot(ctx, state.Snapshot, 0, now); err != nil {
		t.Fatal("prepare Redis test grant:", err)
	}
	t.Log("using disposable Redis over TLS")
	return store
}
