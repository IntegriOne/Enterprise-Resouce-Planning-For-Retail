# Renderer PDF slip gaji

Adapter HRM merender snapshot payroll ke PDF A4 berfont Unicode dan menyerahkan PDF serta SHA-256 ke port DRM privat. Tes dengan DejaVu Sans dan pdftotext membuktikan dokumen dapat dibaca, label Unicode dan gaji bersih muncul, serta input identik menghasilkan byte identik. Tes unit modul dan go vet lulus. Pemasangan client produksi Finance dan DRM masih terbuka.
