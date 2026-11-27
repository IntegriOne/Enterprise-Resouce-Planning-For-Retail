# Memulai wawancara

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy tiruan dan panelis fixture.

| Kasus | Hasil |
| --- | --- |
| Mulai sebelum jadwal | FailedPrecondition |
| Mulai tanpa panelis | FailedPrecondition, tahap lamaran tetap shortlisted |
| Mulai dalam jadwal dengan panelis | Wawancara in_progress dan lamaran interview |
| Mengulang mulai | FailedPrecondition |
| Riwayat tahap | Satu entri transisi ke interview |
