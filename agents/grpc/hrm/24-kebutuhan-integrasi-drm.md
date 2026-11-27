# List operasi

Operasi berikut bergantung pada bukti dokumen atau artefak aturan yang dikelola DRM. Kontrak privat dan adapter produksi masih perlu disepakati dengan pemilik DRM.

## Penjelasan operasi

| Operasi | Butuh GraphQL | Bukti minimum dari DRM |
| --- | --- | --- |
| `PublishJobPosting` | Mutation | Versi immutable dan digest yang cocok, terikat pada posting dan requisition, dengan tujuan `recruitment_job_posting`. |
| `CompleteApplicationAssessment` | Mutation | Jika referensi bukti diberikan, versi immutable dan digest yang cocok, terikat pada asesmen dan lamaran, dengan tujuan `recruitment_assessment_evidence`. |
| `IssueOffer` | Mutation | Dokumen dan versi final immutable, digest SHA-256, serta ikatan dengan offer, lamaran, dan versi kompensasi yang disetujui. Tujuannya `recruitment_offer`. |
| `RecordSignedEmploymentContract` | Mutation | Versi kontrak final immutable yang terikat pada kontrak, pekerja, dan badan hukum pemberi kerja. DRM memverifikasi tanda tangan para pihak dan mengembalikan ID dokumen, ID versi, waktu penandatanganan selesai, dan digest SHA-256 mentah (32 byte). |
| `CompleteOnboardingTask` | Mutation | Versi bukti terverifikasi dan immutable yang terikat pada tugas dan rencana onboarding. |
| `ResolveServiceRequest` | Mutation | Keputusan bahwa versi bukti layak dipakai untuk request, pekerja, dan principal terkait. |
| `CalculatePayrollRun` | Mutation | Akses privat kalkulator ke versi artefak aturan, dengan verifikasi SHA-256 atas isi artefak sebelum dipakai. |

## RBAC

Gunakan izin HRM per operasi sebagaimana tercantum pada [personnel](17-operasi-personnel.md), [rekrutmen](20-operasi-recruitment.md), [onboarding](14-operasi-onboarding.md), [HR Service](10-operasi-hrservice.md), dan [payroll](15-operasi-payroll.md). DRM memeriksa hak akses dokumen sendiri berdasarkan identitas tepercaya, owner, scope, dan tautan objek. Identitas dan tautan pada bukti harus berasal dari data DRM.

## Sinkronisasi kontrak

- Sepakati identitas workload, pemetaan principal, serta cara DRM membuktikan versi, digest, owner, tujuan, dan objek yang ditautkan. Untuk tiga operasi rekrutmen pada tabel, HRM mengharapkan owner `human_resource`.
- Tetapkan perilaku untuk versi hilang, bukti lintas objek atau scope, versi yang masih dapat berubah, serta DRM tidak tersedia. Uji penolakan tersebut dan pastikan transaksi HRM tidak mengubah status atau menulis outbox saat verifikasi gagal.
- Untuk pencatatan tanda tangan, Gateway meneruskan ID kontrak dan ID versi. HRM mengambil bukti dari DRM menggunakan principal tepercaya sebelum menyimpan status `signed`.
- Gateway meneruskan referensi versi dan digest sesuai kontrak tiap RPC. Akses atau pengunduhan konten mengikuti aturan DRM. Verifikasi bukti terjadi melalui koneksi privat.

[Boundary DRM dan HRM](../../../docs/03-managements/documents-and-hrm.md)

