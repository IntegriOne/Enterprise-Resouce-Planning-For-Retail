# Scope baris rencana workforce

Grant baca legal_entity untuk workforce_plan_lines kini memakai badan hukum rencana induk. Repository memperoleh scope lewat join pada query yang sama dan menerapkan filter sebelum pagination dan total. Hasil query membawa atribut scope internal untuk pemeriksaan izin ulang per baris. Atribut tersebut tidak menjadi kolom skema atau field protobuf. Policy menolak pemeriksaan objek tanpa atribut badan hukum yang diperlukan. Izin baca rencana dan baris tetap terpisah.

Integrasi gRPC dengan PostgreSQL dan policy produksi membuktikan detail di luar scope menjadi NotFound, daftar hanya memuat baris yang diizinkan, dan finder rencana di luar scope mengembalikan total nol. Snapshot IAM dan caller transport disimulasikan. Redis dan mTLS runtime masih terbuka.
