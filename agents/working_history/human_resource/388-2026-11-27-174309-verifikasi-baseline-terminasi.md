# Verifikasi baseline terminasi

Memverifikasi SQL baseline HRM lokal pada PostgreSQL disposable. Pemasangan, skenario integritas HRM dan talent, rollback, serta pemasangan ulang lulus. Tes Go dengan race detector juga memverifikasi bahwa runner menolak riwayat dan checksum yang tidak cocok. Database aplikasi aktif tidak diubah.

Mencatat prasyarat kolom pembatalan penugasan dan dampak checksum lama pada daftar operasi terminasi. Commit `470490b` di `origin/main` sudah membawa relay HRM ke IAM. Baseline lokal dan operasi bisnis terminasi membutuhkan publikasi terkoordinasi. Untuk database berisi data yang dipertahankan, siapkan jalur pemindahan data sebelum rollout.

Regresi seluruh modul HRM lokal dengan race detector dan PostgreSQL disposable lulus, dengan 298 tes utama tercatat lulus. Cakupan meliputi offboarding, pembatalan penugasan mendatang, pemeriksaan identitas workload IAM, dan retry relay.
