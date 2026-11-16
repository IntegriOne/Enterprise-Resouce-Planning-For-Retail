# Audit runner migrasi berurutan HRM

## Temuan

`RunMigrationsUp` sebelumnya mengenal `000001` saja dan menolak versi lain dengan query tetap. Approval requisition memerlukan provenance pengusul melalui migrasi aditif; menulis ulang baseline akan menggagalkan checksum instalasi yang sudah ada. Generator DBML saat ini masih menimpa baseline bila dijalankan.

## Perbaikan

- Runner kini memakai daftar migrasi berurutan, memeriksa semua versi terpasang, checksum, dan celah riwayat sebelum menerapkan versi yang belum ada.
- Advisory lock, satu transaksi, pemeriksaan skema lama tanpa tracker, dan sanitasi error tetap berlaku.
- Baseline `000001` tidak diubah. Daftar saat ini masih berisi versi 1; ini persiapan aman untuk `000002`, bukan aktivasi approval requisition.

## Bukti dan batas

Tes PostgreSQL pada skema terisolasi di database `*_test` lulus untuk instalasi awal, pengulangan, versi asing, dan checksum rusak. `go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...`, build Docker dan healthcheck HRM lulus. Tes belum mencakup upgrade ke versi 2, karena migrasi itu belum dibuat. Langkah berikutnya ialah menjaga baseline immutable pada generator, menambah migrasi provenance requisition, lalu command submit/approve/open dengan bukti owner dan SoD.
