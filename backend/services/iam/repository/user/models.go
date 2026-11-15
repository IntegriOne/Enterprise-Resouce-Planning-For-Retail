// Package user mengimplementasikan repositori PostgreSQL GORM untuk pengguna, identitas login, kredensial, TOTP, recovery codes, dan token.
package user

import "time"

type userModel struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey"`
	UserKind     string     `gorm:"column:user_kind"`
	Status       string     `gorm:"column:status"`
	DisplayName  string     `gorm:"column:display_name"`
	AvatarURL    *string    `gorm:"column:avatar_url"`
	Locale       string     `gorm:"column:locale"`
	Timezone     string     `gorm:"column:timezone"`
	AuthEpoch    int64      `gorm:"column:auth_epoch"`
	PermRevision int64      `gorm:"column:perm_revision"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (userModel) TableName() string {
	return "users"
}

type userIdentityModel struct {
	ID         string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID     string     `gorm:"column:user_id;type:uuid"`
	Kind       string     `gorm:"column:kind"`
	Value      string     `gorm:"column:value"`
	IsPrimary  bool       `gorm:"column:is_primary"`
	IsVerified bool       `gorm:"column:is_verified"`
	VerifiedAt *time.Time `gorm:"column:verified_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
}

func (userIdentityModel) TableName() string {
	return "user_identities"
}

type userCredentialModel struct {
	ID                  string     `gorm:"column:id;type:uuid;primaryKey;->"`
	UserID              string     `gorm:"column:user_id;type:uuid"`
	PasswordHash        string     `gorm:"column:password_hash"`
	PasswordChangedAt   time.Time  `gorm:"column:password_changed_at"`
	MustChangePassword  bool       `gorm:"column:must_change_password"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts"`
	LockedUntil         *time.Time `gorm:"column:locked_until"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (userCredentialModel) TableName() string {
	return "user_credentials"
}

type userTOTPFactorModel struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID          string     `gorm:"column:user_id;type:uuid"`
	EncryptedSecret []byte     `gorm:"column:encrypted_secret"`
	KeyVersion      int16      `gorm:"column:key_version"`
	IsEnabled       bool       `gorm:"column:is_enabled"`
	ConfirmedAt     *time.Time `gorm:"column:confirmed_at"`
	LastUsedStep    *int64     `gorm:"column:last_used_step"`
	LastUsedAt      *time.Time `gorm:"column:last_used_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (userTOTPFactorModel) TableName() string {
	return "user_totp_factors"
}

type userRecoveryCodeModel struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string     `gorm:"column:user_id;type:uuid"`
	CodeHash  string     `gorm:"column:code_hash"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (userRecoveryCodeModel) TableName() string {
	return "user_recovery_codes"
}

type userTokenModel struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID      string     `gorm:"column:user_id;type:uuid"`
	IdentityID  *string    `gorm:"column:identity_id;type:uuid"`
	Purpose     string     `gorm:"column:purpose"`
	Status      string     `gorm:"column:status"`
	TokenDigest []byte     `gorm:"column:token_digest"`
	ExpiresAt   time.Time  `gorm:"column:expires_at"`
	ConsumedAt  *time.Time `gorm:"column:consumed_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (userTokenModel) TableName() string {
	return "user_tokens"
}
