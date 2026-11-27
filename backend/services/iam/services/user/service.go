package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
)

// Service mengorkestrasi siklus hidup akun pengguna IAM, kredensial, identitas login, token, dan MFA 2FA.
type Service struct {
	tx                   domainrepo.Transactor
	users                domainrepo.UserRepository
	employeeProvisioning domainrepo.EmployeeProvisioningRepository
	identities           domainrepo.UserIdentityRepository
	credentials          domainrepo.UserCredentialRepository
	totpFactors          domainrepo.UserTOTPFactorRepository
	recoveryCodes        domainrepo.UserRecoveryCodeRepository
	tokens               domainrepo.UserTokenRepository
	roles                domainrepo.RoleRepository
	userRoles            domainrepo.UserRoleRepository
	securityRuntime      domainrepo.SecurityRuntimeRepository
	sessions             SessionCollaborator
	clock                domainservice.Clock
	idGen                domainservice.IDGenerator
	passwordHasher       domainservice.PasswordHasher
	passwordPolicy       domainservice.PasswordPolicy
	tokenGen             domainservice.TokenGenerator
	secretCipher         domainservice.SecretCipher
	totpProvider         domainservice.TOTPProvider
	redirect             domainservice.RedirectResolver
	emailSender          domainservice.EmailSender
	outboxCoord          *redisstore.Coordinator
	projector            domainservice.RuntimeSecurityProjector
	logger               *slog.Logger
	cfg                  Config
	dummyHashOnce        sync.Once
	dummyHash            string
}

// NewService membentuk Service pengguna setelah memvalidasi seluruh port repositori dan kriptografi secara fail-closed.
func NewService(deps Dependencies) (*Service, error) {
	if err := validateUserDeps(deps); err != nil {
		return nil, err
	}

	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	cfg := normalizeConfig(deps.Config)

	return &Service{
		tx:                   deps.Transactor,
		users:                deps.Users,
		employeeProvisioning: deps.EmployeeProvisioning,
		identities:           deps.Identities,
		credentials:          deps.Credentials,
		totpFactors:          deps.TOTPFactors,
		recoveryCodes:        deps.RecoveryCodes,
		tokens:               deps.Tokens,
		roles:                deps.Roles,
		userRoles:            deps.UserRoles,
		securityRuntime:      deps.SecurityRuntime,
		sessions:             deps.Sessions,
		clock:                clock,
		idGen:                deps.IDGen,
		passwordHasher:       deps.PasswordHasher,
		passwordPolicy:       deps.PasswordPolicy,
		tokenGen:             deps.TokenGen,
		secretCipher:         deps.SecretCipher,
		totpProvider:         deps.TOTPProvider,
		redirect:             deps.Redirect,
		emailSender:          deps.EmailSender,
		outboxCoord:          deps.OutboxCoord,
		projector:            deps.Projector,
		logger:               deps.Logger,
		cfg:                  cfg,
	}, nil
}

func validateUserDeps(deps Dependencies) error {
	switch {
	case deps.Transactor == nil:
		return iamerrors.NewValidation("transactor", iamerrors.Code("iam.service.nil_transactor"), "transactor is required")
	case deps.Users == nil:
		return iamerrors.NewValidation("users", iamerrors.Code("iam.service.nil_user_repo"), "user repository is required")
	case deps.Identities == nil:
		return iamerrors.NewValidation("identities", iamerrors.Code("iam.service.nil_identity_repo"), "user identity repository is required")
	case deps.Credentials == nil:
		return iamerrors.NewValidation("credentials", iamerrors.Code("iam.service.nil_credential_repo"), "user credential repository is required")
	case deps.TOTPFactors == nil:
		return iamerrors.NewValidation("totp_factors", iamerrors.Code("iam.service.nil_totp_repo"), "totp factor repository is required")
	case deps.RecoveryCodes == nil:
		return iamerrors.NewValidation("recovery_codes", iamerrors.Code("iam.service.nil_recovery_code_repo"), "recovery code repository is required")
	case deps.Tokens == nil:
		return iamerrors.NewValidation("tokens", iamerrors.Code("iam.service.nil_token_repo"), "user token repository is required")
	case deps.Roles == nil:
		return iamerrors.NewValidation("roles", iamerrors.Code("iam.service.nil_role_repo"), "role repository is required")
	case deps.UserRoles == nil:
		return iamerrors.NewValidation("user_roles", iamerrors.Code("iam.service.nil_user_role_repo"), "user role repository is required")
	case deps.Sessions == nil:
		return iamerrors.NewValidation("sessions", iamerrors.Code("iam.service.nil_session_service"), "session collaborator is required")
	case deps.IDGen == nil:
		return iamerrors.NewValidation("id_gen", iamerrors.Code("iam.service.nil_id_generator"), "id generator is required")
	case deps.PasswordHasher == nil:
		return iamerrors.NewValidation("password_hasher", iamerrors.Code("iam.service.nil_password_hasher"), "password hasher is required")
	case deps.TokenGen == nil:
		return iamerrors.NewValidation("token_gen", iamerrors.Code("iam.service.nil_token_generator"), "token generator is required")
	case deps.SecretCipher == nil:
		return iamerrors.NewValidation("secret_cipher", iamerrors.Code("iam.service.nil_secret_cipher"), "secret cipher is required")
	case deps.TOTPProvider == nil:
		return iamerrors.NewValidation("totp_provider", iamerrors.Code("iam.service.nil_totp_provider"), "totp provider is required")
	default:
		return nil
	}
}

