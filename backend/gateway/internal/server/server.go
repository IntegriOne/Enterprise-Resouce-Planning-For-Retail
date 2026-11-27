// Package server merangkai server HTTP Fiber v3 dan handler eksekusi gqlgen pada endpoint tunggal `/api/graphql`.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/config"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

const graphqlEndpointPath = "/api/graphql"

type clientIPContextKey struct{}

// Server membungkus instance Fiber v3 dan GraphQL handler Gateway.
type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	app    *fiber.App
	gqlSrv *handler.Server
}

// Dependencies menyimpan komponen Zona 3 yang diperlukan oleh server HTTP Gateway.
type Dependencies struct {
	Config          *config.Config
	Logger          *slog.Logger
	Schema          graphql.ExecutableSchema
	AuthManager     *middleware.AuthManager
	OriginValidator *browserauth.OriginValidator
}

// New membentuk Fiber v3 Server beserta handler tunggal `/api/graphql` dan proteksi keamanan lapisan transport.
func New(deps Dependencies) (*Server, error) {
	if deps.Config == nil {
		return nil, errors.New("gateway config is required")
	}
	if deps.Schema == nil {
		return nil, errors.New("executable GraphQL schema is required")
	}
	if deps.AuthManager == nil {
		return nil, errors.New("auth manager is required")
	}
	if deps.OriginValidator == nil {
		return nil, errors.New("origin validator is required")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	gqlSrv := handler.New(deps.Schema)
	gqlSrv.AddTransport(transport.GET{})
	gqlSrv.AddTransport(transport.POST{})
	gqlSrv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	gqlSrv.Use(extension.FixedComplexityLimit(deps.Config.MaxComplexity))
	if deps.Config.EnableIntrospection {
		gqlSrv.Use(extension.Introspection{})
	} else {
		gqlSrv.SetDisableSuggestion(true)
	}
	errorPresenter := buildErrorPresenter(deps.Logger)
	gqlSrv.SetErrorPresenter(errorPresenter)
	gqlSrv.SetRecoverFunc(buildRecoverFunc(deps.Logger))
	gqlSrv.AroundOperations(buildOperationGuard(
		deps.AuthManager,
		deps.OriginValidator,
		errorPresenter,
		deps.Config.EnableIntrospection,
		deps.Config.RateLimitMutationMax,
		deps.Config.RateLimitMutationWindow,
	))

	app := fiber.New(fiber.Config{
		AppName:      deps.Config.ServiceName,
		BodyLimit:    1 * 1024 * 1024,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Wajib menyalin buffer string request agar string tetap imutabel saat diteruskan ke metadata gRPC/HTTP2 lintas goroutine.
		Immutable:          true,
		ProxyHeader:        fiber.HeaderXForwardedFor,
		TrustProxy:         true,
		EnableIPValidation: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback: true,
			Private:  true,
			Proxies:  deps.Config.TrustedProxies,
		},
	})

	app.Use(requestPanicRecoverMiddleware(deps.Logger))
	app.Use(requestSecurityAndLoggingMiddleware(deps.Logger))

	httpGQLHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.Clone(strings.TrimSpace(w.Header().Get("X-Request-ID")))
		if requestID == "" {
			requestID = sanitizeRequestID(r.Header.Get("X-Request-ID"))
			if requestID == "" {
				requestID = uuid.NewString()
			}
			w.Header().Set("X-Request-ID", requestID)
		}

		clientIP, _ := r.Context().Value(clientIPContextKey{}).(string)
		clientIP = normalizeClientIP(clientIP)
		if clientIP == "" {
			clientIP = normalizeClientIP(r.RemoteAddr)
		}
		rawCookie := strings.Clone(browserauth.ReadCookieToken(r, deps.Config.CookieName))
		meta := middleware.RequestMetadata{
			RequestID:  requestID,
			ClientIP:   clientIP,
			UserAgent:  sanitizeUserAgent(r.UserAgent()),
			RawCookie:  rawCookie,
			CookieHash: authsession.HashOpaqueToken(rawCookie),
			CSRFToken:  strings.Clone(strings.TrimSpace(r.Header.Get(browserauth.CSRFHeaderName))),
		}

		effects := browserauth.NewResponseEffects()
		ctx := middleware.WithRequestMetadata(r.Context(), meta)
		ctx = browserauth.WithResponseEffects(ctx, effects)
		ctx = withHTTPRequest(ctx, r)

		ew := browserauth.NewEffectResponseWriter(w, effects)
		gqlSrv.ServeHTTP(ew, r.WithContext(ctx))
	})

	adapted := adaptor.HTTPHandler(httpGQLHandler)
	handleGraphQL := func(c fiber.Ctx) error {
		reqID := strings.Clone(sanitizeRequestID(c.Get("X-Request-ID")))
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set("X-Request-ID", reqID)
		c.RequestCtx().SetUserValue(clientIPContextKey{}, strings.Clone(c.IP()))
		return adapted(c)
	}

	app.Get(graphqlEndpointPath, handleGraphQL)
	app.Post(graphqlEndpointPath, handleGraphQL)

	// Rute lain selain /api/graphql wajib ditolak 404 agar tidak pernah mengekspos /api/auth/* atau /healthz.
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"errors": []fiber.Map{
				{
					"message": "not found",
					"extensions": fiber.Map{
						"code": "NOT_FOUND",
					},
				},
			},
		})
	})

	return &Server{
		cfg:    deps.Config,
		logger: deps.Logger,
		app:    app,
		gqlSrv: gqlSrv,
	}, nil
}

