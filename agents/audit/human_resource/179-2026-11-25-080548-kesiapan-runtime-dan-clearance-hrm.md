# Kesiapan runtime dan clearance HRM

- Pencocokan handoff terkini menemukan 1.006 operasi cocok di 22 service dengan 669 Query dan 337 Mutation berdasarkan proto, handler, service, dan katalog izin. [Inventaris](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md)
- HRM meneruskan identitas rencana, pekerja, run dan hasil untuk verifikasi clearance. Fixture gRPC mencakup status run posted dan hasil individual paid dengan referensi pembayaran. Bukti clearance lintas rencana perlu ditolak verifier produksi. [Catatan Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md)
- Checkout menyediakan service Foundation, HRM, dan IAM. Adapter Finance belum ada. Paket DRM HRM menyediakan renderer dan port penyimpanan privat. Kesiapan produksi jalur slip serta clearance belum dapat dibuktikan dari fixture. [Catatan DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md)
- Pemeriksaan git menemukan perubahan lintas management di checkout bersama. Bukti audit ini merujuk ke worktree lokal.
