# Pencatatan respons penawaran

Tes memakai gRPC privat, PostgreSQL sementara, dan verifier kandidat tiruan.

| Kasus | Hasil |
| --- | --- |
| Kandidat pada bukti belum terverifikasi | FailedPrecondition |
| Respons sebelum penawaran diterbitkan, bertanggal mendatang, atau tepat pada tenggat | FailedPrecondition |
| Mencoba decline dengan bukti penerimaan | FailedPrecondition |
| Bukti penerimaan sesuai | Status accepted |
| Penerimaan diulang | FailedPrecondition |
| Bukti penolakan sesuai | Status declined |
