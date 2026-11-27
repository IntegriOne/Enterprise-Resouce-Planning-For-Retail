# CancelServiceRequest

Tes `TestHRServiceRequestAndEventsWithPostgreSQL` memakai client gRPC melalui `bufconn`, PostgreSQL disposable, dan fixture policy yang memilih principal. Eksekusi dengan race detector lulus.

| Kasus | Respons gRPC | Perilaku |
| --- | --- | --- |
| Principal berbeda mencoba membatalkan tiket terbuka | `PermissionDenied` | Tiket tetap dapat dibatalkan oleh pengusul. |
| Dua permintaan pengusul membatalkan tiket terbuka secara paralel | Satu `OK` dan satu `FailedPrecondition` | Baca ulang menemukan status `cancelled` dan PostgreSQL menyimpan satu event `hr_service_requests.cancel`. |
| Pengusul mengulang pembatalan | `FailedPrecondition` | Transisi kedua ditolak. |
| Principal dengan akses fixture mencoba menugaskan tiket yang dibatalkan | `FailedPrecondition` | Penugasan ditolak. |
| Pengusul mengubah status menjadi `cancelled` melalui update CRUD | `FailedPrecondition` | Field mask tidak dapat menjalankan transisi pembatalan. |

Wire Gateway dan grant IAM resmi belum diuji dalam fixture ini. [Handoff HR Service](../../grpc/hrm/10-operasi-hrservice.md)
