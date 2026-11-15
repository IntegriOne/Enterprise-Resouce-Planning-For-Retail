package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// SessionRepository mendefinisikan operasi baca dan pembaruan siklus hidup sesi pada tabel sessions.
type SessionRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Session, error)
	SelectByTokenDigest(ctx context.Context, tokenDigest []byte) (*entity.Session, error)
	SelectList(ctx context.Context, filter SessionFilter) (PageResult[*entity.Session], error)
	Insert(ctx context.Context, session *entity.Session) error
	UpdateActivity(ctx context.Context, sessionID string, lastActiveAt, idleExpiresAt time.Time) error
	UpdateMFAVerified(ctx context.Context, sessionID string, verifiedAt time.Time, rotatedCSRFDigest []byte) error
	UpdateEnded(ctx context.Context, sessionID string, reason entity.SessionEndReason, endedAt time.Time) error
	UpdateEndedByUserID(ctx context.Context, userID string, reason entity.SessionEndReason, endedAt time.Time) (int64, error)
	DeleteEndedBefore(ctx context.Context, cutoff time.Time) (int64, error)
}
