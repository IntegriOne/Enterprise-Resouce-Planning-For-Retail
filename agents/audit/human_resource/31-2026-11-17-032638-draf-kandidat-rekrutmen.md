# Audit draf kandidat rekrutmen HRM

## Cakupan dan temuan

`hr_candidates` sebelumnya memiliki CRUD gRPC dan repository tetapi belum mempunyai validator produksi sehingga mutasi ditolak oleh `Integrity`. Validator sekarang membuka pencatatan kandidat lokal aktif dengan nomor unik, versi pemberitahuan privasi, waktu persetujuan dan batas retensi yang sah. Nomor, persetujuan, retensi, waktu/pembuat dan status tidak dapat diubah lewat CRUD; nama pilihan dapat dikoreksi. Hapus ditolak jika ada `hr_applications`.

Referensi party/kontak Foundation dan sumber eksternal tetap ditolak karena belum ada verifikasi owner. Metadata persetujuan berasal dari caller dan belum membuktikan penyajian pemberitahuan privasi. Lamaran, seleksi dan hiring belum dibuka. Policy tes adalah tiruan; grant IAM, mTLS/HMAC runtime, Gateway dan scope nyata belum dibuktikan.

## Bukti

- `go fmt ./...`, `go vet ./...`, `go test -race -count=1 ./...`, dan `go build ./...` lulus pada modul HRM.
- `docker compose build human-resource` lulus dan `docker compose up -d --wait human-resource` menghasilkan container HRM `healthy`.
- Tes gRPC `bufconn`→PostgreSQL 18 pada database disposable `*_test` lulus untuk create, duplikasi, persetujuan hilang, referensi kontak tak terverifikasi, list berhalaman, koreksi nama, penolakan perubahan versi notice, outbox dan delete kandidat tanpa lamaran.
- Validator produksi terpasang pada 30/78 resource berkode mutasi; 48 resource lain tetap fail-closed.
