# Audit scope legal entity pada query worker HRM

## Temuan dan perubahan

Policy sebelumnya hanya menerima grant global, sehingga operator HRM dengan grant `legal_entity` tidak dapat membaca worker walaupun boundary dan katalog menyebut scope tersebut. Pembacaan worker kini menerima grant `legal_entity` yang UUID-nya valid. `ReadScope` mengembalikan filter `primary_legal_entity_reference_id` dari snapshot IAM; GetByID menggabungkannya dengan filter `id` dalam satu query, sedangkan GetAll/finder/count menerapkannya sebelum pagination. Tidak ada grant atau perubahan tabel IAM.

Scope parsial dibatasi pada aksi `read`/`list` resource `hr_workers`. Mutasi worker dan resource lain tetap memerlukan grant global serta validator Integrity; dukungan legal entity belum diperluas secara implisit ke kontrak, payroll, atau data sensitif lain. Worker di luar scope diperlakukan sama dengan ID tidak ada pada GetByID.

## Bukti

Tes unit policy mencakup dua legal entity, scope lain, UUID rusak, grant global, serta snapshot nonaktif. Tes service membuktikan GetByID dan list terfilter. Tes gRPC `bufconn` → PostgreSQL sementara `human_resource_scope_test` mencakup GetByID di dalam/luar scope, GetAll limit 1 offset 0/1, finder legal entity lain, serta count; database uji kemudian dihapus. `go fmt`, `go vet`, `go test -race -count=1 ./...`, `go build ./...`, Docker build dan health HRM lulus.

## Batas

Tes policy memakai proyeksi IAM tiruan dan tes SQL memakai policy scope tiruan. Belum ada E2E dengan grant IAM sungguhan melalui Gateway, mTLS/HMAC runtime, atau scope parsial pada domain HRM lain.
