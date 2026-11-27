# Regresi HRM dan katalog runtime

Suite `GOWORK=off go test -p 1 -tags=integration -race -count=1 ./...` lulus pada 67 paket. Database PostgreSQL berakhiran `_test` dibuat khusus untuk pengujian dan dihapus sesudah tes.

Penghitungan langsung terhadap `PermissionCatalog()` menghasilkan 384 kode unik dengan 468 pemetaan operasi. Catatan konteks dan matriks HRM diselaraskan. Pemetaan katalog tidak mengubah role atau grant IAM. Tes ini tidak mencakup verifier Finance/DRM nyata atau sesi IAM melalui Gateway.
