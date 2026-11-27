package server

import (
	"context"
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

const hrmFixtureID = "10000000-0000-4000-8000-000000000001"

type hrmWorkerServer struct {
	personnel.UnimplementedPersonnelServiceServer
}

func (hrmWorkerServer) GetWorkerByID(_ context.Context, req *common.IDRequest) (*personnel.WorkerResponse, error) {
	return &personnel.WorkerResponse{Item: &personnel.Worker{Id: req.GetId(), WorkerNumber: "W-FIXTURE"}}, nil
}

func (hrmWorkerServer) GetAllWorkers(_ context.Context, req *common.PageRequest) (*personnel.WorkersResponse, error) {
	return &personnel.WorkersResponse{
		Items: []*personnel.Worker{{Id: hrmFixtureID, WorkerNumber: "W-FIXTURE"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 2},
	}, nil
}

func (hrmWorkerServer) FindWorkerByWorkerNumber(_ context.Context, req *common.StringFinderRequest) (*personnel.WorkersResponse, error) {
	return &personnel.WorkersResponse{
		Items: []*personnel.Worker{{Id: hrmFixtureID, WorkerNumber: "W-FIXTURE"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 2},
	}, nil
}

const fixtureWorkerEntityID = "20000000-0000-4000-8000-000000000001"

func (hrmWorkerServer) FindWorkerByPrimaryLegalEntityReferenceID(_ context.Context, req *common.StringFinderRequest) (*personnel.WorkersResponse, error) {
	return &personnel.WorkersResponse{
		Items: []*personnel.Worker{{Id: hrmFixtureID, PrimaryLegalEntityReferenceId: fixtureWorkerEntityID}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

type hrmWorkforceServer struct {
	workforce.UnimplementedWorkforcePlanningServiceServer
}

func (hrmWorkforceServer) GetWorkforcePlanByID(_ context.Context, req *common.IDRequest) (*workforce.WorkforcePlanResponse, error) {
	return &workforce.WorkforcePlanResponse{Item: &workforce.WorkforcePlan{Id: req.GetId(), PlanCode: "WP-FIXTURE", PlanVersion: 2}}, nil
}

func (hrmWorkforceServer) GetAllWorkforcePlans(_ context.Context, req *common.PageRequest) (*workforce.WorkforcePlansResponse, error) {
	return &workforce.WorkforcePlansResponse{
		Items: []*workforce.WorkforcePlan{{Id: hrmFixtureID, PlanCode: "WP-FIXTURE"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 2},
	}, nil
}

const fixtureWorkforcePlanID = "40000000-0000-4000-8000-000000000001"

func (hrmWorkforceServer) GetWorkforcePlanLineByID(_ context.Context, req *common.IDRequest) (*workforce.WorkforcePlanLineResponse, error) {
	return &workforce.WorkforcePlanLineResponse{Item: &workforce.WorkforcePlanLine{
		Id: req.GetId(), WorkforcePlanId: fixtureWorkforcePlanID, RequiredFte: "2.50",
	}}, nil
}

func (hrmWorkforceServer) GetAllWorkforcePlanLines(_ context.Context, req *common.PageRequest) (*workforce.WorkforcePlanLinesResponse, error) {
	return &workforce.WorkforcePlanLinesResponse{Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset()}}, nil
}

func (hrmWorkforceServer) FindWorkforcePlanLineByWorkforcePlanID(_ context.Context, req *common.StringFinderRequest) (*workforce.WorkforcePlanLinesResponse, error) {
	return &workforce.WorkforcePlanLinesResponse{
		Items: []*workforce.WorkforcePlanLine{{Id: hrmFixtureID, WorkforcePlanId: req.GetValue()}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

type hrmLeaveServer struct {
	leave.UnimplementedLeaveServiceServer
}

func (hrmLeaveServer) GetLeavePolicyByID(_ context.Context, req *common.IDRequest) (*leave.LeavePolicyResponse, error) {
	return &leave.LeavePolicyResponse{Item: &leave.LeavePolicy{Id: req.GetId(), PolicyCode: "ANNUAL"}}, nil
}

func (hrmLeaveServer) GetAllLeavePolicies(_ context.Context, req *common.PageRequest) (*leave.LeavePoliciesResponse, error) {
	return &leave.LeavePoliciesResponse{
		Items: []*leave.LeavePolicy{{Id: hrmFixtureID, PolicyCode: "ANNUAL"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmLeaveServer) GetLeaveEntitlementByID(_ context.Context, req *common.IDRequest) (*leave.LeaveEntitlementResponse, error) {
	return &leave.LeaveEntitlementResponse{Item: &leave.LeaveEntitlement{
		Id: req.GetId(), WorkerId: hrmFixtureID, LeavePolicyId: fixtureWorkerEntityID, ClosingBalanceSnapshot: "12.50",
	}}, nil
}

func (hrmLeaveServer) FindLeaveEntitlementByWorkerID(_ context.Context, req *common.StringFinderRequest) (*leave.LeaveEntitlementsResponse, error) {
	return &leave.LeaveEntitlementsResponse{
		Items: []*leave.LeaveEntitlement{{Id: hrmFixtureID, WorkerId: hrmFixtureID, LeavePolicyId: fixtureWorkerEntityID, ClosingBalanceSnapshot: "12.50"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmLeaveServer) FindLeaveLedgerEntryByEntitlementID(_ context.Context, req *common.StringFinderRequest) (*leave.LeaveLedgerEntriesResponse, error) {
	return &leave.LeaveLedgerEntriesResponse{
		Items: []*leave.LeaveLedgerEntry{{Id: fixtureWorkerEntityID, EntitlementId: hrmFixtureID, EntrySequence: 9007199254740993, QuantityDelta: "-0.50"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmLeaveServer) FindLeaveAdjustmentByEntitlementID(_ context.Context, req *common.StringFinderRequest) (*leave.LeaveAdjustmentsResponse, error) {
	return &leave.LeaveAdjustmentsResponse{
		Items: []*leave.LeaveAdjustment{{Id: fixtureWorkerEntityID, EntitlementId: hrmFixtureID, QuantityDelta: "-0.50"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmLeaveServer) FindLeaveAdjustmentByIdempotencyKey(_ context.Context, req *common.StringFinderRequest) (*leave.LeaveAdjustmentsResponse, error) {
	return &leave.LeaveAdjustmentsResponse{
		Items: []*leave.LeaveAdjustment{{Id: fixtureWorkerEntityID, EntitlementId: hrmFixtureID, QuantityDelta: "-0.50", IdempotencyKey: "fixture-key"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmLeaveServer) GetLeaveAdjustmentByID(_ context.Context, req *common.IDRequest) (*leave.LeaveAdjustmentResponse, error) {
	if req.GetId() != fixtureWorkerEntityID {
		return nil, status.Error(codes.NotFound, "leave adjustment not found")
	}
	return &leave.LeaveAdjustmentResponse{Item: &leave.LeaveAdjustment{
		Id: req.GetId(), EntitlementId: hrmFixtureID, ReasonCode: "CORRECTION", QuantityDelta: "-0.50",
	}}, nil
}

func (hrmLeaveServer) GetLeaveLedgerEntryByID(_ context.Context, req *common.IDRequest) (*leave.LeaveLedgerEntryResponse, error) {
	if req.GetId() != fixtureWorkerEntityID {
		return nil, status.Error(codes.NotFound, "ledger entry not found")
	}
	return &leave.LeaveLedgerEntryResponse{Item: &leave.LeaveLedgerEntry{
		Id: req.GetId(), EntitlementId: hrmFixtureID, EntrySequence: 9007199254740993, QuantityDelta: "-0.50",
	}}, nil
}

func (hrmLeaveServer) FindLeaveLedgerEntryByAdjustmentID(_ context.Context, req *common.StringFinderRequest) (*leave.LeaveLedgerEntriesResponse, error) {
	response := &leave.LeaveLedgerEntriesResponse{Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset()}}
	if req.GetValue() == fixtureWorkerEntityID {
		adjustmentID := fixtureWorkerEntityID
		response.Items = []*leave.LeaveLedgerEntry{{Id: hrmFixtureID, EntitlementId: hrmFixtureID, AdjustmentId: &adjustmentID}}
		response.Page.Total = 1
	}
	return response, nil
}

func (hrmWorkforceServer) GetJobByID(_ context.Context, req *common.IDRequest) (*workforce.JobResponse, error) {
	return &workforce.JobResponse{Item: &workforce.Job{Id: req.GetId(), JobFamilyId: fixtureWorkerEntityID, JobCode: "OPS"}}, nil
}

func (hrmWorkforceServer) GetAllJobs(_ context.Context, req *common.PageRequest) (*workforce.JobsResponse, error) {
	return &workforce.JobsResponse{
		Items: []*workforce.Job{{Id: hrmFixtureID, JobFamilyId: fixtureWorkerEntityID, JobCode: "OPS"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmWorkforceServer) GetJobFamilyByID(_ context.Context, req *common.IDRequest) (*workforce.JobFamilyResponse, error) {
	return &workforce.JobFamilyResponse{Item: &workforce.JobFamily{Id: req.GetId(), FamilyCode: "RETAIL"}}, nil
}

func (hrmWorkforceServer) GetAllJobFamilies(_ context.Context, req *common.PageRequest) (*workforce.JobFamiliesResponse, error) {
	return &workforce.JobFamiliesResponse{
		Items: []*workforce.JobFamily{{Id: fixtureWorkerEntityID, FamilyCode: "RETAIL"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmWorkforceServer) FindJobByJobFamilyID(_ context.Context, req *common.StringFinderRequest) (*workforce.JobsResponse, error) {
	return &workforce.JobsResponse{
		Items: []*workforce.Job{{Id: hrmFixtureID, JobFamilyId: fixtureWorkerEntityID, JobCode: "FAMILY-JOB"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func fixtureHRMPosition(id string) *workforce.Position {
	return &workforce.Position{
		Id: id, PositionCode: "SHOP-01", JobId: hrmFixtureID, OrganizationUnitReferenceId: fixtureWorkerEntityID,
		LegalEntityReferenceId: fixtureWorkerEntityID, HeadcountAuthorized: "1.25",
	}
}

func (hrmWorkforceServer) GetPositionByID(_ context.Context, req *common.IDRequest) (*workforce.PositionResponse, error) {
	return &workforce.PositionResponse{Item: fixtureHRMPosition(req.GetId())}, nil
}

func (hrmWorkforceServer) GetAllPositions(_ context.Context, req *common.PageRequest) (*workforce.PositionsResponse, error) {
	return &workforce.PositionsResponse{
		Items: []*workforce.Position{fixtureHRMPosition(hrmFixtureID)},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmWorkforceServer) FindPositionByLegalEntityReferenceID(_ context.Context, req *common.StringFinderRequest) (*workforce.PositionsResponse, error) {
	return &workforce.PositionsResponse{
		Items: []*workforce.Position{fixtureHRMPosition(hrmFixtureID)},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func TestHRMQueriesThroughGateway(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	personnel.RegisterPersonnelServiceServer(grpcServer, hrmWorkerServer{})
	workforce.RegisterWorkforcePlanningServiceServer(grpcServer, hrmWorkforceServer{})
	leave.RegisterLeaveServiceServer(grpcServer, hrmLeaveServer{})
	t.Cleanup(grpcServer.Stop)
	go func() { _ = grpcServer.Serve(listener) }()
	conn, err := grpc.NewClient("passthrough:///hrm-fixture",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	const gid = "gid://integrione/hrm-worker/" + hrmFixtureID
	const workerRead = "human_resource.workers.read"
	const planRead = "human_resource.workforce_plans.read"
	const lineRead = "human_resource.workforce_plan_lines.read"
	cases := []struct {
		name, query, want string
		session           bool
		permission        string
	}{
		{"anonymous", "{ workers { total_count } }", "UNAUTHENTICATED", false, ""},
		{"without grant", "{ workers { total_count } }", "FORBIDDEN", true, ""},
		{"worker detail", `{ worker(id: "` + hrmFixtureID + `") { id worker_number } }`, `"id":"` + gid + `","worker_number":"W-FIXTURE"`, true, workerRead},
		{"worker node", `{ node(id: "` + gid + `") { __typename ... on Worker { worker_number } } }`, `"__typename":"Worker","worker_number":"W-FIXTURE"`, true, workerRead},
		{"worker page", "{ workers(first: 1, after: \"djE6MA\") { nodes { worker_number } total_count } }", `"nodes":[{"worker_number":"W-FIXTURE"}],"total_count":2`, true, workerRead},
		{"workforce detail", `{ workforcePlan(id: "` + hrmFixtureID + `") { id plan_code plan_version } }`, `"id":"gid://integrione/workforce-plan/` + hrmFixtureID + `","plan_code":"WP-FIXTURE","plan_version":2`, true, planRead},
		{"workforce node", `{ node(id: "gid://integrione/workforce-plan/` + hrmFixtureID + `") { __typename ... on WorkforcePlan { _id plan_code } } }`, `"__typename":"WorkforcePlan","_id":"` + hrmFixtureID + `","plan_code":"WP-FIXTURE"`, true, planRead},
		{"workforce node requires its own grant", `{ node(id: "gid://integrione/workforce-plan/` + hrmFixtureID + `") { id } }`, "FORBIDDEN", true, workerRead},
		{"wrong GID resource", `{ workforcePlan(id: "gid://integrione/job/` + hrmFixtureID + `") { id } }`, "INVALID_ARGUMENT", true, planRead},
		{"workforce page", "{ workforcePlans(first: 1, after: \"djE6MA\") { nodes { plan_code } total_count } }", `"nodes":[{"plan_code":"WP-FIXTURE"}],"total_count":2`, true, planRead},
		{"worker grant cannot read workforce", "{ workforcePlans { total_count } }", "FORBIDDEN", true, workerRead},
		{"workforce line detail", `{ workforcePlanLine(id: "` + hrmFixtureID + `") { workforce_plan_id required_fte } }`, `"workforce_plan_id":"` + fixtureWorkforcePlanID + `","required_fte":"2.50"`, true, lineRead},
		{"workforce lines by plan", `{ workforcePlanLines(workforce_plan_id: "` + fixtureWorkforcePlanID + `") { nodes { workforce_plan_id } total_count } }`, `"nodes":[{"workforce_plan_id":"` + fixtureWorkforcePlanID + `"}],"total_count":1`, true, lineRead},
		{"all visible workforce lines", "{ workforcePlanLines { nodes { id } total_count } }", `"nodes":[],"total_count":0`, true, lineRead},
		{"plan grant cannot read lines", "{ workforcePlanLines { total_count } }", "FORBIDDEN", true, planRead},
		{"worker number match while total changes", `{ workers(worker_number:"W-FIXTURE") { nodes { worker_number } total_count } }`, `"nodes":[{"worker_number":"W-FIXTURE"}],"total_count":2`, true, workerRead},
		{"blank worker number", `{ workers(worker_number:" ") { total_count } }`, "INVALID_ARGUMENT", true, workerRead},
		{"worker number mismatch", `{ workers(worker_number:"W-OTHER") { total_count } }`, "UNAVAILABLE", true, workerRead},
		{"workers by legal entity", `{ workers(legal_entity_id:"` + fixtureWorkerEntityID + `") { nodes { primary_legal_entity_reference_id } total_count } }`, `"primary_legal_entity_reference_id":"` + fixtureWorkerEntityID + `"}],"total_count":1`, true, workerRead},
		{"worker legal entity mismatch", `{ workers(legal_entity_id:"` + hrmFixtureID + `") { total_count } }`, "UNAVAILABLE", true, workerRead},
		{"both worker filters", `{ workers(legal_entity_id:"` + fixtureWorkerEntityID + `", worker_number:"W-1") { total_count } }`, "INVALID_ARGUMENT", true, workerRead},
		{"leave policy detail", `{ leavePolicy(id:"` + hrmFixtureID + `") { policy_code annual_entitlement } }`, `"policy_code":"ANNUAL","annual_entitlement":null`, true, "human_resource.leave_policies.read"},
		{"leave policy page", "{ leavePolicies { nodes { policy_code } total_count } }", `"nodes":[{"policy_code":"ANNUAL"}],"total_count":1`, true, "human_resource.leave_policies.read"},
		{"worker grant cannot read leave policy", "{ leavePolicies { total_count } }", "FORBIDDEN", true, workerRead},
		{"leave entitlement detail", `{ leaveEntitlement(id:"` + hrmFixtureID + `") { closing_balance } }`, `"closing_balance":"12.50"`, true, "human_resource.leave_entitlements.read"},
		{"leave entitlements by worker", `{ leaveEntitlements(worker_id:"` + hrmFixtureID + `") { nodes { closing_balance } total_count } }`, `"nodes":[{"closing_balance":"12.50"}],"total_count":1`, true, "human_resource.leave_entitlements.read"},
		{"policy grant cannot read entitlement", `{ leaveEntitlements(worker_id:"` + hrmFixtureID + `") { total_count } }`, "FORBIDDEN", true, "human_resource.leave_policies.read"},
		{"wrong entitlement worker", `{ leaveEntitlements(worker_id:"` + fixtureWorkerEntityID + `") { total_count } }`, "UNAVAILABLE", true, "human_resource.leave_entitlements.read"},
		{"leave ledger precision", `{ leaveLedgerEntries(entitlement_id:"` + hrmFixtureID + `") { nodes { entry_sequence quantity_delta } total_count } }`, `"nodes":[{"entry_sequence":"9007199254740993","quantity_delta":"-0.50"}],"total_count":1`, true, "human_resource.leave_ledger_entries.read"},
		{"entitlement grant cannot read ledger", `{ leaveLedgerEntries(entitlement_id:"` + hrmFixtureID + `") { total_count } }`, "FORBIDDEN", true, "human_resource.leave_entitlements.read"},
		{"wrong ledger entitlement", `{ leaveLedgerEntries(entitlement_id:"` + fixtureWorkerEntityID + `") { total_count } }`, "UNAVAILABLE", true, "human_resource.leave_ledger_entries.read"},
		{"pending adjustment", `{ leaveAdjustments(entitlement_id:"` + hrmFixtureID + `") { nodes { quantity_delta approved_at rejected_at } total_count } }`, `"nodes":[{"quantity_delta":"-0.50","approved_at":null,"rejected_at":null}],"total_count":1`, true, "human_resource.leave_adjustments.read"},
		{"ledger grant cannot read adjustments", `{ leaveAdjustments(entitlement_id:"` + hrmFixtureID + `") { total_count } }`, "FORBIDDEN", true, "human_resource.leave_ledger_entries.read"},
		{"wrong adjustment entitlement", `{ leaveAdjustments(entitlement_id:"` + fixtureWorkerEntityID + `") { total_count } }`, "UNAVAILABLE", true, "human_resource.leave_adjustments.read"},
		{"leave adjustment by idempotency key", `{ leaveAdjustments(idempotency_key:"fixture-key") { nodes { quantity_delta } total_count } }`, `"nodes":[{"quantity_delta":"-0.50"}],"total_count":1`, true, "human_resource.leave_adjustments.read"},
		{"adjustment finder returned wrong key", `{ leaveAdjustments(idempotency_key:"wrong") { total_count } }`, "UNAVAILABLE", true, "human_resource.leave_adjustments.read"},
		{"adjustment missing filter", "{ leaveAdjustments { total_count } }", "INVALID_ARGUMENT", true, "human_resource.leave_adjustments.read"},
		{"adjustment both filters", `{ leaveAdjustments(entitlement_id:"` + hrmFixtureID + `",idempotency_key:"fixture-key") { total_count } }`, "INVALID_ARGUMENT", true, "human_resource.leave_adjustments.read"},
		{"adjustment detail", `{ leaveAdjustment(id:"` + fixtureWorkerEntityID + `") { entitlement_id quantity_delta reason_code } }`, `"entitlement_id":"` + hrmFixtureID + `","quantity_delta":"-0.50","reason_code":"CORRECTION"`, true, "human_resource.leave_adjustments.read"},
		{"adjustment detail missing", `{ leaveAdjustment(id:"` + hrmFixtureID + `") { id } }`, "NOT_FOUND", true, "human_resource.leave_adjustments.read"},
		{"create grant cannot read adjustment detail", `{ leaveAdjustment(id:"` + fixtureWorkerEntityID + `") { id } }`, "FORBIDDEN", true, "human_resource.leave_adjustments.create"},
		{"ledger entry detail", `{ leaveLedgerEntry(id:"` + fixtureWorkerEntityID + `") { entry_sequence quantity_delta } }`, `"entry_sequence":"9007199254740993","quantity_delta":"-0.50"`, true, "human_resource.leave_ledger_entries.read"},
		{"ledger entry missing", `{ leaveLedgerEntry(id:"` + hrmFixtureID + `") { id } }`, "NOT_FOUND", true, "human_resource.leave_ledger_entries.read"},
		{"entitlement grant cannot read ledger detail", `{ leaveLedgerEntry(id:"` + fixtureWorkerEntityID + `") { id } }`, "FORBIDDEN", true, "human_resource.leave_entitlements.read"},
		{"posted adjustment ledger", `{ leaveLedgerEntries(adjustment_id:"` + fixtureWorkerEntityID + `") { nodes { adjustment_id } total_count } }`, `"nodes":[{"adjustment_id":"` + fixtureWorkerEntityID + `"}],"total_count":1`, true, "human_resource.leave_ledger_entries.read"},
		{"adjustment with no ledger", `{ leaveLedgerEntries(adjustment_id:"` + hrmFixtureID + `") { nodes { id } total_count } }`, `"nodes":[],"total_count":0`, true, "human_resource.leave_ledger_entries.read"},
		{"both ledger filters", `{ leaveLedgerEntries(entitlement_id:"` + hrmFixtureID + `",adjustment_id:"` + fixtureWorkerEntityID + `") { total_count } }`, "INVALID_ARGUMENT", true, "human_resource.leave_ledger_entries.read"},
		{"job detail", `{ job(id:"` + hrmFixtureID + `") { job_code job_level_code } }`, `"job_code":"OPS","job_level_code":null`, true, "human_resource.jobs.read"},
		{"job list", "{ jobs { nodes { job_code } total_count } }", `"nodes":[{"job_code":"OPS"}],"total_count":1`, true, "human_resource.jobs.read"},
		{"plan grant cannot read jobs", "{ jobs { total_count } }", "FORBIDDEN", true, planRead},
		{"job family detail", `{ jobFamily(id:"` + fixtureWorkerEntityID + `") { family_code description } }`, `"family_code":"RETAIL","description":null`, true, "human_resource.job_families.read"},
		{"job family list", "{ jobFamilies { nodes { family_code } total_count } }", `"nodes":[{"family_code":"RETAIL"}],"total_count":1`, true, "human_resource.job_families.read"},
		{"job grant cannot read job families", "{ jobFamilies { total_count } }", "FORBIDDEN", true, "human_resource.jobs.read"},
		{"jobs by family", `{ jobs(job_family_id:"` + fixtureWorkerEntityID + `") { nodes { job_code } total_count } }`, `"nodes":[{"job_code":"FAMILY-JOB"}],"total_count":1`, true, "human_resource.jobs.read"},
		{"jobs invalid family ID", `{ jobs(job_family_id:"invalid") { total_count } }`, "INVALID_ARGUMENT", true, "human_resource.jobs.read"},
		{"jobs mismatched family", `{ jobs(job_family_id:"` + hrmFixtureID + `") { total_count } }`, "UNAVAILABLE", true, "human_resource.jobs.read"},
		{"position detail", `{ position(id:"` + hrmFixtureID + `") { position_code headcount_authorized } }`, `"position_code":"SHOP-01","headcount_authorized":"1.25"`, true, "human_resource.positions.read"},
		{"position list", "{ positions { nodes { position_code } total_count } }", `"nodes":[{"position_code":"SHOP-01"}],"total_count":1`, true, "human_resource.positions.read"},
		{"job grant cannot read positions", "{ positions { total_count } }", "FORBIDDEN", true, "human_resource.jobs.read"},
		{"positions by entity", `{ positions(legal_entity_id:"` + fixtureWorkerEntityID + `") { nodes { position_code } total_count } }`, `"nodes":[{"position_code":"SHOP-01"}],"total_count":1`, true, "human_resource.positions.read"},
		{"positions invalid entity", `{ positions(legal_entity_id:"invalid") { total_count } }`, "INVALID_ARGUMENT", true, "human_resource.positions.read"},
		{"positions mismatched entity", `{ positions(legal_entity_id:"` + hrmFixtureID + `") { total_count } }`, "UNAVAILABLE", true, "human_resource.positions.read"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, rawToken, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			var cookies []*http.Cookie
			if tc.session {
				cookies = append(cookies, &http.Cookie{Name: config.SessionCookieName, Value: rawToken})
			}
			if tc.permission != "" {
				snap := store.sessions[authsession.HashOpaqueToken(rawToken)].Snapshot
				scope := authz.ScopeGrant{Type: "legal_entity", ID: fixtureWorkerEntityID}
				if strings.HasPrefix(tc.permission, "human_resource.leave_") || tc.permission == "human_resource.jobs.read" || tc.permission == "human_resource.job_families.read" || tc.permission == "human_resource.positions.read" {
					scope = authz.ScopeGrant{Type: authz.ScopeGlobal}
				}
				snap.Permissions = append(snap.Permissions, authz.PermissionGrant{
					Code:  tc.permission,
					Scope: scope,
				})
			}
			resp := doGraphQLPOST(t, srv, map[string]any{"query": tc.query}, nil, cookies)
			body := readBody(t, resp)
			if (strings.HasPrefix(tc.want, `"`) && strings.Contains(body, `"errors"`)) || !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %d %s", tc.want, resp.StatusCode, body)
			}
		})
	}
}

func (hrmWorkerServer) CreateWorker(_ context.Context, req *personnel.CreateWorkerRequest) (*personnel.WorkerResponse, error) {
	in := req.GetInput()
	if in.GetWorkerNumber() == "DUPLICATE" {
		return nil, status.Error(codes.AlreadyExists, "worker number already exists")
	}
	if in.GetPersonPartyReferenceId() != hrmFixtureID || in.GetPrimaryLegalEntityReferenceId() != fixtureWorkerEntityID ||
		in.Status != nil || in.IamPrincipalReferenceId != nil || in.PrimaryAssignmentId != nil {
		return nil, status.Error(codes.Internal, "unexpected worker draft")
	}
	return &personnel.WorkerResponse{Item: &personnel.Worker{Id: hrmFixtureID, Status: "preboarding"}}, nil
}

func (hrmWorkerServer) UpdateWorker(_ context.Context, req *personnel.UpdateWorkerRequest) (*personnel.WorkerResponse, error) {
	if req.GetId() == fixtureWorkerEntityID {
		return nil, status.Error(codes.FailedPrecondition, "existing worker references")
	}
	in := req.GetInput()
	if req.GetId() != hrmFixtureID || strings.Join(req.GetUpdateMask().GetPaths(), ",") != "service_start_date" ||
		in.ServiceStartDate != nil || in.HireDate != nil || in.Status != nil {
		return nil, status.Error(codes.Internal, "unexpected worker patch")
	}
	return &personnel.WorkerResponse{Item: &personnel.Worker{Id: req.GetId(), Status: "preboarding"}}, nil
}

func TestHRMWorkerPreboardingMutationsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	create := `mutation($number: String!) { createWorker(input:{worker_number:$number,worker_category:"employee",person_party_reference_id:"` + hrmFixtureID + `",primary_legal_entity_reference_id:"` + fixtureWorkerEntityID + `"}) { was_successful errors { type } worker { status } } }`
	update := `mutation($id: ID!) { updateWorker(id:$id,input:{service_start_date:null}) { was_successful errors { type } worker { status service_start_date } } }`
	for _, tc := range []struct {
		name, grant, query, key, value, want string
	}{
		{"create preboarding", "create", create, "number", "W-NEW", `"status":"preboarding"`},
		{"duplicate number", "create", create, "number", "DUPLICATE", "CONFLICT"},
		{"clear service date using Relay ID", "update", update, "id", "gid://integrione/hrm-worker/" + hrmFixtureID, `"service_start_date":null`},
		{"worker linked to another record", "update", update, "id", fixtureWorkerEntityID, "CONFLICT"},
		{"read grant used for update", "read", update, "id", hrmFixtureID, "AUTHORIZATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workers." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{"query": tc.query, "variables": map[string]any{tc.key: tc.value}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkerServer) ActivateWorker(_ context.Context, req *common.IDRequest) (*personnel.WorkerResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "missing activation prerequisites")
	}
	return &personnel.WorkerResponse{Item: &personnel.Worker{Id: req.GetId(), Status: "active"}}, nil
}

func (hrmWorkerServer) TerminateWorker(_ context.Context, req *personnel.TerminateWorkerRequest) (*personnel.WorkerResponse, error) {
	if req.GetWorkerId() != hrmFixtureID || req.GetOffboardingPlanId() != fixtureWorkerEntityID {
		return nil, status.Error(codes.FailedPrecondition, "offboarding plan mismatch")
	}
	return &personnel.WorkerResponse{Item: &personnel.Worker{Id: req.GetWorkerId(), Status: "terminated"}}, nil
}

func TestHRMWorkerLifecycleThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	activate := "mutation($id: ID!) { activateWorker(id:$id) { was_successful errors { type } worker { status } } }"
	terminate := "mutation($id: ID!, $plan: ID!) { terminateWorker(id:$id,offboarding_plan_id:$plan) { was_successful errors { type } worker { status } } }"
	for _, tc := range []struct {
		name, grant, query, id, plan, want string
		mfa                                bool
	}{
		{"activation", "activate", activate, "gid://integrione/hrm-worker/" + hrmFixtureID, "", `"status":"active"`, true},
		{"activation prerequisites", "activate", activate, fixtureWorkerEntityID, "", "CONFLICT", true},
		{"termination", "terminate", terminate, hrmFixtureID, fixtureWorkerEntityID, `"status":"terminated"`, true},
		{"offboarding mismatch", "terminate", terminate, hrmFixtureID, hrmFixtureID, "CONFLICT", true},
		{"invalid plan ID", "terminate", terminate, hrmFixtureID, "invalid", "ARGUMENT", true},
		{"activate grant for termination", "activate", terminate, hrmFixtureID, fixtureWorkerEntityID, "AUTHORIZATION", true},
		{"activation without MFA", "activate", activate, hrmFixtureID, "", "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workers." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			variables := map[string]any{"id": tc.id}
			if tc.plan != "" {
				variables["plan"] = tc.plan
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

func (hrmWorkerServer) ApplyPrimaryAssignment(_ context.Context, req *personnel.ApplyPrimaryAssignmentRequest) (*personnel.WorkerResponse, error) {
	if req.GetWorkerId() != hrmFixtureID || req.GetAssignmentId() != fixtureWorkerEntityID {
		return nil, status.Error(codes.FailedPrecondition, "assignment does not match worker")
	}
	return &personnel.WorkerResponse{Item: &personnel.Worker{
		Id: req.GetWorkerId(), Status: "active", PrimaryAssignmentId: &req.AssignmentId,
	}}, nil
}

func TestHRMWorkerPrimaryAssignmentThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	assign := `mutation($reference: ID!) { applyPrimaryAssignment(id:"` + hrmFixtureID + `",assignment_id:$reference) { was_successful errors { type } worker { status } } }`
	for _, tc := range []struct {
		name, action, query, reference, want string
		mfa                                  bool
	}{
		{"assign", "apply_primary_assignment", assign, fixtureWorkerEntityID, `"status":"active"`, true},
		{"assignment mismatch", "apply_primary_assignment", assign, hrmFixtureID, "CONFLICT", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.workers." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{"query": tc.query, "variables": map[string]any{"reference": tc.reference}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) ActivatePosition(_ context.Context, req *common.IDRequest) (*workforce.PositionResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "position is not planned")
	}
	position := fixtureHRMPosition(req.GetId())
	position.PositionStatus = "active"
	return &workforce.PositionResponse{Item: position}, nil
}

func TestHRMActivatePositionThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, id, want string
		mfa                    bool
	}{
		{"planned position", "activate", hrmFixtureID, `"position_status":"active"`, true},
		{"workflow conflict", "activate", fixtureWorkerEntityID, "CONFLICT", true},
		{"invalid ID", "activate", "invalid", "ARGUMENT", true},
		{"without MFA", "activate", hrmFixtureID, "AUTHORIZATION", false},
		{"read grant", "read", hrmFixtureID, "AUTHORIZATION", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.positions." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!) { activatePosition(id:$id) { was_successful errors { type } position { position_status } } }",
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

func (hrmWorkerServer) ApproveWorkerAssignment(_ context.Context, req *common.IDRequest) (*personnel.WorkerAssignmentResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "capacity exceeded")
	}
	approverID := fixturePeriodID
	return &personnel.WorkerAssignmentResponse{Item: &personnel.WorkerAssignment{
		Id: req.GetId(), WorkerId: hrmFixtureID, PositionId: fixtureWorkerEntityID, FtePercent: "87.50",
		ApprovedByPrincipalId: &approverID, ApprovedAt: timestamppb.New(time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)),
	}}, nil
}

func TestHRMApproveWorkerAssignmentThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, id, want string
		mfa                    bool
	}{
		{"approval", "approve", hrmFixtureID, `"fte_percent":"87.50","approved_at":"2026-11-27T00:00:00Z"`, true},
		{"capacity conflict", "approve", fixtureWorkerEntityID, "CONFLICT", true},
		{"invalid ID", "approve", "invalid", "ARGUMENT", true},
		{"without MFA", "approve", hrmFixtureID, "AUTHORIZATION", false},
		{"read grant", "read", hrmFixtureID, "AUTHORIZATION", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.worker_assignments." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!) { approveWorkerAssignment(id:$id) { was_successful errors { type } worker_assignment { fte_percent approved_at } } }",
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

func (hrmWorkerServer) GetWorkerAssignmentByID(_ context.Context, req *common.IDRequest) (*personnel.WorkerAssignmentResponse, error) {
	return &personnel.WorkerAssignmentResponse{Item: &personnel.WorkerAssignment{
		Id: req.GetId(), WorkerId: hrmFixtureID, PositionId: fixtureWorkerEntityID, FtePercent: "87.50",
	}}, nil
}

func (hrmWorkerServer) FindWorkerAssignmentByWorkerID(_ context.Context, req *common.StringFinderRequest) (*personnel.WorkerAssignmentsResponse, error) {
	return &personnel.WorkerAssignmentsResponse{
		Items: []*personnel.WorkerAssignment{{Id: fixturePeriodID, WorkerId: hrmFixtureID, PositionId: fixtureWorkerEntityID}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmWorkerServer) FindWorkerAssignmentByPositionID(_ context.Context, req *common.StringFinderRequest) (*personnel.WorkerAssignmentsResponse, error) {
	return &personnel.WorkerAssignmentsResponse{
		Items: []*personnel.WorkerAssignment{{Id: fixturePeriodID, WorkerId: hrmFixtureID, PositionId: fixtureWorkerEntityID}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func TestHRMWorkerAssignmentQueriesThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	detail := "query($id: ID!) { workerAssignment(id:$id) { fte_percent approved_at } }"
	list := "query($id: ID!) { workerAssignments(worker_id:$id) { nodes { worker_id } total_count } }"
	byPosition := "query($id: ID!) { workerAssignments(position_id:$id) { nodes { position_id } total_count } }"
	both := "query($id: ID!) { workerAssignments(worker_id:$id,position_id:$id) { total_count } }"
	for _, tc := range []struct {
		name, grant, query, id, want string
	}{
		{"unapproved detail", "read", detail, fixturePeriodID, `"fte_percent":"87.50","approved_at":null`},
		{"list using Relay ID", "read", list, "gid://integrione/hrm-worker/" + hrmFixtureID, `"total_count":1`},
		{"RPC returns another worker", "read", list, fixtureWorkerEntityID, "UNAVAILABLE"},
		{"list by position", "read", byPosition, fixtureWorkerEntityID, `"position_id":"` + fixtureWorkerEntityID + `"`},
		{"RPC returns another position", "read", byPosition, hrmFixtureID, "UNAVAILABLE"},
		{"both filters", "read", both, fixtureWorkerEntityID, "INVALID_ARGUMENT"},
		{"approve grant for details", "approve", detail, fixturePeriodID, "FORBIDDEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.worker_assignments." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{"query": tc.query, "variables": map[string]any{"id": tc.id}}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkerServer) CreateWorkerAssignment(_ context.Context, req *personnel.CreateWorkerAssignmentRequest) (*personnel.WorkerAssignmentResponse, error) {
	in := req.GetInput()
	if in.GetPositionId() == hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "position does not meet prerequisites")
	}
	if in.GetWorkerId() != hrmFixtureID || in.GetPositionId() != fixtureWorkerEntityID || in.GetFtePercent() != "87.50" ||
		in.GetManagerWorkerId() != fixturePeriodID || in.GetValidFrom() != "2026-12-01" ||
		in.ApprovedAt != nil || in.ApprovedByPrincipalId != nil || in.EmploymentStatusSnapshot != nil {
		return nil, status.Error(codes.Internal, "invalid assignment request")
	}
	return &personnel.WorkerAssignmentResponse{Item: &personnel.WorkerAssignment{
		Id: fixturePeriodID, WorkerId: in.GetWorkerId(), PositionId: in.GetPositionId(), FtePercent: in.GetFtePercent(),
		RequestedAt: timestamppb.New(time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)),
	}}, nil
}

func TestHRMCreateWorkerAssignmentThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, positionID, want string
	}{
		{"requested", "create", fixtureWorkerEntityID, `"requested_at":"2026-11-27T00:00:00Z","approved_at":null`},
		{"position conflict", "create", hrmFixtureID, "CONFLICT"},
		{"invalid position", "create", "invalid", "ARGUMENT"},
		{"approve grant", "approve", fixtureWorkerEntityID, "AUTHORIZATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.worker_assignments." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreateWorkerAssignmentInput!) { createWorkerAssignment(input:$input) { was_successful errors { type } worker_assignment { requested_at approved_at } } }",
				"variables": map[string]any{"input": map[string]any{
					"worker_id": "gid://integrione/hrm-worker/" + hrmFixtureID, "position_id": tc.positionID,
					"manager_worker_id": "gid://integrione/hrm-worker/" + fixturePeriodID,
					"assignment_type":   "primary", "fte_percent": "87.50", "valid_from": "2026-12-01",
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

func (hrmWorkerServer) SubmitWorkerAssignmentTransfer(_ context.Context, req *personnel.SubmitWorkerAssignmentTransferRequest) (*personnel.SubmitWorkerAssignmentTransferResponse, error) {
	if req.GetIdempotencyKey() == "conflict" {
		return nil, status.Error(codes.FailedPrecondition, "idempotency conflict")
	}
	if req.GetWorkerId() != hrmFixtureID || req.GetSourceAssignmentId() != fixtureWorkerEntityID || req.GetTargetPositionId() != fixturePeriodID {
		return nil, status.Error(codes.Internal, "invalid transfer request")
	}
	item := &personnel.WorkerAssignmentTransfer{
		Id: hrmFixtureID, WorkerId: req.GetWorkerId(), SourceAssignmentId: req.GetSourceAssignmentId(),
		TargetPositionId: req.GetTargetPositionId(), EffectiveDate: req.GetEffectiveDate(), Status: "submitted",
		ReasonCode: "rotation", RequestedAt: timestamppb.New(time.Unix(100, 0)),
	}
	switch req.GetIdempotencyKey() {
	case "retry-key":
	case "already-applied":
		item.Status = "applied"
		item.DecidedAt = timestamppb.New(time.Unix(200, 0))
		id := fixtureWorkforcePlanID
		item.AppliedAssignmentId = &id
	case "already-rejected":
		item.Status = "rejected"
		item.DecidedAt = timestamppb.New(time.Unix(200, 0))
	case "invalid-applied":
		item.Status = "applied"
		item.DecidedAt = timestamppb.New(time.Unix(200, 0))
	default:
		return nil, status.Error(codes.Internal, "unexpected idempotency key")
	}
	return &personnel.SubmitWorkerAssignmentTransferResponse{Item: item}, nil
}

func (hrmWorkerServer) DecideWorkerAssignmentTransfer(_ context.Context, req *personnel.DecideWorkerAssignmentTransferRequest) (*personnel.DecideWorkerAssignmentTransferResponse, error) {
	item := &personnel.WorkerAssignmentTransfer{
		Id: req.GetTransferId(), WorkerId: hrmFixtureID, SourceAssignmentId: fixtureWorkerEntityID, TargetPositionId: fixturePeriodID,
		ReasonCode: "rotation", RequestedAt: timestamppb.New(time.Unix(100, 0)),
	}
	switch req.GetDecision() {
	case personnel.TransferDecision_TRANSFER_DECISION_APPROVE:
		item.Status = "applied"
		item.DecidedAt = timestamppb.New(time.Unix(200, 0))
		appliedID := fixtureWorkforcePlanID
		item.AppliedAssignmentId = &appliedID
	case personnel.TransferDecision_TRANSFER_DECISION_REJECT:
		item.Status = "rejected"
		item.DecidedAt = timestamppb.New(time.Unix(200, 0))
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid decision")
	}
	return &personnel.DecideWorkerAssignmentTransferResponse{Item: item}, nil
}

func TestHRMWorkerTransferThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	submit := `mutation($key: String!) { submitWorkerAssignmentTransfer(input:{
		worker_id:"gid://integrione/hrm-worker/` + hrmFixtureID + `", source_assignment_id:"` + fixtureWorkerEntityID +
		`", target_position_id:"` + fixturePeriodID + `", effective_date:"2026-12-01", reason_code:"staffing", idempotency_key:$key
	}) { was_successful errors { type } worker_assignment_transfer { status applied_assignment_id reason_code requested_at decided_at } } }`
	decide := `mutation($decision: WorkerAssignmentTransferDecision!) { decideWorkerAssignmentTransfer(id:"` + hrmFixtureID +
		`",decision:$decision) { was_successful errors { type } worker_assignment_transfer { status applied_assignment_id reason_code requested_at decided_at } } }`
	for _, tc := range []struct {
		name, grant, query, key, value, want string
		mfa                                  bool
	}{
		{"submit", "create", submit, "key", "retry-key", `"status":"submitted","applied_assignment_id":null,"reason_code":"rotation","requested_at":"1970-01-01T00:01:40Z","decided_at":null`, false},
		{"replay applied", "create", submit, "key", "already-applied", `"status":"applied","applied_assignment_id":"` + fixtureWorkforcePlanID + `"`, false},
		{"replay rejected", "create", submit, "key", "already-rejected", `"status":"rejected","applied_assignment_id":null`, false},
		{"applied replay without assignment", "create", submit, "key", "invalid-applied", "UNAVAILABLE", false},
		{"key conflict", "create", submit, "key", "conflict", "CONFLICT", false},
		{"approve", "transfer", decide, "decision", "APPROVE", `"decided_at":"1970-01-01T00:03:20Z"`, true},
		{"reject", "transfer", decide, "decision", "REJECT", `"status":"rejected","applied_assignment_id":null`, true},
		{"decision without MFA", "transfer", decide, "decision", "APPROVE", "AUTHORIZATION", false},
		{"create grant cannot decide", "create", decide, "decision", "APPROVE", "AUTHORIZATION", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.worker_assignments." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{"query": tc.query, "variables": map[string]any{tc.key: tc.value}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkforceServer) CreatePosition(_ context.Context, req *workforce.CreatePositionRequest) (*workforce.PositionResponse, error) {
	in := req.GetInput()
	if in.GetPositionCode() == "duplicate" {
		return nil, status.Error(codes.AlreadyExists, "duplicate position")
	}
	if in.PositionStatus != nil || in.GetJobId() != hrmFixtureID || in.GetOrganizationUnitReferenceId() != fixtureWorkerEntityID ||
		in.GetLegalEntityReferenceId() != fixtureWorkerEntityID || in.GetHeadcountAuthorized() != "1.25" {
		return nil, status.Error(codes.Internal, "unexpected position input")
	}
	position := fixtureHRMPosition(hrmFixtureID)
	position.PositionStatus = "planned"
	position.WorkLocationReferenceId = in.WorkLocationReferenceId
	position.ReportsToPositionId = in.ReportsToPositionId
	switch in.GetPositionCode() {
	case "wrong-location":
		position.WorkLocationReferenceId = new(fixtureWorkforcePlanID)
	case "missing-parent":
		position.ReportsToPositionId = nil
	}
	return &workforce.PositionResponse{Item: position}, nil
}

func TestHRMCreatePositionThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, action, code, jobID, want string
		mfa                             bool
	}{
		{"planned", "create", "SHOP-01", hrmFixtureID, `"position_status":"planned","headcount_authorized":"1.25"`, true},
		{"wrong location", "create", "wrong-location", hrmFixtureID, "UNAVAILABLE", true},
		{"missing reporting position", "create", "missing-parent", hrmFixtureID, "UNAVAILABLE", true},
		{"duplicate", "create", "duplicate", hrmFixtureID, "CONFLICT", true},
		{"invalid job", "create", "SHOP-01", "invalid", "ARGUMENT", true},
		{"without MFA", "create", "SHOP-01", hrmFixtureID, "AUTHORIZATION", false},
		{"activate grant", "activate", "SHOP-01", hrmFixtureID, "AUTHORIZATION", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.positions." + tc.action, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreatePositionInput!) { createPosition(input:$input) { was_successful errors { type } position { position_status headcount_authorized } } }",
				"variables": map[string]any{"input": map[string]any{
					"position_code": tc.code, "position_title": "Store operator", "job_id": tc.jobID,
					"organization_unit_reference_id": fixtureWorkerEntityID, "legal_entity_reference_id": fixtureWorkerEntityID,
					"headcount_authorized": "1.25", "valid_from": "2026-12-01",
					"work_location_reference_id": fixtureWorkerEntityID, "reports_to_position_id": fixtureWorkforcePlanID,
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

func (hrmWorkforceServer) UpdatePosition(_ context.Context, req *workforce.UpdatePositionRequest) (*workforce.PositionResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "position is not planned")
	}
	in := req.GetInput()
	if strings.Join(req.GetUpdateMask().GetPaths(), ",") == "work_location_reference_id" {
		item := fixtureHRMPosition(req.GetId())
		item.PositionStatus = "planned"
		item.WorkLocationReferenceId = new(fixtureWorkerEntityID)
		return &workforce.PositionResponse{Item: item}, nil
	}
	if strings.Join(req.GetUpdateMask().GetPaths(), ",") != "headcount_authorized,valid_to" ||
		in.GetHeadcountAuthorized() != "2.75" || in.ValidTo != nil || in.PositionStatus != nil || in.CostCenterReferenceId != nil {
		return nil, status.Error(codes.Internal, "unexpected position patch")
	}
	item := fixtureHRMPosition(req.GetId())
	item.PositionStatus, item.HeadcountAuthorized = "planned", in.GetHeadcountAuthorized()
	return &workforce.PositionResponse{Item: item}, nil
}

func TestHRMUpdatePositionThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	patch := map[string]any{"headcount_authorized": "2.75", "valid_to": nil}
	for _, tc := range []struct {
		name, grant, id, want string
		patch                 map[string]any
	}{
		{"partial patch", "update", hrmFixtureID, `"headcount_authorized":"2.75","valid_to":null`, patch},
		{"matching location", "update", hrmFixtureID, `"work_location_reference_id":"` + fixtureWorkerEntityID + `"`, map[string]any{"work_location_reference_id": fixtureWorkerEntityID}},
		{"different location", "update", hrmFixtureID, "UNAVAILABLE", map[string]any{"work_location_reference_id": fixtureWorkforcePlanID}},
		{"location not cleared", "update", hrmFixtureID, "UNAVAILABLE", map[string]any{"work_location_reference_id": nil}},
		{"active position", "update", fixtureWorkerEntityID, "CONFLICT", patch},
		{"null capacity", "update", hrmFixtureID, "ARGUMENT", map[string]any{"headcount_authorized": nil}},
		{"empty patch", "update", hrmFixtureID, "ARGUMENT", map[string]any{}},
		{"create grant", "create", hrmFixtureID, "AUTHORIZATION", patch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = true
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.positions." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $input: UpdatePositionInput!) { updatePosition(id:$id,input:$input) { was_successful errors { type } position { headcount_authorized valid_to work_location_reference_id } } }",
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

func (hrmWorkforceServer) DeletePosition(_ context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	if req.GetId() != hrmFixtureID {
		return nil, status.Error(codes.FailedPrecondition, "position is in use")
	}
	return &common.EmptyResponse{}, nil
}

func (hrmWorkerServer) FindEmploymentContractByWorkerID(_ context.Context, req *common.StringFinderRequest) (*personnel.EmploymentContractsResponse, error) {
	return &personnel.EmploymentContractsResponse{
		Items: []*personnel.EmploymentContract{{
			Id: fixturePeriodID, WorkerId: hrmFixtureID, EmployingLegalEntityReferenceId: fixtureWorkerEntityID, ContractStatus: "signed",
		}},
		Page: &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func (hrmWorkerServer) GetEmploymentContractByID(_ context.Context, req *common.IDRequest) (*personnel.EmploymentContractResponse, error) {
	if req.GetId() == fixtureWorkforcePlanID {
		return nil, status.Error(codes.NotFound, "contract not found")
	}
	return &personnel.EmploymentContractResponse{Item: &personnel.EmploymentContract{
		Id: fixturePeriodID, WorkerId: hrmFixtureID, EmployingLegalEntityReferenceId: fixtureWorkerEntityID, ContractStatus: "signed",
	}}, nil
}

func TestHRMEmploymentContractsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, id, want string
		mfa, detail           bool
	}{
		{"signed contract", "employment_contracts.read", "gid://integrione/hrm-worker/" + hrmFixtureID, `"contract_status":"signed"`, true, false},
		{"mismatched worker", "employment_contracts.read", fixtureWorkerEntityID, "UNAVAILABLE", true, false},
		{"no MFA", "employment_contracts.read", hrmFixtureID, "FORBIDDEN", false, false},
		{"worker read grant", "workers.read", hrmFixtureID, "FORBIDDEN", true, false},
		{"contract detail", "employment_contracts.read", fixturePeriodID, `"contract_status":"signed"`, true, true},
		{"contract detail missing", "employment_contracts.read", fixtureWorkforcePlanID, "NOT_FOUND", true, true},
		{"wrong contract returned", "employment_contracts.read", hrmFixtureID, "UNAVAILABLE", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "query($id: ID!) { employmentContracts(worker_id:$id) { nodes { contract_status } } }",
				"variables": map[string]any{"id": tc.id},
			}
			if tc.detail {
				request["query"] = "query($id: ID!) { employmentContract(id:$id) { contract_status } }"
			}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmWorkerServer) CreateEmploymentContract(_ context.Context, req *personnel.CreateEmploymentContractRequest) (*personnel.EmploymentContractResponse, error) {
	in := req.GetInput()
	if in.GetContractNumber() == "duplicate" {
		return nil, status.Error(codes.AlreadyExists, "contract already exists")
	}
	if in.GetWorkerId() != hrmFixtureID || in.GetEmployingLegalEntityReferenceId() != fixtureWorkerEntityID ||
		in.ContractStatus != nil || in.SignedAt != nil || in.SignedDocumentVersionReferenceId != nil || in.GetNoticePeriodDays() != 30 {
		return nil, status.Error(codes.Internal, "unexpected draft contract input")
	}
	return &personnel.EmploymentContractResponse{Item: &personnel.EmploymentContract{
		Id: fixturePeriodID, WorkerId: in.GetWorkerId(), EmployingLegalEntityReferenceId: in.GetEmployingLegalEntityReferenceId(),
		ContractStatus: "draft", ContractNumber: in.GetContractNumber(),
	}}, nil
}

func TestHRMCreateEmploymentContractThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, number, want string
		mfa                       bool
		days                      int
	}{
		{"draft", "create", "CON-01", `"contract_status":"draft","signed_at":null`, true, 30},
		{"duplicate", "create", "duplicate", "CONFLICT", true, 30},
		{"invalid notice", "create", "CON-01", "ARGUMENT", true, -1},
		{"no MFA", "create", "CON-01", "AUTHORIZATION", false, 30},
		{"read grant", "read", "CON-01", "AUTHORIZATION", true, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.employment_contracts." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreateEmploymentContractInput!) { createEmploymentContract(input:$input) { was_successful errors { type } employment_contract { contract_status signed_at } } }",
				"variables": map[string]any{"input": map[string]any{
					"worker_id": "gid://integrione/hrm-worker/" + hrmFixtureID, "contract_number": tc.number, "contract_type": "fixed_term",
					"employing_legal_entity_reference_id": fixtureWorkerEntityID, "effective_from": "2026-12-01",
					"work_country_code": "ID", "governing_jurisdiction_code": "ID", "notice_period_days": tc.days,
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

const (
	missingSignatureDocumentVersionID = "20000000-0000-4000-8000-000000000002"
	invalidSignatureDigestVersionID   = "20000000-0000-4000-8000-000000000003"
)

func (hrmWorkerServer) RecordSignedEmploymentContract(_ context.Context, req *personnel.RecordSignedEmploymentContractRequest) (*personnel.EmploymentContractResponse, error) {
	if req.GetContractId() != fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "contract is not a draft")
	}
	if req.GetDocumentVersionId() == hrmFixtureID {
		return nil, status.Error(codes.Unavailable, "signature verifier unavailable")
	}
	item := &personnel.EmploymentContract{
		Id: req.GetContractId(), WorkerId: hrmFixtureID, EmployingLegalEntityReferenceId: fixtureWorkerEntityID,
		ContractStatus: "signed", SignedAt: timestamppb.New(time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)),
	}
	version := fixtureWorkforcePlanID
	item.SignedDocumentVersionReferenceId = &version
	documentID := fixtureWorkerEntityID
	item.SignedDocumentReferenceId = &documentID
	item.SignedContentDigest = make([]byte, 32)
	switch req.GetDocumentVersionId() {
	case missingSignatureDocumentVersionID:
		item.SignedDocumentVersionReferenceId = &req.DocumentVersionId
		item.SignedDocumentReferenceId = nil
	case invalidSignatureDigestVersionID:
		item.SignedDocumentVersionReferenceId = &req.DocumentVersionId
		item.SignedContentDigest = []byte{1}
	}
	return &personnel.EmploymentContractResponse{Item: item}, nil
}

func TestHRMRecordSignedEmploymentContractThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, contractID, versionID, want string
		mfa                                      bool
	}{
		{"verified signature", "record_signature", fixturePeriodID, fixtureWorkforcePlanID, `"contract_status":"signed"`, true},
		{"wrong response version", "record_signature", fixturePeriodID, fixtureWorkerEntityID, "UNAVAILABLE", true},
		{"missing signed document", "record_signature", fixturePeriodID, missingSignatureDocumentVersionID, "UNAVAILABLE", true},
		{"invalid signature digest", "record_signature", fixturePeriodID, invalidSignatureDigestVersionID, "UNAVAILABLE", true},
		{"not a draft", "record_signature", hrmFixtureID, fixtureWorkforcePlanID, "CONFLICT", true},
		{"without MFA", "record_signature", fixturePeriodID, fixtureWorkforcePlanID, "AUTHORIZATION", false},
		{"create grant", "create", fixturePeriodID, fixtureWorkforcePlanID, "AUTHORIZATION", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.employment_contracts." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $version: ID!) { recordSignedEmploymentContract(id:$id,document_version_id:$version) { was_successful errors { type } employment_contract { contract_status } } }",
				"variables": map[string]any{"id": tc.contractID, "version": tc.versionID},
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

func (hrmWorkerServer) DeleteEmploymentContract(_ context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
	if req.GetId() != fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "contract is not an unsigned draft")
	}
	return &common.EmptyResponse{}, nil
}

func (hrmWorkerServer) UpdateEmploymentContract(_ context.Context, req *personnel.UpdateEmploymentContractRequest) (*personnel.EmploymentContractResponse, error) {
	if req.GetId() != fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "contract is no longer a draft")
	}
	in := req.GetInput()
	if strings.Join(req.GetUpdateMask().GetPaths(), ",") != "contract_number,effective_to,notice_period_days" ||
		in.GetContractNumber() != "CORRECTED" || in.EffectiveTo != nil || in.NoticePeriodDays != nil || in.ContractStatus != nil {
		return nil, status.Error(codes.Internal, "unexpected contract patch")
	}
	return &personnel.EmploymentContractResponse{Item: &personnel.EmploymentContract{
		Id: req.GetId(), WorkerId: hrmFixtureID, EmployingLegalEntityReferenceId: fixtureWorkerEntityID,
		ContractStatus: "draft", ContractNumber: in.GetContractNumber(),
	}}, nil
}

func TestHRMUpdateEmploymentContractThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	patch := map[string]any{"contract_number": "CORRECTED", "effective_to": nil, "notice_period_days": nil}
	for _, tc := range []struct {
		name, grant, id, want string
		patch                 map[string]any
	}{
		{"partial patch", "update", fixturePeriodID, `"contract_number":"CORRECTED"`, patch},
		{"signed contract", "update", hrmFixtureID, "CONFLICT", patch},
		{"required field null", "update", fixturePeriodID, "ARGUMENT", map[string]any{"contract_number": nil}},
		{"empty patch", "update", fixturePeriodID, "ARGUMENT", map[string]any{}},
		{"read grant", "read", fixturePeriodID, "AUTHORIZATION", patch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = true
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.employment_contracts." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $input: UpdateEmploymentContractInput!) { updateEmploymentContract(id:$id,input:$input) { was_successful errors { type } employment_contract { contract_number } } }",
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

type hrmOffboardingServer struct {
	offboarding.UnimplementedOffboardingServiceServer
}

func fixtureOffboardingPlan() *offboarding.OffboardingPlan {
	day := timestamppb.New(time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC))
	return &offboarding.OffboardingPlan{
		Id: fixtureWorkforcePlanID, WorkerId: hrmFixtureID, PlanNumber: "OFF-01", Status: "in_progress",
		LastWorkingDate: day, TerminationEffectiveDate: day,
	}
}

func (hrmOffboardingServer) GetOffboardingPlanByID(_ context.Context, req *common.IDRequest) (*offboarding.OffboardingPlanResponse, error) {
	if req.GetId() != fixtureWorkforcePlanID {
		return nil, status.Error(codes.NotFound, "offboarding plan not found")
	}
	return &offboarding.OffboardingPlanResponse{Value: fixtureOffboardingPlan()}, nil
}

func (hrmOffboardingServer) FindOffboardingPlanByWorkerID(_ context.Context, req *common.StringFinderRequest) (*offboarding.OffboardingPlansResponse, error) {
	return &offboarding.OffboardingPlansResponse{
		Items: []*offboarding.OffboardingPlan{fixtureOffboardingPlan()},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func TestHRMOffboardingPlanQueriesThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	detail := "query($id: ID!) { offboardingPlan(id:$id) { status termination_effective_date } }"
	list := "query($id: ID!) { offboardingPlans(worker_id:$id) { total_count } }"
	for _, tc := range []struct{ name, grant, query, id, want string }{
		{"detail", "offboarding_plans.read", detail, fixtureWorkforcePlanID, `"status":"in_progress","termination_effective_date":"2026-12-01T00:00:00Z"`},
		{"missing", "offboarding_plans.read", detail, fixturePeriodID, "NOT_FOUND"},
		{"worker Relay ID", "offboarding_plans.read", list, "gid://integrione/hrm-worker/" + hrmFixtureID, `"total_count":1`},
		{"wrong worker", "offboarding_plans.read", list, fixturePeriodID, "UNAVAILABLE"},
		{"termination grant", "workers.terminate", detail, fixtureWorkforcePlanID, "FORBIDDEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{"query": tc.query, "variables": map[string]any{"id": tc.id}}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmOffboardingServer) ApproveOffboardingPlan(_ context.Context, req *common.IDRequest) (*offboarding.OffboardingPlanResponse, error) {
	if req.GetId() != fixtureWorkforcePlanID {
		return nil, status.Error(codes.FailedPrecondition, "offboarding plan is not ready for approval")
	}
	item := fixtureOffboardingPlan()
	item.Status = "approved"
	return &offboarding.OffboardingPlanResponse{Value: item}, nil
}

func (s hrmOffboardingServer) StartOffboardingPlan(ctx context.Context, req *common.IDRequest) (*offboarding.OffboardingPlanResponse, error) {
	return s.GetOffboardingPlanByID(ctx, req)
}

func TestHRMOffboardingPlanCommandsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, field, id, want string
		mfa                          bool
	}{
		{"approve", "approve", "approveOffboardingPlan", fixtureWorkforcePlanID, `"status":"approved"`, true},
		{"approval prerequisites", "approve", "approveOffboardingPlan", fixturePeriodID, "CONFLICT", true},
		{"start", "start", "startOffboardingPlan", fixtureWorkforcePlanID, `"status":"in_progress"`, true},
		{"wrong grant", "approve", "startOffboardingPlan", fixtureWorkforcePlanID, "AUTHORIZATION", true},
		{"without MFA", "start", "startOffboardingPlan", fixtureWorkforcePlanID, "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.offboarding_plans." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!) { " + tc.field + "(id:$id) { was_successful errors { type } offboarding_plan { status } } }",
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

func (hrmOffboardingServer) CreateOffboardingPlan(_ context.Context, req *offboarding.CreateOffboardingPlanRequest) (*offboarding.OffboardingPlanResponse, error) {
	in := req.GetInput()
	if in.GetPlanNumber() == "duplicate" {
		return nil, status.Error(codes.AlreadyExists, "offboarding plan already exists")
	}
	if in.GetWorkerId() != hrmFixtureID || in.GetReasonCode() != "resignation" ||
		in.Status != nil || in.ApprovedAt != nil || in.FinanceClearanceReferenceId != nil ||
		in.GetTerminationEffectiveDate().AsTime().Format(time.DateOnly) != "2026-12-02" {
		return nil, status.Error(codes.Internal, "unexpected draft input")
	}
	item := fixtureOffboardingPlan()
	item.Status = "draft"
	return &offboarding.OffboardingPlanResponse{Value: item}, nil
}

func TestHRMCreateOffboardingPlanThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, grant, number, worker, want string }{
		{"draft", "create", "OFF-01", "gid://integrione/hrm-worker/" + hrmFixtureID, `"status":"draft"`},
		{"duplicate", "create", "duplicate", hrmFixtureID, "CONFLICT"},
		{"invalid worker", "create", "OFF-01", "invalid", "ARGUMENT"},
		{"approve grant", "approve", "OFF-01", hrmFixtureID, "AUTHORIZATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.offboarding_plans." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreateOffboardingPlanInput!) { createOffboardingPlan(input:$input) { was_successful errors { type } offboarding_plan { status } } }",
				"variables": map[string]any{"input": map[string]any{
					"worker_id": tc.worker, "plan_number": tc.number, "reason_code": "resignation",
					"notice_date": "2026-11-27T00:00:00Z", "last_working_date": "2026-12-01T00:00:00Z",
					"termination_effective_date": "2026-12-02T00:00:00Z",
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

func (hrmOffboardingServer) CreateOffboardingTask(_ context.Context, req *offboarding.CreateOffboardingTaskRequest) (*offboarding.OffboardingTaskResponse, error) {
	in := req.GetInput()
	if in.GetOffboardingPlanId() != fixtureWorkforcePlanID {
		return nil, status.Error(codes.FailedPrecondition, "offboarding plan is not a draft")
	}
	if in.Status != nil || in.CompletedAt != nil || in.OwnerConfirmationReferenceId != nil || in.IsMandatory == nil || in.GetIsMandatory() {
		return nil, status.Error(codes.Internal, "unexpected offboarding task input")
	}
	return &offboarding.OffboardingTaskResponse{Value: &offboarding.OffboardingTask{
		Id: hrmFixtureID, OffboardingPlanId: in.GetOffboardingPlanId(), TaskCode: in.GetTaskCode(),
		TaskName: in.GetTaskName(), OwnerManagement: in.GetOwnerManagement(), Status: "pending",
	}}, nil
}

func TestHRMCreateOffboardingTaskThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, grant, plan, want string }{
		{"pending optional task", "create", fixtureWorkforcePlanID, `"status":"pending","is_mandatory":false`},
		{"plan not draft", "create", fixturePeriodID, "CONFLICT"},
		{"invalid plan", "create", "invalid", "ARGUMENT"},
		{"read grant", "read", fixtureWorkforcePlanID, "AUTHORIZATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.offboarding_tasks." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query": "mutation($input: CreateOffboardingTaskInput!) { createOffboardingTask(input:$input) { was_successful errors { type } offboarding_task { status is_mandatory } } }",
				"variables": map[string]any{"input": map[string]any{
					"offboarding_plan_id": tc.plan, "task_code": "HANDOVER", "task_name": "Handover",
					"owner_management": "human_resource", "is_mandatory": false,
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

func (hrmOffboardingServer) FindOffboardingTaskByOffboardingPlanID(_ context.Context, req *common.StringFinderRequest) (*offboarding.OffboardingTasksResponse, error) {
	return &offboarding.OffboardingTasksResponse{
		Items: []*offboarding.OffboardingTask{{Id: hrmFixtureID, OffboardingPlanId: fixtureWorkforcePlanID, Status: "pending"}},
		Page:  &common.PageInfo{Limit: req.GetLimit(), Offset: req.GetOffset(), Total: 1},
	}, nil
}

func TestHRMOffboardingTasksThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, grant, id, want string }{
		{"plan tasks", "offboarding_tasks.read", fixtureWorkforcePlanID, `"nodes":[{"status":"pending"}],"total_count":1`},
		{"unrelated plan", "offboarding_tasks.read", fixturePeriodID, "UNAVAILABLE"},
		{"invalid plan ID", "offboarding_tasks.read", "invalid", "INVALID_ARGUMENT"},
		{"plan only grant", "offboarding_plans.read", fixtureWorkforcePlanID, "FORBIDDEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "query($id: ID!) { offboardingTasks(offboarding_plan_id:$id) { nodes { status } total_count } }",
				"variables": map[string]any{"id": tc.id},
			}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmOffboardingServer) GetOffboardingTaskByID(_ context.Context, req *common.IDRequest) (*offboarding.OffboardingTaskResponse, error) {
	if req.GetId() == fixturePeriodID {
		return nil, status.Error(codes.NotFound, "offboarding task not found")
	}
	return &offboarding.OffboardingTaskResponse{Value: &offboarding.OffboardingTask{
		Id: hrmFixtureID, OffboardingPlanId: fixtureWorkforcePlanID, Status: "pending",
	}}, nil
}

func TestHRMOffboardingTaskThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct{ name, grant, id, want string }{
		{"detail", "offboarding_tasks.read", hrmFixtureID, `"status":"pending"`},
		{"missing", "offboarding_tasks.read", fixturePeriodID, "NOT_FOUND"},
		{"mismatched response", "offboarding_tasks.read", fixtureWorkforcePlanID, "UNAVAILABLE"},
		{"invalid ID", "offboarding_tasks.read", "invalid", "INVALID_ARGUMENT"},
		{"plan only grant", "offboarding_plans.read", hrmFixtureID, "FORBIDDEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, _ := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "query($id: ID!) { offboardingTask(id:$id) { status } }",
				"variables": map[string]any{"id": tc.id},
			}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, nil, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmOffboardingServer) CompleteOffboardingTask(_ context.Context, req *offboarding.CompleteOffboardingTaskRequest) (*offboarding.OffboardingTaskResponse, error) {
	if req.GetOwnerConfirmationReferenceId() != fixturePeriodID {
		return nil, status.Error(codes.Unavailable, "confirmation verifier unavailable")
	}
	return &offboarding.OffboardingTaskResponse{Value: &offboarding.OffboardingTask{
		Id: req.GetId(), OffboardingPlanId: fixtureWorkforcePlanID, Status: "completed",
		OwnerConfirmationReferenceId: &req.OwnerConfirmationReferenceId,
	}}, nil
}

func (hrmOffboardingServer) WaiveOffboardingTask(_ context.Context, req *offboarding.WaiveOffboardingTaskRequest) (*offboarding.OffboardingTaskResponse, error) {
	if req.GetReason() != "fixture reason" {
		return nil, status.Error(codes.InvalidArgument, "reason is required")
	}
	return &offboarding.OffboardingTaskResponse{Value: &offboarding.OffboardingTask{
		Id: req.GetId(), OffboardingPlanId: fixtureWorkforcePlanID, Status: "waived",
	}}, nil
}

func TestHRMOffboardingTaskCommandsThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	complete := `mutation($value: ID!) { completeOffboardingTask(id:"` + hrmFixtureID + `",owner_confirmation_reference_id:$value) { was_successful errors { type } offboarding_task { status } } }`
	waive := `mutation($value: String!) { waiveOffboardingTask(id:"` + hrmFixtureID + `",reason:$value) { was_successful errors { type } offboarding_task { status } } }`
	for _, tc := range []struct {
		name, grant, query, value, want string
		mfa                             bool
	}{
		{"complete", "complete", complete, fixturePeriodID, `"status":"completed"`, true},
		{"invalid confirmation ID", "complete", complete, "invalid", "ARGUMENT", true},
		{"waive", "waive", waive, "fixture reason", `"status":"waived"`, true},
		{"missing reason", "waive", waive, "", "ARGUMENT", true},
		{"wrong grant", "complete", waive, "fixture reason", "AUTHORIZATION", true},
		{"without MFA", "waive", waive, "fixture reason", "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.offboarding_tasks." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{"query": tc.query, "variables": map[string]any{"value": tc.value}}
			headers := map[string]string{"Origin": "https://erp.localhost:8443", browserauth.CSRFHeaderName: csrf}
			cookies := []*http.Cookie{{Name: config.SessionCookieName, Value: token}}
			body := readBody(t, doGraphQLPOST(t, srv, request, headers, cookies))
			if !strings.Contains(body, tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, body)
			}
		})
	}
}

func (hrmOffboardingServer) CompleteOffboardingPlan(_ context.Context, req *offboarding.CompleteOffboardingPlanRequest) (*offboarding.OffboardingPlanResponse, error) {
	if req.GetId() != fixtureWorkforcePlanID || req.GetFinalPayrollRunId() != fixturePeriodID {
		return nil, status.Error(codes.FailedPrecondition, "final payroll run does not match")
	}
	if req.GetFinanceClearanceReferenceId() != hrmFixtureID {
		return nil, status.Error(codes.Unavailable, "clearance verifier unavailable")
	}
	item := fixtureOffboardingPlan()
	item.Status = "completed"
	item.CompletedAt = item.TerminationEffectiveDate
	item.FinalPayrollRunId = &req.FinalPayrollRunId
	item.FinanceClearanceReferenceId = &req.FinanceClearanceReferenceId
	return &offboarding.OffboardingPlanResponse{Value: item}, nil
}

func TestHRMCompleteOffboardingPlanThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, runID, clearanceID, want string
		mfa                                   bool
	}{
		{"completed", "complete", fixturePeriodID, hrmFixtureID, `"status":"completed"`, true},
		{"payroll conflict", "complete", hrmFixtureID, hrmFixtureID, "CONFLICT", true},
		{"verifier missing", "complete", fixturePeriodID, fixtureWorkforcePlanID, "UNAVAILABLE", true},
		{"invalid clearance", "complete", fixturePeriodID, "invalid", "ARGUMENT", true},
		{"wrong grant", "start", fixturePeriodID, hrmFixtureID, "AUTHORIZATION", true},
		{"no MFA", "complete", fixturePeriodID, hrmFixtureID, "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.offboarding_plans." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $run: ID!, $clearance: ID!) { completeOffboardingPlan(id:$id,final_payroll_run_id:$run,finance_clearance_reference_id:$clearance) { was_successful errors { type } offboarding_plan { status } } }",
				"variables": map[string]any{"id": fixtureWorkforcePlanID, "run": tc.runID, "clearance": tc.clearanceID},
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

func (hrmOffboardingServer) CancelOffboardingPlan(_ context.Context, req *offboarding.CancelOffboardingPlanRequest) (*offboarding.OffboardingPlanResponse, error) {
	if req.GetId() != fixtureWorkforcePlanID {
		return nil, status.Error(codes.FailedPrecondition, "plan has started")
	}
	reason := strings.TrimSpace(req.GetReason())
	if reason == "" {
		return nil, status.Error(codes.InvalidArgument, "reason is required")
	}
	item := fixtureOffboardingPlan()
	item.Status = "cancelled"
	item.CancelledAt = item.LastWorkingDate
	item.CancellationReason = &reason
	return &offboarding.OffboardingPlanResponse{Value: item}, nil
}

func TestHRMCancelOffboardingPlanThroughGateway(t *testing.T) {
	conn := newHRMPayrollFixture(t)
	for _, tc := range []struct {
		name, grant, id, reason, want string
		mfa                           bool
	}{
		{"cancel", "cancel", fixtureWorkforcePlanID, " fixture reason ", `"status":"cancelled"`, true},
		{"started plan", "cancel", fixturePeriodID, "fixture reason", "CONFLICT", true},
		{"blank reason", "cancel", fixtureWorkforcePlanID, " ", "ARGUMENT", true},
		{"wrong grant", "start", fixtureWorkforcePlanID, "fixture reason", "AUTHORIZATION", true},
		{"without MFA", "cancel", fixtureWorkforcePlanID, "fixture reason", "AUTHORIZATION", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, store, token, csrf := buildTestGatewayServerWithHRM(t, nil, false, conn)
			session := store.sessions[authsession.HashOpaqueToken(token)]
			session.Token.MFAAuthenticated = tc.mfa
			session.Snapshot.Permissions = append(session.Snapshot.Permissions, authz.PermissionGrant{
				Code: "human_resource.offboarding_plans." + tc.grant, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
			})
			request := map[string]any{
				"query":     "mutation($id: ID!, $reason: String!) { cancelOffboardingPlan(id:$id,reason:$reason) { was_successful errors { type } offboarding_plan { status } } }",
				"variables": map[string]any{"id": tc.id, "reason": tc.reason},
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
