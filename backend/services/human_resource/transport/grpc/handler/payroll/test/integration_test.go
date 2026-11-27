//go:build integration

package payroll_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/contract"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/inbox"
	documentrecords "github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/integration/document_records"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/payroll"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type payslipVerifierStub struct {
	expected          service.PayslipCheck
	publication       *payslipPublicationStub
	missingContent    bool
	reissueReadError  error
	otherEmployment   *service.PayslipCheck
	published         bool
	publishedAt       time.Time
	omitPublicationID bool
	verifiedPeriodID  string
	pdf               service.VerifiedPayslipPDF
	pdfCalls          int
	beforePDFRead     func(context.Context) error
	beforeListReturn  func(context.Context) error
}

func (v *payslipVerifierStub) FindPublicationForReissue(ctx context.Context, check service.PayslipPublicationCheck) (string, error) {
	check.PublicationID = v.publication.publicationID
	if _, err := v.publication.VerifyPayrollPublication(ctx, check); err != nil {
		return "", err
	}
	return check.PublicationID, nil
}

func (v *payslipVerifierStub) ReadPayslipForReissue(ctx context.Context, check service.PayslipCheck, _ string, limit int) (service.VerifiedPayslipPDF, error) {
	if v.reissueReadError != nil {
		return service.VerifiedPayslipPDF{}, v.reissueReadError
	}

	if v.missingContent && v.publication.restoreCalls == 0 {
		return service.VerifiedPayslipPDF{}, service.ErrPayslipContentMissing
	}
	return v.ReadVerifiedPayslipPDF(ctx, check, limit)
}

func (v *payslipVerifierStub) VerifyPublishedPayslip(_ context.Context, check service.PayslipCheck) (service.PublishedPayslip, error) {
	if !v.published || (check != v.expected && (v.otherEmployment == nil || check != *v.otherEmployment)) {
		return service.PublishedPayslip{}, hrerrors.ErrReferenceConflict
	}
	publicationID := check.PublicationID
	if v.omitPublicationID {
		publicationID = ""
	}
	return service.PublishedPayslip{
		PayrollPeriodID: v.verifiedPeriodID, ResultID: check.ResultID, PaymentStatus: "scheduled", PublishedAt: v.publishedAt,
		DocumentVersionID: check.DocumentVersionID, PublicationID: publicationID,
		ScheduledPayDate: time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC),
	}, nil
}

func (v *payslipVerifierStub) ReadVerifiedPayslipPDF(ctx context.Context, check service.PayslipCheck, limit int) (service.VerifiedPayslipPDF, error) {
	v.pdfCalls++
	if check != v.expected || limit != contract.MaxPayslipPDFBytes {
		return service.VerifiedPayslipPDF{}, hrerrors.ErrReferenceConflict
	}
	if v.pdf.Content == nil {
		return service.VerifiedPayslipPDF{}, hrerrors.ErrUnavailable
	}
	if v.beforePDFRead != nil {
		if err := v.beforePDFRead(ctx); err != nil {
			return service.VerifiedPayslipPDF{}, err
		}
	}
	return v.pdf, nil
}

func (v *payslipVerifierStub) ListPublishedPayslips(ctx context.Context, check service.PayslipListCheck) (operation.Page[service.PublishedPayslip], error) {
	if check.PrincipalID != v.expected.PrincipalID || !slices.Contains(check.WorkerIDs, v.expected.WorkerID) {
		return operation.Page[service.PublishedPayslip]{}, hrerrors.ErrReferenceConflict
	}
	records := []service.PayslipCheck{v.expected}
	if v.otherEmployment != nil {
		if !slices.Contains(check.WorkerIDs, v.otherEmployment.WorkerID) || len(check.WorkerIDs) != 2 {
			return operation.Page[service.PublishedPayslip]{}, hrerrors.ErrReferenceConflict
		}
		records = append(records, *v.otherEmployment)
	}
	page := operation.Page[service.PublishedPayslip]{Total: int64(len(records))}
	for _, record := range records[min(check.Offset, len(records)):min(check.Offset+check.Limit, len(records))] {
		slip, err := v.VerifyPublishedPayslip(ctx, record)
		if err != nil {
			return operation.Page[service.PublishedPayslip]{}, err
		}
		page.Items = append(page.Items, slip)
	}
	if v.beforeListReturn != nil {
		if err := v.beforeListReturn(ctx); err != nil {
			return operation.Page[service.PublishedPayslip]{}, err
		}
	}
	return page, nil
}

type payslipPublicationStub struct {
	publicationID, runID, journalID, legalEntityID, resultID, workerID, documentID string
	inputSnapshotID                                                                string
	inputDigest                                                                    []byte
	storeCalls                                                                     atomic.Int32
	restoreCalls                                                                   int
	storeEntered                                                                   chan struct{}
	storeRelease                                                                   chan struct{}
	processor                                                                      *documentrecords.PayslipProcessor
	lostResponseForResult                                                          string
	savedDocument                                                                  *documentrecords.PrivatePayslip
	publishedAt                                                                    time.Time
	beforeVerify                                                                   func()
}

func (s *payslipPublicationStub) VerifyPayrollPublication(_ context.Context, check service.PayslipPublicationCheck) (time.Time, error) {
	if check.PublicationID != s.publicationID || check.RunID != s.runID || check.JournalID != s.journalID ||
		check.LegalEntityID != s.legalEntityID || check.InputSnapshotID != s.inputSnapshotID ||
		!bytes.Equal(check.InputDigest, s.inputDigest) || !bytes.Equal(check.OutputDigest, bytes.Repeat([]byte{0x44}, 32)) {
		return time.Time{}, hrerrors.ErrReferenceConflict
	}
	if s.beforeVerify != nil {
		s.beforeVerify()
	}

	return s.publishedAt, nil
}

func (s *payslipPublicationStub) StorePayslipPDF(ctx context.Context, input service.PayslipDocumentInput) (string, error) {
	attempt := s.storeCalls.Add(1)
	if attempt == 1 && s.storeEntered != nil {
		close(s.storeEntered)
		select {
		case <-s.storeRelease:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if !input.PublishedAt.Equal(s.publishedAt) || input.PublicationID != s.publicationID || input.Result.ID != s.resultID || input.Result.WorkerID != s.workerID ||
		input.IdempotencyKey != "payslip:"+s.publicationID+":"+s.resultID || len(input.Lines) != 1 ||
		input.Period.ID == "" {
		return "", hrerrors.ErrReferenceConflict
	}
	if s.processor != nil {
		return s.processor.StorePayslipPDF(ctx, input)
	}
	return s.documentID, nil
}

func (s *payslipPublicationStub) StorePrivatePayslip(_ context.Context, document documentrecords.PrivatePayslip) (string, error) {
	if document.ResultID != s.resultID || document.WorkerID != s.workerID ||
		document.PublicationID != s.publicationID || document.LegalEntityID != s.legalEntityID ||
		document.IdempotencyKey != "payslip:"+s.publicationID+":"+s.resultID ||
		!bytes.HasPrefix(document.PDF, []byte("%PDF-")) || document.Digest != sha256.Sum256(document.PDF) {
		return "", hrerrors.ErrReferenceConflict
	}
	if document.ResultID == s.lostResponseForResult {
		if s.savedDocument == nil {
			s.savedDocument = &document
			return "", hrerrors.ErrUnavailable
		}
		if document.IdempotencyKey != s.savedDocument.IdempotencyKey ||
			document.Digest != s.savedDocument.Digest || !bytes.Equal(document.PDF, s.savedDocument.PDF) {
			return "", hrerrors.ErrReferenceConflict
		}
	}
	return s.documentID, nil
}

func (s *payslipPublicationStub) RestorePrivatePayslip(_ context.Context, versionID string, document documentrecords.PrivatePayslip) error {
	if versionID != s.documentID || s.savedDocument == nil || document.Digest != s.savedDocument.Digest ||
		!bytes.Equal(document.PDF, s.savedDocument.PDF) {
		return hrerrors.ErrReferenceConflict
	}
	s.restoreCalls++
	return nil
}

type payrollTestPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *payrollTestPolicy) setPrincipal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.principalID = id
}

