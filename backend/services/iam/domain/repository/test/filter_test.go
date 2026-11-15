package repository_test

import (
	"context"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/domain/repository"
)

func TestPageRequestNormalizeAndPageResult(t *testing.T) {
	t.Parallel()

	pZero := (repository.PageRequest{}).Normalize()
	if pZero.Limit != repository.DefaultPageLimit || pZero.Offset != 0 {
		t.Fatalf("zero PageRequest.Normalize() = %+v, want Limit=%d Offset=0", pZero, repository.DefaultPageLimit)
	}

	pClamp := (repository.PageRequest{Limit: 500, Offset: -10}).Normalize()
	if pClamp.Limit != repository.MaxPageLimit || pClamp.Offset != 0 {
		t.Fatalf("clamped PageRequest.Normalize() = %+v, want Limit=%d Offset=0", pClamp, repository.MaxPageLimit)
	}

	res := repository.NewPageResult[string](nil, 25, repository.PageRequest{Limit: 10, Offset: 0})
	if res.Items == nil || len(res.Items) != 0 {
		t.Fatalf("NewPageResult(nil) should produce non-nil empty slice, got %#v", res.Items)
	}
	if !res.HasMore || res.Total != 25 || res.Limit != 10 {
		t.Fatalf("unexpected PageResult metadata: %+v", res)
	}

	lastPage := repository.NewPageResult([]string{"a", "b"}, 12, repository.PageRequest{Limit: 10, Offset: 10})
	if lastPage.HasMore {
		t.Fatal("lastPage.HasMore = true, want false")
	}
}

func TestFiltersNormalizeAndInterfacesCompile(t *testing.T) {
	t.Parallel()

	kind := entity.UserKindEmployee
	uf := (repository.UserFilter{
		Query: "  Budi  ",
		Kind:  &kind,
		Page:  repository.PageRequest{Limit: 1000, Offset: -1},
	}).Normalize()
	if uf.Query != "Budi" || uf.Page.Limit != repository.MaxPageLimit || uf.Page.Offset != 0 {
		t.Fatalf("UserFilter.Normalize() = %+v", uf)
	}

	sf := (repository.SessionFilter{
		UserID: "  11111111-1111-4111-8111-111111111111 ",
		Page:   repository.PageRequest{Limit: -5, Offset: 10},
	}).Normalize()
	if sf.UserID != "11111111-1111-4111-8111-111111111111" || sf.Page.Limit != repository.DefaultPageLimit || sf.Page.Offset != 10 {
		t.Fatalf("SessionFilter.Normalize() = %+v", sf)
	}

	rf := (repository.RoleFilter{
		Query:            "  Admin ",
		ManagementDomain: " IAM ",
	}).Normalize()
	if rf.Query != "Admin" || rf.ManagementDomain != "iam" || rf.Page.Limit != repository.DefaultPageLimit {
		t.Fatalf("RoleFilter.Normalize() = %+v", rf)
	}

	pf := (repository.PermissionFilter{
		Query:    "  User ",
		Resource: " User ",
	}).Normalize()
	if pf.Query != "User" || pf.Resource != "user" || pf.Page.Limit != repository.DefaultPageLimit {
		t.Fatalf("PermissionFilter.Normalize() = %+v", pf)
	}

	gf := (repository.GroupFilter{
		Query: "  Finance ",
	}).Normalize()
	if gf.Query != "Finance" || gf.Page.Limit != repository.DefaultPageLimit {
		t.Fatalf("GroupFilter.Normalize() = %+v", gf)
	}

	// Pastikan seluruh kontrak interface repository dan Transactor terkompilasi konsisten.
	var (
		_ repository.Transactor
		_ repository.UserRepository
		_ repository.UserIdentityRepository
		_ repository.UserCredentialRepository
		_ repository.UserTOTPFactorRepository
		_ repository.UserRecoveryCodeRepository
		_ repository.UserTokenRepository
		_ repository.SessionRepository
		_ repository.RoleRepository
		_ repository.PermissionRepository
		_ repository.RolePermissionRepository
		_ repository.UserRoleRepository
		_ repository.GroupRepository
		_ repository.GroupMemberRepository
		_ repository.GroupRoleRepository
	)
	_ = context.Background()
}
