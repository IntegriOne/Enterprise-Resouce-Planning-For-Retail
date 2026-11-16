package leave_test

import (
	"context"
	"testing"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/leave"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestMalformedInputRejectedBeforeApplicationAccess(t *testing.T) {
	t.Parallel()
	h := &leave.Handler{}
	for _, test := range []struct {
		name string
		call func() error
	}{
		{"nil create", func() error { _, err := h.CreateLeavePolicy(context.Background(), nil); return err }},
		{"invalid calendar date", func() error {
			_, err := h.CreateLeavePolicy(context.Background(), &pb.CreateLeavePolicyRequest{Input: &pb.LeavePolicyInput{EffectiveFrom: text("2026-02-30")}})
			return err
		}},
		{"server field mask", func() error {
			_, err := h.UpdateLeaveEntitlement(context.Background(), &pb.UpdateLeaveEntitlementRequest{Input: &pb.LeaveEntitlementInput{}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"calculated_at"}}})
			return err
		}},
		{"missing ledger effective date", func() error {
			_, err := h.PostLedgerEntry(context.Background(), &pb.PostLedgerEntryRequest{Input: &pb.PostLedgerEntryInput{}})
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if code := status.Code(test.call()); code != codes.InvalidArgument {
				t.Fatalf("got code %s, want InvalidArgument", code)
			}
		})
	}
}

func text(value string) *string { return &value }
