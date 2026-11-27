# Uji webhook outbox dengan PostgreSQL

Uji integrasi paket outbox dengan race detector lulus. Penerima HTTP lokal merespons 503 lalu 204. ID dan isi kedua request sama dan tanda tangan HMAC-SHA256 valid. Status publikasi tersimpan setelah respons berhasil. Hasil ini dicatat di handoff Finance.
