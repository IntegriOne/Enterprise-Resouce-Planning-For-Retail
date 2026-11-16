package payroll_test

import (
	"context"
	"net"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/payroll"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestTypedWirePreservesNullablePatchPaginationAndPolicyDenial(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	profiles := mockdb.NewStore[entity.PayrollProfile, string](h, func(v *entity.PayrollProfile) string { return v.ID })
	components := mockdb.NewStore[entity.PayComponent, string](h, func(v *entity.PayComponent) string { return v.ID })
	periods := mockdb.NewStore[entity.PayrollPeriod, string](h, func(v *entity.PayrollPeriod) string { return v.ID })
	runs := mockdb.NewStore[entity.PayrollRun, string](h, func(v *entity.PayrollRun) string { return v.ID })
	results := mockdb.NewStore[entity.PayrollResult, string](h, func(v *entity.PayrollResult) string { return v.ID })
	lines := mockdb.NewStore[entity.PayrollResultLine, string](h, func(v *entity.PayrollResultLine) string { return v.ID })
	adjustments := mockdb.NewStore[entity.PayrollAdjustment, string](h, func(v *entity.PayrollAdjustment) string { return v.ID })
	groups := mockdb.NewStore[entity.PayrollGroup, string](h, func(v *entity.PayrollGroup) string { return v.ID })
	rules := mockdb.NewStore[entity.PayrollRuleSet, string](h, func(v *entity.PayrollRuleSet) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{
		Runtime: h.Runtime(), PayrollProfile: profiles, PayComponent: components, PayrollPeriod: periods,
		PayrollRun: runs, PayrollResult: results, PayrollResultLine: lines, PayrollAdjustment: adjustments, PayrollGroup: groups, PayrollRuleSet: rules,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if err := payroll.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-payroll", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewPayrollServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	created, err := client.CreatePayrollProfile(ctx, &pb.CreatePayrollProfileRequest{Input: &pb.PayrollProfileInput{
		WorkerId: text(mockdb.ID(20)), PayrollCountryCode: text("ID"), PayrollJurisdictionCode: text("TEST"),
		PayrollGroupId: text(mockdb.ID(20)), EmployingLegalEntityReferenceId: text(mockdb.ID(20)), CurrencyCode: text("IDR"),
		EffectiveFrom: text("2026-11-16"), TaxStatusCode: text("INITIAL"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if created.Item.Id == "" || created.Item.TaxStatusCode == nil {
		t.Fatal("typed create lost ID or nullable field")
	}
	inactive := false
	updated, err := client.UpdatePayrollProfile(ctx, &pb.UpdatePayrollProfileRequest{
		Id: created.Item.Id, Input: &pb.PayrollProfileInput{IsActive: &inactive, PayrollCountryCode: text("US")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"tax_status_code", "is_active"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Item.TaxStatusCode != nil || updated.Item.IsActive || updated.Item.PayrollCountryCode != "ID" {
		t.Fatal("mask must preserve absent, explicit false, and selected NULL semantics")
	}
	page, err := client.GetAllPayrollProfiles(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(page.GetItems()) != 1 || page.GetPage().GetTotal() != 1 {
		t.Fatalf("typed pagination failed: %v", err)
	}
	_, err = client.GetAllPayrollProfiles(ctx, &common.PageRequest{Limit: -1})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("negative pagination code: %s", status.Code(err))
	}
	h.DeniedObject = created.Item.Id
	_, err = client.GetPayrollProfileByID(ctx, &common.IDRequest{Id: created.Item.Id})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("policy denial code: %s", status.Code(err))
	}
}

func TestInputRejectsNarrowingOverflowAndServerFields(t *testing.T) {
	t.Parallel()
	h := &payroll.Handler{}
	scale := int32(65536)
	_, err := h.CreatePayrollGroup(context.Background(), &pb.CreatePayrollGroupRequest{Input: &pb.PayrollGroupInput{CurrencyScale: &scale}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("currency scale overflow accepted")
	}
	_, err = h.UpdatePayrollRun(context.Background(), &pb.UpdatePayrollRunRequest{Input: &pb.PayrollRunInput{}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"started_by_principal_id"}}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("server-owned actor mask accepted")
	}
}

func text(value string) *string { return &value }
