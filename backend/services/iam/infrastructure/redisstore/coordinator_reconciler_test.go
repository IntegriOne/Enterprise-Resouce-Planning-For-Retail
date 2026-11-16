package redisstore_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/db"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

func loadLivePGAndRedisTestEnv(t *testing.T) (string, authsession.RedisConfig) {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("IAM_TEST_DATABASE_DSN"))
	redisAddr := strings.TrimSpace(os.Getenv("IAM_TEST_REDIS_ADDR"))
	if dsn == "" || redisAddr == "" {
		t.Skip("skipping live PostgreSQL+Redis test: set IAM_TEST_DATABASE_DSN and IAM_TEST_REDIS_ADDR to run")
	}

	userName := strings.TrimSpace(os.Getenv("IAM_TEST_REDIS_USERNAME"))
	if userName == "" {
		userName = "iam_writer"
	}
	serverName := strings.TrimSpace(os.Getenv("IAM_TEST_REDIS_SERVER_NAME"))
	if serverName == "" {
		serverName = "redis"
	}

	cfg := authsession.RedisConfig{
		Addr:         redisAddr,
		Username:     userName,
		Password:     os.Getenv("IAM_TEST_REDIS_PASSWORD"),
		CACertPath:   strings.TrimSpace(os.Getenv("IAM_TEST_REDIS_CA_CERT_PATH")),
		ServerName:   serverName,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     16,
		MinIdleConns: 2,
	}
	return dsn, cfg
}

