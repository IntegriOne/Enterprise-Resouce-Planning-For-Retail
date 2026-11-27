# Penolakan koreksi saldo cuti

RPC RejectLeaveAdjustment memakai izin step-up dan mensyaratkan actor manusia yang berbeda dari pengusul. Alasan wajib serta pelaku dan waktu server tercatat atomik bersama event outbox. Koreksi yang ditolak tidak dapat disetujui, dihapus, atau diposting ke ledger. Baseline PostgreSQL menolak keputusan ganda dan metadata penolakan parsial.

Migrasi pra-rilis kembali tunggal di 000001. Efek versi 2–7 dipertahankan dalam baseline. Tes migrasi memeriksa up, pengulangan, down, up kembali, dan penolakan versi atau checksum tidak cocok. Integrasi Leave gRPC ke PostgreSQL lulus dengan policy IAM tiruan. Database development lama tidak dijalankan ulang atau dihapus dalam audit ini.

Suite integrasi serial seluruh modul dengan race detector lulus pada 67 paket di database PostgreSQL disposable. Tes unit dengan race detector lulus pada 48 paket. Build, go vet, dan validasi DBML juga lulus.
