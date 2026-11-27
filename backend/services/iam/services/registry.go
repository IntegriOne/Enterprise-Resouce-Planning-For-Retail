package services

import (
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/mailer"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/totp"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/group"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
	"github.com/integrione/erp-retail/backend/services/iam/services/provisioning"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

// SecurityOptions mengonfigurasi adaptor kriptografi, keyring 2FA/CSRF, kebijakan lockout, allowlist rute redirect, dan proyeksi Redis IAM.
type SecurityOptions struct {
	ScopeVerifier         domainservice.ScopeOwnerVerifier
	TOTPIssuer            string
	TOTPActiveKeyVersion  int16
	TOTPKeyring           map[int16][]byte
	AllowEphemeralDevKey  bool
	CSRFActiveKeyVersion  int
	CSRFKeyring           map[int][]byte
	AllowEphemeralCSRFKey bool
	AllowedRedirectPaths  []string
	MaxFailedAttempts     int
	LockDuration          time.Duration

	Clock          domainservice.Clock
	IDGen          domainservice.IDGenerator
	PasswordHasher domainservice.PasswordHasher
	TokenGen       domainservice.TokenGenerator
	SecretCipher   domainservice.SecretCipher
	TOTPProvider   domainservice.TOTPProvider
	CSRFDeriver    domainservice.CSRFDeriver
	EmailSender    domainservice.EmailSender
	PublicBaseURL  string

	RedisClient *redis.Client
	RedisEnv    string
	Logger      *slog.Logger
}

// Registry menyatukan seluruh layanan domain aplikasi IAM beserta koordinator outbox dan worker rekonsiliasi Redis.
type Registry struct {
	EmployeeProvisioning *provisioning.Service
	Permissions          *permission.Service
	Roles                *role.Service
	Groups               *group.Service
	Sessions             *session.Service
	Users                *user.Service

	Projector       *redisstore.Projector
	OutboxCoord     *redisstore.Coordinator
	Reconciler      *redisstore.Reconciler
	EmailDispatcher mailer.SenderLifecycle
}

// NewRegistryFromConfig merangkai seluruh layanan aplikasi IAM dari repository.Registry dan SecurityConfig runtime.
func NewRegistryFromConfig(repos *iamrepo.Registry, sec config.SecurityConfig) (*Registry, error) {
	return NewRegistryFromConfigWithRuntime(repos, sec, nil, "", nil)
}

// NewRegistryFromConfigWithRuntime merangkai seluruh layanan IAM sekaligus proyektor dan rekonsiler Redis bila tersedia.
func NewRegistryFromConfigWithRuntime(
	repos *iamrepo.Registry,
	sec config.SecurityConfig,
	redisClient *redis.Client,
	redisEnv string,
	logger *slog.Logger,
) (*Registry, error) {
	return NewRegistry(repos, BuildSecurityOptionsFromConfig(sec, redisClient, redisEnv, logger))
}

// BuildSecurityOptionsFromConfig menerjemahkan config.SecurityConfig beserta dependensi runtime ke services.SecurityOptions.
func BuildSecurityOptionsFromConfig(
	sec config.SecurityConfig,
	redisClient *redis.Client,
	redisEnv string,
	logger *slog.Logger,
) SecurityOptions {
	return SecurityOptions{
		TOTPIssuer:            sec.TOTPIssuer,
		TOTPActiveKeyVersion:  sec.TOTPActiveKeyVersion,
		TOTPKeyring:           sec.TOTPKeyring,
		AllowEphemeralDevKey:  sec.AllowEphemeralDevKey,
		CSRFActiveKeyVersion:  sec.CSRFActiveKeyVersion,
		CSRFKeyring:           sec.CSRFKeyring,
		AllowEphemeralCSRFKey: sec.AllowEphemeralCSRFKey,
		AllowedRedirectPaths:  sec.AllowedRedirectPaths,
		MaxFailedAttempts:     sec.MaxFailedAttempts,
		LockDuration:          sec.LockDuration,
		RedisClient:           redisClient,
		RedisEnv:              redisEnv,
		Logger:                logger,
	}
}

// NewRegistry merangkai seluruh 5 layanan aplikasi IAM secara fail-closed dari repository.Registry.
func NewRegistry(repos *iamrepo.Registry, opts SecurityOptions) (*Registry, error) {
	if repos == nil {
		return nil, iamerrors.NewValidation("repository_registry", iamerrors.Code("iam.service.nil_repo_registry"), "repository registry is required")
	}

	clock := opts.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	idGen := opts.IDGen
	if idGen == nil {
		idGen = crypto.NewUUIDGenerator()
	}
	hasher := opts.PasswordHasher
	if hasher == nil {
		hasher = crypto.NewDefaultArgon2idHasher()
	}
	tokenGen := opts.TokenGen
	if tokenGen == nil {
		tokenGen = crypto.NewOpaqueTokenGenerator()
	}

	totpProv, err := resolveTOTPProvider(opts)
	if err != nil {
		return nil, err
	}
	secretCipher, err := resolveSecretCipher(opts)
	if err != nil {
		return nil, err
	}
	csrfDeriver, err := resolveCSRFDeriver(opts)
	if err != nil {
		return nil, err
	}
	redirectResolver, err := resolveRedirectResolver(opts.AllowedRedirectPaths)
	if err != nil {
		return nil, err
	}

	projector, outboxCoord, reconciler, err := resolveRedisRuntime(repos, opts, idGen, clock)
	if err != nil {
		return nil, err
	}

	sessionPolicy := domainservice.NewDefaultSessionPolicy()
	accessEvaluator := domainservice.NewAccessEvaluator(sessionPolicy)
	passwordPolicy := domainservice.NewDefaultPasswordPolicy()

	permSvc, err := permission.NewService(repos.Permissions, clock, idGen)
	if err != nil {
		return nil, err
	}

	roleSvc, err := role.NewService(role.Dependencies{
		Transactor:      repos.Transactor,
		Roles:           repos.Roles,
		RolePermissions: repos.RolePermissions,
		Permissions:     repos.Permissions,
		Users:           repos.Users,
		UserRoles:       repos.UserRoles,
		SecurityRuntime: repos.SecurityRuntime,
		Clock:           clock,
		IDGen:           idGen,
		Redirect:        redirectResolver,
		OutboxCoord:     outboxCoord,
		Projector:       projector,
	})
	if err != nil {
		return nil, err
	}

	groupSvc, err := group.NewService(group.Dependencies{
		Transactor:      repos.Transactor,
		Groups:          repos.Groups,
		GroupMembers:    repos.GroupMembers,
		GroupRoles:      repos.GroupRoles,
		Roles:           repos.Roles,
		Users:           repos.Users,
		SecurityRuntime: repos.SecurityRuntime,
		Clock:           clock,
		IDGen:           idGen,
		OutboxCoord:     outboxCoord,
		Projector:       projector,
	})
	if err != nil {
		return nil, err
	}

	sessSvc, err := session.NewService(session.Dependencies{
		Sessions:        repos.Sessions,
		Users:           repos.Users,
		Transactor:      repos.Transactor,
		SecurityRuntime: repos.SecurityRuntime,
		Clock:           clock,
		IDGen:           idGen,
		Tokens:          tokenGen,
		CSRFDeriver:     csrfDeriver,
		SessionPolicy:   sessionPolicy,
		AccessEvaluator: accessEvaluator,
		Redirect:        redirectResolver,
		OutboxCoord:     outboxCoord,
		Projector:       projector,
	})
	if err != nil {
		return nil, err
	}

	userCfg := user.DefaultConfig()
	if opts.MaxFailedAttempts > 0 {
		userCfg.MaxFailedAttempts = opts.MaxFailedAttempts
	}
	if opts.LockDuration > 0 {
		userCfg.LockDuration = opts.LockDuration
	}
	if opts.TOTPIssuer != "" {
		userCfg.TOTPIssuer = opts.TOTPIssuer
	}
	if opts.PublicBaseURL != "" {
		userCfg.PublicBaseURL = opts.PublicBaseURL
	}

	var emailDisp mailer.SenderLifecycle
	if lc, ok := opts.EmailSender.(mailer.SenderLifecycle); ok {
		emailDisp = lc
	}

	userSvc, err := user.NewService(user.Dependencies{
		EmployeeProvisioning: repos.EmployeeProvisioning,
		Transactor:           repos.Transactor,
		Users:                repos.Users,
		Identities:           repos.UserIdentities,
		Credentials:          repos.UserCredentials,
		TOTPFactors:          repos.UserTOTPFactors,
		RecoveryCodes:        repos.UserRecoveryCodes,
		Tokens:               repos.UserTokens,
		Roles:                repos.Roles,
		UserRoles:            repos.UserRoles,
		SecurityRuntime:      repos.SecurityRuntime,
		Sessions:             sessSvc,
		Clock:                clock,
		IDGen:                idGen,
		PasswordHasher:       hasher,
		PasswordPolicy:       passwordPolicy,
		TokenGen:             tokenGen,
		SecretCipher:         secretCipher,
		TOTPProvider:         totpProv,
		Redirect:             redirectResolver,
		EmailSender:          opts.EmailSender,
		OutboxCoord:          outboxCoord,
		Projector:            projector,
		Logger:               opts.Logger,
		Config:               userCfg,
	})
	if err != nil {
		return nil, err
	}

	reg := &Registry{
		EmployeeProvisioning: provisioning.NewService(repos.EmployeeProvisioning, repos.Users, repos.Roles, opts.ScopeVerifier, clock),
		Permissions:          permSvc,
		Roles:                roleSvc,
		Groups:               groupSvc,
		Sessions:             sessSvc,
		Users:                userSvc,
		Projector:            projector,
		OutboxCoord:          outboxCoord,
		Reconciler:           reconciler,
		EmailDispatcher:      emailDisp,
	}
	if err := reg.Validate(); err != nil {
		return nil, err
	}
	return reg, nil
}

// Validate memastikan seluruh layanan pada Registry telah terinisialisasi tanpa pointer nil.
func (r *Registry) Validate() error {
	if r == nil || r.Permissions == nil || r.Roles == nil || r.Groups == nil || r.Sessions == nil || r.Users == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.service.incomplete_registry"), "iam service registry is incomplete")
	}
	return nil
}

