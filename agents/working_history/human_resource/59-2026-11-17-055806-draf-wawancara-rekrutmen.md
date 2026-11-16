# Draf wawancara rekrutmen

Validator `hr_interviews` dipasang pada wiring HRM. Create/update/delete hanya berlaku untuk status `planned`, application/requisition/kandidat sah, masa retensi memadai, jadwal masa depan dan zona waktu IANA. Referensi lokasi/rapat tanpa verifikasi ditolak; draf dengan panelis atau feedback tidak dapat diubah/dihapus. Tes race seluruh modul, vet, build, gRPC–PostgreSQL disposable, dan Docker healthy lulus. Requisition tes masih fixture SQL dan policy tiruan; transisi lamaran, panelis, feedback, IAM/Gateway nyata belum tersedia.
