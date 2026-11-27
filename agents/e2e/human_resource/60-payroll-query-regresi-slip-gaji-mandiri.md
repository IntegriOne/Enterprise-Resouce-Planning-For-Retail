# Regresi slip gaji mandiri

Tiga pemeriksaan lulus dengan race detector pada checkout aktif. Jalur gRPC memakai PostgreSQL disposable dan fixture penerbitan Finance serta DRM.

| Lapisan | Kasus yang diverifikasi |
| --- | --- |
| Renderer HRM | PDF deterministik dan dapat diekstrak dengan `pdftotext` |
| HRM dan PostgreSQL | Hasil payroll terkait pemilik, publikasi PDF, pengulangan penyimpanan, dan penolakan ketika tautan pemilik berubah |
| GraphQL Gateway dengan fixture gRPC | Detail, daftar Relay, dan PDF milik sendiri. Grant GLOBAL saja atau SELF milik principal lain ditolak. Respons PDF memakai `Cache-Control: no-store`. Pengulangan PDF melalui alias ditolak. |

Uji ini membuktikan perilaku HRM dan Gateway dengan fixture. Uji HTTPS dengan penerbitan Finance dan DRM runtime masih terbuka. [Kontrak akses dan penerbitan](../../grpc/hrm/24-kebutuhan-integrasi-drm.md).

Regresi tambahan mengubah status menjadi `terminated` saat pembaca PDF fixture sedang berjalan, dengan tautan principal tetap sama. Pemeriksaan ulang HRM mengembalikan `PermissionDenied` tanpa mengirim byte PDF. Uji PostgreSQL disposable dengan race detector dan DejaVu Sans lulus.
