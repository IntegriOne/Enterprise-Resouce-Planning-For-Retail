# Uji HTTPS pekerja hiring manager pada detail requisition

Superadmin memberi role dan permission melalui GraphQL IAM. Stack terisolasi membuat draf lalu mengganti hiring manager dengan pekerja aktif dari fixture HRM.

| Kasus | Hasil |
| --- | --- |
| Detail setelah perubahan manager | `hiring_manager_worker_id` berisi GID pekerja pengganti |
| Refetch lewat Relay Node | Referensi pekerja sama |
| Query pekerja memakai referensi tersebut | Nomor pekerja pengganti terbaca |

Integrasi gRPC dengan PostgreSQL juga memverifikasi ID pekerja pada detail. Pengujian Gateway dan Go vet adapter HRM lulus.
