package security

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// FoundationVerifier memvalidasi referensi entitas legal, unit, lokasi, dan party ke pemilik Enterprise Foundation.
type FoundationVerifier interface {
	PositionReferenceVerifier
	WorkerReferenceVerifier
	VerifyLegalEntity(ctx context.Context, legalEntityID string) error
	VerifyOrganizationUnit(ctx context.Context, organizationUnitID, legalEntityID string) error
	VerifyLocation(ctx context.Context, locationID, legalEntityID string) error
	VerifyParty(ctx context.Context, partyID string, requireIndividual bool) error
}

func verifyFoundationLegalEntity(ctx context.Context, verifier FoundationVerifier, legalEntityID string) error {
	if operation.RequireDependency(verifier) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("legal_entity_reference_id", legalEntityID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	return verifier.VerifyLegalEntity(ctx, legalEntityID)
}

func verifyFoundationOrganizationUnit(ctx context.Context, verifier FoundationVerifier, organizationUnitID, legalEntityID string) error {
	if operation.RequireDependency(verifier) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("organization_unit_reference_id", organizationUnitID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	if legalEntityID != "" {
		if err := entity.ValidateUUID("legal_entity_reference_id", legalEntityID); err != nil {
			return hrerrors.ErrReferenceConflict
		}
	}
	return verifier.VerifyOrganizationUnit(ctx, organizationUnitID, legalEntityID)
}

func verifyFoundationParty(ctx context.Context, verifier FoundationVerifier, partyID string, requireIndividual bool) error {
	if operation.RequireDependency(verifier) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("party_reference_id", partyID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	return verifier.VerifyParty(ctx, partyID, requireIndividual)
}
