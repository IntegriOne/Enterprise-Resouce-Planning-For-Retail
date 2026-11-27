# ListMyPayslips melalui gRPC dan PostgreSQL

Tes menggunakan verifier daftar berupa fixture, PostgreSQL sementara, dan race detector.

| Kasus | Hasil |
| --- | --- |
| Daftar pemilik dengan filter periode yang cocok | Satu slip dan total satu |
| Offset setelah halaman terakhir | Daftar kosong dan total tetap satu |
| Verifier mengirim hasil dan referensi dokumen pekerja lain | FailedPrecondition tanpa mengembalikan halaman |

Adapter produksi masih diperlukan untuk menyaring publikasi dan kesiapan dokumen sebelum menghitung total.

Verifier yang mengembalikan ID penerbitan berbeda untuk dokumen milik pekerja yang sama juga ditolak dengan FailedPrecondition.

Regresi gRPC menolak timestamp penerbitan di masa depan dengan FailedPrecondition dan menerima penerbitan yang sudah berlaku.

Tes gRPC memeriksa ID dan kode periode yang diambil dari HRM. Halaman dari verifier yang berisi slip di luar filter periode ditolak dengan FailedPrecondition.

Regresi riwayat kerja memakai dua hasil payroll dengan worker berbeda yang terhubung ke principal dan party orang yang sama. Adapter fixture mengharuskan kedua ID worker tersedia. Pada limit satu, offset 0 dan 1 mengembalikan masing-masing satu hasil dengan total dua. Offset 2 kosong dan total tetap dua. Ketika party orang berbeda, HRM mengembalikan `PermissionDenied`. Tes gRPC dengan PostgreSQL, renderer PDF, dan race detector lulus.

Fixture tambahan membuat 1.001 hubungan kerja lama untuk principal dan party yang sama. Daftar tetap mengembalikan slip dari hubungan kerja yang diminta ketika jumlah ID melampaui batas satu filter repository. HRM membaca hasil dengan beberapa filter worker yang terpisah.
