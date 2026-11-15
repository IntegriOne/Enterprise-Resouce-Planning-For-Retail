package user

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func userToDomain(m *userModel) (*entity.User, error) {
	if m == nil {
		return nil, iamerrors.ErrUserNotFound
	}
	kind, err := entity.ParseUserKind(m.UserKind)
	if err != nil {
		return nil, err
	}
	status, err := entity.ParseUserStatus(m.Status)
	if err != nil {
		return nil, err
	}
	authEpoch := m.AuthEpoch
	if authEpoch < 1 {
		authEpoch = 1
	}
	permRevision := m.PermRevision
	if permRevision < 1 {
		permRevision = 1
	}
	return &entity.User{
		ID:           strings.ToLower(strings.TrimSpace(m.ID)),
		Kind:         kind,
		Status:       status,
		DisplayName:  m.DisplayName,
		AvatarURL:    pgstore.DerefString(m.AvatarURL),
		Locale:       m.Locale,
		Timezone:     m.Timezone,
		AuthEpoch:    authEpoch,
		PermRevision: permRevision,
		LastLoginAt:  pgstore.NormalizeLocalTimePtr(m.LastLoginAt),
		CreatedAt:    pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:    pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:    pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}

func userToModel(u *entity.User) (userModel, error) {
	if u == nil {
		return userModel{}, iamerrors.NewValidation("user", iamerrors.Code("iam.user.nil"), "user must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", u.ID)
	if err != nil {
		return userModel{}, err
	}
	if !u.Kind.IsValid() {
		return userModel{}, iamerrors.ErrInvalidUserKind
	}
	if !u.Status.IsValid() {
		return userModel{}, iamerrors.ErrInvalidUserStatus
	}
	displayName := strings.TrimSpace(u.DisplayName)
	if displayName == "" {
		return userModel{}, iamerrors.ErrInvalidDisplayName
	}
	locale := strings.TrimSpace(u.Locale)
	if locale == "" {
		locale = entity.DefaultUserLocale
	}
	timezone := strings.TrimSpace(u.Timezone)
	if timezone == "" {
		timezone = entity.DefaultUserTimezone
	}
	authEpoch := u.AuthEpoch
	if authEpoch < 1 {
		authEpoch = 1
	}
	permRevision := u.PermRevision
	if permRevision < 1 {
		permRevision = 1
	}
	now := pgstore.MustLocalNow(u.CreatedAt)
	updatedAt := pgstore.MustLocalNow(u.UpdatedAt)
	return userModel{
		ID:           id,
		UserKind:     u.Kind.String(),
		Status:       u.Status.String(),
		DisplayName:  displayName,
		AvatarURL:    pgstore.NullableStringPtr(u.AvatarURL),
		Locale:       locale,
		Timezone:     timezone,
		AuthEpoch:    authEpoch,
		PermRevision: permRevision,
		LastLoginAt:  pgstore.NormalizeLocalTimePtr(u.LastLoginAt),
		CreatedAt:    now,
		UpdatedAt:    updatedAt,
		DeletedAt:    pgstore.NormalizeLocalTimePtr(u.DeletedAt),
	}, nil
}

func identityToDomain(m *userIdentityModel) (*entity.UserIdentity, error) {
	if m == nil {
		return nil, iamerrors.ErrIdentityNotFound
	}
	kind, err := entity.ParseIdentityKind(m.Kind)
	if err != nil {
		return nil, err
	}
	return &entity.UserIdentity{
		ID:         strings.ToLower(strings.TrimSpace(m.ID)),
		UserID:     strings.ToLower(strings.TrimSpace(m.UserID)),
		Kind:       kind,
		Value:      m.Value,
		IsVerified: m.IsVerified,
		IsPrimary:  m.IsPrimary,
		VerifiedAt: pgstore.NormalizeLocalTimePtr(m.VerifiedAt),
		CreatedAt:  pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:  pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}

func identityToModel(i *entity.UserIdentity) (userIdentityModel, error) {
	if i == nil {
		return userIdentityModel{}, iamerrors.NewValidation("identity", iamerrors.Code("iam.identity.nil"), "user identity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", i.ID)
	if err != nil {
		return userIdentityModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", i.UserID)
	if err != nil {
		return userIdentityModel{}, err
	}
	normalizedValue, err := entity.NormalizeIdentityValue(i.Kind, i.Value)
	if err != nil {
		return userIdentityModel{}, err
	}
	verifiedAt := pgstore.NormalizeLocalTimePtr(i.VerifiedAt)
	if i.IsVerified && verifiedAt == nil {
		return userIdentityModel{}, iamerrors.NewValidation("verified_at", iamerrors.Code("iam.identity.missing_verified_at"), "verified_at is required when identity is verified")
	}
	return userIdentityModel{
		ID:         id,
		UserID:     userID,
		Kind:       i.Kind.String(),
		Value:      normalizedValue,
		IsPrimary:  i.IsPrimary,
		IsVerified: i.IsVerified,
		VerifiedAt: verifiedAt,
		CreatedAt:  pgstore.MustLocalNow(i.CreatedAt),
		UpdatedAt:  pgstore.MustLocalNow(i.UpdatedAt),
	}, nil
}
