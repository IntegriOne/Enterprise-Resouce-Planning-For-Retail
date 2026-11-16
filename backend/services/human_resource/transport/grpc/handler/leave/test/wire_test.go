package leave_test

import (
	"context"
	"net"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/leave"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestTypedWireKeepsDecimalNullAndFalseDistinct(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	policies := mockdb.NewStore[entity.LeavePolicy, string](h, func(v *entity.LeavePolicy) string { return v.ID })
	entitlements := mockdb.NewStore[entity.LeaveEntitlement, string](h, func(v *entity.LeaveEntitlement) string { return v.ID })
	adjustments := mockdb.NewStore[entity.LeaveAdjustment, string](h, func(v *entity.LeaveAdjustment) string { return v.ID })
	entries := mockdb.NewStore[entity.LeaveLedgerEntry, string](h, func(v *entity.LeaveLedgerEntry) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{
		Runtime: h.Runtime(), LeavePolicy: policies, LeaveEntitlement: entitlements,
		LeaveAdjustment: adjustments, LeaveLedgerEntry: entries,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if err := leave.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-leave", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewLeaveServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	version := int32(1)
	created, err := client.CreateLeavePolicy(ctx, &pb.CreateLeavePolicyRequest{Input: &pb.LeavePolicyInput{
		PolicyCode: text("ANNUAL"), PolicyVersion: &version, PolicyName: text("Annual leave"),
		LeaveTypeCode: text("ANNUAL"), JurisdictionCode: text("TEST"), AccrualMethod: text("TEST"),
		EffectiveFrom: text("2026-11-16"), AnnualEntitlement: text("12.2500"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if created.Item.GetAnnualEntitlement() != "12.2500" || created.Item.CarryoverLimit != nil {
		t.Fatal("wire conversion lost exact decimal or SQL NULL")
	}
	unpaid := false
	updated, err := client.UpdateLeavePolicy(ctx, &pb.UpdateLeavePolicyRequest{
		Id: created.Item.Id, Input: &pb.LeavePolicyInput{IsPaid: &unpaid, PolicyName: text("ignored")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"annual_entitlement", "is_paid"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Item.AnnualEntitlement != nil || updated.Item.IsPaid || updated.Item.PolicyName != "Annual leave" {
		t.Fatal("wire patch lost selected NULL, explicit false, or unselected field")
	}
	page, err := client.GetAllLeavePolicies(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(page.GetItems()) != 1 || page.GetPage().GetTotal() != 1 {
		t.Fatalf("wire pagination failed: %v", err)
	}
	_, err = client.GetAllLeavePolicies(ctx, &common.PageRequest{Offset: -1})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("negative offset accepted")
	}
}

func TestCommandInputExcludesServerOwnedLedgerFields(t *testing.T) {
	t.Parallel()
	fields := (&pb.PostLedgerEntryInput{}).ProtoReflect().Descriptor().Fields()
	for _, name := range []protoreflect.Name{"entitlement_id", "entry_sequence", "actor_principal_id", "recorded_at", "id"} {
		if fields.ByName(name) != nil {
			t.Fatalf("command input exposes server-owned field %s", name)
		}
	}
}
