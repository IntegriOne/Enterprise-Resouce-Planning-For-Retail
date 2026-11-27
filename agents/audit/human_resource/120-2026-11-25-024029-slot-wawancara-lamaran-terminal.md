# Slot wawancara setelah lamaran terminal

Tes gRPC dan PostgreSQL sementara mereproduksi draf wawancara dari lamaran yang ditarik masih memblokir kandidat. Kueri sekarang mengabaikan jadwal planned ketika lamaran sudah hired, rejected, atau withdrawn. Jadwal berstatus scheduled tetap menghitung benturan sampai status wawancara berubah. Tes Recruitment dengan race detector, vet modul, dan diff check lulus.
