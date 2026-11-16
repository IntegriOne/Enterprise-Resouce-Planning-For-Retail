// Package fixture menyediakan snapshot domain valid untuk tes tanpa database.
package fixture

import (
	"fmt"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

func ID(n int) string { return fmt.Sprintf("00000000-0000-0000-0000-%012d", n) }
func Date() time.Time { return time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC) }
func Worker() entity.Worker {
	return entity.Worker{ID: ID(1), WorkerNumber: "ANDI", PersonPartyReferenceID: ID(90), WorkerCategory: entity.WorkerCategoryEmployee, Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: ID(91), CreatedAt: Date(), UpdatedAt: Date()}
}
func Position() entity.Position {
	return entity.Position{ID: ID(2), PositionCode: "MGR-01", PositionTitle: "Manager", JobID: ID(92), OrganizationUnitReferenceID: ID(93), LegalEntityReferenceID: ID(91), PositionStatus: "active", HeadcountAuthorized: "1", ValidFrom: Date().AddDate(-1, 0, 0)}
}
func Assignment() entity.WorkerAssignment {
	return entity.WorkerAssignment{ID: ID(3), WorkerID: ID(1), PositionID: ID(2), AssignmentType: "primary", FTEPercent: "100", ValidFrom: Date().AddDate(0, -1, 0)}
}
func Entitlement() entity.LeaveEntitlement {
	return entity.LeaveEntitlement{ID: ID(4), WorkerID: ID(1), LeavePolicyID: ID(94), LeaveTypeCodeSnapshot: "ANNUAL", EntitlementPeriodStart: Date().AddDate(0, -1, 0), EntitlementPeriodEnd: Date().AddDate(1, 0, 0), OpeningBalance: "0", AccruedQuantity: "0", UsedQuantitySnapshot: "0", AdjustedQuantity: "0", ClosingBalanceSnapshot: "0", UOMCode: "day", Revision: 1, CalculatedAt: Date()}
}
func Entry(n int, kind string, quantity entity.Decimal) entity.LeaveLedgerEntry {
	return entity.LeaveLedgerEntry{ID: ID(100 + n), EntitlementID: ID(4), EntrySequence: int64(n), EntryType: kind, QuantityDelta: quantity, EffectiveDate: Date(), SourceManagement: "hrm", SourceEventID: ID(200 + n), SourceLineKey: "line", SourceDigest: []byte{1}, ReasonCode: "TEST", ActorPrincipalID: ID(95), RecordedAt: Date()}
}
func PayrollResult() entity.PayrollResult {
	return entity.PayrollResult{ID: ID(5), PayrollRunID: ID(6), PayrollGroupID: ID(7), WorkerID: ID(1), PayrollProfileReferenceID: ID(8), CurrencyCode: "IDR", GrossAmount: "100.123456", ReimbursementAmount: "2.000001", TotalDeductions: "10.000001", EmployerCostAmount: "105.123456", NetPayAmount: "92.123456", ResultStatus: "calculated", CalculationDigest: []byte{1}, CreatedAt: Date()}
}
func Line(n int, kind string, amount entity.Decimal) entity.PayrollResultLine {
	return entity.PayrollResultLine{ID: ID(300 + n), PayrollResultID: ID(5), ComponentID: ID(400 + n), ComponentCodeSnapshot: kind, ComponentNameSnapshot: kind, ComponentTypeSnapshot: kind, Amount: amount, LineSequence: int32(n)}
}
func Assessment() entity.TalentAssessment {
	return entity.TalentAssessment{ID: ID(9), WorkerID: ID(1), TalentReviewCycleID: ID(10), FrameworkID: ID(11), AssessmentVersion: 1, Status: "draft", AssessedByPrincipalID: ID(96), AssessedAt: Date(), EvidenceDocumentVersionReferenceID: ID(97), EvidenceDigest: []byte{1}}
}
func Cycle() entity.TalentReviewCycle {
	return entity.TalentReviewCycle{ID: ID(10), CycleCode: "2026-H2", FrameworkID: ID(11), LegalEntityReferenceID: ID(91), PeriodStart: Date().AddDate(0, -1, 0), PeriodEnd: Date().AddDate(0, 1, 0), Status: "open", OwnerPrincipalID: ID(96)}
}
func Framework() entity.TalentFramework {
	return entity.TalentFramework{ID: ID(11), FrameworkCode: "RETAIL", FrameworkVersion: 1, FrameworkName: "Retail", RubricDocumentVersionReferenceID: ID(97), RubricDigest: []byte{1}, EffectiveFrom: Date().AddDate(-1, 0, 0), ApprovedByPrincipalID: ID(96), ApprovedAt: Date().AddDate(-1, 0, 0)}
}
func Dimension() entity.TalentDimension {
	return entity.TalentDimension{ID: ID(12), FrameworkID: ID(11), DimensionCode: "LEADERSHIP", DimensionName: "Leadership", Definition: "Observed evidence", RatingAnchors: entity.JSON(`{"1":"basic","2":"guided","3":"independent","4":"lead","5":"develop"}`), WeightPercent: "100", IsRequired: true}
}
func Score() entity.TalentAssessmentScore {
	return entity.TalentAssessmentScore{ID: ID(13), AssessmentID: ID(9), FrameworkID: ID(11), DimensionID: ID(12), Score: 4, EvidenceDocumentVersionReferenceID: ID(97), EvidenceDigest: []byte{1}}
}
func Recommendation() entity.PromotionRecommendation {
	return entity.PromotionRecommendation{ID: ID(14), RecommendationNumber: "PROMO-1", WorkerID: ID(1), SourceAssignmentID: ID(3), TargetPositionID: ID(2), AssessmentID: ID(9), RecommendedEffectiveDate: Date(), RationaleDocumentVersionReferenceID: ID(97), RationaleDigest: []byte{1}, RecommendedByPrincipalID: ID(98), RecommendedAt: Date(), Status: "submitted", IdempotencyKey: "PROMO-1"}
}
func Decision() entity.PromotionDecision {
	return entity.PromotionDecision{ID: ID(15), RecommendationID: ID(14), StepNumber: 1, Decision: entity.ApprovalDecisionApprove, DecidedByPrincipalID: ID(99), DecidedAt: Date(), RationaleDocumentVersionReferenceID: ID(97), RationaleDigest: []byte{1}, IdempotencyKey: "DECISION-1"}
}
