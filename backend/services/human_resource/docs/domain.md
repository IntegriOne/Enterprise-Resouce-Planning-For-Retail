# Domain HRM dan wiring repository

Lapisan domain sekarang menyediakan constructor `New<Entity>(snapshot)` serta `(*Entity).Validate()` bagi **86 entity** seluruh 22 domain bisnis dan inbox/outbox, validator sembilan enum, aritmetika decimal eksak, serta aturan agregat untuk sembilan domain utama. Ini melengkapi repository yang sudah tersedia; bukan klaim seluruh workflow aplikasi HRM, otorisasi atau E2E selesai.

## Entity dan invariant

Validasi memeriksa UUID nonzero, field teks wajib dan panjang Unicode, panjang `char(n)`, timestamp/date nonzero, enum, JSON, evidence bytes, identity, serta precision/scale numeric. Constructor tidak membuat ID, default bisnis, timestamp, session atau grant. Identity baru dapat bernilai nol agar PostgreSQL menghasilkan ID; referensi identity wajib positif.

Seluruh **147 CHECK** DBML dikompilasi menjadi predicate Go statis per domain melalui `tools/domain_validation.py`, dipanggil generator repository resmi. Domain tidak menjalankan SQL atau memuat DBML pada runtime. NULL memakai logika tiga nilai agar kondisi opsional tidak berubah artinya; angka dihitung sebagai rasional eksak, tanggal kalender dibandingkan terpisah dari timestamp. Sintaks CHECK yang belum didukung menghentikan generator, bukan dilewati.

`Decimal.ValidatePrecision(p,s)` menolak overflow atau nilai yang akan dibulatkan diam-diam. Nol tambahan di belakang pecahan diperbolehkan jika nilai tetap eksak. Representasi decimal dibatasi sebelum parsing angka besar. `DecimalSum`, `DecimalNegate` dan `Rat` tidak mengubah nilai asal atau memakai float64. Error domain berupa `ValidationError{Field, Rule}` mendukung `errors.Is(err, ErrInvalidInput)` tanpa menyalin nilai PII.

Validasi domain lebih ketat daripada SQL untuk UUID nol, teks wajib yang kosong, evidence bytes kosong dan timestamp zero. Keberadaan referensi luar tetap harus diperiksa lewat kontrak owner. FK, unique, exclusion dan immutable guard PostgreSQL tetap menjadi perlindungan tambahan; `Validate()` tidak membuktikan keberadaan objek, scope atau izin actor.

## Aturan agregat yang tersedia

| Package di `domain/service/` | Operasi dan batas |
|---|---|
| `personnel` | `ValidateAssignment`: worker/position/employer konsisten, periode assignment berada dalam validitas posisi, employment terminal tidak dipakai ulang. Kapasitas serentak dan master Foundation tetap diperiksa use case. |
| `recruitment` | `ValidateHiring`: korelasi application/candidate/requisition/hired worker dan person party; tidak membuat akun IAM. |
| `payroll` | `ValidateResultTotals`, `ValidateRunResults`: earning/deduction/tax/reimbursement/employer contribution direkonsiliasi secara eksak; referensi run/group/currency dan worker/baris unik. Tidak menetapkan tarif BPJS/pajak, membayar atau memberi approval. |
| `leave` | `RebuildBalance`: replay ledger lengkap menjadi salinan projection; urutan kontigu, sumber unik, opening tunggal, reversal tepat jumlah dan satu kali, adjustment cocok dengan approval independen serta tidak diposting dua kali. Reversal atas reversal ditolak; koreksi lanjutan memakai entry/keputusan yang sah. Saldo negatif ditolak. |
| `talent` | `ValidateFinalization`: cycle/framework/worker evidence sesuai, bobot dimensi 100, anchor level 1–5, skor wajib lengkap dan unik, performance evidence final bila dipakai. `ValidateCalibration` mempertahankan skor asli pada sesi terbuka dan assessment final. |
| `performance` | `ValidateReviewGoals`: seluruh goal dalam cakupan review mempunyai skor worker/cycle yang sama, bobot snapshot cocok dan total 100; review final harus memakai revisi baru. |
| `career` | `ValidateAspiration`: career path/transition/target job konsisten, tanpa perubahan jabatan resmi. |
| `learning` | `ValidateCourseOutcomes`: course/competency mapping unik dan sesuai katalog; tidak mengubah skor kompetensi setelah kursus. |
| `promotion` | `ValidateDecision`: independensi pengusul/worker dan nomor tahap meningkat. `BuildAppliedRecommendation`: approval dan assessment final cocok, worker/target/primary assignment sesuai, assignment lama ditutup H-1, event promotion sudah mempunyai identity, tanggal efektif tidak mendatang dan waktu server konsisten. Hasil berupa salinan recommendation applied. |

