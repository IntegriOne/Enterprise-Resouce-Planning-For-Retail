# Cegah persetujuan kompensasi tumpang tindih

Validator menolak periode bertumpang tindih dengan kompensasi approved atau active pada pekerja, penempatan dan komponen yang sama. Pemeriksaan berjalan saat worker terkunci. Tes gRPC PostgreSQL membuktikan status proposal tidak berubah setelah penolakan. Race detector dan go vet lulus.
