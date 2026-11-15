# 2 — E2E HTTPS 22 Query dan 45 Mutation IAM Gateway (1:1 per Operasi)

Menambahkan suite pengujian E2E HTTPS komprehensif pada satu berkas tunggal [iam_operations_e2e_test.go](../../../backend/gateway/internal/e2e/iam_operations_e2e_test.go) (`//go:build e2e`) dengan pemetaan tepat satu fungsi per operasi GraphQL IAM (**22 `TestQuery*`** dan **45 `TestMutation*`**):

- Setiap fungsi menguji multi-skenario terisolasi (`t.Run`), mencetak request/variables/HTTP/JSON result penuh melalui `doGQLOK`/`doGQLErrorLabeled`, memilih seluruh field objek skema `iam.graphql` (`gql*Fields`), serta memvalidasi tiap field/nullability via `assert*Object` dan `assert*Payload`.
- Seluruh fixture pengguna/peran/grup dibersihkan otomatis lewat `t.Cleanup`, menjaga PostgreSQL tetap bersih (`users=1`, `roles=4`, `groups=0`) tanpa kebocoran secret `.env`.
- Verifikasi `gofmt`, `go vet -tags=e2e`, dan eksekusi live HTTPS (`go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/...`) lulus **68/68** (`52.932s`, 0 skip/fail), tersimpan pada [1-alloperations.md](../../e2e/testing/iam/1-alloperations.md) (`13174` baris) dan dirangkum di [2-iam-22-query-dan-45-mutation-operations-e2e.md](../../e2e/gateway/2-iam-22-query-dan-45-mutation-operations-e2e.md).
