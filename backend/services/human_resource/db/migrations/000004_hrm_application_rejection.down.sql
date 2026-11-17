-- Pembalikan hanya untuk database sekali pakai pada pengujian migrasi.
ALTER TABLE hr_applications DROP CONSTRAINT ck_hr_application_rejected;
