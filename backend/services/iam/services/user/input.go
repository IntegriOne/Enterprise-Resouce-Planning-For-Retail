package user

import (
	"log/slog"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
)

// Config menyimpan parameter kebijakan aplikasi untuk autentikasi, lockout, token, dan 2FA pengguna.
type Config struct {
	MaxFailedAttempts       int
	LockDuration            time.Duration
	MFAChallengeTTL         time.Duration
	PasswordResetTTL        time.Duration
	IdentityVerificationTTL time.Duration
	AccountInvitationTTL    time.Duration
	TOTPIssuer              string
	RecoveryCodeCount       int
	PublicBaseURL           string
}

// DefaultConfig mengembalikan konfigurasi aman standar IAM IntegriOne.
func DefaultConfig() Config {
	return Config{
		MaxFailedAttempts:       entity.DefaultMaxFailedLoginAttempts,
		LockDuration:            entity.DefaultLoginLockDuration,
		MFAChallengeTTL:         5 * time.Minute,
		PasswordResetTTL:        15 * time.Minute,
		IdentityVerificationTTL: 24 * time.Hour,
		AccountInvitationTTL:    48 * time.Hour,
		TOTPIssuer:              "IntegriOne ERP",
		RecoveryCodeCount:       8,
		PublicBaseURL:           "https://erp.localhost:8443",
	}
}

// Dependencies mengelompokkan seluruh port repositori, kriptografi, kebijakan domain, dan layanan sesi.
type Dependencies struct {
	Transactor           domainrepo.Transactor
	Users                domainrepo.UserRepository
	EmployeeProvisioning domainrepo.EmployeeProvisioningRepository
	Identities           domainrepo.UserIdentityRepository
	Credentials          domainrepo.UserCredentialRepository
	TOTPFactors          domainrepo.UserTOTPFactorRepository
	RecoveryCodes        domainrepo.UserRecoveryCodeRepository
	Tokens               domainrepo.UserTokenRepository
	Roles                domainrepo.RoleRepository
	UserRoles            domainrepo.UserRoleRepository
	SecurityRuntime      domainrepo.SecurityRuntimeRepository
	Sessions             SessionCollaborator
	Clock                domainservice.Clock
	IDGen                domainservice.IDGenerator
	PasswordHasher       domainservice.PasswordHasher
	PasswordPolicy       domainservice.PasswordPolicy
	TokenGen             domainservice.TokenGenerator
	SecretCipher         domainservice.SecretCipher
	TOTPProvider         domainservice.TOTPProvider
	Redirect             domainservice.RedirectResolver
	EmailSender          domainservice.EmailSender
	OutboxCoord          *redisstore.Coordinator
	Projector            domainservice.RuntimeSecurityProjector
	Logger               *slog.Logger
	Config               Config
}

// CreateUserInput membawa data pendaftaran atau provisioning akun beserta identitas email/username awal.
type CreateUserInput struct {
	Kind                   entity.UserKind
	Status                 entity.UserStatus
	DisplayName            string
	AvatarURL              string
	Locale                 string
	Timezone               string
	Email                  string
	Username               string
	Password               string
	RequirePasswordChange  bool
	MarkIdentitiesVerified bool
	ProvisioningRequestID  string
	WorkerID               string
}

// UpdateProfileInput membawa pembaruan atribut profil tampilan pengguna.
type UpdateProfileInput struct {
	UserID            string
	DisplayName       string
	AvatarURL         string
	PreserveAvatarURL bool
	Locale            string
	Timezone          string
}

// AuthStatus merepresentasikan keluaran status alur login kata sandi maupun tantangan MFA.
type AuthStatus string

const (
	// AuthStatusAuthenticated menandakan autentikasi tuntas dan sesi aktif telah diterbitkan.
	AuthStatusAuthenticated AuthStatus = "authenticated"
	// AuthStatusMFARequired menandakan kata sandi valid namun verifikasi 2FA/recovery code masih dibutuhkan.
	AuthStatusMFARequired AuthStatus = "mfa_required"
)

