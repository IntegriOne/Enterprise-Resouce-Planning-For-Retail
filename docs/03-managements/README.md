# Peta 13 Management Produk Retail

> **Dokumen management retail · diperiksa 13 November 2026.** Daftar resmi mengikuti [keputusan owner](../00-governance/decisions.md). Kepemilikan bisnis mengikuti [agents/boundary.md](../../agents/boundary.md). Daftar service dan dokumen tidak otomatis membuktikan API, implementasi, E2E, atau deployment siap.

## Management dalam scope aktif

Terdapat **13 management**. Halaman rincian sengaja digabung berdasarkan alur bisnis, sehingga **satu halaman dapat membahas lebih dari satu management**; tidak ada kewajiban satu berkas Markdown untuk tiap management.

| No. | Management | Dokumentasi terkait | Direktori source |
|---:|---|---|---|
| 1 | IAM Management | [Identity & Access](iam.md) | [`iam_management`](../../backend/services/iam_management/README.md) |
| 2 | Enterprise Foundation | [Legal Entity, Party & Location](enterprise-foundation.md) | [`enterprise_foundation`](../../backend/services/enterprise_foundation/README.md) |
| 3 | Product Lifecycle Management (PLM) | [Item, Variant & Revision](product-and-pricing.md) | [`product_lifecycle_management`](../../backend/services/product_lifecycle_management/README.md) |
| 4 | E-Commerce Commerce | [Channel, Cart & Quote](oms-and-commerce.md) | [`ecommerce_commerce`](../../backend/services/ecommerce_commerce/README.md) |
| 5 | Pricing & Revenue Management | [Price & Promotion](product-and-pricing.md) | [`pricing_revenue_management`](../../backend/services/pricing_revenue_management/README.md) |
| 6 | Order Management (OMS) | [Sales Order & Return](oms-and-commerce.md) | [`order_management`](../../backend/services/order_management/README.md) |
| 7 | Inventory & Warehouse Management (WMS) | [Receiving, Stock & Fulfillment](procurement-and-inventory.md) | [`inventory_warehouse_management`](../../backend/services/inventory_warehouse_management/README.md) |
| 8 | Finance & Accounting | [Money & Ledger](finance.md) | [`finance_accounting`](../../backend/services/finance_accounting/README.md) |
| 9 | Procurement & Sourcing | [Supplier, PO & ASN](procurement-and-inventory.md) | [`procurement_sourcing`](../../backend/services/procurement_sourcing/README.md) |
| 10 | Customer Service Management | [Case & Ticket](crm-and-customer-service.md) | [`customer_service_management`](../../backend/services/customer_service_management/README.md) |
| 11 | CRM & Sales Management | [Lead, Consent & Customer](crm-and-customer-service.md) | [`crm_sales_management`](../../backend/services/crm_sales_management/README.md) |
| 12 | Document & Records Management (DRM) | [Document Version & Retention](documents-and-hrm.md) | [`document_records_management`](../../backend/services/document_records_management/README.md) |
| 13 | Human Resource Management (HRM) | [Worker & Employment](documents-and-hrm.md) | [`human_resource_management`](../../backend/services/human_resource_management/README.md) |

## Di luar daftar 13 management

**Manufacturing Operations PAUSED** untuk pengembangan baru dan **tidak mempunyai direktori service aktif** di `backend/services/` pada tree `main` yang diperiksa 13 November 2026. Keputusan owner tidak mengizinkan penghapusan riwayat; rujukan historis tidak dianggap service yang tersedia sekarang. Management lain yang tidak ada pada tabel bukan anggota roadmap aktif; kebutuhan baru memerlukan keputusan owner.

Integrasi Shopee/TikTok Shop, webhook, analytics dan website merupakan kapabilitas lintas-management menurut [fase produk F1–F4](../01-product/vision-and-scope.md), **bukan management tambahan secara otomatis**. Status implementasi harus diverifikasi dari kode, kontrak dan tes, bukan keberadaan README atau direktori.
