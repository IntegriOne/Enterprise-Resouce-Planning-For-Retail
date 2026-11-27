// Package mailer menyediakan implementasi pengiriman email transaksional IAM untuk Amazon SES v2,
// SMTP (SES SMTP, SendGrid SMTP, Mailpit), Twilio SendGrid API, Resend API, Postmark API, serta dispatcher asinkron.
package mailer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"net/url"
	"strings"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

// Provider menentukan backend pengiriman email yang aktif.
type Provider string

const (
	ProviderDisabled Provider = "disabled"
	ProviderLog      Provider = "log"
	ProviderSES      Provider = "ses"
	ProviderSMTP     Provider = "smtp"
	ProviderSendGrid Provider = "sendgrid"
	ProviderResend   Provider = "resend"
	ProviderPostmark Provider = "postmark"
)

// DeliveryMode menentukan apakah pengiriman dijalankan asinkron lewat worker antrean atau sinkron.
type DeliveryMode string

const (
	DeliveryModeAsync DeliveryMode = "async"
	DeliveryModeSync  DeliveryMode = "sync"
)

// SMTPTLSMode mengontrol perilaku enkripsi koneksi SMTP keluar.
type SMTPTLSMode string

const (
	SMTPTLSStartTLS SMTPTLSMode = "starttls"
	SMTPTLSImplicit SMTPTLSMode = "tls"
	SMTPTLSNone     SMTPTLSMode = "none"
)

// Config menyimpan seluruh parameter konfigurasi provider email serta dispatcher IAM.
type Config struct {
	Provider      Provider
	DeliveryMode  DeliveryMode
	From          string
	ReplyTo       string
	PublicBaseURL string

	Timeout             time.Duration
	QueueSize           int
	Workers             int
	MaxAttempts         int
	InitialRetryBackoff time.Duration

	// Konfigurasi Amazon SES v2 API.
	SESRegion           string
	SESAccessKeyID      string
	SESSecretAccessKey  string
	SESSessionToken     string
	SESConfigurationSet string
	SESEndpointURL      string

	// Konfigurasi SMTP (AWS SES SMTP, SendGrid SMTP, Mailpit, Postfix).
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	SMTPTLSMode    SMTPTLSMode
	SMTPServerName string
	AllowPlaintext bool

	// Konfigurasi API pihak ketiga (Twilio SendGrid, Resend, Postmark).
	APIKey     string
	APIBaseURL string
}

// String menyamarkan kredensial SMTP, SES, dan API key email agar tidak bocor lewat fmt/log.
func (c Config) String() string {
	return fmt.Sprintf("mailer.Config{Provider:%s, DeliveryMode:%s, Host:%s, Port:%d, AccessKey:%s, Secret:%s, SessionToken:%s, SMTPPassword:%s, APIKey:%s}",
		c.Provider, c.DeliveryMode, c.SMTPHost, c.SMTPPort,
		maskConfigFlag(c.SESAccessKeyID), maskConfigFlag(c.SESSecretAccessKey),
		maskConfigFlag(c.SESSessionToken), maskConfigFlag(c.SMTPPassword), maskConfigFlag(c.APIKey))
}

// GoString menyamarkan kredensial pada format %#v.
func (c Config) GoString() string {
	return c.String()
}

// LogValue memancarkan metadata konfigurasi email aman untuk slog tanpa kredensial mentah.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("provider", string(c.Provider)),
		slog.String("delivery_mode", string(c.DeliveryMode)),
		slog.String("smtp_host", c.SMTPHost),
		slog.Int("smtp_port", c.SMTPPort),
		slog.String("ses_access_key_id", maskConfigFlag(c.SESAccessKeyID)),
		slog.String("ses_secret_access_key", maskConfigFlag(c.SESSecretAccessKey)),
		slog.String("ses_session_token", maskConfigFlag(c.SESSessionToken)),
		slog.String("smtp_password", maskConfigFlag(c.SMTPPassword)),
		slog.String("api_key", maskConfigFlag(c.APIKey)),
	)
}

func maskConfigFlag(val string) string {
	if strings.TrimSpace(val) == "" {
		return "<empty>"
	}
	return "[REDACTED]"
}

// DefaultConfig mengembalikan konfigurasi bawaan yang aman untuk lingkungan pengembangan lokal.
func DefaultConfig() Config {
	return Config{
		Provider:            ProviderDisabled,
		DeliveryMode:        DeliveryModeAsync,
		From:                "IntegriOne Security <no-reply@integrione.local>",
		PublicBaseURL:       "https://erp.localhost:8443",
		Timeout:             8 * time.Second,
		QueueSize:           256,
		Workers:             2,
		MaxAttempts:         3,
		InitialRetryBackoff: 250 * time.Millisecond,
		SMTPPort:            587,
		SMTPTLSMode:         SMTPTLSStartTLS,
	}
}

