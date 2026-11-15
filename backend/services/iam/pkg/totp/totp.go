package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/url"
	"strings"
	"time"
	"unicode"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const (
	defaultIssuer        = "IntegriOne ERP"
	defaultDigits        = 6
	defaultPeriodSeconds = int64(30)
	defaultSkewSteps     = int64(1)
	defaultSecretBytes   = 20
	minSecretBytes       = 16
)

var base32NoPad = base32.StdEncoding.WithPadding(base32.NoPadding)

// Config mengatur parameter pembangkitan secret dan jendela verifikasi TOTP RFC 6238.
type Config struct {
	DefaultIssuer string
	Digits        int
	PeriodSeconds int64
	SkewSteps     int64
	SecretBytes   int
}

// DefaultConfig mengembalikan konfigurasi standar Google Authenticator (SHA-1, 6 digit, 30 detik, skew +-1).
func DefaultConfig() Config {
	return Config{
		DefaultIssuer: defaultIssuer,
		Digits:        defaultDigits,
		PeriodSeconds: defaultPeriodSeconds,
		SkewSteps:     defaultSkewSteps,
		SecretBytes:   defaultSecretBytes,
	}
}

// RFC6238Provider mengimplementasikan domainservice.TOTPProvider menggunakan HMAC-SHA1 standar RFC 6238.
type RFC6238Provider struct {
	cfg        Config
	randSource io.Reader
}

// NewDefaultProvider membentuk RFC6238Provider dengan konfigurasi standar Google Authenticator.
func NewDefaultProvider() *RFC6238Provider {
	return &RFC6238Provider{
		cfg:        DefaultConfig(),
		randSource: rand.Reader,
	}
}

// NewProvider membentuk RFC6238Provider setelah memvalidasi batas aman konfigurasi.
func NewProvider(cfg Config) (*RFC6238Provider, error) {
	if strings.TrimSpace(cfg.DefaultIssuer) == "" {
		cfg.DefaultIssuer = defaultIssuer
	}
	if cfg.Digits == 0 {
		cfg.Digits = defaultDigits
	}
	if cfg.Digits != 6 && cfg.Digits != 8 {
		return nil, iamerrors.NewValidation("totp.digits", iamerrors.Code("iam.totp.invalid_digits_config"), "totp digits must be 6 or 8")
	}
	if cfg.PeriodSeconds == 0 {
		cfg.PeriodSeconds = defaultPeriodSeconds
	}
	if cfg.PeriodSeconds < 15 || cfg.PeriodSeconds > 300 {
		return nil, iamerrors.NewValidation("totp.period_seconds", iamerrors.Code("iam.totp.invalid_period_config"), "totp period must be between 15 and 300 seconds")
	}
	if cfg.SkewSteps < 0 || cfg.SkewSteps > 3 {
		return nil, iamerrors.NewValidation("totp.skew_steps", iamerrors.Code("iam.totp.invalid_skew_config"), "totp skew steps must be between 0 and 3")
	}
	if cfg.SecretBytes == 0 {
		cfg.SecretBytes = defaultSecretBytes
	}
	if cfg.SecretBytes < minSecretBytes || cfg.SecretBytes > 64 {
		return nil, iamerrors.NewValidation("totp.secret_bytes", iamerrors.Code("iam.totp.invalid_secret_bytes"), "totp secret length must be between 16 and 64 bytes")
	}
	return &RFC6238Provider{
		cfg:        cfg,
		randSource: rand.Reader,
	}, nil
}

