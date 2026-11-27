# Query detail pekerja HRM

Query hrmWorker dan Relay Node memakai adapter HRM menuju GetWorkerByID. Gateway memeriksa sesi employee dan permission read dengan cakupan global atau legal_entity sesuai grant. Scope objek diputuskan HRM. Dependensi module ditambahkan ke Dockerfile, dan koneksi HRM memakai circuit breaker tersendiri. Go generate, tes Gateway dengan race detector, vet, serta kompilasi tanpa go.work lulus. Grant IAM resmi dan HTTPS lintas service belum diuji.
