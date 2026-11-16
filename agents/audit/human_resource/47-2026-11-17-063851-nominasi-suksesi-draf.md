# Audit nominasi suksesi draf HRM

## Temuan

RPC dan repository `hr_succession_candidates` tersedia tetapi Integrity belum mendaftarkan validator, sehingga mutasi gagal tertutup. Rencana suksesi draf sudah memiliki validator, namun belum dapat mempunyai kandidat lewat jalur HRM.

## Perbaikan

- Validator kandidat mengunci plan, posisi, dan worker. Nominasi hanya untuk rencana belum disetujui dengan kerahasiaan `highly_restricted`, posisi aktif pada tanggal penilaian, worker aktif dari legal entity yang sama, serta pasangan plan–worker unik.
- Readiness harus beralasan dan salah satu nilai terdaftar. Kategori selain `development_needed` memerlukan asesmen talent final milik worker tersebut. Rencana pengembangan opsional harus milik worker dan berstatus aktif/selesai dengan approval. `risk_of_loss` belum dapat diisi lewat CRUD.
- Identitas plan dan worker tidak dapat diganti; plan yang sudah berisi kandidat tidak dapat diubah/dihapus melalui CRUD. Kandidat dalam plan draf dapat dihapus.

## Bukti dan batas

Unit mock menguji nominasi sah, duplikasi, kesiapan tanpa asesmen, asesmen final, rencana pengembangan, plan disetujui, worker tidak aktif, dan cancellation. Tes gRPC `bufconn` → PostgreSQL disposable `_test` menguji create/update/finder/delete, larangan mengubah plan yang dipakai, dan tiga outbox event. `go fmt`, `go vet`, `go build`, seluruh `go test -race`, serta Docker HRM healthy lulus.

Tes memakai policy tiruan serta posisi/worker fixture SQL. Belum ada approval suksesi, mandat panel, E2E IAM/Gateway, atau bukti integritas referensi Foundation untuk posisi. Ini nominasi draf internal, bukan keputusan promosi atau grant IAM. Validator terdaftar 39/78.
