# Working History Frontend #4: Migrasi `/register` ke `createUser` dengan Input `username` dan Redirect Anti-Enumerasi

**Tanggal:** 2026-11-15 09:40 WIB

## Ringkasan Perubahan
- **Klien GraphQL Browser ([`iam-browser-auth.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-browser-auth.ts)):** Menghapus `RegisterCustomerPayload`, `REGISTER_CUSTOMER_MUTATION`, dan `submitRegisterCustomer`; menggantinya dengan `CreateUserPayload` (`user` + `redirectLocation`), `CREATE_USER_MUTATION`, dan `submitCreateUser` yang juga membaca header HTTP `Location`.
- **Halaman `/register` ([`register-experience.tsx`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/app/register/register-experience.tsx)):** Menambahkan field wajib `username` (`id="reg-username"`, validasi client sesuai pola IAM `[a-z0-9][a-z0-9._-]{1,62}[a-z0-9]`), mengganti pengiriman formulir ke `submitCreateUser`, serta mengarahkan ke `/login` melalui `useRouter().push("/login")` ketika menerima header `Location: /login` atau `redirectLocation: "/login"`.
- **Unit Test ([`iam-browser-auth.test.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/lib/graphql/iam-browser-auth.test.ts)):** Menambahkan pengujian kontrak `submitCreateUser` untuk sukses publik (`user: null`), duplikat email (`Location: /login` + `redirectLocation: "/login"`), dan konflik username (`CONFLICT` `"username has been taken"`).

## Verifikasi
- `node --test` (`5/5` PASS), `npx next typegen`, `npm run lint` (`0` errors, `0` warnings), `npx tsc --noEmit`, `npm run build`, serta smoke `/register` HTTP `200` via Caddy HTTPS `PASS`.
