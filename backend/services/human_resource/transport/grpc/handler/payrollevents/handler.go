package payrollevents

import (
	"context"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/payroll/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/inbox"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Receiver memproses bukti Finance setelah identitas workload terverifikasi.
type Receiver interface {
	inbox.PayrollPostingConfirmer
	inbox.PayrollPayslipPublisher
	inbox.PayrollPaymentConfirmer
	inbox.PayrollSettlementConfirmer
}

type handler struct {
	pb.UnimplementedPayrollEventServiceServer
	receiver Receiver
}

// Register mendaftarkan penerima payroll pada listener workload mTLS privat.
func Register(registrar grpc.ServiceRegistrar, receiver Receiver) error {
	if operation.RequireDependency(registrar) != nil || operation.RequireDependency(receiver) != nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterPayrollEventServiceServer(registrar, &handler{receiver: receiver})
	return nil
}

func (h *handler) DeliverPayrollEvent(ctx context.Context, req *pb.DeliverPayrollEventRequest) (*pb.DeliverPayrollEventResponse, error) {
	caller, ok := peer.FromContext(ctx)
	if !ok || caller == nil {
		return nil, status.Error(codes.Unauthenticated, "verified workload certificate required")
	}
	info, ok := caller.AuthInfo.(credentials.TLSInfo)
	if !ok || len(info.State.VerifiedChains) == 0 || len(info.State.VerifiedChains[0]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "verified workload certificate required")
	}
	workload := info.State.VerifiedChains[0][0]
	if workload == nil || len(workload.URIs) != 1 || workload.URIs[0] == nil || workload.URIs[0].String() != "spiffe://integrione/finance-accounting" {
		return nil, status.Error(codes.PermissionDenied, "workload is not authorized to deliver payroll events")
	}
	const source = "finance_accounting"
	runID, version := req.GetPayrollRunId(), req.GetSchemaVersion()
	var err error
	switch proof := req.GetEvidence().(type) {
	case *pb.DeliverPayrollEventRequest_Posting:
		err = inbox.HandlePayrollPosting(ctx, h.receiver, inbox.PayrollPostingDelivery{
			SourceManagement: source, EventType: "payroll.posted", SchemaVersion: version,
			PayrollRunID: runID, JournalID: proof.Posting.GetJournalId(),
		})
	case *pb.DeliverPayrollEventRequest_Publication:
		err = inbox.HandlePayrollPublication(ctx, h.receiver, inbox.PayrollPublicationDelivery{
			SourceManagement: source, EventType: "payroll.published", SchemaVersion: version,
			PayrollRunID: runID, PublicationID: proof.Publication.GetPublicationId(),
		})
	case *pb.DeliverPayrollEventRequest_Payment:
		err = inbox.HandlePayrollPayment(ctx, h.receiver, inbox.PayrollPaymentDelivery{
			SourceManagement: source, EventType: "payroll.result_paid", SchemaVersion: version,
			PayrollRunID: runID, ResultID: proof.Payment.GetResultId(), PaymentID: proof.Payment.GetPaymentId(),
		})
	case *pb.DeliverPayrollEventRequest_Settlement:
		err = inbox.HandlePayrollSettlement(ctx, h.receiver, inbox.PayrollSettlementDelivery{
			SourceManagement: source, EventType: "payroll.settled", SchemaVersion: version,
			PayrollRunID: runID, BatchID: proof.Settlement.GetBatchId(), SettlementID: proof.Settlement.GetSettlementId(),
		})
	default:
		return nil, status.Error(codes.InvalidArgument, "payroll event evidence required")
	}
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.DeliverPayrollEventResponse{}, nil
}
