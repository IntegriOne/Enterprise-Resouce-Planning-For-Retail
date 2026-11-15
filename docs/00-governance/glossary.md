# Glosarium ERP Retail dan Keamanan

> **Dokumentasi Integrione · 12 November 2026.** Penetapan owner, desain, implementasi, verifikasi dan deployment adalah status berbeda. Kebijakan normatif mengikuti [AGENTS.md](../../AGENTS.md), [arsitektur keamanan](../../agents/arsitektur.md), dan [batas management](../../agents/boundary.md).

| Istilah | Makna |
|---|---|
| Principal | Subjek human/workload di IAM; **tidak sama** dengan worker HRM, party supplier atau customer account Commerce |
| Party | Identitas organisasi/orang mitra bisnis dalam Enterprise Foundation |
| Realm | Batas isolasi keamanan IAM, **bukan** legal entity |
| Legal entity | Badan hukum penjual, pembeli atau pemilik transaksi sesuai scope |
| Permission | `resource.action` milik management dan IAM katalog; efektif bersama scope/grant |
| RBAC/ABAC | Role sebagai administrasi izin dan kebijakan atribut subject/resource/action/environment |
| Opaque session | Secret browser acak tanpa daftar izin, state server-side IAM/Redis |
| Snapshot authorization | Proyeksi grant berversi yang berlaku pada session/realm/management |
| SKU/varian | Identitas unit jual berbeda dari konten listing per channel |
| PO/ASN/receipt | Purchase order (Procurement), pemberitahuan kirim supplier (Procurement), penerimaan fisik (WMS) |
| OMS sales order | Order resmi yang diterima/dikelola Order Management, bukan cart/submission Commerce |
| Quote | Penawaran harga final yang diseal dengan versi sumber dan digest; bukan order accepted |
| Outbox/inbox | Penulisan event dalam transaksi lokal dan deduplikasi di konsumen |
| Webhook | HTTP callback provider; bisa berulang, terlambat atau tidak datang sama sekali |
| Reconciliation | Proses membandingkan fakta owner/provider dan memperbaiki proyeksi dengan command sah |
| POC | Eksperimen membuktikan kelayakan teknis/produk, bukan produksi |
| MRD/BRD/PRD/SRS/FRD/TRD | Kebutuhan pasar/bisnis/produk/sistem/fungsi/teknis, berbeda tetapi saling ditelusuri |
| ADR/RFC | Keputusan arsitektur yang dicatat/proposal yang meminta review |
| E2E | Test alur API hingga service dan database owner nyata; tidak identik dengan mock |
| RaaS | Model instalasi dan penyerahan sistem pelanggan menurut keputusan komersial yang disetujui; tidak otomatis shared multi-tenant |
