package party_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
)

type harness struct {
	svc      *party.Service
	parties  *memPartyRepo
	roles    *memPartyRoleRepo
	contacts *memPartyContactRepo
	links    *memPartyAddressRepo
	addrs    *stubAddressRepo
	outbox   *stubOutboxRepo
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	roles := newMemPartyRoleRepo()
	parties := newMemPartyRepo(roles)
	contacts := newMemPartyContactRepo()
	links := newMemPartyAddressRepo()
	addrs := &stubAddressRepo{items: make(map[string]*entity.Address)}
	outbox := &stubOutboxRepo{}

	svc, err := party.NewService(party.Dependencies{
		Transactor:     noopTransactor{},
		Parties:        parties,
		PartyRoles:     roles,
		PartyContacts:  contacts,
		PartyAddresses: links,
		Addresses:      addrs,
		Outbox:         outbox,
		Clock:          fixedClock{now: time.Date(2026, 11, 15, 14, 0, 0, 0, time.UTC)},
		IDGen:          &seqIDGen{},
	})
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	return &harness{
		svc:      svc,
		parties:  parties,
		roles:    roles,
		contacts: contacts,
		links:    links,
		addrs:    addrs,
		outbox:   outbox,
	}
}

func strPtr(v string) *string { return &v }

func TestNewServiceFailClosedAndNilReceiver(t *testing.T) {
	roles := newMemPartyRoleRepo()
	base := party.Dependencies{
		Transactor:     noopTransactor{},
		Parties:        newMemPartyRepo(roles),
		PartyRoles:     roles,
		PartyContacts:  newMemPartyContactRepo(),
		PartyAddresses: newMemPartyAddressRepo(),
		Addresses:      &stubAddressRepo{},
		Outbox:         &stubOutboxRepo{},
		IDGen:          &seqIDGen{},
	}

	cases := []struct {
		name  string
		mut   func(*party.Dependencies)
		field string
	}{
		{"nil transactor", func(d *party.Dependencies) { d.Transactor = nil }, "transactor"},
		{"nil parties", func(d *party.Dependencies) { d.Parties = nil }, "parties"},
		{"nil party_roles", func(d *party.Dependencies) { d.PartyRoles = nil }, "party_roles"},
		{"nil party_contacts", func(d *party.Dependencies) { d.PartyContacts = nil }, "party_contacts"},
		{"nil party_addresses", func(d *party.Dependencies) { d.PartyAddresses = nil }, "party_addresses"},
		{"nil addresses", func(d *party.Dependencies) { d.Addresses = nil }, "addresses"},
		{"nil outbox", func(d *party.Dependencies) { d.Outbox = nil }, "outbox"},
		{"nil id_gen", func(d *party.Dependencies) { d.IDGen = nil }, "id_gen"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := base
			tc.mut(&deps)
			if _, err := party.NewService(deps); err == nil || eferrors.KindOf(err) != eferrors.KindInvalidInput || eferrors.FieldOf(err) != tc.field {
				t.Fatalf("expected invalid input for %s, got %v", tc.field, err)
			}
		})
	}

	var nilSvc *party.Service
	if _, err := nilSvc.GetAll(context.Background(), domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindUnavailable {
		t.Fatalf("expected KindUnavailable on nil receiver, got %v", err)
	}
}

