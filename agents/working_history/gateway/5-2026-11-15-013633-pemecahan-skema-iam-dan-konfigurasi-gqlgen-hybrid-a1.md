# Pemecahan Skema GraphQL IAM dan Implementasi Config/Splitter gqlgen Hybrid A1

- Memecah [iam.graphql] menjadi [iam_auth.graphql](../../../backend/gateway/graphql/iam/iam_auth.graphql), [iam_user.graphql](../../../backend/gateway/graphql/iam/iam_user.graphql), dan [iam_rbac.graphql](../../../backend/gateway/graphql/iam/iam_rbac.graphql) serta mengonfirmasi validitas `26` query root dan `46` mutation root lewat `gqlparser`.
- Memperbarui [gqlgen.yml](../../../backend/gateway/gqlgen.yml) (`schema: graphql/**/*.graphql`, `exec.layout: follow-schema`, `fast_validation: true`, `use_buffer_pooling: true`).
- Membuat plugin AST splitter [split_modelgen.go](../../../backend/gateway/internal/gqltools/split_modelgen.go), CLI [cmd/gqlgen/main.go](../../../backend/gateway/cmd/gqlgen/main.go), dan memperbarui [generate.go](../../../backend/gateway/generate.go) termasuk pemetaan otomatis `_id -> RawID` pada tipe Relay.
- Memverifikasi lewat [split_modelgen_test.go](../../../backend/gateway/internal/gqltools/test/split_modelgen_test.go), `go vet ./...`, `go build ./cmd/...`, dan unit test Gateway tanpa merusak kompilasi aktif sebelum regenerasi penuh.