// GenerateSecret membangkitkan seed acak dan URI otpauth:// standar untuk Google Authenticator.
func (p *RFC6238Provider) GenerateSecret(accountName, issuer string) ([]byte, string, error) {
	if p == nil {
		return nil, "", iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.totp.nil_provider"), "totp provider is nil")
	}
	trimmedAccount := strings.TrimSpace(accountName)
	if trimmedAccount == "" || hasControlRune(trimmedAccount) {
		return nil, "", iamerrors.NewValidation("account_name", iamerrors.Code("iam.totp.invalid_account_name"), "account name must not be empty or contain control characters")
	}

	resolvedIssuer := strings.TrimSpace(issuer)
	if resolvedIssuer == "" {
		resolvedIssuer = p.cfg.DefaultIssuer
	}
	if hasControlRune(resolvedIssuer) {
		return nil, "", iamerrors.NewValidation("issuer", iamerrors.Code("iam.totp.invalid_issuer"), "issuer must not contain control characters")
	}

	src := p.randSource
	if src == nil {
		src = rand.Reader
	}
	secret := make([]byte, p.cfg.SecretBytes)
	if _, err := io.ReadFull(src, secret); err != nil {
		return nil, "", iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.totp.secret_entropy_failed"), "failed to generate totp secret bytes")
	}

	encodedSecret := base32NoPad.EncodeToString(secret)
	query := url.Values{}
	query.Set("secret", encodedSecret)
	query.Set("issuer", resolvedIssuer)
	query.Set("algorithm", "SHA1")
	query.Set("digits", fmt.Sprintf("%d", p.cfg.Digits))
	query.Set("period", fmt.Sprintf("%d", p.cfg.PeriodSeconds))

	label := url.PathEscape(resolvedIssuer + ":" + trimmedAccount)
	rawQuery := strings.ReplaceAll(query.Encode(), "+", "%20")
	uri := fmt.Sprintf("otpauth://totp/%s?%s", label, rawQuery)
	return secret, uri, nil
}

// VerifyCode mengevaluasi kode OTP terhadap jendela waktu [step-skew .. step+skew] secara constant-time.
func (p *RFC6238Provider) VerifyCode(secret []byte, code string, at time.Time) (bool, int64, error) {
	if p == nil {
		return false, 0, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.totp.nil_provider"), "totp provider is nil")
	}
	if len(secret) < minSecretBytes {
		return false, 0, iamerrors.NewValidation("secret", iamerrors.Code("iam.totp.invalid_secret_length"), "totp secret must be at least 16 bytes")
	}
	if at.IsZero() {
		return false, 0, iamerrors.NewValidation("at", iamerrors.Code("iam.totp.zero_timestamp"), "verification timestamp must not be zero")
	}

	trimmedCode := strings.TrimSpace(code)
	if len(trimmedCode) != p.cfg.Digits || !isASCIIDigits(trimmedCode) {
		return false, 0, nil
	}

	currentStep := at.Unix() / p.cfg.PeriodSeconds
	candidateBytes := []byte(trimmedCode)

	var matched bool
	var matchedStep int64
	for offset := -p.cfg.SkewSteps; offset <= p.cfg.SkewSteps; offset++ {
		step := currentStep + offset
		if step < 0 {
			continue
		}
		expected := p.computeCodeForStep(secret, step)
		if subtle.ConstantTimeCompare([]byte(expected), candidateBytes) == 1 && !matched {
			matched = true
			matchedStep = step
		}
	}

	return matched, matchedStep, nil
}

// GenerateCodeAt menghasilkan kode TOTP untuk timestamp lokal/epoch tertentu terutama guna pengujian dan verifikasi sinkron.
func (p *RFC6238Provider) GenerateCodeAt(secret []byte, at time.Time) (string, error) {
	if p == nil {
		return "", iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.totp.nil_provider"), "totp provider is nil")
	}
	if len(secret) < minSecretBytes {
		return "", iamerrors.NewValidation("secret", iamerrors.Code("iam.totp.invalid_secret_length"), "totp secret must be at least 16 bytes")
	}
	if at.IsZero() {
		return "", iamerrors.NewValidation("at", iamerrors.Code("iam.totp.zero_timestamp"), "timestamp must not be zero")
	}
	step := at.Unix() / p.cfg.PeriodSeconds
	return p.computeCodeForStep(secret, step), nil
}

func (p *RFC6238Provider) computeCodeForStep(secret []byte, step int64) string {
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(step))

	mac := hmac.New(sha1.New, secret)
	_, _ = mac.Write(counter[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])

	mod := uint32(math.Pow10(p.cfg.Digits))
	return fmt.Sprintf("%0*d", p.cfg.Digits, binCode%mod)
}

func isASCIIDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func hasControlRune(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
