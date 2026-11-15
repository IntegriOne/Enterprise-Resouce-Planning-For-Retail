package session

import (
	"bytes"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func sessionToDomain(m *sessionModel) (*entity.Session, error) {
	if m == nil {
		return nil, iamerrors.ErrSessionNotFound
	}
	var endReason *entity.SessionEndReason
	if m.EndReason != nil {
		parsed, err := entity.ParseSessionEndReason(*m.EndReason)
		if err != nil {
			return nil, err
		}
		endReason = &parsed
	}
	sess, err := entity.RehydrateSession(
		strings.ToLower(strings.TrimSpace(m.ID)),
		strings.ToLower(strings.TrimSpace(m.UserID)),
		m.TokenDigest,
		m.CSRFTokenDigest,
		m.MFAVerified,
		pgstore.NormalizeLocalTimePtr(m.MFAVerifiedAt),
		pgstore.ParseIPAddressPtr(m.IPAddress),
		pgstore.DerefString(m.UserAgent),
		pgstore.NormalizeLocalTime(m.AuthenticatedAt),
		pgstore.NormalizeLocalTime(m.LastSeenAt),
		pgstore.NormalizeLocalTime(m.IdleExpiresAt),
		pgstore.NormalizeLocalTime(m.AbsoluteExpiresAt),
		pgstore.NormalizeLocalTimePtr(m.EndedAt),
		endReason,
		pgstore.NormalizeLocalTime(m.CreatedAt),
	)
	if err != nil {
		return nil, err
	}
	sess.SetSecurityMetadata(m.Generation, m.CSRFKeyVersion)
	return sess, nil
}

func sessionToModel(s *entity.Session) (sessionModel, error) {
	if s == nil {
		return sessionModel{}, iamerrors.NewValidation("session", iamerrors.Code("iam.session.nil"), "session must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", s.ID)
	if err != nil {
		return sessionModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", s.UserID)
	if err != nil {
		return sessionModel{}, err
	}
	tokenDigest := s.TokenDigest()
	if len(tokenDigest) == 0 {
		return sessionModel{}, iamerrors.NewValidation("token_digest", iamerrors.Code("iam.session.empty_token_digest"), "session token digest must not be empty")
	}
	csrfDigest := s.CSRFTokenDigest()
	if len(csrfDigest) == 0 {
		return sessionModel{}, iamerrors.NewValidation("csrf_token_digest", iamerrors.Code("iam.session.empty_csrf_digest"), "csrf token digest must not be empty")
	}

	var endReason *string
	if s.EndReason != nil {
		if !s.EndReason.IsValid() {
			return sessionModel{}, iamerrors.ErrInvalidSessionEndReason
		}
		r := s.EndReason.String()
		endReason = &r
	}

	gen := s.Generation
	if gen < 1 {
		gen = 1
	}
	csrfVer := s.CSRFKeyVersion
	if csrfVer < 1 {
		csrfVer = 1
	}
	return sessionModel{
		ID:                id,
		UserID:            userID,
		TokenDigest:       bytes.Clone(tokenDigest),
		CSRFTokenDigest:   bytes.Clone(csrfDigest),
		Generation:        gen,
		CSRFKeyVersion:    csrfVer,
		IPAddress:         pgstore.FormatIPAddressPtr(s.IPAddress),
		UserAgent:         pgstore.NullableStringPtr(s.UserAgent),
		AuthenticatedAt:   pgstore.MustLocalNow(s.AuthenticatedAt),
		LastSeenAt:        pgstore.MustLocalNow(s.LastActiveAt),
		IdleExpiresAt:     pgstore.NormalizeLocalTime(s.IdleExpiresAt),
		AbsoluteExpiresAt: pgstore.NormalizeLocalTime(s.AbsoluteExpiresAt),
		MFAVerified:       s.MFAVerified,
		MFAVerifiedAt:     pgstore.NormalizeLocalTimePtr(s.MFAVerifiedAt),
		EndedAt:           pgstore.NormalizeLocalTimePtr(s.EndedAt),
		EndReason:         endReason,
		CreatedAt:         pgstore.MustLocalNow(s.CreatedAt),
	}, nil
}
