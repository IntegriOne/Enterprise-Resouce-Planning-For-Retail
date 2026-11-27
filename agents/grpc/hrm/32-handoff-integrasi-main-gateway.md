# Handoff publikasi wiring Gateway HRM

Status terkini dan bukti dari main ada di [status runtime dan coverage](43-status-runtime-dan-coverage.md). Catatan di bawah merekam tahapan kandidat sebelum publikasi.

## Status checkout dan main

Pada commit `4aaa4ae`, `origin/main` belum memuat direktori SDL HRM dan adapter HRM Gateway. Schema GraphQL yang terlacak hanya schema bersama dan IAM. Image development yang sudah diverifikasi dibangun dari worktree lokal yang juga memuat perubahan IAM dan Enterprise Foundation. Bukti E2E pada [matriks Query](31-bukti-https-query-gateway.md) dan [matriks Mutation](30-bukti-https-mutation-gateway.md) berlaku untuk image lokal tersebut.

## File yang perlu direkonsiliasi

| Bagian | Isi yang diperlukan |
| --- | --- |
| Domain HRM Gateway | SDL HRM, resolver per domain, adapter HRM, klien gRPC HRM, dan pengujian HRM |
| Kontrak | Modul Go HRM dan versi proto/generated yang dipakai adapter |
| Gateway bersama | Konfigurasi source connection, bootstrap, wiring resolver, registrasi schema, batas kompleksitas, dan dependency module |
| Kode generated bersama | Root, schema, model, dan file per domain yang dibuat ulang dengan generator dari kombinasi SDL yang akan dipublikasikan |
| Runtime | Koneksi HRM mTLS dan konfigurasi lokal. Manifest IAM uji ada pada image khusus yang dibangun dari konteks di luar checkout. |

File bootstrap dan generated Gateway memuat perubahan gabungan HRM, IAM, dan Foundation. Publikasi perlu memisahkan perubahan milik domain dari wiring bersama, lalu memeriksa hasil build dan E2E dari revisi yang akan dipush. [Status wiring lokal](26-status-wiring-gateway-hrm.md).

## Prasyarat publikasi dari Enterprise Foundation

Pada baseline `origin/main` `4aaa4ae`, direktori Enterprise Foundation hanya berisi DBML. Server HRM lokal mengimpor kontrak `enterprise_foundation/contract/gen/enterprise_foundation/v1` untuk memverifikasi referensi bisnis. Dalam salinan baseline dengan perubahan HRM, `GOWORK=off go test -count=1 -mod=mod ./...` gagal pada paket startup dan adapter Enterprise Foundation karena modul tersebut belum memiliki `go.mod` di baseline.

Setelah menambahkan kandidat `go.mod`, `go.sum`, dan `contract/` milik Enterprise Foundation pada salinan, suite modul HRM dan `go vet` lulus. Kandidat tersebut masih berasal dari worktree Enterprise Foundation. Publikasi server HRM perlu menunggu kontrak modul resmi tersebut masuk `main`. [Uji HTTPS payload Workforce Plan](../../e2e/human_resource/56-workforce-query-mutation-payload-gateway-isolasi.md).

Untuk pemeriksaan referensi HRM, kontrak resmi Enterprise Foundation perlu menyediakan `GetLegalEntityById`, `GetOrganizationUnitById`, `GetLocationById`, dan `GetPartyById`. HRM membutuhkan modul serta kode generated pemilik yang memuat keempat operasi. Untuk E2E yang menulis data, deployment juga membutuhkan server Enterprise Foundation yang melayani kontrak itu, termasuk konfigurasi trust dan HMAC internal yang sesuai. [Klien referensi HRM](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go).

