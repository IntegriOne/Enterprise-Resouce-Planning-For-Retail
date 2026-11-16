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

// NewRegistry mengikat service domain ke registry repository yang telah tersedia.
func NewRegistry(repos *repository.Registry, runtime operation.Dependencies) (*Registry, error) {
	if repos == nil {
		return nil, hrerrors.ErrUnavailable
	}
	runtime.Transactor = repos.Transactor
	runtime.Outbox = repos.OutboxEvent
	r := &Registry{}
	var err error
	r.WorkforcePlanning, err = workforceplanningservice.NewService(workforceplanningservice.Dependencies{
		Runtime: runtime, JobFamily: repos.JobFamily, Job: repos.Job,
		Position: repos.Position, WorkforcePlan: repos.WorkforcePlan,
		WorkforcePlanLine: repos.WorkforcePlanLine,
	})
	if err != nil {
		return nil, err
	}
	r.Personnel, err = personnelservice.NewService(personnelservice.Dependencies{
		Runtime: runtime, Worker: repos.Worker, EmploymentContract: repos.EmploymentContract,
		WorkerAssignment: repos.WorkerAssignment, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
		WorkerDocument: repos.WorkerDocument, EmergencyContact: repos.EmergencyContact,
	})
	if err != nil {
		return nil, err
	}
	r.Recruitment, err = recruitmentservice.NewService(recruitmentservice.Dependencies{
		Runtime: runtime, Requisition: repos.Requisition, Candidate: repos.Candidate,
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
		Runtime: runtime, PerformanceCycle: repos.PerformanceCycle, PerformanceGoal: repos.PerformanceGoal,
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
		Runtime: runtime, OffboardingPlan: repos.OffboardingPlan, OffboardingTask: repos.OffboardingTask,
	})
	if err != nil {
		return nil, err
	}
	r.Career, err = careerservice.NewService(careerservice.Dependencies{
		Runtime: runtime, DevelopmentPlan: repos.DevelopmentPlan, DevelopmentAction: repos.DevelopmentAction,
		CareerPath: repos.CareerPath, CareerPathTransition: repos.CareerPathTransition,
		WorkerCareerAspiration: repos.WorkerCareerAspiration,
	})
	if err != nil {
		return nil, err
	}
	r.Policy, err = policyservice.NewService(policyservice.Dependencies{
		Runtime: runtime, Policy: repos.Policy, PolicyAcknowledgement: repos.PolicyAcknowledgement,
	})
	if err != nil {
		return nil, err
	}
	r.HRService, err = hrserviceservice.NewService(hrserviceservice.Dependencies{
		Runtime: runtime, ServiceRequest: repos.ServiceRequest, ServiceRequestEvent: repos.ServiceRequestEvent,
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
		Runtime: runtime, AccessRequest: repos.AccessRequest,
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
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
