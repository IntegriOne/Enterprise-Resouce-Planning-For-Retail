package handler

import (
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

func toProtoUser(u *entity.User) *iamv1.User {
	if u == nil {
		return nil
	}
	return &iamv1.User{
		Id:           u.ID,
		Kind:         toProtoUserKind(u.Kind),
		Status:       toProtoUserStatus(u.Status),
		DisplayName:  u.DisplayName,
		AvatarUrl:    u.AvatarURL,
		Locale:       u.Locale,
		Timezone:     u.Timezone,
		LastLoginAt:  toProtoTimestampPtr(u.LastLoginAt),
		CreatedAt:    toProtoTimestamp(u.CreatedAt),
		UpdatedAt:    toProtoTimestamp(u.UpdatedAt),
		DeletedAt:    toProtoTimestampPtr(u.DeletedAt),
		AuthEpoch:    u.AuthEpoch,
		PermRevision: u.PermRevision,
	}
}

func toProtoUsers(items []*entity.User) []*iamv1.User {
	out := make([]*iamv1.User, 0, len(items))
	for _, u := range items {
		out = append(out, toProtoUser(u))
	}
	return out
}

func toProtoUserIdentity(id *entity.UserIdentity) *iamv1.UserIdentity {
	if id == nil {
		return nil
	}
	return &iamv1.UserIdentity{
		Id:         id.ID,
		UserId:     id.UserID,
		Kind:       toProtoIdentityKind(id.Kind),
		Value:      id.Value,
		IsVerified: id.IsVerified,
		IsPrimary:  id.IsPrimary,
		VerifiedAt: toProtoTimestampPtr(id.VerifiedAt),
		CreatedAt:  toProtoTimestamp(id.CreatedAt),
		UpdatedAt:  toProtoTimestamp(id.UpdatedAt),
	}
}

func toProtoUserIdentities(items []*entity.UserIdentity) []*iamv1.UserIdentity {
	out := make([]*iamv1.UserIdentity, 0, len(items))
	for _, id := range items {
		out = append(out, toProtoUserIdentity(id))
	}
	return out
}

func toProtoIssuedToken(res *user.IssuedTokenResult) *iamv1.IssuedToken {
	if res == nil {
		return nil
	}
	var tok *iamv1.UserToken
	if res.Token != nil {
		tok = &iamv1.UserToken{
			Id:         res.Token.ID,
			UserId:     res.Token.UserID,
			Purpose:    toProtoTokenPurpose(res.Token.Purpose),
			ExpiresAt:  toProtoTimestamp(res.Token.ExpiresAt),
			UsedAt:     toProtoTimestampPtr(res.Token.UsedAt),
			CreatedAt:  toProtoTimestamp(res.Token.CreatedAt),
			IdentityId: res.Token.IdentityID,
			Status:     toProtoTokenStatus(res.Token.Status),
		}
	}
	return &iamv1.IssuedToken{
		Token:          tok,
		PlaintextToken: res.PlaintextToken,
	}
}

func toProtoSession(s *entity.Session) *iamv1.Session {
	if s == nil {
		return nil
	}
	var ipAddr string
	if s.IPAddress != nil {
		ipAddr = s.IPAddress.String()
	}
	var endReason iamv1.SessionEndReason
	if s.EndReason != nil {
		endReason = toProtoSessionEndReason(*s.EndReason)
	}
	return &iamv1.Session{
		Id:                s.ID,
		UserId:            s.UserID,
		MfaVerified:       s.MFAVerified,
		MfaVerifiedAt:     toProtoTimestampPtr(s.MFAVerifiedAt),
		IpAddress:         ipAddr,
		UserAgent:         s.UserAgent,
		AuthenticatedAt:   toProtoTimestamp(s.AuthenticatedAt),
		LastActiveAt:      toProtoTimestamp(s.LastActiveAt),
		IdleExpiresAt:     toProtoTimestamp(s.IdleExpiresAt),
		AbsoluteExpiresAt: toProtoTimestamp(s.AbsoluteExpiresAt),
		EndedAt:           toProtoTimestampPtr(s.EndedAt),
		EndReason:         endReason,
		CreatedAt:         toProtoTimestamp(s.CreatedAt),
		Generation:        s.Generation,
		CsrfKeyVersion:    int32(s.CSRFKeyVersion),
	}
}

func toProtoSessions(items []*entity.Session) []*iamv1.Session {
	out := make([]*iamv1.Session, 0, len(items))
	for _, s := range items {
		out = append(out, toProtoSession(s))
	}
	return out
}

func toProtoIssuedSession(is *session.IssuedSession) *iamv1.IssuedSession {
	if is == nil {
		return nil
	}
	return &iamv1.IssuedSession{
		Session:      toProtoSession(is.Session),
		SessionToken: is.SessionToken,
		CsrfToken:    is.CSRFToken,
	}
}

func toProtoTOTPFactor(f *entity.UserTOTPFactor) *iamv1.UserTotpFactor {
	if f == nil {
		return nil
	}
	var lastStep *int64
	if f.LastUsedStep != nil {
		v := *f.LastUsedStep
		lastStep = &v
	}
	return &iamv1.UserTotpFactor{
		Id:           f.ID,
		UserId:       f.UserID,
		KeyVersion:   int32(f.KeyVersion),
		IsEnabled:    f.IsEnabled,
		ConfirmedAt:  toProtoTimestampPtr(f.ConfirmedAt),
		LastUsedStep: lastStep,
		CreatedAt:    toProtoTimestamp(f.CreatedAt),
		UpdatedAt:    toProtoTimestamp(f.UpdatedAt),
	}
}

func toProtoAuthStatus(s user.AuthStatus) iamv1.AuthStatus {
	switch s {
	case user.AuthStatusAuthenticated:
		return iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED
	case user.AuthStatusMFARequired:
		return iamv1.AuthStatus_AUTH_STATUS_MFA_REQUIRED
	default:
		return iamv1.AuthStatus_AUTH_STATUS_UNSPECIFIED
	}
}

func toProtoVerifyEmailStatus(s user.VerifyEmailStatus) iamv1.VerifyEmailStatus {
	switch s {
	case user.VerifyEmailStatusVerified:
		return iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_VERIFIED
	case user.VerifyEmailStatusExpired:
		return iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_EXPIRED
	case user.VerifyEmailStatusAlreadyUsed:
		return iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_ALREADY_USED
	case user.VerifyEmailStatusInvalid:
		return iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_INVALID
	default:
		return iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_UNSPECIFIED
	}
}
