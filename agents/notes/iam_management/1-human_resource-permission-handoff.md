# Handoff katalog izin Human Resource Management untuk IAM

HRM memeriksa 408 kode izin unik melalui 492 pemetaan operasi di runtime. Format kodenya `human_resource.<resource>.<action>`. Query detail, daftar, dan finder menggunakan izin baca resource. Command workflow memakai izin bernama. Sumber runtime dan rancangan peran tersedia pada [katalog Go](../../../backend/services/human_resource/infrastructure/security/catalog.go), [matriks peran HRM](../../roles-permissions/human_resource/1-roles-permissions.md), dan [inventaris operasi Gateway](../../grpc/hrm/1-inventaris-operasi-grpc-hrm.md).

| Operasi | Izin HRM | Scope dan validasi |
| --- | --- | --- |
| `CancelServiceRequest` | `human_resource.service_requests.cancel` | `self` yang cocok dengan principal sesi atau grant global eksplisit. Command selalu mencocokkan principal pengusul. MFA terbaru wajib. |
| `GetMyPayslip`, `ListMyPayslips` | `human_resource.payslips.read_self` | Hanya `self` yang cocok dengan principal sesi. HRM memeriksa hubungan principal–worker, kepemilikan hasil, penerbitan Finance, dan akses versi DRM. |
| Baca pekerja, rencana tenaga kerja, dan baris rencana | Kode `read` pada resource masing-masing | Grant global eksplisit atau `legal_entity` dengan filter HRM sebelum query dan hitungan. |

Untuk operasi lainnya, cocokkan scope dengan implementasi policy HRM dan uji di Gateway–IAM–gRPC sebelum menerbitkan grant. Gunakan sensitivitas minimum dari `IsSensitiveOperation`, termasuk aturan MFA dan batas usia autentikasi sepuluh menit pada command yang ditandai sensitif. Template peran umum tidak memberikan `delete`.

Manifest resmi `backend/shared/rbac/catalogs/human_resource.json` belum tersedia. Saat resolver Gateway aktif dan pemilik management menyetujui katalog final, buat manifest dengan kode aktif, sensitivitas, scope yang didukung dan template role yang disetujui. Uji kesesuaian manifest dengan katalog runtime dan evaluasi IAM untuk scope `self` sebelum provisioning. Assignment pengguna tetap melalui mutasi resmi IAM. [Tata kelola manifest](../../users.md#5-tata-kelola-permission-manifest-kode--pemberian-akses-lintas-management)

`LinkWorkerPrincipal` memerlukan verifikasi keputusan provisioning yang mengikat ID bukti, principal employee, worker, party orang, dan principal pemohon pengaitan. `GetUserById` menyediakan tipe dan status akun. Kontrak IAM saat ini belum memberikan bukti hubungan akun dengan worker atau party yang dibutuhkan runtime HRM.