func (p *payrollTestPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *payrollTestPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*payrollTestPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type payrollSettlementVerifierFunc func(context.Context, service.PayrollSettlementCheck) (time.Time, error)

func (f payrollSettlementVerifierFunc) VerifyPayrollSettlement(ctx context.Context, check service.PayrollSettlementCheck) (time.Time, error) {
	return f(ctx, check)
}

type payrollPaymentVerifierFunc func(context.Context, service.PayrollPaymentCheck) error

func (f payrollPaymentVerifierFunc) VerifyPayrollPayment(ctx context.Context, check service.PayrollPaymentCheck) error {
	return f(ctx, check)
}

type payrollPostingVerifierFunc func(context.Context, service.PayrollPostingCheck) (time.Time, error)

func (f payrollPostingVerifierFunc) VerifyPayrollPosting(ctx context.Context, check service.PayrollPostingCheck) (time.Time, error) {
	return f(ctx, check)
}

type payrollFoundationStub struct{ rejectWindow atomic.Bool }

func (s *payrollFoundationStub) VerifyLegalEntityWindow(_ context.Context, legalID string, start, end time.Time) error {
	if s.rejectWindow.Load() || entity.ValidateUUID("legal_entity_id", legalID) != nil ||
		start.Format(time.DateOnly) != "2026-11-01" || end.Format(time.DateOnly) != "2026-11-30" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (payrollFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (payrollFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (payrollFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (payrollFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (payrollFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (payrollFoundationStub) VerifyParty(context.Context, string, bool) error      { return nil }

const payrollInputFixture = "frozen payroll input fixture"
const replacementPayrollInputFixture = "corrected payroll input fixture"
const payrollRuleFixture = "approved payroll rule artifact fixture"

type payrollCalculatorStub struct {
	snapshotScope         service.CalculationSnapshot
	replacementSnapshotID string
	result                entity.PayrollResult
	line                  entity.PayrollResultLine
	mismatched            atomic.Bool
	wrongRuleArtifact     atomic.Bool
	changedInputBytes     atomic.Bool
	changedRuleBytes      atomic.Bool
	precisionMismatch     atomic.Bool
	prematurePublication  atomic.Bool
	calls                 atomic.Int32
}

func (s *payrollCalculatorStub) ReadCalculationSnapshot(_ context.Context, snapshotID string) (service.CalculationSnapshot, error) {
	result := s.snapshotScope
	result.ID = snapshotID
	switch {
	case s.changedInputBytes.Load():
		result.Content = []byte("changed input fixture")
	case snapshotID == s.replacementSnapshotID:
		result.Content = []byte(replacementPayrollInputFixture)
	default:
		result.Content = []byte(payrollInputFixture)
	}
	return result, nil
}

func (s *payrollCalculatorStub) ReadRuleArtifactVersion(context.Context, string) ([]byte, error) {
	if s.changedRuleBytes.Load() {
		return []byte("changed rule fixture"), nil
	}
	return []byte(payrollRuleFixture), nil
}

func (s *payrollCalculatorStub) CalculateVerified(_ context.Context, request service.CalculationRequest, snapshot, artifact []byte) (service.CalculationResult, error) {
	s.calls.Add(1)
	expectedSnapshot := payrollInputFixture
	if request.InputSnapshotID == s.replacementSnapshotID {
		expectedSnapshot = replacementPayrollInputFixture
	}
	if string(snapshot) != expectedSnapshot || string(artifact) != payrollRuleFixture {
		return service.CalculationResult{}, fmt.Errorf("calculator received unexpected input or rules")
	}
	if request.CurrencyScale != 2 {
		return service.CalculationResult{}, fmt.Errorf("unexpected payroll currency scale: %d", request.CurrencyScale)
	}
	if request.PeriodStart.Format(time.DateOnly) != "2026-11-01" ||
		request.PeriodEnd.Format(time.DateOnly) != "2026-11-30" ||
		request.PayDate.Format(time.DateOnly) != "2026-11-28" {
		return service.CalculationResult{}, fmt.Errorf("unexpected payroll dates: %v, %v, %v", request.PeriodStart, request.PeriodEnd, request.PayDate)
	}
	result := s.result
	result.PayrollRunID, result.PayrollGroupID, result.CurrencyCode = request.RunID, request.GroupID, request.CurrencyCode
	output := service.CalculationResult{
		RunID: request.RunID, RuleSetID: request.RuleSetID, InputSnapshotID: request.InputSnapshotID,
		InputDigest: request.InputDigest, RuleArtifactDigest: request.RuleArtifactDigest, OutputDigest: bytes.Repeat([]byte{0x56}, 32),
		Results: []entity.PayrollResult{result}, Lines: []entity.PayrollResultLine{s.line},
	}
	if s.prematurePublication.Load() {
		output.Results[0].PayslipPublicationReferenceID = &request.RunID
	}
	if s.precisionMismatch.Load() {
		output.Results[0].GrossAmount = "100.001"
		output.Results[0].EmployerCostAmount = "100.001"
		output.Results[0].NetPayAmount = "100.001"
		output.Lines[0].Amount = "100.001"
	}
	if s.mismatched.Load() {
		output.InputDigest = bytes.Repeat([]byte{0x57}, 32)
	}
	if s.wrongRuleArtifact.Load() {
		output.RuleArtifactDigest = bytes.Repeat([]byte{0x58}, 32)
	}
	return output, nil
}

func TestPayrollRPCAndApprovalWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.PayComponentValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	foundation := &payrollFoundationStub{}
	payrollVals, err := security.PayrollValidators(repos, foundation)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range payrollVals {
		validators[k] = v
	}

	starterID := "00000000-0000-4000-8000-000000000021"
	approverID := "00000000-0000-4000-8000-000000000022"
	policy := &payrollTestPolicy{principalID: starterID}
	publication := &payslipPublicationStub{}
	if fontPath := os.Getenv("HRM_PAYSLIP_TEST_FONT"); fontPath != "" {
		font, err := os.ReadFile(fontPath)
		if err != nil {
			t.Fatal(err)
		}
		publication.processor, err = documentrecords.NewPayslipProcessor(publication, publication, font)
		if err != nil {
			t.Fatal(err)
		}
	}

	payslipVerifier := &payslipVerifierStub{publication: publication}
	calculator := &payrollCalculatorStub{}
	verifiedCalculator, err := service.NewVerifiedCalculator(calculator, calculator, calculator)
	if err != nil {
		t.Fatal(err)
	}
	inputDigest := sha256.Sum256([]byte(payrollInputFixture))
	ruleDigest := sha256.Sum256([]byte(payrollRuleFixture))
	var verifyPosting payrollPostingVerifierFunc
	var verifyPayment payrollPaymentVerifierFunc
	var verifySettlement payrollSettlementVerifierFunc
	svc, err := service.NewService(service.Dependencies{
		PayrollSettlement: payrollSettlementVerifierFunc(func(ctx context.Context, check service.PayrollSettlementCheck) (time.Time, error) {
			if verifySettlement == nil {
				return time.Time{}, hrerrors.ErrUnavailable
			}
			return verifySettlement(ctx, check)
		}),
		PayrollPayment: payrollPaymentVerifierFunc(func(ctx context.Context, check service.PayrollPaymentCheck) error {
			if verifyPayment == nil {
				return hrerrors.ErrUnavailable
			}
			return verifyPayment(ctx, check)
		}),
		PayrollPosting: payrollPostingVerifierFunc(func(ctx context.Context, check service.PayrollPostingCheck) (time.Time, error) {
			if verifyPosting == nil {
				return time.Time{}, hrerrors.ErrUnavailable
			}
			return verifyPosting(ctx, check)
		}),
		Calculator:         verifiedCalculator,
		PayslipPublication: publication,
		PayslipReissue:     payslipVerifier,
		Workers:            repos.Worker, PayslipVerifier: payslipVerifier,
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		PayrollProfile:    repos.PayrollProfile,
		PayComponent:      repos.PayComponent,
		PayrollPeriod:     repos.PayrollPeriod,
		PayrollRun:        repos.PayrollRun,
		PayrollResult:     repos.PayrollResult,
		PayrollResultLine: repos.PayrollResultLine,
		PayrollAdjustment: repos.PayrollAdjustment,
		PayrollGroup:      repos.PayrollGroup,
		PayrollRuleSet:    repos.PayrollRuleSet,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: policy.currentPrincipal(), Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-payroll-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewPayrollServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	personID, _ := ids.NewID()
	ruleSetID, _ := ids.NewID()
	futureRuleSetID, _ := ids.NewID()
	expiredRuleSetID, _ := ids.NewID()
	artifactDocID, _ := ids.NewID()
	legalDocID, _ := ids.NewID()
	snapshotID, _ := ids.NewID()
	resultID, _ := ids.NewID()
	resultLineID, _ := ids.NewID()
	secondWorkerID, _ := ids.NewID()
	secondResultID, _ := ids.NewID()
	if resultID > secondResultID {
		resultID, secondResultID = secondResultID, resultID
	}
	secondResultLineID, _ := ids.NewID()
	var secondProfileID, fixtureResultID, fixtureLineID, replacementResultID, replacementLineID string
	var groupID, compID, periodID, profileID, adjID, draftRunID, calcRunID, replacementRunID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		err := database.WithContext(cleanupCtx).Transaction(func(tx *gorm.DB) error {
			for _, ddl := range []string{
				"ALTER TABLE hr_payroll_result_lines DISABLE TRIGGER tr_hr_payroll_line_guard",
				"ALTER TABLE hr_payroll_results DISABLE TRIGGER tr_hr_payroll_result_guard",
				"ALTER TABLE hr_payroll_runs DISABLE TRIGGER tr_hr_payroll_run_guard",
				"ALTER TABLE hr_payroll_rule_sets DISABLE TRIGGER tr_hr_payroll_rule_immutable",
			} {
				if err := tx.Exec(ddl).Error; err != nil {
					return err
				}
			}
			for _, stmt := range []struct {
				q string
				v string
			}{
				{"DELETE FROM hr_payroll_result_lines WHERE id = ?", replacementLineID},
				{"DELETE FROM hr_payroll_results WHERE id = ?", replacementResultID},
				{"DELETE FROM hr_payroll_result_lines WHERE id = ?", fixtureLineID},
				{"DELETE FROM hr_payroll_results WHERE id = ?", fixtureResultID},
				{"DELETE FROM hr_payroll_result_lines WHERE id = ?", resultLineID},
				{"DELETE FROM hr_payroll_result_lines WHERE id = ?", secondResultLineID},
				{"DELETE FROM hr_payroll_results WHERE id = ?", resultID},
				{"DELETE FROM hr_payroll_results WHERE id = ?", secondResultID},
				{"DELETE FROM hr_payroll_adjustments WHERE id = ?", adjID},
				{"DELETE FROM hr_payroll_runs WHERE id = ?", replacementRunID},
				{"DELETE FROM hr_payroll_runs WHERE id = ?", calcRunID},
				{"DELETE FROM hr_payroll_runs WHERE id = ?", draftRunID},
				{"DELETE FROM hr_payroll_profiles WHERE id = ?", profileID},
				{"DELETE FROM hr_payroll_profiles WHERE id = ?", secondProfileID},
				{"DELETE FROM hr_payroll_periods WHERE id = ?", periodID},
				{"DELETE FROM hr_pay_components WHERE id = ?", compID},
				{"DELETE FROM hr_payroll_groups WHERE id = ?", groupID},
				{"DELETE FROM hr_payroll_rule_sets WHERE id = ?", ruleSetID},
				{"DELETE FROM hr_payroll_rule_sets WHERE id = ?", futureRuleSetID},
				{"DELETE FROM hr_payroll_rule_sets WHERE id = ?", expiredRuleSetID},
				{"DELETE FROM hr_workers WHERE id = ?", workerID},
				{"DELETE FROM hr_workers WHERE id = ?", secondWorkerID},
			} {
				if stmt.v != "" {
					if err := tx.Exec(stmt.q, stmt.v).Error; err != nil {
						return err
					}
				}
			}
			for _, agg := range []string{adjID, calcRunID, draftRunID, replacementRunID, profileID, secondProfileID, periodID, compID, groupID} {
				if agg != "" {
					if err := tx.Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
						return err
					}
				}
			}
			for _, ddl := range []string{
				"ALTER TABLE hr_payroll_result_lines ENABLE TRIGGER tr_hr_payroll_line_guard",
				"ALTER TABLE hr_payroll_results ENABLE TRIGGER tr_hr_payroll_result_guard",
				"ALTER TABLE hr_payroll_runs ENABLE TRIGGER tr_hr_payroll_run_guard",
				"ALTER TABLE hr_payroll_rule_sets ENABLE TRIGGER tr_hr_payroll_rule_immutable",
			} {
				if err := tx.Exec(ddl).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Errorf("payroll integration cleanup failed: %v", err)
		}
	})

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date)
		 VALUES (?, ?, ?, 'employee', 'active', ?, ?, '2026-01-01', '2026-01-01')`,
		workerID, "PAY-W-"+workerID, personID, legalID, starterID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_payroll_rule_sets (id, rule_code, rule_version, jurisdiction_code, effective_from,
		 artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id,
		 approved_by_principal_id, approved_at)
		 VALUES (?, ?, 'v1', 'ID', '2026-01-01', ?, ?, ?, ?, NOW())`,
		ruleSetID, "PAY-RULE-"+ruleSetID, artifactDocID, ruleDigest[:], legalDocID, approverID,
	).Error; err != nil {
		t.Fatal(err)
	}

	// Rule set tidak boleh dibuat langsung tanpa sertifikasi DRM.
	if _, err := client.CreatePayrollRuleSet(ctx, &pb.CreatePayrollRuleSetRequest{Input: &pb.PayrollRuleSetInput{
		RuleCode: text("BLOCKED-" + ruleSetID), RuleVersion: text("v1"), JurisdictionCode: text("ID"),
		EffectiveFrom: text("2026-01-01"), ArtifactDocumentVersionReferenceId: &artifactDocID,
		ArtifactDigest: ruleDigest[:], LegalBasisDocumentVersionReferenceId: &legalDocID,
		ApprovedByPrincipalId: &approverID, ApprovedAt: timestamppb.Now(),
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct rule set create = %v, want FailedPrecondition", status.Code(err))
	}

	scale := int32(2)
	active := true
	if _, err := client.CreatePayrollGroup(ctx, &pb.CreatePayrollGroupRequest{Input: &pb.PayrollGroupInput{
		GroupCode: text("PAY-GRP-BAD-" + legalID), GroupName: text("Invalid TZ"),
		EmployingLegalEntityReferenceId: &legalID, JurisdictionCode: text("ID"),
		CurrencyCode: text("IDR"), CurrencyScale: &scale, PayFrequency: text("monthly"),
		Timezone: text("Local"), IsActive: &active,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Local timezone status = %v, want InvalidArgument", status.Code(err))
	}

	grpRes, err := client.CreatePayrollGroup(ctx, &pb.CreatePayrollGroupRequest{Input: &pb.PayrollGroupInput{
		GroupCode: text("PAY-GRP-" + legalID), GroupName: text("Monthly HQ"),
		EmployingLegalEntityReferenceId: &legalID, JurisdictionCode: text("ID"),
		CurrencyCode: text("IDR"), CurrencyScale: &scale, PayFrequency: text("monthly"),
		Timezone: text("Asia/Jakarta"), IsActive: &active,
	}})
	if err != nil {
		t.Fatalf("create payroll group: %v", err)
	}
	groupID = grpRes.GetItem().GetId()

	compVer := int32(1)
	compRes, err := client.CreatePayComponent(ctx, &pb.CreatePayComponentRequest{Input: &pb.PayComponentInput{
		ComponentCode: text("BASE-" + legalID), ComponentVersion: &compVer,
		ComponentName: text("Base Salary"), ComponentType: text("earning"),
		CalculationMethod: text("fixed"), JurisdictionCode: text("ID"),
		CurrencyCode: text("IDR"), EffectiveFrom: text("2026-01-01"), IsActive: &active,
	}})
	if err != nil {
		t.Fatalf("create pay component: %v", err)
	}
	compID = compRes.GetItem().GetId()

	periodInput := &pb.PayrollPeriodInput{
		PayrollGroupId: &groupID, EmployingLegalEntityReferenceId: &legalID,
		CurrencyCode: text("IDR"), JurisdictionCode: text("ID"), PeriodCode: text("2026-11"),
		PeriodStart: text("2026-11-01"), PeriodEnd: text("2026-11-30"),
		PayDate: text("2026-11-27"), PeriodStatus: text("open"),
	}
	outsideEntityLifetime := proto.CloneOf(periodInput)
	outsideEntityLifetime.PeriodStart = text("2026-10-31")
	if _, err := client.CreatePayrollPeriod(ctx, &pb.CreatePayrollPeriodRequest{Input: outsideEntityLifetime}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("payroll period outside EF legal entity lifetime: %v", err)
	}
	perRes, err := client.CreatePayrollPeriod(ctx, &pb.CreatePayrollPeriodRequest{Input: periodInput})
	if err != nil {
		t.Fatalf("create payroll period: %v", err)
	}
	periodID = perRes.GetItem().GetId()
	calculator.snapshotScope = service.CalculationSnapshot{PeriodID: periodID, GroupID: groupID, EmployerID: legalID}
	reschedule := &pb.UpdatePayrollPeriodRequest{
		Id: periodID, Input: &pb.PayrollPeriodInput{PayDate: text("2026-11-28")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"pay_date"}},
	}
	updatedPeriod, err := client.UpdatePayrollPeriod(ctx, reschedule)
	if err != nil || updatedPeriod.GetItem().GetPayDate() != "2026-11-28" {
		t.Fatalf("reschedule unreferenced period: %v", err)
	}

	profInput := &pb.PayrollProfileInput{
		WorkerId: &workerID, PayrollCountryCode: text("ID"), PayrollJurisdictionCode: text("ID"),
		PayrollGroupId: &groupID, EmployingLegalEntityReferenceId: &legalID,
		CurrencyCode: text("IDR"), EffectiveFrom: text("2026-11-01"), IsActive: &active,
	}
	profRes, err := client.CreatePayrollProfile(ctx, &pb.CreatePayrollProfileRequest{Input: profInput})
	if err != nil {
		t.Fatalf("create payroll profile: %v", err)
	}
	profileID = profRes.GetItem().GetId()
	if _, err := client.CreatePayrollProfile(ctx, &pb.CreatePayrollProfileRequest{Input: profInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("overlapping payroll profile = %v, want AlreadyExists", status.Code(err))
	}

	adjRes, err := client.CreatePayrollAdjustment(ctx, &pb.CreatePayrollAdjustmentRequest{Input: &pb.PayrollAdjustmentInput{
		WorkerId: &workerID, PayrollPeriodId: &periodID, ComponentId: &compID,
		AdjustmentType: text("bonus"), Amount: text("250000.000000"), CurrencyCode: text("IDR"),
		Reason: text("Monthly performance bonus"), IdempotencyKey: text("idem-adj-" + workerID),
	}})
	if err != nil {
		t.Fatalf("create payroll adjustment: %v", err)
	}
	adjID = adjRes.GetItem().GetId()
	reschedule.Input.PayDate = text("2026-11-29")
	if _, err := client.UpdatePayrollPeriod(ctx, reschedule); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("reschedule period referenced by adjustment: %v", err)
	}

	runNum1 := int32(1)
	draftInput := &pb.PayrollRunInput{
		PayrollPeriodId: &periodID, PayrollGroupId: &groupID, EmployingLegalEntityReferenceId: &legalID,
		CurrencyCode: text("IDR"), RuleSetId: &ruleSetID, RunNumber: &runNum1, RunType: text("regular"),
		Status: text("draft"), JurisdictionCode: text("ID"), RuleEngineName: text("id-tax"),
		RuleEngineVersion: text("1.0"), InputSnapshotReferenceId: &snapshotID,
		InputDigest: inputDigest[:], IdempotencyKey: text("idem-run-1-" + periodID),
	}
	runRes, err := client.CreatePayrollRun(ctx, &pb.CreatePayrollRunRequest{Input: draftInput})
	if err != nil {
		t.Fatalf("create draft payroll run: %v", err)
	}
	draftRunID = runRes.GetItem().GetId()
	if _, err := client.CreatePayrollRun(ctx, &pb.CreatePayrollRunRequest{Input: draftInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate run_number = %v, want AlreadyExists", status.Code(err))
	}

	runNum2 := int32(2)
	calcInput := proto.CloneOf(draftInput)
	calcInput.RunNumber = &runNum2
	calcInput.IdempotencyKey = text("idem-run-2-" + periodID)
	calcRes, err := client.CreatePayrollRun(ctx, &pb.CreatePayrollRunRequest{Input: calcInput})
	if err != nil {
		t.Fatalf("create second payroll run: %v", err)
	}
	calcRunID = calcRes.GetItem().GetId()

	if err := database.WithContext(ctx).Exec(
		`UPDATE hr_payroll_runs SET status = 'calculated', output_digest = ?, calculated_at = NOW() WHERE id = ?`,
		bytes.Repeat([]byte{0x44}, 32), calcRunID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_payroll_results (id, payroll_run_id, payroll_group_id, worker_id, payroll_profile_reference_id,
		 currency_code, gross_amount, reimbursement_amount, total_deductions, employer_cost_amount, net_pay_amount,
		 result_status, calculation_digest)
		 VALUES (?, ?, ?, ?, ?, 'IDR', 100.000000, 0.000000, 0.000000, 100.000000, 100.000000, 'calculated', ?)`,
		resultID, calcRunID, groupID, workerID, profileID, bytes.Repeat([]byte{0x55}, 32),
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_payroll_result_lines (id, payroll_result_id, component_id, component_code_snapshot,
		 component_name_snapshot, component_type_snapshot, amount, line_sequence)
		 VALUES (?, ?, ?, ?, 'Base Salary', 'earning', 100.000000, 1)`,
		resultLineID, resultID, compID, "BASE-"+legalID,
	).Error; err != nil {
		t.Fatal(err)
	}

	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_workers
		(id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id, hire_date)
		VALUES (?, ?, ?, 'employee', 'active', ?, '2026-01-01')`,
		secondWorkerID, "PAY-SECOND-"+secondWorkerID, secondWorkerID, legalID).Error; err != nil {
		t.Fatal(err)
	}
	secondProfileInput := proto.CloneOf(profInput)
	secondProfileInput.WorkerId = &secondWorkerID
	secondProfile, err := client.CreatePayrollProfile(ctx, &pb.CreatePayrollProfileRequest{Input: secondProfileInput})
	if err != nil {
		t.Fatalf("create payroll profile for second worker: %v", err)
	}
	secondProfileID = secondProfile.GetItem().GetId()
	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_payroll_results
		(id, payroll_run_id, payroll_group_id, worker_id, payroll_profile_reference_id,
		currency_code, gross_amount, reimbursement_amount, total_deductions, employer_cost_amount, net_pay_amount,
		result_status, calculation_digest)
		SELECT ?, payroll_run_id, payroll_group_id, ?, ?, currency_code, gross_amount, reimbursement_amount,
		total_deductions, employer_cost_amount, net_pay_amount, result_status, calculation_digest
		FROM hr_payroll_results WHERE id = ?`, secondResultID, secondWorkerID, secondProfileID, resultID).Error; err != nil {
		t.Fatal(err)
	}

	reviewRequest := &common.IDRequest{Id: calcRunID}
	if _, err := client.SubmitPayrollRunForReview(ctx, reviewRequest); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("submit review with incomplete payroll result: %v", err)
	}

	policy.setPrincipal(approverID)
	if _, err := client.ApprovePayrollRun(ctx, reviewRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve without submitting review: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_runs SET status = 'approved', approved_by_principal_id = ?, approved_at = NOW() WHERE id = ?", approverID, calcRunID).Error; err == nil ||
		!strings.Contains(err.Error(), "Payroll approval requires review") {
		t.Fatalf("database accepted approval without review: %v", err)
	}

	if err := database.WithContext(ctx).Exec(`INSERT INTO hr_payroll_result_lines
		(id, payroll_result_id, component_id, component_code_snapshot, component_name_snapshot,
		component_type_snapshot, amount, line_sequence)
		SELECT ?, ?, component_id, component_code_snapshot, component_name_snapshot,
		component_type_snapshot, amount, line_sequence FROM hr_payroll_result_lines WHERE id = ?`,
		secondResultLineID, secondResultID, resultLineID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.SubmitPayrollRunForReview(ctx, reviewRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("review submission by another principal: %v", err)
	}
	policy.setPrincipal(starterID)
	review, err := client.SubmitPayrollRunForReview(ctx, reviewRequest)
	if err != nil || review.GetItem().GetStatus() != "review" || review.GetItem().GetApprovedAt() != nil {
		t.Fatalf("submit review: %+v, %v", review, err)
	}
	if _, err := client.ApprovePayrollRun(ctx, reviewRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approval after review: %v", err)
	}
	for _, mutation := range []struct{ query, id string }{
		{"UPDATE hr_payroll_runs SET output_digest = decode('01', 'hex') WHERE id = ?", calcRunID},
		{"UPDATE hr_payroll_results SET calculation_digest = decode('01', 'hex') WHERE id = ?", resultID},
		{"UPDATE hr_payroll_result_lines SET component_name_snapshot = 'Modified' WHERE id = ?", resultLineID},
		{"UPDATE hr_payroll_runs SET status = 'calculated' WHERE id = ?", calcRunID},
	} {
		if err := database.WithContext(ctx).Exec(mutation.query, mutation.id).Error; err == nil {
			t.Fatalf("database accepted mutation of payroll snapshot under review: %s", mutation.query)
		}
	}
	if _, err := client.SubmitPayrollRunForReview(ctx, reviewRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat review submission: %v", err)
	}
	policy.setPrincipal(approverID)
	if _, err := client.RequestPayrollPosting(ctx, reviewRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("posting request before approval: %v", err)
	}
	approved, err := client.ApprovePayrollRun(ctx, &common.IDRequest{Id: calcRunID})
	if err != nil || approved.GetItem().GetStatus() != "approved" || approved.GetItem().GetApprovedByPrincipalId() != approverID {
		t.Fatalf("approve payroll run: res=%+v err=%v", approved, err)
	}
	if _, err := client.ApprovePayrollRun(ctx, &common.IDRequest{Id: calcRunID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat approval = %v, want FailedPrecondition", status.Code(err))
	}
	posting, err := client.RequestPayrollPosting(ctx, reviewRequest)
	if err != nil || posting.GetItem().GetStatus() != "approved" || posting.GetItem().GetPostingRequestedAt() == nil ||
		posting.GetItem().GetPostingRequestedByPrincipalId() != approverID || posting.GetItem().GetFinanceJournalReferenceId() != "" {
		t.Fatalf("request payroll posting: %+v, %v", posting, err)
	}
	if _, err := client.RequestPayrollPosting(ctx, reviewRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat posting request: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_runs SET posting_requested_by_principal_id = ? WHERE id = ?", starterID, calcRunID).Error; err == nil {
		t.Fatal("database accepted overwriting existing posting requester")
	}
	var outboxCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", calcRunID).Count(&outboxCount).Error; err != nil || outboxCount != 4 {
		t.Fatalf("payroll run outbox count=%d err=%v", outboxCount, err)
	}
	var matchingPostingEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type = ?", calcRunID, "hr_payroll_runs.request_posting").
		Where("payload->'details'->>'output_digest_sha256' = ?", strings.Repeat("44", 32)).
		Where("payload->'details'->>'input_digest_sha256' = ?", fmt.Sprintf("%x", inputDigest)).
		Where("payload->'details'->>'input_snapshot_reference_id' = ?", snapshotID).
		Where("payload->'details'->>'payroll_period_id' = ?", periodID).
		Where("payload->'details'->>'payroll_group_id' = ?", groupID).
		Where("payload->'details'->>'employing_legal_entity_reference_id' = ?", legalID).
		Where("payload->'details'->>'currency_code' = ? AND payload->'details'->>'settlement_direction' = ?", "IDR", "payable").
		Count(&matchingPostingEvents).Error; err != nil || matchingPostingEvents != 1 {
		t.Fatalf("posting event must identify the approved payroll snapshot: count=%d err=%v", matchingPostingEvents, err)
	}

	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET payslip_document_reference_id = ? WHERE id = ?", artifactDocID, resultID).Error; err == nil ||
		!strings.Contains(err.Error(), "Payroll must be posted before attaching a payslip") {
		t.Fatalf("attach payslip before posting: %v", err)
	}
	delivery := inbox.PayrollPublicationDelivery{
		SourceManagement: "finance_accounting", EventType: "payroll.published",
		SchemaVersion: 1, PublicationID: legalDocID, PayrollRunID: calcRunID,
	}
	if err := inbox.HandlePayrollPublication(ctx, svc, delivery); !errors.Is(err, hrerrors.ErrWorkflowRequired) ||
		publication.storeCalls.Load() != 0 {
		t.Fatalf("publication before posting must remain unacknowledged: %v, stores=%d", err, publication.storeCalls.Load())
	}
	postingDelivery := inbox.PayrollPostingDelivery{
		SourceManagement: "finance_accounting", EventType: "payroll.posted", SchemaVersion: 1,
		PayrollRunID: calcRunID, JournalID: snapshotID,
	}
	forgedPosting := postingDelivery
	forgedPosting.SourceManagement = "human_resource"
	if err := inbox.HandlePayrollPosting(ctx, svc, forgedPosting); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("posting from non-Finance workload: %v", err)
	}
	if err := inbox.HandlePayrollPosting(ctx, svc, postingDelivery); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("posting without proof: %v", err)
	}
	var postingChecks int
	verifyPosting = func(_ context.Context, check service.PayrollPostingCheck) (time.Time, error) {
		postingChecks++
		if check.RunID != calcRunID || check.JournalID != snapshotID || check.LegalEntityID != legalID ||
			check.PeriodID != periodID || check.GroupID != groupID || check.InputSnapshotID != snapshotID ||
			check.CurrencyCode != "IDR" || check.SettlementDirection != "payable" || check.RequestedAt.IsZero() ||
			!bytes.Equal(check.InputDigest, inputDigest[:]) || !bytes.Equal(check.OutputDigest, bytes.Repeat([]byte{0x44}, 32)) {
			return time.Time{}, hrerrors.ErrReferenceConflict
		}
		return check.RequestedAt, nil
	}
	if err := inbox.HandlePayrollPosting(ctx, svc, postingDelivery); err != nil {
		t.Fatalf("project verified Finance journal: %v", err)
	}
	if err := inbox.HandlePayrollPosting(ctx, svc, postingDelivery); err != nil || postingChecks != 1 {
		t.Fatalf("replay Finance journal: %v, checks=%d", err, postingChecks)
	}
	conflictingPosting := postingDelivery
	conflictingPosting.JournalID = legalDocID
	if err := inbox.HandlePayrollPosting(ctx, svc, conflictingPosting); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("replace verified Finance journal: %v", err)
	}
	publication.publicationID, publication.runID, publication.journalID = legalDocID, calcRunID, snapshotID
	publication.inputSnapshotID, publication.inputDigest = snapshotID, inputDigest[:]
	publication.legalEntityID, publication.resultID, publication.workerID, publication.documentID = legalID, resultID, workerID, artifactDocID
	if _, err := svc.PublishPayslipForResult(ctx, snapshotID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 0 {
		t.Fatalf("verify publication before storing payslip: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.publishedAt = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 0 {
		t.Fatalf("publication predating payroll approval: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.publishedAt = approved.GetItem().GetApprovedAt().AsTime().Truncate(time.Microsecond)
	if !publication.publishedAt.Before(posting.GetItem().GetPostingRequestedAt().AsTime()) {
		t.Fatal("fixture posting request must follow approval")
	}
	if _, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 0 {
		t.Fatalf("publication predating posting request: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.publishedAt = time.Now().UTC().Add(time.Hour)
	if _, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 0 {
		t.Fatalf("future payroll publication: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.publishedAt = time.Now().UTC()
	publication.inputSnapshotID = legalDocID
	if _, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 0 {
		t.Fatalf("publication with a different input snapshot: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.inputSnapshotID = snapshotID
	publication.inputDigest = bytes.Repeat([]byte{0x34}, 32)
	if _, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 0 {
		t.Fatalf("publication with a different input digest: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.inputDigest = inputDigest[:]
	type publishResult struct {
		documentID string
		err        error
	}
	publication.storeEntered, publication.storeRelease = make(chan struct{}), make(chan struct{})
	attempts := make(chan publishResult, 2)
	publish := func() {
		documentID, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID)
		attempts <- publishResult{documentID: documentID, err: err}
	}
	go publish()
	select {
	case <-publication.storeEntered:
	case <-ctx.Done():
		t.Fatal("first publication did not reach storage")
	}
	go publish()
	lockContext, cancelLock := context.WithTimeout(ctx, 5*time.Second)
	var lockObserved bool
	for lockContext.Err() == nil {
		err := database.WithContext(lockContext).Raw(
			"SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE datname = current_database() " +
				"AND wait_event_type = 'Lock' AND query LIKE '%hr_payroll_runs%FOR UPDATE%')",
		).Scan(&lockObserved).Error
		if err != nil || lockObserved {
			break
		}
		select {
		case <-lockContext.Done():
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancelLock()
	close(publication.storeRelease)
	for range 2 {
		select {
		case attempt := <-attempts:
			if attempt.err != nil || attempt.documentID != artifactDocID {
				t.Fatalf("concurrent publication failed: %+v", attempt)
			}
		case <-ctx.Done():
			t.Fatal("concurrent publication did not finish")
		}
	}
	if !lockObserved || publication.storeCalls.Load() != 1 {
		t.Fatalf("expected a competing publication and one storage attempt: lock=%t, stores=%d", lockObserved, publication.storeCalls.Load())
	}
	issuedDocument := artifactDocID
	publication.publicationID = snapshotID
	if _, err := svc.PublishPayslipForResult(ctx, snapshotID, resultID); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 1 {
		t.Fatalf("different verified publication must not reuse the first slip: %v, stores=%d", err, publication.storeCalls.Load())
	}
	publication.publicationID = legalDocID
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET payslip_publication_reference_id = ? WHERE id = ?", snapshotID, resultID).Error; err == nil {
		t.Fatal("database allowed replacing the slip publication reference")
	}
	replayedDocument, err := svc.PublishPayslipForResult(ctx, legalDocID, resultID)
	if err != nil || replayedDocument != issuedDocument || publication.storeCalls.Load() != 1 {
		t.Fatalf("replay payslip publication: %s, %v, stores=%d", replayedDocument, err, publication.storeCalls.Load())
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET payslip_document_reference_id = ? WHERE id = ?", legalDocID, resultID).Error; err == nil {
		t.Fatal("database allowed replacing issued payslip reference")
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET payslip_document_reference_id = ? WHERE id = ?", artifactDocID, secondResultID).Error; err == nil ||
		!strings.Contains(err.Error(), "uq_hr_payroll_result_payslip") {
		t.Fatalf("attach one payslip to two workers: %v", err)
	}
	storedSlip, err := client.GetPayrollResultByID(ctx, &common.IDRequest{Id: resultID})
	if err != nil || storedSlip.GetItem().GetPayslipDocumentReferenceId() != artifactDocID {
		t.Fatalf("read issued payslip reference: %v", err)
	}

	policy.setPrincipal(starterID)
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("iam_principal_reference_id", starterID).Error; err != nil {
		t.Fatalf("link owner in payslip fixture: %v", err)
	}
	payslipVerifier.expected = service.PayslipCheck{
		ResultID: resultID, RunID: calcRunID, WorkerID: workerID, PrincipalID: starterID, DocumentVersionID: artifactDocID, PublicationID: legalDocID,
	}
	mine := &pb.GetMyPayslipRequest{PayrollResultId: resultID}
	if _, err := client.GetMyPayslip(ctx, mine); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("posted run without Finance publication: %v", err)
	}
	payslipVerifier.published = true
	payslipVerifier.publishedAt = time.Now().UTC().Add(time.Hour)
	if _, err := client.GetMyPayslip(ctx, mine); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("read payslip before verified publication date: %v", err)
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("list payslip before verified publication date: %v", err)
	}
	payslipVerifier.publishedAt = approved.GetItem().GetApprovedAt().AsTime().Add(-time.Second)
	if _, err := client.GetMyPayslip(ctx, mine); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("read slip published before approval: %v", err)
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("list slip published before approval: %v", err)
	}
	payslipVerifier.publishedAt = publication.publishedAt
	payslipVerifier.omitPublicationID = true
	if _, err := client.GetMyPayslip(ctx, mine); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("read payslip with missing Finance publication ID: %v", err)
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("list payslip with missing Finance publication ID: %v", err)
	}
	payslipVerifier.omitPublicationID = false
	payslipVerifier.verifiedPeriodID = snapshotID
	if _, err := client.GetMyPayslip(ctx, mine); status.Code(err) != codes.FailedPrecondition {
		t.Errorf("read payslip with a different verified period: %v", err)
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Errorf("list payslip with a different verified period: %v", err)
	}
	payslipVerifier.verifiedPeriodID = periodID
	payslip, err := client.GetMyPayslip(ctx, mine)
	if err != nil || payslip.GetPayrollResultId() != resultID || payslip.GetDocumentVersionId() != artifactDocID ||
		payslip.GetPaymentStatus() != "scheduled" || payslip.GetScheduledPayDate() != "2026-11-25" ||
		payslip.GetPayrollPeriodId() != periodID || payslip.GetPeriodStart() != "2026-11-01" || payslip.GetPeriodEnd() != "2026-11-30" {
		t.Fatalf("published own payslip: %+v, %v", payslip, err)
	}
	if len(payslip.GetPdf()) != 0 || payslipVerifier.pdfCalls != 0 {
		t.Fatal("metadata response retrieved the private PDF")
	}
	pdfRequest := &pb.GetMyPayslipRequest{PayrollResultId: resultID, IncludePdf: true}
	if _, err := client.GetMyPayslip(ctx, pdfRequest); status.Code(err) != codes.Unavailable {
		t.Fatalf("read payslip with unavailable DRM: %v", err)
	}
	content := []byte("%PDF-1.7\nfixture")
	digest := sha256.Sum256(content)
	payslipVerifier.pdf = service.VerifiedPayslipPDF{DocumentVersionID: artifactDocID, Content: content, Digest: digest}
	opened, err := client.GetMyPayslip(ctx, pdfRequest)
	if err != nil || !bytes.Equal(opened.GetPdf(), content) || opened.GetDocumentVersionId() != artifactDocID {
		t.Fatalf("open verified PDF: %v", err)
	}
	payslipVerifier.beforePDFRead = func(readCtx context.Context) error {
		return database.WithContext(readCtx).Table("hr_workers").Where("id = ?", workerID).
			Update("iam_principal_reference_id", nil).Error
	}
	stale, err := client.GetMyPayslip(ctx, pdfRequest)
	if status.Code(err) != codes.PermissionDenied || len(stale.GetPdf()) != 0 {
		t.Fatalf("ownership changed during PDF read: %+v, %v", stale, err)
	}
	payslipVerifier.beforePDFRead = nil
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("iam_principal_reference_id", starterID).Error; err != nil {
		t.Fatal(err)
	}
	payslipVerifier.beforePDFRead = func(readCtx context.Context) error {
		return database.WithContext(readCtx).Table("hr_workers").Where("id = ?", workerID).
			Updates(map[string]any{"status": "terminated", "termination_date": time.Now().UTC()}).Error
	}
	stale, err = client.GetMyPayslip(ctx, pdfRequest)
	if status.Code(err) != codes.PermissionDenied || len(stale.GetPdf()) != 0 {
		t.Fatalf("employment ended during PDF read: %+v, %v", stale, err)
	}
	payslipVerifier.beforePDFRead = nil
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Updates(map[string]any{"status": "active", "termination_date": nil}).Error; err != nil {
		t.Fatal(err)
	}
	reissued, err := client.ReissueEmployeePayslip(ctx, &pb.ReissueEmployeePayslipRequest{WorkerId: workerID, PayrollResultId: resultID})
	if err != nil || !bytes.Equal(reissued.GetPdf(), content) || reissued.GetDocumentVersionId() != artifactDocID || publication.storeCalls.Load() != 1 {
		t.Fatalf("reissue existing private document: %v", err)
	}
	var reissueEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("event_type = ? AND payload->>'payroll_result_id' = ?", "hr_payslip_reissues.recorded", resultID).
		Where("payload->>'outcome' = ?", "reused").
		Count(&reissueEvents).Error; err != nil || reissueEvents != 1 {
		t.Fatalf("reissue audit event: count=%d, err=%v", reissueEvents, err)
	}
	for _, invalid := range []service.VerifiedPayslipPDF{
		{DocumentVersionID: legalDocID, Content: content, Digest: digest},
		{DocumentVersionID: artifactDocID, Content: content},
		{DocumentVersionID: artifactDocID, Content: bytes.Repeat(content, contract.MaxPayslipPDFBytes/len(content)+1)},
	} {
		payslipVerifier.pdf = invalid
		if _, err := client.GetMyPayslip(ctx, pdfRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("unsafe PDF content was accepted: %v", err)
		}
	}
	for _, tc := range []struct {
		name   string
		change func(*inbox.PayrollPublicationDelivery)
		want   error
	}{
		{"wrong source", func(d *inbox.PayrollPublicationDelivery) { d.SourceManagement = "human_resource" }, hrerrors.ErrForbidden},
		{"wrong event type", func(d *inbox.PayrollPublicationDelivery) { d.EventType = "payroll.paid" }, hrerrors.ErrInvalidInput},
		{"unsupported version", func(d *inbox.PayrollPublicationDelivery) { d.SchemaVersion = 2 }, hrerrors.ErrInvalidInput},
		{"invalid publication", func(d *inbox.PayrollPublicationDelivery) { d.PublicationID = "invalid" }, hrerrors.ErrInvalidInput},
		{"invalid run", func(d *inbox.PayrollPublicationDelivery) { d.PayrollRunID = "invalid" }, hrerrors.ErrInvalidInput},
	} {
		invalid := delivery
		tc.change(&invalid)
		if err := inbox.HandlePayrollPublication(ctx, svc, invalid); !errors.Is(err, tc.want) || publication.storeCalls.Load() != 1 {
			t.Fatalf("reject %s: %v, stores=%d", tc.name, err, publication.storeCalls.Load())
		}
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET result_status = 'paid' WHERE id = ?", secondResultID).Error; err != nil {
		t.Fatal(err)
	}
	if err := inbox.HandlePayrollPublication(ctx, svc, delivery); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != 1 {
		t.Fatalf("paid result without a payment reference: %v, stores=%d", err, publication.storeCalls.Load())
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET result_status = 'calculated' WHERE id = ?", secondResultID).Error; err != nil {
		t.Fatal(err)
	}
	paymentDelivery := inbox.PayrollPaymentDelivery{
		SourceManagement: "finance_accounting", EventType: "payroll.result_paid", SchemaVersion: 1,
		PayrollRunID: calcRunID, ResultID: secondResultID, PaymentID: snapshotID,
	}
	if err := inbox.HandlePayrollPayment(ctx, svc, paymentDelivery); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("payment without Finance proof: %v", err)
	}
	confirmationChecks := 0
	verifyPayment = func(_ context.Context, check service.PayrollPaymentCheck) error {
		confirmationChecks++
		amount, err := check.NetPayAmount.Rat()
		if err != nil || amount.RatString() != "100" || check.RunID != calcRunID || check.JournalID != snapshotID ||
			check.ResultID != secondResultID || check.WorkerID != secondWorkerID || check.PaymentID != snapshotID ||
			check.LegalEntityID != legalID || check.CurrencyCode != "IDR" || check.SettlementDirection != "payable" ||
			!bytes.Equal(check.RunOutputDigest, bytes.Repeat([]byte{0x44}, 32)) ||
			!bytes.Equal(check.CalculationDigest, bytes.Repeat([]byte{0x55}, 32)) {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	if err := inbox.HandlePayrollPayment(ctx, svc, paymentDelivery); err != nil {
		t.Fatalf("project verified Finance result payment: %v", err)
	}
	if err := inbox.HandlePayrollPayment(ctx, svc, paymentDelivery); err != nil || confirmationChecks != 1 {
		t.Fatalf("replay result payment: %v, checks=%d", err, confirmationChecks)
	}
	conflictingPayment := paymentDelivery
	conflictingPayment.PaymentID = legalDocID
	if err := inbox.HandlePayrollPayment(ctx, svc, conflictingPayment); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("replace verified payment: %v", err)
	}
	resultPayment, err := client.GetPayrollResultByID(ctx, &common.IDRequest{Id: secondResultID})
	if err != nil || resultPayment.GetItem().GetResultStatus() != "paid" || resultPayment.GetItem().GetPaymentReferenceId() != snapshotID {
		t.Fatalf("paid result did not retain its verified reference: %v", err)
	}
	postedRun, err := client.GetPayrollRunByID(ctx, &common.IDRequest{Id: calcRunID})
	if err != nil || postedRun.GetItem().GetStatus() != "posted" {
		t.Fatalf("individual payment must keep aggregate run posted: %v", err)
	}
	settlement := inbox.PayrollSettlementDelivery{
		SourceManagement: "finance_accounting", EventType: "payroll.settled", SchemaVersion: 1,
		PayrollRunID: calcRunID, BatchID: artifactDocID, SettlementID: legalDocID,
	}
	if err := inbox.HandlePayrollSettlement(ctx, svc, settlement); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("settlement before all individual confirmations: %v", err)
	}

	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET result_status = 'paid' WHERE id = ?", resultID).Error; err != nil {
		t.Fatal(err)
	}
	publication.resultID, publication.workerID, publication.documentID = secondResultID, secondWorkerID, "invalid"
	expectedStoreError := hrerrors.ErrReferenceConflict
	if publication.processor != nil {
		publication.documentID = legalDocID
		publication.lostResponseForResult = secondResultID
		expectedStoreError = hrerrors.ErrUnavailable
	}
	reissueMissing := &pb.ReissueEmployeePayslipRequest{WorkerId: secondWorkerID, PayrollResultId: secondResultID}
	mismatched := &pb.ReissueEmployeePayslipRequest{WorkerId: workerID, PayrollResultId: secondResultID}
	if _, err := client.ReissueEmployeePayslip(ctx, mismatched); status.Code(err) != codes.NotFound ||
		publication.storeCalls.Load() != 1 {
		t.Fatalf("mismatched worker and result: %v, stores=%d", err, publication.storeCalls.Load())
	}

	verificationCalls := 0
	publication.beforeVerify = func() {
		verificationCalls++
		if verificationCalls == 2 {
			policy.setPrincipal(approverID)
		}
	}
	if _, err := client.ReissueEmployeePayslip(ctx, reissueMissing); status.Code(err) != codes.PermissionDenied ||
		publication.storeCalls.Load() != 1 || verificationCalls != 2 {
		t.Fatalf("actor changed during Finance verification: %v, stores=%d, verifications=%d", err, publication.storeCalls.Load(), verificationCalls)
	}
	publication.beforeVerify = nil
	policy.setPrincipal(starterID)

	wantStoreCode := codes.FailedPrecondition
	if errors.Is(expectedStoreError, hrerrors.ErrUnavailable) {
		wantStoreCode = codes.Unavailable
	}
	if _, err := client.ReissueEmployeePayslip(ctx, reissueMissing); status.Code(err) != wantStoreCode || publication.storeCalls.Load() != 2 {
		t.Fatalf("reissue must attempt to store missing slip: %v, stores=%d", err, publication.storeCalls.Load())
	}
	if publication.processor != nil && publication.savedDocument == nil {
		t.Fatal("the storage fixture did not persist the PDF before losing the response")
	}
	var withoutDocument int64
	if err := database.WithContext(ctx).Model(&entity.PayrollResult{}).
		Where("id = ? AND payslip_document_reference_id IS NULL AND payslip_publication_reference_id IS NULL", secondResultID).
		Count(&withoutDocument).Error; err != nil || withoutDocument != 1 {
		t.Fatalf("interrupted publication left inconsistent references: count=%d, err=%v", withoutDocument, err)
	}
	publication.documentID = legalDocID
	originalCheck, originalPDF := payslipVerifier.expected, payslipVerifier.pdf
	payslipVerifier.expected = service.PayslipCheck{
		ResultID: secondResultID, RunID: calcRunID, WorkerID: secondWorkerID, PrincipalID: starterID,
		DocumentVersionID: legalDocID, PublicationID: publication.publicationID,
	}
	wantPDF := content
	if publication.savedDocument != nil {
		wantPDF = publication.savedDocument.PDF
	}
	payslipVerifier.pdf = service.VerifiedPayslipPDF{DocumentVersionID: legalDocID, Content: wantPDF, Digest: sha256.Sum256(wantPDF)}
	recovered, err := client.ReissueEmployeePayslip(ctx, reissueMissing)
	if err != nil || !bytes.Equal(recovered.GetPdf(), wantPDF) || recovered.GetDocumentVersionId() != legalDocID || publication.storeCalls.Load() != 3 {
		t.Fatalf("reissue retry did not return the stored version: %v, stores=%d", err, publication.storeCalls.Load())
	}
	for _, readError := range []error{hrerrors.ErrUnavailable, hrerrors.ErrForbidden, hrerrors.ErrNotFound} {
		payslipVerifier.reissueReadError = readError
		if slip, err := client.ReissueEmployeePayslip(ctx, reissueMissing); err == nil || len(slip.GetPdf()) != 0 ||
			publication.storeCalls.Load() != 3 || publication.restoreCalls != 0 {
			t.Fatalf("ambiguous DRM response triggered restore: %v, stores=%d", err, publication.storeCalls.Load())
		}
	}
	payslipVerifier.reissueReadError = nil

	if publication.processor != nil {
		payslipVerifier.missingContent = true
		restored, err := client.ReissueEmployeePayslip(ctx, reissueMissing)
		if err != nil || !bytes.Equal(restored.GetPdf(), wantPDF) || restored.GetDocumentVersionId() != legalDocID ||
			publication.restoreCalls != 1 || publication.storeCalls.Load() != 4 {
			t.Fatalf("restore immutable version: %v, restores=%d", err, publication.restoreCalls)
		}
		payslipVerifier.missingContent = false
	}
	outcomes := []string{"created"}
	if publication.processor != nil {
		outcomes = append(outcomes, "restored")
	}
	for _, outcome := range outcomes {
		if err := database.WithContext(ctx).Table("hr_outbox_events").
			Where("event_type = ? AND payload->>'payroll_result_id' = ? AND payload->>'outcome' = ?",
				"hr_payslip_reissues.recorded", secondResultID, outcome).
			Count(&reissueEvents).Error; err != nil || reissueEvents != 1 {
			t.Fatalf("reissue audit outcome %s: count=%d, err=%v", outcome, reissueEvents, err)
		}
	}

	storeAttempts := publication.storeCalls.Load()

	payslipVerifier.expected, payslipVerifier.pdf = originalCheck, originalPDF
	if err := inbox.HandlePayrollPublication(ctx, svc, delivery); !errors.Is(err, hrerrors.ErrReferenceConflict) || publication.storeCalls.Load() != storeAttempts {
		t.Fatalf("publish eligible slip while another result is invalid: %v, stores=%d", err, publication.storeCalls.Load())
	}
	var completedDocument int64
	if err := database.WithContext(ctx).Model(&entity.PayrollResult{}).
		Where("id = ? AND payslip_document_reference_id = ? AND payslip_publication_reference_id = ?", secondResultID, legalDocID, publication.publicationID).
		Count(&completedDocument).Error; err != nil || completedDocument != 1 {
		t.Fatalf("retry did not record the original version and publication: count=%d, err=%v", completedDocument, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET result_status = 'calculated' WHERE id = ?", resultID).Error; err != nil {
		t.Fatal(err)
	}
	if err := inbox.HandlePayrollPublication(ctx, svc, delivery); err != nil || publication.storeCalls.Load() != storeAttempts {
		t.Fatalf("replay completed run publication: %v, stores=%d", err, publication.storeCalls.Load())
	}
	verifyPayment = func(_ context.Context, check service.PayrollPaymentCheck) error {
		if check.ResultID != resultID || check.WorkerID != workerID || check.PaymentID != legalDocID {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	firstPayment := paymentDelivery
	firstPayment.ResultID, firstPayment.PaymentID = resultID, legalDocID
	if err := inbox.HandlePayrollPayment(ctx, svc, firstPayment); err != nil {
		t.Fatalf("confirm final individual payment: %v", err)
	}
	if err := inbox.HandlePayrollSettlement(ctx, svc, settlement); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("settlement without Finance proof: %v", err)
	}
	settlementChecks := 0
	verifySettlement = func(_ context.Context, check service.PayrollSettlementCheck) (time.Time, error) {
		settlementChecks++
		if check.RunID != calcRunID || check.JournalID != snapshotID || check.BatchID != artifactDocID ||
			check.SettlementID != legalDocID || check.LegalEntityID != legalID ||
			check.CurrencyCode != "IDR" || check.SettlementDirection != "payable" ||
			!bytes.Equal(check.OutputDigest, bytes.Repeat([]byte{0x44}, 32)) || len(check.Results) != 2 {
			return time.Time{}, hrerrors.ErrReferenceConflict
		}
		expected := map[string]string{resultID: legalDocID, secondResultID: snapshotID}
		for _, row := range check.Results {
			if expected[row.ResultID] != row.PaymentID {
				return time.Time{}, hrerrors.ErrReferenceConflict
			}
			delete(expected, row.ResultID)
		}
		return time.Now().UTC(), nil
	}
	if err := inbox.HandlePayrollSettlement(ctx, svc, settlement); err != nil {
		t.Fatalf("project verified Finance run settlement: %v", err)
	}
	if err := inbox.HandlePayrollSettlement(ctx, svc, settlement); err != nil || settlementChecks != 1 {
		t.Fatalf("replay run settlement: %v, checks=%d", err, settlementChecks)
	}
	conflictingSettlement := settlement
	conflictingSettlement.SettlementID = snapshotID
	if err := inbox.HandlePayrollSettlement(ctx, svc, conflictingSettlement); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("replace verified settlement: %v", err)
	}
	paidRun, err := client.GetPayrollRunByID(ctx, &common.IDRequest{Id: calcRunID})
	if err != nil || paidRun.GetItem().GetStatus() != "paid" || paidRun.GetItem().GetPaymentBatchReferenceId() != artifactDocID ||
		paidRun.GetItem().GetFinanceSettlementEventReferenceId() != legalDocID {
		t.Fatalf("paid run did not retain Finance evidence: %v", err)
	}
	if _, err := client.GetMyPayslip(ctx, &pb.GetMyPayslipRequest{PayrollResultId: secondResultID}); status.Code(err) != codes.NotFound {
		t.Fatalf("read another worker's payslip: %v", err)
	}
	pdfCalls := payslipVerifier.pdfCalls
	if _, err := client.GetMyPayslip(ctx, &pb.GetMyPayslipRequest{PayrollResultId: secondResultID, IncludePdf: true}); status.Code(err) != codes.NotFound ||
		payslipVerifier.pdfCalls != pdfCalls {
		t.Fatalf("read another worker's PDF: %v, reads=%d", err, payslipVerifier.pdfCalls-pdfCalls)
	}
	mySlips, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1, PayrollPeriodId: &periodID})
	if err != nil || len(mySlips.GetItems()) != 1 || mySlips.GetPage().GetTotal() != 1 ||
		mySlips.GetItems()[0].GetPayrollResultId() != resultID ||
		mySlips.GetItems()[0].GetPayrollPeriodId() != periodID || mySlips.GetItems()[0].GetPeriodCode() != "2026-11" {
		t.Fatalf("list published worker slips: %+v, %v", mySlips, err)
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1, PayrollPeriodId: &snapshotID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("list verifier returned a slip outside the requested payroll period: %v", err)
	}
	emptySlips, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1, Offset: 1})
	if err != nil || len(emptySlips.GetItems()) != 0 || emptySlips.GetPage().GetTotal() != 1 {
		t.Fatalf("list slips past final page: %+v, %v", emptySlips, err)
	}
	payslipVerifier.expected.PublicationID = snapshotID
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("list verifier returned a different publication: %v", err)
	}
	payslipVerifier.expected.PublicationID = legalDocID
	payslipVerifier.expected.ResultID, payslipVerifier.expected.DocumentVersionID = secondResultID, legalDocID
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("list verifier returned another worker's slip: %v", err)
	}
	historicalWorker := database.WithContext(ctx).Table("hr_workers").Where("id = ?", secondWorkerID)
	if err := historicalWorker.Updates(map[string]any{
		"status": "terminated", "iam_principal_reference_id": starterID, "termination_date": time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetMyPayslip(ctx, mine); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("read slip with conflicting worker identity: %v", err)
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("list slips with conflicting worker identity: %v", err)
	}
	if err := historicalWorker.Update("person_party_reference_id", personID).Error; err != nil {
		t.Fatal(err)
	}
	payslipVerifier.expected.WorkerID = secondWorkerID
	historicalSlip, err := client.GetMyPayslip(ctx, &pb.GetMyPayslipRequest{PayrollResultId: secondResultID})
	if err != nil || historicalSlip.GetPayrollResultId() != secondResultID {
		t.Fatalf("read slip from former employment: %+v, %v", historicalSlip, err)
	}
	currentWorker := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID)
	if err := currentWorker.Updates(map[string]any{"status": "terminated", "termination_date": time.Now().UTC()}).Error; err != nil {
		t.Fatal(err)
	}
	for _, request := range []*pb.GetMyPayslipRequest{mine, pdfRequest} {
		if _, err := client.GetMyPayslip(ctx, request); status.Code(err) != codes.PermissionDenied {
			t.Fatalf("read slip after all employments ended: %v", err)
		}
	}
	if _, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("list slips after all employments ended: %v", err)
	}
	if err := currentWorker.Updates(map[string]any{"status": "active", "termination_date": nil}).Error; err != nil {
		t.Fatal(err)
	}
	// Riwayat uji melebihi batas jumlah nilai dalam satu filter repository.
	if err := database.WithContext(ctx).Exec(`
		INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, iam_principal_reference_id, termination_date)
		SELECT CAST(lpad(to_hex(n), 32, '0') AS uuid), 'PAY-HISTORY-' || n, ?, 'employee', 'terminated', ?, ?, CURRENT_DATE
		FROM generate_series(1, 1001) AS n`, personID, legalID, starterID).Error; err != nil {
		t.Fatal(err)
	}
	historyPage, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1})
	if err != nil || len(historyPage.GetItems()) != 1 || historyPage.GetItems()[0].GetPayrollResultId() != secondResultID {
		t.Fatalf("list former slip with long employment history: %+v, %v", historyPage, err)
	}
	if err := database.WithContext(ctx).Exec(
		"DELETE FROM hr_workers WHERE worker_number LIKE ? AND iam_principal_reference_id = ?", "PAY-HISTORY-%", starterID,
	).Error; err != nil {
		t.Fatal(err)
	}
	payslipVerifier.otherEmployment = &service.PayslipCheck{
		ResultID: resultID, RunID: calcRunID, WorkerID: workerID, PrincipalID: starterID,
		DocumentVersionID: artifactDocID, PublicationID: legalDocID,
	}
	for offset, expectedID := range []string{secondResultID, resultID} {
		page, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1, Offset: int32(offset)})
		if err != nil || page.GetPage().GetTotal() != 2 || len(page.GetItems()) != 1 ||
			page.GetItems()[0].GetPayrollResultId() != expectedID {
			t.Fatalf("list slips from both employments at offset %d: %+v, %v", offset, page, err)
		}
	}
	tailPage, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1, Offset: 2})
	if err != nil || tailPage.GetPage().GetTotal() != 2 || len(tailPage.GetItems()) != 0 {
		t.Fatalf("list slips past both employments: %+v, %v", tailPage, err)
	}
	payslipVerifier.beforeListReturn = func(listCtx context.Context) error {
		return database.WithContext(listCtx).Table("hr_workers").Where("id = ?", secondWorkerID).
			Update("iam_principal_reference_id", nil).Error
	}
	stalePage, err := client.ListMyPayslips(ctx, &pb.ListMyPayslipsRequest{Limit: 1, Offset: 2})
	if status.Code(err) != codes.PermissionDenied || stalePage != nil {
		t.Fatalf("list returned page or total after losing one of two worker links: %+v, %v", stalePage, err)
	}
	payslipVerifier.beforeListReturn = nil
	payslipVerifier.otherEmployment = nil
	if err := historicalWorker.Updates(map[string]any{
		"status": "active", "iam_principal_reference_id": nil, "termination_date": nil, "person_party_reference_id": secondWorkerID,
	}).Error; err != nil {
		t.Fatal(err)
	}
	policy.setPrincipal(approverID)

	fixtureResultID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	fixtureLineID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	calculator.result = entity.PayrollResult{
		ID: fixtureResultID, WorkerID: workerID, PayrollProfileReferenceID: profileID,
		GrossAmount: "100", ReimbursementAmount: "0", TotalDeductions: "0", EmployerCostAmount: "100", NetPayAmount: "100",
		ResultStatus: "calculated", CalculationDigest: bytes.Repeat([]byte{0x58}, 32),
	}
	calculator.line = entity.PayrollResultLine{
		ID: fixtureLineID, PayrollResultID: fixtureResultID, ComponentID: compID,
		ComponentCodeSnapshot: "BASE-" + legalID, ComponentNameSnapshot: "Base Salary",
		ComponentTypeSnapshot: "earning", Amount: "100", LineSequence: 1,
	}
	calculate := &common.IDRequest{Id: draftRunID}
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("calculate another principal's run: %v", err)
	}
	policy.setPrincipal(starterID)
	if err := database.WithContext(ctx).Table("hr_payroll_groups").Where("id = ?", groupID).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	calls := calculator.calls.Load()
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition || calculator.calls.Load() != calls {
		t.Fatalf("inactive payroll group must be rejected before invoking calculator: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_payroll_groups").Where("id = ?", groupID).Update("is_active", true).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_payroll_rule_sets (id, rule_code, rule_version, jurisdiction_code, effective_from,
		 artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id,
		 approved_by_principal_id, approved_at)
		 SELECT ?, ?, rule_version, jurisdiction_code, effective_from,
		 artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id,
		 approved_by_principal_id, NOW() + INTERVAL '1 day'
		 FROM hr_payroll_rule_sets WHERE id = ?`,
		futureRuleSetID, "PAY-RULE-"+futureRuleSetID, ruleSetID,
	).Error; err != nil {
		t.Fatalf("seed future rule approval: %v", err)
	}
	futureRuleInput := proto.CloneOf(draftInput)
	futureRuleInput.RuleSetId = &futureRuleSetID
	futureRuleInput.RunNumber = new(int32(99))
	futureKey := "future-rule-" + periodID
	futureRuleInput.IdempotencyKey = &futureKey
	var futureRunCount int64
	if created, err := client.CreatePayrollRun(ctx, &pb.CreatePayrollRunRequest{Input: futureRuleInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create payroll run before rule approval: %+v, %v", created, err)
	}
	if err := database.WithContext(ctx).Table("hr_payroll_runs").Where("idempotency_key = ?", futureKey).
		Count(&futureRunCount).Error; err != nil || futureRunCount != 0 {
		t.Fatalf("rejected creation persisted a run: count=%d, err=%v", futureRunCount, err)
	}
	if err := database.WithContext(ctx).Table("hr_payroll_runs").Where("id = ?", draftRunID).
		Update("rule_set_id", futureRuleSetID).Error; err != nil {
		t.Fatalf("bind draft run to future rule approval: %v", err)
	}
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition || calculator.calls.Load() != calls {
		t.Fatalf("future rule approval must be rejected before invoking calculator: %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_payroll_rule_sets (id, rule_code, rule_version, jurisdiction_code, effective_from, effective_to,
		 artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id,
		 approved_by_principal_id, approved_at)
		 SELECT ?, ?, rule_version, jurisdiction_code, effective_from, DATE '2026-11-29',
		 artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id,
		 approved_by_principal_id, approved_at FROM hr_payroll_rule_sets WHERE id = ?`,
		expiredRuleSetID, "PAY-RULE-"+expiredRuleSetID, ruleSetID,
	).Error; err != nil {
		t.Fatalf("seed rule ending before payroll period closes: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_payroll_runs").Where("id = ?", draftRunID).
		Update("rule_set_id", expiredRuleSetID).Error; err != nil {
		t.Fatalf("bind draft run to rule expiring during period: %v", err)
	}
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition || calculator.calls.Load() != calls {
		t.Fatalf("rule that does not cover payroll period must be rejected before invoking calculator: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_payroll_runs").Where("id = ?", draftRunID).
		Update("rule_set_id", ruleSetID).Error; err != nil {
		t.Fatalf("restore approved rule set for calculation: %v", err)
	}
	foundation.rejectWindow.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition || calculator.calls.Load() != calls {
		t.Fatalf("changed EF window must be rejected before executing rules: %v", err)
	}
	foundation.rejectWindow.Store(false)
	calculator.changedInputBytes.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition || calculator.calls.Load() != calls {
		t.Fatalf("changed snapshot must be rejected before executing rules: %v", err)
	}
	calculator.changedInputBytes.Store(false)
	calculator.changedRuleBytes.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition || calculator.calls.Load() != calls {
		t.Fatalf("changed rule artifact must be rejected before executing rules: %v", err)
	}
	calculator.changedRuleBytes.Store(false)
	calculator.mismatched.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("calculate with a mismatched snapshot digest: %v", err)
	}
	var resultCount int64
	if err := database.WithContext(ctx).Table("hr_payroll_results").Where("id = ?", fixtureResultID).Count(&resultCount).Error; err != nil || resultCount != 0 {
		t.Fatalf("rejected calculation persisted a result: count=%d err=%v", resultCount, err)
	}
	calculator.mismatched.Store(false)
	calculator.wrongRuleArtifact.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("calculate with a mismatched rule artifact digest: %v", err)
	}
	calculator.wrongRuleArtifact.Store(false)
	calculator.precisionMismatch.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("calculate with excess net pay precision: %v", err)
	}
	calculator.precisionMismatch.Store(false)
	calculator.prematurePublication.Store(true)
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("calculate with a premature payslip publication reference: %v", err)
	}
	calculator.prematurePublication.Store(false)
	if err := database.WithContext(ctx).Exec("UPDATE hr_pay_components SET component_name = 'New Base Salary' WHERE id = ?", compID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("calculate with stale component snapshot: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_pay_components SET component_name = 'Base Salary' WHERE id = ?", compID).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_profiles SET is_active = false WHERE id = ?", profileID).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("calculate with an inactive worker payroll profile: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_profiles SET is_active = true WHERE id = ?", profileID).Error; err != nil {
		t.Fatal(err)
	}
	calculated, err := client.CalculatePayrollRun(ctx, calculate)
	if err != nil || calculated.GetItem().GetStatus() != "calculated" ||
		!bytes.Equal(calculated.GetItem().GetOutputDigest(), bytes.Repeat([]byte{0x56}, 32)) {
		t.Fatalf("calculate approved snapshot: %+v, %v", calculated, err)
	}
	if err := database.WithContext(ctx).Table("hr_payroll_result_lines").
		Where("id = ? AND payroll_result_id = ?", fixtureLineID, fixtureResultID).Count(&resultCount).Error; err != nil || resultCount != 1 {
		t.Fatalf("calculation did not persist a line: count=%d err=%v", resultCount, err)
	}
	if _, err := client.CalculatePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat payroll calculation: %v", err)
	}

	cancelRequest := &pb.CancelPayrollRunRequest{Id: draftRunID, Reason: "Input snapshot requires correction"}
	if _, err := client.CancelPayrollRun(ctx, cancelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("cancel without review: %v", err)
	}
	if _, err := client.SubmitPayrollRunForReview(ctx, calculate); err != nil {
		t.Fatalf("submit second run for cancellation review: %v", err)
	}
	if _, err := client.CancelPayrollRun(ctx, cancelRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("initiator cancelled own run: %v", err)
	}
	policy.setPrincipal(approverID)
	if _, err := client.CancelPayrollRun(ctx, &pb.CancelPayrollRunRequest{Id: draftRunID, Reason: " "}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("cancel without reason: %v", err)
	}
	cancelled, err := client.CancelPayrollRun(ctx, cancelRequest)
	if err != nil || cancelled.GetItem().GetStatus() != "cancelled" ||
		cancelled.GetItem().GetCancellationReason() != cancelRequest.Reason ||
		cancelled.GetItem().GetCancelledByPrincipalId() != approverID || cancelled.GetItem().GetCancelledAt() == nil ||
		!bytes.Equal(cancelled.GetItem().GetOutputDigest(), calculated.GetItem().GetOutputDigest()) {
		t.Fatalf("cancel reviewed payroll: %+v, %v", cancelled, err)
	}
	if _, err := client.ApprovePayrollRun(ctx, calculate); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve cancelled payroll: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_result_lines SET amount = 1 WHERE id = ?", fixtureLineID).Error; err == nil {
		t.Fatal("database updated a cancelled payroll snapshot")
	}
	if _, err := client.CancelPayrollRun(ctx, cancelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("cancel already cancelled payroll: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_runs SET status = 'review' WHERE id = ?", draftRunID).Error; err == nil {
		t.Fatal("database reopened cancelled payroll")
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type = ?", draftRunID, "hr_payroll_runs.cancel").Count(&outboxCount).Error; err != nil || outboxCount != 1 {
		t.Fatalf("cancel outbox count=%d error=%v", outboxCount, err)
	}

	policy.setPrincipal(starterID)
	replacement := proto.CloneOf(draftInput)
	replacement.RunNumber = new(int32(3))
	replacement.IdempotencyKey = text("replacement-" + periodID)
	replacement.InputSnapshotReferenceId = &legalDocID
	replacementDigest := sha256.Sum256([]byte(replacementPayrollInputFixture))
	replacement.InputDigest = replacementDigest[:]
	calculator.replacementSnapshotID = legalDocID
	newRun, err := client.CreatePayrollRun(ctx, &pb.CreatePayrollRunRequest{Input: replacement})
	if err != nil || newRun.GetItem().GetStatus() != "draft" || newRun.GetItem().GetId() == draftRunID ||
		newRun.GetItem().GetInputSnapshotReferenceId() != legalDocID {
		t.Fatalf("create replacement on same payroll period: %+v, %v", newRun, err)
	}
	replacementRunID = newRun.GetItem().GetId()
	prior, err := client.GetPayrollRunByID(ctx, calculate)
	if err != nil || prior.GetItem().GetStatus() != "cancelled" ||
		!bytes.Equal(prior.GetItem().GetOutputDigest(), calculated.GetItem().GetOutputDigest()) {
		t.Fatalf("cancelled snapshot changed after creating a replacement: %+v, %v", prior, err)
	}

	replacementResultID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	replacementLineID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	calculator.result.ID = replacementResultID
	calculator.line.ID, calculator.line.PayrollResultID = replacementLineID, replacementResultID
	replacementRequest := &common.IDRequest{Id: replacementRunID}
	if _, err := client.CalculatePayrollRun(ctx, replacementRequest); err != nil {
		t.Fatalf("calculate replacement payroll: %v", err)
	}
	if _, err := client.SubmitPayrollRunForReview(ctx, replacementRequest); err != nil {
		t.Fatalf("submit replacement payroll: %v", err)
	}
	policy.setPrincipal(approverID)
	startDecision := make(chan struct{})
	decisions := make(chan error, 2)
	go func() {
		<-startDecision
		_, err := client.ApprovePayrollRun(ctx, replacementRequest)
		decisions <- err
	}()
	go func() {
		<-startDecision
		_, err := client.CancelPayrollRun(ctx, &pb.CancelPayrollRunRequest{Id: replacementRunID, Reason: "Review rejected fixture"})
		decisions <- err
	}()
	close(startDecision)
	decisionStatuses := make(map[codes.Code]int)
	for range 2 {
		decisionStatuses[status.Code(<-decisions)]++
	}
	if decisionStatuses[codes.OK] != 1 || decisionStatuses[codes.FailedPrecondition] != 1 {
		t.Fatalf("simultaneous payroll decisions returned %v", decisionStatuses)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type IN ?", replacementRunID, []string{"hr_payroll_runs.approve", "hr_payroll_runs.cancel"}).
		Count(&outboxCount).Error; err != nil || outboxCount != 1 {
		t.Fatalf("simultaneous payroll decisions emitted %d events: %v", outboxCount, err)
	}

}
