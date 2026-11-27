# Pengujian pembuatan dan koreksi pekerja melalui HTTPS

Uji memakai superadmin, role sementara dengan izin `workers.create`, `workers.update`, `workers.read`, dan `offboarding_plans.create`, serta dua database disposable untuk Enterprise Foundation dan HRM. Semua mutation memakai sesi HTTPS dan CSRF yang sah.

| Kasus | Hasil |
| --- | --- |
| Membuat badan hukum aktif dan party individu aktif melalui API Foundation | Berhasil |
| `hrmCreateWorker` dengan kedua referensi | Berhasil, status `preboarding` |
| Mengulang pembuatan dengan input yang sama | `CONFLICT` |
| `hrmUpdateWorker` mengoreksi tanggal masuk | Berhasil |
| `hrmWorker` membaca ID pekerja yang sama | Tanggal terkoreksi, `2026-12-02` |
| `hrmCreateOffboardingPlan` untuk pekerja preboarding | `CONFLICT` |
| Pemeriksaan PostgreSQL | Tepat satu pekerja `preboarding` dengan tanggal terkoreksi. Nol rencana offboarding. |

Assignment dicabut, role dihapus, dan sesi ditutup. Enterprise Foundation dan HRM kembali sehat pada database development semula. Kedua database disposable dihapus. [Uji workforce plan](97-hrm-mutation-referensi-badan-hukum-workforce.md).
