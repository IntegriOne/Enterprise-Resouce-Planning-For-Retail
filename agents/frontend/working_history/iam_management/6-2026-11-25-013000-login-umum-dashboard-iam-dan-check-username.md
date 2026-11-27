# Working History 6 — Login Umum Netral, Dashboard IAM Lengkap, & Check Username

Memperbarui permukaan masuk publik agar sepenuhnya umum serta membangun Dashboard IAM (`/identity/*`) menyeluruh:

- **Login & Registrasi Netral Anti-Kebocoran:** Menghapus rujukan tipe akun (`customer`, `vendor`, `supplier`, `internal`, `mitra`, `portal`, `peran`, `Gateway`) pada `/login`, `/register`, dan `/verify-email`, serta menstandarkan `100dvh`, skeleton sesi, dan feedback `:active`.
- **Dashboard IAM, Sidebar, & Navbar (`/identity/*`):** Membangun guard `<Suspense>` (`identity/layout.tsx`), sidebar collapsible + mobile drawer, sticky navbar (`Ctrl+K` & `Step-Up MFA`), serta 6 halaman penuh (`/identity`, `users`, `roles`, `groups`, `permissions`, `sessions`) bertopografi `Geist` + `Geist Mono` tanpa menyentuh `backend/gateway/`.
- **Check Username GraphQL:** Menambahkan pengecekan username live (`users(query)`) saat selesai mengetik pada Dashboard IAM, error field konflik pada `/login`/`/register`, serta permintaan kontrak `checkUsername` publik ke IAM & Gateway (`agents/notes/`).
- **Verifikasi:** `17/17` unit test, `typegen`, `lint`, `tsc`, `next build`, Docker `frontend-1`/`frontend-2`/`caddy` `Healthy`, dan Chrome Headless E2E HTTPS `PASS`.
