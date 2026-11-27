# Uji awal Query pekerja dengan superadmin melalui HTTPS

Caddy, kedua replika Gateway, IAM, dan HRM development aktif. Permintaan HTTPS memakai CA lokal, cookie sesi, Origin, dan token CSRF untuk mutasi sesi. Kredensial dibaca dari konfigurasi lokal.

| Langkah | Hasil aktual |
| --- | --- |
| `browserSession` lalu `userLogin` sebagai superadmin | `AUTHENTICATED` |
| `hrmWorkers { total_count }` sebelum assignment role HRM | `FORBIDDEN` |
| `userLogout` | Berhasil |

SDL IAM menyediakan `assignRolePermission` dan `assignUserRole`. Pendaftaran kode permission yang belum ada dilakukan melalui manifest yang dimuat saat bootstrap IAM. Uji positif masih memerlukan pemuatan permission HRM, assignment role uji, dan login ulang. [Prosedur superadmin](../../../AGENTS.md#uji-fungsional-https-dengan-superadmin).

## Uji setelah pemberian role

Permission `human_resource.workers.read` dimuat melalui bootstrap dengan manifest pada konteks build image IAM uji lokal. Manifest uji berada di luar checkout. IAM kembali sehat dan Query `permissions` menemukan kode tersebut.

| Langkah | Hasil aktual |
| --- | --- |
| `createRole`, `assignRolePermission`, `assignUserRole` dengan scope global lalu login ulang | Berhasil |
| `hrmWorkers` saat koneksi HRM masih dimatikan di Gateway | `UNAVAILABLE` |
| Aktifkan koneksi di kedua replika dengan `HRM_GATEWAY_ENABLED=true docker compose up -d --no-deps --no-build --wait gateway-1 gateway-2` | Kedua replika sehat |
| `hrmWorkers` dengan sesi superadmin dan grant uji | Berhasil, `total_count: 0` pada database HRM yang kosong. Log HRM mengonfirmasi `PersonnelService/GetAllWorkers` berstatus `OK`. |
| `revokeUserRole`, `deleteRole`, `userLogout` | Berhasil |

Image IAM uji masih berjalan di development. Image sebelum pengujian tersimpan sebagai `integrione-retail-iam:before-hrm-test`. Permission katalog telah tercatat dan role uji sudah dihapus. Uji ini membuktikan jalur HTTPS berakun sampai gRPC HRM dan PostgreSQL untuk daftar kosong. Jalur filter dan penolakan lintas badan hukum memerlukan data pekerja dan grant terbatas.
