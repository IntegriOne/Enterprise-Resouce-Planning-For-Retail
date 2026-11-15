# Verifikasi penyatuan Talent Management dalam HRM

16 November 2026 WIB. Scope: skema dan baseline `human_resource`, dokumentasi domain, tanpa runtime atau migrasi database bisnis. Keputusan owner menyatukan Talent dan recruitment dalam HRM; worker master tunggal.

## Perubahan

69→86 tabel, 130 FK internal, 147 CHECK DBML, 9 enum, 12 kelompok diagram. Baseline PostgreSQL menambahkan total lima exclusion constraint, 14 trigger, delapan function. Detail fungsi, alasan dan batas aturan: [desain HRM–Talent](../../../backend/services/human_resource/docs/talent-management.md).

## Verifikasi yang dijalankan

Generator CLI DBML 12.0.0 berhasil. `db/test/verify_schema.sh` menjalankan baseline up, tes HRM lama, tes Talent, down, dan up ulang pada database disposable Docker. Semua 70 skenario negatif lulus: 38 sebelumnya dan 32 berikut.

1. rubrik dimensi terpakai dibekukan.
2. manifest rubrik disahkan immutable.
3. finalisasi tanpa skor wajib.
4. dimensi dari rubrik lain ditolak.
5. skor talent di luar rubrik.
6. skor talent final immutable.
7. bukti asesmen final immutable.
8. kalibrasi lintas siklus.
9. skor awal kalibrasi palsu.
10. keputusan kalibrasi append-only.
11. pool memakai asesmen pekerja lain.
12. pool aktif tanpa approval.
13. keanggotaan pool aktif bertumpuk.
14. jalur karier menunjuk jabatan yang sama.
15. aspirasi dan tujuan jalur berbeda.
16. rekomendasi dengan assignment pekerja lain.
17. draft langsung approved.
18. usulan berubah sesudah submit.
19. approved tanpa keputusan nyata.
20. pengusul menyetujui usulan sendiri.
21. keputusan defer dianggap approve.
22. applied tanpa bukti penempatan dan event.
23. keputusan promosi immutable.
24. event hire tidak membuktikan promosi.
25. tanggal event tidak sesuai keputusan promosi.
26. event promosi milik pekerja lain.
27. status applied dikembalikan.
28. bukti penerapan ditimpa.
29. review menilai goal pekerja lain.
30. hasil KPI historis immutable.
31. koreksi pengukuran lintas goal.
32. mapping kursus kompetensi ganda.

Fixture positif: asesmen final berskor lengkap; kalibrasi menyimpan skor awal dan hasil terpisah; pool aktif dengan approval; aspirasi terikat tujuan; defer→approve oleh aktor berbeda; **approval tidak mengubah penempatan resmi**; transaksi penerapan menutup assignment lama, mencatat assignment/event baru dan menandai applied; kursus mendukung dua kompetensi. Database sementara dibersihkan.

## Batas klaim

Tes ini menguji constraint dan trigger SQL, bukan IAM/ABAC atau E2E GraphQL. Source authority, seluruh tahap approval, status pekerja/posisi, budget/kompensasi, kalibrasi pada sesi terbuka, rubrik lengkap dan total bobot, rekonsiliasi KPI, koreksi employment sesudah applied serta outbox atomik masih memerlukan service dan tes concurrency. Tidak ada microservice Talent baru atau pemindahan ownership Foundation/IAM/Finance/DRM/Workforce.
