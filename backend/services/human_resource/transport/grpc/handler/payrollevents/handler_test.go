package payrollevents

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/url"
	"testing"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/payroll/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type payrollReceiverFixture struct {
	Receiver
	calls int
	err   error
}

func (r *payrollReceiverFixture) ConfirmPayrollPosting(_ context.Context, runID, journalID string) error {
	r.calls++
	if runID != "run" || journalID != "journal" {
		return hrerrors.ErrReferenceConflict
	}
	return r.err
}

func TestDeliverPayrollEventRequiresFinanceWorkload(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		identity string
		version  int32
		err      error
		code     codes.Code
		calls    int
	}{
		{name: "missing certificate", version: 1, code: codes.Unauthenticated},
		{name: "other workload", identity: "spiffe://integrione/iam", version: 1, code: codes.PermissionDenied},
		{name: "wrong version", identity: "spiffe://integrione/finance-accounting", version: 2, code: codes.InvalidArgument},
		{name: "verifier unavailable", identity: "spiffe://integrione/finance-accounting", version: 1, err: hrerrors.ErrUnavailable, code: codes.Unavailable, calls: 1},
		{name: "verified event", identity: "spiffe://integrione/finance-accounting", version: 1, code: codes.OK, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			if tc.identity != "" {
				identity, err := url.Parse(tc.identity)
				if err != nil {
					t.Fatal(err)
				}
				certificate := &x509.Certificate{URIs: []*url.URL{identity}}
				ctx = peer.NewContext(ctx, &peer.Peer{AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{certificate}}},
				}})
			}
			fixture := &payrollReceiverFixture{err: tc.err}
			h := handler{receiver: fixture}
			got, err := h.DeliverPayrollEvent(ctx, &pb.DeliverPayrollEventRequest{
				PayrollRunId: "run", SchemaVersion: tc.version,
				Evidence: &pb.DeliverPayrollEventRequest_Posting{Posting: &pb.PostingEvidence{JournalId: "journal"}},
			})
			if status.Code(err) != tc.code || fixture.calls != tc.calls || (tc.code == codes.OK) != (got != nil) {
				t.Fatalf("delivery code=%v, calls=%d, response=%v", status.Code(err), fixture.calls, got)
			}
		})
	}
}

func (r *payrollReceiverFixture) PublishPayslipsForRun(context.Context, string, string) error {
	r.calls++
	return r.err
}

func TestPublicationPreservesRetryablePartialFailure(t *testing.T) {
	uri, err := url.Parse("spiffe://integrione/finance-accounting")
	if err != nil {
		t.Fatal(err)
	}
	ctx := peer.NewContext(t.Context(), &peer.Peer{AuthInfo: credentials.TLSInfo{
		State: tls.ConnectionState{VerifiedChains: [][]*x509.Certificate{{{URIs: []*url.URL{uri}}}}},
	}})
	for _, tc := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{"retryable partial failure", errors.Join(hrerrors.ErrNotFound, hrerrors.ErrUnavailable), codes.Unavailable},
		{"permanent failure", hrerrors.ErrNotFound, codes.NotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			receiver := &payrollReceiverFixture{err: tc.err}
			h := handler{receiver: receiver}
			got, err := h.DeliverPayrollEvent(ctx, &pb.DeliverPayrollEventRequest{
				PayrollRunId: "10000000-0000-4000-8000-000000000001", SchemaVersion: 1,
				Evidence: &pb.DeliverPayrollEventRequest_Publication{
					Publication: &pb.PublicationEvidence{PublicationId: "10000000-0000-4000-8000-000000000002"},
				},
			})
			if status.Code(err) != tc.code || got != nil || receiver.calls != 1 {
				t.Fatalf("publication delivery: status=%v response=%v calls=%d", status.Code(err), got, receiver.calls)
			}
		})
	}
}
