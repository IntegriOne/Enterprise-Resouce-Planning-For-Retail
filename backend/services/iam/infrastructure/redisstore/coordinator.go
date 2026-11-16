package redisstore

import (
	"context"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

// OutboxPayload menyimpan metadata ringan untuk event outbox keamanan IAM di PostgreSQL.
type OutboxPayload struct {
	UserID          string   `json:"user_id,omitempty"`
	SessionID       string   `json:"session_id,omitempty"`
	TokenDigestsHex []string `json:"token_digests_hex,omitempty"`
	AffectedUserIDs []string `json:"affected_user_ids,omitempty"`
	InvalidateAuthz bool     `json:"invalidate_authz,omitempty"`
}

// Coordinator menjembatani penulisan outbox transaksional PG (pending) -> proyeksi cepat Redis -> penandaan effective.
type Coordinator struct {
	runtimeRepo domainrepo.SecurityRuntimeRepository
	projector   *Projector
	idGen       domainservice.IDGenerator
	clock       domainservice.Clock
	logger      *slog.Logger
	notifyCh    chan struct{}
}

// NewCoordinator membentuk koordinator outbox dan proyeksi keamanan IAM->Redis.
func NewCoordinator(
	runtimeRepo domainrepo.SecurityRuntimeRepository,
	projector *Projector,
	idGen domainservice.IDGenerator,
	clock domainservice.Clock,
	logger *slog.Logger,
) *Coordinator {
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Coordinator{
		runtimeRepo: runtimeRepo,
		projector:   projector,
		idGen:       idGen,
		clock:       clock,
		logger:      logger,
		notifyCh:    make(chan struct{}, 1),
	}
}

// RecordInTx mencatat event outbox berstatus pending di dalam transaksi database yang sama dengan mutasi keamanan.
func (c *Coordinator) RecordInTx(
	txCtx context.Context,
	eventType string,
	aggregateType string,
	aggregateID string,
	payload OutboxPayload,
) (string, error) {
	if c == nil || c.runtimeRepo == nil || c.idGen == nil {
		return "", iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.outbox.coordinator_unavailable"), "outbox coordinator is unavailable")
	}
	evID, err := c.idGen.NewUUID()
	if err != nil {
		return "", err
	}
	ev, err := entity.NewOutboxEvent(evID, eventType, aggregateType, aggregateID, payload, c.clock.Now())
	if err != nil {
		return "", err
	}
	if err := c.runtimeRepo.Insert(txCtx, ev); err != nil {
		return "", err
	}
	return evID, nil
}

// CompleteAfterCommit memproyeksikan mutasi ke Redis segera setelah commit PG lalu menandai event sebagai effective (atau jadwalkan retry bila gagal).
func (c *Coordinator) CompleteAfterCommit(ctx context.Context, eventID string, projectFn func(recEpoch int64) error) {
	if c == nil {
		return
	}
	if c.projector == nil || !c.projector.IsEnabled() {
		if eventID != "" && c.runtimeRepo != nil {
			_ = c.runtimeRepo.MarkEffective(ctx, eventID, c.clock.Now())
		}
		return
	}

	var recEpoch int64 = 1
	if c.runtimeRepo != nil {
		if loaded, err := c.runtimeRepo.GetRecoveryEpoch(ctx); err == nil && loaded >= 1 {
			recEpoch = loaded
		}
	}

	if err := projectFn(recEpoch); err != nil {
		c.logger.Warn("redis security projection failed after PG commit; retaining pending outbox",
			slog.String("event_id", eventID),
			slog.String("error", err.Error()),
		)
		if eventID != "" && c.runtimeRepo != nil {
			now := c.clock.Now()
			_ = c.runtimeRepo.MarkRetry(ctx, eventID, 1, now.Add(time.Second), err.Error())
		}
		c.triggerReconcile()
		return
	}

	if eventID != "" && c.runtimeRepo != nil {
		if err := c.runtimeRepo.MarkEffective(ctx, eventID, c.clock.Now()); err != nil {
			c.logger.Warn("failed to mark outbox event effective",
				slog.String("event_id", eventID),
				slog.String("error", err.Error()),
			)
		}
	}
}

// NotifyChannel mengembalikan channel bangun bagi background reconciler.
func (c *Coordinator) NotifyChannel() <-chan struct{} {
	if c == nil {
		return nil
	}
	return c.notifyCh
}

func (c *Coordinator) triggerReconcile() {
	if c == nil || c.notifyCh == nil {
		return
	}
	select {
	case c.notifyCh <- struct{}{}:
	default:
	}
}

// EncodeDigestsHex mengonversi potongan SHA-256 digest biner menjadi string hex lowercase untuk disimpan aman di payload JSON outbox.
func EncodeDigestsHex(digests [][]byte) []string {
	if len(digests) == 0 {
		return nil
	}
	out := make([]string, 0, len(digests))
	for _, d := range digests {
		if len(d) > 0 {
			out = append(out, hex.EncodeToString(d))
		}
	}
	return out
}
