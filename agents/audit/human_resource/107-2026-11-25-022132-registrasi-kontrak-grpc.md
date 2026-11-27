# Registrasi kontrak HRM gRPC

Tes membangun registry aplikasi dengan repository offline, menjalankan `RegisterAll` pada server gRPC, dan membandingkan nama layanan, method, serta arah streaming dengan descriptor protobuf terkompilasi. Kontrak saat ini mencakup 22 layanan dan 985 method. `go test -race ./cmd -count=1` dan `go vet ./cmd` lulus. Tes memeriksa registrasi server dan tidak memanggil method bisnis atau IAM.