func resolveTOTPProvider(opts SecurityOptions) (domainservice.TOTPProvider, error) {
	if opts.TOTPProvider != nil {
		return opts.TOTPProvider, nil
	}
	cfg := totp.DefaultConfig()
	if opts.TOTPIssuer != "" {
		cfg.DefaultIssuer = opts.TOTPIssuer
	}
	return totp.NewProvider(cfg)
}

func resolveSecretCipher(opts SecurityOptions) (domainservice.SecretCipher, error) {
	if opts.SecretCipher != nil {
		return opts.SecretCipher, nil
	}
	if len(opts.TOTPKeyring) > 0 {
		ver := opts.TOTPActiveKeyVersion
		if ver <= 0 {
			ver = 1
		}
		return crypto.NewAESGCMKeyringCipher(ver, opts.TOTPKeyring)
	}
	if opts.AllowEphemeralDevKey {
		return crypto.NewEphemeralAES256Cipher()
	}
	return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.service.missing_secret_cipher"), "totp encryption keyring is required")
}

func resolveCSRFDeriver(opts SecurityOptions) (domainservice.CSRFDeriver, error) {
	if opts.CSRFDeriver != nil {
		return opts.CSRFDeriver, nil
	}
	ver := opts.CSRFActiveKeyVersion
	if ver <= 0 {
		ver = 1
	}
	allowEphemeral := opts.AllowEphemeralCSRFKey
	return crypto.NewHMACCSRFDeriver(ver, opts.CSRFKeyring, allowEphemeral)
}

