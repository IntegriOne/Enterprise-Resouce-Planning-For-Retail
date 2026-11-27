package handler_test

import (
	"testing"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPartyService_FullRPCCoverage(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := h.signedCtx(t)

	// Guard: empty payload, invalid timestamp / unknown optional enum ditolak InvalidArgument.
	if _, err := h.parties.CreateParty(ctx, &efv1.CreatePartyRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on empty CreateParty, got %v", status.Code(err))
	}
	if _, err := h.parties.UpdatePartyTaxProfile(ctx, &efv1.UpdatePartyTaxProfileRequest{
		Id:            "10000000-0000-4000-8000-000000000001",
		TaxVerifiedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: -5},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on invalid tax_verified_at timestamp, got %v", status.Code(err))
	}
	unknownRole := efv1.PartyRoleKind(99)
	if _, err := h.parties.ListParties(ctx, &efv1.ListPartiesRequest{Role: &unknownRole}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on unknown optional party role filter, got %v", status.Code(err))
	}

	// Siapkan alamat untuk PartyAddress RPCs.
	addr1Resp, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		AddressLine_1: "Jl. Industri Textile No. 5",
		City:          "Bandung",
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("CreateAddress 1 failed: %v", err)
	}
	addr2Resp, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		AddressLine_1: "Jl. Gudang Mitra No. 9",
		City:          "Tangerang",
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("CreateAddress 2 failed: %v", err)
	}

	taxTS := timestamppb.New(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))

	// --- 14 Party Core RPCs ---
	p1Resp, err := h.parties.CreateParty(ctx, &efv1.CreatePartyRequest{
		PartnerCode:       "SUP-KAIN-01",
		Kind:              efv1.PartyKind_PARTY_KIND_ORGANIZATION,
		DisplayName:       "PT Kain Nusantara",
		LegalName:         strPtr("PT Kain Nusantara Jaya"),
		Email:             strPtr("sales@kainnusantara.id"),
		Phone:             strPtr("+6222444555"),
		WebsiteUrl:        strPtr("https://kainnusantara.id"),
		TaxIdNumber:       strPtr("11.222.333.4-555.000"),
		TaxRegisteredName: strPtr("PT KAIN NUSANTARA JAYA"),
		IsPkp:             true,
		TaxVerifiedAt:     taxTS,
		PreferredLanguage: "id",
		Status:            efv1.PartyStatus_PARTY_STATUS_PENDING_VERIFICATION,
		Notes:             strPtr("Pemasok utama kain katun"),
	})
	if err != nil {
		t.Fatalf("CreateParty p1 failed: %v", err)
	}
	p1 := p1Resp.GetParty()

	p2Resp, err := h.parties.CreateParty(ctx, &efv1.CreatePartyRequest{
		PartnerCode:       "VND-LOG-01",
		Kind:              efv1.PartyKind_PARTY_KIND_ORGANIZATION,
		DisplayName:       "CV Ekspedisi Cepat",
		Email:             strPtr("ops@ekspedisicepat.id"),
		PreferredLanguage: "id",
		Status:            efv1.PartyStatus_PARTY_STATUS_ACTIVE,
	})
	if err != nil {
		t.Fatalf("CreateParty p2 failed: %v", err)
	}
	p2 := p2Resp.GetParty()

	if r, err := h.parties.GetPartyById(ctx, &efv1.GetPartyByIdRequest{Id: p1.GetId()}); err != nil || r.GetParty().GetPartnerCode() != "sup-kain-01" {
		t.Fatalf("GetPartyById failed: %v / %+v", err, r.GetParty())
	}
	if r, err := h.parties.GetPartiesByIds(ctx, &efv1.GetPartiesByIdsRequest{Ids: []string{p1.GetId(), p2.GetId()}}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("GetPartiesByIds failed: %v", err)
	}
	if r, err := h.parties.GetPartyByCode(ctx, &efv1.GetPartyByCodeRequest{PartnerCode: "SUP-KAIN-01"}); err != nil || r.GetParty().GetId() != p1.GetId() {
		t.Fatalf("GetPartyByCode failed: %v", err)
	}
	if r, err := h.parties.GetPartyByTaxIdNumber(ctx, &efv1.GetPartyByTaxIdNumberRequest{TaxIdNumber: "11.222.333.4-555.000"}); err != nil || r.GetParty().GetId() != p1.GetId() {
		t.Fatalf("GetPartyByTaxIdNumber failed: %v", err)
	}
	if r, err := h.parties.ListPartiesByEmail(ctx, &efv1.ListPartiesByEmailRequest{Email: "sales@kainnusantara.id"}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartiesByEmail failed: %v", err)
	}
	if r, err := h.parties.GetAllParties(ctx, &efv1.GetAllPartiesRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || len(r.GetItems()) != 1 || r.GetPage().GetTotal() != 2 || !r.GetPage().GetHasMore() {
		t.Fatalf("GetAllParties pagination failed: %v / %+v", err, r.GetPage())
	}
	if r, err := h.parties.UpdatePartyProfile(ctx, &efv1.UpdatePartyProfileRequest{
		Id:                p1.GetId(),
		Kind:              efv1.PartyKind_PARTY_KIND_ORGANIZATION,
		DisplayName:       "PT Kain Nusantara Prima",
		LegalName:         strPtr("PT Kain Nusantara Jaya Tbk"),
		Email:             strPtr("sales@kainnusantara.id"),
		PreferredLanguage: "en",
	}); err != nil || r.GetParty().GetDisplayName() != "PT Kain Nusantara Prima" {
		t.Fatalf("UpdatePartyProfile failed: %v", err)
	}
	if r, err := h.parties.UpdatePartyTaxProfile(ctx, &efv1.UpdatePartyTaxProfileRequest{
		Id:                 p1.GetId(),
		TaxIdNumber:        strPtr("11.222.333.4-555.000"),
		TaxRegisteredName:  strPtr("PT KAIN NUSANTARA PRIMA"),
		IsPkp:              false,
		IsTaxExempt:        true,
		TaxExemptionReason: strPtr("SKB PPN Export"),
		TaxVerifiedAt:      taxTS,
	}); err != nil || !r.GetParty().GetIsTaxExempt() {
		t.Fatalf("UpdatePartyTaxProfile failed: %v", err)
	}
	if r, err := h.parties.VerifyAndActivateParty(ctx, &efv1.VerifyAndActivatePartyRequest{Id: p1.GetId()}); err != nil || r.GetParty().GetStatus() != efv1.PartyStatus_PARTY_STATUS_ACTIVE {
		t.Fatalf("VerifyAndActivateParty failed: %v", err)
	}
	if r, err := h.parties.SuspendParty(ctx, &efv1.SuspendPartyRequest{Id: p2.GetId()}); err != nil || r.GetParty().GetStatus() != efv1.PartyStatus_PARTY_STATUS_SUSPENDED {
		t.Fatalf("SuspendParty failed: %v", err)
	}
	if r, err := h.parties.DeactivateParty(ctx, &efv1.DeactivatePartyRequest{Id: p2.GetId()}); err != nil || r.GetParty().GetStatus() != efv1.PartyStatus_PARTY_STATUS_INACTIVE {
		t.Fatalf("DeactivateParty failed: %v", err)
	}

	// --- 13 PartyRole RPCs ---
	supRoleResp, err := h.parties.AssignPartyRole(ctx, &efv1.AssignPartyRoleRequest{
		PartyId:       p1.GetId(),
		Role:          efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER,
		ReferenceNote: strPtr("Supplier Kain Utama"),
	})
	if err != nil {
		t.Fatalf("AssignPartyRole supplier failed: %v", err)
	}
	supRole := supRoleResp.GetRole()

	b2bRoleResp, err := h.parties.AssignPartyRole(ctx, &efv1.AssignPartyRoleRequest{
		PartyId:       p1.GetId(),
		Role:          efv1.PartyRoleKind_PARTY_ROLE_KIND_CUSTOMER_B2B,
		ReferenceNote: strPtr("Pembeli grosir"),
	})
	if err != nil {
		t.Fatalf("AssignPartyRole customer_b2b failed: %v", err)
	}
	b2bRole := b2bRoleResp.GetRole()

	if r, err := h.parties.GetPartyRoleById(ctx, &efv1.GetPartyRoleByIdRequest{RoleId: supRole.GetId()}); err != nil || r.GetRole().GetRole() != efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER {
		t.Fatalf("GetPartyRoleById failed: %v", err)
	}
	if r, err := h.parties.ListPartyRolesByParty(ctx, &efv1.ListPartyRolesByPartyRequest{PartyId: p1.GetId()}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("ListPartyRolesByParty failed: %v", err)
	}
	if r, err := h.parties.GetPartyRoleByPartyAndKind(ctx, &efv1.GetPartyRoleByPartyAndKindRequest{PartyId: p1.GetId(), Kind: efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER}); err != nil || r.GetRole().GetId() != supRole.GetId() {
		t.Fatalf("GetPartyRoleByPartyAndKind failed: %v", err)
	}
	supKind := efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER
	if r, err := h.parties.ListPartyRoles(ctx, &efv1.ListPartyRolesRequest{PartyId: p1.GetId(), Role: &supKind, ActiveOnly: true}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartyRoles failed: %v", err)
	}
	if r, err := h.parties.GetAllPartyRoles(ctx, &efv1.GetAllPartyRolesRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || !r.GetPage().GetHasMore() {
		t.Fatalf("GetAllPartyRoles failed: %v", err)
	}
	if r, err := h.parties.ListPartyRolesByKind(ctx, &efv1.ListPartyRolesByKindRequest{Kind: efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER, ActiveOnly: true}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartyRolesByKind failed: %v", err)
	}
	if r, err := h.parties.GetAllPartyRolesByKind(ctx, &efv1.GetAllPartyRolesByKindRequest{Kind: efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER, ActiveOnly: true}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("GetAllPartyRolesByKind failed: %v", err)
	}
	if r, err := h.parties.ListParties(ctx, &efv1.ListPartiesRequest{Role: &supKind}); err != nil || len(r.GetItems()) != 1 || r.GetItems()[0].GetId() != p1.GetId() {
		t.Fatalf("ListParties filtered by role failed: %v", err)
	}
	if r, err := h.parties.ValidatePartyRoleScope(ctx, &efv1.ValidatePartyRoleScopeRequest{
		PartyId:      p1.GetId(),
		RequiredRole: efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER,
	}); err != nil || !r.GetDecision().GetEligible() || r.GetDecision().GetScopeKind() != efv1.ScopeTargetKind_SCOPE_TARGET_KIND_SUPPLIER_PARTY {
		t.Fatalf("ValidatePartyRoleScope failed: err=%v decision=%+v", err, r.GetDecision())
	}
	if r, err := h.parties.UpdatePartyRoleNote(ctx, &efv1.UpdatePartyRoleNoteRequest{
		Id:            supRole.GetId(),
		ReferenceNote: strPtr("Supplier Kain Nasional"),
	}); err != nil || r.GetRole().GetReferenceNote() != "Supplier Kain Nasional" {
		t.Fatalf("UpdatePartyRoleNote failed: %v", err)
	}
	if r, err := h.parties.DeactivatePartyRole(ctx, &efv1.DeactivatePartyRoleRequest{RoleId: b2bRole.GetId()}); err != nil || r.GetRole().GetIsActive() {
		t.Fatalf("DeactivatePartyRole failed: %v", err)
	}
	if r, err := h.parties.ActivatePartyRole(ctx, &efv1.ActivatePartyRoleRequest{RoleId: b2bRole.GetId()}); err != nil || !r.GetRole().GetIsActive() {
		t.Fatalf("ActivatePartyRole failed: %v", err)
	}

	// --- 18 PartyContact & Portal Access RPCs ---
	iamUserID := "30000000-0000-4000-8000-000000000001"
	c1Resp, err := h.parties.AddPartyContact(ctx, &efv1.AddPartyContactRequest{
		PartyId:            p1.GetId(),
		FullName:           "Budi Santoso",
		JobTitle:           strPtr("Sales Manager"),
		ContactRole:        "portal_admin",
		Email:              strPtr("budi@kainnusantara.id"),
		Phone:              strPtr("+628111222333"),
		PortalAccessStatus: efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE,
		IsPrimary:          true,
	})
	if err != nil {
		t.Fatalf("AddPartyContact c1 failed: %v", err)
	}
	c1 := c1Resp.GetContact()

	c2Resp, err := h.parties.AddPartyContact(ctx, &efv1.AddPartyContactRequest{
		PartyId:            p1.GetId(),
		FullName:           "Siti Rahma",
		ContactRole:        "finance_pic",
		Email:              strPtr("siti@kainnusantara.id"),
		PortalAccessStatus: efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE,
		IsPrimary:          false,
	})
	if err != nil {
		t.Fatalf("AddPartyContact c2 failed: %v", err)
	}
	c2 := c2Resp.GetContact()

	// Link & Activate Portal Access -> EvaluateRepresentativePortalAccess
	if r, err := h.parties.RequestPartyPortalLink(ctx, &efv1.RequestPartyPortalLinkRequest{
		ContactId: c1.GetId(),
		IamUserId: iamUserID,
	}); err != nil || r.GetContact().GetPortalAccessStatus() != efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_PENDING_LINK {
		t.Fatalf("RequestPartyPortalLink failed: %v", err)
	}
	// Saat status masih pending_link, EvaluateRepresentativePortalAccess harus gagal fail-closed.
	if _, err := h.parties.EvaluateRepresentativePortalAccess(ctx, &efv1.EvaluateRepresentativePortalAccessRequest{
		PartyId:      p1.GetId(),
		IamUserId:    iamUserID,
		RequiredRole: efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER,
	}); err == nil {
		t.Fatal("expected EvaluateRepresentativePortalAccess to reject pending_link contact")
	}
	if r, err := h.parties.ActivatePartyPortalAccess(ctx, &efv1.ActivatePartyPortalAccessRequest{
		ContactId: c1.GetId(),
		IamUserId: iamUserID,
	}); err != nil || r.GetContact().GetPortalAccessStatus() != efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_ACTIVE {
		t.Fatalf("ActivatePartyPortalAccess failed: %v", err)
	}
	if r, err := h.parties.EvaluateRepresentativePortalAccess(ctx, &efv1.EvaluateRepresentativePortalAccessRequest{
		PartyId:      p1.GetId(),
		IamUserId:    iamUserID,
		RequiredRole: efv1.PartyRoleKind_PARTY_ROLE_KIND_SUPPLIER,
	}); err != nil || !r.GetEvaluation().GetAllowed() || len(r.GetEvaluation().GetActiveRoles()) == 0 {
		t.Fatalf("EvaluateRepresentativePortalAccess active failed: err=%v eval=%+v", err, r.GetEvaluation())
	}

	// Lookup & update contact RPCs
	if r, err := h.parties.GetPartyContactById(ctx, &efv1.GetPartyContactByIdRequest{ContactId: c1.GetId()}); err != nil || r.GetContact().GetFullName() != "Budi Santoso" {
		t.Fatalf("GetPartyContactById failed: %v", err)
	}
	if r, err := h.parties.ListPartyContactsByParty(ctx, &efv1.ListPartyContactsByPartyRequest{PartyId: p1.GetId()}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("ListPartyContactsByParty failed: %v", err)
	}
	if r, err := h.parties.GetPrimaryPartyContact(ctx, &efv1.GetPrimaryPartyContactRequest{PartyId: p1.GetId()}); err != nil || r.GetContact().GetId() != c1.GetId() {
		t.Fatalf("GetPrimaryPartyContact failed: %v", err)
	}
	if r, err := h.parties.ListPartyContactsByIamUserId(ctx, &efv1.ListPartyContactsByIamUserIdRequest{IamUserId: iamUserID}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartyContactsByIamUserId failed: %v", err)
	}
	if r, err := h.parties.ListPartyContactsByEmail(ctx, &efv1.ListPartyContactsByEmailRequest{Email: "budi@kainnusantara.id"}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartyContactsByEmail failed: %v", err)
	}
	activePortal := efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_ACTIVE
	if r, err := h.parties.ListPartyContacts(ctx, &efv1.ListPartyContactsRequest{PartyId: p1.GetId(), PortalAccessStatus: &activePortal}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartyContacts failed: %v", err)
	}
	if r, err := h.parties.GetAllPartyContacts(ctx, &efv1.GetAllPartyContactsRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || !r.GetPage().GetHasMore() {
		t.Fatalf("GetAllPartyContacts failed: %v", err)
	}
	if r, err := h.parties.UpdatePartyContact(ctx, &efv1.UpdatePartyContactRequest{
		Id:          c2.GetId(),
		FullName:    "Siti Rahmawati",
		JobTitle:    strPtr("Finance Lead"),
		ContactRole: "finance_pic",
		Email:       strPtr("siti@kainnusantara.id"),
		IsPrimary:   false,
	}); err != nil || r.GetContact().GetFullName() != "Siti Rahmawati" {
		t.Fatalf("UpdatePartyContact failed: %v", err)
	}
	if r, err := h.parties.SetPrimaryPartyContact(ctx, &efv1.SetPrimaryPartyContactRequest{PartyId: p1.GetId(), ContactId: c2.GetId()}); err != nil || !r.GetContact().GetIsPrimary() {
		t.Fatalf("SetPrimaryPartyContact failed: %v", err)
	}
	if r, err := h.parties.RevokePartyPortalAccess(ctx, &efv1.RevokePartyPortalAccessRequest{ContactId: c1.GetId()}); err != nil || r.GetContact().GetPortalAccessStatus() != efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_REVOKED {
		t.Fatalf("RevokePartyPortalAccess failed: %v", err)
	}
	if r, err := h.parties.UnlinkPartyPortalUser(ctx, &efv1.UnlinkPartyPortalUserRequest{ContactId: c1.GetId()}); err != nil || r.GetContact().GetPortalAccessStatus() != efv1.PortalAccessStatus_PORTAL_ACCESS_STATUS_NONE {
		t.Fatalf("UnlinkPartyPortalUser failed: %v", err)
	}
	if r, err := h.parties.DeactivatePartyContact(ctx, &efv1.DeactivatePartyContactRequest{ContactId: c1.GetId()}); err != nil || r.GetContact().GetIsActive() {
		t.Fatalf("DeactivatePartyContact failed: %v", err)
	}
	if r, err := h.parties.ActivatePartyContact(ctx, &efv1.ActivatePartyContactRequest{ContactId: c1.GetId()}); err != nil || !r.GetContact().GetIsActive() {
		t.Fatalf("ActivatePartyContact failed: %v", err)
	}
	if r, err := h.parties.DeletePartyContact(ctx, &efv1.DeletePartyContactRequest{ContactId: c1.GetId()}); err != nil || !r.GetDeleted() {
		t.Fatalf("DeletePartyContact failed: %v", err)
	}

	// --- 11 PartyAddress RPCs ---
	link1Resp, err := h.parties.LinkPartyAddress(ctx, &efv1.LinkPartyAddressRequest{
		PartyId:        p1.GetId(),
		AddressId:      addr1Resp.GetAddress().GetId(),
		Usage:          efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING,
		IsPrimary:      true,
		ReplacePrimary: false,
	})
	if err != nil {
		t.Fatalf("LinkPartyAddress link1 failed: %v", err)
	}
	link1 := link1Resp.GetLink()

	link2Resp, err := h.parties.LinkPartyAddress(ctx, &efv1.LinkPartyAddressRequest{
		PartyId:        p1.GetId(),
		AddressId:      addr2Resp.GetAddress().GetId(),
		Usage:          efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING,
		IsPrimary:      true,
		ReplacePrimary: true,
	})
	if err != nil {
		t.Fatalf("LinkPartyAddress link2 replace_primary failed: %v", err)
	}
	link2 := link2Resp.GetLink()

	if r, err := h.parties.GetPartyAddressLinkById(ctx, &efv1.GetPartyAddressLinkByIdRequest{LinkId: link2.GetId()}); err != nil || !r.GetLink().GetIsPrimary() {
		t.Fatalf("GetPartyAddressLinkById failed: %v", err)
	}
	if r, err := h.parties.ListPartyAddressLinksByParty(ctx, &efv1.ListPartyAddressLinksByPartyRequest{PartyId: p1.GetId()}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("ListPartyAddressLinksByParty failed: %v", err)
	}
	if r, err := h.parties.ListPartyAddressLinksByPartyAndUsage(ctx, &efv1.ListPartyAddressLinksByPartyAndUsageRequest{PartyId: p1.GetId(), Usage: efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING}); err != nil || len(r.GetItems()) != 2 {
		t.Fatalf("ListPartyAddressLinksByPartyAndUsage failed: %v", err)
	}
	if r, err := h.parties.GetPrimaryPartyAddressLink(ctx, &efv1.GetPrimaryPartyAddressLinkRequest{PartyId: p1.GetId(), Usage: efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING}); err != nil || r.GetLink().GetId() != link2.GetId() {
		t.Fatalf("GetPrimaryPartyAddressLink failed: %v", err)
	}
	if r, err := h.parties.ListPartyAddressLinksByAddressId(ctx, &efv1.ListPartyAddressLinksByAddressIdRequest{AddressId: addr1Resp.GetAddress().GetId()}); err != nil || len(r.GetItems()) != 1 {
		t.Fatalf("ListPartyAddressLinksByAddressId failed: %v", err)
	}
	shipUsage := efv1.AddressUsageKind_ADDRESS_USAGE_KIND_SHIPPING
	primaryOnly := true
	if r, err := h.parties.ListPartyAddressLinks(ctx, &efv1.ListPartyAddressLinksRequest{
		PartyId:     p1.GetId(),
		Usage:       &shipUsage,
		PrimaryOnly: &primaryOnly,
	}); err != nil || len(r.GetItems()) != 1 || r.GetItems()[0].GetId() != link2.GetId() {
		t.Fatalf("ListPartyAddressLinks failed: %v", err)
	}
	if r, err := h.parties.GetAllPartyAddressLinks(ctx, &efv1.GetAllPartyAddressLinksRequest{Page: &efv1.PageRequest{Limit: 1, Offset: 0}}); err != nil || !r.GetPage().GetHasMore() {
		t.Fatalf("GetAllPartyAddressLinks failed: %v", err)
	}
	if r, err := h.parties.UpdatePartyAddressLink(ctx, &efv1.UpdatePartyAddressLinkRequest{
		Id:             link1.GetId(),
		Usage:          efv1.AddressUsageKind_ADDRESS_USAGE_KIND_BILLING,
		IsPrimary:      false,
		ReplacePrimary: false,
	}); err != nil || r.GetLink().GetUsage() != efv1.AddressUsageKind_ADDRESS_USAGE_KIND_BILLING {
		t.Fatalf("UpdatePartyAddressLink failed: %v", err)
	}
	if r, err := h.parties.SetPrimaryPartyAddressLink(ctx, &efv1.SetPrimaryPartyAddressLinkRequest{LinkId: link1.GetId()}); err != nil || !r.GetLink().GetIsPrimary() {
		t.Fatalf("SetPrimaryPartyAddressLink failed: %v", err)
	}
	if r, err := h.parties.UnlinkPartyAddress(ctx, &efv1.UnlinkPartyAddressRequest{LinkId: link1.GetId()}); err != nil || !r.GetUnlinked() {
		t.Fatalf("UnlinkPartyAddress failed: %v", err)
	}

	// RemoveRole & DeleteParty akhir.
	if r, err := h.parties.RemovePartyRole(ctx, &efv1.RemovePartyRoleRequest{RoleId: b2bRole.GetId()}); err != nil || !r.GetRemoved() {
		t.Fatalf("RemovePartyRole failed: %v", err)
	}
	if r, err := h.parties.DeleteParty(ctx, &efv1.DeletePartyRequest{Id: p2.GetId()}); err != nil || !r.GetDeleted() {
		t.Fatalf("DeleteParty failed: %v", err)
	}
}
