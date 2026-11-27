// Composition root ini dipelihara langsung bersama kontrak service setiap domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	accessrequestservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/accessrequest"
	analyticsservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/analytics"
	benefitsservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/benefits"
	careerservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	compensationservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/compensation"
	competencyservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	employeerelationsservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/employeerelations"
	engagementservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/engagement"
	hrserviceservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/hrservice"
	learningservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/learning"
	leaveservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	offboardingservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/offboarding"
	onboardingservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/onboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	payrollservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	performanceservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/performance"
	personnelservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	policyservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/policy"
	promotionservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/promotion"
	recruitmentservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	successionservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/succession"
	talentservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	workforceplanningservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// Registry menghimpun seluruh application service HRM.
type Registry struct {
	WorkforcePlanning *workforceplanningservice.Service
	Personnel         *personnelservice.Service
	Recruitment       *recruitmentservice.Service
	Onboarding        *onboardingservice.Service
	Compensation      *compensationservice.Service
	Benefits          *benefitsservice.Service
	Payroll           *payrollservice.Service
	Leave             *leaveservice.Service
	Performance       *performanceservice.Service
	Learning          *learningservice.Service
	Competency        *competencyservice.Service
	Succession        *successionservice.Service
	EmployeeRelations *employeerelationsservice.Service
	Analytics         *analyticsservice.Service
	Offboarding       *offboardingservice.Service
	Career            *careerservice.Service
	Policy            *policyservice.Service
	HRService         *hrserviceservice.Service
	Engagement        *engagementservice.Service
	AccessRequests    *accessrequestservice.Service
	Talent            *talentservice.Service
	Promotion         *promotionservice.Service
}

// RecruitmentVerifiers membawa adapter pemilik bukti untuk workflow rekrutmen.
type RecruitmentVerifiers struct {
	RequisitionApproval recruitmentservice.RequisitionApprovalVerifier
	RequisitionMandate  recruitmentservice.RequisitionMandateVerifier
	PostingDocument     recruitmentservice.PostingDocumentVerifier
	AssessmentEvidence  recruitmentservice.AssessmentEvidenceVerifier
	OfferCompensation   recruitmentservice.OfferCompensationVerifier
	OfferDocument       recruitmentservice.OfferDocumentVerifier
	OfferResponse       recruitmentservice.OfferResponseVerifier
}

// OffboardingVerifiers membawa adapter bukti tugas dan clearance offboarding.
type OffboardingVerifiers struct {
	Confirmation offboardingservice.ConfirmationVerifier
	Clearance    offboardingservice.ClearanceVerifier
}

type registryOptions struct {
	payrollSettlement   payrollservice.PayrollSettlementVerifier
	payrollPayment      payrollservice.PayrollPaymentVerifier
	payrollPosting      payrollservice.PayrollPostingVerifier
	contractSignature   personnelservice.ContractSignatureVerifier
	payslipPublication  payrollservice.PayslipPublicationProcessor
	payslipReissue      payrollservice.PayslipReissueVerifier
	payslipVerifier     payrollservice.PayslipVerifier
	principalLink       personnelservice.PrincipalLinkVerifier
	hrServiceResolution hrserviceservice.ResolutionVerifier
	onboardingEvidence  onboardingservice.EvidenceVerifier
	workforceBudget     workforceplanningservice.WorkforceBudgetVerifier
	offboarding         OffboardingVerifiers
	payrollCalculator   payrollservice.Calculator
	successionApproval  successionservice.ApprovalVerifier
	recruitment         RecruitmentVerifiers
	requisitionMandate  recruitmentservice.RequisitionMandateVerifier
}

// RegistryOption mengatur adapter tambahan tanpa mengubah repository atau policy.
type RegistryOption func(*registryOptions)

