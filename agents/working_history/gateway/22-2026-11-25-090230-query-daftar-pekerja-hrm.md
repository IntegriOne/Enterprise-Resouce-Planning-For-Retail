# Query daftar pekerja HRM

Query hrmWorkers meneruskan limit dan offset ke GetAllWorkers dan memakai total hasil setelah scope dari HRM. Gateway memakai aturan izin yang sama dengan query detail serta membatasi kompleksitas berdasarkan ukuran halaman. Tes adapter membuktikan request tanpa grant dan pagination invalid berhenti sebelum RPC, serta grant legal_entity dapat meneruskan daftar. HTTPS lintas service masih terbuka.
