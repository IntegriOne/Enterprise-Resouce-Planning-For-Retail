package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApproveWorkerAssignment mengesahkan permintaan tanpa mengubah penempatan primer saat ini.
func (s *Service) ApproveWorkerAssignment(ctx context.Context, assignmentID string) (*entity.WorkerAssignment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.Command(ctx, assignmentID, "approve", func(_ context.Context, assignment *entity.WorkerAssignment, actor operation.Actor, now time.Time) (*entity.WorkerAssignment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if assignment.RequestedByPrincipalID == nil || assignment.RequestedAt == nil ||
			assignment.ApprovedByPrincipalID != nil || assignment.ApprovedAt != nil ||
			assignment.EmploymentStatusSnapshot != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if strings.EqualFold(*assignment.RequestedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if now.Before(*assignment.RequestedAt) {
			return nil, hrerrors.ErrReferenceConflict
		}
		status := string(entity.WorkerStatusActive)
		assignment.EmploymentStatusSnapshot = &status
		assignment.ApprovedByPrincipalID = &actor.PrincipalID
		assignment.ApprovedAt = &now
		return assignment, nil
	})
}
