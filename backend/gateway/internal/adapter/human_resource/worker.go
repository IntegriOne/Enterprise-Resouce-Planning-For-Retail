package human_resource

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	accessrequest "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/accessrequest/v1"
	analytics "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/analytics/v1"
	benefits "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	career "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/career/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	compensation "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/compensation/v1"
	competency "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	employeerelations "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/employeerelations/v1"
	engagement "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/engagement/v1"
	hrservice "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/hrservice/v1"
	learning "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	offboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	onboarding "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/onboarding/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	performance "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	policy "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/policy/v1"
	promotion "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/promotion/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	succession "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/succession/v1"
	talent "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const workerGIDPrefix = "gid://integrione/hrm-worker/"
const workerReadPermission = "human_resource.workers.read"

type Adapter struct {
	iam               source_connections.IAMClientSet
	accessRequests    accessrequest.AccessRequestServiceClient
	analytics         analytics.AnalyticsServiceClient
	auth              *middleware.AuthManager
	authorizer        *middleware.Authorizer
	benefits          benefits.BenefitsServiceClient
	career            career.CareerServiceClient
	compensation      compensation.CompensationServiceClient
	competency        competency.CompetencyServiceClient
	employeeRelations employeerelations.EmployeeRelationsServiceClient
	engagement        engagement.EngagementServiceClient
	hrService         hrservice.HRServiceClient
	learning          learning.LearningServiceClient
	leave             leave.LeaveServiceClient
	offboarding       offboarding.OffboardingServiceClient
	onboarding        onboarding.OnboardingServiceClient
	payroll           payroll.PayrollServiceClient
	performance       performance.PerformanceServiceClient
	personnel         personnel.PersonnelServiceClient
	policy            policy.PolicyServiceClient
	promotion         promotion.PromotionServiceClient
	recruitment       recruitment.RecruitmentServiceClient
	succession        succession.SuccessionServiceClient
	talent            talent.TalentServiceClient
	workforce         workforce.WorkforcePlanningServiceClient
}

type Option func(*Adapter)

// WithEmployeeRequestDelivery memasang koneksi penerimaan pengajuan IAM.
func WithEmployeeRequestDelivery(clients source_connections.IAMClientSet) Option {
	return func(adapter *Adapter) { adapter.iam = clients }
}

func NewAdapter(auth *middleware.AuthManager, authorizer *middleware.Authorizer, conn *grpc.ClientConn, options ...Option) *Adapter {
	a := &Adapter{auth: auth, authorizer: authorizer}
	if conn != nil {
		a.accessRequests = accessrequest.NewAccessRequestServiceClient(conn)
		a.analytics = analytics.NewAnalyticsServiceClient(conn)
		a.benefits = benefits.NewBenefitsServiceClient(conn)
		a.career = career.NewCareerServiceClient(conn)
		a.compensation = compensation.NewCompensationServiceClient(conn)
		a.competency = competency.NewCompetencyServiceClient(conn)
		a.employeeRelations = employeerelations.NewEmployeeRelationsServiceClient(conn)
		a.engagement = engagement.NewEngagementServiceClient(conn)
		a.hrService = hrservice.NewHRServiceClient(conn)
		a.learning = learning.NewLearningServiceClient(conn)
		a.leave = leave.NewLeaveServiceClient(conn)
		a.offboarding = offboarding.NewOffboardingServiceClient(conn)
		a.onboarding = onboarding.NewOnboardingServiceClient(conn)
		a.payroll = payroll.NewPayrollServiceClient(conn)
		a.performance = performance.NewPerformanceServiceClient(conn)
		a.personnel = personnel.NewPersonnelServiceClient(conn)
		a.policy = policy.NewPolicyServiceClient(conn)
		a.promotion = promotion.NewPromotionServiceClient(conn)
		a.recruitment = recruitment.NewRecruitmentServiceClient(conn)
		a.succession = succession.NewSuccessionServiceClient(conn)
		a.talent = talent.NewTalentServiceClient(conn)
		a.workforce = workforce.NewWorkforcePlanningServiceClient(conn)
	}
	for _, option := range options {
		option(a)
	}
	return a
}

func IsWorkerNodeID(raw string) bool {
	return strings.HasPrefix(raw, workerGIDPrefix)
}

