package entity_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const testUserID = "11111111-1111-4111-8111-111111111111"
const testIdentityID = "22222222-2222-4222-8222-222222222222"

func TestNewUserAndLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	nonLocal := time.Date(2026, 11, 14, 10, 0, 0, 0, time.FixedZone("TEST", 3600))
	user, err := entity.NewUser(testUserID, entity.UserKindEmployee, entity.UserStatusPending, "  Budi Santoso  ", nonLocal)
	if err != nil {
		t.Fatalf("NewUser() unexpected error: %v", err)
	}
	if user.DisplayName != "Budi Santoso" {
		t.Fatalf("DisplayName = %q, want trimmed Budi Santoso", user.DisplayName)
	}
	if user.Locale != entity.DefaultUserLocale || user.Timezone != entity.DefaultUserTimezone {
		t.Fatalf("Locale/Timezone = (%q, %q), want (%q, %q)", user.Locale, user.Timezone, entity.DefaultUserLocale, entity.DefaultUserTimezone)
	}
	if user.CreatedAt.Location() != time.Local {
		t.Fatalf("CreatedAt.Location() = %v, want %v", user.CreatedAt.Location(), time.Local)
	}
	if err := user.CanAuthenticate(); !errors.Is(err, iamerrors.ErrUserInactive) {
		t.Fatalf("pending user CanAuthenticate() = %v, want ErrUserInactive", err)
	}

	if err := user.Activate(now.Add(time.Minute)); err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	if err := user.CanAuthenticate(); err != nil {
		t.Fatalf("active user CanAuthenticate() = %v, want nil", err)
	}

	user.RecordLogin(now.Add(2 * time.Minute))
	if user.LastLoginAt == nil || !user.LastLoginAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("LastLoginAt = %v, want recorded login time", user.LastLoginAt)
	}

	if err := user.Suspend(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("Suspend() unexpected error: %v", err)
	}
	if err := user.CanAuthenticate(); !errors.Is(err, iamerrors.ErrUserSuspended) {
		t.Fatalf("suspended user CanAuthenticate() = %v, want ErrUserSuspended", err)
	}

	if err := user.Disable(now.Add(4 * time.Minute)); err != nil {
		t.Fatalf("Disable() unexpected error: %v", err)
	}
	if err := user.CanAuthenticate(); !errors.Is(err, iamerrors.ErrUserDisabled) {
		t.Fatalf("disabled user CanAuthenticate() = %v, want ErrUserDisabled", err)
	}

	if err := user.SoftDelete(now.Add(5 * time.Minute)); err != nil {
		t.Fatalf("SoftDelete() unexpected error: %v", err)
	}
	if !user.IsDeleted() {
		t.Fatal("IsDeleted() = false, want true")
	}
	if err := user.CanAuthenticate(); !errors.Is(err, iamerrors.ErrUserDeleted) {
		t.Fatalf("deleted user CanAuthenticate() = %v, want ErrUserDeleted", err)
	}
	if err := user.Activate(now.Add(6 * time.Minute)); !errors.Is(err, iamerrors.ErrUserDeleted) {
		t.Fatalf("Activate() after SoftDelete() = %v, want ErrUserDeleted", err)
	}

	sysUser, err := entity.NewUser(testUserID, entity.UserKindSystem, entity.UserStatusActive, "Auto System IAM", now)
	if err != nil {
		t.Fatalf("NewUser(system) unexpected error: %v", err)
	}
	if err := sysUser.CanAuthenticate(); !errors.Is(err, iamerrors.ErrSystemCannotLogin) || err.Error() != "system cannot login" {
		t.Fatalf("system user CanAuthenticate() = %v, want ErrSystemCannotLogin (system cannot login)", err)
	}
}

func TestUserValidationAndProfileUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	if _, err := entity.NewUser(testUserID, entity.UserKindEmployee, entity.UserStatusActive, "   ", now); !errors.Is(err, iamerrors.ErrInvalidDisplayName) {
		t.Fatalf("empty display name error = %v, want ErrInvalidDisplayName", err)
	}
	if _, err := entity.NewUser(testUserID, entity.UserKindEmployee, entity.UserStatusActive, strings.Repeat("a", 201), now); !errors.Is(err, iamerrors.ErrInvalidDisplayName) {
		t.Fatalf("too long display name error = %v, want ErrInvalidDisplayName", err)
	}

	user, err := entity.NewUser(testUserID, entity.UserKindEmployee, entity.UserStatusActive, "Budi", now)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}
	if err := user.UpdateProfile("Budi Pratama", "https://cdn.integrione.local/a.png", "en-US", "UTC", now.Add(time.Minute)); err != nil {
		t.Fatalf("UpdateProfile() unexpected error: %v", err)
	}
	if err := user.UpdateProfile("Budi Pratama", "https://cdn.integrione.local/a.png", "", "", now.Add(90*time.Second)); err != nil {
		t.Fatalf("UpdateProfile(empty locale/timezone) unexpected error: %v", err)
	}
	if user.Locale != "en-US" || user.Timezone != "UTC" {
		t.Fatalf("UpdateProfile(empty locale/timezone) overwrote values: (%q, %q)", user.Locale, user.Timezone)
	}
	if err := user.UpdateProfile("Budi", "javascript:alert(1)", "id-ID", "Asia/Jakarta", now.Add(2*time.Minute)); !errors.Is(err, iamerrors.ErrInvalidAvatarURL) {
		t.Fatalf("UpdateProfile() bad url error = %v, want ErrInvalidAvatarURL", err)
	}
	if err := user.UpdateProfile("Budi", "", strings.Repeat("x", 21), "Asia/Jakarta", now.Add(2*time.Minute)); !errors.Is(err, iamerrors.ErrInvalidLocale) {
		t.Fatalf("UpdateProfile() bad locale error = %v, want ErrInvalidLocale", err)
	}
	if err := user.UpdateProfile("Budi", "", "id-ID", strings.Repeat("x", 81), now.Add(2*time.Minute)); !errors.Is(err, iamerrors.ErrInvalidTimezone) {
		t.Fatalf("UpdateProfile() bad timezone error = %v, want ErrInvalidTimezone", err)
	}
}

func TestUserIdentity(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	emailID, err := entity.NewUserIdentity(testIdentityID, testUserID, entity.IdentityKindEmail, "  Budi.Santoso@IntegriOne.ID ", true, now)
	if err != nil {
		t.Fatalf("NewUserIdentity(email) error = %v", err)
	}
	if emailID.Value != "budi.santoso@integrione.id" {
		t.Fatalf("normalized email = %q, want budi.santoso@integrione.id", emailID.Value)
	}
	if emailID.IsVerified || emailID.VerifiedAt != nil {
		t.Fatal("new identity should start unverified")
	}

	emailID.MarkVerified(now.Add(time.Minute))
	if !emailID.IsVerified || emailID.VerifiedAt == nil {
		t.Fatal("MarkVerified() must set IsVerified=true and VerifiedAt")
	}
	emailID.SetPrimary(false, now.Add(2*time.Minute))
	if emailID.IsPrimary {
		t.Fatal("SetPrimary(false) must update IsPrimary")
	}

	userIdentity, err := entity.NewUserIdentity(testIdentityID, testUserID, entity.IdentityKindUsername, "  Budi.Admin_01 ", false, now)
	if err != nil {
		t.Fatalf("NewUserIdentity(username) error = %v", err)
	}
	if userIdentity.Value != "budi.admin_01" {
		t.Fatalf("normalized username = %q, want budi.admin_01", userIdentity.Value)
	}

	if _, err := entity.NewUserIdentity(testIdentityID, testUserID, entity.IdentityKindEmail, "bad-email", true, now); !errors.Is(err, iamerrors.ErrInvalidEmail) {
		t.Fatalf("invalid email error = %v, want ErrInvalidEmail", err)
	}
	if _, err := entity.NewUserIdentity(testIdentityID, testUserID, entity.IdentityKindUsername, "ab", true, now); !errors.Is(err, iamerrors.ErrInvalidUsername) {
		t.Fatalf("short username error = %v, want ErrInvalidUsername", err)
	}
	if _, err := entity.NewUserIdentity(testIdentityID, testUserID, entity.IdentityKind("phone"), "08123", true, now); !errors.Is(err, iamerrors.ErrInvalidIdentityKind) {
		t.Fatalf("invalid kind error = %v, want ErrInvalidIdentityKind", err)
	}

	formatted := fmt.Sprintf("%v %#v", emailID, emailID)
	if strings.Contains(formatted, "budi.santoso@integrione.id") {
		t.Fatalf("UserIdentity string/GoString leaked raw email PII: %s", formatted)
	}
	if !strings.Contains(formatted, "b***@integrione.id") {
		t.Fatalf("UserIdentity string/GoString missing masked email: %s", formatted)
	}
}
