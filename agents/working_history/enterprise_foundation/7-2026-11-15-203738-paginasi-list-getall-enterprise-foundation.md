# 7. Penyeragaman Paginasi List dan GetAll Enterprise Foundation

**Tanggal:** 15 November 2026 20:37 WIB

## Ringkasan Pekerjaan

Menyeragamkan seluruh operasi daftar (`List*`, `GetAll*`, dan *multi-row lookup*) di `backend/services/enterprise_foundation` memakai `PageRequest{Limit, Offset}` (default `20`, maks `100`) menuju `PageResult[T]{Items, Total, Limit, Offset, HasMore}`:

- **`domain/repository`**: Mengubah kontrak kueri multi-baris menjadi berhalaman, menambah `PartyRoleFilter`/`PartyAddressFilter` beserta `SelectList`, dan mengamankan `NewPageResult`.
- **`repository`**: Menambahkan *short-circuit* pada `pgstore.Paginate` saat `total == 0` atau `offset >= total`, serta menerapkan pengurutan deterministik di seluruh repositori.
- **`services` & Pengujian**: Mengubah seluruh `GetAll*` dan `List*` agar menerima `PageRequest` (pemuatan bertahap untuk frontend), menambah `ListRoles`/`GetAllRoles`/`ListAddressLinks`/`GetAllAddressLinks`, serta memperbarui unit test dan E2E PostgreSQL 18.

## Verifikasi

- `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -p 1 -race -count=1 ./...` pada `enterprise_foundation_test` lulus semua paket (`PASS`).
