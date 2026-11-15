# Penyatuan Talent Management dalam HRM

Keputusan owner 16 November 2026: Talent Management dan recruitment tetap dalam satu HRM. Memperluas DBML 69→86 tabel dengan rubrik/asesmen talent, kalibrasi, talent pool, jalur/aspirasi karier, rekomendasi/keputusan promosi, riwayat KPI dan pemetaan banyak kompetensi per kursus. Menghubungkan suksesi ke asesmen worker yang sama.

Baseline tunggal 000001 diselaraskan. Ditambahkan perlindungan rubrik terpakai/asesmen final, keputusan append-only, overlap pool serta state machine promosi. Approval tidak mengubah assignment; applied membutuhkan keputusan dan bukti penempatan/event yang sesuai.

Verifikasi PostgreSQL Docker: 70 skenario negatif (38 HRM + 32 Talent), fixture positif dan up/down/up lulus; database sementara dihapus. Dokumentasi HRM–Talent, ringkasan management, konteks dan audit diperbarui. Runtime Go/GraphQL/IAM, otorisasi/mandat panel, transaksi outbox serta concurrency/E2E belum dibuat. Tidak mengubah pekerjaan developer lain atau database bisnis.