Slip gaji mandiri membutuhkan penyimpanan PDF immutable dan akses dokumen yang dibatasi ke pekerja pemilik. Trigger dan alurnya dirinci di [handoff payroll](15-operasi-payroll.md#slip-gaji-pdf-mandiri).

## Penjelasan akses PDF slip gaji mandiri

| Tahap | Kebutuhan kontrak |
| --- | --- |
| Penerbitan Finance | Event tepercaya mengikat ID penerbitan ke run dan snapshot yang diverifikasi HRM. Penerbitan dan status transfer pekerja merupakan dua data terpisah. |
| Penyimpanan privat DRM | Terima byte PDF, digest SHA-256, referensi hasil payroll dan penerbitan, serta kunci idempotensi. Kembalikan ID versi immutable yang sama saat permintaan identik diulang. |
| Pembukaan PDF | Verifikasi principal employee, hubungan dengan pekerja pemilik hasil, serta kecocokan versi DRM dengan hasil payroll terbit. Berikan akses hanya ke versi tersebut melalui kontrak privat yang disepakati. |
| Kesiapan portal | Daftar slip hanya menampilkan hasil terbit dengan versi siap akses. Portal membuka PDF yang telah lolos pemeriksaan dan menyediakan cetak melalui penampil PDF. |

## Butuh GraphQL Query

`myPayslip` dan `myPayslips` menyediakan metadata serta ID versi. `myPayslipPdf` menyediakan konten lewat `/api/graphql` dengan respons `no-store`. Portal di `/account` mengambil satu PDF per permintaan dan membuka dokumen pada penampil browser untuk dicetak. Ketiga query membutuhkan sesi employee dengan grant `human_resource.payslips.read_self` berscope SELF yang cocok. ID versi saja tidak memberi akses ke PDF.

## Wiring runtime Finance dan DRM

| Titik pemasangan HRM | Kontrak yang harus dipenuhi |
| --- | --- |
| `WithPayslipPublication` | `VerifyPayrollPublication` memeriksa ID penerbitan, run, jurnal, badan hukum, snapshot input, dan dua digest ke sumber Finance. `StorePayslipPDF` memakai penyimpanan DRM privat dan mengembalikan ID versi immutable setelah berhasil. |
| `WithPayslipVerifier` | Adapter yang sama mengimplementasikan `VerifyPublishedPayslip`, `ListPublishedPayslips`, dan `ReadVerifiedPayslipPDF`. Ketiga metode memakai bukti penerbitan serta pemeriksaan kepemilikan dokumen untuk principal sesi. |
| `HandlePayrollPublication` | Transport privat memberi identitas sumber yang telah diautentikasi, tipe `payroll.published`, versi 1, ID penerbitan, dan ID run. Sinyal sukses dikirim ke transport setelah semua hasil diproses. Kegagalan sementara dijadwalkan ulang. |

Daftar menyaring semua hubungan kerja milik principal, periode opsional, hasil yang sudah terbit, dan versi DRM yang siap diakses sebelum menghitung total dan menerapkan paginasi. Detail dan PDF mencocokkan hasil, pekerja, run, penerbitan, serta versi yang sama. Pembaca PDF menghormati batas ukuran dari HRM dan mengembalikan digest versi bersama byte. Registry dan handler tersedia. Startup standar belum memasang adapter Finance atau DRM produksi.

[Registry HRM](../../../backend/services/human_resource/internal/service/registry.go) · [Verifier metadata dan PDF](../../../backend/services/human_resource/internal/service/payroll/payslip.go) · [Kontrak daftar](../../../backend/services/human_resource/internal/service/payroll/payslip_list.go) · [Kontrak penerbitan](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go) · [Handler inbox](../../../backend/services/human_resource/infrastructure/inbox/payroll_publication.go).

## Verifikasi akses slip

Uji akses sendiri saat transfer masih terjadwal, penolakan untuk hasil pekerja lain dan versi yang salah, penolakan sebelum penerbitan atau saat PDF belum siap, serta retry setelah DRM menyimpan PDF tetapi sebelum HRM menyimpan referensi versi.

Renderer HRM lulus uji PDF dengan DejaVu Sans dan ekstraksi `pdftotext`. Dokumen stabil saat urutan input berubah, menampilkan nomor pekerja, nilai gaji bersih, dan teks beraksen tanpa UUID internal, serta menyembunyikan kontribusi yang hanya untuk pemberi kerja. Penyimpanan dan akses PDF melalui runtime DRM masih memerlukan kontrak privat dan pengujian lintas layanan. [Tes renderer](../../../backend/services/human_resource/infrastructure/integration/document_records/payslip_test.go).

Uji HRM dengan PostgreSQL dan renderer PDF menyimulasikan penyimpanan dokumen berhasil di DRM lalu respons hilang. Referensi versi dan penerbitan di HRM tetap kosong. Pengulangan event untuk run yang sama mengirim PDF identik dengan digest dan kunci idempotensi yang sama. Fixture mengembalikan versi awal dan HRM mencatatnya. Pengulangan setelah pencatatan selesai tidak memanggil penyimpanan lagi.

## Urutan verifikasi saat integrasi tersedia

1. Jalankan penerbitan Finance dan periksa bahwa HRM menerima event dari workload Finance, mencocokkannya dengan run dan jurnal, lalu meminta penyimpanan PDF privat.
2. Catat ID versi yang dikembalikan DRM. Daftar dan detail dengan sesi pekerja harus menunjuk ke hasil payroll serta versi yang sama. Jika verifier mengembalikan ID periode, nilainya harus cocok dengan periode run HRM.
3. Buka dan cetak PDF dari portal memakai sesi pemilik. Ulangi dengan pekerja lain dan pastikan konten ditolak. Ulangi event penerbitan dan pastikan versi dokumen tetap sama.

## Kontrak baca PDF pada HRM

RPC privat `GetMyPayslip` menerima `include_pdf: true` untuk membuka versi PDF yang sudah terverifikasi. Adapter pembaca DRM menerima principal, pekerja, hasil, run, dan ID penerbitan serta versi dokumen yang telah diperiksa. Adapter mengembalikan versi yang sama, isi PDF, dan digest SHA-256 dari versi tersebut. Batas RPC adalah 2 MiB. HRM memeriksa kecocokan versi, digest, dan format dasar, lalu memeriksa otorisasi ulang. Query daftar dan detail metadata tidak membuka konten. Gateway menyajikan konten melalui Query `myPayslipPdf`. Implementasi adapter DRM produksi masih diperlukan.

Tes `TestHRMOwnPayslipPDFOverHTTPS` menerima fixture akun pada `E2E_HRM_PAYSLIP_EMPLOYEE_IDENTIFIER` dan `E2E_HRM_PAYSLIP_EMPLOYEE_PASSWORD`, ID hasil yang berbeda pada `E2E_HRM_OWN_PAYSLIP_RESULT_ID` dan `E2E_HRM_OTHER_PAYSLIP_RESULT_ID`, serta digest versi DRM pada `E2E_HRM_OWN_PAYSLIP_PDF_SHA256`. Jalankan dari modul Gateway dengan `go test -tags=e2e ./internal/e2e -run '^TestHRMOwnPayslipPDFOverHTTPS$' -count=1`. Tes memeriksa isi dan header cache dokumen milik sendiri, serta penolakan terhadap slip pekerja lain.

Klien meminta satu PDF per operasi GraphQL. Gateway membatasi setiap operasi agar satu dokumen atau kolom base64 tidak diminta berulang melalui alias. Portal sudah menggunakan satu permintaan untuk satu slip.

Batas 2 MiB berlaku sejak HRM membuat dokumen. HRM menolak hasil render yang melampaui batas sebelum menghubungi DRM. Renderer, pembaca HRM, dan Gateway memakai konstanta kontrak yang sama.

Regresi dua pemroses serentak memblokir pemanggilan simpan pertama sampai PostgreSQL menunjukkan pemroses kedua sedang menunggu kunci run. Kedua pemroses kemudian menerima ID versi yang sama dan fixture penyimpanan dipanggil sekali. Hasil ini berlaku pada transaksi HRM. Deduplikasi di DRM tetap diperlukan jika respons penyimpanan hilang sebelum commit HRM.

Sesudah verifier atau pembaca PDF selesai, HRM membaca ulang sesi dan tautan principal ke pekerja. Jika tautan pemilik berubah, detail dan isi PDF tidak dikirim. Query daftar juga menolak halaman apabila himpunan pekerja berubah saat pemrosesan. Uji gRPC PostgreSQL terisolasi mencakup pelepasan tautan saat pembaca PDF berjalan.

Gateway menerima hasil `recordSignedEmploymentContract` hanya jika HRM mengembalikan kontrak yang diminta, status `signed`, ID versi yang diminta, UUID dokumen yang valid, waktu tanda tangan yang valid, dan digest 32 byte. Regresi Gateway membuktikan respons tanpa dokumen atau dengan digest salah ditolak. Jalur positif masih memerlukan verifier DRM runtime.

### Pembacaan artefak aturan payroll

Provider DRM untuk `ReadRuleArtifactVersion` perlu mengembalikan byte versi immutable yang diminta melalui identitas workload HRM. Referensi yang tersedia pada HRM adalah UUID versi, bukan UUID dokumen induk. Respons dari DRM harus terikat pada versi yang diminta, termasuk saat dokumen induk memiliki versi baru. HRM memverifikasi digest SHA-256 terhadap byte respons sebelum engine membaca aturan. Versi hilang, akses ditolak, gangguan DRM, atau digest berbeda menghentikan kalkulasi dan mempertahankan run sebagai draf. [Kontrak pembaca HRM](../../../backend/services/human_resource/internal/service/payroll/verified_calculator.go)

Regresi PostgreSQL payroll dengan renderer aktif memakai `HRM_PAYSLIP_TEST_FONT` yang menunjuk ke DejaVu Sans. Suite lulus dengan race detector, termasuk retry publikasi yang mengirim PDF identik dan jalur cetak ulang. Suite renderer terpisah juga lulus dengan `pdftotext`. Penyimpanan, akses dan pemulihan produksi tetap membutuhkan service DRM.
