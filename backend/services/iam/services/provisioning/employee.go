package provisioning

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// WorkerAccountLinkVerifier memeriksa tautan berdasarkan catatan pekerja pada HRM.
type WorkerAccountLinkVerifier interface {
	VerifyWorkerAccountLink(context.Context, string, string, string) error
}

type Service struct {
	workerLink WorkerAccountLinkVerifier
	requests   domainrepo.EmployeeProvisioningRepository
	users      domainrepo.UserRepository
	roles      domainrepo.RoleRepository
	verifier   domainservice.ScopeOwnerVerifier
	clock      domainservice.Clock
}

func NewService(requests domainrepo.EmployeeProvisioningRepository, users domainrepo.UserRepository, roles domainrepo.RoleRepository, verifier domainservice.ScopeOwnerVerifier, clock domainservice.Clock) *Service {
	return &Service{requests: requests, users: users, roles: roles, verifier: verifier, clock: clock}
}

// ConfigureWorkerAccountLinkVerifier mengikat verifier saat startup sebelum server menerima RPC.
func (s *Service) ConfigureWorkerAccountLinkVerifier(verifier WorkerAccountLinkVerifier) {
	s.workerLink = verifier
}

// Receive menyimpan pengajuan HRM setelah memeriksa kembali izin dan target scope.
func (s *Service) Receive(ctx context.Context, hrmRequestID, workerID, legalEntityID, workerNumber, catalogID string, requestedAt time.Time) (*entity.EmployeeProvisioningRequest, error) {
	hrmRequestID = strings.ToLower(strings.TrimSpace(hrmRequestID))
	workerID = strings.ToLower(strings.TrimSpace(workerID))
	legalEntityID = strings.ToLower(strings.TrimSpace(legalEntityID))
	workerNumber = strings.TrimSpace(workerNumber)
	if !entity.IsValidUUID(hrmRequestID) || !entity.IsValidUUID(workerID) ||
		!entity.IsValidUUID(legalEntityID) || requestedAt.IsZero() || workerNumber == "" || utf8.RuneCountInString(workerNumber) > 100 {
		return nil, iamerrors.ErrInvalidInput
	}
	snapshot, catalog, err := s.authorizedEmployeeCatalog(ctx, legalEntityID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(catalog.ID, catalogID) {
		return nil, iamerrors.ErrConflict
	}

	return s.requests.InsertReceipt(ctx, entity.EmployeeProvisioningRequest{
		HRMRequestID: hrmRequestID, WorkerID: workerID, WorkerNumber: workerNumber, LegalEntityID: legalEntityID,
		RequestedByPrincipalID: snapshot.UserID, RequestedAt: requestedAt, CatalogRoleID: catalog.ID,
	})
}

// ResolveCatalog mengembalikan referensi kebijakan self-service yang dikelola IAM.
func (s *Service) ResolveCatalog(ctx context.Context, legalEntityID string) (string, error) {
	_, catalog, err := s.authorizedEmployeeCatalog(ctx, legalEntityID)
	if err != nil {
		return "", err
	}
	return catalog.ID, nil
}

func (s *Service) authorizedEmployeeCatalog(ctx context.Context, legalEntityID string) (*entity.AccessSnapshot, *entity.Role, error) {
	snapshot, err := s.employeeSnapshot(ctx)
	if err != nil {
		return nil, nil, err
	}
	scope, err := entity.NewScope(entity.ScopeTypeLegalEntity, &legalEntityID)
	if err != nil {
		return nil, nil, err
	}
	if _, allowed := snapshot.HasPermission("human_resource.access_requests.create", scope); !allowed {
		return nil, nil, iamerrors.ErrForbidden
	}
	if err := domainservice.VerifyScope(ctx, s.verifier, scope); err != nil {
		return nil, nil, err
	}
	if s.roles == nil {
		return nil, nil, iamerrors.ErrUnavailable
	}
	role, err := s.roles.SelectByCode(ctx, entity.BaselineSelfRoleCodeForUserKind(entity.UserKindEmployee))
	if errors.Is(err, iamerrors.ErrRoleNotFound) {
		return nil, nil, iamerrors.ErrUnavailable
	}
	if err != nil {
		return nil, nil, err
	}
	if role == nil || !role.IsSystem || !role.IsActive || role.ManagementDomain != "iam" {
		return nil, nil, iamerrors.ErrUnavailable
	}
	return snapshot, role, nil
}

// ListOpen mengembalikan antrean bagi operator provisioning IAM dengan izin global.
func (s *Service) ListOpen(ctx context.Context, limit, offset int) ([]entity.EmployeeProvisioningRequest, int64, error) {
	if limit < 0 || limit > 100 || offset < 0 {
		return nil, 0, iamerrors.ErrInvalidInput
	}
	if err := s.authorizeQueue(ctx); err != nil {
		return nil, 0, err
	}
	return s.requests.SelectOpen(ctx, limit, offset)
}

// Get mengembalikan detail pengajuan bagi operator provisioning IAM.
func (s *Service) Get(ctx context.Context, id string) (*entity.EmployeeProvisioningRequest, error) {
	if err := entity.ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := s.authorizeQueue(ctx); err != nil {
		return nil, err
	}
	return s.requests.SelectByID(ctx, id)
}

// GetStatusForHRM membuka status IAM kepada operator HRM sesuai cakupan pengajuan.
func (s *Service) GetStatusForHRM(ctx context.Context, hrmRequestID, legalEntityID string) (*entity.EmployeeProvisioningRequest, error) {
	if err := entity.ValidateUUID("hrm_request_id", hrmRequestID); err != nil {
		return nil, err
	}
	scope, err := entity.NewScope(entity.ScopeTypeLegalEntity, &legalEntityID)
	if err != nil {
		return nil, err
	}
	snapshot, err := s.employeeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	if _, allowed := snapshot.HasPermission("human_resource.access_requests.read", scope); !allowed {
		return nil, iamerrors.ErrForbidden
	}
	return s.requests.SelectByHRMRequest(ctx, hrmRequestID, legalEntityID)
}

func (s *Service) authorizeQueue(ctx context.Context) error {
	snapshot, err := s.employeeSnapshot(ctx)
	if err != nil {
		return err
	}
	if _, allowed := snapshot.HasPermission("iam.users.write", entity.GlobalScope()); !allowed {
		return iamerrors.ErrForbidden
	}
	return nil
}

func (s *Service) employeeSnapshot(ctx context.Context) (*entity.AccessSnapshot, error) {
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated || caller.SessionID == "" {
		return nil, iamerrors.ErrUnauthorized
	}
	if caller.UserKind != string(entity.UserKindEmployee) {
		return nil, iamerrors.ErrForbidden
	}
	if s == nil || s.requests == nil || s.users == nil || s.clock == nil {
		return nil, iamerrors.ErrUnavailable
	}
	snapshot, err := s.users.SelectAccessSnapshotByUserID(ctx, caller.PrincipalID, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if snapshot == nil || snapshot.UserKind != entity.UserKindEmployee ||
		snapshot.AuthEpoch != caller.AuthEpoch || snapshot.PermRevision != caller.PermRevision {
		return nil, iamerrors.ErrForbidden
	}
	return snapshot, nil
}

// Complete memverifikasi tautan langsung ke HRM sebelum menandai pengajuan tertaut.
func (s *Service) Complete(ctx context.Context, id, accountID string) error {
	if !entity.IsValidUUID(id) || !entity.IsValidUUID(accountID) {
		return iamerrors.ErrInvalidInput
	}
	if err := s.authorizeQueue(ctx); err != nil {
		return err
	}
	if s.workerLink == nil {
		return iamerrors.ErrUnavailable
	}
	caller, _ := authcontext.FromContext(ctx)
	request, err := s.requests.SelectByID(ctx, id)
	if err != nil {
		return err
	}
	if request == nil || request.ProcessedByUserID == nil || request.AccountUserID == nil ||
		!strings.EqualFold(*request.ProcessedByUserID, caller.PrincipalID) || !strings.EqualFold(*request.AccountUserID, accountID) ||
		(request.Status != entity.EmployeeProvisioningProcessing && request.Status != entity.EmployeeProvisioningLinked) {
		return iamerrors.ErrConflict
	}
	if err := s.workerLink.VerifyWorkerAccountLink(ctx, request.WorkerID, accountID, request.LegalEntityID); err != nil {
		return err
	}
	return s.requests.UpdateLinked(ctx, id, caller.PrincipalID, accountID, s.clock.Now())
}
