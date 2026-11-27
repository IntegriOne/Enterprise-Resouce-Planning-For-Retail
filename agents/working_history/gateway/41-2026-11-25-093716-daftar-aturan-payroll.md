# Wiring daftar aturan payroll

Query `hrmPayrollRuleSets` mengambil versi, masa berlaku dan waktu persetujuan dari Payroll gRPC. Gateway memakai izin global baca aturan. Uji fixture membuktikan hasil berpaginasi dan penolakan grant baca grup saja.