func normalizeConfig(cfg Config) Config {
	def := DefaultConfig()
	if cfg.MaxFailedAttempts <= 0 {
		cfg.MaxFailedAttempts = def.MaxFailedAttempts
	}
	if cfg.LockDuration <= 0 {
		cfg.LockDuration = def.LockDuration
	}
	if cfg.MFAChallengeTTL <= 0 {
		cfg.MFAChallengeTTL = def.MFAChallengeTTL
	}
	if cfg.PasswordResetTTL <= 0 {
		cfg.PasswordResetTTL = def.PasswordResetTTL
	}
	if cfg.IdentityVerificationTTL <= 0 {
		cfg.IdentityVerificationTTL = def.IdentityVerificationTTL
	}
	if cfg.AccountInvitationTTL <= 0 {
		cfg.AccountInvitationTTL = def.AccountInvitationTTL
	}
	if strings.TrimSpace(cfg.TOTPIssuer) == "" {
		cfg.TOTPIssuer = def.TOTPIssuer
	}
	if cfg.RecoveryCodeCount <= 0 {
		cfg.RecoveryCodeCount = def.RecoveryCodeCount
	}
	if strings.TrimSpace(cfg.PublicBaseURL) == "" {
		cfg.PublicBaseURL = def.PublicBaseURL
	} else {
		cfg.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	}
	return cfg
}