func (a *Adapter) authorizeRead(ctx context.Context, permission string) (context.Context, error) {
	if a == nil || a.auth == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	sess, err := a.auth.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.Snapshot == nil || sess.Caller.UserKind != "employee" {
		return nil, middleware.ErrForbidden("employee session required", nil)
	}
	ctx = middleware.WithResolvedSession(ctx, sess)
	verified, err := a.authorizer.RequirePermission(ctx, permission, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	if gwErr, ok := middleware.AsGatewayError(err); ok && gwErr.Code == middleware.CodeForbidden {
		// HRM menentukan kecocokan data dengan badan hukum saat membaca record.
		for _, grant := range sess.Snapshot.Permissions {
			if grant.Code != permission || grant.Scope.Type != "legal_entity" {
				continue
			}
			if _, parseErr := uuid.Parse(grant.Scope.ID); parseErr != nil {
				continue
			}
			candidate, scopeErr := a.authorizer.RequirePermission(ctx, permission, grant.Scope, false)
			if scopeErr == nil {
				verified, err = candidate, nil
				break
			}
			if scopeFailure, ok := middleware.AsGatewayError(scopeErr); !ok || scopeFailure.Code != middleware.CodeForbidden {
				return nil, scopeErr
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return middleware.WithResolvedSession(ctx, verified), nil
}

// Worker memetakan detail pekerja dari HRM setelah otorisasi Gateway.
func (a *Adapter) Worker(ctx context.Context, rawID string) (*model.Worker, error) {
	rpcCtx, err := a.authorizeRead(ctx, workerReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawID, workerGIDPrefix))
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.GetWorkerByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "worker")
	}
	worker := response.GetItem()
	if worker == nil || worker.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource response", nil)
	}
	return toModelWorker(worker), nil
}

func toModelWorker(worker *personnel.Worker) *model.Worker {
	var partyID *string
	if id := worker.GetPersonPartyReferenceId(); id != "" {
		partyID = &id
	}
	return &model.Worker{
		PersonPartyReferenceID: partyID,
		ID:                     workerGIDPrefix + worker.GetId(), RawID: worker.GetId(), WorkerNumber: worker.GetWorkerNumber(),
		WorkerCategory: worker.GetWorkerCategory(), Status: worker.GetStatus(),
		PrimaryLegalEntityReferenceID: worker.GetPrimaryLegalEntityReferenceId(),
		HireDate:                      worker.HireDate, ServiceStartDate: worker.ServiceStartDate,
	}
}

func mapReadError(err error, resource string) error {
	code := status.Code(err)
	switch code {
	case codes.NotFound:
		return middleware.ErrNotFound(resource+" not found", err)
	case codes.PermissionDenied:
		return middleware.ErrForbidden("forbidden", err)
	case codes.InvalidArgument:
		return middleware.ErrInvalidArgument("invalid "+resource+" request", err)
	default:
		// Pesan transport dapat memuat alamat privat atau detail upstream.
		return middleware.ErrUnavailable("human resource service unavailable", fmt.Errorf("human resource RPC status: %s", code))
	}
}

// Workers mengambil halaman pekerja dan total yang sudah dibatasi scope oleh HRM.
func (a *Adapter) Workers(ctx context.Context, limit, offset int, workerNumber, legalEntityID *string) (*Page[model.Worker], error) {
	rpcCtx, err := a.authorizeRead(ctx, workerReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	if workerNumber != nil && legalEntityID != nil {
		return nil, middleware.ErrInvalidArgument("use only one worker filter", nil)
	}
	var response *personnel.WorkersResponse
	var legalEntityFilter string
	switch {
	case legalEntityID != nil:
		id, parseErr := parseNodeID(*legalEntityID, "legal-entity")
		if parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid legal entity ID", nil)
		}
		legalEntityFilter = id.String()
		response, err = a.personnel.FindWorkerByPrimaryLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{
			Value: &legalEntityFilter, Limit: int32(limit), Offset: int32(offset),
		})
	case workerNumber != nil:
		number := *workerNumber
		if !utf8.ValidString(number) || strings.ContainsRune(number, 0) || strings.TrimSpace(number) == "" || utf8.RuneCountInString(number) > 100 {
			return nil, middleware.ErrInvalidArgument("invalid worker number", nil)
		}
		response, err = a.personnel.FindWorkerByWorkerNumber(rpcCtx, &common.StringFinderRequest{
			Value: workerNumber, Limit: int32(limit), Offset: int32(offset),
		})
	default:
		response, err = a.personnel.GetAllWorkers(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource worker page", nil)
	}
	result := &Page[model.Worker]{
		Items: make([]*model.Worker, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, worker := range items {
		if parsed, err := uuid.Parse(worker.GetId()); err != nil || parsed == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource worker page", nil)
		}
		if (workerNumber != nil && worker.GetWorkerNumber() != *workerNumber) ||
			(legalEntityID != nil && worker.GetPrimaryLegalEntityReferenceId() != legalEntityFilter) {
			return nil, middleware.ErrUnavailable("invalid human resource worker match", nil)
		}
		result.Items = append(result.Items, toModelWorker(worker))
	}
	return result, nil
}
