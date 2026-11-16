// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateSuccessionPlanInput adalah object input aplikasi, bukan entity ORM.
type CreateSuccessionPlanInput struct {
	PositionID            string
	PlanCode              string
	PlanVersion           int32
	ReviewDate            time.Time
	ConfidentialityClass  *string
	ApprovedByPrincipalID *string
	ApprovedAt            *time.Time
	SuccessionRisk        *string
}

// UpdateSuccessionPlanInput membedakan field absen, null dan nilai eksplisit.
type UpdateSuccessionPlanInput struct {
	PositionID            operation.Field[string]
	PlanCode              operation.Field[string]
	PlanVersion           operation.Field[int32]
	ReviewDate            operation.Field[time.Time]
	ConfidentialityClass  operation.Field[string]
	ApprovedByPrincipalID operation.Field[string]
	ApprovedAt            operation.Field[time.Time]
	SuccessionRisk        operation.Field[string]
}

// CreateSuccessionCandidateInput adalah object input aplikasi, bukan entity ORM.
type CreateSuccessionCandidateInput struct {
	SuccessionPlanID    string
	CandidateWorkerID   string
	ReadinessCode       string
	ReadinessAssessedAt time.Time
	ReadinessRationale  *string
	TalentAssessmentID  *string
	DevelopmentPlanID   *string
	RiskOfLoss          *string
}

// UpdateSuccessionCandidateInput membedakan field absen, null dan nilai eksplisit.
type UpdateSuccessionCandidateInput struct {
	SuccessionPlanID    operation.Field[string]
	CandidateWorkerID   operation.Field[string]
	ReadinessCode       operation.Field[string]
	ReadinessAssessedAt operation.Field[time.Time]
	ReadinessRationale  operation.Field[string]
	TalentAssessmentID  operation.Field[string]
	DevelopmentPlanID   operation.Field[string]
	RiskOfLoss          operation.Field[string]
}
