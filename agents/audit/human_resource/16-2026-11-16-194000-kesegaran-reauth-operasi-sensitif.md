# Audit kesegaran autentikasi operasi sensitif HRM

## Temuan

`Policy.Authorize` sebelumnya mewajibkan `MFAAuthenticated` untuk grant atau operasi sensitif, tetapi menerima sesi MFA lama tanpa jendela re-autentikasi 10 menit yang ditetapkan IAM. `SessionTokenRecord.IssuedAtUnix` tersedia dalam proyeksi Redis; IAM projector mengisinya dengan waktu login atau verifikasi MFA terbaru. Field ini memungkinkan pemeriksaan lokal HRM tanpa membaca tabel IAM.

## Perbaikan

HRM sekarang menolak operasi sensitif jika timestamp kosong, berada di masa depan, atau umur autentikasi melebihi 10 menit. Permission global, snapshot segar, identitas dan revisi sesi tetap diperiksa seperti sebelumnya. Pembacaan biasa yang tidak sensitif tetap dapat memakai sesi sah yang lebih tua.

## Bukti dan batas

Tes tabel mencakup auth baru/lama, timestamp hilang/masa depan, MFA hilang, flag sensitif pada grant, dan pembacaan biasa. `go fmt`, `go vet`, `go test -race -count=1 ./...`, `go build`, Docker build, serta health check container HRM lulus pada 16 November 2026. Ini belum E2E nyata dengan grant IAM dan Gateway; validator mutasi selain 14 resource yang telah terpasang tetap fail-closed.
