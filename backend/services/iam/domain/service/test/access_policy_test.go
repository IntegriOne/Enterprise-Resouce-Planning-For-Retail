package service_test

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

func TestAccessEvaluatorAuthorize(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	userID := "11111111-1111-4111-8111-111111111111"
	otherUserID := "99999999-9999-4999-8999-999999999999"
	facID := "22222222-2222-4222-8222-222222222222"
	otherFacID := "33333333-3333-4333-8333-333333333333"

	user, _ := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Budi", now)
	suspendedUser, _ := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusSuspended, "Budi", now)

	tok := sha256.Sum256([]byte("tok"))
	csrf := sha256.Sum256([]byte("csrf"))
	sess, _ := entity.NewSession("44444444-4444-4444-8444-444444444444", userID, tok[:], csrf[:], true, nil, "ua", now)

	facScope, _ := entity.NewScope(entity.ScopeTypeFacility, &facID)
	otherFacScope, _ := entity.NewScope(entity.ScopeTypeFacility, &otherFacID)

	snapshot := &entity.AccessSnapshot{
		UserID:     userID,
		UserKind:   entity.UserKindEmployee,
		UserStatus: entity.UserStatusActive,
		MFAEnabled: true,
		Permissions: []entity.ScopedPermission{
			{PermissionCode: "wms.stock.read", IsSensitive: false, Scope: facScope},
			{PermissionCode: "iam.role.grant", IsSensitive: true, Scope: entity.GlobalScope()},
		},
		CheckedAt: now,
	}

	eval := service.NewAccessEvaluator(service.NewDefaultSessionPolicy())

	t.Run("grants valid non-sensitive scoped permission", func(t *testing.T) {
		t.Parallel()
		err := eval.Authorize(service.AuthorizeInput{
			User:               user,
			Session:            sess,
			Snapshot:           snapshot,
			RequiredPermission: "wms.stock.read",
			RequiredScope:      facScope,
			Now:                now.Add(15 * time.Second),
		})
		if err != nil {
			t.Fatalf("Authorize() unexpected error: %v", err)
		}
	})

	t.Run("rejects suspended user fail-closed", func(t *testing.T) {
		t.Parallel()
		err := eval.Authorize(service.AuthorizeInput{
			User:               suspendedUser,
			Session:            sess,
			Snapshot:           snapshot,
			RequiredPermission: "wms.stock.read",
			RequiredScope:      facScope,
			Now:                now.Add(15 * time.Second),
		})
		if !errors.Is(err, iamerrors.ErrUserSuspended) {
			t.Fatalf("Authorize() = %v, want ErrUserSuspended", err)
		}
	})

	t.Run("rejects stale snapshot beyond 30s fail-closed", func(t *testing.T) {
		t.Parallel()
		err := eval.Authorize(service.AuthorizeInput{
			User:               user,
			Session:            sess,
			Snapshot:           snapshot,
			RequiredPermission: "wms.stock.read",
			RequiredScope:      facScope,
			Now:                now.Add(35 * time.Second),
		})
		if !errors.Is(err, iamerrors.ErrSnapshotStale) {
			t.Fatalf("Authorize() = %v, want ErrSnapshotStale", err)
		}
	})

	t.Run("rejects mismatched scope", func(t *testing.T) {
		t.Parallel()
		err := eval.Authorize(service.AuthorizeInput{
			User:               user,
			Session:            sess,
			Snapshot:           snapshot,
			RequiredPermission: "wms.stock.read",
			RequiredScope:      otherFacScope,
			Now:                now.Add(10 * time.Second),
		})
		if !errors.Is(err, iamerrors.ErrPermissionDenied) {
			t.Fatalf("Authorize() = %v, want ErrPermissionDenied", err)
		}
	})

	t.Run("rejects sensitive permission when session authentication is older than reauth window", func(t *testing.T) {
		t.Parallel()
		freshSnap := snapshot.Clone()
		checkTime := now.Add(12 * time.Minute)
		freshSnap.CheckedAt = checkTime
		err := eval.Authorize(service.AuthorizeInput{
			User:               user,
			Session:            sess,
			Snapshot:           &freshSnap,
			RequiredPermission: "iam.role.grant",
			RequiredScope:      entity.GlobalScope(),
			Now:                checkTime,
		})
		if !errors.Is(err, iamerrors.ErrSensitiveReauthRequired) {
			t.Fatalf("Authorize() = %v, want ErrSensitiveReauthRequired", err)
		}
	})

	t.Run("rejects cross-user session/snapshot mismatch", func(t *testing.T) {
		t.Parallel()
		wrongSnap := snapshot.Clone()
		wrongSnap.UserID = otherUserID
		err := eval.Authorize(service.AuthorizeInput{
			User:               user,
			Session:            sess,
			Snapshot:           &wrongSnap,
			RequiredPermission: "wms.stock.read",
			RequiredScope:      facScope,
			Now:                now.Add(10 * time.Second),
		})
		if !errors.Is(err, iamerrors.ErrForbidden) {
			t.Fatalf("Authorize() = %v, want ErrForbidden", err)
		}
	})

	t.Run("allows self-service sensitive reauth after step-up MFA without RBAC permission code", func(t *testing.T) {
		t.Parallel()
		steppedSess := *sess
		stepUpAt := now.Add(12 * time.Minute)
		if err := steppedSess.MarkMFAVerified(stepUpAt); err != nil {
			t.Fatalf("MarkMFAVerified() error = %v", err)
		}
		freshSnap := snapshot.Clone()
		freshSnap.CheckedAt = stepUpAt.Add(5 * time.Second)
		err := eval.Authorize(service.AuthorizeInput{
			User:                   user,
			Session:                &steppedSess,
			Snapshot:               &freshSnap,
			RequiredPermission:     "",
			RequiredScope:          entity.GlobalScope(),
			RequireSensitiveReauth: true,
			Now:                    stepUpAt.Add(10 * time.Second),
		})
		if err != nil {
			t.Fatalf("Authorize() unexpected error after MFA step-up = %v", err)
		}
	})
}
