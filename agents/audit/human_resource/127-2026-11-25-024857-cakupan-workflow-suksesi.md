# Cakupan workflow suksesi HRM

Kontrak saat ini menyediakan CRUD draf rencana dan kandidat serta finder. Validator memeriksa posisi aktif, badan hukum kandidat, tanggal asesmen, bukti readiness, dan kerahasiaan rencana. Tes validator serta integrasi gRPC dengan PostgreSQL sementara dan race detector lulus.

Persetujuan rencana, nominasi resmi, dan penarikan resmi tercantum pada [matriks izin HRM](../../roles-permissions/human_resource/1-roles-permissions.md) sebagai rancangan. Ketiganya belum punya RPC atau pemetaan runtime. Persetujuan perlu mandat panel dan pemeriksaan konflik kepentingan. Bukti pengujian saat ini hanya mencakup nominasi draf serta penjagaan data yang sudah disetujui, sehingga belum membuktikan workflow persetujuan berjalan.
