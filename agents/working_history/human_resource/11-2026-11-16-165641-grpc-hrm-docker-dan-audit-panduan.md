# gRPC HRM dan Docker development

Pada 16 November 2026, transport gRPC HRM untuk 22 domain dan 84 resource ditambahkan dengan registrasi eksplisit, DTO typed, pagination, finder, FieldMask, mTLS, proof HMAC, error mapper aman, health dan shutdown. Kontrak JSON memakai byte UTF-8 agar angka besar tidak berubah. Katalog permission HRM disusun sebagai draf review tanpa grant IAM. Adapter policy/integrity tetap menolak operasi bisnis yang belum disetujui atau divalidasi.

`go generate`, vet, race test dan Docker build lulus. Container HRM healthy, baseline `000001` terpasang sekali dan tetap satu migrasi setelah restart. RPC tanpa sertifikat klien atau konteks Gateway ditolak. Audit dan laporan E2E terbatas disimpan terpisah; GraphQL dan CRUD positif belum diklaim.
