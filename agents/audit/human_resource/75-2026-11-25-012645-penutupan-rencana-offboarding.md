# Penutupan rencana offboarding

RPC CompleteOffboardingPlan memerlukan izin step-up dan rencana in_progress yang telah mencapai tanggal efektif. HRM memeriksa bahwa pekerja memiliki hasil pada run payroll yang dirujuk. Verifier otoritatif harus memastikan run tersebut payroll final dan membuktikan clearance Finance serta keputusan IAM untuk pekerja dan rencana yang sama. Semua tugas wajib harus selesai atau di-waive sebelum status dan outbox dicatat dalam transaksi.

Tes unit memeriksa guard tugas wajib. Integrasi gRPC ke PostgreSQL membuktikan jalur produksi menolak penutupan dengan Unavailable saat verifier belum tersedia. Jalur positif yang mencakup clearance nyata masih terbuka. Penutupan tidak menetapkan status pembayaran, pencabutan sesi, atau termination worker.
