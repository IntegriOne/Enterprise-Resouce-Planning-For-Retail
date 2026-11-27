# Working History #8 — Hapus Menu `/verify-email` & Dukungan Pengalihan Login Belum Terverifikasi

- **Tanggal:** 2026-11-25 01:55 WIB
- **Ruang Lingkup:** Frontend `/login`, `/verify-email`, & catatan kontrak IAM/Gateway.

## Ringkasan Perubahan

1. **Penghapusan Menu `/verify-email`:**
   - Menghapus tautan `Verifikasi Email` dari Navbar ([`login-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login-experience.tsx)), Footbar ([`login-footbar.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login-footbar.tsx)), dan tautan inline pada [`auth-forms.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/auth-forms.tsx).
2. **Pengalihan Login Belum Terverifikasi & Mode Resend di `/verify-email`:**
   - Menambahkan `isUnverifiedLoginResponse` dan `isUnverifiedLoginError` pada [`iam-browser-auth.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-browser-auth.ts) serta `window.location.replace("/verify-email")` pada `handleLoginSubmit`.
   - Memperbarui [`verify-email-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/verify-email/verify-email-experience.tsx) agar akses tanpa `#token=` merender formulir kirim ulang verifikasi (`submitResendValidationEmailToken`) dan tetap memproses `#token=` lewat event `hashchange` maupun muat awal.
   - Mencatat gap backend (`ErrInvalidCredentials` & `normalizeAuthError` di IAM/Gateway) di [`4-frontend.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/notes/iam_management/4-frontend.md) & [`2-frontend.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/notes/gateway/2-frontend.md) tanpa menyentuh `backend/`.
3. **Verifikasi:** `18/18` unit test, `typegen`, `eslint`, `tsc`, `next build` (`12/12`), Docker `Healthy`, dan Chrome Headless E2E PASS.
