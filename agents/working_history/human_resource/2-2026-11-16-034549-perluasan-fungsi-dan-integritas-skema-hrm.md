# Perluasan skema HRM berdasarkan riset fungsi HR

16 November 2026 WIB. Memakai skill Database Engineer `cognitive-schema-mapper` dan rujukan resmi CIPD untuk memetakan recruitment serta fungsi HR lainnya. DBML aktif `human_resource` diperluas 45→69 tabel dengan 11 kelompok bisnis dan satu integrasi. Ditambahkan perencanaan tenaga kerja, kompetensi, posting/asesmen, offboarding, grup/aturan payroll, ledger cuti, development plan, kebijakan/PKB, layanan, engagement, access request serta outbox/inbox.

Diperkuat FK kepemilikan pekerja, scope payroll, persamaan total, approval, bukti settlement, temporal overlap dan snapshot immutable. Baseline tunggal 000001 up/down digenerate memakai CLI DBML 12.0.0. Pengujian PostgreSQL Docker 18.6: 38 skenario negatif serta fixture positif lulus; siklus up/down/up berhasil dan database sementara dihapus. Compose config lulus.

Dokumentasi desain dan konteks mencatat batas service/otorisasi/aturan hukum yang belum diimplementasikan. Tidak mengubah database bisnis atau pekerjaan Enterprise Foundation/Gateway milik developer lain.
