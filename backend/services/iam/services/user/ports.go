package user

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
)

// SessionCollaborator mengabstraksikan operasi penerbitan, pembaruan MFA, dan pencabutan sesi milik layanan sesi.
type SessionCollaborator interface {
	Issue(ctx context.Context, in session.IssueInput) (*session.IssuedSession, error)
	RevokeAllUserSessions(ctx context.Context, userID string, reason entity.SessionEndReason) (int64, error)
	MarkUserMFAVerified(ctx context.Context, userID, sessionID string) (*session.CSRFTokenResult, error)
}
