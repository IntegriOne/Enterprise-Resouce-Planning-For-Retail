package resolvers

import (
	"github.com/integrione/erp-retail/backend/gateway/graph/generated"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	efadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/enterprise_foundation"
	hrmadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/human_resource"
	iamadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/iam"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

// Resolver menyimpan dependensi adapter domain yang melayani eksekusi GraphQL Gateway.
type Resolver struct {
	IAMAdapter *iamadapter.Adapter
	EFAdapter  *efadapter.Adapter
	HRMAdapter *hrmadapter.Adapter
}

func (r *Resolver) iam() (*iamadapter.Adapter, error) {
	if r == nil || r.IAMAdapter == nil {
		return nil, middleware.ErrUnavailable("IAM adapter is not initialized", nil)
	}
	return r.IAMAdapter, nil
}

func (r *Resolver) ef() (*efadapter.Adapter, error) {
	if r == nil || r.EFAdapter == nil {
		return nil, middleware.ErrUnavailable("enterprise foundation adapter is not initialized", nil)
	}
	return r.EFAdapter, nil
}

// NewComplexityRoot membentuk aturan penghitungan kompleksitas kueri Relay terpaginasi untuk batas FixedComplexityLimit.
func NewComplexityRoot(maxComplexity int) generated.ComplexityRoot {
	var c generated.ComplexityRoot

	// Satu operasi boleh membaca dan menyajikan paling banyak satu PDF berukuran besar.
	pdfCost := maxComplexity/2 + 1
	c.Query.MyPayslipPDF = func(childComplexity int, _ string) int {
		return max(pdfCost, childComplexity)
	}
	c.Mutation.ReissueEmployeePayslip = func(childComplexity int, _, _ string) int {
		return max(pdfCost, childComplexity)
	}
	c.MyPayslipPDF.ContentBase64 = func(_ int) int {
		return pdfCost
	}

	paginated := func(childComplexity int, first, last *int) int {
		limit := 20
		if first != nil && *first > 0 {
			limit = *first
		} else if last != nil && *last > 0 {
			limit = *last
		}
		if limit > 100 {
			// Menolak permintaan limit halaman berlebih langsung pada lapis kompleksitas GraphQL.
			return limit * 10
		}
		factor := (limit + 3) / 4
		if factor < 1 {
			factor = 1
		}
		if childComplexity < 1 {
			childComplexity = 1
		}
		return factor * childComplexity
	}

	c.Query.OffboardingPlans = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.OffboardingTasks = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Jobs = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.JobsByIds = func(childComplexity int, ids []string) int {
		size := len(ids)
		return paginated(childComplexity, &size, nil)
	}
	c.Query.JobFamilies = func(childComplexity int, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Positions = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LeavePolicies = func(childComplexity int, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LeaveEntitlements = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LeaveLedgerEntries = func(childComplexity int, _ *string, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LeaveAdjustments = func(childComplexity int, _ *string, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PayrollResults = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PayrollResultLines = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Candidates = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Requisitions = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PayrollRuns = func(childComplexity int, _ *string, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PayrollPeriods = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PayrollGroups = func(childComplexity int, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PayrollRuleSets = func(childComplexity int, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.WorkforcePlans = func(childComplexity int, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.WorkforcePlanLines = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Workers = func(childComplexity int, _ *string, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.WorkerAssignments = func(childComplexity int, _ *string, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.EmploymentContracts = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.MyPayslips = func(childComplexity int, _ *string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.WorkerAssignmentTransfers = func(childComplexity int, _ string, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.EmployeeProvisioningRequests = func(childComplexity int, _, _ *string, first, last *int) int {
		return paginated(childComplexity, first, last)
	}

	c.Query.Users = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *model.UserKind, _ *model.UserStatus, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserIdentities = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserRoles = func(childComplexity int, _ string, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserEffectiveRoles = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserGroups = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.UserSessions = func(childComplexity int, _ string, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.RolePermissions = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Permissions = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Groups = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.GroupMembers = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.GroupRoles = func(childComplexity int, _ string, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.MySessions = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}

	c.User.Identities = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.User.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.User.EffectiveRoles = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.User.Groups = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.User.Sessions = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Role.Permissions = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Group.Members = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Group.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *bool) int {
		return paginated(childComplexity, first, last)
	}

	// Kompleksitas kueri dan relasi terpaginasi Enterprise Foundation.
	c.Query.Addresses = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *string, _ *string, _ *string) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllAddresses = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LegalEntitiesByNib = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LegalEntities = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *model.LegalEntityStatus, _ *bool, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllLegalEntities = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LegalEntityChildren = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LocationsByLegalEntityAndKind = func(childComplexity int, _ string, _ model.LocationKind, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LocationsByAddressID = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.LocationChildren = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Locations = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *string, _ *string, _ *model.LocationKind, _ *bool, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllLocations = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.OrganizationUnitsByHeadPersonID = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.OrganizationUnitsByPrimaryLocationID = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.OrganizationUnitChildren = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.OrganizationUnits = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *string, _ *string, _ *model.OrgUnitKind, _ *bool, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllOrganizationUnits = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartiesByEmail = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.Parties = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *model.PartyKind, _ *model.PartyStatus, _ *model.PartyRoleKind, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllParties = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyRolesByParty = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyRoles = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *model.PartyRoleKind, _ *bool, _ *string) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllPartyRoles = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyRolesByKind = func(childComplexity int, _ model.PartyRoleKind, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllPartyRolesByKind = func(childComplexity int, _ model.PartyRoleKind, _ *bool, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyContactsByParty = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyContactsByIamUserID = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyContactsByEmail = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyContacts = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *string, _ *model.PortalAccessStatus, _ *bool, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllPartyContacts = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyAddressLinksByParty = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyAddressLinksByPartyAndUsage = func(childComplexity int, _ string, _ model.AddressUsageKind, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyAddressLinksByAddressID = func(childComplexity int, _ string, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.PartyAddressLinks = func(childComplexity int, _ *string, _ *string, first *int, last *int, _ *string, _ *string, _ *model.AddressUsageKind, _ *bool) int {
		return paginated(childComplexity, first, last)
	}
	c.Query.AllPartyAddressLinks = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}

	c.LegalEntity.Children = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Location.Children = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.OrganizationUnit.Children = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Party.Roles = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Party.Contacts = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}
	c.Party.Addresses = func(childComplexity int, _ *string, _ *string, first *int, last *int) int {
		return paginated(childComplexity, first, last)
	}

	return c
}
