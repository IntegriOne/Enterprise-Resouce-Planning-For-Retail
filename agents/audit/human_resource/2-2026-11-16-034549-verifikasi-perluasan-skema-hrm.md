# Verifikasi perluasan fungsi dan integritas skema HRM

Tanggal: 16 November 2026 WIB. Scope: `backend/services/human_resource/` saja. Permintaan owner: riset internet fungsi HR, sesuaikan dan update DBML. Skill: Database Engineer `cognitive-schema-mapper`; bukti pemetaan dan sumber terdapat pada [dokumen desain](../../../backend/services/human_resource/docs/schema-design.md).

## Hasil struktural

- Sebelum: 45 tabel, 9 enum, 56 FK, 43 CHECK pada export audit awal.
- Sesudah: 69 tabel, 9 enum, 95 FK internal, 113 CHECK DBML, 12 kelompok diagram.
- Tambahan baseline: empat exclusion constraint, lima trigger dan tiga function.
- Tidak ada FK lintas service, tenant master, role/grant lokal atau pengambilalihan roster/absensi. Identitas worker berbeda dari principal IAM.
- DBML + invariant PostgreSQL menghasilkan baseline 000001 up/down yang bisa diregenerasi. Belum diterapkan pada database aplikasi.

## Bukti pengujian aktual

`bash backend/services/human_resource/db/test/verify_schema.sh` lulus pada PostgreSQL Docker 18.6. Nama database unik `hrm_schema_*_test`, up → test → down → up → cleanup. Semua fixture di dalam transaksi di-rollback. Script memverifikasi SQLSTATE penolakan, bukan hanya error umum.

38 kasus negatif yang lulus:

1. Primary assignment overlap pada batas tanggal inklusif.
2. FTE lebih dari 100.
3. Primary assignment milik pekerja lain.
4. Kompensasi menunjuk assignment pekerja lain.
5. Kompensasi overlap lintas versi komponen yang berkode sama.
6. Kompensasi aktif tanpa penyetuju.
7. Profil payroll memakai employer berbeda dari grup.
8. Profil payroll overlap.
9. Run koreksi tanpa asal.
10. Run memakai grup berbeda dari periode.
11. Hasil payroll memakai profil pekerja lain.
12. Hasil payroll memakai profil grup lain.
13. Net tidak sesuai gross/reimbursement/deductions.
14. Approval payroll tanpa rincian.
15. Tipe snapshot baris berbeda dari komponen.
16. Approval tanpa aktor.
17. Approval dengan total tidak sama dengan rincian.
18. Perubahan baris setelah approval.
19. Perubahan nilai hasil setelah approval.
20. Pengembalian status approved ke draft.
21. Overwrite bukti payslip yang sudah tercatat.
22. Status paid tanpa bukti Finance.
23. Pengubahan artefak aturan payroll immutable.
24. Persamaan saldo cuti salah.
25. Pengubahan ledger cuti.
26. Pemakaian cuti dengan source event/line ganda.
27. Reversal lintas entitlement.
28. Entry usage dengan jumlah positif.
29. Snapshot metrik dengan scope NULL ganda.
30. Kasus hubungan karyawan dengan worker yang tidak ada.
31. Access request menunjuk event pekerja lain.
32. Access request applied tanpa keputusan IAM.
33. Level kompetensi di luar 1–5.
34. Suksesi memakai development plan pekerja lain.
35. Task offboarding selesai tanpa konfirmasi owner.
36. Waiver offboarding tanpa alasan/aktor.
37. Inbox event Finance ganda.
38. Entitlement jenis cuti sama bertumpuk lintas versi kebijakan.

Fixture positif mencakup periode primary bersebelahan, secondary paralel, dua komponen kompensasi berbeda, grup payroll badan hukum berbeda, tanggal bayar sebelum akhir bulan, net di atas gross akibat reimbursement, approval/posted/paid lengkap, penambahan metadata payslip setelah approval, dan versi metrik baru. Baseline down berhasil menghapus seluruh tabel HR; baseline up ulang berhasil. `docker compose config --quiet` lulus.

## Batas hasil

Ini verifikasi skema/SQL, bukan benchmark, concurrency test atau E2E HTTP GraphQL. Tidak ada runtime/image HRM pada folder aktif untuk build/deploy. Dependency hukum, formula/rounding, validitas bisnis lintas entitas, source authenticity, ABAC, layanan eksternal, transaksi saldo cuti dan relay/inbox masih harus dikerjakan dan diuji. Lihat daftar lengkap pada dokumen desain. Saldo bucket yang aritmetis benar belum otomatis cocok dengan seluruh SUM ledger; implementasi service harus membuktikannya.

Penambahan cakupan tidak berarti semua 69 tabel harus diaktifkan saat F1 awal. Prioritas tetap worker/kontrak/assignment/onboarding-offboarding, lalu operasional payroll/cuti, kemudian pengembangan dan analytics sesuai kebutuhan. Sumber riset menunjukkan fungsi, bukan jumlah divisi wajib.
