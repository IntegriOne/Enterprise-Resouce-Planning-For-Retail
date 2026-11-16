package session

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func (s *Service) runInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if s.tx != nil {
		return s.tx.WithinTransaction(ctx, fn)
	}
	return fn(ctx)
}

func (s *Service) deriveCSRF(sessionID string, tokenDigest []byte, generation int64) (string, []byte, int, error) {
	if s.csrfDeriver != nil {
		ver := s.csrfDeriver.ActiveKeyVersion()
		plaintext, digest, err := s.csrfDeriver.DeriveToken(ver, sessionID, tokenDigest, generation)
		if err != nil {
			return "", nil, 0, err
		}
		return plaintext, digest, ver, nil
	}
	plaintext, digest, err := s.tokens.GenerateOpaqueToken(opaqueTokenBytes)
	if err != nil {
		return "", nil, 0, err
	}
	return plaintext, digest, 1, nil
}

func (s *Service) recordSessionIssuedInTx(txCtx context.Context, sess *entity.Session) error {
	if s.outboxCoord == nil && s.projector == nil {
		return nil
	}
	payload := redisstore.OutboxPayload{
		UserID:          sess.UserID,
		SessionID:       sess.ID,
		TokenDigestsHex: redisstore.EncodeDigestsHex([][]byte{sess.TokenDigest()}),
	}
	var evID string
	if s.outboxCoord != nil {
		var err error
		evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventSessionIssued, "session", sess.ID, payload)
		if err != nil {
			return err
		}
	}
	u, uErr := s.users.SelectByID(txCtx, sess.UserID)
	sessCopy := *sess
	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		if uErr != nil {
			loaded, err := s.users.SelectByID(postCtx, sessCopy.UserID)
			if err != nil {
				return
			}
			u = loaded
		}
		if s.outboxCoord != nil {
			s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
				if s.projector == nil {
					return nil
				}
				return s.projector.ProjectSessionIssued(postCtx, u, &sessCopy, recEpoch)
			})
			return
		}
		if s.projector != nil && s.projector.IsEnabled() {
			_ = s.projector.ProjectSessionIssued(postCtx, u, &sessCopy, 1)
		}
	})
	return nil
}

func (s *Service) recordSessionUpdatedInTx(txCtx context.Context, sess *entity.Session) error {
	if s.outboxCoord == nil && s.projector == nil {
		return nil
	}
	payload := redisstore.OutboxPayload{
		UserID:    sess.UserID,
		SessionID: sess.ID,
	}
	var evID string
	if s.outboxCoord != nil {
		var err error
		evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventSessionUpdated, "session", sess.ID, payload)
		if err != nil {
			return err
		}
	}
	u, uErr := s.users.SelectByID(txCtx, sess.UserID)
	sessCopy := *sess
	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		if uErr != nil {
			loaded, err := s.users.SelectByID(postCtx, sessCopy.UserID)
			if err != nil {
				return
			}
			u = loaded
		}
		if s.outboxCoord != nil {
			s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
				if s.projector == nil {
					return nil
				}
				return s.projector.ProjectSessionUpdated(postCtx, u, &sessCopy, recEpoch)
			})
			return
		}
		if s.projector != nil && s.projector.IsEnabled() {
			_ = s.projector.ProjectSessionUpdated(postCtx, u, &sessCopy, 1)
		}
	})
	return nil
}

func (s *Service) recordSessionRevokedInTx(txCtx context.Context, sess *entity.Session) error {
	if s.outboxCoord == nil && s.projector == nil {
		return nil
	}
	digestCopy := make([]byte, len(sess.TokenDigest()))
	copy(digestCopy, sess.TokenDigest())
	payload := redisstore.OutboxPayload{
		UserID:          sess.UserID,
		SessionID:       sess.ID,
		TokenDigestsHex: redisstore.EncodeDigestsHex([][]byte{digestCopy}),
	}
	var evID string
	if s.outboxCoord != nil {
		var err error
		evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventSessionRevoked, "session", sess.ID, payload)
		if err != nil {
			return err
		}
	}
	u, uErr := s.users.SelectByID(txCtx, sess.UserID)
	uid := sess.UserID
	pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
		if uErr != nil {
			loaded, _ := s.users.SelectByID(postCtx, uid)
			u = loaded
		}
		if s.outboxCoord != nil {
			s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
				if s.projector == nil {
					return nil
				}
				return s.projector.ProjectSessionRevoked(postCtx, u, digestCopy, recEpoch)
			})
			return
		}
		if s.projector != nil && s.projector.IsEnabled() {
			_ = s.projector.ProjectSessionRevoked(postCtx, u, digestCopy, 1)
		}
	})
	return nil
}

