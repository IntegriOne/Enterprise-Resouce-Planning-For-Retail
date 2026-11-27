# ID pengajuan wajib pada createEmployeeUser

Subtes `TestMutationCreateEmployeeUser/ProvisioningRequestRequired` berjalan lewat HTTPS dengan login superadmin development. Pengujian mengirim field wajib yang lain dan sengaja menghilangkan `provisioning_request_id`. Gateway mengembalikan HTTP 422 dengan `GRAPHQL_VALIDATION_FAILED`.

`go test -count=1 -tags=e2e -run '^TestMutationCreateEmployeeUser/ProvisioningRequestRequired$' ./internal/e2e` lulus. Fixture penolakan pada suite Enterprise Foundation juga diselaraskan untuk mengirim bentuk input terbaru. Bukti positif onboarding lengkap tersedia pada [uji aktivasi dan provisioning](135-hrm-mutation-aktivasi-dan-provisioning-akun.md). Suite positif lama yang membuat employee tanpa pengajuan masih membutuhkan fixture baru.

## Regresi positif yang tersimpan di repo

`TestHRMEmployeeProvisioningOverHTTPS` menerima URL stack uji, pengajuan HRM berstatus pending, ID pekerja, dan sesi operator MFA dari runner disposable. Tes membaca detail antrean, menjalankan mutation, memeriksa akun employee PENDING, lalu mengulang mutation dan memeriksa ID tetap sama serta pengajuan berubah menjadi linked. Runner menyiapkan akun superadmin dan grant melalui GraphQL. Sesi uji diteruskan lewat environment proses dan tidak disimpan di repo.

`/tmp/erp-cross-service.UUJx1n/run-durable-provisioning.py` menjalankan tes baru dengan `-tags=e2e` dan keluar dengan status 0. Runner juga memeriksa satu akun dan satu penerimaan di PostgreSQL. Fixture positif suite IAM lama tetap membutuhkan migrasi ke alur ini.

## Menjalankan ulang dari repo

Dari root, jalankan `python3 backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py` dengan stack Compose development hidup. Runner membangun empat binary dari checkout, membuat database serta Redis disposable, dan memakai Caddy loopback pada port 18443. Worker serta referensi EF dibuat lewat GraphQL. Runner memeriksa penolakan aktivasi sebelum MFA dan tanpa kontrak. Database HRM terisolasi menyediakan fixture kontrak bertanda tangan karena verifier DRM belum tersedia. Command `hrmActivateWorker` menghasilkan status aktif dan peristiwa perekrutan. IAM uji memakai manifest resmi untuk izin aktivasi sensitif. Role dan assignment uji tetap melalui GraphQL. Tanggal masuk fixture memakai tanggal UTC sebelumnya agar uji tetap berjalan saat tanggal host WIB sudah berganti. Role dan permission tetap melalui GraphQL. Runner ini telah lulus dan membersihkan container serta database yang dibuatnya.
