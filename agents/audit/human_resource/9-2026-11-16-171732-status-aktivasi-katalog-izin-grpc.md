# Audit aktivasi katalog izin gRPC HRM

Tanggal: 16 November 2026. Acuan: `agents/how_to_code_grpc_managements.md`, katalog HRM, source aplikasi dan hasil build Docker.

## Bukti yang sudah ada

- Kontrak 22 domain/84 resource dan tiga command khusus tersedia; generator proto, vet dan tes race modul HRM lulus.
- `PermissionCatalog()` memetakan 84 resource aplikasi ke 321 kode IAM: `list` berbagi izin `read`; enam resource riwayat/ledger/hasil hanya baca; `post_leave_entry`, `approve`, `finalize` dipetakan eksplisit.
- Peta resource dicocokkan satu per satu dengan 84 `spec.Table` aplikasi. Tes policy membuktikan izin baca tidak memberi create, serta delete/approve tetap butuh MFA meskipun flag sensitivitas snapshot keliru.
- Image HRM dibangun ulang dan container kembali `healthy` setelah wiring katalog awal.

## Belum memenuhi kriteria selesai

- `cmd/main.go` masih memasang `security.NewIntegrity(nil)`. Mutasi bisnis tetap fail-closed; validasi parent, owner, status, dan referensi lintas management belum tersedia untuk seluruh resource.
- Scope `legal_entity`, `organization_unit`, dan `self` belum diimplementasikan; policy service hanya menerima grant global. Belum ada grant IAM resmi atau E2E positif dengan principal berizin.
- Relay outbox, cleanup/retensi, dan delivery lintas service belum dipasang. Listener sehat tidak membuktikan pengiriman event.
- Gateway GraphQL HRM dan kasus E2E lintas Gateway/IAM/DB belum tersedia. Karena itu status ini bukan penyelesaian goal gRPC HRM penuh.