// AuthenticatePasswordInput membawa kredensial login terpadu (email atau username) dan konteks jaringan klien.
type AuthenticatePasswordInput struct {
	Identifier string
	Password   string
	IPAddress  string
	UserAgent  string
}

// AuthenticateResult memuat hasil autentikasi baik sesi terbit ataupun token tantangan MFA sementara.
type AuthenticateResult struct {
	Status                AuthStatus
	User                  *entity.User
	IssuedSession         *session.IssuedSession
	RedirectPath          string
	MustChangePassword    bool
	MFAChallengeToken     string
	MFAChallengeExpiresAt *time.Time
}

// VerifyMFAChallengeInput membawa token tantangan MFA sementara dan kode OTP 6 digit.
type VerifyMFAChallengeInput struct {
	ChallengeToken string
	Code           string
	IPAddress      string
	UserAgent      string
}

// VerifyRecoveryCodeInput membawa token tantangan MFA sementara dan salah satu kode cadangan pemulihan 2FA.
type VerifyRecoveryCodeInput struct {
	ChallengeToken string
	RecoveryCode   string
	IPAddress      string
	UserAgent      string
}

// ChangePasswordInput membawa permintaan penggantian kata sandi mandiri pengguna terautentikasi.
type ChangePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

// ChangePasswordResult memuat ringkasan hasil penggantian kata sandi mandiri pengguna.
type ChangePasswordResult struct {
	Changed           bool
	SessionsRevoked   int64
	PasswordChangedAt time.Time
}

// UpdateEmailInput membawa permintaan perubahan email dengan penyimpanan pending email sampai diverifikasi.
type UpdateEmailInput struct {
	UserID          string
	NewEmail        string
	CurrentPassword string
}

// UpdateEmailResult membawa identitas pending email dan token verifikasi aktif tunggal yang diterbitkan.
type UpdateEmailResult struct {
	PendingIdentity   *entity.UserIdentity
	VerificationToken string
	ExpiresAt         time.Time
}

// UpdateUsernameInput membawa permintaan perubahan atau penambahan username pengguna tanpa memicu ulang verifikasi email.
type UpdateUsernameInput struct {
	UserID   string
	Username string
}

// ResendValidationEmailInput mendukung kirim ulang token verifikasi baik via sesi login maupun email registrasi pending.
type ResendValidationEmailInput struct {
	UserID string
	Email  string
}

// ResendValidationEmailResult membawa status pengiriman ulang dan metadata token aktif bila diterbitkan.
type ResendValidationEmailResult struct {
	IsResend          bool
	PendingEmail      string
	VerificationToken string
	ExpiresAt         *time.Time
}

// ConfirmPasswordResetInput membawa token pemulihan sekali pakai beserta kata sandi baru.
type ConfirmPasswordResetInput struct {
	RawToken    string
	NewPassword string
}

// AddIdentityInput membawa penambahan pengenal login email atau username baru untuk pengguna.
type AddIdentityInput struct {
	UserID    string
	Kind      entity.IdentityKind
	Value     string
	IsPrimary bool
}

// IssuedTokenResult membawa entitas token tersimpan bersama token polos sekali tampil untuk dikirim ke pengguna.
type IssuedTokenResult struct {
	Token          *entity.UserToken
	PlaintextToken string
}

// EnrollTOTPResult membawa ID faktor TOTP tersimpan beserta URI QR Google Authenticator.
type EnrollTOTPResult struct {
	FactorID   string
	OTPAuthURI string
}

// ConfirmTOTPResult mengembalikan faktor TOTP yang aktif bersama kode cadangan pemulihan sekali tampil.
type ConfirmTOTPResult struct {
	Factor        *entity.UserTOTPFactor
	RecoveryCodes []string
}
