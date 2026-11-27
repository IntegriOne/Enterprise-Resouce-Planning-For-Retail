# Pengujian HTTPS guard assignment dan posisi

`TestHRMWorkerAndPositionGuardsOverHTTPS` mengirim permintaan ke Caddy lokal dengan CA pengembangan dan tanpa sesi. Empat belas skenario lulus melalui `/api/graphql`.

| Operasi | Hasil |
| --- | --- |
| `hrmEmploymentContract` | UNAUTHENTICATED |
| `hrmCreateEmploymentContract` | FORBIDDEN karena CSRF absen |
| `hrmRecordSignedEmploymentContract` | FORBIDDEN karena CSRF absen |
| `hrmDeleteEmploymentContract` | FORBIDDEN karena CSRF absen |
| `hrmEmploymentContracts` | UNAUTHENTICATED |
| `hrmWorkerAssignment` | UNAUTHENTICATED |
| `hrmWorkerAssignments` | UNAUTHENTICATED |
| `hrmWorkerAssignments(position_id: ...)` | UNAUTHENTICATED |
| `hrmApproveWorkerAssignment` | FORBIDDEN karena CSRF absen |
| `hrmApplyPrimaryAssignment` | FORBIDDEN karena CSRF absen |
| `hrmDecideWorkerAssignmentTransfer` | FORBIDDEN karena CSRF absen |
| `hrmActivatePosition` | FORBIDDEN karena CSRF absen |
| `hrmUpdatePosition` | FORBIDDEN karena CSRF absen |
| `hrmDeletePosition` | FORBIDDEN karena CSRF absen |

Uji positif dengan akun employee dan grant IAM resmi masih tertunda.

Kedua replika sehat setelah rollout image `sha256:953ec3e2f712a3920ac2648e6c869fb19a2b19f7c08f890133683ef00083b688`. Filter posisi divalidasi melalui schema aktif dan mencapai guard sesi. Image lama tersedia dengan tag `hrm-before-position-filter`.
