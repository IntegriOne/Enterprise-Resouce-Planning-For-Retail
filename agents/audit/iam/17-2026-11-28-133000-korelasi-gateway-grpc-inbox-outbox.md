# Audit koneksi IAM — Gateway, gRPC dan korelasi bisnis

**Kesimpulan.** IAM sudah memverifikasi lima jenis scope eksternal di EF dan menyelesaikan provisioning pegawai dengan membaca linkage HR. Identitas target pada respons scope dan pemulihan pengiriman request pegawai masih memerlukan penguatan.

**Lingkup pemeriksaan.** Pembacaan statis checkout bersama pada 28 November 2026 WIB, HEAD `22bc75143003d6d745f386d9808eefb3c5c5c26c`, termasuk source lokal yang belum di-commit. Bukti runtime terdahulu disebut terpisah bila tersedia. Tidak dilakukan transaksi HTTPS, transfer, deployment atau perubahan aplikasi pada audit ini.

## Jalur yang terkonfirmasi dari source

`role` dan `group` memanggil `VerifyScope` saat mengikat role. Legal entity, unit, facility, supplier dan vendor diperiksa di EF; GLOBAL dan SELF ditangani lokal. Outbox keamanan IAM ditulis dan direkonsiliasi ke Redis untuk snapshot/revoke. `Complete` memerlukan verifier worker–account dan diaktifkan oleh `startHRMWorkloads` melalui konfigurasi akses terminasi dan listener mandat. Jalur ini sensitif terhadap konfigurasi, namun sudah ada implementasi. IAM tidak memerlukan inbox data gaji, stok atau invoice. [role/service.go L508](../../../backend/services/iam/services/role/service.go#L508) · [group/service.go L378](../../../backend/services/iam/services/group/service.go#L378) · [outbox/repository.go L162](../../../backend/services/iam/repository/outbox/repository.go#L162) · [redisstore/reconciler.go](../../../backend/services/iam/infrastructure/redisstore/reconciler.go) · [cmd/main.go L103](../../../backend/services/iam/cmd/main.go#L103) · [cmd/requisition_mandate.go L22](../../../backend/services/iam/cmd/requisition_mandate.go#L22) · [provisioning/employee.go L179](../../../backend/services/iam/services/provisioning/employee.go#L179)

**Permukaan kontrak.** Parser gqlparser membaca 25 root Query dan 48 root Mutation dari SDL `backend/gateway/graphql/iam/`. Descriptor dari `buf build` memuat 91 RPC dalam 9 service proto milik IAM. RPC privat/callback ikut terhitung. Angka ini mengukur deklarasi dan tidak dipakai sebagai bukti E2E.

## Kecocokan dengan rancangan bisnis

Procurement §23 membutuhkan aktor, role, permission dan scope, sedangkan PLM §16 memerlukan pemeriksaan akses teknis dan pemisahan reviewer. IAM menyuplai identitas dan grant; fakta pihak/unit tetap diverifikasi di EF. Tidak ada logic.md khusus IAM di folder yang diperiksa. Aturan provisioning dan revoke juga ditelusuri melalui users.md serta implementasi IAM–HRM. [procurement/logic.md](../../logic/procurement/logic.md) · [inventory_warehouse/logic.md](../../logic/inventory_warehouse/logic.md) · [product_life_cycle/logic.md](../../logic/product_life_cycle/logic.md)

## Temuan dan dampak

### IAM-01 — P1 · Keputusan scope belum dicocokkan dengan ID yang diminta

Semua cabang `EFVerifier.VerifyScope` meneruskan expected kind ke `validateDecision`. Fungsi memeriksa eligible dan kind, tetapi tidak membandingkan `target_id` dengan UUID request. EF mengisi field tersebut dari objek yang dinilai. Respons sah untuk objek lain dengan jenis sama dapat disalahartikan sebagai persetujuan target request. Client juga memakai plaintext gRPC dan interceptor penandatangan. [scopeverifier/ef_verifier.go L124](../../../backend/services/iam/infrastructure/scopeverifier/ef_verifier.go#L124) · [v1/common.proto L102](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/common.proto#L102) · [handler/mapper_common.go L253](../../../backend/services/enterprise_foundation/transport/grpc/handler/mapper_common.go#L253) · [cmd/wiring.go](../../../backend/services/iam/cmd/wiring.go)

### IAM-02 — P1 · Jalur penerimaan provisioning untuk pemulihan durable

`Receive` memperoleh aktor dari sesi pegawai aktif, memeriksa katalog dan memasukkan receipt HR. Jalur ini cocok bagi Gateway yang meneruskan permintaan langsung. Worker HR setelah restart tidak membawa sesi browser yang sama. Menjalankan ulang RPC ini dengan identitas manusia sintetis akan mengubah batas kepercayaan. [provisioning/employee.go](../../../backend/services/iam/services/provisioning/employee.go) · [handler/employee_provisioning_handler.go](../../../backend/services/iam/transport/grpc/handler/employee_provisioning_handler.go) · [cmd/requisition_mandate.go](../../../backend/services/iam/cmd/requisition_mandate.go) · [human_resource/access_request.go L100](../../../backend/gateway/internal/adapter/human_resource/access_request.go#L100)

## Tindak lanjut

[Instruksi per temuan, pemilik, pengujian dan kriteria penutupan](../../instructions/iam/1-leader-perbaikan-korelasi-gateway-grpc-instructions.md). [Matriks lintas management dan perhitungan permukaan kontrak](../gateway/2-2026-11-28-133000-korelasi-gateway-grpc-inbox-outbox.md).

**Jejak baseline lokal.** SHA-256 gabungan 16 file backend/logic yang ditautkan pada audit ini adalah `4a9373a0ab76c9d16221d9a80263964900307a4488c8155291be29767cb81065`. Urutkan path relatif unik, lalu untuk tiap path gabungkan UTF-8 path, byte NUL dan digest SHA-256 biner dari isi file. Hash gabungan membantu mencocokkan checkout ketika source lokal belum dipublikasikan.
