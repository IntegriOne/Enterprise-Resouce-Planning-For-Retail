# Dua operator memproses satu pengajuan employee

Superadmin menyiapkan role dan grant melalui GraphQL dalam stack HTTPS disposable. Operator kedua dibuat melalui pengajuan HRM dan penautan IAM. Verifikasi identitas loginnya disiapkan sebagai fixture IAM. Kontrak bertanda tangan menggunakan fixture HRM. Kedua operator login dengan MFA pada sesi terpisah.

Dua koneksi HTTPS mengirim `createEmployeeUser` secara bersamaan untuk pekerja dan pengajuan kedua yang sama, dengan kredensial akun baru yang berbeda. Tepat satu mutation berhasil dan satu menerima `CONFLICT`. PostgreSQL menyimpan tepat satu akun dari persaingan tersebut. Principal pada pekerja cocok dengan akun dari mutation yang berhasil.

Runner `/tmp/erp-cross-service.UUJx1n/run-two-operators.py` keluar dengan status 0. Stack dan database disposable dibersihkan.
