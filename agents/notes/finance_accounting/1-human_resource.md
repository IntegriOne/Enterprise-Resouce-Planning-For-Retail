# Verifikasi clearance Finance saat penutupan offboarding

HRM memerlukan operasi privat Finance untuk membaca keputusan clearance berdasarkan ID referensi, worker, dan payroll run yang terkait. Finance perlu membuktikan identitas objek, legal entity dan scope, keputusan final, versi atau waktu validasi, serta hubungan clearance dengan run terminasi. HRM mengirim identitas workload dan actor terverifikasi serta deadline. Referensi salah menghasilkan konflik bisnis. Jika verifikasi tidak tersedia, HRM mengembalikan Unavailable tanpa mengubah status atau outbox. Clearance Finance tidak membuktikan revoke IAM.

Snapshot HRM juga membawa ID rencana offboarding, tanggal efektif terminasi, ID hasil payroll pekerja, mata uang, digest output run, digest kalkulasi hasil, serta referensi pembayaran individual bila tercatat. Finance perlu mengikat clearance pada hasil individual tersebut dan menetapkan tahap payroll yang memenuhi syarat clearance. Status run agregat dan pembayaran pekerja dapat berbeda.

Uji positif lintas layanan memerlukan keputusan clearance yang cocok dengan snapshot. Uji penolakan mencakup hasil pekerja lain, digest berbeda, clearance dicabut, dan gangguan verifier. [Kontrak dan matriks uji HRM](../../grpc/hrm/38-kebutuhan-finance-clearance-offboarding.md).
