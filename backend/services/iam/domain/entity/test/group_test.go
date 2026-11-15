package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const testGroupID = "99999999-9999-4999-8999-999999999999"

func TestGroupAndGroupMember(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	group, err := entity.NewGroup(testGroupID, " Finance_Approvers ", "Finance Approvers", "Approval group", now)
	if err != nil {
		t.Fatalf("NewGroup() unexpected error: %v", err)
	}
	if group.Code != "finance_approvers" || !group.IsActive {
		t.Fatalf("unexpected group fields: %+v", group)
	}

	if err := group.UpdateDetails("Regional Finance Approvers", "Updated desc", now.Add(time.Minute)); err != nil {
		t.Fatalf("UpdateDetails() error = %v", err)
	}
	group.Deactivate(now.Add(2 * time.Minute))
	if group.IsActive {
		t.Fatal("group must be inactive after Deactivate")
	}
	group.Activate(now.Add(3 * time.Minute))
	if !group.IsActive {
		t.Fatal("group must be active after Activate")
	}

	if _, err := entity.NewGroup(testGroupID, "bad code!", "Name", "", now); !errors.Is(err, iamerrors.ErrInvalidGroupCode) {
		t.Fatalf("invalid group code error = %v, want ErrInvalidGroupCode", err)
	}

	member, err := entity.NewGroupMember(testGroupID, testUserID, now)
	if err != nil {
		t.Fatalf("NewGroupMember() unexpected error: %v", err)
	}
	if member.GroupID != testGroupID || member.UserID != testUserID {
		t.Fatalf("GroupMember mismatch: %+v", member)
	}
}
