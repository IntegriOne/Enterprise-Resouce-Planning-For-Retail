# Inventaris kontrak dan registrasi gRPC HRM

Pencocokan langsung 22 berkas proto domain dengan deklarasi method pada handler menemukan 985 RPC dengan method eksplisit. Pemeriksaan langsung mencakup proto Leave karena indeks graf menandai satu baris sebagai parse parsial. Proto Common berisi tipe bersama dan tidak memiliki RPC.

`TestRegisterAllMatchesHRMProtoServices` lulus dengan race detector. Tes membandingkan descriptor protobuf Go dengan registry gRPC, termasuk nama layanan, nama method, dan flag streaming. `TestBuildMutationValidatorsCoversAllCatalogMutatableResources` juga lulus dan membuktikan pemasangan validator untuk sumber daya mutasi yang tercakup katalog.

Cakupan ini membuktikan keberadaan handler, registrasi transport, dan pemasangan validator. Operasi dapat tetap dibatasi oleh lifecycle dan dependency pemilik. Penyelesaian bisnis memerlukan verifikasi jalur sukses aktual, khususnya workflow kontrak bertanda tangan DRM dan settlement Finance. Integrasi Gateway dan provisioning IAM juga masih terbuka.
