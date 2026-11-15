# Sitemap Backoffice, Supplier Portal dan Website

> **UX/UI design brief · belum berupa file Figma atau prototipe visual jadi.** Keputusan fase mengikuti [produk](../01-product/roadmap.md) dan kebijakan akses [IAM](../02-architecture/security-and-sessions.md).

## Internal ERP
```text
Dashboard (jualan, stok, finance, channel health)
Produk ─ item/SKU/varian (PLM), listing (Commerce), harga (Pricing)
Pembelian ─ supplier/party, PO, ASN, discrepancy (Procurement)
Gudang ─ receiving, stok, reservation, picking/packing, dispatch (WMS)
Penjualan ─ website + marketplace orders (OMS), fulfillment dan return status
Keuangan ─ AP/AR, receipt, GL, tax, report (Finance)
Pelanggan ─ CRM, consent, tickets (CRM + Customer Service)
Integrasi ─ akun channel, token expiry, webhook queue, reconciliation, analytics
Pengaturan ─ IAM staff, company/party, dokumen/retensi, HRM
```

## Vendor/supplier portal (F2)
Login IAM→select authorized party/contract→PO or deliverable list→detail/version→ASN/dokumen→status/notifications→support. Supplier tidak pernah memperoleh akses Finance internal tanpa grant yang eksplisit.

## Website customer (F4 terakhir) & `/profile` universal
- **Website customer:** Home→Category/Collection→Search/Filter→Product/Variant/Size Chart→Guest Cart (HMAC)→Login di `/login` (`commerceClaimGuestCart`)→Checkout (shipping, tax, discount)→Quote final review→OMS Order (`AWAITING_PAYMENT`) + WMS Stock Hold TTL→Finance Payment Gateway (Xendit Invoice/VA/QRIS)→Webhook PAID→OMS `CONFIRMED` & Tracking→Unduh Invoice Customer (DRM terisolasi)→Return/Support. Legal/Privacy dan consent tersedia sebelum mengaktifkan tracking pemasaran.
- **Route umum `/profile`:** Tersedia untuk seluruh pengguna terautentikasi (**Internal Employee, Customer Web Commerce, Vendor CMT, dan Supplier**) untuk kelola profil akun dasar, kata sandi, TOTP/MFA, sesi aktif, serta akses dokumen pribadi terotorisasi (mis. slip gaji karyawan bagi internal), dengan isolasi RBAC–ABAC ketat di server.

## Design states yang wajib
Loading, empty, permission denied, service unavailable, quote expired, price refreshed, stock unavailable, webhook delayed, payment expired/failed, order unknown, partial refund dan retry idempoten. Interface tidak boleh menampilkan “Berhasil” hanya karena proses asinkron atau sesi Payment Gateway berstatus `PENDING`.

## Prototype acceptance
Paths yang dapat diklik internal: SKU→PO→receipt, supplier PO→ASN, channel link→analytics; buyer journey hanya prototype sampai fase F4. Uji Figma/prototype URL setelah benar-benar dibuat, jangan menulis URL asumsi.
