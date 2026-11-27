package enterprise_foundation

import (
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPortalUserProvisioning_ValidationAndLinking(t *testing.T) {
	suite, reg := newEFMockSuite(t)
	authzEng := newTestAuthorizer(t, nil)
	adp, err := NewAdapter(reg.EnterpriseFoundation, authzEng, 2*time.Second)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	nowTS := timestamppb.New(time.Now())
	vendorPartyID := "aaaaaaaa-1111-4111-8111-111111111111"
	unlinkedContactID := "bbbbbbbb-2222-4222-8222-222222222222"
	vendorRoleID := "dddddddd-3333-4333-8333-333333333333"
	contactEmail := "pic.vendor@integrione.test"

	suite.mu.Lock()
	suite.parties[vendorPartyID] = &efv1.Party{
		Id:          vendorPartyID,
		PartnerCode: "VND-001",
		Kind:        efv1.PartyKind_PARTY_KIND_ORGANIZATION,
		DisplayName: "PT Vendor Jahit",
		Status:      efv1.PartyStatus_PARTY_STATUS_ACTIVE,
		CreatedAt:   nowTS,
		UpdatedAt:   nowTS,
	}
	suite.partyRoles[vendorRoleID] = &efv1.PartyRole{
		Id:        vendorRoleID,
		PartyId:   vendorPartyID,
		Role:      efv1.PartyRoleKind_PARTY_ROLE_KIND_VENDOR,
		IsActive:  true,
		CreatedAt: nowTS,
		UpdatedAt: nowTS,
	}
	suite.partyContacts[unlinkedContactID] = &efv1.PartyContact{
		Id:                 unlinkedContactID,
		PartyId:            vendorPartyID,
		FullName:           "Dewi Sartika",
		ContactRole:        "account_lead",
		Email:              &contactEmail,
		PortalAccessStatus: efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE,
		IsActive:           true,
		CreatedAt:          nowTS,
		UpdatedAt:          nowTS,
	}
	suite.mu.Unlock()

	globalCtx := newTestSessionContext(t, false, false, time.Now(), globalEFGrants(PermPartiesWrite))
	vendorScopedCtx := newTestSessionContext(t, false, false, time.Now(), []authz.PermissionGrant{
		{Code: PermPartiesWrite, Scope: authz.ScopeGrant{Type: scopeTypeVendorParty, ID: vendorPartyID}},
	})
	noPermCtx := newTestSessionContext(t, false, false, time.Now(), nil)

	// 1. RBAC fail-closed: tanpa izin -> AUTHORIZATION dengan 0 panggilan RPC EF.
	beforeCalls := suite.calls.Load()
	if _, _, err := adp.PreparePortalUserProvisioning(noPermCtx, vendorPartyID, nil, "VENDOR"); err == nil {
		t.Fatalf("expected AUTHORIZATION error when parties.write missing")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected CodeForbidden, got %v", err)
	}
	if _, _, err := adp.PreparePortalUserProvisioning(noPermCtx, "bad-id", nil, "VENDOR"); err == nil {
		t.Fatalf("expected AUTHORIZATION before ID oracle check")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected CodeForbidden on malformed ID without perm, got %v", err)
	}
	if suite.calls.Load() != beforeCalls {
		t.Fatalf("expected 0 EF RPC calls on authorization denial")
	}

	// 2. Role mismatch (party testUUID5 hanya punya SUPPLIER, dipanggil sebagai VENDOR) -> CONFLICT.
	if _, _, err := adp.PreparePortalUserProvisioning(globalCtx, EncodeGID(ResourceParty, testUUID5), nil, "VENDOR"); err == nil {
		t.Fatalf("expected CONFLICT when party lacks VENDOR role")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeConflict {
		t.Fatalf("expected CodeConflict on missing VENDOR role, got %v", err)
	}

	// 3. Contact sudah tertaut (testUUID7 pada testUUID5) -> CONFLICT sebelum CreateUser.
	if _, _, err := adp.PreparePortalUserProvisioning(globalCtx, testUUID5, &[]string{testUUID7}[0], "SUPPLIER"); err == nil {
		t.Fatalf("expected CONFLICT for already-linked party contact")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeConflict {
		t.Fatalf("expected CodeConflict for linked contact, got %v", err)
	}

	// 4. Scoped vendor_party berhasil untuk party cocok dan menautkan existing contact_id via RequestPartyPortalLink.
	normParty, normContact, err := adp.PreparePortalUserProvisioning(vendorScopedCtx, EncodeGID(ResourceParty, vendorPartyID), &[]string{EncodeGID(ResourcePartyContact, unlinkedContactID)}[0], "VENDOR")
	if err != nil || normParty != vendorPartyID || normContact == nil || *normContact != unlinkedContactID {
		t.Fatalf("PreparePortalUserProvisioning(existing contact) = (%q, %v, %v)", normParty, normContact, err)
	}
	linkedExisting, err := adp.CompletePortalUserProvisioning(vendorScopedCtx, PortalUserProvisioningInput{
		PartyID:      normParty,
		ContactID:    normContact,
		RequiredRole: "VENDOR",
		FullName:     "Dewi Sartika",
		Email:        contactEmail,
		IAMUserID:    testUUID8,
	})
	if err != nil || linkedExisting == nil || linkedExisting.RawID != unlinkedContactID ||
		linkedExisting.IamUserID == nil || *linkedExisting.IamUserID != EncodeGID(ResourceIAMUser, testUUID8) ||
		linkedExisting.PortalAccessStatus != model.PortalAccessStatusPendingLink {
		t.Fatalf("CompletePortalUserProvisioning(existing contact) unexpected: contact=%+v err=%v", linkedExisting, err)
	}

	// 5. Auto-create PIC contact saat contact_id nil -> memanggil AddPartyContact dengan iam_user_id & PENDING_LINK.
	normPartySup, normContactNil, err := adp.PreparePortalUserProvisioning(globalCtx, EncodeGID(ResourceParty, testUUID5), nil, "SUPPLIER")
	if err != nil || normPartySup != testUUID5 || normContactNil != nil {
		t.Fatalf("PreparePortalUserProvisioning(auto contact) = (%q, %v, %v)", normPartySup, normContactNil, err)
	}
	newUserUUID := "eeeeeeee-4444-4444-8444-444444444444"
	linkedNew, err := adp.CompletePortalUserProvisioning(globalCtx, PortalUserProvisioningInput{
		PartyID:      normPartySup,
		RequiredRole: "SUPPLIER",
		FullName:     "PIC Baru Supplier",
		Email:        "pic.sup@integrione.test",
		IAMUserID:    newUserUUID,
	})
	if err != nil || linkedNew == nil ||
		linkedNew.IamUserID == nil || *linkedNew.IamUserID != EncodeGID(ResourceIAMUser, newUserUUID) ||
		linkedNew.PortalAccessStatus != model.PortalAccessStatusPendingLink ||
		linkedNew.ContactRole != "portal_pic" {
		t.Fatalf("CompletePortalUserProvisioning(new contact) unexpected: contact=%+v err=%v", linkedNew, err)
	}

	// 6. Degraded EF -> UNAVAILABLE.
	suite.mu.Lock()
	suite.forcedErr = status.Error(codes.Unavailable, "dial tcp enterprise-foundation:9002: connection refused")
	suite.mu.Unlock()
	if _, _, err := adp.PreparePortalUserProvisioning(globalCtx, testUUID5, nil, "SUPPLIER"); err == nil {
		t.Fatalf("expected UNAVAILABLE when EF server fails")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected CodeUnavailable, got %v", err)
	}
}
