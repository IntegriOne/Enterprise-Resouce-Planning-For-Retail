# Application service HRM

Menambahkan 22 package domain di `internal/service`, 84 resource bisnis, DTO create/patch, guard workflow, scoped pagination, transaksi/outbox dan registry repository. Command posting cuti, approval payroll dan finalisasi talent memakai aturan agregat serta menolak dataset parsial. Memperbaiki penyalinan JSON nullable agar SQL NULL tidak berubah menjadi literal JSON null.

Unit/mock representatif setiap domain, rollback dependency/outbox, scope, revoke dan concurrency lulus. `GOWORK=off go fmt ./...`, `go vet ./...`, `go test -race ./...`, `go build ./...` lulus. Benchmark 26 operasi, enam pengulangan, tersimpan di `agents/benchmark/human_resource/`.

Arahan owner untuk perubahan kode langsung diterapkan setelah kerangka awal; service selanjutnya dipelihara lewat patch, tanpa generator service tersimpan. Belum ada adapter keamanan/referensi produksi, kontrak/runtime gRPC atau E2E. Query scoped baru belum diuji terhadap PostgreSQL. Workflow promosi/kalibrasi dan lifecycle lengkap tetap pekerjaan lanjutan.