// WithWorkforceBudgetVerifier memasukkan verifier untuk domain WorkforcePlanning.
// Finance memiliki keputusan anggaran; cocokkan bukti dengan versi workforce plan.
func WithWorkforceBudgetVerifier(verifier workforceplanningservice.WorkforceBudgetVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.workforceBudget = verifier
	}
}

// WithOnboardingEvidenceVerifier memasukkan verifier untuk domain Onboarding.
// DRM memverifikasi versi bukti yang terikat pada tugas dan rencana onboarding.
func WithOnboardingEvidenceVerifier(verifier onboardingservice.EvidenceVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.onboardingEvidence = verifier
	}
}

// WithHRServiceResolutionVerifier memasukkan verifier untuk domain HRService.
// DRM memverifikasi versi bukti, permintaan, pekerja, serta akses principal.
func WithHRServiceResolutionVerifier(verifier hrserviceservice.ResolutionVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.hrServiceResolution = verifier
	}
}

// WithOffboardingVerifiers memasukkan adapter bukti yang tersedia saat startup.
// Finance memberikan clearance payroll; bukti tugas diverifikasi oleh pemilik tugas.
func WithOffboardingVerifiers(verifiers OffboardingVerifiers) RegistryOption {
	return func(options *registryOptions) {
		options.offboarding = verifiers
	}
}

// WithRequisitionMandateVerifier memasang client IAM untuk pemeriksaan grant peninjau.
func WithRequisitionMandateVerifier(verifier recruitmentservice.RequisitionMandateVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.requisitionMandate = verifier
	}
}

// WithRecruitmentVerifiers memasukkan adapter rekrutmen yang tersedia pada startup.
// Finance memverifikasi anggaran requisition. DRM memverifikasi versi bukti posting, asesmen, dan penawaran.
func WithRecruitmentVerifiers(verifiers RecruitmentVerifiers) RegistryOption {
	return func(options *registryOptions) {
		options.recruitment = verifiers
	}
}

// WithSuccessionApprovalVerifier memasukkan adapter pemeriksaan mandat panel suksesi.
func WithSuccessionApprovalVerifier(verifier successionservice.ApprovalVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.successionApproval = verifier
	}
}

// WithPayrollCalculator memasukkan adapter mesin kalkulasi payroll.
// Kalkulator memerlukan snapshot immutable dan artefak aturan DRM dengan digest yang cocok.
func WithPayrollCalculator(calculator payrollservice.Calculator) RegistryOption {
	return func(options *registryOptions) {
		options.payrollCalculator = calculator
	}
}

// WithPrincipalLinkVerifier memasukkan pemeriksa bukti pengaitan akun pekerja.
func WithPrincipalLinkVerifier(verifier personnelservice.PrincipalLinkVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.principalLink = verifier
	}
}

// WithPayslipVerifier memasang pemeriksa penerbitan Finance dan akses dokumen DRM.
// DRM memeriksa akses principal pemilik sebelum membuka versi slip yang sudah diterbitkan.
func WithPayslipVerifier(verifier payrollservice.PayslipVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.payslipVerifier = verifier
	}
}

// WithPayslipReissue memasang verifier untuk cetak ulang slip oleh petugas HR.
// Petugas HR dapat memulihkan blob yang hilang pada versi yang sama setelah DRM mencocokkan digest.
func WithPayslipReissue(verifier payrollservice.PayslipReissueVerifier) RegistryOption {
	return func(cfg *registryOptions) {
		cfg.payslipReissue = verifier
	}
}

// WithPayrollSettlementVerifier memasang pembaca bukti settlement run Finance.
func WithPayrollSettlementVerifier(verifier payrollservice.PayrollSettlementVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.payrollSettlement = verifier
	}
}

// WithPayrollPaymentVerifier memasang pembaca bukti pembayaran individual Finance.
func WithPayrollPaymentVerifier(verifier payrollservice.PayrollPaymentVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.payrollPayment = verifier
	}
}

