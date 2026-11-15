package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// SecurityControlRepository mengelola epoch pemulihan global singleton pada tabel iam_security_control.
type SecurityControlRepository interface {
	GetRecoveryEpoch(ctx context.Context) (int64, error)
	IncrementRecoveryEpoch(ctx context.Context) (int64, error)
}

// OutboxRepository mengelola antrean event keamanan pending/effective/failed pada tabel iam_outbox_events.
type OutboxRepository interface {
	Insert(ctx context.Context, ev *entity.OutboxEvent) error
	SelectDuePending(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEvent, error)
	MarkEffective(ctx context.Context, id string, processedAt time.Time) error
	MarkRetry(ctx context.Context, id string, attemptCount int, nextRetryAt time.Time, errMsg string) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
	CountPending(ctx context.Context) (int64, error)
	CountFailed(ctx context.Context) (int64, error)
}

// SecurityRuntimeRepository menyediakan operasi atomik epoch keamanan, revisi izin, sesi pemulihan, dan outbox Redis.
type SecurityRuntimeRepository interface {
	SecurityControlRepository
	OutboxRepository
	IncrementUserSecurityCounters(ctx context.Context, userID string, bumpAuthEpoch, bumpPermRevision bool) (authEpoch int64, permRevision int64, err error)
	BumpPermRevisionForRole(ctx context.Context, roleID string) ([]string, error)
	BumpPermRevisionForGroup(ctx context.Context, groupID string) ([]string, error)
	SelectActiveSessionDigestsByUserID(ctx context.Context, userID string) ([][]byte, error)
	SelectActiveSessionsForRecovery(ctx context.Context, now time.Time, limit int) ([]*entity.Session, error)
	UpdateSessionSecurityMetadata(ctx context.Context, sessionID string, generation int64, csrfKeyVersion int, csrfDigest []byte) error
	ResolveUserRoleUserID(ctx context.Context, assignmentID string) (string, error)
	ResolveGroupRoleGroupID(ctx context.Context, assignmentID string) (string, error)
}
