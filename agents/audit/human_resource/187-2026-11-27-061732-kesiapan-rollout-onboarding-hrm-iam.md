# Kesiapan rollout onboarding HRM dan IAM

Build Gateway, HRM, IAM, dan Enterprise Foundation dari checkout berhasil menjalankan onboarding lewat Caddy HTTPS pada stack terisolasi. IAM menggunakan katalog dari source tanpa overlay. Browser mengajukan akun, IAM menyimpan satu akun dan satu penerimaan, lalu browser HRM menampilkan status tertaut. Role serta penugasan izin dilakukan lewat GraphQL. [Alur browser](../../frontend/e2e/human_resource/1-mutation-pengajuan-akun-pekerja.md).

Pemeriksaan development menunjukkan dua replika Gateway mengaktifkan HRM. HRM serta IAM mengaktifkan rute privat ke Enterprise Foundation. Checksum baseline HRM development cocok dengan file checkout. Gateway development masih mengembalikan `GRAPHQL_VALIDATION_FAILED` untuk `hrmWorkerEmployeeAccountRequest`. Pemeriksaan baca-saja pada database IAM development menunjukkan tabel `employee_provisioning_requests` belum ada.

Rollout bersama memerlukan database IAM dengan schema provisioning yang kompatibel. Pertahankan data dan versi migrasi IAM yang sudah terpasang saat menyiapkan jalur upgrade. Hasil ini belum membuktikan upgrade pada database development berisi data. Stack development dipertahankan selama uji dan audit.

## Rehearsal database IAM dengan baseline terpasang

Database IAM `*_test` memakai SQL baseline dari commit `e28491c`, tracker versi 1 bersih, dan satu akun fixture aktif. Selisih SQL terhadap baseline checkout hanya blok tabel provisioning dan dua indeks. Blok yang sama diterapkan dua kali dalam transaksi. Pemeriksaan membuktikan akun lama tetap aktif, tracker tetap versi 1, dan aplikasi terbaru dapat memproses onboarding melalui HTTPS dan browser. IAM menyimpan satu akun baru serta satu penerimaan. Pemeriksaan akhir kembali menemukan akun lama utuh.

Rehearsal memakai database disposable, tanpa data development. Jalur aditif memakai SQL kanonik pada baseline `000001`. Runner lokal `/tmp/erp-cross-service.UUJx1n/run-existing-iam-onboarding.py` selesai dengan status 0.

## Rollout development

Blok DDL provisioning dari baseline checkout diterapkan dalam transaksi pada database `iam` development. Tabel dan kedua indeks tersedia setelah commit. Jumlah akun sebelum dan sesudah DDL tetap delapan dan tracker tetap versi 1 bersih.

Container IAM dan HRM diperbarui, disusul dua replika Gateway dan frontend. Masing-masing lulus healthcheck. Enterprise Foundation yang sudah berjalan tetap sehat. Compose dijalankan dengan `HRM_GATEWAY_ENABLED=true`; flag perlu disertakan pada recreate berikutnya selama default Compose masih false.

Caddy HTTPS mengenali `hrmWorkerEmployeeAccountRequest` dan mengembalikan `UNAUTHENTICATED` pada sesi anonim. Empat panggilan `hrmWorkers` dari superadmin berhasil. Role uji, grant, dan assignment dibuat dan dibersihkan lewat GraphQL. Percobaan pertama runner sempat salah membaca ID assignment. Assignment dan role yang tertinggal dicabut melalui GraphQL sebelum pengujian ulang berhasil. Provisioning lengkap diuji di stack disposable, berikut rehearsal skema lama; development memakai uji baca.

## Recreate Compose tanpa flag shell

Checkout development kini membaca `HRM_GATEWAY_ENABLED=true` dari `.env` lokal yang dikecualikan dari Git. Kedua replika dibuat ulang secara bergiliran melalui Compose tanpa flag shell dan lulus healthcheck. Caddy HTTPS kembali melayani empat permintaan daftar pekerja dengan superadmin. Pemberian dan pencabutan grant uji dilakukan melalui GraphQL. Query status pengajuan anonim tetap ditolak.
