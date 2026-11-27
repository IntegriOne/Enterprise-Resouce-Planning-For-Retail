# Uji HTTPS guard penautan principal pekerja

Superadmin fixture memakai sesi TOTP dan izin `human_resource.workers.link_principal` yang diberikan melalui mutasi GraphQL IAM. Target employee dibuat dan diaktifkan melalui API IAM. Status aktif pekerja dan verifikasi identitas target disiapkan pada database disposable. Runtime HRM standar tidak memasang verifier bukti provisioning.

| Kasus | Hasil |
| --- | --- |
| Tautkan worker ke principal operator sendiri | `FORBIDDEN` |
| Tautkan ke employee lain dengan ID bukti sintetis | `UNAVAILABLE` |
| Pemeriksaan PostgreSQL | `iam_principal_reference_id` tetap kosong dan nol event lifecycle `principal_linked` |

Database dan Redis uji dibersihkan setelah layanan development kembali sehat. Jalur berhasil memerlukan bukti provisioning dari IAM yang mengikat akun, worker, party, dan pemohon. [Handoff IAM](../../notes/iam_management/1-human_resource-permission-handoff.md).
