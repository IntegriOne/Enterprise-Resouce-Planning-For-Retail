package config

import "testing"

func validTestConfig() *Config {
	cfg := &Config{
		AppEnv:                    "production",
		ServiceName:               "integrione-gateway",
		HTTPHost:                  "0.0.0.0",
		HTTPPort:                  8080,
		MaxComplexity:             250,
		EnableIntrospection:       false,
		AllowedOrigins:            []string{"https://erp.localhost:8443"},
		CookieName:                SessionCookieName,
		CookieSecure:              true,
		IAMGRPCTarget:             "iam:9001",
		InternalGatewayHMACSecret: "0123456789abcdef0123456789abcdef",
		RedisEnabled:              true,
		RedisEnv:                  "production",
	}
	cfg.Redis.Addr = "redis:6379"
	cfg.Redis.Password = "strong-redis-password"
	cfg.Redis.CACertPath = "/etc/ssl/redis/ca.crt"
	return cfg
}

func TestValidateRuntimeStrict(t *testing.T) {
	t.Run("konfigurasi valid lolos", func(t *testing.T) {
		if err := validTestConfig().ValidateRuntimeStrict(); err != nil {
			t.Fatalf("expected valid config, got: %v", err)
		}
	})

	t.Run("tolak cookie non-Host", func(t *testing.T) {
		c := validTestConfig()
		c.CookieName = "integrione_session"
		if err := c.ValidateRuntimeStrict(); err == nil {
			t.Fatal("expected error for non __Host- cookie name")
		}
	})

	t.Run("tolak cookie secure false", func(t *testing.T) {
		c := validTestConfig()
		c.CookieSecure = false
		if err := c.ValidateRuntimeStrict(); err == nil {
			t.Fatal("expected error when CookieSecure is false")
		}
	})

	t.Run("tolak hmac secret pendek", func(t *testing.T) {
		c := validTestConfig()
		c.InternalGatewayHMACSecret = "too-short"
		if err := c.ValidateRuntimeStrict(); err == nil {
			t.Fatal("expected error for short HMAC secret")
		}
	})

	t.Run("tolak redis tanpa tls ca verified", func(t *testing.T) {
		c := validTestConfig()
		c.Redis.CACertPath = ""
		if err := c.ValidateRuntimeStrict(); err == nil {
			t.Fatal("expected error when Redis CA cert is empty")
		}
		c2 := validTestConfig()
		c2.Redis.TLSInsecureSkipVerify = true
		if err := c2.ValidateRuntimeStrict(); err == nil {
			t.Fatal("expected error when TLSInsecureSkipVerify is true")
		}
	})

	t.Run("tolak introspection aktif di production", func(t *testing.T) {
		c := validTestConfig()
		c.EnableIntrospection = true
		if err := c.ValidateRuntimeStrict(); err == nil {
			t.Fatal("expected error when introspection enabled outside development")
		}
	})
}

func TestLoadStrictEnvRejectsWildcardOrigin(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("GATEWAY_ALLOWED_ORIGINS", "*")
	t.Setenv("GATEWAY_INTERNAL_HMAC_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("GATEWAY_REDIS_ADDR", "redis:6379")
	t.Setenv("GATEWAY_REDIS_PASSWORD", "secret")
	t.Setenv("GATEWAY_REDIS_CA_CERT_PATH", "/tmp/ca.crt")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load to reject wildcard origin")
	}
}
