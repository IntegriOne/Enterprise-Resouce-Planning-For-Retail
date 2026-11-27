# Uji HTTPS Query job, job family, dan posisi dengan data

Role uji dipasang melalui GraphQL IAM. Job family dan job sintetis dibuat di PostgreSQL HRM disposable untuk diuji sebagai data katalog. Posisi terkait dibuat dan diaktifkan lewat GraphQL dengan badan hukum dan unit organisasi dari Foundation.

| Query | Hasil |
| --- | --- |
| `hrmPosition` | ID posisi yang sama dan status aktif |
| `hrmJob` | ID job dan ID family yang sesuai |
| `hrmJobFamily` | ID dan kode family yang sesuai |
| `hrmJobFamilies` | Satu family dan total satu |
| `hrmJobs` dengan filter family yang ada | Satu job dan total satu |
| `hrmJobs` dengan filter UUID sintetis yang tidak ada | Daftar kosong, total nol |

Percobaan awal menerima `UNAVAILABLE` dari Foundation saat membuat prasyarat segera setelah startup. Pengulangan setelah jeda inisialisasi berhasil. Database dan Redis uji dihapus setelah development kembali sehat. [Uji perubahan posisi](112-hrm-query-mutation-lifecycle-posisi-mfa.md).
