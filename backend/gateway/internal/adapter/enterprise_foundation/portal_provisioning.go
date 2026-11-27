package enterprise_foundation

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
)

// PortalUserProvisioningInput menyimpan parameter penautan akun portal IAM ke PIC mitra Enterprise Foundation.
type PortalUserProvisioningInput struct {
	PartyID          string
	ContactID        *string
	RequiredRole     string
	FullName         string
	Email            string
	ContactRole      *string
	JobTitle         *string
	Phone            *string
	IsPrimaryContact *bool
	IAMUserID        string
}

// PreparePortalUserProvisioning menegakkan izin `enterprise_foundation.parties.write` dan memverifikasi kelayakan mitra, peran, serta PIC opsional sebelum pembuatan akun IAM.
func (a *Adapter) PreparePortalUserProvisioning(
	ctx context.Context,
	partyID string,
	contactID *string,
	requiredRole string,
) (string, *string, error) {
	if a == nil {
		return "", nil, middleware.ErrUnavailable("enterprise foundation service unavailable", source_connections.ErrEFUnavailable)
	}
	protoRole, err := parsePortalRequiredPartyRole(requiredRole)
	if err != nil {
		return "", nil, err
	}

	cleanPartyID, idErr := ParseTypedID(partyID, ResourceParty)
	if idErr != nil {
		if _, authErr := a.requirePerm(ctx, PermPartiesWrite, false); authErr != nil {
			return "", nil, authErr
		}
		return "", nil, idErr
	}

	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return "", nil, err
	}

	cleanContactID, err := ParseOptionalTypedID(contactID, ResourcePartyContact)
	if err != nil {
		return "", nil, err
	}

	partyResp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyByIdResponse, error) {
		return c.Parties.GetPartyById(rpcCtx, &efv1.GetPartyByIdRequest{Id: cleanPartyID})
	})
	if err != nil {
		return "", nil, err
	}
	party := partyResp.GetParty()
	if party == nil {
		return "", nil, middleware.ErrNotFound("party not found", nil)
	}
	if party.GetDeletedAt() != nil ||
		party.GetStatus() == efv1.PartyStatus_PARTY_STATUS_SUSPENDED ||
		party.GetStatus() == efv1.PartyStatus_PARTY_STATUS_INACTIVE {
		return "", nil, middleware.ErrConflict("target party is not eligible for portal user provisioning", nil)
	}

	roleResp, err := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyRoleByPartyAndKindResponse, error) {
		return c.Parties.GetPartyRoleByPartyAndKind(rpcCtx, &efv1.GetPartyRoleByPartyAndKindRequest{
			PartyId: cleanPartyID,
			Kind:    protoRole,
		})
	})
	if err != nil {
		if gwErr, ok := middleware.AsGatewayError(err); ok && gwErr.Code == middleware.CodeNotFound {
			return "", nil, middleware.ErrConflict("target party does not hold required "+strings.ToLower(strings.TrimSpace(requiredRole))+" role", err)
		}
		return "", nil, err
	}
	role := roleResp.GetRole()
	if role == nil || !role.GetIsActive() {
		return "", nil, middleware.ErrConflict("target party "+strings.ToLower(strings.TrimSpace(requiredRole))+" role is inactive", nil)
	}

	if cleanContactID != nil {
		contactResp, contactErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.GetPartyContactByIdResponse, error) {
			return c.Parties.GetPartyContactById(rpcCtx, &efv1.GetPartyContactByIdRequest{ContactId: *cleanContactID})
		})
		if contactErr != nil {
			return "", nil, contactErr
		}
		if err := validateExistingContactForProvisioning(contactResp.GetContact(), cleanPartyID); err != nil {
			return "", nil, err
		}
	}

	return cleanPartyID, cleanContactID, nil
}

