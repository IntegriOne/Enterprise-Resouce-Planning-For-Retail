package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/config"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	offboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const fixturePeriodID = "20000000-0000-4000-8000-000000000001"
const fixtureDocumentVersion = "30000000-0000-4000-8000-000000000001"

type hrmPayrollServer struct {
	payroll.UnimplementedPayrollServiceServer
}

func (hrmPayrollServer) GetMyPayslip(_ context.Context, req *payroll.GetMyPayslipRequest) (*payroll.GetMyPayslipResponse, error) {
	if req.GetPayrollResultId() != hrmFixtureID {
		return nil, status.Error(codes.NotFound, "payslip not found")
	}
	item := fixturePayslip()
	if req.GetIncludePdf() {
		item.Pdf = []byte("%PDF-1.7\nfixture")
	}
	return item, nil
}

func (hrmPayrollServer) ListMyPayslips(_ context.Context, req *payroll.ListMyPayslipsRequest) (*payroll.ListMyPayslipsResponse, error) {
	if req.GetPayrollPeriodId() != fixturePeriodID {
		return nil, status.Error(codes.Unavailable, "payslip verifier unavailable")
	}
	item := fixturePayslip()
	switch req.GetOffset() {
	case 1:
		item.PeriodEnd = "2026-10-31"
	case 2:
		item.ScheduledPayDate = ""
	case 3:
		item.PaymentStatus = "unknown"
	case 4:
		return &payroll.ListMyPayslipsResponse{
			Items: []*payroll.GetMyPayslipResponse{item, fixturePayslip()},
			Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 6},
		}, nil
	}
	return &payroll.ListMyPayslipsResponse{
		Items: []*payroll.GetMyPayslipResponse{item},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: int64(req.GetOffset()) + 1},
	}, nil
}

func fixturePayslip() *payroll.GetMyPayslipResponse {
	return &payroll.GetMyPayslipResponse{
		PayrollResultId: hrmFixtureID, PayrollPeriodId: fixturePeriodID, DocumentVersionId: fixtureDocumentVersion,
		PublishedAt:   timestamppb.New(time.Date(2026, time.November, 25, 1, 0, 0, 0, time.UTC)),
		PaymentStatus: "scheduled", ScheduledPayDate: "2026-11-25",
		PeriodStart: "2026-11-01", PeriodEnd: "2026-11-30", PeriodCode: "NOV-2026",
	}
}

func newHRMPayrollFixture(t *testing.T) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	payroll.RegisterPayrollServiceServer(grpcServer, hrmPayrollServer{})
	personnel.RegisterPersonnelServiceServer(grpcServer, hrmWorkerServer{})
	workforce.RegisterWorkforcePlanningServiceServer(grpcServer, hrmWorkforceServer{})
	leave.RegisterLeaveServiceServer(grpcServer, hrmLeaveServer{})
	offboarding.RegisterOffboardingServiceServer(grpcServer, hrmOffboardingServer{})
	t.Cleanup(grpcServer.Stop)
	go func() { _ = grpcServer.Serve(listener) }()
	conn, err := grpc.NewClient("passthrough:///hrm-payroll-fixture",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestHRMMyPayslipsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)

	const detail = `{ myPayslip(payroll_result_id: "` + hrmFixtureID + `") { document_version_id payment_status } }`
	const pdf = `{ myPayslipPdf(payroll_result_id: "` + hrmFixtureID + `") { content_base64 } }`
	const twoPDFs = `{ first: myPayslipPdf(payroll_result_id: "` + hrmFixtureID + `") { content_base64 }
		second: myPayslipPdf(payroll_result_id: "` + hrmFixtureID + `") { content_base64 } }`
	const twoContents = `{ myPayslipPdf(payroll_result_id: "` + hrmFixtureID + `") {
		first: content_base64 second: content_base64 } }`
	const list = `{ myPayslips(payroll_period_id: "` + fixturePeriodID + `", first: 1) { nodes { payroll_period_id } total_count } }`
	cases := []struct {
		name, query, want string
		scope             authz.ScopeGrant
	}{
		{"own detail", detail, `"document_version_id":"` + fixtureDocumentVersion + `","payment_status":"scheduled"`, authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"own PDF", pdf, `"content_base64":"JVBERi0xLjcKZml4dHVyZQ=="`, authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"PDF request budget", twoPDFs, "COMPLEXITY_LIMIT_EXCEEDED", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"PDF content budget", twoContents, "COMPLEXITY_LIMIT_EXCEEDED", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"PDF global only", pdf, "FORBIDDEN", authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"PDF different principal", pdf, "FORBIDDEN", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "other-user"}},
		{"own filtered list", list, `"nodes":[{"payroll_period_id":"` + fixturePeriodID + `"}],"total_count":1`, authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"global does not grant self", detail, "FORBIDDEN", authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"other principal scope", list, "FORBIDDEN", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "other-user"}},
		{"reversed period", `{ myPayslips(payroll_period_id: "` + fixturePeriodID + `", first: 1, after: "djE6MA") { total_count } }`, "UNAVAILABLE", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"missing scheduled date", `{ myPayslips(payroll_period_id: "` + fixturePeriodID + `", first: 1, after: "djE6MQ") { total_count } }`, "UNAVAILABLE", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"unknown payment status", `{ myPayslips(payroll_period_id: "` + fixturePeriodID + `", first: 1, after: "djE6Mg") { total_count } }`, "UNAVAILABLE", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"duplicate results", `{ myPayslips(payroll_period_id: "` + fixturePeriodID + `", first: 2, after: "djE6Mw") { total_count } }`, "UNAVAILABLE", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"unavailable verifier", "{ myPayslips { total_count } }", "UNAVAILABLE", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
		{"missing result", `{ myPayslip(payroll_result_id: "` + fixturePeriodID + `") { document_version_id } }`, "NOT_FOUND", authz.ScopeGrant{Type: authz.ScopeSelf, ID: "user-admin"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, rawToken, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			snap := store.sessions[authsession.HashOpaqueToken(rawToken)].Snapshot
			snap.Permissions = append(snap.Permissions, authz.PermissionGrant{Code: "human_resource.payslips.read_self", Scope: tc.scope})
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: rawToken}}
			resp := doGraphQLPOST(t, srv, map[string]any{"query": tc.query}, nil, cookies)
			body := readBody(t, resp)
			if tc.name == "own PDF" && !strings.Contains(resp.Header.Get("Cache-Control"), "no-store") {
				t.Fatal("private PDF response was cacheable")
			}
			if (strings.HasPrefix(tc.want, `"`) && strings.Contains(body, `"errors"`)) || !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %d %s", tc.want, resp.StatusCode, body)
			}
		})
	}
}