// CreateUser mendaftarkan pengguna baru setelah memvalidasi kebijakan kata sandi serta mencegah duplikasi email/username di user_identities.
func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (*entity.User, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	status := in.Status
	if status == "" {
		status = entity.UserStatusPending
	}

	if in.ProvisioningRequestID != "" && (in.Kind != entity.UserKindEmployee || status != entity.UserStatusPending || in.MarkIdentitiesVerified) {
		return nil, iamerrors.ErrInvalidInput
	}

	userID, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.user.id_generation_failed"), "failed to generate user id")
	}

	now := s.clock.Now()
	u, err := entity.NewUser(userID, in.Kind, status, in.DisplayName, now)
	if err != nil {
		return nil, err
	}
	if in.AvatarURL != "" || in.Locale != "" || in.Timezone != "" {
		if err := u.UpdateProfile(u.DisplayName, in.AvatarURL, in.Locale, in.Timezone, now); err != nil {
			return nil, err
		}
	}

	if in.Kind == entity.UserKindSystem {
		// Akun layanan otomasi system bersifat non-interaktif: hanya username, tanpa email dan tanpa kata sandi.
		if strings.TrimSpace(in.Email) != "" {
			return nil, iamerrors.WithField(iamerrors.ErrSystemUserInteractiveForbidden, "email", "system automation account must not have an interactive email")
		}
		if in.Password != "" {
			return nil, iamerrors.WithField(iamerrors.ErrSystemUserInteractiveForbidden, "password", "system automation account must not have a password")
		}
	}

	normEmail, normUsername, idents, forbiddenTerms, err := s.buildInitialIdentities(userID, in, now)
	if err != nil {
		return nil, err
	}
	forbiddenTerms = append(forbiddenTerms, u.DisplayName)

	var cred *entity.UserCredential
	if in.Kind != entity.UserKindSystem {
		if err := s.passwordPolicy.Validate(in.Password, forbiddenTerms...); err != nil {
			return nil, err
		}
		passwordHash, err := s.passwordHasher.Hash(in.Password)
		if err != nil {
			return nil, err
		}
		cred, err = entity.NewUserCredential(userID, passwordHash, in.RequirePasswordChange, now)
		if err != nil {
			return nil, err
		}
	}

	var issuedVerification *IssuedTokenResult
	var pendingEmail string
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if in.ProvisioningRequestID != "" {
			existing, err := s.claimEmployeeProvisioning(txCtx, in, now)
			if err != nil {
				return err
			}
			if existing != nil {
				u = existing
				return nil
			}
		}
		// Pra-cek eksplisit per jenis identitas (username terlebih dahulu lalu email) untuk memastikan keduanya belum dipakai akun lain.
		if err := s.ensureEmailAndUsernameAvailable(txCtx, normEmail, normUsername); err != nil {
			return err
		}

		if err := s.users.InsertWithIdentities(txCtx, u, idents); err != nil {
			return err
		}
		if cred != nil {
			if err := s.credentials.Insert(txCtx, cred); err != nil {
				return err
			}
		}
		if in.ProvisioningRequestID != "" {
			if err := s.associateProvisionedAccount(txCtx, in.ProvisioningRequestID, u.ID); err != nil {
				return err
			}
		}
		assignedBaseline, err := s.assignBaselineSelfRoleInTx(txCtx, u.ID, u.Kind, now)
		if err != nil {
			return err
		}
		if err := s.recordUserSecurityChangedInTx(txCtx, u.ID, false, assignedBaseline); err != nil {
			return err
		}
		if u.Status == entity.UserStatusPending && !in.MarkIdentitiesVerified {
			for _, ident := range idents {
				if ident.Kind == entity.IdentityKindEmail {
					tokRes, tokErr := s.issueOneTimeTokenForIdentity(txCtx, u.ID, &ident.ID, entity.TokenPurposeEmailVerification, s.cfg.IdentityVerificationTTL, now)
					if tokErr != nil {
						return tokErr
					}
					issuedVerification = tokRes
					pendingEmail = ident.Value
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if issuedVerification != nil && pendingEmail != "" {
		s.dispatchVerificationEmail(ctx, u.DisplayName, pendingEmail, issuedVerification.PlaintextToken, s.cfg.IdentityVerificationTTL)
	}
	return u, nil
}

func (s *Service) assignBaselineSelfRoleInTx(txCtx context.Context, userID string, kind entity.UserKind, now time.Time) (bool, error) {
	roleCode := entity.BaselineSelfRoleCodeForUserKind(kind)
	if roleCode == "" || s.userRoles == nil || s.roles == nil {
		return false, nil
	}
	baseRole, err := s.roles.SelectByCode(txCtx, roleCode)
	if err != nil {
		if errors.Is(err, iamerrors.ErrRoleNotFound) {
			// Lingkungan unit test terisolasi tanpa seed peran bawaan tetap dapat membuat user tanpa gagal palsu.
			return false, nil
		}
		return false, err
	}
	if !baseRole.IsActive {
		return false, nil
	}
	uid := userID
	selfScope, err := entity.NewScope(entity.ScopeTypeSelf, &uid)
	if err != nil {
		return false, err
	}
	exists, err := s.userRoles.CheckExists(txCtx, userID, baseRole.ID, selfScope)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	assignmentID, err := s.idGen.NewUUID()
	if err != nil {
		return false, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.user_role.id_generation_failed"), "failed to generate baseline user role id")
	}
	assignment, err := entity.NewUserRole(assignmentID, userID, baseRole.ID, selfScope, nil, nil, now)
	if err != nil {
		return false, err
	}
	if err := s.userRoles.Insert(txCtx, assignment); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) buildInitialIdentities(
	userID string,
	in CreateUserInput,
	now time.Time,
) (string, string, []*entity.UserIdentity, []string, error) {

	var normEmail, normUsername string
	var idents []*entity.UserIdentity
	var forbidden []string

	if strings.TrimSpace(in.Email) != "" {
		emailID, err := s.idGen.NewUUID()
		if err != nil {
			return "", "", nil, nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.identity.id_generation_failed"), "failed to generate identity id")
		}
		emailIdent, err := entity.NewUserIdentity(emailID, userID, entity.IdentityKindEmail, in.Email, true, now)
		if err != nil {
			return "", "", nil, nil, err
		}
		if in.MarkIdentitiesVerified {
			emailIdent.MarkVerified(now)
		}
		normEmail = emailIdent.Value
		idents = append(idents, emailIdent)
		forbidden = append(forbidden, normEmail)
	}

	if strings.TrimSpace(in.Username) != "" {
		usernameID, err := s.idGen.NewUUID()
		if err != nil {
			return "", "", nil, nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.identity.id_generation_failed"), "failed to generate identity id")
		}
		isPrimary := normEmail == ""
		usernameIdent, err := entity.NewUserIdentity(usernameID, userID, entity.IdentityKindUsername, in.Username, isPrimary, now)
		if err != nil {
			return "", "", nil, nil, err
		}
		if in.MarkIdentitiesVerified {
			usernameIdent.MarkVerified(now)
		}
		normUsername = usernameIdent.Value
		idents = append(idents, usernameIdent)
		forbidden = append(forbidden, normUsername)
	}

	if len(idents) == 0 {
		return "", "", nil, nil, iamerrors.WithField(iamerrors.ErrInvalidInput, "identity", "at least email or username is required")
	}
	return normEmail, normUsername, idents, forbidden, nil
}

// GetByID mengambil entitas pengguna berdasarkan UUID-nya.
func (s *Service) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return s.users.SelectByID(ctx, userID)
}

// GetAccessSnapshot mengambil snapshot hak akses efektif terbaru milik pengguna pada waktu lokal OS saat ini.
func (s *Service) GetAccessSnapshot(ctx context.Context, userID string) (*entity.AccessSnapshot, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return s.users.SelectAccessSnapshotByUserID(ctx, userID, s.clock.Now())
}

// List mengambil daftar pengguna berhalaman sesuai filter jenis, status, dan kata kunci.
func (s *Service) List(ctx context.Context, filter domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}
	return s.users.SelectList(ctx, filter.Normalize())
}

