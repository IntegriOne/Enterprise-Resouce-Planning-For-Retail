package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"log/slog"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWireIAMServices_WithGormRegistry(t *testing.T) {
	// Tidak menggunakan t.Parallel karena t.Setenv memengaruhi environment proses.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open(sqlite) error = %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Gagal tertutup jika Production mode tanpa IAM_TOTP_KEYRING.
	t.Setenv("IAM_TOTP_KEYRING", "")
	if _, err := wireIAMServices(db, &config.Config{AppEnv: "production"}, logger); err == nil {
		t.Fatal("expected fail-closed error in production without IAM_TOTP_KEYRING")
	}

	// Gagal tertutup jika Production mode tanpa IAM_CSRF_HMAC_KEYRING walaupun IAM_TOTP_KEYRING tersedia.
	spec := "1:" + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x99}, 32))
	t.Setenv("IAM_TOTP_KEYRING", spec)
	t.Setenv("IAM_TOTP_ACTIVE_KEY_VERSION", "1")
	t.Setenv("IAM_CSRF_HMAC_KEYRING", "")
	if _, err := wireIAMServices(db, &config.Config{AppEnv: "production"}, logger); err == nil {
		t.Fatal("expected fail-closed error in production without IAM_CSRF_HMAC_KEYRING")
	}

	// Berhasil dengan kedua keyring TOTP & CSRF terkonfigurasi lewat repository.New -> services.NewRegistryFromConfig.
	t.Setenv("IAM_CSRF_HMAC_KEYRING", spec)
	t.Setenv("IAM_CSRF_ACTIVE_KEY_VERSION", "1")

	reg, err := wireIAMServices(db, &config.Config{AppEnv: "production"}, logger)
	if err != nil {
		t.Fatalf("wireIAMServices(production with keyring) error = %v", err)
	}
	if err := reg.Validate(); err != nil {
		t.Fatalf("Registry.Validate() error = %v", err)
	}

	// Gagal tertutup jika gormDB nil.
	if _, err := wireIAMServices(nil, &config.Config{AppEnv: "development"}, logger); err == nil {
		t.Fatal("expected error when gormDB is nil")
	}
}

func TestWireGRPCServer_RegistersAllIAMServices(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open(sqlite) error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if _, err := wireGRPCServer(nil, nil, logger); err == nil {
		t.Fatal("expected fail-closed error when cfg is nil")
	}

	cfg := &config.Config{
		AppEnv:      "development",
		ServiceName: "integrione-iam",
	}
	reg, err := wireIAMServices(db, cfg, logger)
	if err != nil {
		t.Fatalf("wireIAMServices() error = %v", err)
	}

	if _, err := wireGRPCServer(cfg, nil, logger); err == nil {
		t.Fatal("expected fail-closed error when service registry is nil")
	}

	srv, err := wireGRPCServer(cfg, reg, logger)
	if err != nil {
		t.Fatalf("wireGRPCServer() error = %v", err)
	}

	info := srv.GRPCServer().GetServiceInfo()
	expectedServices := []string{
		"iam.v1.PermissionService",
		"iam.v1.RoleService",
		"iam.v1.GroupService",
		"iam.v1.SessionService",
		"iam.v1.UserService",
		"iam.v1.AuthService",
		"grpc.health.v1.Health",
	}
	for _, svcName := range expectedServices {
		if _, ok := info[svcName]; !ok {
			t.Fatalf("expected gRPC service %q to be registered, got %v", svcName, info)
		}
	}
}
