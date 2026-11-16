# Repository eksplisit dan cleanup HRM

Atas arahan owner, mengganti embedding 86 repository dengan field DB dan CRUD GORM langsung. Menambahkan nama objek/kolom pada port dan implementasi, misalnya `FindLeavePoliciesByUUID`, `DeleteLeavePoliciesByUUID` dan finder kolom; alias lama menjaga kompatibilitas service. Primary key integer/composite mempertahankan tipe dan predicate lengkap.

Menghapus 25 placeholder, dua helper generik lama dan adapter DomainWriters tanpa caller aktif. Tool model dibatasi agar tidak menimpa port/repository manual atau membuat kembali adapter tersebut. Seluruh penulisan kode memakai patch langsung.

Fmt/vet/test-race/build lulus. PostgreSQL Docker18.6: dua database disposable, 86 port, 503 finder lama+503 bernama objek, scoped list/count dan 162 percobaan write per fixture lulus. Masing-masing fixture mengisi 27 tabel; ini bukan positive CRUD seluruh tabel atau E2E. Tidak mengubah skema, migrasi, service, Gateway atau management lain. Dokumentasi, konteks dan audit diperbarui.
