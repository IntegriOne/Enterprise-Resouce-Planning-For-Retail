package config

import (
	"testing"
	"time"
)

func TestHumanResourceConfigIsOptional(t *testing.T) {
	setMinimalStrictGatewayEnv(t)
	t.Setenv("HRM_GRPC_ENABLED", "")
	t.Setenv("HRM_GRPC_TIMEOUT_SECONDS", "bad")
	cfg, err := Load()
	if err != nil || cfg.HRMGRPCEnabled || cfg.HRMGRPCConfigError != nil || cfg.HRMGRPCTimeout != 5*time.Second {
		t.Fatalf("disabled HRM should use the default timeout: %v", err)
	}
	t.Setenv("HRM_GRPC_ENABLED", "true")
	cfg, err = Load()
	if err != nil || !cfg.HRMGRPCEnabled || cfg.HRMGRPCConfigError == nil {
		t.Fatalf("invalid HRM timeout should degrade only HRM: %v", err)
	}
}
