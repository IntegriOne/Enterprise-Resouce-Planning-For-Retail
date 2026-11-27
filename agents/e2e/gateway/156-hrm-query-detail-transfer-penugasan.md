# Uji HTTPS detail transfer penugasan

Superadmin membuat grant baca dan assignment melalui GraphQL IAM pada stack terisolasi. Worker diprovisioning melalui IAM–HRM dan unit dibuat melalui EF. Row transfer berasal dari fixture PostgreSQL tes. Uji gRPC PostgreSQL terpisah membaca transfer yang dibuat melalui command submit.

| Kasus | Hasil |
| --- | --- |
| Lookup sebelum pemberian grant | `FORBIDDEN` |
| Lookup dengan UUID dan GID transfer | Hasil identik, status `submitted` |
| Refetch melalui `node` | Tipe `WorkerAssignmentTransfer`, GID dan status cocok |
| Query detail memakai GID pekerja | `INVALID_ARGUMENT` |
| UUID transfer tidak ditemukan | `NOT_FOUND` |