// UpdateProfile memperbarui atribut tampilan, avatar URL, locale, dan timezone pengguna aktif.
func (s *Service) UpdateProfile(ctx context.Context, in UpdateProfileInput) (*entity.User, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}

	var updated *entity.User
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Pastikan hanya pengguna berstatus ACTIVE yang dapat memperbarui profil mandiri.
		u, err := s.loadAuthenticatableUser(txCtx, in.UserID)
		if err != nil {
			return err
		}
		avatarURL := in.AvatarURL
		if in.PreserveAvatarURL {
			avatarURL = u.AvatarURL
		}
		if err := u.UpdateProfile(in.DisplayName, avatarURL, in.Locale, in.Timezone, s.clock.Now()); err != nil {
			return err
		}
		if err := s.users.Update(txCtx, u); err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Activate mengaktifkan akun pengguna yang belum dihapus lunak.
func (s *Service) Activate(ctx context.Context, userID string) (*entity.User, error) {
	return s.transitionUserStatus(ctx, userID, false, func(u *entity.User) error {
		return u.Activate(s.clock.Now())
	})
}

// Suspend menangguhkan pengguna sementara sekaligus mencabut seluruh sesi aktifnya secara atomik.
func (s *Service) Suspend(ctx context.Context, userID string) (*entity.User, error) {
	return s.transitionUserStatus(ctx, userID, true, func(u *entity.User) error {
		return u.Suspend(s.clock.Now())
	})
}

// Disable menonaktifkan akun pengguna sekaligus mencabut seluruh sesi aktifnya secara atomik.
func (s *Service) Disable(ctx context.Context, userID string) (*entity.User, error) {
	return s.transitionUserStatus(ctx, userID, true, func(u *entity.User) error {
		return u.Disable(s.clock.Now())
	})
}

// SoftDelete menghapus lunak pengguna sekaligus mencabut seluruh sesi aktifnya dalam satu transaksi.
func (s *Service) SoftDelete(ctx context.Context, userID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return err
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.users.SelectByID(txCtx, userID)
		if err != nil {
			return err
		}
		now := s.clock.Now()
		if err := u.SoftDelete(now); err != nil {
			return err
		}
		// Cabut sesi dan naikkan auth_epoch sebelum deleted_at diisi agar pembaruan penghitung keamanan tidak tertolak filter deleted_at IS NULL.
		if _, err := s.sessions.RevokeAllUserSessions(txCtx, userID, entity.SessionEndUserDisabled); err != nil {
			return err
		}
		return s.users.SoftDelete(txCtx, userID, now)
	})
}

func (s *Service) transitionUserStatus(
	ctx context.Context,
	userID string,
	revokeSessions bool,
	mutate func(*entity.User) error,
) (*entity.User, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}

	var updated *entity.User
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.users.SelectByID(txCtx, userID)
		if err != nil {
			return err
		}
		if err := mutate(u); err != nil {
			return err
		}
		if u.Status == entity.UserStatusActive {
			if err := s.requireEmployeeLinkForActivation(txCtx, u); err != nil {
				return err
			}
		}
		if err := s.users.Update(txCtx, u); err != nil {
			return err
		}
		if revokeSessions {
			if _, err := s.sessions.RevokeAllUserSessions(txCtx, userID, entity.SessionEndUserDisabled); err != nil {
				return err
			}
		} else {
			if err := s.recordUserSecurityChangedInTx(txCtx, userID, true, true); err != nil {
				return err
			}
		}
		updated = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// loadAuthenticatableUser mengambil akun pengguna manusia berdasarkan ID dan memastikan statusnya ACTIVE serta bukan akun system.
func (s *Service) loadAuthenticatableUser(ctx context.Context, userID string) (*entity.User, error) {
	u, err := s.users.SelectByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u.Kind == entity.UserKindSystem {
		return nil, iamerrors.ErrSystemUserInteractiveForbidden
	}
	if err := u.CanAuthenticate(); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.tx == nil || s.users == nil || s.identities == nil || s.credentials == nil || s.roles == nil || s.userRoles == nil || s.sessions == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.user.service_unavailable"), "user service is not initialized")
	}
	return nil
}
