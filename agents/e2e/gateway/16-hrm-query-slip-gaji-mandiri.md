# Pengujian integrasi Query slip gaji mandiri

Endpoint GraphQL Gateway diuji dengan sesi fixture dan Payroll gRPC melalui bufconn.

| Kasus | Hasil |
| --- | --- |
| Detail dengan izin self milik sendiri | Referensi versi DRM dan status `scheduled` |
| Daftar dengan filter periode dan izin self milik sendiri | Item pada periode yang diminta dan total dari HRM |
| Izin dengan scope global | `FORBIDDEN` |
| Izin self milik principal lain | `FORBIDDEN` |
| Verifier upstream tidak tersedia | `UNAVAILABLE` |
| Hasil payroll tidak ditemukan | `NOT_FOUND` |

Enam kasus lulus dalam `TestHRMMyPayslipsThroughGateway`. Suite Gateway dengan race detector dan vet juga lulus. Publikasi Finance, penyimpanan PDF, dan unduhan privat DRM belum diuji pada lingkungan ini.
