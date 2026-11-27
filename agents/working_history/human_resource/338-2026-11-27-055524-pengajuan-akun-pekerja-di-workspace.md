# Pengajuan akun pekerja di workspace HRM

Menambah halaman Pekerja untuk mencari nomor pekerja, melihat nama dan badan hukum pada detail jika diizinkan, mengajukan akun, serta memeriksa statusnya. ID teknis dipakai sebagai parameter GraphQL. Query `hrmWorkerEmployeeAccountRequest` mencari pengajuan berdasarkan pekerja dan membaca status akhir dari IAM. Guard workspace server dan browser menerima izin baca pekerja. Izin pengajuan dan baca status divalidasi oleh Gateway.

Pengujian HTTPS dengan superadmin dan role sementara melalui GraphQL membuktikan pengajuan dari browser, penolakan sebelum MFA pada antrean IAM, penolakan pekerja yang berbeda, idempotensi penerbitan akun, serta perubahan status HRM menjadi `applied`.
