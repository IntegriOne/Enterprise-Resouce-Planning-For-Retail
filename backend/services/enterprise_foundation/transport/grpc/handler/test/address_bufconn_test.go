package handler_test

import (
	"testing"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAddressService_FullRPCCoverage(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	ctx := h.signedCtx(t)

	// Validasi input kosong pada endpoint CreateAddress wajib ditolak fail-closed.
	if _, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on empty CreateAddress, got %v", status.Code(err))
	}

	label1 := "HQ Jakarta"
	line2 := "Floor 10"
	prov := "DKI Jakarta"
	postal := "10220"
	lat := -6.2088
	lng := 106.8456

	// 1. CreateAddress
	createResp, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		Label:         &label1,
		AddressLine_1: "Jl. Jenderal Sudirman Kav 1",
		AddressLine_2: &line2,
		City:          "Jakarta Pusat",
		Province:      &prov,
		PostalCode:    &postal,
		CountryCode:   "ID",
		Latitude:      &lat,
		Longitude:     &lng,
	})
	if err != nil {
		t.Fatalf("CreateAddress failed: %v", err)
	}
	addr1 := createResp.GetAddress()
	if addr1.GetId() == "" || addr1.GetCity() != "Jakarta Pusat" || addr1.GetLatitude() != lat {
		t.Fatalf("unexpected CreateAddress payload: %+v", addr1)
	}

	// Duplikasi CreateAddress harus ditolak dengan AlreadyExists.
	if _, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		Label:         &label1,
		AddressLine_1: "jl. jenderal sudirman kav 1",
		AddressLine_2: &line2,
		City:          "jakarta pusat",
		Province:      &prov,
		PostalCode:    &postal,
		CountryCode:   "id",
	}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists on duplicate address, got %v (%v)", status.Code(err), err)
	}

	// 2. FindOrCreateAddress (dedup hit lalu miss baru)
	focHit, err := h.addresses.FindOrCreateAddress(ctx, &efv1.FindOrCreateAddressRequest{
		Label:         &label1,
		AddressLine_1: "Jl. Jenderal Sudirman Kav 1",
		AddressLine_2: &line2,
		City:          "Jakarta Pusat",
		Province:      &prov,
		PostalCode:    &postal,
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("FindOrCreateAddress hit failed: %v", err)
	}
	if focHit.GetCreated() || focHit.GetAddress().GetId() != addr1.GetId() {
		t.Fatalf("expected existing address from FindOrCreateAddress hit, got created=%v id=%s", focHit.GetCreated(), focHit.GetAddress().GetId())
	}

	focNew, err := h.addresses.FindOrCreateAddress(ctx, &efv1.FindOrCreateAddressRequest{
		AddressLine_1: "Jl. Asia Afrika No. 8",
		City:          "Bandung",
		Province:      strPtr("Jawa Barat"),
		PostalCode:    strPtr("40111"),
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("FindOrCreateAddress new failed: %v", err)
	}
	if !focNew.GetCreated() || focNew.GetAddress().GetId() == "" {
		t.Fatalf("expected created=true on new FindOrCreateAddress, got %+v", focNew)
	}
	addr2 := focNew.GetAddress()

	// Tambah alamat ketiga untuk menguji paginasi limit/offset/has_more.
	create3, err := h.addresses.CreateAddress(ctx, &efv1.CreateAddressRequest{
		AddressLine_1: "Jl. Tunjungan No. 12",
		City:          "Surabaya",
		CountryCode:   "ID",
	})
	if err != nil {
		t.Fatalf("CreateAddress 3 failed: %v", err)
	}
	addr3 := create3.GetAddress()

	// 3. GetAddressById
	gotResp, err := h.addresses.GetAddressById(ctx, &efv1.GetAddressByIdRequest{Id: addr1.GetId()})
	if err != nil || gotResp.GetAddress().GetId() != addr1.GetId() {
		t.Fatalf("GetAddressById failed: %v / %+v", err, gotResp)
	}

	// 4. GetAddressesByIds
	batchResp, err := h.addresses.GetAddressesByIds(ctx, &efv1.GetAddressesByIdsRequest{
		Ids: []string{addr1.GetId(), addr3.GetId()},
	})
	if err != nil || len(batchResp.GetItems()) != 2 {
		t.Fatalf("GetAddressesByIds failed: err=%v len=%d", err, len(batchResp.GetItems()))
	}

	// 5. ListAddresses dengan filter kota & paginasi.
	listResp, err := h.addresses.ListAddresses(ctx, &efv1.ListAddressesRequest{
		City: "Bandung",
		Page: &efv1.PageRequest{Limit: 10, Offset: 0},
	})
	if err != nil || len(listResp.GetItems()) != 1 || listResp.GetPage().GetTotal() != 1 || listResp.GetPage().GetHasMore() {
		t.Fatalf("ListAddresses failed: err=%v items=%d page=%+v", err, len(listResp.GetItems()), listResp.GetPage())
	}

	// 6. GetAllAddresses dengan paginasi bertahap.
	page1, err := h.addresses.GetAllAddresses(ctx, &efv1.GetAllAddressesRequest{
		Page: &efv1.PageRequest{Limit: 2, Offset: 0},
	})
	if err != nil || len(page1.GetItems()) != 2 || page1.GetPage().GetTotal() != 3 || !page1.GetPage().GetHasMore() {
		t.Fatalf("GetAllAddresses page 1 failed: err=%v items=%d page=%+v", err, len(page1.GetItems()), page1.GetPage())
	}
	page2, err := h.addresses.GetAllAddresses(ctx, &efv1.GetAllAddressesRequest{
		Page: &efv1.PageRequest{Limit: 2, Offset: 2},
	})
	if err != nil || len(page2.GetItems()) != 1 || page2.GetPage().GetHasMore() {
		t.Fatalf("GetAllAddresses page 2 failed: err=%v items=%d page=%+v", err, len(page2.GetItems()), page2.GetPage())
	}

	// 7. UpdateAddress
	updResp, err := h.addresses.UpdateAddress(ctx, &efv1.UpdateAddressRequest{
		Id:               addr2.GetId(),
		Label:            strPtr("Bandung Warehouse"),
		AddressLine_1:    "Jl. Asia Afrika No. 8",
		City:             "Bandung",
		Province:         strPtr("Jawa Barat"),
		PostalCode:       strPtr("40111"),
		CountryCode:      "ID",
		FormattedAddress: strPtr("Jl. Asia Afrika No. 8, Bandung"),
	})
	if err != nil || updResp.GetAddress().GetLabel() != "Bandung Warehouse" {
		t.Fatalf("UpdateAddress failed: %v / %+v", err, updResp)
	}

	// 8. IsAddressReferenced sebelum dan sesudah ditautkan ke LegalEntity.
	refBefore, err := h.addresses.IsAddressReferenced(ctx, &efv1.IsAddressReferencedRequest{Id: addr1.GetId()})
	if err != nil || refBefore.GetReferenced() {
		t.Fatalf("expected address unreferenced initially, err=%v ref=%v", err, refBefore.GetReferenced())
	}

	_, err = h.legalEntities.CreateLegalEntity(ctx, &efv1.CreateLegalEntityRequest{
		EntityCode:              "PT-ADDR-REF",
		LegalName:               "PT Address Ref",
		Status:                  efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE,
		JurisdictionCountryCode: "ID",
		RegisteredAddressId:     strPtr(addr1.GetId()),
		FiscalYearStartMonth:    1,
	})
	if err != nil {
		t.Fatalf("CreateLegalEntity for ref check failed: %v", err)
	}

	refAfter, err := h.addresses.IsAddressReferenced(ctx, &efv1.IsAddressReferencedRequest{Id: addr1.GetId()})
	if err != nil || !refAfter.GetReferenced() {
		t.Fatalf("expected address referenced after legal entity link, err=%v ref=%v", err, refAfter.GetReferenced())
	}

	// 9. DeleteAddress gagal saat masih direferensikan, berhasil saat bebas.
	if _, err := h.addresses.DeleteAddress(ctx, &efv1.DeleteAddressRequest{Id: addr1.GetId()}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition deleting referenced address, got %v (%v)", status.Code(err), err)
	}
	delResp, err := h.addresses.DeleteAddress(ctx, &efv1.DeleteAddressRequest{Id: addr3.GetId()})
	if err != nil || !delResp.GetDeleted() {
		t.Fatalf("DeleteAddress failed: err=%v deleted=%v", err, delResp.GetDeleted())
	}
	if _, err := h.addresses.GetAddressById(ctx, &efv1.GetAddressByIdRequest{Id: addr3.GetId()}); status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound after DeleteAddress, got %v", status.Code(err))
	}
}

func strPtr(v string) *string { return &v }
