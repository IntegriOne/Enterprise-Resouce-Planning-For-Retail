# Aksi pengajuan dan penghapusan draf requisition

Detail requisition menampilkan aksi Ajukan dan Hapus sesuai izin GLOBAL masing-masing. Kedua tindakan meminta konfirmasi dan menggunakan kontrak GraphQL yang memerlukan sesi MFA. Jika Gateway meminta verifikasi ulang, UI membuka dialog MFA yang sudah digunakan workspace.

Uji browser HTTPS dengan superadmin memverifikasi pembatalan konfirmasi, status setelah pengajuan, dan penghapusan draf lain. GraphQL membuktikan kedua perubahan tersimpan.
