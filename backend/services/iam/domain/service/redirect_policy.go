package service

import (
	"slices"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// defaultAllowedRedirectRoutes memuat daftar rute portal resmi aplikasi terpadu IntegriOne.
var defaultAllowedRedirectRoutes = []string{
	"/",
	"/identity",
	"/human-resource",
	"/procurement",
	"/warehouse",
	"/finance",
	"/accounting",
	"/product",
	"/supplier",
	"/vendor",
	"/commerce",
	"/account",
	"/profile",
}

// RedirectResolver menentukan tujuan navigasi pasca-login secara deterministik dan terlindung dari open redirect.
type RedirectResolver struct {
	allowedPrefixes []string
}

// NewDefaultRedirectResolver membentuk RedirectResolver dengan allowlist rute resmi IntegriOne.
func NewDefaultRedirectResolver() RedirectResolver {
	return RedirectResolver{
		allowedPrefixes: slices.Clone(defaultAllowedRedirectRoutes),
	}
}

// NewRedirectResolver membentuk RedirectResolver dengan allowlist khusus yang terlebih dahulu divalidasi keamanannya.
func NewRedirectResolver(allowedRoutes []string) (RedirectResolver, error) {
	cleaned := make([]string, 0, len(allowedRoutes))
	for _, route := range allowedRoutes {
		trimmed := strings.TrimSpace(route)
		if err := entity.ValidateInternalRedirectPath(trimmed); err != nil {
			return RedirectResolver{}, err
		}
		cleaned = append(cleaned, trimmed)
	}
	return RedirectResolver{allowedPrefixes: cleaned}, nil
}

// ValidateAllowedPath memverifikasi bahwa jalur redirect bersifat relatif internal sekaligus masuk dalam allowlist resmi.
func (r RedirectResolver) ValidateAllowedPath(path string) error {
	if err := entity.ValidateInternalRedirectPath(path); err != nil {
		return err
	}
	for _, allowed := range r.allowedPrefixes {
		if path == allowed {
			return nil
		}
		if allowed != "/" && strings.HasPrefix(path, allowed+"/") {
			return nil
		}
	}
	return iamerrors.ErrRedirectPathInvalid
}

// ResolvePostLoginPath memilih rute awal paling prioritas di antara peran aktif, atau fallback aman sesuai user_kind.
func (r RedirectResolver) ResolvePostLoginPath(userKind entity.UserKind, roles []entity.Role) (string, error) {
	if !userKind.IsValid() {
		return "", iamerrors.ErrInvalidUserKind
	}
	if userKind == entity.UserKindSystem {
		// Akun otomasi system non-interaktif tidak pernah memperoleh rute navigasi dashboard interaktif.
		roles = nil
	}

	var best *entity.Role
	for i := range roles {
		role := &roles[i]
		candidate := strings.TrimSpace(role.DefaultRedirectPath)
		if !role.IsActive || candidate == "" {
			continue
		}
		if err := r.ValidateAllowedPath(candidate); err != nil {
			return "", err
		}
		if best == nil || isHigherPriorityRedirect(role, best) {
			best = role
		}
	}

	if best != nil {
		return strings.TrimSpace(best.DefaultRedirectPath), nil
	}

	fallback := defaultFallbackPathByUserKind(userKind)
	if err := r.ValidateAllowedPath(fallback); err != nil {
		return "", err
	}
	return fallback, nil
}

func isHigherPriorityRedirect(candidate, current *entity.Role) bool {
	if candidate.RedirectPriority != current.RedirectPriority {
		return candidate.RedirectPriority < current.RedirectPriority
	}
	candPath := strings.TrimSpace(candidate.DefaultRedirectPath)
	currPath := strings.TrimSpace(current.DefaultRedirectPath)
	if candPath != currPath {
		return candPath < currPath
	}
	return candidate.Code < current.Code
}

func defaultFallbackPathByUserKind(kind entity.UserKind) string {
	switch kind {
	case entity.UserKindVendor:
		return "/vendor"
	case entity.UserKindSupplier:
		return "/supplier"
	case entity.UserKindEmployee:
		return "/account"
	case entity.UserKindCustomer, entity.UserKindSystem:
		return "/"
	default:
		return "/"
	}
}
