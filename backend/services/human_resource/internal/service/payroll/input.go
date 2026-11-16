// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreatePayrollProfileInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollProfileInput struct {
	WorkerID                        string
	PayrollCountryCode              string
	PayrollJurisdictionCode         string
	PayrollGroupID                  string
	EmployingLegalEntityReferenceID string
	CurrencyCode                    string
	TaxProfileReferenceID           *string
	BankAccountTokenReferenceID     *string
	TaxStatusCode                   *string
	WithholdingElectionReferenceID  *string
	EffectiveFrom                   time.Time
	EffectiveTo                     *time.Time
	IsActive                        *bool
}

// UpdatePayrollProfileInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollProfileInput struct {
	WorkerID                        operation.Field[string]
	PayrollCountryCode              operation.Field[string]
	PayrollJurisdictionCode         operation.Field[string]
	PayrollGroupID                  operation.Field[string]
	EmployingLegalEntityReferenceID operation.Field[string]
	CurrencyCode                    operation.Field[string]
	TaxProfileReferenceID           operation.Field[string]
	BankAccountTokenReferenceID     operation.Field[string]
	TaxStatusCode                   operation.Field[string]
	WithholdingElectionReferenceID  operation.Field[string]
	EffectiveFrom                   operation.Field[time.Time]
	EffectiveTo                     operation.Field[time.Time]
	IsActive                        operation.Field[bool]
}

// CreatePayComponentInput adalah object input aplikasi, bukan entity ORM.
type CreatePayComponentInput struct {
	ComponentCode      string
	ComponentVersion   int32
	ComponentName      string
	ComponentType      string
	CalculationMethod  string
	TaxTreatmentCode   *string
	JurisdictionCode   string
	CurrencyCode       *string
	RuleSetID          *string
	FormulaReferenceID *string
	Taxable            *bool
	Pensionable        *bool
	EffectiveFrom      time.Time
	EffectiveTo        *time.Time
	IsActive           *bool
}

// UpdatePayComponentInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayComponentInput struct {
	ComponentCode      operation.Field[string]
	ComponentVersion   operation.Field[int32]
	ComponentName      operation.Field[string]
	ComponentType      operation.Field[string]
	CalculationMethod  operation.Field[string]
	TaxTreatmentCode   operation.Field[string]
	JurisdictionCode   operation.Field[string]
	CurrencyCode       operation.Field[string]
	RuleSetID          operation.Field[string]
	FormulaReferenceID operation.Field[string]
	Taxable            operation.Field[bool]
	Pensionable        operation.Field[bool]
	EffectiveFrom      operation.Field[time.Time]
	EffectiveTo        operation.Field[time.Time]
	IsActive           operation.Field[bool]
}

// CreatePayrollPeriodInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollPeriodInput struct {
	PayrollGroupID                  string
	EmployingLegalEntityReferenceID string
	CurrencyCode                    string
	JurisdictionCode                string
	PeriodCode                      string
	PeriodStart                     time.Time
	PeriodEnd                       time.Time
	PayDate                         time.Time
	PeriodStatus                    *string
	CalendarReferenceID             *string
}

// UpdatePayrollPeriodInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollPeriodInput struct {
	PayrollGroupID                  operation.Field[string]
	EmployingLegalEntityReferenceID operation.Field[string]
	CurrencyCode                    operation.Field[string]
	JurisdictionCode                operation.Field[string]
	PeriodCode                      operation.Field[string]
	PeriodStart                     operation.Field[time.Time]
	PeriodEnd                       operation.Field[time.Time]
	PayDate                         operation.Field[time.Time]
	PeriodStatus                    operation.Field[string]
	CalendarReferenceID             operation.Field[string]
}

// CreatePayrollRunInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollRunInput struct {
	PayrollPeriodID                   string
	PayrollGroupID                    string
	EmployingLegalEntityReferenceID   string
	CurrencyCode                      string
	RuleSetID                         string
	OriginalRunID                     *string
	SettlementDirection               *string
	RunNumber                         int32
	RunType                           string
	Status                            *entity.PayrollRunStatus
	JurisdictionCode                  string
	RuleEngineName                    string
	RuleEngineVersion                 string
	InputSnapshotReferenceID          string
	InputDigest                       []byte
	OutputDigest                      []byte
	CalculatedAt                      *time.Time
	ApprovedByPrincipalID             *string
	ApprovedAt                        *time.Time
	FinanceJournalReferenceID         *string
	PaymentBatchReferenceID           *string
	FinanceSettlementEventReferenceID *string
	FinanceSettledAt                  *time.Time
	IdempotencyKey                    string
}

// UpdatePayrollRunInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollRunInput struct {
	PayrollPeriodID                   operation.Field[string]
	PayrollGroupID                    operation.Field[string]
	EmployingLegalEntityReferenceID   operation.Field[string]
	CurrencyCode                      operation.Field[string]
	RuleSetID                         operation.Field[string]
	OriginalRunID                     operation.Field[string]
	SettlementDirection               operation.Field[string]
	RunNumber                         operation.Field[int32]
	RunType                           operation.Field[string]
	Status                            operation.Field[entity.PayrollRunStatus]
	JurisdictionCode                  operation.Field[string]
	RuleEngineName                    operation.Field[string]
	RuleEngineVersion                 operation.Field[string]
	InputSnapshotReferenceID          operation.Field[string]
	InputDigest                       operation.Field[[]byte]
	OutputDigest                      operation.Field[[]byte]
	CalculatedAt                      operation.Field[time.Time]
	ApprovedByPrincipalID             operation.Field[string]
	ApprovedAt                        operation.Field[time.Time]
	FinanceJournalReferenceID         operation.Field[string]
	PaymentBatchReferenceID           operation.Field[string]
	FinanceSettlementEventReferenceID operation.Field[string]
	FinanceSettledAt                  operation.Field[time.Time]
	IdempotencyKey                    operation.Field[string]
}

// CreatePayrollResultInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollResultInput struct {
	PayrollRunID               string
	PayrollGroupID             string
	WorkerID                   string
	PayrollProfileReferenceID  string
	CurrencyCode               string
	GrossAmount                entity.Decimal
	ReimbursementAmount        *entity.Decimal
	TotalDeductions            entity.Decimal
	EmployerCostAmount         entity.Decimal
	NetPayAmount               entity.Decimal
	ResultStatus               *string
	PaymentReferenceID         *string
	PayslipDocumentReferenceID *string
	CalculationDigest          []byte
}

// UpdatePayrollResultInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollResultInput struct {
	PayrollRunID               operation.Field[string]
	PayrollGroupID             operation.Field[string]
	WorkerID                   operation.Field[string]
	PayrollProfileReferenceID  operation.Field[string]
	CurrencyCode               operation.Field[string]
	GrossAmount                operation.Field[entity.Decimal]
	ReimbursementAmount        operation.Field[entity.Decimal]
	TotalDeductions            operation.Field[entity.Decimal]
	EmployerCostAmount         operation.Field[entity.Decimal]
	NetPayAmount               operation.Field[entity.Decimal]
	ResultStatus               operation.Field[string]
	PaymentReferenceID         operation.Field[string]
	PayslipDocumentReferenceID operation.Field[string]
	CalculationDigest          operation.Field[[]byte]
}

// CreatePayrollResultLineInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollResultLineInput struct {
	PayrollResultID       string
	ComponentID           string
	ComponentCodeSnapshot string
	ComponentNameSnapshot string
	ComponentTypeSnapshot string
	Amount                entity.Decimal
	Quantity              *entity.Decimal
	Rate                  *entity.Decimal
	TaxableBase           *entity.Decimal
	CalculationReference  *string
	LineSequence          int32
}

// UpdatePayrollResultLineInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollResultLineInput struct {
	PayrollResultID       operation.Field[string]
	ComponentID           operation.Field[string]
	ComponentCodeSnapshot operation.Field[string]
	ComponentNameSnapshot operation.Field[string]
	ComponentTypeSnapshot operation.Field[string]
	Amount                operation.Field[entity.Decimal]
	Quantity              operation.Field[entity.Decimal]
	Rate                  operation.Field[entity.Decimal]
	TaxableBase           operation.Field[entity.Decimal]
	CalculationReference  operation.Field[string]
	LineSequence          operation.Field[int32]
}

// CreatePayrollAdjustmentInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollAdjustmentInput struct {
	WorkerID              string
	PayrollPeriodID       string
	ComponentID           string
	AdjustmentType        string
	Amount                entity.Decimal
	CurrencyCode          string
	Reason                string
	SourceReference       *string
	ApprovedByPrincipalID *string
	ApprovedAt            *time.Time
	IdempotencyKey        string
}

// UpdatePayrollAdjustmentInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollAdjustmentInput struct {
	WorkerID              operation.Field[string]
	PayrollPeriodID       operation.Field[string]
	ComponentID           operation.Field[string]
	AdjustmentType        operation.Field[string]
	Amount                operation.Field[entity.Decimal]
	CurrencyCode          operation.Field[string]
	Reason                operation.Field[string]
	SourceReference       operation.Field[string]
	ApprovedByPrincipalID operation.Field[string]
	ApprovedAt            operation.Field[time.Time]
	IdempotencyKey        operation.Field[string]
}

// CreatePayrollGroupInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollGroupInput struct {
	GroupCode                       string
	GroupName                       string
	EmployingLegalEntityReferenceID string
	JurisdictionCode                string
	CurrencyCode                    string
	CurrencyScale                   int16
	PayFrequency                    string
	Timezone                        string
	IsActive                        *bool
}

// UpdatePayrollGroupInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollGroupInput struct {
	GroupCode                       operation.Field[string]
	GroupName                       operation.Field[string]
	EmployingLegalEntityReferenceID operation.Field[string]
	JurisdictionCode                operation.Field[string]
	CurrencyCode                    operation.Field[string]
	CurrencyScale                   operation.Field[int16]
	PayFrequency                    operation.Field[string]
	Timezone                        operation.Field[string]
	IsActive                        operation.Field[bool]
}

// CreatePayrollRuleSetInput adalah object input aplikasi, bukan entity ORM.
type CreatePayrollRuleSetInput struct {
	RuleCode                             string
	RuleVersion                          string
	JurisdictionCode                     string
	EffectiveFrom                        time.Time
	EffectiveTo                          *time.Time
	ArtifactDocumentVersionReferenceID   string
	ArtifactDigest                       []byte
	LegalBasisDocumentVersionReferenceID string
	ApprovedByPrincipalID                string
	ApprovedAt                           time.Time
}

// UpdatePayrollRuleSetInput membedakan field absen, null dan nilai eksplisit.
type UpdatePayrollRuleSetInput struct {
	RuleCode                             operation.Field[string]
	RuleVersion                          operation.Field[string]
	JurisdictionCode                     operation.Field[string]
	EffectiveFrom                        operation.Field[time.Time]
	EffectiveTo                          operation.Field[time.Time]
	ArtifactDocumentVersionReferenceID   operation.Field[string]
	ArtifactDigest                       operation.Field[[]byte]
	LegalBasisDocumentVersionReferenceID operation.Field[string]
	ApprovedByPrincipalID                operation.Field[string]
	ApprovedAt                           operation.Field[time.Time]
}
