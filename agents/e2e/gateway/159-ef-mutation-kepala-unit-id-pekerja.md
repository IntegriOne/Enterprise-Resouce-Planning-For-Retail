# Uji HTTPS ID pekerja untuk kepala unit

Pada stack terisolasi, superadmin membuat unit Enterprise Foundation dan pekerja aktif. IAM menautkan akun melalui provisioning. Role dan permission uji diberikan melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Assign memakai Global ID akun IAM | Ditolak dengan `ARGUMENT` |
| Assign memakai Global ID pekerja HRM | Unit menyimpan UUID pekerja |
| Query unit berdasarkan kepala unit memakai Global ID pekerja | Unit yang sama dikembalikan |
| Assign dengan nilai null | Kepala unit dikosongkan |

Go test dengan race detector untuk adapter Enterprise Foundation dan server Gateway lulus.