func (hrmPayrollServer) RequestPayrollPosting(_ context.Context, req *common.IDRequest) (*payroll.PayrollRunResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "workflow required")
	}
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: req.GetId(), Status: "approved", PostingRequestedAt: fixturePayslip().PublishedAt,
	}}, nil
}

func TestHRMRequestPayrollPostingThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	cases := []struct {
		name, id, want   string
		grant, mfa, csrf bool
	}{
		{"approved run", hrmFixtureID, `"status":"approved","posting_requested_at":"2026-11-25T01:00:00Z"`, true, true, true},
		{"ineligible run", fixturePeriodID, "CONFLICT", true, true, true},
		{"missing grant", hrmFixtureID, "AUTHORIZATION", false, true, true},
		{"missing MFA", hrmFixtureID, "employee MFA authentication required", true, false, true},
		{"missing CSRF", hrmFixtureID, "missing X-CSRF-Token header", true, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, rawToken, rawCSRF := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(rawToken)]
			session.Token.MFAAuthenticated = tc.mfa
			if tc.grant {
				session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
					Code: "human_resource.payroll_runs.request_posting", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}, IsSensitive: true,
				})
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443"}
			if tc.csrf {
				headers[browserauth.CSRFHeaderName] = rawCSRF
			}
			query := `mutation ($id: ID!) { requestPayrollPosting(payroll_run_id: $id) { was_successful errors { type message } payroll_posting_request { status posting_requested_at } } }`
			request := map[string]any{"query": query, "variables": map[string]any{"id": tc.id}}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: rawToken}}
			resp := doGraphQLPOST(t, srv, request, headers, cookies)
			body := readBody(t, resp)
			if (strings.HasPrefix(tc.want, `"`) && !strings.Contains(body, `"was_successful":true`)) || !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %d %s", tc.want, resp.StatusCode, body)
			}
		})
	}
}

func (hrmPayrollServer) GetPayrollRunByID(_ context.Context, req *common.IDRequest) (*payroll.PayrollRunResponse, error) {
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: req.GetId(), PayrollPeriodId: fixturePeriodID, Status: "approved",
		PostingRequestedAt: fixturePayslip().PublishedAt,
	}}, nil
}

func (hrmPayrollServer) GetAllPayrollRuns(_ context.Context, req *common.PageRequest) (*payroll.PayrollRunsResponse, error) {
	journal := fixtureDocumentVersion
	return &payroll.PayrollRunsResponse{
		Items: []*payroll.PayrollRun{{
			Id: hrmFixtureID, PayrollPeriodId: fixturePeriodID, Status: "posted", FinanceJournalReferenceId: &journal,
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmPayrollServer) FindPayrollRunByIdempotencyKey(_ context.Context, req *common.StringFinderRequest) (*payroll.PayrollRunsResponse, error) {
	response := &payroll.PayrollRunsResponse{Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset()}}
	if req.GetValue() == "draft-key" {
		response.Items = []*payroll.PayrollRun{{Id: hrmFixtureID, PayrollPeriodId: fixturePeriodID, Status: "draft", IdempotencyKey: "draft-key"}}
		response.Page.Total = 1
	}
	return response, nil
}

func (hrmPayrollServer) FindPayrollRunByPayrollPeriodID(_ context.Context, req *common.StringFinderRequest) (*payroll.PayrollRunsResponse, error) {
	return &payroll.PayrollRunsResponse{
		Items: []*payroll.PayrollRun{{Id: hrmFixtureID, PayrollPeriodId: fixturePeriodID, Status: "review"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func TestHRMPayrollRunsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	const detail = `{ payrollRun(id: "` + hrmFixtureID + `") { status posting_requested_at finance_journal_reference_id finance_settled_at } }`
	const list = "{ payrollRuns(first: 1) { nodes { status finance_journal_reference_id } total_count } }"
	for _, tc := range []struct {
		name, query, want string
		scope             authz.ScopeGrant
	}{
		{"requested but awaiting journal", detail, `"status":"approved","posting_requested_at":"2026-11-25T01:00:00Z","finance_journal_reference_id":null,"finance_settled_at":null`, authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"posted run in list", list, `"status":"posted","finance_journal_reference_id":"` + fixtureDocumentVersion + `"}],"total_count":1`, authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"unsupported legal entity scope", list, "FORBIDDEN", authz.ScopeGrant{Type: "legal_entity", ID: fixturePeriodID}},
		{"draft recovery", `{ payrollRuns(idempotency_key:"draft-key") { nodes { id status } total_count } }`, `"id":"gid://integrione/payroll-run/` + hrmFixtureID + `","status":"draft"}],"total_count":1`, authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"unknown key", `{ payrollRuns(idempotency_key:"unknown") { nodes { id } total_count } }`, `"nodes":[],"total_count":0`, authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"blank key", `{ payrollRuns(idempotency_key:" ") { total_count } }`, "INVALID_ARGUMENT", authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"filter period", `{ payrollRuns(payroll_period_id:"` + fixturePeriodID + `") { nodes { status } total_count } }`, `"nodes":[{"status":"review"}],"total_count":1`, authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"period mismatch", `{ payrollRuns(payroll_period_id:"` + hrmFixtureID + `") { total_count } }`, "UNAVAILABLE", authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{"both filters", `{ payrollRuns(payroll_period_id:"` + fixturePeriodID + `", idempotency_key:"draft-key") { total_count } }`, "INVALID_ARGUMENT", authz.ScopeGrant{Type: authz.ScopeGlobal}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, rawToken, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			snapshot := store.sessions[authsession.HashOpaqueToken(rawToken)].Snapshot
			snapshot.Permissions = append(snapshot.Permissions, authz.PermissionGrant{Code: "human_resource.payroll_runs.read", Scope: tc.scope})
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: rawToken}}
			body := readBody(t, doGraphQLPOST(t, srv, map[string]any{"query": tc.query}, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) SubmitPayrollRunForReview(_ context.Context, req *common.IDRequest) (*payroll.PayrollRunResponse, error) {
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: req.GetId(), PayrollPeriodId: fixturePeriodID, Status: "review",
	}}, nil
}

func (hrmPayrollServer) ApprovePayrollRun(_ context.Context, req *common.IDRequest) (*payroll.PayrollRunResponse, error) {
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: req.GetId(), PayrollPeriodId: fixturePeriodID, Status: "approved",
	}}, nil
}

func (hrmPayrollServer) CancelPayrollRun(_ context.Context, req *payroll.CancelPayrollRunRequest) (*payroll.PayrollRunResponse, error) {
	if req.GetReason() != "Review found an error" {
		return nil, status.Error(codes.FailedPrecondition, "unexpected reason")
	}
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: req.GetId(), PayrollPeriodId: fixturePeriodID, Status: "cancelled", CancelledAt: fixturePayslip().PublishedAt,
	}}, nil
}

func (hrmPayrollServer) CalculatePayrollRun(_ context.Context, req *common.IDRequest) (*payroll.PayrollRunResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.Unavailable, "payroll calculator unavailable")
	}
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: req.GetId(), PayrollPeriodId: fixturePeriodID, Status: "calculated", CalculatedAt: fixturePayslip().PublishedAt,
	}}, nil
}

func TestHRMPayrollReviewThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	reason, empty := "  Review found an error  ", "  "
	for _, tc := range []struct {
		name, field, action, want, id string
		reason                        *string
	}{
		{"calculate", "calculatePayrollRun", "calculate", `"status":"calculated","calculated_at":"2026-11-25T01:00:00Z"`, hrmFixtureID, nil},
		{"calculator unavailable", "calculatePayrollRun", "calculate", "UNAVAILABLE", fixturePeriodID, nil},
		{"approve grant cannot calculate", "calculatePayrollRun", "approve", "AUTHORIZATION", hrmFixtureID, nil},
		{"submit", "submitPayrollRunForReview", "submit_review", `"status":"review"`, hrmFixtureID, nil},
		{"approve", "approvePayrollRun", "approve", `"status":"approved"`, hrmFixtureID, nil},
		{"submit grant cannot approve", "approvePayrollRun", "submit_review", "AUTHORIZATION", hrmFixtureID, nil},
		{"cancel in review", "cancelPayrollRun", "cancel", `"status":"cancelled","cancelled_at":"2026-11-25T01:00:00Z"`, hrmFixtureID, &reason},
		{"blank cancellation reason", "cancelPayrollRun", "cancel", "ARGUMENT", hrmFixtureID, &empty},
		{"approve grant cannot cancel", "cancelPayrollRun", "approve", "AUTHORIZATION", hrmFixtureID, &reason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, rawToken, rawCSRF := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(rawToken)]
			session.Token.MFAAuthenticated = true
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.payroll_runs." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := "mutation ($id: ID!) { " + tc.field + "(payroll_run_id: $id) { was_successful errors { type } payroll_run { status calculated_at } } }"
			variables := map[string]any{"id": tc.id}
			if tc.reason != nil {
				query = "mutation ($id: ID!, $reason: String!) { " + tc.field + "(payroll_run_id: $id, reason: $reason) { was_successful errors { type } payroll_run { status cancelled_at } } }"
				variables["reason"] = *tc.reason
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: rawCSRF}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: rawToken}}
			request := map[string]any{"query": query, "variables": variables}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) FreezePayrollInput(_ context.Context, req *payroll.FreezePayrollInputRequest) (*payroll.FreezePayrollInputResponse, error) {
	if req.GetPayrollPeriodId() != fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "payroll period cannot be frozen")
	}
	return &payroll.FreezePayrollInputResponse{
		Item: &payroll.PayrollInputSnapshotMetadata{
			Id:                              hrmFixtureID,
			PayrollPeriodId:                 req.GetPayrollPeriodId(),
			PayrollGroupId:                  hrmFixtureID,
			EmployingLegalEntityReferenceId: hrmFixtureID,
			JurisdictionCode:                "ID",
			CurrencyCode:                    "IDR",
			Digest:                          bytes.Repeat([]byte{0xab}, 32),
			FrozenAt:                        fixturePayslip().PublishedAt,
		},
	}, nil
}

func TestHRMFreezePayrollInputThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	const validIdem = "99999999-9999-4999-8999-999999999999"
	for _, tc := range []struct {
		name       string
		periodID   string
		idemKey    string
		permission string
		mfa        bool
		withCSRF   bool
		want       string
	}{
		{"success", fixturePeriodID, validIdem, "human_resource.payroll_runs.create", true, true, `"was_successful":true`},
		{"closed or empty period", hrmFixtureID, validIdem, "human_resource.payroll_runs.create", true, true, `"type":"CONFLICT"`},
		{"invalid idempotency key", fixturePeriodID, "not-a-uuid", "human_resource.payroll_runs.create", true, true, `"type":"ARGUMENT"`},
		{"missing permission", fixturePeriodID, validIdem, "human_resource.payroll_runs.read", true, true, `"type":"AUTHORIZATION"`},
		{"missing mfa", fixturePeriodID, validIdem, "human_resource.payroll_runs.create", false, true, `"type":"AUTHORIZATION"`},
		{"missing csrf", fixturePeriodID, validIdem, "human_resource.payroll_runs.create", true, false, `"FORBIDDEN"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: tc.permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := `mutation ($input: FreezePayrollInputInput!) { freezePayrollInput(input: $input) { was_successful errors { type } frozen_payroll_input { input_snapshot_reference_id payroll_period_id payroll_group_id employing_legal_entity_reference_id jurisdiction_code currency_code input_digest_sha256 frozen_at } } }`
			request := map[string]any{"query": query, "variables": map[string]any{"input": map[string]any{
				"payroll_period_id": tc.periodID, "idempotency_key": tc.idemKey,
			}}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443"}
			if tc.withCSRF {
				headers[browserauth.CSRFHeaderName] = csrf
			}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) CreatePayrollRun(_ context.Context, req *payroll.CreatePayrollRunRequest) (*payroll.PayrollRunResponse, error) {
	input := req.GetInput()
	if input.GetIdempotencyKey() == "duplicate" {
		return nil, status.Error(codes.AlreadyExists, "duplicate run")
	}
	if len(input.GetInputDigest()) != 32 || input.GetInputDigest()[0] != 0xab ||
		input.GetIdempotencyKey() != "draft-key" || input.Status != nil || input.FinanceJournalReferenceId != nil {
		return nil, status.Error(codes.InvalidArgument, "unexpected draft input")
	}
	return &payroll.PayrollRunResponse{Item: &payroll.PayrollRun{
		Id: hrmFixtureID, PayrollPeriodId: input.GetPayrollPeriodId(), RunNumber: input.GetRunNumber(), Status: "draft",
	}}, nil
}

func TestHRMCreatePayrollRunThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, key, digest, want string }{
		{"draft", "draft-key", strings.Repeat("ab", 32), `"status":"draft","run_number":1`},
		{"duplicate", "duplicate", strings.Repeat("ab", 32), "CONFLICT"},
		{"invalid digest", "draft-key", "wrong", "ARGUMENT"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = true
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.payroll_runs.create", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			input := map[string]any{
				"payroll_period_id": fixturePeriodID, "payroll_group_id": hrmFixtureID,
				"employing_legal_entity_reference_id": hrmFixtureID, "currency_code": "IDR", "rule_set_id": hrmFixtureID,
				"run_number": 1, "run_type": "regular", "jurisdiction_code": "ID", "rule_engine_name": "rule-engine",
				"rule_engine_version": "1", "input_snapshot_reference_id": hrmFixtureID,
				"input_digest_sha256": tc.digest, "idempotency_key": tc.key,
			}
			query := `mutation ($input: CreatePayrollRunInput!) { createPayrollRun(input: $input) { was_successful errors { type } payroll_run { status run_number } } }`
			request := map[string]any{"query": query, "variables": map[string]any{"input": input}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) GetAllPayrollPeriods(_ context.Context, req *common.PageRequest) (*payroll.PayrollPeriodsResponse, error) {
	return &payroll.PayrollPeriodsResponse{
		Items: []*payroll.PayrollPeriod{{
			Id: fixturePeriodID, PayrollGroupId: hrmFixtureID, EmployingLegalEntityReferenceId: hrmFixtureID,
			PeriodCode: "FIXTURE", PeriodStatus: "open",
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 2},
	}, nil
}

func (hrmPayrollServer) GetAllPayrollGroups(_ context.Context, req *common.PageRequest) (*payroll.PayrollGroupsResponse, error) {
	return &payroll.PayrollGroupsResponse{
		Items: []*payroll.PayrollGroup{{
			Id: hrmFixtureID, EmployingLegalEntityReferenceId: fixturePeriodID, GroupCode: "GROUP", IsActive: true,
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 2},
	}, nil
}

func (hrmPayrollServer) GetAllPayrollRuleSets(_ context.Context, req *common.PageRequest) (*payroll.PayrollRuleSetsResponse, error) {
	return &payroll.PayrollRuleSetsResponse{
		Items: []*payroll.PayrollRuleSet{{
			Id: hrmFixtureID, RuleCode: "RULE", RuleVersion: "2", ApprovedAt: fixturePayslip().PublishedAt,
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 2},
	}, nil
}

func (hrmPayrollServer) GetPayrollPeriodByID(_ context.Context, req *common.IDRequest) (*payroll.PayrollPeriodResponse, error) {
	if req.GetId() != fixturePeriodID {
		return nil, status.Error(codes.NotFound, "period not found")
	}
	return &payroll.PayrollPeriodResponse{Item: &payroll.PayrollPeriod{
		Id: req.GetId(), PayrollGroupId: hrmFixtureID, EmployingLegalEntityReferenceId: hrmFixtureID,
		PeriodCode: "FIXTURE", PeriodStatus: "open",
	}}, nil
}

func (hrmPayrollServer) GetPayrollGroupByID(_ context.Context, req *common.IDRequest) (*payroll.PayrollGroupResponse, error) {
	return &payroll.PayrollGroupResponse{Item: &payroll.PayrollGroup{
		Id: req.GetId(), EmployingLegalEntityReferenceId: fixturePeriodID, GroupCode: "GROUP", IsActive: true,
	}}, nil
}

func (hrmPayrollServer) GetPayrollRuleSetByID(_ context.Context, req *common.IDRequest) (*payroll.PayrollRuleSetResponse, error) {
	return &payroll.PayrollRuleSetResponse{Item: &payroll.PayrollRuleSet{
		Id: req.GetId(), RuleCode: "RULE", RuleVersion: "2",
	}}, nil
}

func (hrmPayrollServer) FindPayrollPeriodByPayrollGroupID(_ context.Context, req *common.StringFinderRequest) (*payroll.PayrollPeriodsResponse, error) {
	return &payroll.PayrollPeriodsResponse{
		Items: []*payroll.PayrollPeriod{{
			Id: fixturePeriodID, PayrollGroupId: hrmFixtureID, EmployingLegalEntityReferenceId: fixturePeriodID, PeriodCode: "GROUP-PERIOD",
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func TestHRMPayrollReferenceQueriesThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	const periods = "{ payrollPeriods(first: 1, after: \"djE6MA\") { nodes { period_code period_status } total_count } }"
	const rules = "{ payrollRuleSets(first: 1, after: \"djE6MA\") { nodes { rule_code rule_version approved_at } total_count } }"
	const groups = "{ payrollGroups(first: 1, after: \"djE6MA\") { nodes { group_code is_active } total_count } }"
	for _, tc := range []struct{ name, permission, query, want string }{
		{"list periods", "human_resource.payroll_periods.read", periods, `"period_code":"FIXTURE","period_status":"open"}],"total_count":2`},
		{"run permission only", "human_resource.payroll_runs.read", periods, "FORBIDDEN"},
		{"periods by group", "human_resource.payroll_periods.read", `{ payrollPeriods(payroll_group_id:"` + hrmFixtureID + `") { nodes { period_code } total_count } }`, `"nodes":[{"period_code":"GROUP-PERIOD"}],"total_count":1`},
		{"invalid group filter", "human_resource.payroll_periods.read", `{ payrollPeriods(payroll_group_id:"bad") { total_count } }`, "INVALID_ARGUMENT"},
		{"mismatched group response", "human_resource.payroll_periods.read", `{ payrollPeriods(payroll_group_id:"` + fixturePeriodID + `") { total_count } }`, "UNAVAILABLE"},
		{"period detail", "human_resource.payroll_periods.read", `{ payrollPeriod(id:"` + fixturePeriodID + `") { period_code period_status } }`, `"period_code":"FIXTURE","period_status":"open"`},
		{"period missing", "human_resource.payroll_periods.read", `{ payrollPeriod(id:"` + hrmFixtureID + `") { id } }`, "NOT_FOUND"},
		{"period invalid ID", "human_resource.payroll_periods.read", `{ payrollPeriod(id:"invalid") { id } }`, "INVALID_ARGUMENT"},
		{"list groups", "human_resource.payroll_groups.read", groups, `"group_code":"GROUP","is_active":true}],"total_count":2`},
		{"period permission cannot list groups", "human_resource.payroll_periods.read", groups, "FORBIDDEN"},
		{"list rules", "human_resource.payroll_rule_sets.read", rules, `"rule_code":"RULE","rule_version":"2","approved_at":"2026-11-25T01:00:00Z"}],"total_count":2`},
		{"group permission cannot list rules", "human_resource.payroll_groups.read", rules, "FORBIDDEN"},
		{"group detail", "human_resource.payroll_groups.read", `{ payrollGroup(id:"` + hrmFixtureID + `") { group_code is_active } }`, `"group_code":"GROUP","is_active":true`},
		{"rule detail", "human_resource.payroll_rule_sets.read", `{ payrollRuleSet(id:"` + hrmFixtureID + `") { rule_version approved_at } }`, `"rule_version":"2","approved_at":null`},
		{"group permission cannot read rule detail", "human_resource.payroll_groups.read", `{ payrollRuleSet(id:"` + hrmFixtureID + `") { id } }`, "FORBIDDEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: tc.permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, map[string]any{"query": tc.query}, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) FindPayrollResultByPayrollRunID(_ context.Context, req *common.StringFinderRequest) (*payroll.PayrollResultsResponse, error) {
	return &payroll.PayrollResultsResponse{
		Items: []*payroll.PayrollResult{{
			Id: hrmFixtureID, PayrollRunId: hrmFixtureID, WorkerId: fixturePeriodID,
			CurrencyCode: "IDR", NetPayAmount: "123456789012345.67",
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmPayrollServer) GetPayrollResultByID(_ context.Context, req *common.IDRequest) (*payroll.PayrollResultResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.NotFound, "payroll result not found")
	}
	return &payroll.PayrollResultResponse{Item: &payroll.PayrollResult{
		Id: req.GetId(), PayrollRunId: fixturePeriodID, WorkerId: fixtureDocumentVersion, NetPayAmount: "123456789012345.67",
	}}, nil
}

func TestHRMPayrollResultsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, permission, id, want string
		mfa, detail                bool
	}{
		{"precise amount", "payroll_results.read", hrmFixtureID, `"net_pay_amount":"123456789012345.67"`, true, false},
		{"missing MFA", "payroll_results.read", hrmFixtureID, "FORBIDDEN", false, false},
		{"run read only", "payroll_runs.read", hrmFixtureID, "FORBIDDEN", true, false},
		{"mismatched result run", "payroll_results.read", fixturePeriodID, "UNAVAILABLE", true, false},
		{"detail with MFA", "payroll_results.read", hrmFixtureID, `"net_pay_amount":"123456789012345.67"`, true, true},
		{"detail missing", "payroll_results.read", fixturePeriodID, "NOT_FOUND", true, true},
		{"detail without MFA", "payroll_results.read", hrmFixtureID, "FORBIDDEN", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := `query ($run: ID!) { payrollResults(payroll_run_id: $run) { nodes { net_pay_amount } } }`
			if tc.detail {
				query = `query ($run: ID!) { payrollResult(id: $run) { net_pay_amount } }`
			}
			request := map[string]any{"query": query, "variables": map[string]any{"run": tc.id}}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) FindPayrollResultLineByPayrollResultID(_ context.Context, req *common.StringFinderRequest) (*payroll.PayrollResultLinesResponse, error) {
	return &payroll.PayrollResultLinesResponse{
		Items: []*payroll.PayrollResultLine{{
			Id: fixturePeriodID, PayrollResultId: hrmFixtureID, ComponentCodeSnapshot: "BASE", Amount: "123456789012345.67",
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmPayrollServer) GetPayrollResultLineByID(_ context.Context, req *common.IDRequest) (*payroll.PayrollResultLineResponse, error) {
	if req.GetId() != fixturePeriodID {
		return nil, status.Error(codes.NotFound, "line not found")
	}
	return &payroll.PayrollResultLineResponse{Item: &payroll.PayrollResultLine{
		Id: req.GetId(), PayrollResultId: hrmFixtureID, ComponentCodeSnapshot: "BASE", Amount: "123456789012345.67",
	}}, nil
}

func TestHRMPayrollResultLinesThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, permission, id, want string
		mfa, detail                bool
	}{
		{"precise component", "payroll_result_lines.read", hrmFixtureID, `"component_code":"BASE","amount":"123456789012345.67","rate":null`, true, false},
		{"result permission only", "payroll_results.read", hrmFixtureID, "FORBIDDEN", true, false},
		{"missing MFA", "payroll_result_lines.read", hrmFixtureID, "FORBIDDEN", false, false},
		{"mismatched parent", "payroll_result_lines.read", fixturePeriodID, "UNAVAILABLE", true, false},
		{"line detail", "payroll_result_lines.read", fixturePeriodID, `"component_code":"BASE","amount":"123456789012345.67"`, true, true},
		{"line missing", "payroll_result_lines.read", hrmFixtureID, "NOT_FOUND", true, true},
		{"line detail without MFA", "payroll_result_lines.read", fixturePeriodID, "FORBIDDEN", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := `query ($id: ID!) { payrollResultLines(payroll_result_id: $id) { nodes { component_code amount rate } } }`
			if tc.detail {
				query = `query ($id: ID!) { payrollResultLine(id: $id) { component_code amount } }`
			}
			request := map[string]any{"query": query, "variables": map[string]any{"id": tc.id}}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmLeaveServer) ApproveLeaveAdjustment(_ context.Context, req *common.IDRequest) (*leave.LeaveAdjustmentResponse, error) {
	return &leave.LeaveAdjustmentResponse{Item: &leave.LeaveAdjustment{
		Id: req.GetId(), EntitlementId: fixturePeriodID, ApprovedAt: fixturePayslip().PublishedAt,
	}}, nil
}

func (hrmLeaveServer) RejectLeaveAdjustment(_ context.Context, req *leave.RejectLeaveAdjustmentRequest) (*leave.LeaveAdjustmentResponse, error) {
	return &leave.LeaveAdjustmentResponse{Item: &leave.LeaveAdjustment{
		Id: req.GetId(), EntitlementId: fixturePeriodID, RejectedAt: fixturePayslip().PublishedAt, RejectionReasonCode: &req.ReasonCode,
	}}, nil
}

func (hrmLeaveServer) CloseLeaveEntitlement(_ context.Context, req *common.IDRequest) (*leave.LeaveEntitlementResponse, error) {
	return &leave.LeaveEntitlementResponse{Item: &leave.LeaveEntitlement{
		Id: req.GetId(), WorkerId: fixturePeriodID, LeavePolicyId: fixtureDocumentVersion, ClosedAt: fixturePayslip().PublishedAt,
	}}, nil
}

func TestHRMLeaveAdjustmentDecisionsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, permission, selection, want string
		mfa, csrf                         bool
	}{
		{"approve", "leave_adjustments.approve", `approveLeaveAdjustment(id:$id) { was_successful errors { type } leave_adjustment { approved_at } }`, `"approved_at":"2026-11-25T01:00:00Z"`, true, true},
		{"reject", "leave_adjustments.reject", `rejectLeaveAdjustment(id:$id, reason_code:" WRONG_BALANCE ") { was_successful errors { type } leave_adjustment { rejection_reason_code } }`, `"rejection_reason_code":"WRONG_BALANCE"`, true, true},
		{"blank reason", "leave_adjustments.reject", `rejectLeaveAdjustment(id:$id, reason_code:" ") { was_successful errors { type } leave_adjustment { id } }`, "ARGUMENT", true, true},
		{"wrong grant", "leave_adjustments.approve", `rejectLeaveAdjustment(id:$id, reason_code:"WRONG_BALANCE") { was_successful errors { type } leave_adjustment { id } }`, "AUTHORIZATION", true, true},
		{"missing MFA", "leave_adjustments.approve", `approveLeaveAdjustment(id:$id) { was_successful errors { type } leave_adjustment { id } }`, "AUTHORIZATION", false, true},
		{"missing CSRF", "leave_adjustments.approve", `approveLeaveAdjustment(id:$id) { was_successful errors { type } leave_adjustment { id } }`, "FORBIDDEN", true, false},
		{"close entitlement", "leave_entitlements.close", `closeLeaveEntitlement(id:$id) { was_successful errors { type } leave_entitlement { closed_at } }`, `"closed_at":"2026-11-25T01:00:00Z"`, true, true},
		{"close without MFA", "leave_entitlements.close", `closeLeaveEntitlement(id:$id) { was_successful errors { type } leave_entitlement { id } }`, "AUTHORIZATION", false, true},
		{"read grant cannot close", "leave_entitlements.read", `closeLeaveEntitlement(id:$id) { was_successful errors { type } leave_entitlement { id } }`, "AUTHORIZATION", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := "mutation($id: ID!) { " + tc.selection + " }"
			request := map[string]any{"query": query, "variables": map[string]any{"id": hrmFixtureID}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443"}
			if tc.csrf {
				headers[browserauth.CSRFHeaderName] = csrf
			}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmLeaveServer) CreateLeaveAdjustment(_ context.Context, req *leave.CreateLeaveAdjustmentRequest) (*leave.LeaveAdjustmentResponse, error) {
	input := req.GetInput()
	if input.GetIdempotencyKey() == "duplicate" {
		return nil, status.Error(codes.AlreadyExists, "duplicate adjustment")
	}
	if input.GetQuantityDelta() != "-0.50" || input.GetIdempotencyKey() != "adjust-1" || input.ApprovedAt != nil {
		return nil, status.Error(codes.InvalidArgument, "unexpected adjustment input")
	}
	return &leave.LeaveAdjustmentResponse{Item: &leave.LeaveAdjustment{
		Id: fixtureDocumentVersion, EntitlementId: input.GetEntitlementId(), QuantityDelta: input.GetQuantityDelta(),
	}}, nil
}

func TestHRMCreateLeaveAdjustmentThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, key, want string }{
		{"draft", "adjust-1", `"quantity_delta":"-0.50","approved_at":null,"rejected_at":null`},
		{"duplicate", "duplicate", "CONFLICT"},
		{"blank key", " ", "ARGUMENT"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.leave_adjustments.create", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := `mutation($input: CreateLeaveAdjustmentInput!) {
				createLeaveAdjustment(input:$input) { was_successful errors { type } leave_adjustment { quantity_delta approved_at rejected_at } }
			}`
			input := map[string]any{
				"entitlement_id": fixturePeriodID, "quantity_delta": "-0.50",
				"reason_code": "CORRECTION", "reason": "Fixture", "idempotency_key": tc.key,
			}
			request := map[string]any{"query": query, "variables": map[string]any{"input": input}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmLeaveServer) PostLedgerEntry(_ context.Context, req *leave.PostLedgerEntryRequest) (*leave.LeaveEntitlementResponse, error) {
	input := req.GetInput()
	expected := sha256.Sum256([]byte(`["adjustment-posting-v1","` + fixturePeriodID + `","` + hrmFixtureID + `","-0.50","2026-11-25","CORRECTION"]`))
	if req.GetEntitlementId() != fixturePeriodID || input.GetAdjustmentId() != hrmFixtureID ||
		input.GetSourceEventId() != hrmFixtureID || input.GetSourceManagement() != "hrm" ||
		input.GetEntryType() != "adjustment" || input.GetSourceLineKey() != "adjustment-posting-v1" ||
		!bytes.Equal(input.GetSourceDigest(), expected[:]) {
		return nil, status.Error(codes.InvalidArgument, "unexpected posting provenance")
	}
	return &leave.LeaveEntitlementResponse{Item: &leave.LeaveEntitlement{
		Id: req.GetEntitlementId(), WorkerId: hrmFixtureID, LeavePolicyId: fixtureDocumentVersion, ClosingBalanceSnapshot: "11.50",
	}}, nil
}

func TestHRMPostLeaveAdjustmentThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, permission, date, want string }{
		{"post", "leave_entitlements.post", "2026-11-25", `"closing_balance":"11.50"`},
		{"bad date", "leave_entitlements.post", "2026-13-01", "ARGUMENT"},
		{"approve grant cannot post", "leave_adjustments.approve", "2026-11-25", "AUTHORIZATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = true
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := `mutation($input: PostLeaveAdjustmentInput!) { postLeaveAdjustment(input:$input) { was_successful errors { type } leave_entitlement { closing_balance } } }`
			input := map[string]any{
				"entitlement_id": fixturePeriodID, "adjustment_id": hrmFixtureID,
				"quantity_delta": "-0.50", "effective_date": tc.date, "reason_code": "CORRECTION",
			}
			request := map[string]any{"query": query, "variables": map[string]any{"input": input}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) SubmitWorkforcePlan(_ context.Context, req *common.IDRequest) (*workforce.WorkforcePlanResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "plan transition not permitted")
	}
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{Id: req.GetId(), Status: "review"}}, nil
}

func (hrmWorkforceServer) CancelWorkforcePlan(_ context.Context, req *common.IDRequest) (*workforce.WorkforcePlanResponse, error) {
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{Id: req.GetId(), Status: "cancelled"}}, nil
}

func TestHRMWorkforcePlanCommandsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, field, id, want string
		mfa                           bool
	}{
		{"submit", "submit", "submitWorkforcePlan", hrmFixtureID, `"status":"review"`, true},
		{"cancel", "cancel", "cancelWorkforcePlan", hrmFixtureID, `"status":"cancelled"`, true},
		{"invalid transition", "submit", "submitWorkforcePlan", fixturePeriodID, "CONFLICT", true},
		{"submit grant cannot cancel", "submit", "cancelWorkforcePlan", hrmFixtureID, "AUTHORIZATION", true},
		{"cancel without MFA", "cancel", "cancelWorkforcePlan", hrmFixtureID, "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workforce_plans." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			query := "mutation($id: ID!) { " + tc.field + "(id:$id) { was_successful errors { type } workforce_plan { status } } }"
			request := map[string]any{"query": query, "variables": map[string]any{"id": tc.id}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) CreateWorkforcePlan(_ context.Context, req *workforce.CreateWorkforcePlanRequest) (*workforce.WorkforcePlanResponse, error) {
	in := req.GetInput()
	if in.GetPlanCode() == "duplicate" {
		return nil, status.Error(codes.AlreadyExists, "duplicate plan")
	}
	if in.GetPlanCode() != "WF-TEST" || in.GetPlanVersion() != 1 || in.GetLegalEntityReferenceId() != hrmFixtureID ||
		in.GetPeriodStart() != "2026-11-01" || in.GetPeriodEnd() != "2026-11-30" ||
		in.FinanceBudgetReferenceId != nil || in.Status != nil || in.ApprovedByPrincipalId != nil || in.ApprovedAt != nil {
		return nil, status.Error(codes.Internal, "unexpected draft input")
	}
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{
		Id: fixturePeriodID, LegalEntityReferenceId: in.GetLegalEntityReferenceId(), Status: "draft",
	}}, nil
}

func TestHRMCreateWorkforcePlanThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, code, want string
		version                  int
	}{
		{"draft", "create", "WF-TEST", `"status":"draft"`, 1},
		{"duplicate", "create", "duplicate", "CONFLICT", 1},
		{"invalid version", "create", "WF-TEST", "ARGUMENT", 0},
		{"submit grant", "submit", "WF-TEST", "AUTHORIZATION", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workforce_plans." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreateWorkforcePlanInput!) { createWorkforcePlan(input:$input) { was_successful errors { type } workforce_plan { status } } }",
				"variables": map[string]any{"input": map[string]any{
					"plan_code": tc.code, "plan_version": tc.version, "legal_entity_reference_id": hrmFixtureID,
					"period_start": "2026-11-01", "period_end": "2026-11-30",
				}},
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) CreateWorkforcePlanLine(_ context.Context, req *workforce.CreateWorkforcePlanLineRequest) (*workforce.WorkforcePlanLineResponse, error) {
	in := req.GetInput()
	if in.GetWorkforcePlanId() == fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "plan is not a draft")
	}
	if in.GetWorkforcePlanId() != hrmFixtureID || in.GetLineNumber() != 1 || in.GetJobId() != fixturePeriodID ||
		in.GetOrganizationUnitReferenceId() != hrmFixtureID || in.GetRequiredFte() != "1.25" ||
		in.GetEstimatedCost() != "1200.50" || in.GetCurrencyCode() != "IDR" || in.GetNeededFrom() != "2026-11-01" {
		return nil, status.Error(codes.Internal, "unexpected line input")
	}
	return &workforce.WorkforcePlanLineResponse{Item: &workforce.WorkforcePlanLine{
		Id: fixturePeriodID, WorkforcePlanId: in.GetWorkforcePlanId(), LineNumber: in.GetLineNumber(), RequiredFte: in.GetRequiredFte(),
	}}, nil
}

func TestHRMCreateWorkforcePlanLineThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, planID, want string
	}{
		{"draft line", "create", hrmFixtureID, `"required_fte":"1.25"`},
		{"plan in review", "create", fixturePeriodID, "CONFLICT"},
		{"invalid plan", "create", "invalid", "ARGUMENT"},
		{"read grant", "read", hrmFixtureID, "AUTHORIZATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workforce_plan_lines." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreateWorkforcePlanLineInput!) { createWorkforcePlanLine(input:$input) { was_successful errors { type } workforce_plan_line { required_fte } } }",
				"variables": map[string]any{"input": map[string]any{
					"workforce_plan_id": tc.planID, "line_number": 1, "organization_unit_reference_id": hrmFixtureID,
					"job_id": fixturePeriodID, "worker_category": "employee", "current_fte": "1.00", "required_fte": "1.25",
					"estimated_cost": "1200.50", "currency_code": "IDR", "needed_from": "2026-11-01",
					"action_code": "recruit", "rationale": "Staffing gap",
				}},
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) DeleteWorkforcePlanLine(_ context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	switch req.GetId() {
	case hrmFixtureID:
		return &common.EmptyResponse{}, nil
	case fixturePeriodID:
		return nil, status.Error(codes.FailedPrecondition, "line cannot be deleted")
	default:
		return nil, status.Error(codes.NotFound, "line not found")
	}
}

func TestHRMDeleteWorkforceRecordsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, id, want, resource, field string
		mfa                                     bool
	}{
		{"draft line", "delete", hrmFixtureID, `"deleted":true`, "workforce_plan_lines", "deleteWorkforcePlanLine", true},
		{"workflow conflict", "delete", fixturePeriodID, "CONFLICT", "workforce_plan_lines", "deleteWorkforcePlanLine", true},
		{"invalid ID", "delete", "invalid", "ARGUMENT", "workforce_plan_lines", "deleteWorkforcePlanLine", true},
		{"no MFA", "delete", hrmFixtureID, "AUTHORIZATION", "workforce_plan_lines", "deleteWorkforcePlanLine", false},
		{"create grant", "create", hrmFixtureID, "AUTHORIZATION", "workforce_plan_lines", "deleteWorkforcePlanLine", true},
		{"unreferenced position", "delete", hrmFixtureID, `"deleted":true`, "positions", "deletePosition", true},
		{"referenced position", "delete", fixturePeriodID, "CONFLICT", "positions", "deletePosition", true},
		{"position delete without MFA", "delete", hrmFixtureID, "AUTHORIZATION", "positions", "deletePosition", false},
		{"workforce plan grant for position", "delete", hrmFixtureID, "AUTHORIZATION", "workforce_plans", "deletePosition", true},
		{"draft contract", "delete", fixturePeriodID, `"deleted":true`, "employment_contracts", "deleteEmploymentContract", true},
		{"signed contract", "delete", hrmFixtureID, "CONFLICT", "employment_contracts", "deleteEmploymentContract", true},
		{"contract delete without MFA", "delete", fixturePeriodID, "AUTHORIZATION", "employment_contracts", "deleteEmploymentContract", false},
		{"worker assignment delete grant for contract", "delete", fixturePeriodID, "AUTHORIZATION", "worker_assignments", "deleteEmploymentContract", true},
		{"empty plan", "delete", hrmFixtureID, `"deleted":true`, "workforce_plans", "deleteWorkforcePlan", true},
		{"plan with lines", "delete", fixturePeriodID, "CONFLICT", "workforce_plans", "deleteWorkforcePlan", true},
		{"line delete grant for plan", "delete", hrmFixtureID, "AUTHORIZATION", "workforce_plan_lines", "deleteWorkforcePlan", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.resource + "." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!) { " + tc.field + "(id:$id) { was_successful errors { type } deleted } }",
				"variables": map[string]any{"id": tc.id},
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) DeleteWorkforcePlan(_ context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	if req.GetId() == hrmFixtureID {
		return &common.EmptyResponse{}, nil
	}
	return nil, status.Error(codes.FailedPrecondition, "draft is not empty")
}

func (hrmWorkforceServer) UpdateWorkforcePlanLine(_ context.Context, req *workforce.UpdateWorkforcePlanLineRequest) (*workforce.WorkforcePlanLineResponse, error) {
	if req.GetId() == fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "plan is not a draft")
	}
	in := req.GetInput()
	if req.GetId() != hrmFixtureID || strings.Join(req.GetUpdateMask().GetPaths(), ",") != "required_fte,needed_until" ||
		in.GetRequiredFte() != "2.125" || in.NeededUntil != nil || in.WorkforcePlanId != nil {
		return nil, status.Error(codes.Internal, "unexpected patch")
	}
	return &workforce.WorkforcePlanLineResponse{Item: &workforce.WorkforcePlanLine{
		Id: req.GetId(), RequiredFte: in.GetRequiredFte(),
	}}, nil
}

func TestHRMUpdateWorkforcePlanLineThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, id, want string
		patch                  map[string]any
	}{
		{"partial patch and explicit null", "update", hrmFixtureID, `"required_fte":"2.125","needed_until":null`, map[string]any{"required_fte": "2.125", "needed_until": nil}},
		{"empty patch", "update", hrmFixtureID, "ARGUMENT", map[string]any{}},
		{"plan in review", "update", fixturePeriodID, "CONFLICT", map[string]any{"required_fte": "2.125"}},
		{"create grant", "create", hrmFixtureID, "AUTHORIZATION", map[string]any{"required_fte": "2.125"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workforce_plan_lines." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $input: UpdateWorkforcePlanLineInput!) { updateWorkforcePlanLine(id:$id,input:$input) { was_successful errors { type } workforce_plan_line { required_fte needed_until } } }",
				"variables": map[string]any{"id": tc.id, "input": tc.patch},
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) UpdateWorkforcePlan(_ context.Context, req *workforce.UpdateWorkforcePlanRequest) (*workforce.WorkforcePlanResponse, error) {
	if req.GetId() == fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "plan is not a draft")
	}
	in := req.GetInput()
	if req.GetId() != hrmFixtureID || strings.Join(req.GetUpdateMask().GetPaths(), ",") != "period_end" ||
		in.GetPeriodEnd() != "2026-12-31" || in.PeriodStart != nil || in.Status != nil || in.FinanceBudgetReferenceId != nil {
		return nil, status.Error(codes.Internal, "unexpected draft patch")
	}
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{
		Id: req.GetId(), Status: "draft", PeriodEnd: in.GetPeriodEnd(),
	}}, nil
}

func TestHRMUpdateWorkforcePlanThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, id, want string
		patch                  map[string]any
	}{
		{"period correction", "update", hrmFixtureID, `"period_end":"2026-12-31"`, map[string]any{"period_end": "2026-12-31"}},
		{"null date", "update", hrmFixtureID, "ARGUMENT", map[string]any{"period_end": nil}},
		{"plan in review", "update", fixturePeriodID, "CONFLICT", map[string]any{"period_end": "2026-12-31"}},
		{"create grant", "create", hrmFixtureID, "AUTHORIZATION", map[string]any{"period_end": "2026-12-31"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workforce_plans." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $input: UpdateWorkforcePlanInput!) { updateWorkforcePlan(id:$id,input:$input) { was_successful errors { type } workforce_plan { period_end } } }",
				"variables": map[string]any{"id": tc.id, "input": tc.patch},
			}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) ApproveWorkforcePlan(_ context.Context, req *common.IDRequest) (*workforce.WorkforcePlanResponse, error) {
	if req.GetId() == fixturePeriodID {
		return nil, status.Error(codes.Unavailable, "budget verifier unavailable")
	}
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{Id: req.GetId(), Status: "approved"}}, nil
}

func (hrmWorkforceServer) SupersedeWorkforcePlan(_ context.Context, req *workforce.SupersedeWorkforcePlanRequest) (*workforce.WorkforcePlanResponse, error) {
	if req.GetId() != hrmFixtureID || req.GetReplacementPlanId() != fixturePeriodID || req.GetReasonCode() != "new-version" {
		return nil, status.Error(codes.FailedPrecondition, "invalid replacement")
	}
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{
		Id: req.GetId(), Status: "superseded", SupersededByPlanId: &req.ReplacementPlanId,
	}}, nil
}

func TestHRMWorkforcePlanApprovalAndSupersessionThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	approve := "mutation($id: ID!) { approveWorkforcePlan(id:$id) { was_successful errors { type } workforce_plan { status } } }"
	supersede := "mutation($id: ID!, $replacement: ID!) { supersedeWorkforcePlan(id:$id,replacement_plan_id:$replacement,reason_code:\" new-version \") { was_successful errors { type } workforce_plan { status superseded_by_plan_id } } }"
	for _, tc := range []struct {
		name, grant, query, id, replacement, want string
		mfa                                       bool
	}{
		{"approved", "approve", approve, hrmFixtureID, "", `"status":"approved"`, true},
		{"Finance verifier unavailable", "approve", approve, fixturePeriodID, "", "UNAVAILABLE", true},
		{"superseded", "supersede", supersede, hrmFixtureID, fixturePeriodID, `"superseded_by_plan_id":"` + fixturePeriodID + `"`, true},
		{"self replacement", "supersede", supersede, hrmFixtureID, hrmFixtureID, "ARGUMENT", true},
		{"submit grant", "submit", approve, hrmFixtureID, "", "AUTHORIZATION", true},
		{"no MFA", "supersede", supersede, hrmFixtureID, fixturePeriodID, "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workforce_plans." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			variables := map[string]any{"id": tc.id}
			if tc.replacement != "" {
				variables["replacement"] = tc.replacement
			}
			request := map[string]any{"query": tc.query, "variables": variables}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmPayrollServer) ReissueEmployeePayslip(_ context.Context, req *payroll.ReissueEmployeePayslipRequest) (*payroll.ReissueEmployeePayslipResponse, error) {
	if req.GetWorkerId() != fixtureWorkerEntityID || req.GetPayrollResultId() != hrmFixtureID {
		return nil, status.Error(codes.NotFound, "payslip not found")
	}
	return &payroll.ReissueEmployeePayslipResponse{DocumentVersionId: fixtureDocumentVersion, Pdf: []byte("%PDF-1.7\nfixture")}, nil
}

func TestHRMReissueEmployeePayslipThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	const selection = `reissueEmployeePayslip(worker_id: "gid://integrione/hrm-worker/` + fixtureWorkerEntityID + `", payroll_result_id: "` + hrmFixtureID +
		`") { was_successful errors { type } payslip_pdf { content_base64 } }`
	for _, tc := range []struct {
		name, query, scope, want string
		mfa                      bool
	}{
		{"scoped employee", "mutation { " + selection + " }", "legal_entity", "JVBERi0xLjcKZml4dHVyZQ==", true},
		{"global only", "mutation { " + selection + " }", "global", "AUTHORIZATION", true},
		{"without MFA", "mutation { " + selection + " }", "legal_entity", "AUTHORIZATION", false},
		{"multiple PDF fields", "mutation { first:" + selection + " second:" + selection + " }", "legal_entity", "COMPLEXITY_LIMIT_EXCEEDED", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.payslips.reissue", Scope: authz.ScopeGrant{Type: tc.scope, ID: fixtureWorkerEntityID},
			})
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			resp := doGraphQLPOST(t, srv, map[string]any{"query": tc.query}, headers, cookies)
			if body := readBody(t, resp); !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}
