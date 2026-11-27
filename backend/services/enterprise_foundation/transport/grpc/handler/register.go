package handler

import (
	"log/slog"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	"google.golang.org/grpc"
)

// Dependencies menampung seluruh service aplikasi domain yang dibutuhkan oleh handler gRPC Enterprise Foundation.
type Dependencies struct {
	Logger   *slog.Logger
	Services *services.Registry
}

// RegisterAll mendaftarkan seluruh handler gRPC Enterprise Foundation ke ServiceRegistrar secara fail-closed.
func RegisterAll(registrar grpc.ServiceRegistrar, deps Dependencies) error {
	if registrar == nil {
		return eferrors.NewValidation("registrar", eferrors.Code("enterprise_foundation.handler.nil_registrar"), "grpc service registrar is required")
	}
	if deps.Services == nil {
		return eferrors.NewValidation("services", eferrors.Code("enterprise_foundation.handler.nil_services_registry"), "services registry is required")
	}
	if err := deps.Services.Validate(); err != nil {
		return err
	}

	errMapper := NewErrorMapper(deps.Logger)

	addressHandler, err := NewAddressHandler(deps.Services.Addresses, errMapper)
	if err != nil {
		return err
	}
	legalEntityHandler, err := NewLegalEntityHandler(deps.Services.LegalEntities, errMapper)
	if err != nil {
		return err
	}
	locationHandler, err := NewLocationHandler(deps.Services.Locations, errMapper)
	if err != nil {
		return err
	}
	organizationUnitHandler, err := NewOrganizationUnitHandler(deps.Services.OrganizationUnits, errMapper)
	if err != nil {
		return err
	}
	partyHandler, err := NewPartyHandler(deps.Services.Parties, errMapper)
	if err != nil {
		return err
	}

	efv1.RegisterAddressServiceServer(registrar, addressHandler)
	efv1.RegisterLegalEntityServiceServer(registrar, legalEntityHandler)
	efv1.RegisterLocationServiceServer(registrar, locationHandler)
	efv1.RegisterOrganizationUnitServiceServer(registrar, organizationUnitHandler)
	efv1.RegisterPartyServiceServer(registrar, partyHandler)

	return nil
}