// WithPayrollPostingVerifier memasang pembaca bukti jurnal Finance.
func WithPayrollPostingVerifier(verifier payrollservice.PayrollPostingVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.payrollPosting = verifier
	}
}

// WithPayslipPublication memasang adapter penerbitan dan penyimpanan PDF.
// Finance mengonfirmasi penerbitan. DRM menyimpan PDF privat secara idempoten setelah penerbitan.
func WithPayslipPublication(processor payrollservice.PayslipPublicationProcessor) RegistryOption {
	return func(options *registryOptions) {
		options.payslipPublication = processor
	}
}

// WithContractSignatureVerifier memasang verifier tanda tangan kontrak dari DRM.
// DRM memverifikasi tanda tangan dan ikatan versi immutable dengan kontrak serta pekerja.
func WithContractSignatureVerifier(verifier personnelservice.ContractSignatureVerifier) RegistryOption {
	return func(options *registryOptions) {
		options.contractSignature = verifier
	}
}

// NewRegistry mengikat service domain ke registry repository yang telah tersedia.
func NewRegistry(repos *repository.Registry, runtime operation.Dependencies, options ...RegistryOption) (*Registry, error) {
	if repos == nil {
		return nil, hrerrors.ErrUnavailable
	}
	var cfg registryOptions
	for _, apply := range options {
		if apply == nil {
			return nil, hrerrors.ErrInvalidInput
		}
		apply(&cfg)
	}
	runtime.Transactor = repos.Transactor
	runtime.Outbox = repos.OutboxEvent
	r := &Registry{}
	var err error
	r.WorkforcePlanning, err = workforceplanningservice.NewService(workforceplanningservice.Dependencies{
		Runtime: runtime, JobFamily: repos.JobFamily, Job: repos.Job,
		Position: repos.Position, WorkforcePlan: repos.WorkforcePlan,
		WorkforcePlanLine: repos.WorkforcePlanLine,
		BudgetVerifier:    cfg.workforceBudget,
	})
	if err != nil {
		return nil, err
	}
	r.Personnel, err = personnelservice.NewService(personnelservice.Dependencies{
		ContractSignatureVerifier: cfg.contractSignature,
		PrincipalLinkVerifier:     cfg.principalLink,
		WorkerAssignmentTransfer:  repos.WorkerAssignmentTransfer, TransferPosition: repos.Position,
		OffboardingPlan: repos.OffboardingPlan,
		Runtime:         runtime, Worker: repos.Worker, EmploymentContract: repos.EmploymentContract,
		WorkerAssignment: repos.WorkerAssignment, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
		WorkerDocument: repos.WorkerDocument, EmergencyContact: repos.EmergencyContact,
	})
	if err != nil {
		return nil, err
	}
	if cfg.requisitionMandate == nil {
		cfg.requisitionMandate = cfg.recruitment.RequisitionMandate
	}
	r.Recruitment, err = recruitmentservice.NewService(recruitmentservice.Dependencies{
		ApprovalVerifier:           cfg.recruitment.RequisitionApproval,
		MandateVerifier:            cfg.requisitionMandate,
		PostingVerifier:            cfg.recruitment.PostingDocument,
		AssessmentEvidenceVerifier: cfg.recruitment.AssessmentEvidence,
		OfferCompensationVerifier:  cfg.recruitment.OfferCompensation,
		OfferDocumentVerifier:      cfg.recruitment.OfferDocument,
		OfferResponseVerifier:      cfg.recruitment.OfferResponse,
		Runtime:                    runtime, Workers: repos.Worker, Requisition: repos.Requisition, Candidate: repos.Candidate,
		Application: repos.Application, ApplicationStageHistory: repos.ApplicationStageHistory,
		Interview: repos.Interview, InterviewPanelist: repos.InterviewPanelist,
		InterviewFeedback: repos.InterviewFeedback, Offer: repos.Offer,
		JobPosting: repos.JobPosting, ApplicationAssessment: repos.ApplicationAssessment,
	})
	if err != nil {
		return nil, err
	}
	r.Onboarding, err = onboardingservice.NewService(onboardingservice.Dependencies{
		Runtime: runtime, OnboardingPlan: repos.OnboardingPlan, OnboardingTask: repos.OnboardingTask,
		Evidence: cfg.onboardingEvidence,
	})
	if err != nil {
		return nil, err
	}
	r.Compensation, err = compensationservice.NewService(compensationservice.Dependencies{
		Runtime: runtime, CompensationBand: repos.CompensationBand, CompensationRecord: repos.CompensationRecord,
	})
	if err != nil {
		return nil, err
	}
	r.Benefits, err = benefitsservice.NewService(benefitsservice.Dependencies{
		Runtime: runtime, BenefitPlan: repos.BenefitPlan, BenefitEnrollment: repos.BenefitEnrollment,
	})
	if err != nil {
		return nil, err
	}
	r.Payroll, err = payrollservice.NewService(payrollservice.Dependencies{
		PayrollSettlement:  cfg.payrollSettlement,
		PayrollPayment:     cfg.payrollPayment,
		PayrollPosting:     cfg.payrollPosting,
		PayslipPublication: cfg.payslipPublication,
		PayslipReissue:     cfg.payslipReissue,
		Calculator:         cfg.payrollCalculator,
		Workers:            repos.Worker, PayslipVerifier: cfg.payslipVerifier,
		Runtime: runtime, PayrollProfile: repos.PayrollProfile, PayComponent: repos.PayComponent,
		PayrollPeriod: repos.PayrollPeriod, PayrollRun: repos.PayrollRun,
		PayrollResult: repos.PayrollResult, PayrollResultLine: repos.PayrollResultLine,
		PayrollAdjustment: repos.PayrollAdjustment, PayrollGroup: repos.PayrollGroup,
		PayrollRuleSet: repos.PayrollRuleSet,
	})
	if err != nil {
		return nil, err
	}
	r.Leave, err = leaveservice.NewService(leaveservice.Dependencies{
		Runtime: runtime, LeavePolicy: repos.LeavePolicy, LeaveEntitlement: repos.LeaveEntitlement,
		LeaveAdjustment: repos.LeaveAdjustment, LeaveLedgerEntry: repos.LeaveLedgerEntry,
	})
	if err != nil {
		return nil, err
	}
	r.Performance, err = performanceservice.NewService(performanceservice.Dependencies{
		Runtime: runtime, Worker: repos.Worker,
		PerformanceCycle: repos.PerformanceCycle, PerformanceGoal: repos.PerformanceGoal,
		PerformanceReview: repos.PerformanceReview, PerformanceGoalUpdate: repos.PerformanceGoalUpdate,
		PerformanceReviewGoalScore: repos.PerformanceReviewGoalScore,
	})
	if err != nil {
		return nil, err
	}
	r.Learning, err = learningservice.NewService(learningservice.Dependencies{
		Runtime: runtime, LearningCourse: repos.LearningCourse, LearningEnrollment: repos.LearningEnrollment,
		LearningCourseCompetency: repos.LearningCourseCompetency,
	})
	if err != nil {
		return nil, err
	}
	r.Competency, err = competencyservice.NewService(competencyservice.Dependencies{
		Runtime: runtime, WorkerQualification: repos.WorkerQualification, Competency: repos.Competency,
		JobCompetency: repos.JobCompetency, WorkerCompetencyAssessment: repos.WorkerCompetencyAssessment,
	})
	if err != nil {
		return nil, err
	}
	r.Succession, err = successionservice.NewService(successionservice.Dependencies{
		ApprovalVerifier: cfg.successionApproval, ApprovalWorkers: repos.Worker,
		Runtime: runtime, SuccessionPlan: repos.SuccessionPlan, SuccessionCandidate: repos.SuccessionCandidate,
	})
	if err != nil {
		return nil, err
	}
	r.EmployeeRelations, err = employeerelationsservice.NewService(employeerelationsservice.Dependencies{
		Runtime: runtime, EmployeeRelationsCase: repos.EmployeeRelationsCase,
		EmployeeRelationsEvent: repos.EmployeeRelationsEvent, CollectiveAgreement: repos.CollectiveAgreement,
	})
	if err != nil {
		return nil, err
	}
	r.Analytics, err = analyticsservice.NewService(analyticsservice.Dependencies{
		Runtime: runtime, HumanCapitalMetricDefinition: repos.HumanCapitalMetricDefinition,
		HumanCapitalMetricSnapshot: repos.HumanCapitalMetricSnapshot,
	})
	if err != nil {
		return nil, err
	}
	r.Offboarding, err = offboardingservice.NewService(offboardingservice.Dependencies{
		Runtime: runtime, OffboardingPlan: repos.OffboardingPlan, OffboardingTask: repos.OffboardingTask, PayrollResults: repos.PayrollResult, PayrollRuns: repos.PayrollRun,
		Workers:              repos.Worker,
		ConfirmationVerifier: cfg.offboarding.Confirmation, ClearanceVerifier: cfg.offboarding.Clearance,
	})
	if err != nil {
		return nil, err
	}
	r.Career, err = careerservice.NewService(careerservice.Dependencies{
		Runtime: runtime, Worker: repos.Worker, DevelopmentPlan: repos.DevelopmentPlan, DevelopmentAction: repos.DevelopmentAction,
		CareerPath: repos.CareerPath, CareerPathTransition: repos.CareerPathTransition,
		WorkerCareerAspiration: repos.WorkerCareerAspiration,
	})
	if err != nil {
		return nil, err
	}
	r.Policy, err = policyservice.NewService(policyservice.Dependencies{
		Runtime: runtime, Policy: repos.Policy, PolicyAcknowledgement: repos.PolicyAcknowledgement,
		Worker: repos.Worker,
	})
	if err != nil {
		return nil, err
	}
	r.HRService, err = hrserviceservice.NewService(hrserviceservice.Dependencies{
		Runtime: runtime, ServiceRequest: repos.ServiceRequest, ServiceRequestEvent: repos.ServiceRequestEvent,
		ResolutionVerifier: cfg.hrServiceResolution,
	})
	if err != nil {
		return nil, err
	}
	r.Engagement, err = engagementservice.NewService(engagementservice.Dependencies{
		Runtime: runtime, EngagementSurvey: repos.EngagementSurvey, EngagementSurveyResult: repos.EngagementSurveyResult,
	})
	if err != nil {
		return nil, err
	}
	r.AccessRequests, err = accessrequestservice.NewService(accessrequestservice.Dependencies{
		Runtime: runtime, AccessRequest: repos.AccessRequest, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
	})
	if err != nil {
		return nil, err
	}
	r.Talent, err = talentservice.NewService(talentservice.Dependencies{
		Runtime: runtime, PerformanceReviewReader: r.Performance,
		TalentFramework: repos.TalentFramework, TalentDimension: repos.TalentDimension,
		TalentReviewCycle: repos.TalentReviewCycle, TalentAssessment: repos.TalentAssessment,
		TalentAssessmentScore: repos.TalentAssessmentScore, TalentCalibrationSession: repos.TalentCalibrationSession,
		TalentCalibrationDecision: repos.TalentCalibrationDecision, TalentPool: repos.TalentPool,
		TalentPoolMembership: repos.TalentPoolMembership,
	})
	if err != nil {
		return nil, err
	}
	r.Promotion, err = promotionservice.NewService(promotionservice.Dependencies{
		Runtime: runtime, PromotionRecommendation: repos.PromotionRecommendation, PromotionDecision: repos.PromotionDecision,
		Worker: repos.Worker, WorkerAssignment: repos.WorkerAssignment, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
		Position: repos.Position, TalentAssessment: repos.TalentAssessment,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
