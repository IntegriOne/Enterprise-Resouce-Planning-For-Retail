# Pembacaan referensi slip gaji pada hasil payroll

Tes gRPC dan PostgreSQL sementara lulus dengan race detector. Fixture database mensimulasikan status posted, referensi jurnal Finance, serta referensi versi DRM.

| Kasus | Hasil |
| --- | --- |
| Catat referensi slip pertama pada hasil run posted | Berhasil |
| Catat referensi saat run masih approved | PostgreSQL menolak |
| Ganti referensi slip yang sudah tercatat | PostgreSQL menolak |
| Tautkan versi DRM yang sama ke hasil pekerja lain | PostgreSQL menolak duplikasi |
| Panggil `GetPayrollResultByID` dengan policy staf tiruan | Referensi awal dikembalikan |

Cakupan ini memeriksa penyimpanan referensi dan pembacaan staf. Penerbitan PDF, adapter DRM, dan akses mandiri masih terbuka.

Baseline `000001_hrm` juga lulus siklus up → down → up pada database PostgreSQL sementara. Indeks `uq_hr_payroll_result_payslip` tercatat setelah penerapan ulang. Pemeriksaan ini memakai database baru. Database yang sudah mencatat checksum baseline lama memerlukan jalur migrasi yang disepakati sebelum menerima perubahan. [Pemeriksaan checksum migration runner](../../../backend/services/human_resource/db/migrate.go)

Tes PostgreSQL dengan race detector memakai `PublishPayslipForResult` dan fixture Finance–DRM untuk mengisi referensi pertama. Bukti yang salah ditolak sebelum penyimpanan. Bukti yang cocok membuat satu panggilan penyimpanan dan referensi terbaca melalui gRPC. Retry mengembalikan referensi yang sama tanpa memanggil penyimpanan lagi.

Regresi penerbitan satu run membuktikan slip pertama tetap tertaut saat adapter mengembalikan referensi tidak valid untuk hasil kedua. Retry setelah adapter pulih menyimpan hasil kedua saja, lalu replay berikutnya tidak melakukan panggilan penyimpanan tambahan.

Timestamp penerbitan satu jam setelah waktu server ditolak sebelum adapter PDF–DRM dipanggil. Penerbitan yang sudah berlaku tetap menghasilkan satu dokumen.
