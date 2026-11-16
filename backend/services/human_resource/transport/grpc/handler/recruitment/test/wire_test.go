package recruitment_test

import (
	"context"
	"net"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/recruitment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestTypedWirePreservesJSONPrecisionNullAndPagination(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	requisitions := mockdb.NewStore[entity.Requisition, string](h, func(v *entity.Requisition) string { return v.ID })
	candidates := mockdb.NewStore[entity.Candidate, string](h, func(v *entity.Candidate) string { return v.ID })
	applications := mockdb.NewStore[entity.Application, string](h, func(v *entity.Application) string { return v.ID })
	history := mockdb.NewStore[entity.ApplicationStageHistory, int64](h, func(v *entity.ApplicationStageHistory) int64 { return v.ID })
	interviews := mockdb.NewStore[entity.Interview, string](h, func(v *entity.Interview) string { return v.ID })
	panelists := mockdb.NewStore[entity.InterviewPanelist, string](h, func(v *entity.InterviewPanelist) string { return v.ID })
	feedback := mockdb.NewStore[entity.InterviewFeedback, string](h, func(v *entity.InterviewFeedback) string { return v.ID })
	offers := mockdb.NewStore[entity.Offer, string](h, func(v *entity.Offer) string { return v.ID })
	postings := mockdb.NewStore[entity.JobPosting, string](h, func(v *entity.JobPosting) string { return v.ID })
	assessments := mockdb.NewStore[entity.ApplicationAssessment, string](h, func(v *entity.ApplicationAssessment) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{
		Runtime: h.Runtime(), Requisition: requisitions, Candidate: candidates, Application: applications, ApplicationStageHistory: history,
		Interview: interviews, InterviewPanelist: panelists, InterviewFeedback: feedback, Offer: offers, JobPosting: postings, ApplicationAssessment: assessments,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if err := recruitment.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-recruitment", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewRecruitmentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const precise = `{"identifier":9007199254740993}`
	created, err := client.CreateRequisition(ctx, &pb.CreateRequisitionRequest{Input: &pb.RequisitionInput{
		RequisitionNumber: text("TEST"), JobId: text(mockdb.ID(20)), OrganizationUnitReferenceId: text(mockdb.ID(20)),
		HiringManagerPrincipalId: text(mockdb.ID(20)), RequisitionType: text("TEST"), EmploymentCategory: text("employee"), IdempotencyKey: text("TEST"),
		MinimumQualifications: &common.JsonDocument{Utf8: []byte(precise)}, DesiredQualifications: &common.JsonDocument{Utf8: []byte("null")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if string(created.Item.MinimumQualifications.Utf8) != precise || string(created.Item.DesiredQualifications.Utf8) != "null" {
		t.Fatal("JSON precision or literal null changed on wire")
	}
	updated, err := client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{
		Id: created.Item.Id, Input: &pb.RequisitionInput{}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"minimum_qualifications"}},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("required JSON column accepted SQL NULL")
	}
	updated, err = client.UpdateRequisition(ctx, &pb.UpdateRequisitionRequest{
		Id:         created.Item.Id,
		Input:      &pb.RequisitionInput{MinimumQualifications: &common.JsonDocument{Utf8: []byte("null")}},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"minimum_qualifications"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Item.MinimumQualifications == nil || string(updated.Item.MinimumQualifications.Utf8) != "null" || updated.Item.DesiredQualifications == nil {
		t.Fatal("literal JSON null or unselected JSON value was lost")
	}
	page, err := client.GetAllRequisitions(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(page.GetItems()) != 1 || page.GetPage().GetTotal() != 1 {
		t.Fatalf("pagination failed: %v", err)
	}
	_, err = client.GetAllRequisitions(ctx, &common.PageRequest{Limit: -1})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("negative page size accepted")
	}
}

func text(v string) *string { return &v }
