# Query Personnel dengan scope badan hukum

Tes menggunakan gRPC bufconn, interceptor dan policy HRM produksi, proyeksi sesi sintetis pada Redis TLS disposable, dan PostgreSQL sementara dengan race detector.

| Kasus | Hasil |
| --- | --- |
| Get worker tanpa bukti Gateway | Unauthenticated |
| Get worker dengan bukti Gateway tetapi tanpa hash sesi | Unauthenticated |
| Get worker dengan bukti Gateway dan hash yang tidak dikenali | Unauthenticated |
| Get worker dalam badan hukum pada grant | Satu worker ditemukan |
| Get worker di badan hukum lain | NotFound |
| List limit satu | Satu worker dalam scope dan total satu |
| List offset satu | Halaman kosong dan total tetap satu |
| Finder untuk badan hukum di luar scope | Halaman kosong dan total nol |
| List ketika gerbang keamanan Redis ditutup | Unavailable |
| Get setelah gerbang dibuka kembali dengan epoch dan sesi yang sama | Worker dalam scope ditemukan |
| Get setelah epoch Redis naik sementara proyeksi masih memakai epoch lama | PermissionDenied |
| Get setelah proyeksi keamanan dan izin diselaraskan ke epoch baru | Worker dalam scope ditemukan |
| Get dan list setelah token sesi dicabut dari Redis | Unauthenticated pada kedua RPC |

Bagian mutasi pada suite memakai policy tiruan. Tes query memakai bukti Gateway yang ditandatangani dengan kunci fixture. Jalur query berhasil dengan adapter penyimpanan Redis produksi ketika variabel `HRM_TEST_DISPOSABLE_REDIS_*` diisi. Gateway aktual dan grant yang diprovisikan melalui IAM belum tercakup.

Verifikasi ulang dengan PostgreSQL sementara, Redis TLS sementara, dan race detector lulus. Log tes mengonfirmasi cabang Redis aktif dan kedua RPC get/list ditolak setelah token dicabut. Proyeksi sesi dan grant tetap disiapkan dari fixture tes.

Interceptor produksi dengan bukti Gateway fixture dan Redis TLS juga diuji saat revisi izin berubah. Detail dan daftar pekerja menolak bukti revisi lama dengan `PermissionDenied`. Bukti baru dengan revisi yang cocok kembali mengizinkan detail sesuai scope.

Pada revisi izin baru, grant fixture dipindahkan dari badan hukum A ke B. Bukti lama ditolak. Dengan bukti baru, detail pekerja A memberi `NotFound` dan daftar hanya menghitung serta mengembalikan pekerja B. Tes PostgreSQL dan Redis TLS dengan race detector lulus.

Regresi finder nomor pekerja memakai dua pekerja dari badan hukum berbeda. Nomor di dalam grant mengembalikan satu item dan total satu. Nomor di luar grant mengembalikan halaman kosong dan total nol. Verifikasi dengan PostgreSQL sementara, race detector dan policy HRM produksi lulus. Sesi pada pengujian ini memakai proyeksi fixture.
