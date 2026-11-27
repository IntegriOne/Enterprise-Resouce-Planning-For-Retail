# Wiring pembatalan payroll saat review

Hubungkan `hrmCancelPayrollRun` ke gRPC HRM dengan izin khusus, MFA, dan CSRF. Gateway menolak alasan kosong, meneruskan alasan yang sudah dirapikan, dan menampilkan waktu pembatalan. Tiga kasus pembatalan lulus dalam pengujian integrasi fixture. Race test dan vet Gateway lulus.
