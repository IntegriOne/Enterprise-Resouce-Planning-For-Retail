# Pengajuan jurnal payroll

Tes memakai gRPC privat, PostgreSQL sementara dan policy tiruan.

| Kasus | Hasil |
| --- | --- |
| Pengajuan sebelum approval | FailedPrecondition |
| Pengajuan setelah approval | Requester, waktu dan outbox tersimpan, status tetap approved dan jurnal kosong |
| Pengajuan diulang | FailedPrecondition, tanpa outbox tambahan |
| Menimpa principal pengaju dengan SQL | PostgreSQL menolak perubahan |
