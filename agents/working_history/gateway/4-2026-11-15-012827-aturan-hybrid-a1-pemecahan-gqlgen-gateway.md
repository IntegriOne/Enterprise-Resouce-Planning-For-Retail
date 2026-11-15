# Working History — Penetapan Aturan Modularisasi Codegen GraphQL Gateway (Hybrid A1)

Setelah audit kemampuan native `gqlgen v0.17.97` dan panduan `gqlgen.com`, aturan arsitektur Gateway pada [AGENTS.md](../../../AGENTS.md) diperbarui terlebih dahulu sebelum mengubah konfigurasi generator:

1. Menetapkan pemecahan skema `graphql/**/*.graphql` dengan prefix file unik per management/sub-domain (`graphql/{management}/{management}_{domain}.graphql`) agar tidak bentrok karena pemetaan berbasis `filepath.Base` di `gqlgen`.
2. Mewajibkan `exec.layout: follow-schema` (`graph/generated/{name}.generated.go` + `root_.generated.go`) serta `resolver.layout: follow-schema` sebagai delegator tipis menuju adapter/resolver Zona 1 (`internal/adapter/{management}/` atau `graph/resolvers/{management}/{domain}/`).
3. Menetapkan pola Hybrid A1 untuk `modelgen`: memecah model ter-generate 1:1 per skema (`graph/model/{management}_{domain}_models_gen.go`) dalam satu `package model` via custom plugin di [generate.go](../../../backend/gateway/generate.go) (+ `autobind` opsional) untuk menghapus motherfile tanpa memicu *Go import cycle* lintas domain.
4. Menyelaraskan Zona 1–3 Gateway, CI Gate GraphQL (`go generate ./...`), dan larangan menaruh `.graphql` aktif bagi modul scaffold/*PAUSED*.
