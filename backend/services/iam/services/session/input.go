package session

import (
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
)

// Dependencies mengelompokkan port repositori, kebijakan domain, dan pembangkit token layanan sesi.
type Dependencies struct {
	Sessions        domainrepo.SessionRepository
	Users           domainrepo.UserRepository
	Transactor      domainrepo.Transactor
	SecurityRuntime domainrepo.SecurityRuntimeRepository
	Clock           domainservice.Clock
	IDGen           domainservice.IDGenerator
	Tokens          domainservice.TokenGenerator
	CSRFDeriver     domainservice.CSRFDeriver
	SessionPolicy   domainservice.SessionPolicy
	AccessEvaluator domainservice.AccessEvaluator
	Redirect        domainservice.RedirectResolver
	OutboxCoord     *redisstore.Coordinator
	Projector       domainservice.RuntimeSecurityProjector
}

// RefreshSnapshotResult memuat AccessSnapshot terkini beserta otorisasi akses dasbor /identity dan rute redirect.
type RefreshSnapshotResult struct {
	User                 *entity.User
	Snapshot             *entity.AccessSnapshot
	CanAccessIdentity    bool
	DefaultRedirectPath  string
	AllowedRedirectPaths []string
}

// IssueInput membawa parameter pembentukan sesi baru setelah autentikasi berhasil.
type IssueInput struct {
	UserID      string
	MFAVerified bool
	IPAddress   string
	UserAgent   string
}

// IssuedSession memuat entitas sesi serta token buram dan token CSRF polos untuk dikembalikan sekali saat pembuatan.
type IssuedSession struct {
	Session      *entity.Session
	SessionToken string
	CSRFToken    string
}

// CSRFTokenResult memuat hasil rotasi token CSRF setelah peningkatan status keamanan sesi.
type CSRFTokenResult struct {
	Session   *entity.Session
	CSRFToken string
}

// AuthorizeRequest merangkum konteks verifikasi sesi, CSRF, izin berscope, dan re-autentikasi sensitif.
type AuthorizeRequest struct {
	SessionToken           string
	CSRFToken              string
	RequireCSRF            bool
	RequiredPermission     string
	RequiredScope          entity.Scope
	RequireSensitiveReauth bool
}

// AuthorizeResult mengembalikan konteks terverifikasi lengkap bila evaluasi otorisasi berhasil.
type AuthorizeResult struct {
	User     *entity.User
	Session  *entity.Session
	Snapshot *entity.AccessSnapshot
}
