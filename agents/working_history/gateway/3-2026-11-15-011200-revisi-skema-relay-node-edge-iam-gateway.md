# Working History — Revisi Kontrak GraphQL IAM & Gateway Core (Relay Node/Edges Opsi A)

Pada sesi `2026-11-15`, fokus perubahan dibatasi pada kontrak skema GraphQL [schema.graphql](../../../backend/gateway/graphql/schema.graphql) dan [iam.graphql](../../../backend/gateway/graphql/iam.graphql) untuk review sebelum sinkronisasi Go/`gqlgen`/E2E:

1. Menjaga [schema.graphql](../../../backend/gateway/graphql/schema.graphql) tetap minimal (*non-motherfile*) berisi `Time`, `ScopeType`, `Scope`, Relay `PageInfo`, `interface Node`, `Query.node(id: ID!): Node`, `Query.health`, dan placeholder `Mutation`.
2. Menyederhanakan keempat `Create*UserInput` menjadi `{ full_name!, email!, username!, password! }` (`PENDING` hingga `verifyEmail`), memisahkan `uploadAvatarUser(input: UserAvatarUploadInput!): UserAvatarPayload!`, menyelaraskan `VerifyEmailStatus`/`VerifyEmailInput`/`VerifyEmailPayload`, dan menghapus `addUserIdentity`/`AddUserIdentityInput` (email otomatis primer).
3. Menerapkan pola Relay penuh (Opsi A) pada [iam.graphql](../../../backend/gateway/graphql/iam.graphql): 6 tipe `Node` (`id` GID + `_id`), 10 pasang `*Edge`/`*Connection` (`edges`, `nodes`, `pageInfo`, `total_count`), relasi grafik terhubung pada `User`/`Role`/`Group`/`UserIdentity`/`Session`, serta pemetaan 4 mutasi relasi ke `*Edge!`.
4. Validasi parse `gqlparser` (`SCHEMA AUDIT PASS`: 24 query, 45 mutasi IAM, + 8 operasi setara frontend; `frontend` `verifyCustomerEmail` perlu sinkronisasi pada tahap implementasi).
