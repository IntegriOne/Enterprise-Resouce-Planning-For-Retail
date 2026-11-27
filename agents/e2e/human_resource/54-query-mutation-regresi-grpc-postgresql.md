# Regresi handler gRPC HRM dengan PostgreSQL

Perintah `go test -p=1 -race -tags=integration ./transport/grpc/handler/... -count=1 -v` lulus. Output memuat 24 paket dengan tes dan 46 tes tingkat atas lulus, tanpa skip. Suite dijalankan berurutan pada database PostgreSQL Docker sementara bernama `hrm_disposable_*_test` dengan opt-in `HRM_TEST_DATABASE_DISPOSABLE=1`. Database telah dihapus setelah tes.

| Area | Cakupan pada run ini |
| --- | --- |
| Offboarding | Jalur penutupan ikut dijalankan dengan data payroll terminasi dan verifier clearance fixture. |
| Payroll dan slip mandiri | Font DejaVu Sans mengaktifkan renderer PDF. Fixture memeriksa format PDF dan digest sebelum penyimpanan, penolakan sebelum bukti penerbitan valid, replay penerbitan, serta akses hanya ke slip milik pekerja. |
| Personnel dan Workforce Planning | Tes handler memeriksa aturan perubahan data dan scope dengan PostgreSQL. |
| Domain lain | Seluruh tes handler bertag integrasi yang ditemukan dijalankan sebagai bagian dari perintah di atas. |

Verifikasi penerbitan payroll Finance, penyimpanan dokumen DRM, dan login IAM melalui Gateway HTTPS masih memerlukan runtime dependency masing-masing. [Handoff Finance](../../grpc/hrm/25-kebutuhan-integrasi-finance-payroll.md) · [Handoff DRM](../../grpc/hrm/24-kebutuhan-integrasi-drm.md).

Setelah perbaikan kegagalan sebagian penerbitan slip, 68 paket HRM lulus `go test -race -p 1 -count=1 -tags=integration ./...` dengan PostgreSQL disposable dan font PDF uji. `go vet ./...` lulus dan container HRM development sehat. Stack HTTPS terisolasi memakai image aktif dan membuktikan persetujuan independen serta satu event pengajuan posting. Penerbitan parsial dan replay diuji dengan fixture Finance serta DRM pada regresi PostgreSQL.

## Regresi build terkini

`go test -race -p=1 -count=1 -tags=integration -v ./...` selesai sukses pada PostgreSQL disposable. Output mencatat 70 paket bertes dan 288 tes tingkat atas lulus. Satu tes renderer dilewati karena font belum disetel pada eksekusi penuh. Uji renderer dijalankan ulang dengan DejaVu Sans dan `pdftotext`, lalu lulus. Suite handler payroll juga lulus ulang dengan font yang sama.

Suite Personnel dijalankan kembali dengan PostgreSQL dan Redis TLS disposable memakai password sementara. Query gRPC menolak revisi izin lama, menggunakan scope yang baru setelah bukti sesi disegarkan, dan menolak baca setelah pencabutan sesi di Redis. `go vet -tags=integration ./...` lulus. Uji Finance dan DRM di suite ini menggunakan fixture.
