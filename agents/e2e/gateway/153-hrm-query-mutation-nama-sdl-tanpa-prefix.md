# Uji HTTPS nama publik SDL sumber daya manusia

Gateway dengan kontrak baru diuji melalui Caddy dan `/api/graphql` pada stack terisolasi. Superadmin menyiapkan permission dan assignment melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| `createWorker`, `activateWorker`, `requestEmployeeAccount`, dan `createEmployeeUser` | Provisioning menghasilkan satu akun tertaut dan receipt tunggal |
| Query pekerja, jabatan, dan unit | Nomor pekerja serta label dari HRM dan Enterprise Foundation terbaca |
| `createRequisition`, `updateRequisition`, dan Query detail | Draf terbentuk, headcount berubah, tanggal target dapat dikosongkan, dan detail cocok |
| `submitRequisition` dan edit setelah pengajuan | Pengajuan berhasil setelah MFA dan edit berikutnya ditolak |
| `deleteRequisition` | Draf dapat dihapus dan requisition yang sudah diajukan dipertahankan |

SDL domain diperiksa untuk memastikan deklarasi tipe dan nama operasi tidak lagi berawalan HRM. Paginasi Relay dan cakupan refetch melalui Node masih dalam penyelarasan.
