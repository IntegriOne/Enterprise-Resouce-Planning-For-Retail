# Pengujian Query metadata kontrak pekerja

| Kasus fixture | Hasil |
| --- | --- |
| Grant baca kontrak global, MFA, dan Relay ID | Metadata kontrak pekerja dikembalikan |
| RPC mengembalikan pekerja berbeda | UNAVAILABLE |
| Sesi tanpa MFA | FORBIDDEN |
| Grant baca profil pekerja dipakai untuk kontrak | FORBIDDEN |
| Detail dengan ID kontrak yang cocok | Metadata kontrak dikembalikan |
| Detail kontrak tidak ditemukan | NOT_FOUND |
| RPC detail mengembalikan ID yang salah | UNAVAILABLE |

Verifikasi menggunakan sesi employee dan fixture gRPC.

## HTTPS lokal

Kedua replika sehat setelah rollout image `sha256:07ff732cc4db8f9f182fec45ec95caa786ca9d8f7ac7985e0c8187803a9b914e`. Query `hrmEmploymentContracts` dikenali di Caddy dan ditolak dengan `UNAUTHENTICATED` saat sesi absen. Suite guard pekerja dan posisi meluluskan sepuluh skenario termasuk Query ini. Uji positif masih memerlukan fixture IAM resmi.

Kedua replika kemudian diperbarui ke image `sha256:1361577971b2011931529149ed6a065f1587309cfeb64325f9a06d07484c52de` yang menambahkan Query detail. Keduanya sehat. Sebelas skenario guard HTTPS lulus, termasuk `hrmEmploymentContract` dan `hrmEmploymentContracts` yang menolak akses anonim. Uji positif grant badan hukum dilewati karena fixture belum tersedia.
