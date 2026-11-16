package succession_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/succession/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/succession"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/succession"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setup(t *testing.T) (pb.SuccessionServiceClient, *mockdb.Harness, *mockdb.Store[entity.SuccessionPlan, string]) {
	t.Helper()
	h := mockdb.New()
	plans := mockdb.NewStore(h, func(v *entity.SuccessionPlan) string { return v.ID })
	candidates := mockdb.NewStore(h, func(v *entity.SuccessionCandidate) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), SuccessionPlan: plans, SuccessionCandidate: candidates})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	connection, err := grpc.NewClient("passthrough:///succession-test",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	return pb.NewSuccessionServiceClient(connection), h, plans
}
func createRequest() *pb.CreateSuccessionPlanRequest {
	return &pb.CreateSuccessionPlanRequest{Input: &pb.SuccessionPlanInput{
		PositionId: proto.String(mockdb.ID(20)), PlanCode: proto.String("SUCCESSION"), PlanVersion: proto.Int32(1),
		ReviewDate: timestamppb.New(mockdb.Now()), SuccessionRisk: proto.String("low"),
	}}
}
func TestWireCRUDPaginationFinderAndPatchPresence(t *testing.T) {
	t.Parallel()
	client, h, store := setup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	created, err := client.CreateSuccessionPlan(ctx, createRequest())
	if err != nil {
		t.Fatal(err)
	}
	if created.Value.Id == "" || created.Value.GetSuccessionRisk() != "low" || len(h.Events) != 1 {
		t.Fatal("create lost values or outbox")
	}
	id := created.Value.Id
	patched, err := client.UpdateSuccessionPlan(ctx, &pb.UpdateSuccessionPlanRequest{Id: id,
		Input:      &pb.SuccessionPlanInput{PlanCode: proto.String("CHANGED"), ReviewDate: &timestamppb.Timestamp{Nanos: -1}},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"plan_code", "succession_risk"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if patched.Value.PlanCode != "CHANGED" || patched.Value.SuccessionRisk != nil || !patched.Value.ReviewDate.AsTime().Equal(mockdb.Now()) {
		t.Fatal("patch must clear selected nullable field and ignore unselected invalid timestamp")
	}
	page, err := client.GetAllSuccessionPlans(ctx, &commonv1.PageRequest{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Page.Total != 1 || page.Page.Limit != 1 {
		t.Fatalf("wrong page: %v", page)
	}
	found, err := client.FindSuccessionPlanByPlanCode(ctx, &commonv1.StringFinderRequest{Value: proto.String("CHANGED"), Limit: 1})
	if err != nil || len(found.GetItems()) != 1 {
		t.Fatalf("finder: %v %v", found, err)
	}
	empty, err := client.GetAllSuccessionPlans(ctx, &commonv1.PageRequest{Limit: 1, Offset: 1})
	if err != nil || len(empty.GetItems()) != 0 || empty.Page.Total != 1 {
		t.Fatalf("offset: %v %v", empty, err)
	}
	if _, err := client.DeleteSuccessionPlan(ctx, &commonv1.IDRequest{Id: id}); err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 0 || len(h.Events) != 3 {
		t.Fatal("delete lost atomic outbox")
	}
	if _, err := client.GetSuccessionPlanByID(ctx, &commonv1.IDRequest{Id: id}); status.Code(err) != codes.NotFound {
		t.Fatalf("missing: %v", err)
	}
}
func TestWireRejectsInvalidPatchAndPreservesRows(t *testing.T) {
	t.Parallel()
	client, h, store := setup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	created, err := client.CreateSuccessionPlan(ctx, createRequest())
	if err != nil {
		t.Fatal(err)
	}
	for _, paths := range [][]string{nil, {"id"}, {"plan_code", "plan_code"}, {"*"}, {"review_date"}, {"plan_code"}} {
		_, err := client.UpdateSuccessionPlan(ctx, &pb.UpdateSuccessionPlanRequest{Id: created.Value.Id,
			Input: &pb.SuccessionPlanInput{}, UpdateMask: &fieldmaskpb.FieldMask{Paths: paths}})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("mask %v: %v", paths, err)
		}
	}
	if len(h.Events) != 1 || len(store.Rows) != 1 || store.Rows[created.Value.Id].PlanCode != "SUCCESSION" {
		t.Fatal("invalid patch changed state")
	}
}
func TestWireDeniedAndRollbackDoNotLeakData(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"authorize", "outbox"} {
		t.Run(stage, func(t *testing.T) {
			client, h, store := setup(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			expected := codes.Internal
			if stage == "authorize" {
				h.Failures[stage] = hrerrors.ErrForbidden
				expected = codes.PermissionDenied
			} else {
				h.Failures[stage] = errors.New("private SQL and employee detail")
			}
			value, err := client.CreateSuccessionPlan(ctx, createRequest())
			if value != nil || status.Code(err) != expected {
				t.Fatalf("failure: %v %v", value, err)
			}
			if len(store.Rows) != 0 || len(h.Events) != 0 {
				t.Fatal("failed RPC committed record or outbox")
			}
			if stage == "outbox" && status.Convert(err).Message() != "internal error" {
				t.Fatalf("unsafe error: %v", err)
			}
		})
	}
}
