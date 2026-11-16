// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateBenefitPlanInput adalah object input aplikasi, bukan entity ORM.
type CreateBenefitPlanInput struct {
	PlanCode                            string
	PlanVersion                         int32
	PlanName                            string
	BenefitType                         string
	ProviderPartyReferenceID            *string
	JurisdictionCode                    string
	EligibilityRule                     entity.JSON
	EmployeeContributionRuleReferenceID *string
	EmployerContributionRuleReferenceID *string
	ValidFrom                           time.Time
	ValidTo                             *time.Time
	IsActive                            *bool
}

// UpdateBenefitPlanInput membedakan field absen, null dan nilai eksplisit.
type UpdateBenefitPlanInput struct {
	PlanCode                            operation.Field[string]
	PlanVersion                         operation.Field[int32]
	PlanName                            operation.Field[string]
	BenefitType                         operation.Field[string]
	ProviderPartyReferenceID            operation.Field[string]
	JurisdictionCode                    operation.Field[string]
	EligibilityRule                     operation.Field[entity.JSON]
	EmployeeContributionRuleReferenceID operation.Field[string]
	EmployerContributionRuleReferenceID operation.Field[string]
	ValidFrom                           operation.Field[time.Time]
	ValidTo                             operation.Field[time.Time]
	IsActive                            operation.Field[bool]
}

// CreateBenefitEnrollmentInput adalah object input aplikasi, bukan entity ORM.
type CreateBenefitEnrollmentInput struct {
	WorkerID                   string
	BenefitPlanID              string
	EnrollmentStatus           *string
	EffectiveFrom              time.Time
	EffectiveTo                *time.Time
	EnrollmentReference        *string
	ProviderMemberReference    *string
	EmployeeContributionAmount *entity.Decimal
	EmployerContributionAmount *entity.Decimal
	CurrencyCode               *string
	ElectedAt                  *time.Time
	VerifiedAt                 *time.Time
}

// UpdateBenefitEnrollmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateBenefitEnrollmentInput struct {
	WorkerID                   operation.Field[string]
	BenefitPlanID              operation.Field[string]
	EnrollmentStatus           operation.Field[string]
	EffectiveFrom              operation.Field[time.Time]
	EffectiveTo                operation.Field[time.Time]
	EnrollmentReference        operation.Field[string]
	ProviderMemberReference    operation.Field[string]
	EmployeeContributionAmount operation.Field[entity.Decimal]
	EmployerContributionAmount operation.Field[entity.Decimal]
	CurrencyCode               operation.Field[string]
	ElectedAt                  operation.Field[time.Time]
	VerifiedAt                 operation.Field[time.Time]
}
