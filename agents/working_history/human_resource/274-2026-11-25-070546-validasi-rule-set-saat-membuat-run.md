# Validasi rule set saat membuat payroll run

Pembuatan run memerlukan rule set yang telah disetujui saat run dibuat. Regresi validator mencakup persetujuan pada waktu pembuatan dan penolakan persetujuan yang baru berlaku sesudahnya. Tes validator dengan race detector, integrasi gRPC PostgreSQL, dan go vet lulus. [Handoff payroll](../../grpc/hrm/15-operasi-payroll.md)