Agregat tidak melakukan IO, tidak menyimpan record dan tidak memanggil IAM/Foundation/Finance/DRM. Input record harus berasal dari pembacaan terotorisasi yang konsisten dalam transaksi; waktu bisnis berasal dari server. Memanggil validator tidak memutuskan bahwa actor berhak menyetujui operasi. Mandat panel, tahap approval wajib, eligibility perusahaan dan recheck keamanan tetap milik use case.

Replay cuti membutuhkan **seluruh** ledger dan approval adjustment terkait; caller harus memuat semua halaman di bawah lock entitlement. Input kosong/terpotong bukan bukti seluruh ledger telah dimuat. Projection merupakan hasil replay, bukan saldo yang ditambah ulang ke projection lama. Expiry dihitung sebagai debit pada used snapshot karena skema tidak mempunyai bucket expiry tersendiri. Revisi projection meningkat satu kali tiap rebuild; commit masih harus atomik bersama entry/outbox.

## Wiring nyata

`repository.Registry.Writers()` menghasilkan **86 `record.Writer[Entity,*Entity]`**. Batas generic memastikan validator dan port Insert/Update menerima tipe entity yang persis sama. Constructor menolak dependency nil, termasuk typed nil. Writer memeriksa context, kemudian `Validate()`, baru memanggil repository dengan context transaksi yang sama; error port dikembalikan apa adanya.

```go
store, err := repository.New(db)
if err != nil {
    return err
}
writers, err := store.Writers()
if err != nil {
    return err
}
worker, err := entity.NewWorker(snapshot)
if err != nil {
    return err
}
return store.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
    return writers.Worker.Insert(txCtx, worker)
})
```

Alur aplikasi selanjutnya: otorisasi → baca/lock lewat port repository → periksa aturan agregat → write melalui writer → outbox → commit. `writers.*.Create` adalah alias Insert; `Update` tetap snapshot lengkap. Read/GetAll/FindBy/Count, locking dan delete memakai port repository yang sudah tersedia. Repository mentah tetap operasi persistence murni dan tidak memanggil aturan bisnis; use case wajib memilih jalur validasi ini untuk write.

Writer tidak mengatur transisi antarstatus atau menjalankan semua validator agregat secara otomatis. Constructor record tidak membuat model menjadi immutable di memori. Perubahan hubungan kerja, penerapan promosi dan saldo cuti tetap memerlukan transaksi lengkap pada `services/`; seluruh assignment/event/outbox harus disimpan bersama. Approval saja tidak mengubah jabatan atau grant.

## Verifikasi

- Unit test di `domain/entity/test/` dan masing-masing `domain/service/<domain>/test/`: field/enum/precision/evidence, persamaan payroll, ledger reversal/approval, rubrik/goal, korelasi hubungan kerja, self-approval dan tanggal promosi.
- Mock port membuktikan record invalid tidak mencapai repository, error port tidak berubah, typed nil ditolak dan context dibatalkan diteruskan. Tes race writer menjalankan 16 operasi serentak dengan sinkronisasi deterministik.
- `GOWORK=off go fmt ./...`, `go vet ./...`, `go test -race ./...`, dan `go build ./...` lulus pada modul HRM mandiri.
- PostgreSQL Docker 18.6, dua database sementara `hrm_repository_*_test`: 86 writer dibangun dan menolak nil sebelum persistensi; semua snapshot yang dimuat dari masing-masing 27 tabel fixture lolos `Validate`; update memakai writer dalam transaksi rollback. Matrix repository sebelumnya (86 GetAll/Count, 503 finder dan 81 percobaan write per fixture), serta 70 kasus penolakan SQL tetap lulus. Database dibuang.
- Generator deterministik; assertion bertipe mengikat entity, validator, port dan implementasi repository. Model/schema SQL tidak diubah.

Bukti tersebut belum merupakan positive CRUD semua 86 tabel, pengujian seluruh invariant agregat terhadap concurrency database, atau E2E HRM melalui Gateway. Lapisan `services/`, runtime/config/bootstrap DB, adapter owner, otorisasi, relay inbox/outbox dan API tetap pekerjaan berikutnya.
