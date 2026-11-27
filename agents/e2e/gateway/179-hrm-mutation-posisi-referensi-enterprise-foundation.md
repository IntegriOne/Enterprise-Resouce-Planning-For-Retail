# Uji HTTPS posisi dengan referensi Enterprise Foundation

Pengujian memakai build Gateway, HRM, IAM dan Enterprise Foundation pada stack disposable. Superadmin memasang dan mencabut role serta izin HRM melalui GraphQL IAM. Dua badan hukum, satu unit, dan dua lokasi dibuat melalui GraphQL Enterprise Foundation. Job family dan jabatan dibuat melalui GraphQL HRM.

| Kasus | Hasil |
| --- | --- |
| Posisi memakai unit organisasi milik badan hukum lain | `CONFLICT` |
| Posisi memakai lokasi kerja milik badan hukum lain | `CONFLICT` |
| Jabatan, unit dan lokasi valid pada badan hukum yang sama | Posisi tersimpan sebagai `planned`, dengan kode, judul dan referensi lokasi yang benar |

Kedua penolakan memakai kode posisi yang sama dengan percobaan berhasil untuk memastikan percobaan yang ditolak tidak mengonsumsi kode. Suite adapter Enterprise Foundation dan policy HRM juga lulus dengan race detector.

Regresi Gateway dengan respons gRPC fixture memastikan `createPosition` mengembalikan `UNAVAILABLE` saat lokasi pada respons berbeda atau posisi atasan yang diminta hilang. Respons dengan kedua referensi yang cocok tetap berhasil. Kedua kasus inkonsisten sebelumnya dilaporkan berhasil.

Regresi `updatePosition` membuktikan bahwa respons lokasi yang cocok diterima, sedangkan lokasi yang berbeda atau masih terisi setelah permintaan pengosongan ditolak dengan `UNAVAILABLE`. Patch yang tidak mengirim lokasi tetap dapat memperbarui field lain. Pengujian ini memakai fixture gRPC.

Uji HTTPS tambahan memakai stack disposable dengan Enterprise Foundation asli. Patch lintas badan hukum ditolak. Patch judul yang tidak memuat lokasi mempertahankan lokasi lama, `null` mengosongkannya, dan GID lokasi yang sah menetapkannya kembali. Role dan permission superadmin dipasang dan dibersihkan lewat GraphQL IAM.