func (s *Service) revokeAllSessionsWithOutbox(txCtx context.Context, userID string, reason entity.SessionEndReason) (int64, error) {
	var digests [][]byte
	if s.securityRuntime != nil {
		loadedDigests, err := s.securityRuntime.SelectActiveSessionDigestsByUserID(txCtx, userID)
		if err != nil {
			return 0, err
		}
		digests = loadedDigests
	}

	revoked, err := s.sessions.UpdateEndedByUserID(txCtx, userID, reason, s.clock.Now())
	if err != nil {
		return 0, err
	}

	if s.securityRuntime != nil {
		if _, _, err := s.securityRuntime.IncrementUserSecurityCounters(txCtx, userID, true, false); err != nil {
			return 0, err
		}
	}

	if s.outboxCoord != nil || s.projector != nil {
		payload := redisstore.OutboxPayload{
			UserID:          userID,
			TokenDigestsHex: redisstore.EncodeDigestsHex(digests),
		}
		var evID string
		if s.outboxCoord != nil {
			evID, err = s.outboxCoord.RecordInTx(txCtx, entity.OutboxEventUserSessionsRevoked, "user", userID, payload)
			if err != nil {
				return 0, err
			}
		}
		var txSnapshot *entity.User
		if s.users != nil {
			txSnapshot, _ = s.users.SelectByID(txCtx, userID)
		}
		pgstore.RegisterAfterCommit(txCtx, func(postCtx context.Context) {
			var u *entity.User
			if s.users != nil {
				u, _ = s.users.SelectByID(postCtx, userID)
			}
			// Bila pengguna dihapus lunak sebelum commit, pertahankan snapshot in-tx berstatus disabled untuk proyeksi keamanan Redis.
			if u == nil && txSnapshot != nil {
				cp := *txSnapshot
				cp.Status = entity.UserStatusDisabled
				u = &cp
			}
			if s.outboxCoord != nil {
				s.outboxCoord.CompleteAfterCommit(postCtx, evID, func(recEpoch int64) error {
					if s.projector == nil {
						return nil
					}
					return s.projector.ProjectUserSessionsRevoked(postCtx, u, digests, recEpoch)
				})
				return
			}
			if s.projector != nil && s.projector.IsEnabled() {
				_ = s.projector.ProjectUserSessionsRevoked(postCtx, u, digests, 1)
			}
		})
	}
	return revoked, nil
}

// GetSessionCSRF mengekstrak ulang token CSRF deterministik milik sesi aktif untuk sinkronisasi antar-tab aman.
func (s *Service) GetSessionCSRF(ctx context.Context, sessionID string) (*CSRFTokenResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("session_id", sessionID); err != nil {
		return nil, err
	}
	if s.csrfDeriver == nil {
		return nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.csrf_deriver_unavailable"), "deterministic csrf deriver is not configured")
	}

	sess, err := s.sessions.SelectByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := sess.CheckStatus(now); err != nil {
		s.recordNaturalExpiryIfApplicable(ctx, sess.ID, err, now)
		return nil, err
	}
	if err := s.ensureSessionOwnerActive(ctx, sess); err != nil {
		return nil, err
	}

	keyVer := sess.CSRFKeyVersion
	if keyVer <= 0 {
		keyVer = 1
	}
	plaintext, digest, err := s.csrfDeriver.DeriveToken(keyVer, sess.ID, sess.TokenDigest(), sess.Generation)
	if err != nil {
		return nil, err
	}
	if !sess.MatchesCSRFDigest(digest) {
		return nil, iamerrors.ErrCSRFMismatch
	}
	return &CSRFTokenResult{
		Session:   sess,
		CSRFToken: plaintext,
	}, nil
}

// RefreshAccessSnapshot menghitung AccessSnapshot + izin dasbor /identity lalu memproyeksikannya ke Redis.
func (s *Service) RefreshAccessSnapshot(ctx context.Context, userID string) (*RefreshSnapshotResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}

	u, err := s.users.SelectByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u.IsDeleted() {
		return nil, iamerrors.ErrUserDeleted
	}

	now := s.clock.Now()
	snapshot, err := s.users.SelectAccessSnapshotByUserID(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	defaultRedirect, err := s.redirect.ResolvePostLoginPath(u.Kind, snapshot.Roles)
	if err != nil {
		return nil, err
	}

	if s.securityRuntime != nil {
		recEpoch, err := s.securityRuntime.GetRecoveryEpoch(ctx)
		if err != nil {
			return nil, err
		}
		if recEpoch < 1 {
			return nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.security.invalid_recovery_epoch"), "security recovery epoch is unavailable")
		}
		snapshot.RecoveryEpoch = recEpoch
	}

	canAccessIdentity := evaluateIdentityDashboardAccess(u, snapshot, defaultRedirect)
	allowedRoutes := s.redirect.AllowedPaths()

	if s.projector != nil && s.projector.IsEnabled() {
		if err := s.projector.ProjectAccessSnapshot(ctx, snapshot, canAccessIdentity, defaultRedirect, allowedRoutes); err != nil {
			return nil, err
		}
	}

	return &RefreshSnapshotResult{
		User:                 u,
		Snapshot:             snapshot,
		CanAccessIdentity:    canAccessIdentity,
		DefaultRedirectPath:  defaultRedirect,
		AllowedRedirectPaths: allowedRoutes,
	}, nil
}

func evaluateIdentityDashboardAccess(u *entity.User, snapshot *entity.AccessSnapshot, defaultRedirect string) bool {
	if u == nil || snapshot == nil {
		return false
	}
	if u.Status != entity.UserStatusActive || u.IsDeleted() {
		return false
	}
	if u.Kind == entity.UserKindCustomer {
		return false
	}
	if strings.TrimSpace(defaultRedirect) == "/identity" {
		return true
	}
	for _, perm := range snapshot.Permissions {
		// Izin self-service (*self) dimiliki seluruh karyawan/mitra untuk akun pribadi dan tidak memberi akses ke dashboard admin /identity.
		if entity.IsSelfServicePermissionCode(perm.PermissionCode) {
			continue
		}
		if strings.HasPrefix(perm.PermissionCode, "iam.") {
			return true
		}
	}
	return false
}
