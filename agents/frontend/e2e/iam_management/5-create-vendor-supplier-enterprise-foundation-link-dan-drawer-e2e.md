# Laporan E2E Frontend #5: Provisioning `VENDOR` & `SUPPLIER` di `/identity/users/vendors`, Drawer Tautan EF, & Perbaikan Complexity Query

- **Tanggal:** 25 November 2026 (09:00 WIB)
- **Lingkungan Uji:** Caddy HTTPS (`https://erp.localhost:8443`), `frontend-1` & `frontend-2` (`Next.js 16.4.6` / `React 19.3.0`), `gateway-1` & `gateway-2`, `iam`, `enterprise-foundation`, `postgres`, `redis`.
- **Rute & Operasi Teruji:** `/login`, `/identity/users/vendors`, `UserCreateModal`, `UserDetailDrawer`, query `parties`, `partyContactsByParty`, `partyContactsByIamUserId`, dan mutasi `createVendorUser` / `createSupplierUser`.

## Matriks Kasus Uji

| No | Kasus Uji | Kondisi / Aktor | Hasil Ekspektasi | Hasil Aktual | Status |
|---|---|---|---|---|---|
| 1 | Unit test, typegen, ESLint, TSC, & build | `frontend/` (`13` test `node --test`, `next typegen`, `npm run lint`, `tsc --noEmit`, `npm run build`) | Exit `0`, zero error/warning, seluruh `13` test termasuk `lookupPartyContactsByIamUserId` PASS | `13/13` PASS, `next build` (`15/15` rute) exit `0` | PASS |
| 2 | Create Vendor via UI (`/identity/users/vendors`) | Chrome Headless login SuperAdmin -> modal `Create Vendor` -> pilih Party aktif ber-role `VENDOR` -> isi PIC baru & akun -> submit | `createVendorUser` sukses, drawer terbuka tanpa error complexity (`<= 250`), `party_contacts.iam_user_id` tampil & tersimpan di PostgreSQL EF | Drawer tampilkan `party_id`, `contact_id`, `iam_user_id`, `PENDING_LINK`; Postgres: `iam.users.id = enterprise_foundation.party_contacts.iam_user_id` | PASS |
| 3 | Create Supplier via UI (taut PIC eksisting) | Chrome Headless modal `Create Supplier` -> pilih Party `SUPPLIER` -> muat PIC `lookupUnlinkedPartyContacts` -> pilih `PartyContact` aktif -> submit | `createSupplierUser` menautkan PIC lama tanpa membuat duplikat, drawer menampilkan `party_contacts.iam_user_id` cocok, DB PostgreSQL EF memperbarui `iam_user_id` & `portal_access_status = 'pending_link'` | PIC eksisting terisi `iam_user_id = iam.users.id`, drawer sinkron, `0` JS/console error & `0` HTTP `>= 400` | PASS |
| 4 | Isolasi kompleksitas query seluruh drawer `/identity` | Penurunan paginasi & penghapusan duplikasi field `nodes`/`edges` berlebih di `USER_ROLES_QUERY`, `ROLE_PERMISSIONS_QUERY`, `GROUP_MEMBERS_QUERY`, `GROUP_ROLES_QUERY` | Semua query drawer berada di bawah batas Gateway `MaxComplexity = 250` | Error awal `complexity 336 exceeds 250` teratasi tuntas pada runtime container | PASS |
