# Penyelesaian wawancara gRPC

`CompleteInterview` tersedia untuk recruiter yang ditugaskan. HRM memeriksa versi feedback terakhir dari setiap panelis di bawah lock wawancara, lalu menyimpan status completed dan outbox secara atomik. Tes gRPC–PostgreSQL dengan race detector menolak feedback kosong atau sebelum jadwal dan membuktikan satu event pada penyelesaian valid. Go test dan go vet lulus.
