package security

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type contractWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type contractNumberReader interface {
	FindByContractNumber(context.Context, string, int, int) ([]entity.EmploymentContract, error)
}

type employmentContractValidator struct {
	workers   contractWorkerReader
	contracts contractNumberReader
}

// PersonnelValidators mengaktifkan aturan kontrak draf yang hanya memakai referensi pekerja HRM.
func PersonnelValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.EmploymentContract) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_employment_contracts": &employmentContractValidator{
			workers: repos.Worker, contracts: repos.EmploymentContract,
		},
	}, nil
}

// ValidateMutation membatasi CRUD ke kontrak draf tanpa bukti tanda tangan.
func (v *employmentContractValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_employment_contracts" ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.contracts) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.EmploymentContract
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.EmploymentContract)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if !draftUnsignedContract(previous) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	contract := previous
	if after != nil {
		var ok bool
		contract, ok = after.(*entity.EmploymentContract)
		if !ok || contract == nil {
			return hrerrors.ErrInvalidInput
		}
		if !draftUnsignedContract(contract) || !supportedContractType(contract.ContractType) {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != contract.ID || previous.WorkerID != contract.WorkerID ||
			previous.EmployingLegalEntityReferenceID != contract.EmployingLegalEntityReferenceID ||
			!previous.CreatedAt.Equal(contract.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous == nil {
		return hrerrors.ErrInvalidInput
	}
	if contract.CollectiveAgreementID != nil {
		return hrerrors.ErrUnavailable
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, contract.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || !strings.EqualFold(worker.PrimaryLegalEntityReferenceID, contract.EmployingLegalEntityReferenceID) {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if worker.Status != entity.WorkerStatusPreboarding && worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	matches, err := v.contracts.FindByContractNumber(ctx, contract.ContractNumber, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != contract.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func draftUnsignedContract(contract *entity.EmploymentContract) bool {
	return contract != nil && contract.ContractStatus == "draft" && contract.SignedAt == nil &&
		contract.SignedDocumentReferenceID == nil && contract.SignedDocumentVersionReferenceID == nil &&
		contract.SignedContentDigest == nil
}

func supportedContractType(value string) bool {
	switch value {
	case "permanent", "fixed_term", "hourly", "contractor", "apprenticeship", "collective_agreement", "other":
		return true
	default:
		return false
	}
}
