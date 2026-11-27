# Approval payroll hanya pada periode terbuka

Validator persetujuan run kini memeriksa status periode setelah mengambil row lock. Regresi menunjukkan run yang telah dihitung sebelumnya dapat disetujui walau periode sudah ditutup. Periode tertutup kini menghasilkan FailedPrecondition, dan periode terbuka tetap menerima approval yang sah.
