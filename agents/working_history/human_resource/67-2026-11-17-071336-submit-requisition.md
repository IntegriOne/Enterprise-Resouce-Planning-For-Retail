# Submit requisition HRM

`SubmitRequisition` tersedia sebagai command dan RPC privat. Pembuat draf manusia harus menjadi pengusul; status, waktu, actor, dan outbox ditulis atomik setelah validasi ulang referensi. Permission `.submit` sensitif menambah katalog menjadi 329 kode tanpa grant IAM. Tes mock serta gRPC–PostgreSQL `_test`, race/vet/build, dan Docker HRM lulus. Runtime produksi tetap gagal tertutup karena verifier Foundation belum terpasang. Approval/open dan E2E lintas management masih berikutnya.
