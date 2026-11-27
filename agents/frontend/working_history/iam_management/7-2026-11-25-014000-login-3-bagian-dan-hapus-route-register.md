# Working History #7 — Login 3 Bagian & Penghapusan Rute `/register`

- **Tanggal:** 2026-11-25 01:40 WIB
- **Ruang Lingkup:** Frontend `/login` & `/verify-email` (`frontend/src/app/`).

## Ringkasan Perubahan

1. **Penyederhanaan `/login` Menjadi 3 Bagian Saja:**
   - Menghapus `login-editorial-sections.tsx`, registrasi `ScrollTrigger`, serta seluruh CSS Chapters 2–6 pada [`login.module.css`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login.module.css).
   - Menambahkan [`login-footbar.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login-footbar.tsx) sehingga `.enterpriseShell` hanya berisi Navbar (`<header>`), Login Section (`<main>` -> 1 `<section>`), dan Footbar (`<footer>`).
2. **Penghapusan Rute `/register` (Switch ke Register View di `/login`):**
   - Menghapus direktori `frontend/src/app/register/`.
   - Mengubah tombol pendaftaran di Footbar & Navbar agar mengganti `step` menjadi `"register"` langsung di `/login`, serta memastikan [`login-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login-experience.tsx) menangani `/login?mode=register`.
   - Mengubah tautan `Daftar Ulang` di [`verify-email-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/verify-email/verify-email-experience.tsx) dari `/register` ke `/login?mode=register`.
3. **Verifikasi:** `17/17` unit test, `typegen`, `eslint`, `tsc`, `next build` (`12/12`), build Docker `frontend-1`/`frontend-2`/`caddy` `Healthy`, `/register` HTTP `404`, dan Chrome Headless E2E PASS (`0` JS/HTTP errors, desktop `1440x900` fit `900px`, mobile `320x740` zero overflow).