// App mengembalikan *fiber.App internal terutama untuk pengujian HTTP terintegrasi dengan app.Test.
func (s *Server) App() *fiber.App {
	return s.app
}

// Listen menjalankan HTTP server Fiber pada alamat konfigurasi hingga dihentikan.
func (s *Server) Listen() error {
	addr := s.cfg.HTTPAddress()
	s.logger.Info("starting IntegriOne GraphQL Gateway",
		slog.String("service", s.cfg.ServiceName),
		slog.String("address", addr),
		slog.String("env", s.cfg.AppEnv),
		slog.Int("max_complexity", s.cfg.MaxComplexity),
		slog.Bool("introspection", s.cfg.EnableIntrospection),
	)
	return s.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: true,
	})
}

// Shutdown menutup server HTTP secara anggun dalam batas konteks timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.app == nil {
		return nil
	}
	return s.app.ShutdownWithContext(ctx)
}

func requestPanicRecoverMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if logger != nil {
					logger.Error("gateway http panic recovered",
						slog.String("method", c.Method()),
						slog.String("path", c.Path()),
						slog.String("request_id", c.GetRespHeader("X-Request-ID")),
						slog.Any("panic", recovered),
						slog.String("stack", string(debug.Stack())),
					)
				}
				c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				c.Set("X-Content-Type-Options", "nosniff")
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"errors": []fiber.Map{
						{
							"message": "internal server error",
							"extensions": fiber.Map{
								"code": string(middleware.CodeInternalError),
								"type": "INTERNAL_ERROR",
							},
						},
					},
				})
			}
		}()
		return c.Next()
	}
}

func requestSecurityAndLoggingMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Set("Pragma", "no-cache")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Set("Referrer-Policy", "same-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Set("Cross-Origin-Resource-Policy", "same-origin")
		c.Set("Vary", "Cookie, Origin")

		err := c.Next()
		statusCode := c.Response().StatusCode()
		reqID := c.GetRespHeader("X-Request-ID")

		if logger != nil {
			attrs := []any{
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.String("client_ip", c.IP()),
			}
			if reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}
			if err != nil {
				logger.Warn("gateway http request error", append(attrs, slog.String("error", err.Error()))...)
			} else {
				logger.Info("gateway http request completed", attrs...)
			}
		}
		return err
	}
}

func sanitizeRequestID(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || len(trimmed) > 128 {
		return ""
	}
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
		default:
			return ""
		}
	}
	return trimmed
}

// normalizeClientIP membersihkan alamat IP klien serta melepaskan port dari RemoteAddr bila ada.
func normalizeClientIP(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		trimmed = host
	}
	trimmed = strings.TrimSpace(trimmed)
	if parsed := net.ParseIP(trimmed); parsed != nil {
		return strings.Clone(parsed.String())
	}
	cleaned := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, trimmed)
	if len(cleaned) > 128 {
		cleaned = cleaned[:128]
	}
	return strings.Clone(cleaned)
}

// sanitizeUserAgent membersihkan karakter kontrol dari User-Agent serta membatasi panjang header ke 512 byte.
func sanitizeUserAgent(raw string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, strings.TrimSpace(raw))
	if len(cleaned) > 512 {
		cleaned = strings.ToValidUTF8(cleaned[:512], "")
	}
	return strings.Clone(cleaned)
}
