# Persetujuan rencana offboarding

Tes memakai PostgreSQL sementara, policy tiruan, dan race detector. Tautan principal pekerja disiapkan langsung pada fixture.

| Kasus | Hasil |
| --- | --- |
| Mulai rencana sebelum disetujui | FailedPrecondition |
| Pembuat menyetujui rencana sendiri | PermissionDenied |
| Subjek pekerja menyetujui rencana yang dibuat orang lain | PermissionDenied |
| Penyetuju independen menyetujui draf | Status approved, waktu dan principal tersimpan |
| Persetujuan diulang | FailedPrecondition |
