package iam

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
)

// Publish mengirim event terminasi ke IAM dan memeriksa korelasi konfirmasinya.
func (c *RequisitionMandateClient) Publish(ctx context.Context, event *entity.OutboxEvent) error {
	if c == nil || c.termination == nil {
		return hrerrors.ErrUnavailable
	}
	if event == nil || event.EventType != entity.WorkerTerminationRecorded ||
		event.AggregateType != "hr_worker_terminations" || event.SchemaVersion != 1 {
		return hrerrors.ErrInvalidInput
	}
	var data struct {
		WorkerID string `json:"worker_id"`
		PlanID   string `json:"offboarding_plan_id"`
		Date     string `json:"termination_effective_date"`
	}
	if err := json.Unmarshal(event.Payload, &data); err != nil {
		return hrerrors.ErrInvalidInput
	}
	date, err := time.Parse(time.DateOnly, data.Date)
	if err != nil || date.IsZero() || !strings.EqualFold(data.PlanID, event.AggregateID) ||
		entity.ValidateUUID("worker_id", data.WorkerID) != nil ||
		entity.ValidateUUID("offboarding_plan_id", data.PlanID) != nil {
		return hrerrors.ErrInvalidInput
	}
	result, err := c.termination.ApplyEmployeeTermination(ctx, &iamv1.ApplyEmployeeTerminationRequest{
		WorkerId: data.WorkerID, OffboardingPlanId: data.PlanID, EffectiveDate: data.Date,
	})
	if err != nil {
		return err
	}
	if !strings.EqualFold(result.GetWorkerId(), data.WorkerID) || !strings.EqualFold(result.GetOffboardingPlanId(), data.PlanID) || result.GetEffectiveDate() != data.Date {
		return hrerrors.ErrUnavailable
	}
	switch result.GetOutcome() {
	case iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_ACCOUNT_DISABLED,
		iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_ACCOUNT_RETAINED,
		iamv1.EmployeeTerminationOutcome_EMPLOYEE_TERMINATION_OUTCOME_NO_LINKED_ACCOUNT:
		return nil
	default:
		return hrerrors.ErrUnavailable
	}
}