func TestPartyCRUDDuplicatesTaxAndGetAll(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	p1, err := h.svc.Create(ctx, party.CreateInput{
		PartnerCode:       "VND-GARMENT",
		Kind:              entity.PartyKindOrganization,
		DisplayName:       "CV Jahit Sejahtera",
		LegalName:         strPtr("CV Jahit Sejahtera Mandiri"),
		Email:             strPtr("Sales@Jahit.id"),
		TaxIDNumber:       strPtr("03.111.222.3-444.000"),
		IsPKP:             true,
		PreferredLanguage: "id",
		Status:            entity.PartyStatusPendingVerification,
	})
	if err != nil {
		t.Fatalf("Create p1 failed: %v", err)
	}
	if p1.PartnerCode != "vnd-garment" {
		t.Fatalf("expected normalized code vnd-garment, got %q", p1.PartnerCode)
	}

	callsBefore := h.parties.insertCalls

	// Duplikasi partner_code dan tax_id ditolak sebelum Insert.
	_, err = h.svc.Create(ctx, party.CreateInput{
		PartnerCode: " vnd-garment ",
		Kind:        entity.PartyKindOrganization,
		DisplayName: "Duplikat Kode",
		Status:      entity.PartyStatusActive,
	})
	if !errors.Is(err, eferrors.ErrPartyCodeExists) {
		t.Fatalf("expected ErrPartyCodeExists, got %v", err)
	}
	_, err = h.svc.Create(ctx, party.CreateInput{
		PartnerCode: "vnd-other",
		Kind:        entity.PartyKindOrganization,
		DisplayName: "Duplikat Tax ID",
		TaxIDNumber: strPtr("03.111.222.3-444.000"),
		Status:      entity.PartyStatusActive,
	})
	if !errors.Is(err, eferrors.ErrPartyTaxIDExists) {
		t.Fatalf("expected ErrPartyTaxIDExists, got %v", err)
	}
	if h.parties.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on duplicate party, before=%d after=%d", callsBefore, h.parties.insertCalls)
	}

	// Buat mitra kedua untuk uji bentrok NPWP saat UpdateTaxProfile.
	p2, err := h.svc.Create(ctx, party.CreateInput{
		PartnerCode: "sup-kain",
		Kind:        entity.PartyKindOrganization,
		DisplayName: "PT Tekstil Nusantara",
		TaxIDNumber: strPtr("04.999.888.7-666.000"),
		Status:      entity.PartyStatusActive,
	})
	if err != nil {
		t.Fatalf("Create p2 failed: %v", err)
	}

	if _, err := h.svc.UpdateTaxProfile(ctx, party.UpdateTaxProfileInput{
		ID:          p1.ID,
		TaxIDNumber: p2.TaxIDNumber,
		IsPKP:       true,
	}); !errors.Is(err, eferrors.ErrPartyTaxIDExists) {
		t.Fatalf("expected ErrPartyTaxIDExists on UpdateTaxProfile, got %v", err)
	}

	// Lookup, UpdateProfile, dan siklus hidup.
	if got, err := h.svc.GetByCode(ctx, "VND-GARMENT"); err != nil || got.ID != p1.ID {
		t.Fatalf("GetByCode failed: got=%+v err=%v", got, err)
	}
	if got, err := h.svc.GetByTaxIDNumber(ctx, "03.111.222.3-444.000"); err != nil || got.ID != p1.ID {
		t.Fatalf("GetByTaxIDNumber failed: got=%+v err=%v", got, err)
	}
	if byEmail, err := h.svc.ListByEmail(ctx, "sales@jahit.id", domainrepo.PageRequest{}); err != nil || byEmail.Total != 1 || len(byEmail.Items) != 1 {
		t.Fatalf("ListByEmail failed: page=%+v err=%v", byEmail, err)
	}
	if byIDs, err := h.svc.GetByIDs(ctx, []string{p1.ID, p2.ID, p1.ID}); err != nil || len(byIDs) != 2 {
		t.Fatalf("GetByIDs failed: len=%d err=%v", len(byIDs), err)
	}

	p1, err = h.svc.UpdateProfile(ctx, party.UpdateProfileInput{
		ID:                p1.ID,
		Kind:              entity.PartyKindOrganization,
		DisplayName:       "CV Jahit Sejahtera Baru",
		Email:             strPtr("sales@jahit.id"),
		PreferredLanguage: "id",
	})
	if err != nil || p1.DisplayName != "CV Jahit Sejahtera Baru" {
		t.Fatalf("UpdateProfile failed: got=%+v err=%v", p1, err)
	}
	if p1, err = h.svc.VerifyAndActivate(ctx, p1.ID); err != nil || p1.Status != entity.PartyStatusActive {
		t.Fatalf("VerifyAndActivate failed: got=%+v err=%v", p1, err)
	}
	if p1, err = h.svc.Suspend(ctx, p1.ID); err != nil || p1.Status != entity.PartyStatusSuspended {
		t.Fatalf("Suspend failed: got=%+v err=%v", p1, err)
	}
	if p1, err = h.svc.Deactivate(ctx, p1.ID); err != nil || p1.Status != entity.PartyStatusInactive {
		t.Fatalf("Deactivate failed: got=%+v err=%v", p1, err)
	}
	if err := h.svc.Delete(ctx, p1.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := h.svc.GetByID(ctx, p1.ID); !errors.Is(err, eferrors.ErrPartyNotFound) {
		t.Fatalf("expected deleted party not found, got %v", err)
	}

	// GetAll multi-page drain (>100 mitra).
	for i := 0; i < 110; i++ {
		id := fmt.Sprintf("dddddddd-dddd-4ddd-8ddd-%012d", i+1)
		p, err := entity.NewParty(id, entity.PartyParams{
			PartnerCode:       fmt.Sprintf("pty-bulk-%03d", i+1),
			Kind:              entity.PartyKindOrganization,
			DisplayName:       fmt.Sprintf("Mitra Bulk %03d", i+1),
			PreferredLanguage: "id",
			Status:            entity.PartyStatusActive,
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewParty seed error: %v", err)
		}
		h.parties.seed(p)
	}
	p1Page, err := h.svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 0})
	if err != nil || p1Page.Total != 111 || len(p1Page.Items) != 50 || !p1Page.HasMore {
		t.Fatalf("GetAll page 1 failed: page=%+v err=%v", p1Page, err)
	}
	p3Page, err := h.svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 100})
	if err != nil || p3Page.Total != 111 || len(p3Page.Items) != 11 || p3Page.HasMore {
		t.Fatalf("GetAll page 3 failed: page=%+v err=%v", p3Page, err)
	}
}