// CompletePortalUserProvisioning menetapkan FK logis `party_contacts.iam_user_id` di Enterprise Foundation baik dengan menautkan kontak yang ada maupun menambahkan kontak PIC baru berstatus PENDING_LINK.
func (a *Adapter) CompletePortalUserProvisioning(
	ctx context.Context,
	in PortalUserProvisioningInput,
) (*model.PartyContact, error) {
	if a == nil {
		return nil, middleware.ErrUnavailable("enterprise foundation service unavailable", source_connections.ErrEFUnavailable)
	}
	if _, err := parsePortalRequiredPartyRole(in.RequiredRole); err != nil {
		return nil, err
	}

	cleanPartyID, err := ParseTypedID(in.PartyID, ResourceParty)
	if err != nil {
		return nil, err
	}
	sess, err := a.requireAnyScope(ctx, PermPartiesWrite, false, partyScopes(cleanPartyID)...)
	if err != nil {
		return nil, err
	}
	cleanContactID, err := ParseOptionalTypedID(in.ContactID, ResourcePartyContact)
	if err != nil {
		return nil, err
	}
	cleanIAMUserID, err := ParseUUIDOrGID(in.IAMUserID, ResourceIAMUser)
	if err != nil {
		return nil, err
	}

	if cleanContactID != nil {
		resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.RequestPartyPortalLinkResponse, error) {
			return c.Parties.RequestPartyPortalLink(rpcCtx, &efv1.RequestPartyPortalLinkRequest{
				ContactId: *cleanContactID,
				IamUserId: cleanIAMUserID,
			})
		})
		if rpcErr != nil {
			return nil, rpcErr
		}
		return toModelPartyContact(resp.GetContact()), nil
	}

	fullName := strings.TrimSpace(in.FullName)
	email := strings.TrimSpace(in.Email)
	if fullName == "" || email == "" {
		return nil, middleware.ErrInvalidArgument("full_name and email are required to create party contact", nil)
	}
	contactRole := "portal_pic"
	if in.ContactRole != nil && strings.TrimSpace(*in.ContactRole) != "" {
		contactRole = strings.TrimSpace(*in.ContactRole)
	}

	resp, rpcErr := callEF(ctx, a, sess, func(rpcCtx context.Context, c *source_connections.EnterpriseFoundationClientSet) (*efv1.AddPartyContactResponse, error) {
		return c.Parties.AddPartyContact(rpcCtx, &efv1.AddPartyContactRequest{
			PartyId:            cleanPartyID,
			IamUserId:          &cleanIAMUserID,
			FullName:           fullName,
			JobTitle:           trimOptionalStringPtr(in.JobTitle),
			ContactRole:        contactRole,
			Email:              &email,
			Phone:              trimOptionalStringPtr(in.Phone),
			PortalAccessStatus: efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_PENDING_LINK,
			IsPrimary:          derefBool(in.IsPrimaryContact, false),
		})
	})
	if rpcErr != nil {
		return nil, rpcErr
	}
	return toModelPartyContact(resp.GetContact()), nil
}

func parsePortalRequiredPartyRole(requiredRole string) (efv1.PartyRoleKind, error) {
	switch strings.ToUpper(strings.TrimSpace(requiredRole)) {
	case "VENDOR":
		return efv1.PartyRoleKind_PARTY_ROLE_KIND_VENDOR, nil
	case "SUPPLIER":
		return efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER, nil
	default:
		return efv1.PartyRoleKind_PARTY_ROLE_KIND_UNSPECIFIED, middleware.ErrInvalidArgument("unsupported party role for portal user provisioning", nil)
	}
}

func validateExistingContactForProvisioning(contact *efv1.PartyContact, expectedPartyID string) error {
	if contact == nil {
		return middleware.ErrNotFound("party contact not found", nil)
	}
	if contact.GetDeletedAt() != nil {
		return middleware.ErrConflict("party contact has been deleted", nil)
	}
	if !contact.GetIsActive() {
		return middleware.ErrConflict("party contact is inactive", nil)
	}
	if !strings.EqualFold(strings.TrimSpace(contact.GetPartyId()), expectedPartyID) {
		return middleware.ErrConflict("party contact does not belong to target party", nil)
	}
	if strings.TrimSpace(contact.GetIamUserId()) != "" {
		return middleware.ErrConflict("party contact is already linked to another IAM user", nil)
	}
	if contact.GetPortalAccessStatus() != efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE {
		return middleware.ErrConflict("party contact portal access status must be NONE before linking a new user", nil)
	}
	return nil
}

func trimOptionalStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