// Normalize mengisi nilai bawaan yang wajar dan menormalkan identifier konfigurasi.
func (c Config) Normalize() Config {
	def := DefaultConfig()
	if strings.TrimSpace(string(c.Provider)) == "" {
		c.Provider = def.Provider
	} else {
		c.Provider = Provider(strings.ToLower(strings.TrimSpace(string(c.Provider))))
	}
	if strings.TrimSpace(string(c.DeliveryMode)) == "" {
		c.DeliveryMode = def.DeliveryMode
	} else {
		c.DeliveryMode = DeliveryMode(strings.ToLower(strings.TrimSpace(string(c.DeliveryMode))))
	}
	c.From = strings.TrimSpace(c.From)
	if c.From == "" {
		c.From = def.From
	}
	c.ReplyTo = strings.TrimSpace(c.ReplyTo)
	c.PublicBaseURL = strings.TrimRight(strings.TrimSpace(c.PublicBaseURL), "/")
	if c.PublicBaseURL == "" {
		c.PublicBaseURL = def.PublicBaseURL
	}
	if c.Timeout <= 0 {
		c.Timeout = def.Timeout
	}
	if c.QueueSize <= 0 {
		c.QueueSize = def.QueueSize
	}
	if c.Workers <= 0 {
		c.Workers = def.Workers
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = def.MaxAttempts
	}
	if c.InitialRetryBackoff <= 0 {
		c.InitialRetryBackoff = def.InitialRetryBackoff
	}
	if c.SMTPPort <= 0 {
		if c.SMTPTLSMode == SMTPTLSImplicit {
			c.SMTPPort = 465
		} else {
			c.SMTPPort = def.SMTPPort
		}
	}
	if strings.TrimSpace(string(c.SMTPTLSMode)) == "" {
		c.SMTPTLSMode = def.SMTPTLSMode
	} else {
		c.SMTPTLSMode = SMTPTLSMode(strings.ToLower(strings.TrimSpace(string(c.SMTPTLSMode))))
	}
	c.SMTPHost = strings.TrimSpace(c.SMTPHost)
	c.SMTPServerName = strings.TrimSpace(c.SMTPServerName)
	c.SESRegion = strings.TrimSpace(c.SESRegion)
	c.SESConfigurationSet = strings.TrimSpace(c.SESConfigurationSet)
	c.SESEndpointURL = strings.TrimSpace(c.SESEndpointURL)
	c.APIBaseURL = strings.TrimRight(strings.TrimSpace(c.APIBaseURL), "/")
	return c
}

// Validate memeriksa keabsahan konfigurasi provider, alamat pengirim, URL dasar tautan, dan keamanan transport.
func (c Config) Validate(isProduction bool) error {
	cfg := c.Normalize()

	if cfg.DeliveryMode != DeliveryModeAsync && cfg.DeliveryMode != DeliveryModeSync {
		return fmt.Errorf("unsupported email delivery mode %q", cfg.DeliveryMode)
	}

	if err := validatePublicBaseURL(cfg.PublicBaseURL); err != nil {
		return err
	}

	if cfg.Provider == ProviderDisabled {
		return nil
	}

	if cfg.Provider == ProviderLog {
		return nil
	}

	if err := validateHeaderSafe("from", cfg.From); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(cfg.From); err != nil {
		return fmt.Errorf("invalid IAM_EMAIL_FROM address: %w", err)
	}
	if cfg.ReplyTo != "" {
		if err := validateHeaderSafe("reply_to", cfg.ReplyTo); err != nil {
			return err
		}
		if _, err := mail.ParseAddress(cfg.ReplyTo); err != nil {
			return fmt.Errorf("invalid IAM_EMAIL_REPLY_TO address: %w", err)
		}
	}

	switch cfg.Provider {
	case ProviderSES:
		if cfg.SESRegion == "" {
			return errors.New("SES region is required when IAM_EMAIL_PROVIDER=ses")
		}
		if cfg.SESEndpointURL != "" {
			if _, err := url.ParseRequestURI(cfg.SESEndpointURL); err != nil {
				return fmt.Errorf("invalid SES endpoint url: %w", err)
			}
		}
		if (cfg.SESAccessKeyID == "") != (cfg.SESSecretAccessKey == "") {
			return errors.New("both access key id and secret access key must be provided together when using static AWS credentials")
		}
	case ProviderSMTP:
		if cfg.SMTPHost == "" {
			return errors.New("SMTP host is required when IAM_EMAIL_PROVIDER=smtp")
		}
		if cfg.SMTPPort <= 0 || cfg.SMTPPort > 65535 {
			return fmt.Errorf("invalid SMTP port %d", cfg.SMTPPort)
		}
		switch cfg.SMTPTLSMode {
		case SMTPTLSStartTLS, SMTPTLSImplicit:
		case SMTPTLSNone:
			if isProduction || !cfg.AllowPlaintext {
				return errors.New("plaintext SMTP (tls_mode=none) is only allowed in non-production environments when IAM_SMTP_ALLOW_PLAINTEXT=true")
			}
		default:
			return fmt.Errorf("unsupported SMTP TLS mode %q", cfg.SMTPTLSMode)
		}
	case ProviderSendGrid, ProviderResend, ProviderPostmark:
		if strings.TrimSpace(cfg.APIKey) == "" {
			return fmt.Errorf("API key is required when IAM_EMAIL_PROVIDER=%s", cfg.Provider)
		}
		if cfg.APIBaseURL != "" {
			u, err := url.Parse(cfg.APIBaseURL)
			if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
				return fmt.Errorf("invalid API base URL for provider %s", cfg.Provider)
			}
		}
	default:
		return fmt.Errorf("unsupported email provider %q", cfg.Provider)
	}

	return nil
}

func validatePublicBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return fmt.Errorf("invalid IAM_PUBLIC_BASE_URL %q: must be valid http or https URL", raw)
	}
	return nil
}

// PermanentError menandai kegagalan pengiriman permanen (mis. 4xx payload ditolak) agar tidak diulang sia-sia.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	if e == nil || e.Err == nil {
		return "permanent email delivery error"
	}
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsPermanent melaporkan apakah error pengiriman bersifat permanen sehingga tidak perlu retry.
func IsPermanent(err error) bool {
	_, ok := errors.AsType[*PermanentError](err)
	return ok
}

// SenderLifecycle menggabungkan EmailSender dengan antrean/worker lifecycle agar dapat didrain saat shutdown.
type SenderLifecycle interface {
	domainservice.EmailSender
	Start(ctx context.Context)
	Close(ctx context.Context) error
}

// NewSender membangun sender email beserta dispatcher sesuai konfigurasi runtime IAM.
func NewSender(ctx context.Context, rawCfg Config, isProduction bool, logger *slog.Logger) (SenderLifecycle, error) {
	cfg := rawCfg.Normalize()
	if err := cfg.Validate(isProduction); err != nil {
		return nil, err
	}

	var base domainservice.EmailSender
	switch cfg.Provider {
	case ProviderDisabled:
		base = noopSender{}
	case ProviderLog:
		base = &logSender{logger: logger}
	case ProviderSES:
		sesSender, err := NewSESSender(ctx, cfg)
		if err != nil {
			return nil, err
		}
		base = sesSender
	case ProviderSMTP:
		smtpSender, err := NewSMTPSender(cfg)
		if err != nil {
			return nil, err
		}
		base = smtpSender
	case ProviderSendGrid, ProviderResend, ProviderPostmark:
		httpSender, err := NewHTTPProviderSender(cfg)
		if err != nil {
			return nil, err
		}
		base = httpSender
	default:
		return nil, fmt.Errorf("unsupported email provider %q", cfg.Provider)
	}

	return NewDispatcher(base, DispatcherConfig{
		Mode:                cfg.DeliveryMode,
		Workers:             cfg.Workers,
		QueueSize:           cfg.QueueSize,
		SendTimeout:         cfg.Timeout,
		MaxAttempts:         cfg.MaxAttempts,
		InitialRetryBackoff: cfg.InitialRetryBackoff,
	}, logger), nil
}

type noopSender struct{}

func (noopSender) ProviderName() string {
	return string(ProviderDisabled)
}

func (noopSender) Send(_ context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	if err := validateOutboundMessage(msg); err != nil {
		return nil, &PermanentError{Err: err}
	}
	return &domainservice.EmailSendResult{
		Provider:  string(ProviderDisabled),
		MessageID: "disabled",
	}, nil
}

type logSender struct {
	logger *slog.Logger
}

func (s *logSender) ProviderName() string {
	return string(ProviderLog)
}

func (s *logSender) Send(_ context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	if err := validateOutboundMessage(msg); err != nil {
		return nil, &PermanentError{Err: err}
	}
	// Keamanan: jangan pernah mencatat token reset/verifikasi ataupun isi HTML mentah ke dalam log.
	if s.logger != nil {
		s.logger.Info("transactional email simulated in log mode",
			slog.String("provider", string(ProviderLog)),
			slog.String("purpose", string(msg.Purpose)),
			slog.String("recipient", domainservice.MaskEmailForLog(msg.To)),
			slog.String("subject", msg.Subject),
		)
	}
	return &domainservice.EmailSendResult{
		Provider:  string(ProviderLog),
		MessageID: "log-simulated",
	}, nil
}
