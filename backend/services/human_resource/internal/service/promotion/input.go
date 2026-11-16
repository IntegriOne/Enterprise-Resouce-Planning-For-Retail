// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreatePromotionRecommendationInput adalah object input aplikasi, bukan entity ORM.
type CreatePromotionRecommendationInput struct {
	RecommendationNumber                string
	WorkerID                            string
	SourceAssignmentID                  string
	TargetPositionID                    string
	AssessmentID                        string
	SupersedesRecommendationID          *string
	RecommendedEffectiveDate            time.Time
	RationaleDocumentVersionReferenceID string
	RationaleDigest                     []byte
	Status                              *string
	ApprovalDecisionID                  *string
	AppliedAssignmentID                 *string
	AppliedLifecycleEventID             *int64
	AppliedAt                           *time.Time
	IdempotencyKey                      string
}

// UpdatePromotionRecommendationInput membedakan field absen, null dan nilai eksplisit.
type UpdatePromotionRecommendationInput struct {
	RecommendationNumber                operation.Field[string]
	WorkerID                            operation.Field[string]
	SourceAssignmentID                  operation.Field[string]
	TargetPositionID                    operation.Field[string]
	AssessmentID                        operation.Field[string]
	SupersedesRecommendationID          operation.Field[string]
	RecommendedEffectiveDate            operation.Field[time.Time]
	RationaleDocumentVersionReferenceID operation.Field[string]
	RationaleDigest                     operation.Field[[]byte]
	Status                              operation.Field[string]
	ApprovalDecisionID                  operation.Field[string]
	AppliedAssignmentID                 operation.Field[string]
	AppliedLifecycleEventID             operation.Field[int64]
	AppliedAt                           operation.Field[time.Time]
	IdempotencyKey                      operation.Field[string]
}

// CreatePromotionDecisionInput adalah object input aplikasi, bukan entity ORM.
type CreatePromotionDecisionInput struct {
	RecommendationID                    string
	StepNumber                          int32
	Decision                            entity.ApprovalDecision
	RationaleDocumentVersionReferenceID string
	RationaleDigest                     []byte
	IdempotencyKey                      string
}

// UpdatePromotionDecisionInput membedakan field absen, null dan nilai eksplisit.
type UpdatePromotionDecisionInput struct {
	RecommendationID                    operation.Field[string]
	StepNumber                          operation.Field[int32]
	Decision                            operation.Field[entity.ApprovalDecision]
	RationaleDocumentVersionReferenceID operation.Field[string]
	RationaleDigest                     operation.Field[[]byte]
	IdempotencyKey                      operation.Field[string]
}