func TestPartyRolesAndScopeEvaluation(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	p, err := h.svc.Create(ctx, party.CreateInput{
		PartnerCode:       "vnd-scope",
		Kind:              entity.PartyKindOrganization,
		DisplayName:       "PT Vendor Scope",
		PreferredLanguage: "id",
		Status:            entity.PartyStatusPendingVerification,
	})
	if err != nil {
		t.Fatalf("Create party failed: %v", err)
	}

	role, err := h.svc.AssignRole(ctx, party.AssignRoleInput{
		PartyID:       p.ID,
		Role:          entity.PartyRoleKindVendor,
		ReferenceNote: strPtr("Vendor utama"),
	})
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	callsBefore := h.roles.insertCalls
	if _, err := h.svc.AssignRole(ctx, party.AssignRoleInput{
		PartyID: p.ID,
		Role:    entity.PartyRoleKindVendor,
	}); !errors.Is(err, eferrors.ErrPartyRoleExists) {
		t.Fatalf("expected ErrPartyRoleExists, got %v", err)
	}
	if h.roles.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on duplicate role, before=%d after=%d", callsBefore, h.roles.insertCalls)
	}

	// Mitra pending ditolak sebagai scope sampai diaktifkan.
	if _, err := h.svc.ValidatePartyRoleScope(ctx, p.ID, entity.PartyRoleKindVendor); !errors.Is(err, eferrors.ErrPartyInactive) {
		t.Fatalf("expected ErrPartyInactive before VerifyAndActivate, got %v", err)
	}
	if _, err := h.svc.VerifyAndActivate(ctx, p.ID); err != nil {
		t.Fatalf("VerifyAndActivate failed: %v", err)
	}
	if dec, err := h.svc.ValidatePartyRoleScope(ctx, p.ID, entity.PartyRoleKindVendor); err != nil || !dec.Eligible {
		t.Fatalf("expected eligible scope for active vendor role, dec=%+v err=%v", dec, err)
	}

	if _, err := h.svc.DeactivateRole(ctx, role.ID); err != nil {
		t.Fatalf("DeactivateRole failed: %v", err)
	}
	if _, err := h.svc.ValidatePartyRoleScope(ctx, p.ID, entity.PartyRoleKindVendor); !errors.Is(err, eferrors.ErrPartyRoleInactive) {
		t.Fatalf("expected ErrPartyRoleInactive after DeactivateRole, got %v", err)
	}
	if role, err = h.svc.ActivateRole(ctx, role.ID); err != nil || !role.IsActive {
		t.Fatalf("ActivateRole failed: got=%+v err=%v", role, err)
	}
	if role, err = h.svc.UpdateRoleNote(ctx, party.UpdateRoleInput{ID: role.ID, ReferenceNote: strPtr("Updated note")}); err != nil || *role.ReferenceNote != "Updated note" {
		t.Fatalf("UpdateRoleNote failed: got=%+v err=%v", role, err)
	}
	if byPartyRoles, err := h.svc.ListRolesByParty(ctx, p.ID, domainrepo.PageRequest{}); err != nil || byPartyRoles.Total != 1 || len(byPartyRoles.Items) != 1 {
		t.Fatalf("ListRolesByParty failed: page=%+v err=%v", byPartyRoles, err)
	}
	if listedRoles, err := h.svc.ListRoles(ctx, domainrepo.PartyRoleFilter{PartyID: p.ID, Query: "updated"}); err != nil || listedRoles.Total != 1 || len(listedRoles.Items) != 1 {
		t.Fatalf("ListRoles failed: page=%+v err=%v", listedRoles, err)
	}
	if allRoles, err := h.svc.GetAllRoles(ctx, domainrepo.PageRequest{Limit: 10, Offset: 0}); err != nil || allRoles.Total != 1 || len(allRoles.Items) != 1 {
		t.Fatalf("GetAllRoles failed: page=%+v err=%v", allRoles, err)
	}
	if list, err := h.svc.GetAllRolesByKind(ctx, entity.PartyRoleKindVendor, true, domainrepo.PageRequest{}); err != nil || list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("GetAllRolesByKind failed: page=%+v err=%v", list, err)
	}
	if err := h.svc.RemoveRole(ctx, role.ID); err != nil {
		t.Fatalf("RemoveRole failed: %v", err)
	}
}