Matriks operasi Gateway sekarang memuat nama payload dan field hasil untuk 47 Mutation. [Kontrak Gateway HRM](26-status-wiring-gateway-hrm.md#penjelasan-mutation).

## Kandidat wiring Gateway lebih dulu

Kandidat Gateway berbasis `4aaa4ae` kini mencakup resolver Relay `node` untuk pekerja HRM, pemuatan modul HRM pada Dockerfile Gateway, dan suite server HRM untuk kontrak payload terbaru. Suite Gateway lulus pada kandidat ini saat memakai kontrak HRM lokal. Percobaan build Gateway dengan kontrak HRM murni dari `origin/main` gagal karena versi itu belum menyediakan `GetMyPayslipResponse`, `WorkerAssignmentTransfer`, `RecordSignedEmploymentContract`, dan `RejectLeaveAdjustment` beserta tipe terkait. Publikasi wiring penuh perlu membawa kontrak HRM yang sesuai.

## Pemeriksaan ulang baseline publikasi

`origin/main` pada `e28491c` sudah memuat modul dan kontrak gRPC Enterprise Foundation yang sebelumnya diperlukan. Proto terlacak mendeklarasikan `GetLegalEntityById`, `GetOrganizationUnitById`, `GetLocationById`, dan `GetPartyById`. Baseline yang sama masih hanya memuat SDL Gateway bersama dan IAM. Bukti HTTPS HRM tetap berasal dari build worktree lokal. Ini menghapus prasyarat publikasi kontrak EF yang dicatat pada baseline lama. Kandidat publikasi berikutnya masih perlu membawa wiring Gateway HRM, perubahan kompatibel IAM dan Foundation, serta hasil generator dari satu schema gabungan.

## Verifikasi kandidat service HRM

Salinan sementara memadukan modul HRM lokal dengan dependensi bersama dan Enterprise Foundation dari `origin/main` `e28491c`. Unit test modul lulus dengan `GOWORK=off`. Gateway, IAM, dan frontend tetap memerlukan verifikasi kandidat publikasi bersama karena mesin uji ini hanya memuat modul HRM serta dependensi Go-nya.

Uji PostgreSQL disposable memasang tujuh migrasi HRM dari `e28491c`, lalu menjalankan runner kandidat dengan baseline tunggal. Runner kandidat mengembalikan `unsupported HRM migration version` dan mempertahankan tujuh entri riwayat. Instalasi baru memakai baseline `000001`. Lingkungan yang masih memakai riwayat lama perlu mengikuti keputusan migrasi pra-rilis di AGENTS.md. Periksa data HRM dan dampak sebelum memilih cara transisi. Uji tidak mengubah database development yang sedang dipakai.

## Verifikasi kandidat Gateway gabungan

Salinan sementara memuat wiring Gateway serta modul HRM, IAM, Enterprise Foundation, dan shared dari checkout lokal di atas baseline `e28491c`. Berkas environment, sertifikat, dan secret tidak disalin. `GOWORK=off go test -count=1 ./...` menjalankan sembilan paket pengujian. Generator resmi `GOWORK=off go generate .` menghasilkan hash yang sama untuk file Go pada direktori generated, model, dan resolver. Bukti ini memeriksa bahwa kode hasil generator sesuai SDL gabungan dan bahwa modul Gateway dapat diuji tanpa workspace lokal atau credential development.

Pada salinan yang sama, `go vet ./...` dan `go build ./cmd/...` Gateway lulus. Regresi dengan race detector lulus untuk tiga adapter management, perkakas GraphQL, dan koneksi sumber. Unit test IAM dan Enterprise Foundation lulus tanpa workspace dengan 24 dan 21 paket yang menjalankan tes. Salinan ini tidak dipublikasikan dan tidak menjalankan ulang HTTPS. Bukti runtime tetap berasal dari stack lokal.

## Uji HTTPS kandidat

Empat binary dari salinan kandidat diuji bersama pada stack disposable melalui Caddy. Provisioning IAM menautkan akun ke pekerja HRM. Request ulang menghasilkan akun yang sama dan refetch Relay mengembalikan pengajuan yang sudah diterapkan. IAM juga memeriksa target grant badan hukum terhadap Enterprise Foundation. [Matriks uji HTTPS](../../e2e/gateway/169-iam-mutation-create-employee-user-kandidat-publikasi.md).

Pada stack kandidat yang sama, seluruh 52 Query dan 55 Mutation lulus pemeriksaan guard HTTPS. Jalur positif provisioning ikut lulus setelah jendela rate limit terlewati. [Uji guard dan provisioning kandidat](../../e2e/gateway/170-hrm-query-mutation-guard-dan-provisioning-kandidat.md).

Ulangi provisioning dengan runner yang sama untuk checkout lokal atau kandidat di direktori terpisah. Opsi `--source-root` membangun keempat binary dari kandidat dengan `GOWORK=off`, sementara fixture dan pengaturan Docker development memakai checkout tempat runner dijalankan.

```sh
python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py --source-root /path/ke/kandidat
```

## Verifikasi snapshot kandidat terbaru

`git fetch origin` mengonfirmasi `origin/main` pada `e28491c`. Direktori GraphQL remote masih memuat schema dasar dan IAM. Empat modul Go lokal lulus tes dan build dengan `GOWORK=off`. Salinan kandidat terpisah hanya memakai sumber backend, schema, dan konfigurasi dari daftar file Git, termasuk file baru yang tidak diabaikan. Berkas environment, sertifikat, dan secret tidak disalin. Generator Gateway menghasilkan hash identik untuk 74 berkas Go generated, model, dan resolver. Gateway pada salinan itu lulus tes dan build tanpa workspace.

Runner HTTPS membangun empat binary dari salinan kandidat. Provisioning serentak, verifikasi tanggal pekerja melalui EF, penolakan cetak ulang tanpa provider, terminasi otomatis, bukti clearance IAM melalui mTLS, penolakan sesi lama, dan penolakan login ulang lulus. Provider Finance dan DRM produksi belum diuji. Kandidat tidak dipublikasikan. Workspace bersama masih memuat perubahan gabungan HRM, IAM, EF, Gateway, dan frontend yang perlu direkonsiliasi sebelum commit/push.

### Dependensi shared yang sudah dipublikasikan

`origin/main` sekarang memuat commit `0054b4a` untuk pembacaan atomik proyeksi user dan security gate Redis. Dua berkas `backend/shared/authsession` diuji pada salinan baseline remote dengan `GOWORK=off`, race detector, dan `go vet` sebelum commit. SHA dikonfirmasi di remote sesudah push. Perubahan IAM dan HRM yang memanggil API baru masih berada di workspace bersama.

Kontrak gRPC terminasi HRM dan IAM beserta binding Go sudah berada di `origin/main` pada `8547df4` dan `133c698`. Keduanya lulus `buf lint`, reproduksi hash generator, dan kompilasi paket pada salinan baseline remote. Kontrak ini memisahkan dependensi protobuf dari publikasi handler dan konfigurasi listener yang masih perlu direkonsiliasi.

Commit `8598e31` memublikasikan handler dan logika IAM untuk terminasi serta clearance. Pemeriksa proyeksi Redis dan otorisasi sertifikat dipisahkan ke berkas milik integrasi ini agar perubahan startup dan proyeksi IAM lain tetap di workspace. Race test serta `go vet` lulus pada salinan baseline sebelum push. Registrasi listener IAM, client verifier HRM, dan pengiriman outbox HRM masih perlu dipublikasikan.

Commit `55062f7` memublikasikan client mTLS IAM untuk verifier HRM, serta dependensi modul dan build Docker yang diperlukan. Perubahan EF di `go.mod` dan Dockerfile IAM tetap ada di workspace. Kandidat terisolasi lulus race test, `go vet`, build command dan build image Docker. Registrasi listener pada daemon IAM tetap perlu dipublikasikan.

Commit `e8e1d8a` telah memasang listener mTLS privat pada daemon IAM. Listener opsional melayani mandat requisition per unit dan mendaftarkan terminasi bila client callback HRM dikonfigurasi. Kandidat hanya membawa diff startup dan konfigurasi yang diperlukan serta fixture tes mandat. Race test, `go vet`, reproduksi binding protobuf dan build command lulus pada baseline terisolasi. SHA remote dikonfirmasi. Sisi listener dan relay HRM serta deployment overlay belum ikut commit ini.

Commit `f9e6e96` memublikasikan client HRM untuk pemeriksaan mandat requisition dan pengiriman event terminasi ke IAM, berikut dependensi Go dan Docker yang dapat dibangun pada baseline remote. Race test, `go vet`, build Go, dan image Docker lulus. Workflow terminasi bisnis, callback, relay, dan overlay deployment masih perlu ikut dipublikasikan.

Commit `3ab8aff` memublikasikan seleksi dan pengiriman outbox khusus terminasi. Constructor lama tetap tersedia untuk jalur event umum, yang melewati terminasi. Kandidat baseline lulus empat mode seleksi PostgreSQL disposable, pemeriksaan digest setelah round trip JSONB, race test, dan `go vet`. Startup HRM dan workflow terminasi masih memerlukan commit lanjutan.

Commit `65f3473` memublikasikan callback verifier terminasi dan listener mTLS opsional HRM untuk IAM. Kandidat lulus race test, `go vet`, dan build command. Percobaan membawa command terminasi bisnis sekaligus menemukan baseline remote belum memiliki field `WorkerAssignment.CancelledAt` yang diperlukan untuk membatalkan penugasan mendatang. Workflow penulisan terminasi, relay daemon, dan overlay deployment masih terbuka. Publikasi berikutnya harus menyertakan model serta migrasi yang kompatibel dengan baseline.

Commit `470490b` memasang relay terminasi pada daemon HRM. Aktivasi memerlukan callback HRM, koneksi IAM, dan pool PostgreSQL minimal dua koneksi. Race test, `go vet`, dan build command lulus pada baseline terisolasi. Pembuat event di workflow terminasi serta konfigurasi deployment belum ikut commit.
