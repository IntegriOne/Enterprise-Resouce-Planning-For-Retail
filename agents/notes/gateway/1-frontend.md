# Permintaan Kontrak GraphQL Gateway: Query `checkUsername` Aman Anti-Enumerasi

## Task
Mempublikasikan query GraphQL `checkUsername` pada `/api/graphql` (`backend/gateway/graphql/iam/iam_auth.graphql` / `iam_user.graphql`) setelah RPC `CheckUsernameAvailability` tersedia pada service IAM.

## Request
- Schema: `backend/gateway/graphql/iam/iam_auth.graphql` atau `backend/gateway/graphql/iam/iam_user.graphql`
- Operasi yang diminta:
  ```graphql
  type UsernameAvailabilityPayload {
    available: Boolean!
    normalized_username: String!
    message: String!
  }

  extend type Query {
    checkUsername(username: String!): UsernameAvailabilityPayload!
  }
  ```

## Action
1. Tambahkan resolver `CheckUsername` pada `backend/gateway/internal/adapter/iam/` yang memanggil `UserService.CheckUsernameAvailability`.
2. Terapkan *rate limiting* IP + sliding window yang selaras dengan `LoginRateMax` serta pastikan respons publik **tidak** mengembalikan `user_kind`, `email`, atau `display_name` akun yang ada (untuk mencegah pembocoran 3+ jenis akun IntegriOne pada endpoint publik).

## Context
- Kebutuhan diminta oleh halaman `/login` (tab Daftar Akun) dan `/register` agar dapat memberi umpan balik langsung saat pengguna selesai mengetik (`onBlur`) di `/api/graphql`.
- Sesuai batas ownership Frontend (`frontend/AGENTS.md` §8B & §10D), Frontend tidak memodifikasi `backend/gateway/` secara sepihak dan menunggu kontrak schema Gateway diterbitkan.