func TestPartyContactsPortalLifecycleAndIAMPolicy(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	p, err := h.svc.Create(ctx, party.CreateInput{
		PartnerCode:       "sup-portal",
		Kind:              entity.PartyKindOrganization,
		DisplayName:       "PT Pemasok Portal",
		PreferredLanguage: "id",
		Status:            entity.PartyStatusPendingVerification,
	})
	if err != nil {
		t.Fatalf("Create party failed: %v", err)
	}
	if _, err := h.svc.AssignRole(ctx, party.AssignRoleInput{
		PartyID: p.ID,
		Role:    entity.PartyRoleKindSupplier,
	}); err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	iamUser1 := "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1"
	iamUser2 := "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2"

	c1, err := h.svc.AddContact(ctx, party.AddContactInput{
		PartyID:            p.ID,
		IAMUserID:          &iamUser1,
		FullName:           "Budi Santoso",
		ContactRole:        "account_manager",
		Email:              strPtr("budi@supplier.id"),
		PortalAccessStatus: entity.PortalAccessStatusPendingLink,
		IsPrimary:          true,
	})
	if err != nil {
		t.Fatalf("AddContact c1 failed: %v", err)
	}

	// Principal IAM yang sama tidak boleh ditautkan ke kontak kedua pada mitra sama.
	callsBefore := h.contacts.insertCalls
	if _, err := h.svc.AddContact(ctx, party.AddContactInput{
		PartyID:            p.ID,
		IAMUserID:          &iamUser1,
		FullName:           "Duplikat IAM",
		ContactRole:        "admin",
		PortalAccessStatus: entity.PortalAccessStatusPendingLink,
	}); !errors.Is(err, eferrors.ErrPartyContactIAMUserExists) {
		t.Fatalf("expected ErrPartyContactIAMUserExists, got %v", err)
	}
	if h.contacts.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on duplicate iam_user_id, before=%d after=%d", callsBefore, h.contacts.insertCalls)
	}

	// Kontak kedua dengan IsPrimary=true menggeser primary contact lama.
	c2, err := h.svc.AddContact(ctx, party.AddContactInput{
		PartyID:            p.ID,
		FullName:           "Siti Rahma",
		ContactRole:        "finance",
		Phone:              strPtr("+628123456789"),
		PortalAccessStatus: entity.PortalAccessStatusNone,
		IsPrimary:          true,
	})
	if err != nil {
		t.Fatalf("AddContact c2 failed: %v", err)
	}
	if prim, err := h.svc.GetPrimaryContact(ctx, p.ID); err != nil || prim.ID != c2.ID {
		t.Fatalf("expected c2 as primary contact, got=%+v err=%v", prim, err)
	}

	// ActivatePortalAccess gagal selama party masih pending_verification.
	if _, err := h.svc.ActivatePortalAccess(ctx, c1.ID, iamUser1); !errors.Is(err, eferrors.ErrPartyInactive) {
		t.Fatalf("expected ErrPartyInactive on portal activate while party pending, got %v", err)
	}

	// Aktifkan mitra, lalu aktifkan portal c1 -> evaluasi representative lolos!
	if _, err := h.svc.VerifyAndActivate(ctx, p.ID); err != nil {
		t.Fatalf("VerifyAndActivate failed: %v", err)
	}
	c1, err = h.svc.ActivatePortalAccess(ctx, c1.ID, iamUser1)
	if err != nil || c1.PortalAccessStatus != entity.PortalAccessStatusActive {
		t.Fatalf("ActivatePortalAccess failed: got=%+v err=%v", c1, err)
	}
	eval, err := h.svc.EvaluateRepresentativePortalAccess(ctx, p.ID, iamUser1, entity.PartyRoleKindSupplier)
	if err != nil || !eval.Allowed || eval.ContactID != c1.ID {
		t.Fatalf("EvaluateRepresentativePortalAccess failed: eval=%+v err=%v", eval, err)
	}

	// RequestPortalLink c2 dengan iamUser1 ditolak karena iamUser1 sudah dipakai c1; dengan iamUser2 berhasil.
	if _, err := h.svc.RequestPortalLink(ctx, c2.ID, iamUser1); !errors.Is(err, eferrors.ErrPartyContactIAMUserExists) {
		t.Fatalf("expected ErrPartyContactIAMUserExists on RequestPortalLink conflict, got %v", err)
	}
	if _, err := h.svc.RequestPortalLink(ctx, c2.ID, iamUser2); err != nil {
		t.Fatalf("RequestPortalLink iamUser2 failed: %v", err)
	}

	// Revoke lalu DeleteContact mencabut akses portal aktif dan memancarkan outbox.
	if c1, err = h.svc.RevokePortalAccess(ctx, c1.ID); err != nil || c1.PortalAccessStatus != entity.PortalAccessStatusRevoked {
		t.Fatalf("RevokePortalAccess failed: got=%+v err=%v", c1, err)
	}
	if _, err := h.svc.EvaluateRepresentativePortalAccess(ctx, p.ID, iamUser1, entity.PartyRoleKindSupplier); err == nil {
		t.Fatalf("expected revoked contact portal access to be denied")
	}
	if byPartyContacts, err := h.svc.ListContactsByParty(ctx, p.ID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || byPartyContacts.Total != 2 || len(byPartyContacts.Items) != 1 || !byPartyContacts.HasMore {
		t.Fatalf("ListContactsByParty page 1 failed: page=%+v err=%v", byPartyContacts, err)
	}
	if byIAM, err := h.svc.ListContactsByIAMUserID(ctx, iamUser1, domainrepo.PageRequest{}); err != nil || byIAM.Total != 1 || len(byIAM.Items) != 1 {
		t.Fatalf("ListContactsByIAMUserID failed: page=%+v err=%v", byIAM, err)
	}
	if byContactEmail, err := h.svc.ListContactsByEmail(ctx, "budi@supplier.id", domainrepo.PageRequest{}); err != nil || byContactEmail.Total != 1 || len(byContactEmail.Items) != 1 {
		t.Fatalf("ListContactsByEmail failed: page=%+v err=%v", byContactEmail, err)
	}
	if allContacts, err := h.svc.GetAllContacts(ctx, domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || allContacts.Total != 2 || len(allContacts.Items) != 1 || allContacts.HasMore {
		t.Fatalf("GetAllContacts page 2 failed: page=%+v err=%v", allContacts, err)
	}
	if err := h.svc.DeleteContact(ctx, c2.ID); err != nil {
		t.Fatalf("DeleteContact c2 failed: %v", err)
	}
}

func TestPartyAddressLinksAndPrimaryConflict(t *testing.T) {
	ctx := context.Background()
	h := newHarness(t)

	p, err := h.svc.Create(ctx, party.CreateInput{
		PartnerCode:       "sup-addr",
		Kind:              entity.PartyKindOrganization,
		DisplayName:       "PT Logistik Mitra",
		PreferredLanguage: "id",
		Status:            entity.PartyStatusActive,
	})
	if err != nil {
		t.Fatalf("Create party failed: %v", err)
	}

	addr1ID := "ffffffff-ffff-4fff-8fff-fffffffffff1"
	addr2ID := "ffffffff-ffff-4fff-8fff-fffffffffff2"
	for _, spec := range []struct{ id, line string }{{addr1ID, "Jl. Pelabuhan 1"}, {addr2ID, "Jl. Pelabuhan 2"}} {
		a, err := entity.NewAddress(spec.id, entity.AddressParams{
			AddressLine1: spec.line,
			City:         "Surabaya",
			CountryCode:  "ID",
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewAddress failed: %v", err)
		}
		h.addrs.items[spec.id] = a
	}

	link1, err := h.svc.LinkAddress(ctx, party.LinkAddressInput{
		PartyID:   p.ID,
		AddressID: addr1ID,
		Usage:     entity.AddressUsageKindBilling,
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("LinkAddress 1 failed: %v", err)
	}

	callsBefore := h.links.insertCalls
	// Duplikasi (party_id, address_id, usage) ditolak sebelum Insert.
	if _, err := h.svc.LinkAddress(ctx, party.LinkAddressInput{
		PartyID:   p.ID,
		AddressID: addr1ID,
		Usage:     entity.AddressUsageKindBilling,
		IsPrimary: false,
	}); !errors.Is(err, eferrors.ErrPartyAddressExists) {
		t.Fatalf("expected ErrPartyAddressExists, got %v", err)
	}
	// Konflik primary pada usage sama jika ReplacePrimary=false ditolak sebelum Insert.
	if _, err := h.svc.LinkAddress(ctx, party.LinkAddressInput{
		PartyID:        p.ID,
		AddressID:      addr2ID,
		Usage:          entity.AddressUsageKindBilling,
		IsPrimary:      true,
		ReplacePrimary: false,
	}); !errors.Is(err, eferrors.ErrPrimaryPartyAddressConflict) {
		t.Fatalf("expected ErrPrimaryPartyAddressConflict, got %v", err)
	}
	if h.links.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on invalid address links, before=%d after=%d", callsBefore, h.links.insertCalls)
	}

	// ReplacePrimary=true mengganti primary billing ke alamat ke-2 secara atomik.
	link2, err := h.svc.LinkAddress(ctx, party.LinkAddressInput{
		PartyID:        p.ID,
		AddressID:      addr2ID,
		Usage:          entity.AddressUsageKindBilling,
		IsPrimary:      true,
		ReplacePrimary: true,
	})
	if err != nil {
		t.Fatalf("LinkAddress 2 ReplacePrimary failed: %v", err)
	}
	if prim, err := h.svc.GetPrimaryAddressLink(ctx, p.ID, entity.AddressUsageKindBilling); err != nil || prim.ID != link2.ID {
		t.Fatalf("expected link2 as primary billing address, got=%+v err=%v", prim, err)
	}

	// Kembalikan primary ke link1, verifikasi paginasi tautan alamat, lalu hapus link2.
	if _, err := h.svc.SetPrimaryAddressLink(ctx, link1.ID); err != nil {
		t.Fatalf("SetPrimaryAddressLink failed: %v", err)
	}
	if byPartyP1, err := h.svc.ListAddressLinksByParty(ctx, p.ID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || byPartyP1.Total != 2 || len(byPartyP1.Items) != 1 || !byPartyP1.HasMore {
		t.Fatalf("ListAddressLinksByParty page 1 failed: page=%+v err=%v", byPartyP1, err)
	} else if byPartyP2, err := h.svc.ListAddressLinksByParty(ctx, p.ID, domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || byPartyP2.Total != 2 || len(byPartyP2.Items) != 1 || byPartyP2.HasMore || byPartyP2.Items[0].ID == byPartyP1.Items[0].ID {
		t.Fatalf("ListAddressLinksByParty page 2 failed: page=%+v p1=%+v err=%v", byPartyP2, byPartyP1, err)
	}
	if byUsage, err := h.svc.ListAddressLinksByPartyAndUsage(ctx, p.ID, entity.AddressUsageKindBilling, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || byUsage.Total != 2 || len(byUsage.Items) != 1 || !byUsage.HasMore {
		t.Fatalf("ListAddressLinksByPartyAndUsage failed: page=%+v err=%v", byUsage, err)
	}
	if byAddr, err := h.svc.ListAddressLinksByAddressID(ctx, addr1ID, domainrepo.PageRequest{}); err != nil || byAddr.Total != 1 || len(byAddr.Items) != 1 {
		t.Fatalf("ListAddressLinksByAddressID failed: page=%+v err=%v", byAddr, err)
	}
	billingUsage := entity.AddressUsageKindBilling
	if listedLinks, err := h.svc.ListAddressLinks(ctx, domainrepo.PartyAddressFilter{PartyID: p.ID, Usage: &billingUsage, Page: domainrepo.PageRequest{Limit: 1, Offset: 1}}); err != nil || listedLinks.Total != 2 || len(listedLinks.Items) != 1 || listedLinks.HasMore {
		t.Fatalf("ListAddressLinks failed: page=%+v err=%v", listedLinks, err)
	}
	if allLinks, err := h.svc.GetAllAddressLinks(ctx, domainrepo.PageRequest{Limit: 10}); err != nil || allLinks.Total != 2 || len(allLinks.Items) != 2 || allLinks.HasMore {
		t.Fatalf("GetAllAddressLinks failed: page=%+v err=%v", allLinks, err)
	}
	if err := h.svc.UnlinkAddress(ctx, link2.ID); err != nil {
		t.Fatalf("UnlinkAddress failed: %v", err)
	}
	if byParty, err := h.svc.ListAddressLinksByParty(ctx, p.ID, domainrepo.PageRequest{}); err != nil || byParty.Total != 1 || len(byParty.Items) != 1 {
		t.Fatalf("ListAddressLinksByParty failed: page=%+v err=%v", byParty, err)
	}
}
