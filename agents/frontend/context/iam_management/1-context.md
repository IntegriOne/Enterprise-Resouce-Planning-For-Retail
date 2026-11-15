# Konteks Aktif Frontend IAM Management (`frontend/`)

## Apa yang sudah dikerjakan

- **Migrasi Pendaftaran `/register` ke `createUser` dengan Input `username` & `poweredByHeader: false`:**
  - Menghapus `RegisterCustomerPayload` dan `submitRegisterCustomer` di [`iam-browser-auth.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-browser-auth.ts), menggantinya dengan `CreateUserPayload` (`user` + `redirectLocation`) serta `submitCreateUser` yang meneruskan header HTTP `Location`.
  - Memperbarui [`register-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/register/register-experience.tsx) dengan input wajib `username` (`id="reg-username"`, validasi `[a-z0-9][a-z0-9._-]{1,62}[a-z0-9]`), pengiriman melalui `submitCreateUser`, dan pengalihan ke `/login` lewat `useRouter().push("/login")` saat server mengembalikan header `Location: /login` atau `redirectLocation: "/login"`.
  - Memastikan `poweredByHeader: false` aktif pada [`next.config.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/next.config.ts) & [`Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/Dockerfile) sehingga header `X-Powered-By` absen di container runtime & Caddy HTTPS.
- **Sinkronisasi Kontrak `MutationResult`, `userResetPassword` & `changeUsername`:**
  - Memelihara kontrak `MutationResult` (`errors { type message }`, `was_successful`), `submitUserResetPassword`, dan `submitChangeUsername`.
- **Verifikasi Build, Docker Compose & Rute Live HTTPS:**
  - `node --test src/lib/graphql/iam-browser-auth.test.ts` (`5/5` PASS), `npx next typegen`, `npm run lint` (`0` errors, `0` warnings), `npx tsc --noEmit`, dan `npm run build` keluar dengan kode `0`.
  - Container `frontend-1` & `frontend-2` berstatus `healthy`; smoke HTTPS mengembalikan `/login` `HTTP 200`, `/register` `HTTP 200` (field `reg-username` aktif, `X-Powered-By` absen), `/verify-email` `HTTP 200`, dan `/identity` (anonim) `HTTP 307`.

## Apa yang belum dikerjakan

- UI lengkap pengaturan akun/username (`changeUsername`) & *pending email management* serta dashboard penuh `/identity` untuk pengelolaan user/role/group mengikuti jadwal sprint IAM UI selanjutnya.
- Pengujian E2E browser Playwright untuk alur form interaktif di peramban belum dibuat.

## Apa yang harus dikerjakan selanjutanya

1. Tambahkan antarmuka pengaturan profil mandiri (`changeUsername`, `updateUserEmail`, `deletePendingUserEmail`, `resendValidationEmailToken`, `uploadAvatarUser`) pada rute akun terautentikasi dan bangun halaman dashboard `/identity`.
2. Implementasikan pengujian E2E browser (`agents/frontend/e2e/iam_management/`) sesuai protokol TestAB saat dashboard `/identity` dan manajemen profil mandiri difinalisasi.
