# Audit persetujuan penempatan pekerja — 17 November 2026

## Cakupan dan hasil

`ApproveWorkerAssignment` tersedia sebagai command service dan RPC gRPC privat. Permission baru `human_resource.worker_assignments.approve` memerlukan pemeriksaan sensitif pada policy produksi. Pengaju dan penyetuju harus principal manusia berbeda; waktu serta snapshot status ditetapkan server. Validator menolak perubahan field draf lain, worker/posisi yang belum aktif, legal entity/periode tidak cocok, total FTE pekerja di atas 100%, atau kapasitas posisi di atas `headcount_authorized × 100%`. Baris worker dan posisi dikunci pada transaksi yang sama dengan update assignment dan outbox. Proposal kompensasi membaca assignment approved tanpa mengunci baris assignment sehingga urutan lock tidak berlawanan dengan approval.

## Bukti verifikasi

- `go generate ./contract` dan `go fmt ./...`: lulus.
- `go vet ./...` dan `go test -race ./...`: lulus.
- Tes gRPC `bufconn` → PostgreSQL Docker pada database `_test` sekali pakai: lulus, termasuk self approval, posisi/worker belum aktif, retry, dan dua approval paralel pada posisi kapasitas 100% dengan masing-masing 60% (satu sukses, satu ditolak).
- `docker compose build human-resource` dan `docker compose up -d --wait human-resource`: lulus, container healthy.

## Batas

Tes integrasi memakai policy tiruan; grant IAM dan E2E Gateway belum diuji. Aktivasi posisi lewat API produksi, penerapan penempatan primer ke worker, penutupan periode, dan verifikasi manager pada posisi pelaporan masih pekerjaan berikutnya. Kumpulan assignment dibatasi 10.000 baris saat pemeriksaan; jika terlampaui, approval gagal tertutup.
