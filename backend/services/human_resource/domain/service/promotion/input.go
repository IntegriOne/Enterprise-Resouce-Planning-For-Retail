package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// ApplicationInput membawa bukti hasil perubahan hubungan kerja dalam transaksi yang sama.
// BusinessDate wajib berasal dari clock server, bukan input browser.
type ApplicationInput struct {
	Recommendation   *entity.PromotionRecommendation
	Decision         *entity.PromotionDecision
	Assessment       *entity.TalentAssessment
	Worker           *entity.Worker
	SourceAssignment *entity.WorkerAssignment
	NewAssignment    *entity.WorkerAssignment
	TargetPosition   *entity.Position
	Event            *entity.WorkerLifecycleEvent
	BusinessDate     time.Time
	AppliedAt        time.Time
}