func TestLivePGAndTLSRedis_OutboxBootstrapAndOutageRecovery(t *testing.T) {
	dsn, redisCfg := loadLivePGAndRedisTestEnv(t)
	if os.Getenv("IAM_TEST_DATABASE_DISPOSABLE") != "1" {
		t.Skip("live integration test requires an explicitly disposable IAM test database")
	}
	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	gormDB, err := db.Init(ctx, dsn, true)
	if err != nil {
		t.Fatalf("db.Init(test) error = %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("gormDB.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	var databaseName string
	if err := sqlDB.QueryRowContext(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("identify disposable test database: %v", err)
	}
	if !strings.HasPrefix(databaseName, "iam_disposable_") || !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("refusing to truncate non-disposable database %q", databaseName)
	}

	// Bersihkan tabel domain pada DB _test saja sambil mempertahankan baris kontrol id=1 pada iam_security_control.
	if err := gormDB.Exec(`TRUNCATE TABLE iam_outbox_events, user_roles, group_roles, group_members, role_permissions, sessions, user_tokens, user_recovery_codes, user_totp_factors, user_credentials, user_identities, users, roles, permissions, groups RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("truncate iam_test tables error = %v", err)
	}

	redisClient, err := authsession.NewTLSClient(redisCfg)
	if err != nil {
		t.Fatalf("authsession.NewTLSClient() error = %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })

	testEnv := fmt.Sprintf("test_iam_%d", time.Now().UnixNano())
	ks := authsession.NewKeyspace(testEnv)
	store, err := authsession.NewStore(redisClient, ks)
	if err != nil {
		t.Fatalf("authsession.NewStore() error = %v", err)
	}
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = store.DeleteTestNamespace(cleanCtx)
	})

	repos, err := repository.New(gormDB)
	if err != nil {
		t.Fatalf("repository.New() error = %v", err)
	}

	secCfg := config.SecurityConfig{
		TOTPIssuer:            "IntegriOne ERP",
		TOTPActiveKeyVersion:  1,
		AllowEphemeralDevKey:  true,
		CSRFActiveKeyVersion:  1,
		AllowEphemeralCSRFKey: true,
		MaxFailedAttempts:     5,
		LockDuration:          15 * time.Minute,
	}

	reg, err := services.NewRegistryFromConfigWithRuntime(repos, secCfg, redisClient, testEnv, logger)
	if err != nil {
		t.Fatalf("NewRegistryFromConfigWithRuntime() error = %v", err)
	}

	// 1. Uji idempotensi BootstrapSystemPrincipalAndRBAC sebanyak dua kali pada PostgreSQL nyata.
	sysCfg := config.SystemBootstrapConfig{
		UserUID:     "00000000-0000-4000-8000-000000000001",
		Username:    "systemiam",
		Email:       "systemiam@integrione.local",
		Password:    "SysAdm#LiveTest2026!Z",
		DisplayName: "SystemIAM",
		UserKind:    "system",
		RoleCode:    "superadmin",
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, reg, sysCfg, logger); err != nil {
		t.Fatalf("BootstrapSystemPrincipalAndRBAC(1) error = %v", err)
	}
	if err := services.BootstrapSystemPrincipalAndRBAC(ctx, repos, reg, sysCfg, logger); err != nil {
		t.Fatalf("BootstrapSystemPrincipalAndRBAC(2) error = %v", err)
	}
	if err := reg.Reconciler.BootstrapGateAndDrain(ctx); err != nil {
		t.Fatalf("BootstrapGateAndDrain(initial) error = %v", err)
	}

	// 2. Uji login systemiam -> penerbitan sesi -> outbox langsung effective dan terproyeksi ke Redis TLS.
	authRes, err := reg.Users.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "systemiam",
		Password:   sysCfg.Password,
		IPAddress:  "127.0.0.1",
		UserAgent:  "live-pg-redis-test",
	})
	if err != nil || authRes.User == nil || authRes.IssuedSession == nil {
		t.Fatalf("AuthenticatePassword(systemiam) error = (%+v, %v)", authRes, err)
	}
	if authRes.RedirectPath != "/identity" {
		t.Fatalf("expected /identity redirect for systemiam, got %q", authRes.RedirectPath)
	}

	sum := sha256.Sum256([]byte(authRes.IssuedSession.SessionToken))
	tokHashHex := hex.EncodeToString(sum[:])
	sessCheck, err := store.ReadSessionContext(ctx, tokHashHex, time.Now())
	if err != nil {
		t.Fatalf("store.ReadSessionContext() after live login error = %v", err)
	}
	if !sessCheck.Gate.Open || sessCheck.Token == nil || sessCheck.Token.UserID != authRes.User.ID {
		t.Fatalf("unexpected live Redis session context: %+v", sessCheck)
	}
	if pending, err := repos.SecurityRuntime.CountPending(ctx); err != nil || pending != 0 {
		t.Fatalf("expected 0 pending outbox after healthy commit, got pending=%d err=%v", pending, err)
	}

	// 3. Simulasi outage Redis saat mutasi terjadi: buat pelanggan baru dan terbitkan sesi memakai klien Redis yang gagal.
	downRedisCfg := redisCfg
	downRedisCfg.Addr = "127.0.0.1:19998"
	downRedisCfg.DialTimeout = 100 * time.Millisecond
	downRedisCfg.ReadTimeout = 100 * time.Millisecond
	downRedisCfg.WriteTimeout = 100 * time.Millisecond
	downClient, err := authsession.NewTLSClient(downRedisCfg)
	if err != nil {
		t.Fatalf("authsession.NewTLSClient(down) error = %v", err)
	}
	t.Cleanup(func() { _ = downClient.Close() })

	outageReg, err := services.NewRegistryFromConfigWithRuntime(repos, secCfg, downClient, testEnv, logger)
	if err != nil {
		t.Fatalf("NewRegistryFromConfigWithRuntime(outage) error = %v", err)
	}
	regCust, err := outageReg.Users.CreateUser(ctx, user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		DisplayName: "Buyer Live",
		Email:       "buyer-live@integrione.local",
		Username:    "buyerlive",
		Password:    "VaultRandom#2026!Z",
	})
	if err != nil {
		t.Fatalf("outageReg.Users.CreateUser() error = %v", err)
	}
	resendCust, err := outageReg.Users.ResendValidationEmailToken(ctx, user.ResendValidationEmailInput{
		Email: "buyer-live@integrione.local",
	})
	if err != nil || resendCust.VerificationToken == "" {
		t.Fatalf("outageReg.Users.ResendValidationEmailToken() error = (%+v, %v)", resendCust, err)
	}
	if _, err := outageReg.Users.ConfirmEmailVerificationByToken(ctx, resendCust.VerificationToken); err != nil {
		t.Fatalf("outageReg.Users.ConfirmEmailVerificationByToken() error = %v", err)
	}
	custAuth, err := outageReg.Users.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: "buyerlive",
		Password:   "VaultRandom#2026!Z",
		IPAddress:  "127.0.0.1",
		UserAgent:  "outage-client",
	})
	if err != nil || custAuth.IssuedSession == nil {
		t.Fatalf("outageReg.Users.AuthenticatePassword() error = (%+v, %v)", custAuth, err)
	}
	if custAuth.RedirectPath != "/" {
		t.Fatalf("expected '/' redirect for customer, got %q", custAuth.RedirectPath)
	}

	pendingDuringOutage, err := repos.SecurityRuntime.CountPending(ctx)
	if err != nil || pendingDuringOutage == 0 {
		t.Fatalf("expected pending outbox events > 0 during Redis outage, got %d (err=%v)", pendingDuringOutage, err)
	}

	// Bersihkan namespace Redis seakan Redis baru restart kosong, lalu pulihkan dengan Reconciler sehat.
	if err := store.DeleteTestNamespace(ctx); err != nil {
		t.Fatalf("DeleteTestNamespace() error = %v", err)
	}
	prevEpoch := sessCheck.Gate.RecoveryEpoch

	if err := reg.Reconciler.BootstrapGateAndDrain(ctx); err != nil {
		t.Fatalf("BootstrapGateAndDrain(recovery) error = %v", err)
	}

	pendingAfterRecover, err := repos.SecurityRuntime.CountPending(ctx)
	if err != nil || pendingAfterRecover != 0 {
		t.Fatalf("expected 0 pending outbox after recovery, got %d (err=%v)", pendingAfterRecover, err)
	}

	custSum := sha256.Sum256([]byte(custAuth.IssuedSession.SessionToken))
	custRes, err := store.ReadSessionContext(ctx, hex.EncodeToString(custSum[:]), time.Now())
	if err != nil {
		t.Fatalf("store.ReadSessionContext(customer after recovery) error = %v", err)
	}
	if !custRes.Gate.Open || custRes.Gate.RecoveryEpoch <= prevEpoch {
		t.Fatalf("expected reopened gate with bumped recovery_epoch > %d, got %+v", prevEpoch, custRes.Gate)
	}
	if custRes.Token == nil || custRes.Token.UserID != regCust.ID {
		t.Fatalf("expected customer session hydrated after recovery, got %+v", custRes.Token)
	}
	if _, err := reg.Sessions.RefreshAccessSnapshot(ctx, authRes.User.ID); err != nil {
		t.Fatalf("RefreshAccessSnapshot(systemiam) error = %v", err)
	}
	_ = entity.UserKindCustomer
}
