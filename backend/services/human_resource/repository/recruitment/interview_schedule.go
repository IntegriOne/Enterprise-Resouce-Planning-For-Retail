package recruitment

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
	"gorm.io/gorm"
)

// HasCandidateScheduleOverlap memeriksa irisan jadwal aktif dalam transaksi pemanggil.
func (r *InterviewRepository) HasCandidateScheduleOverlap(ctx context.Context, candidateID, interviewID string, start, end time.Time) (bool, error) {
	db, err := r.query(ctx)
	if err != nil {
		return false, err
	}
	return hasInterviewScheduleOverlap(db.Where("hr_applications.candidate_id = ?", candidateID), interviewID, start, end)
}

// HasInterviewerScheduleOverlap memeriksa reservasi panelis pada wawancara lain.
func (r *InterviewRepository) HasInterviewerScheduleOverlap(ctx context.Context, principalID, interviewID string, start, end time.Time) (bool, error) {
	db, err := r.query(ctx)
	if err != nil {
		return false, err
	}
	db = db.Joins("JOIN hr_interview_panelists ON hr_interview_panelists.interview_id = hr_interviews.id").
		Where("hr_interview_panelists.interviewer_principal_id = ?", principalID)
	return hasInterviewScheduleOverlap(db, interviewID, start, end)
}

func hasInterviewScheduleOverlap(db *gorm.DB, interviewID string, start, end time.Time) (bool, error) {
	terminal := []entity.ApplicationStatus{
		entity.ApplicationStatusHired, entity.ApplicationStatusRejected, entity.ApplicationStatusWithdrawn,
	}
	var matches []string
	err := db.Joins("JOIN hr_applications ON hr_applications.id = hr_interviews.application_id").
		Where("hr_interviews.id <> ?", interviewID).
		Where("(hr_interviews.status IN ? OR (hr_interviews.status = ? AND hr_applications.status NOT IN ?))",
			[]string{"scheduled", "in_progress"}, "planned", terminal).
		Where("hr_interviews.scheduled_start < ? AND hr_interviews.scheduled_end > ?", end, start).
		Limit(1).Pluck("hr_interviews.id", &matches).Error
	return len(matches) > 0, pgstore.MapError(err)
}
