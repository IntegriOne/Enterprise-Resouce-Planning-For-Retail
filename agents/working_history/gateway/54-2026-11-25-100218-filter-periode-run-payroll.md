# Filter periode pada daftar run payroll

Tambahkan filter periode opsional pada `hrmPayrollRuns`. Gateway mendelegasikan filter dan paginasi ke gRPC, memeriksa kecocokan periode, serta menolak penggunaan dua filter sekaligus. Tes mencakup periode valid, respons tidak cocok, dan filter ganda.
