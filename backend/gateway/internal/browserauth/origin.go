package browserauth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	// ErrOriginRequired menandakan mutasi GraphQL wajib membawa header Origin tepercaya.
	ErrOriginRequired = errors.New("origin header is required")
	// ErrOriginForbidden menandakan header Origin atau Sec-Fetch-Site tidak termasuk origin yang diizinkan.
	ErrOriginForbidden = errors.New("origin is not allowed")
)

// OriginValidator menegakkan allowlist Origin serta pengecekan Sec-Fetch-Site pada permintaan browser.
type OriginValidator struct {
	allowed map[string]struct{}
}

// NewOriginValidator membentuk validator origin berdasarkan daftar skema://host[:port] tepercaya.
func NewOriginValidator(allowedOrigins []string) (*OriginValidator, error) {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, raw := range allowedOrigins {
		norm, err := normalizeOrigin(raw)
		if err != nil {
			return nil, err
		}
		allowed[norm] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, errors.New("at least one allowed origin is required")
	}
	return &OriginValidator{allowed: allowed}, nil
}

// ValidateRequest memeriksa keabsahan header Origin dan Sec-Fetch-Site.
// Jika requireOrigin bernilai false (misalnya query SSR cookie-only), Origin hanya divalidasi bila hadir.
func (v *OriginValidator) ValidateRequest(h http.Header, requireOrigin bool) error {
	if v == nil || len(v.allowed) == 0 {
		return ErrOriginForbidden
	}

	secFetchSite := strings.ToLower(strings.TrimSpace(h.Get("Sec-Fetch-Site")))
	if secFetchSite != "" {
		switch secFetchSite {
		case "same-origin", "none":
		default:
			return ErrOriginForbidden
		}
	}

	rawOrigin := strings.TrimSpace(h.Get("Origin"))
	if rawOrigin == "" {
		if requireOrigin {
			return ErrOriginRequired
		}
		return nil
	}
	if strings.EqualFold(rawOrigin, "null") {
		return ErrOriginForbidden
	}

	normOrigin, err := normalizeOrigin(rawOrigin)
	if err != nil {
		return ErrOriginForbidden
	}
	if _, ok := v.allowed[normOrigin]; !ok {
		return ErrOriginForbidden
	}
	return nil
}

func normalizeOrigin(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "*" {
		return "", fmt.Errorf("invalid origin %q", raw)
	}
	u, err := url.Parse(trimmed)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid origin URL %q", raw)
	}
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host), nil
}
