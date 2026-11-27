# Working History #10 — Sidebar Users 3 Segmen (Customers, Vendors, Internals) & Tombol Create Employee / Create Vendor

- **Tanggal:** 2026-11-25 02:50 WIB
- **Ruang Lingkup:** Sidebar IAM, Navbar, Overview, & Direktori Pengguna `/identity/users/*`.

## Ringkasan Perubahan

1. **Pemecahan Menu Sidebar Pengguna Menjadi 3 Segmen (`UserKind`):**
   - Menambahkan grup `Pengguna` pada [`identity-sidebar.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/components/identity/identity-sidebar.tsx) dengan 3 tautan plus lencana `kind`: **Customers** (`/identity/users/customers`), **Vendors** (`/identity/users/vendors`), dan **Internals** (`/identity/users/internals`).
   - Menambahkan ikon `"building"` & `"badge"` di [`identity-icons.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/components/identity/identity-icons.tsx), metadata judul rute di [`identity-navbar.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/components/identity/identity-navbar.tsx), serta 3 halaman App Router di `/identity/users/{customers,vendors,internals}/page.tsx`.
2. **Filter `UserKind` & Tombol Create Sesuai GraphQL (`createEmployeeUser` / `createVendorUser`):**
   - Memparameterisasi [`users-client.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/identity/users/users-client.tsx) per segmen dengan sub-tab kategori `VENDOR`/`SUPPLIER` dan `EMPLOYEE`/`SYSTEM` serta lencana `total_count`.
   - Memperbarui [`user-create-modal.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/identity/users/user-create-modal.tsx) dan [`overview-client.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/identity/overview-client.tsx) agar menyediakan tombol eksplisit **Create Employee** (`createEmployeeUser`), **Create Vendor** (`createVendorUser`), **Create Supplier** (`createSupplierUser`), dan **Create Customer** (`createUser`).
3. **Verifikasi:** `19/19` unit test, `typegen`, `eslint`, `tsc`, `next build` (`15/15`), `frontend-1`/`frontend-2`/`caddy` `Healthy`, `307 -> /login` anonim pada 3 rute baru, serta Chrome E2E rute publik PASS.
