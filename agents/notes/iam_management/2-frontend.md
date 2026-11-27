# Permintaan Kontrak gRPC IAM: Query Publik `CheckUsernameAvailability`

## Task
Menambahkan RPC ter-rate-limit pada server gRPC `iam` (`backend/services/iam`) untuk mengecek ketersediaan username ketika pengguna selesai mengetik pada formulir pendaftaran publik (`/login` / `/register`) tanpa membocorkan data identitas atau jenis akun (`user_kind`).

## Request
- Service: `backend/services/iam/contract/proto/iam/v1/user.proto` (`UserService`)
- Operasi yang diminta: `rpc CheckUsernameAvailability(CheckUsernameAvailabilityRequest) returns (CheckUsernameAvailabilityResponse)` (mengembalikan `{ available: bool, normalized_username: string, reason: string }` — tanpa mengekspos `user_kind`, `email`, atau `display_name` kepada pemanggil publik anonim).

## Action
1. Normalisasi `username` dan validasi format sesuai `entity.UpdateUsernameValue` (`3-64` karakter huruf kecil/angka/`.`/`_`/`-`).
2. Periksa tabel `user_identities` (`kind = 'username' AND LOWER(value) = ?`) via repository `FindUserIdentityByUsernameKind`.
3. Kembalikan respons konstan `available: true|false` yang aman terhadap *email oracle* dan hanya memeriksa tabel username (jangan menyinggung email atau tipe akun).
4. Terapkan rate limiting ketat di tingkat Gateway/IAM agar endpoint tidak disalahgunakan untuk *username enumeration* massal.

## Context
- UI terautentikasi (`/identity/users` dan `/identity/sessions`) telah memakai query `users(query)` + `identities` yang ada.
- Formulir pendaftaran publik (`/login` tab Daftar Akun & `/register`) belum dapat memanggil `users(query)` karena menolak pemanggil anonim (`UNAUTHENTICATED`), sehingga membutuhkan query publik khusus yang aman dari *enumeration/user_kind leak*.
