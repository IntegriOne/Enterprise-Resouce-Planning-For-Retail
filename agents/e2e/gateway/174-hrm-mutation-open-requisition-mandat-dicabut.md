# Uji HTTPS pembukaan requisition dan pencabutan mandat

Gateway, HRM, IAM, dan Enterprise Foundation berjalan pada stack disposable. HRM memakai klien mTLS ke listener privat IAM. Superadmin MFA menjadi operator pembukaan. Peninjau pada requisition approved adalah akun employee terpisah. Role dan grant diatur melalui GraphQL IAM. Status awal requisition disiapkan di PostgreSQL HRM disposable.

| Kasus | Hasil |
| --- | --- |
| Operator pembukaan memiliki grant yang sesuai, peninjau belum punya mandat unit | Pembukaan ditolak dengan `AUTHORIZATION`. |
| Berikan grant unit yang tepat kepada peninjau tercatat | Pembukaan mencapai pemeriksaan anggaran dan mengembalikan `UNAVAILABLE` karena verifier Finance belum terpasang. |
| Cabut assignment peninjau melalui GraphQL | Percobaan berikutnya kembali ditolak dengan `AUTHORIZATION`. |
| Periksa status | Requisition tetap approved. Draf submitted pada skenario approval juga tidak berubah. |

Unit test memastikan penolakan mandat tidak memanggil verifier anggaran. Race test service rekrutmen dan go vet lulus.

Regresi dengan build checkout aktif dan sertifikat overlay development lulus pada stack disposable. Setelah grant unit diberikan lewat GraphQL IAM, `openRequisition` mencapai pemeriksaan Finance. Setelah grant peninjau tercatat dicabut, pembukaan kembali ditolak pada pemeriksaan IAM. Tes listener IAM, handler mandat, grant efektif, dan korelasi respons HRM juga lulus dengan race detector.
