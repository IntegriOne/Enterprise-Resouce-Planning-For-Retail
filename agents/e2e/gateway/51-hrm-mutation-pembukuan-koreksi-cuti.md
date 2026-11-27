# Pengujian Mutation pembukuan koreksi cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Izin post, MFA dan CSRF | Sumber HRM dan digest stabil diteruskan. Saldo snapshot baru dikembalikan |
| Tanggal tidak sah | INVALID_ARGUMENT |
| Izin persetujuan saja | FORBIDDEN |

Tes fixture memakai race detector. Pada runtime, HRM memvalidasi kecocokan koreksi yang disetujui, saldo dan duplikasi, serta mencatat ledger dan perubahan saldo secara atomik.

Regresi gRPC HRM dengan PostgreSQL sementara memakai format sumber dan digest yang sama dengan Gateway. Pembukuan pertama memperbarui saldo menjadi 11.0000 dan revisi menjadi 3. Permintaan identik ditolak dengan AlreadyExists. Pembacaan ulang mengonfirmasi saldo dan revisi tetap sama. Race detector dan vet paket integrasi lulus.
