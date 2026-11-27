package service

import (
	"context"
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type IAMClearanceRequest struct {
	PlanID, WorkerID, WorkerPrincipalID string
	TerminationEffectiveDate            time.Time
}

type FinanceClearanceRequest struct {
	PlanID, WorkerID, FinalPayrollRunID, FinanceClearanceReferenceID string
	FinalPayrollResultID, EmployingLegalEntityID, CurrencyCode       string
	PaymentReferenceID                                               string
	RunOutputDigest, ResultCalculationDigest                         []byte
	TerminationEffectiveDate                                         time.Time
}

type IAMClearanceVerifier interface {
	VerifyIAMClearance(context.Context, IAMClearanceRequest) error
}

type FinanceClearanceVerifier interface {
	VerifyFinanceClearance(context.Context, FinanceClearanceRequest) error
}

type ownerClearance struct {
	iam     IAMClearanceVerifier
	finance FinanceClearanceVerifier
}

// NewOwnerClearance menggabungkan keputusan clearance dari dua pemilik data.
func NewOwnerClearance(iam IAMClearanceVerifier, finance FinanceClearanceVerifier) (ClearanceVerifier, error) {
	if err := operation.RequireDependency(iam); err != nil {
		return nil, err
	}
	if err := operation.RequireDependency(finance); err != nil {
		return nil, err
	}
	return ownerClearance{iam: iam, finance: finance}, nil
}

func (v ownerClearance) VerifyOffboardingClearance(ctx context.Context, req ClearanceRequest) error {
	if err := v.iam.VerifyIAMClearance(ctx, IAMClearanceRequest{
		PlanID: req.PlanID, WorkerID: req.WorkerID, WorkerPrincipalID: req.WorkerPrincipalID,
		TerminationEffectiveDate: req.TerminationEffectiveDate,
	}); err != nil {
		return err
	}
	return v.finance.VerifyFinanceClearance(ctx, FinanceClearanceRequest{
		PlanID: req.PlanID, WorkerID: req.WorkerID, FinalPayrollRunID: req.FinalPayrollRunID,
		FinanceClearanceReferenceID: req.FinanceClearanceReferenceID, FinalPayrollResultID: req.FinalPayrollResultID,
		EmployingLegalEntityID: req.EmployingLegalEntityID, CurrencyCode: req.CurrencyCode,
		PaymentReferenceID: req.PaymentReferenceID, RunOutputDigest: req.RunOutputDigest,
		ResultCalculationDigest: req.ResultCalculationDigest, TerminationEffectiveDate: req.TerminationEffectiveDate,
	})
}
