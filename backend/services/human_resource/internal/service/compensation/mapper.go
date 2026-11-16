// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createCompensationBand(in CreateCompensationBandInput) (*entity.CompensationBand, error) {
	v := &entity.CompensationBand{}
	v.BandCode = in.BandCode
	v.BandVersion = in.BandVersion
	v.JobLevelCode = in.JobLevelCode
	v.JurisdictionCode = in.JurisdictionCode
	v.CurrencyCode = in.CurrencyCode
	v.PayFrequency = in.PayFrequency
	v.MinimumAmount = in.MinimumAmount
	v.MidpointAmount = in.MidpointAmount
	v.MaximumAmount = in.MaximumAmount
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	v.ApprovalReferenceID = in.ApprovalReferenceID
	return v, nil
}
func patchCompensationBand(v *entity.CompensationBand, in UpdateCompensationBandInput) error {
	if in.BandCode.Set {
		if in.BandCode.Value == nil {
			return hrerrors.Invalid("band_code", "must not be null")
		}
		v.BandCode = *in.BandCode.Value
	}
	if in.BandVersion.Set {
		if in.BandVersion.Value == nil {
			return hrerrors.Invalid("band_version", "must not be null")
		}
		v.BandVersion = *in.BandVersion.Value
	}
	if in.JobLevelCode.Set {
		if in.JobLevelCode.Value == nil {
			return hrerrors.Invalid("job_level_code", "must not be null")
		}
		v.JobLevelCode = *in.JobLevelCode.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.PayFrequency.Set {
		if in.PayFrequency.Value == nil {
			return hrerrors.Invalid("pay_frequency", "must not be null")
		}
		v.PayFrequency = *in.PayFrequency.Value
	}
	if in.MinimumAmount.Set {
		if in.MinimumAmount.Value == nil {
			return hrerrors.Invalid("minimum_amount", "must not be null")
		}
		v.MinimumAmount = *in.MinimumAmount.Value
	}
	if in.MidpointAmount.Set {
		if in.MidpointAmount.Value == nil {
			return hrerrors.Invalid("midpoint_amount", "must not be null")
		}
		v.MidpointAmount = *in.MidpointAmount.Value
	}
	if in.MaximumAmount.Set {
		if in.MaximumAmount.Value == nil {
			return hrerrors.Invalid("maximum_amount", "must not be null")
		}
		v.MaximumAmount = *in.MaximumAmount.Value
	}
	if in.ValidFrom.Set {
		if in.ValidFrom.Value == nil {
			return hrerrors.Invalid("valid_from", "must not be null")
		}
		v.ValidFrom = *in.ValidFrom.Value
	}
	if in.ValidTo.Set {
		v.ValidTo = in.ValidTo.Value
	}
	if in.ApprovalReferenceID.Set {
		v.ApprovalReferenceID = in.ApprovalReferenceID.Value
	}
	return nil
}
func specCompensationBand() operation.Specification[entity.CompensationBand, string] {
	return operation.Specification[entity.CompensationBand, string]{
		Domain: "compensation", Table: "hr_compensation_bands",
		Key:         func(v *entity.CompensationBand) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.CompensationBand) string { return v.ID },
		Initialize: func(v *entity.CompensationBand, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.CompensationBand, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.CompensationBand) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			if v.ApprovalReferenceID != nil {
				a["approval_reference_id"] = fmt.Sprint(*v.ApprovalReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.CompensationBand) error {
			return nil
		},
	}
}
func createCompensationRecord(in CreateCompensationRecordInput) (*entity.CompensationRecord, error) {
	v := &entity.CompensationRecord{}
	v.WorkerID = in.WorkerID
	v.AssignmentID = in.AssignmentID
	v.ComponentID = in.ComponentID
	v.ComponentCodeSnapshot = in.ComponentCodeSnapshot
	v.CompensationBandID = in.CompensationBandID
	v.CompensationType = in.CompensationType
	v.Amount = in.Amount
	v.CurrencyCode = in.CurrencyCode
	v.PayFrequency = in.PayFrequency
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "proposed"
	}
	v.ReasonCode = in.ReasonCode
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.SourceChangeReferenceID = in.SourceChangeReferenceID
	if in.AccessClass != nil {
		v.AccessClass = *in.AccessClass
	} else {
		v.AccessClass = "restricted"
	}
	return v, nil
}
func patchCompensationRecord(v *entity.CompensationRecord, in UpdateCompensationRecordInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.AssignmentID.Set {
		if in.AssignmentID.Value == nil {
			return hrerrors.Invalid("assignment_id", "must not be null")
		}
		v.AssignmentID = *in.AssignmentID.Value
	}
	if in.ComponentID.Set {
		if in.ComponentID.Value == nil {
			return hrerrors.Invalid("component_id", "must not be null")
		}
		v.ComponentID = *in.ComponentID.Value
	}
	if in.ComponentCodeSnapshot.Set {
		if in.ComponentCodeSnapshot.Value == nil {
			return hrerrors.Invalid("component_code_snapshot", "must not be null")
		}
		v.ComponentCodeSnapshot = *in.ComponentCodeSnapshot.Value
	}
	if in.CompensationBandID.Set {
		v.CompensationBandID = in.CompensationBandID.Value
	}
	if in.CompensationType.Set {
		if in.CompensationType.Value == nil {
			return hrerrors.Invalid("compensation_type", "must not be null")
		}
		v.CompensationType = *in.CompensationType.Value
	}
	if in.Amount.Set {
		if in.Amount.Value == nil {
			return hrerrors.Invalid("amount", "must not be null")
		}
		v.Amount = *in.Amount.Value
	}
	if in.CurrencyCode.Set {
		if in.CurrencyCode.Value == nil {
			return hrerrors.Invalid("currency_code", "must not be null")
		}
		v.CurrencyCode = *in.CurrencyCode.Value
	}
	if in.PayFrequency.Set {
		if in.PayFrequency.Value == nil {
			return hrerrors.Invalid("pay_frequency", "must not be null")
		}
		v.PayFrequency = *in.PayFrequency.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.ReasonCode.Set {
		v.ReasonCode = in.ReasonCode.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.SourceChangeReferenceID.Set {
		v.SourceChangeReferenceID = in.SourceChangeReferenceID.Value
	}
	if in.AccessClass.Set {
		if in.AccessClass.Value == nil {
			return hrerrors.Invalid("access_class", "must not be null")
		}
		v.AccessClass = *in.AccessClass.Value
	}
	return nil
}
func specCompensationRecord() operation.Specification[entity.CompensationRecord, string] {
	return operation.Specification[entity.CompensationRecord, string]{
		Domain: "compensation", Table: "hr_compensation_records",
		Key:         func(v *entity.CompensationRecord) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.CompensationRecord) string { return v.ID },
		Initialize: func(v *entity.CompensationRecord, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.CompensationRecord, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.CompensationRecord) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["assignment_id"] = fmt.Sprint(v.AssignmentID)
			a["component_id"] = fmt.Sprint(v.ComponentID)
			if v.CompensationBandID != nil {
				a["compensation_band_id"] = fmt.Sprint(*v.CompensationBandID)
			}
			a["status"] = fmt.Sprint(v.Status)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			if v.SourceChangeReferenceID != nil {
				a["source_change_reference_id"] = fmt.Sprint(*v.SourceChangeReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.CompensationRecord) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "proposed" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.Status) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			if action != "create" && old.Status == "active" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
