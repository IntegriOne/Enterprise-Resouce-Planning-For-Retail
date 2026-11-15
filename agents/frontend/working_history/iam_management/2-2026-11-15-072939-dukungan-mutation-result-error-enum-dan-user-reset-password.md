# 2 - Dukungan `MutationResult`, `ErrorTypeEnum`, dan `userResetPassword` di Frontend

Tanggal: 2026-11-15 07:29 WIB

## Ringkasan Perubahan
- Memperbarui [`iam-browser-auth.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-browser-auth.ts) agar semua mutasi IAM mengambil `errors { type message }` serta `was_successful`, mengekstrak `AuthRequestError.errorType`, dan merender pesan kesalahan yang mencakup tipe error serta pesan dari server di `formatAuthErrorMessage`.
- Mengganti pemanggilan reset kata sandi menjadi `submitUserResetPassword` (`userResetPassword`), menambahkan `submitChangeUsername` (`changeUsername`), serta menyelaraskan [`login-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/login-experience.tsx), [`verify-email-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/verify-email/verify-email-experience.tsx), dan [`login/README.md`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/login/README.md).
- Memperbarui skenario unit test pada [`iam-browser-auth.test.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-browser-auth.test.ts) untuk kontrak payload `MutationResult`, `userResetPassword`, dan `changeUsername`.

## Verifikasi
- `node --test src/lib/graphql/iam-browser-auth.test.ts` (`4/4` lulus), `npx next typegen`, `npm run lint`, dan `npx tsc --noEmit` keluar dengan kode `0`.
- Image Docker `integrione-retail-frontend:dev` berhasil dibuild & dijalankan (`frontend-1`/`frontend-2` `healthy`); `/login`, `/register`, `/verify-email` merespons `HTTP 200` via HTTPS Caddy.