func resolveRedirectResolver(allowedPaths []string) (domainservice.RedirectResolver, error) {
	if len(allowedPaths) > 0 {
		return domainservice.NewRedirectResolver(allowedPaths)
	}
	return domainservice.NewDefaultRedirectResolver(), nil
}

func resolveRedisRuntime(
	repos *iamrepo.Registry,
	opts SecurityOptions,
	idGen domainservice.IDGenerator,
	clock domainservice.Clock,
) (*redisstore.Projector, *redisstore.Coordinator, *redisstore.Reconciler, error) {
	var store *authsession.Store
	if opts.RedisClient != nil {
		var err error
		store, err = authsession.NewStore(opts.RedisClient, authsession.NewKeyspace(opts.RedisEnv))
		if err != nil {
			return nil, nil, nil, err
		}
	}
	projector := redisstore.NewProjector(store, clock)
	var coord *redisstore.Coordinator
	var reconciler *redisstore.Reconciler
	if repos.SecurityRuntime != nil {
		coord = redisstore.NewCoordinator(repos.SecurityRuntime, projector, idGen, clock, opts.Logger)
		reconciler = redisstore.NewReconciler(
			repos.SecurityRuntime,
			repos.Users,
			repos.Sessions,
			projector,
			coord,
			clock,
			opts.Logger,
		)
	}
	return projector, coord, reconciler, nil
}
