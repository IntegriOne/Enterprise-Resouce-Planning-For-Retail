# Uji HTTPS transfer assignment primer

Dua akun fixture dengan MFA memakai IAM, Foundation, HRM dan Redis disposable. Posisi sumber dan tujuan diaktifkan melalui API. Assignment primer sumber sudah disetujui dan diterapkan. Job, status aktif pekerja, dan verifikasi identitas akun kedua memakai SQL fixture pada database uji.

| Kasus | Hasil |
| --- | --- |
| `hrmSubmitWorkerAssignmentTransfer` untuk posisi tujuan aktif pada badan hukum yang sama | Status `submitted` |
| Ulangi pengajuan dengan idempotency key dan input yang sama | ID permintaan tetap |
| Pemohon mencoba menyetujui transfer sendiri | `FORBIDDEN` |
| Principal independen menyetujui transfer pada tanggal efektif | Status `applied` dan ID assignment baru |
| Ulangi keputusan yang sama | ID assignment baru tetap |
| PostgreSQL | Assignment primer menunjuk assignment baru, pemohon dan pengambil keputusan berbeda, dan hanya ada dua assignment, sumber dan tujuan |

Percobaan awal sempat menerima `UNAVAILABLE` dari IAM setelah startup. Pengulangan menunggu inisialisasi selesai dan memakai status kontrak `submitted`. Suite lulus, database serta Redis uji dihapus, dan stack development kembali sehat. [Uji assignment primer](113-hrm-query-mutation-assignment-primer-mfa.md).
