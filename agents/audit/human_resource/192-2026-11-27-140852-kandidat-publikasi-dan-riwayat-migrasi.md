# Kandidat publikasi HRM dan riwayat migrasi

Salinan sementara berisi modul HRM lokal dan dependensi Enterprise Foundation serta shared dari `origin/main` `e28491c`. Dengan `GOWORK=off`, unit test modul dan `go vet` lulus. Suite integrasi `go test -race -p 1 -count=1 -tags=integration ./...` juga lulus pada instalasi PostgreSQL baru dengan 69 paket pengujian.

Uji terpisah memasang migrasi resmi `000001` sampai `000007` pada database disposable. Runner kandidat dengan baseline tunggal menolak riwayat tersebut. Ketujuh entri tetap utuh. Database disposable kedua uji sudah dibersihkan.

Kandidat service dapat diuji dengan dependensi main yang tersedia. Pembuktian ini mencakup instalasi baru, sementara lingkungan dengan riwayat lama memerlukan keputusan transisi sesuai kebijakan pra-rilis. Wiring lintas Gateway, IAM, dan frontend masih berada di checkout bersama. [Handoff](../../grpc/hrm/32-handoff-integrasi-main-gateway.md).
