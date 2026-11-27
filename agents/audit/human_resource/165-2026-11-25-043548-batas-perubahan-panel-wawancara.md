# Batas perubahan panel wawancara

Validator sebelumnya memakai batas status yang sama untuk perubahan panelis dan pencatatan feedback. Akibatnya, panelis masih dapat ditambahkan ketika wawancara berstatus `in_progress`, atau dihapus setelah waktu mulai apabila statusnya masih `planned`.

Validator kini memeriksa status dan waktu mulai setelah mengunci wawancara dan lamaran. Perubahan panelis setelah salah satu batas tersebut menghasilkan `FailedPrecondition`. Validator feedback tetap memakai aturan yang sudah ada.

Uji regresi mereproduksi celah status sebelum perbaikan. Tes keamanan dan handler dengan race detector, integrasi gRPC dengan PostgreSQL sementara, dan `go vet ./...` lulus.
