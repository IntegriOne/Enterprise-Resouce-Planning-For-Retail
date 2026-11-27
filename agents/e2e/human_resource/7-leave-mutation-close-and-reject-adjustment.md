# Penutupan hak cuti dan penolakan koreksi

Tes gRPC menggunakan PostgreSQL sementara, race detector, dan policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Setujui koreksi yang membuat saldo -1 | FailedPrecondition. Outbox koreksi hanya memuat event pengajuan. |
| Tolak koreksi yang gagal disetujui dengan alasan INSUFFICIENT_BALANCE | Berhasil |
| Ajukan koreksi ketika saldo nol dan hak cuti belum ditutup | Berhasil |
| Tutup hak cuti setelah akhir periode dengan koreksi belum diputuskan | Berhasil |
| Setujui koreksi setelah penutupan | FailedPrecondition |
| Tolak koreksi setelah penutupan dengan alasan PERIOD_CLOSED | Berhasil |
| Posting atau hapus hak cuti yang telah ditutup | FailedPrecondition |
| Hitung ledger dan outbox hak cuti setelah penolakan | Dua entri ledger dan empat event outbox hak cuti |
