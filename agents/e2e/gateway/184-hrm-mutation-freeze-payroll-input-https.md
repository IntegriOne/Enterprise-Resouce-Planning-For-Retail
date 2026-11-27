# Pembekuan input payroll dan pembuatan draf melalui HTTPS

## Lingkungan

Binary EF, IAM, HRM, dan Gateway dibangun dari checkout kandidat (`--source-root .`) pada stack disposable dengan database PostgreSQL sementara, Redis TLS, Caddy HTTPS pada port `18443`, serta listener mTLS terminasi (`--termination --guards --freeze-payroll`).

## Matriks kasus

| Kasus | Hasil |
| --- | --- |
| Guard seluruh 52 Query dan 62 Mutation tanpa sesi/CSRF (`--guards`) | Lulus (`24 Node`, `24 Relay Connection`) |
| `freezePayrollInput` tanpa MFA (`TC-03`) | Ditolak `AUTHORIZATION` |
| `freezePayrollInput` tanpa `X-CSRF-Token` (`TC-03`) | Ditolak `FORBIDDEN` |
| `freezePayrollInput` pada periode tertutup (`TC-02`) | Ditolak `CONFLICT` |
| `freezePayrollInput` pada periode terbuka dengan profil layak (`TC-01`, `TC-05`) | Berhasil, mengembalikan metadata tanpa token rekening atau isi payroll |
| Pengulangan `freezePayrollInput` dengan `idempotency_key` sama (`TC-04`) | Mengembalikan `input_snapshot_reference_id` dan `input_digest_sha256` identik |
| `createPayrollRun` dengan digest berbeda dari snapshot HRM (`TC-02`) | Ditolak `CONFLICT` |
| `createPayrollRun` memakai ID dan digest dari `freezePayrollInput` (`TC-01`) | Berhasil membentuk draf berstatus `draft` |
| Regresi provisioning dan terminasi employee (`--termination`) | Lulus hingga pencabutan sesi dan verifikasi mTLS |

## Perintah

- `python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py --source-root . --termination --guards --freeze-payroll`
