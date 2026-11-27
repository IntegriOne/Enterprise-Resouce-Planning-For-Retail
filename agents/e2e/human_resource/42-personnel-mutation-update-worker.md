# UpdateWorker melalui gRPC dan PostgreSQL

| Kasus | Hasil |
| --- | --- |
| Koreksi tanggal mulai bekerja pada worker preboarding | Sukses dan tanggal baru tersimpan |
| Koreksi tanggal disertai pengaitan principal IAM | FailedPrecondition |
| Baca worker setelah penolakan | Tanggal lama tetap tersimpan dan principal IAM tetap kosong |

Tes lulus dengan PostgreSQL sementara dan race detector.
