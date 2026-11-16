# Inventaris permission workflow HRM lanjutan

Katalog HRM kini menyebut kode tindakan khusus per workflow untuk workforce planning, personnel, recruitment, onboarding, offboarding, payroll, leave, performance, learning, succession, employee relations, career, policy, HR service, engagement, access request, talent dan promosi. Kode rancangan mencakup `approve`, `submit`, `finalize`, `apply`, `terminate` dan tindakan terkait dengan batas SoD, bukti lintas owner, scope dan transaksi.

Semua kode tambahan masih **rancangan nonaktif**. Runtime tetap memetakan 323 kode yang sudah punya operasi; tidak ada grant atau perubahan IAM. `approve` akses sengaja tidak diberikan kepada HRM karena keputusan grant milik IAM. Dokumentasi ini menjadi daftar kerja untuk command gRPC dan validator berikutnya, bukan klaim workflow sudah berjalan.

Verifikasi: `git diff --check` pada berkas HRM lulus. Tidak ada perubahan Go, skema, migrasi, atau container dalam pekerjaan ini.
