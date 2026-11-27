# Uji mutation draf workforce dengan referensi badan hukum yang tidak ada

Uji HTTPS memakai superadmin dengan role uji global yang hanya berisi `human_resource.workforce_plans.create`. Setelah assignment dan login ulang, permintaan `hrmCreateWorkforcePlan` memakai UUID sintetik valid untuk badan hukum yang tidak ada dan periode draf valid.

| Pemeriksaan | Hasil |
| --- | --- |
| Query Enterprise Foundation untuk badan hukum aktif | Total nol |
| Mutation Gateway | `CONFLICT` |
| Log gRPC HRM | `WorkforcePlanningService/CreateWorkforcePlan` mengembalikan `FailedPrecondition` |
| Jumlah draf pada PostgreSQL HRM setelah permintaan | Nol |
| Pencabutan assignment, penghapusan role dan logout | Berhasil |

Percobaan pertama berhenti pada rate limit sebelum role dibuat. Uji berhasil dijalankan setelah jendela 60 detik berakhir. Kasus ini membuktikan penolakan referensi lintas layanan tanpa membuat draf. Uji penyimpanan positif masih memerlukan badan hukum aktif dari Enterprise Foundation dan data HRM disposable. [Matriks mutation sebelumnya](96-hrm-mutation-seluruh-schema-superadmin-https.md).

## Uji positif dengan database disposable

Enterprise Foundation dan HRM menggunakan database baru berakhiran `_test` melalui override Compose sementara. IAM tetap memakai image pengujian lokal. Superadmin menerima role sementara dengan izin `workforce_plans.create`, `workforce_plans.read`, dan `workforce_plans.update`. Session diperbarui sebelum permintaan.

| Kasus HTTPS | Hasil |
| --- | --- |
| `createLegalEntity` melalui API Enterprise Foundation dengan status aktif | Berhasil |
| `hrmCreateWorkforcePlan` dengan badan hukum tersebut | Berhasil dengan status `draft` |
| Ulangi kode, versi, dan badan hukum yang sama | `CONFLICT` |
| `hrmUpdateWorkforcePlan` untuk mengoreksi skenario | Berhasil |
| `hrmWorkforcePlan` membaca ulang ID yang sama | Mengembalikan skenario `E2E-UPDATED` |
| Query baca-saja PostgreSQL untuk status dan skenario | Tepat satu baris `draft|E2E-UPDATED` |
| Cabut assignment, hapus role, dan logout | Berhasil |

Kedua layanan dipulihkan ke database development semula dan kembali sehat. Kedua database uji dihapus.

## Uji ulang dengan katalog resmi

Regresi stack disposable memakai IAM yang dibangun dari manifest HRM resmi. Superadmin memasang tiga izin `workforce_plans.create`, `workforce_plans.read` dan `workforce_plans.update` melalui role sementara GraphQL. Pembuatan badan hukum aktif di EF, pembuatan dan koreksi draf, penolakan draf duplikat dan pembacaan ulang skenario lulus. Permission dan role uji dibersihkan via GraphQL. Persetujuan anggaran Finance berada pada alur terpisah.
