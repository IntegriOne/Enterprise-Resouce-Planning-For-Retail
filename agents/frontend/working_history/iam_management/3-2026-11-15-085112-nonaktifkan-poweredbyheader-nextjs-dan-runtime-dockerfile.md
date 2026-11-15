# Penonaktifan `poweredByHeader: false` & Paritas Runtime `Dockerfile` Frontend

- **Waktu:** `2026-11-15 08:51 WIB`
- **Cakupan:** [`frontend/next.config.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/next.config.ts), [`frontend/Dockerfile`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/Dockerfile).
- **Perubahan:** Menambahkan `poweredByHeader: false` pada `next.config.ts` dan menyalin `next.config.ts` ke stage `runtime` pada `Dockerfile` karena `next start` memuat konfigurasi runtime dari `/app/next.config.ts` (sebelumnya tidak tersalin sehingga container masih mengirim `X-Powered-By: Next.js`).
- **Verifikasi:**
  - `npx next typegen && npm run lint && npx tsc --noEmit && npm run build` (`Next.js 16.4.6`) dan `node --test src/lib/graphql/iam-browser-auth.test.ts` (`4/4`) `PASS`.
  - Rebuild `frontend-1` (`16.5s`) dan restart `frontend-1`/`frontend-2` (`healthy`); log mencatat `✓ Running next.config.ts took 11ms`.
  - Probe `curl -skI` ke `:3333/login`, `:3334/login`, dan `https://erp.localhost:8443/{,login,register,identity}` membuktikan header `X-Powered-By` tidak lagi dikirim; E2E HTTPS (`71`/`140`) `PASS`.
