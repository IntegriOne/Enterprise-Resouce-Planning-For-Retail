# Gerbang integrasi HRM

Menjalankan 16 paket integrasi HRM secara terisolasi dengan PostgreSQL Docker `_test`, `-race`, dan `-count=1`; seluruhnya lulus. Dua paket memerlukan opt-in nama database `hrm_disposable_*_test` dan variabel `HRM_TEST_DATABASE_DISPOSABLE=1`, lalu lulus pada pengulangan. Database uji dihapus dan container HRM tetap healthy. Tes ini membuktikan regresi yang tertulis pada paket terkait; IAM/Gateway dan verifier owner produksi belum dibuktikan. Tidak ada kode bisnis atau skema diubah.
