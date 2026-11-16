# Benchmark service HRM — 16 November 2026

Target: `backend/services/human_resource/internal/service/<domain>/test/service_bench_test.go` pada 22 domain; command kompleks pada `payroll/test/approval_bench_test.go`. Host Linux amd64, AMD Ryzen 9 9900X 12-Core Processor, Go 1.27.4, GOMAXPROCS 24. Benchmark dijalankan serial antar package (`-p 1`), enam pengulangan, 100 ms per sampel, tanpa race detector.

Perintah aktual dari modul HRM:

```bash
GOWORK=off go test -p 1 -run '^$' -bench '^BenchmarkService' -benchmem -benchtime=100ms -count=6 ./internal/service/... > /tmp/hrm-service-bench-final.txt
/home/muriarfad/go/bin/benchstat /tmp/hrm-service-bench-final.txt
```

Dataset mock: list satu record representatif per domain, page pertama; update satu record dengan validasi, policy dan outbox, mempertahankan event terakhir agar state tetap terbatas. Approval: satu run calculated, satu result dan empat component line; reset status/event dikeluarkan dari timer tiap iterasi. Policy mock memberikan scope penuh secara eksplisit; tidak ada DB, Redis atau RPC. Angka berikut adalah median waktu benchstat enam sampel; alokasi mencakup penyalinan snapshot pada mock transaksi. [Output mentah seluruh sampel beserta N](1-service-domain-pagination-dan-payroll-approval-raw.txt) disimpan untuk reproduksi dan review.

| Domain — operasi | Median µs/op | Alokasi KiB/op | allocs/op |
|---|---:|---:|---:|
| accessrequest — list | 5,646 | 3,779 | 37 |
| analytics — list | 4,603 | 2,687 | 33 |
| benefits — list | 5,511 | 3,127 | 33 |
| career — list | 4,862 | 2,896 | 33 |
| compensation — list | 4,930 | 3,128 | 29 |
| competency — list | 4,693 | 2,844 | 33 |
| employeerelations — list | 5,413 | 3,462 | 33 |
| engagement — list | 5,007 | 3,268 | 35 |
| hrservice — list | 5,332 | 3,650 | 37 |
| learning — list | 4,824 | 2,812 | 29 |
| leave — list | 5,184 | 3,063 | 29 |
| offboarding — list | 6,059 | 4,218 | 35 |
| onboarding — list | 4,577 | 2,738 | 37 |
| payroll — list | 5,519 | 3,760 | 35 |
| payroll — update | 21,67 | 15,40 | 156 |
| payroll — approve | 98,42 | 72,64 | 1.073 |
| performance — list | 5,477 | 3,214 | 35 |
| personnel — list | 5,991 | 3,969 | 41 |
| personnel — update | 22,55 | 15,93 | 180 |
| policy — list | 4,434 | 2,561 | 33 |
| promotion — list | 6,968 | 5,117 | 45 |
| recruitment — list | 7,651 | 4,912 | 45 |
| succession — list | 4,353 | 2,624 | 31 |
| talent — list | 5,226 | 3,332 | 37 |
| workforceplanning — list | 3,564 | 1,838 | 29 |
| workforceplanning — update | 16,92 | 11,37 | 127 |

Enam sampel mentah command approval (N aktual, tanpa pembulatan median):

| Sampel | N | ns/op | B/op | allocs/op |
|---|---:|---:|---:|---:|
| 1 | 1084 | 100541 | 74429 | 1073 |
| 2 | 1161 | 100392 | 74381 | 1073 |
| 3 | 1280 | 97731 | 74380 | 1073 |
| 4 | 1264 | 99110 | 74410 | 1073 |
| 5 | 1224 | 97333 | 74383 | 1072 |
| 6 | 1230 | 96315 | 74372 | 1072 |

Approval ±2% pada enam sampel; list sekitar ±1–6%, update ±1–5%. Dataset lebih besar belum diukur. Beban approval mencakup validasi decimal eksak, pembacaan lengkap child, penggandaan model, serialisasi payload dan snapshot mock. Ini kandidat sumber alokasi berdasarkan kode, **belum hasil profiling**; tidak ada klaim optimasi atau perbandingan terhadap baseline sebelumnya. Ukuran snapshot mock ikut memperbesar hasil, sehingga angka ini tidak boleh dipakai sebagai kapasitas DB/Redis produksi, SLA atau estimasi throughput jaringan. Uji skala produksi dan profiling menunggu adapter serta runtime nyata.
