package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// EmployeeRelationsCaseRepository adalah port persistensi hr_employee_relations_cases.
type EmployeeRelationsCaseRepository interface {
	Repository[entity.EmployeeRelationsCase, string]
	InsertEmployeeRelationsCases(context.Context, *entity.EmployeeRelationsCase) error
	UpdateEmployeeRelationsCases(context.Context, *entity.EmployeeRelationsCase) error
	FindEmployeeRelationsCasesByUUID(context.Context, string) (*entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByUUIDForUpdate(context.Context, string) (*entity.EmployeeRelationsCase, error)
	DeleteEmployeeRelationsCasesByUUID(context.Context, string) error
	GetAllEmployeeRelationsCases(context.Context, int, int) ([]entity.EmployeeRelationsCase, error)
	CountEmployeeRelationsCases(context.Context) (int64, error)
	SelectEmployeeRelationsCasesWhere(context.Context, []Filter, int, int) ([]entity.EmployeeRelationsCase, error)
	CountEmployeeRelationsCasesWhere(context.Context, []Filter) (int64, error)
	FindEmployeeRelationsCasesByCaseNumber(context.Context, string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByCaseType(context.Context, string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByStatus(context.Context, entity.CaseStatus, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesBySubjectWorkerID(context.Context, *string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByReportingPartyReferenceID(context.Context, *string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByAssignedInvestigatorPrincipalID(context.Context, *string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByOutcomeCode(context.Context, *string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByRestrictedSummaryReferenceID(context.Context, string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByLegalHoldReferenceID(context.Context, *string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindEmployeeRelationsCasesByIdempotencyKey(context.Context, string, int, int) ([]entity.EmployeeRelationsCase, error)
	FindByCaseNumber(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByCaseType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByStatus(ctx context.Context, value entity.CaseStatus, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindBySubjectWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByReportingPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByAssignedInvestigatorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByOutcomeCode(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByRestrictedSummaryReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByLegalHoldReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error)
}

// EmployeeRelationsEventRepository adalah port persistensi hr_employee_relations_events.
type EmployeeRelationsEventRepository interface {
	Repository[entity.EmployeeRelationsEvent, int64]
	InsertEmployeeRelationsEvents(context.Context, *entity.EmployeeRelationsEvent) error
	UpdateEmployeeRelationsEvents(context.Context, *entity.EmployeeRelationsEvent) error
	FindEmployeeRelationsEventsByID(context.Context, int64) (*entity.EmployeeRelationsEvent, error)
	FindEmployeeRelationsEventsByIDForUpdate(context.Context, int64) (*entity.EmployeeRelationsEvent, error)
	DeleteEmployeeRelationsEventsByID(context.Context, int64) error
	GetAllEmployeeRelationsEvents(context.Context, int, int) ([]entity.EmployeeRelationsEvent, error)
	CountEmployeeRelationsEvents(context.Context) (int64, error)
	SelectEmployeeRelationsEventsWhere(context.Context, []Filter, int, int) ([]entity.EmployeeRelationsEvent, error)
	CountEmployeeRelationsEventsWhere(context.Context, []Filter) (int64, error)
	FindEmployeeRelationsEventsByCaseID(context.Context, string, int, int) ([]entity.EmployeeRelationsEvent, error)
	FindEmployeeRelationsEventsByEventType(context.Context, string, int, int) ([]entity.EmployeeRelationsEvent, error)
	FindEmployeeRelationsEventsByActorPrincipalID(context.Context, *string, int, int) ([]entity.EmployeeRelationsEvent, error)
	FindEmployeeRelationsEventsByEventReferenceID(context.Context, *string, int, int) ([]entity.EmployeeRelationsEvent, error)
	FindByCaseID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error)
	FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error)
	FindByActorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error)
	FindByEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error)
}

// CollectiveAgreementRepository adalah port persistensi hr_collective_agreements.
type CollectiveAgreementRepository interface {
	Repository[entity.CollectiveAgreement, string]
	InsertCollectiveAgreements(context.Context, *entity.CollectiveAgreement) error
	UpdateCollectiveAgreements(context.Context, *entity.CollectiveAgreement) error
	FindCollectiveAgreementsByUUID(context.Context, string) (*entity.CollectiveAgreement, error)
	FindCollectiveAgreementsByUUIDForUpdate(context.Context, string) (*entity.CollectiveAgreement, error)
	DeleteCollectiveAgreementsByUUID(context.Context, string) error
	GetAllCollectiveAgreements(context.Context, int, int) ([]entity.CollectiveAgreement, error)
	CountCollectiveAgreements(context.Context) (int64, error)
	SelectCollectiveAgreementsWhere(context.Context, []Filter, int, int) ([]entity.CollectiveAgreement, error)
	CountCollectiveAgreementsWhere(context.Context, []Filter) (int64, error)
	FindCollectiveAgreementsByAgreementCode(context.Context, string, int, int) ([]entity.CollectiveAgreement, error)
	FindCollectiveAgreementsByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.CollectiveAgreement, error)
	FindCollectiveAgreementsByRepresentativePartyReferenceID(context.Context, string, int, int) ([]entity.CollectiveAgreement, error)
	FindCollectiveAgreementsByJurisdictionCode(context.Context, string, int, int) ([]entity.CollectiveAgreement, error)
	FindCollectiveAgreementsByDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.CollectiveAgreement, error)
	FindByAgreementCode(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error)
	FindByRepresentativePartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error)
	FindByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error)
}
