# Wiring pembuatan draf payroll

Hubungkan `hrmCreatePayrollRun` ke gRPC HRM dengan grant create, MFA dan CSRF. Input memuat referensi snapshot dan digest heksadesimal serta kunci idempotensi. HRM menentukan status dan keputusan workflow. Tes integrasi memverifikasi kontrak draf, duplikasi dan digest invalid. Race test dan vet Gateway lulus.
