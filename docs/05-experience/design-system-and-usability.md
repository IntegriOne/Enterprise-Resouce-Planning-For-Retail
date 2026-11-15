# Design System, Style Guide dan Usability QA

> **UX/UI design brief · belum berupa file Figma atau prototipe visual jadi.** Keputusan fase mengikuti [produk](../01-product/roadmap.md) dan kebijakan akses [IAM](../02-architecture/security-and-sessions.md).

## Design tokens (per brand, belum final)
| Token | Status |
|---|---|
| `brand.primary/secondary` | Diisi owner merchant [HEX] |
| `surface/text/muted/error/success` | Semantik; hasil contrast perlu diuji |
| `font.ui` | Diisi font berlisensi |
| `spacing.base` | 8px sebagai hipotesis |
| `radius.base` | Ditentukan setelah prototype |
| `motion.reduced` | Wajib hormati preferensi `prefers-reduced-motion` (CSS media query + `gsap.matchMedia()`) |

## Standar Animasi & Motion (`gsap` + `@gsap/react`)
- Stack animasi frontend resmi memakai **`gsap` (`^3.15.1`)** dan **`@gsap/react` (`^2.1.2`)** (`useGSAP`), hanya pada Client Component (`"use client"`) dengan cleanup kontekstual dan `gsap.matchMedia()` agar pengguna `prefers-reduced-motion: reduce` melihat state akhir yang stabil tanpa animasi mengganggu.
- **Dashboard internal & portal mitra (`/identity`, `/profile`, ERP backoffice, `/supplier`, `/vendor`):** motion harus **subtil, cepat (`~150–250ms`), dan fungsional**; utamakan properti GPU (`transform`, `opacity`), tanpa *scroll-jacking*, *intro delay*, atau animasi pemicu layout thrashing di tabel/grid operasional.
- **Web Commerce storefront & `/login`:** boleh memakai koreografi *editorial* lebih kaya (`ScrollTrigger`, transisi hero/galeri) sambil menjaga CLS, kecepatan input, dan aksesibilitas keyboard.

Komponen: Button primary/secondary/destructive, form field/validation, table+server pagination, status badge berlabel teks, drawer/modal, tabs, toast/error, skeleton, date/currency input, accessible chart dan mobile product gallery.

## Accessibility
Target **WCAG 2.2 AA sebagai usulan**, perlu disetujui owner; semantik heading, keyboard focus, contrast, status tidak bergantung pada warna, form error terkait field, alt text gambar produk, responsive zoom, ukuran target sentuh. Hindari screenshot yang mengandung data supplier atau customer sungguhan.

## Usability test plan
| Scenario | Kriteria awal |
|---|---|
| Operator membuat item+varian | Tidak membuat SKU duplikat, pesan error jelas |
| Supplier upload ASN | Scope dan status delivery benar |
| Manager channel melihat event gagal | Bisa membedakan gagal signature vs retry |
| Owner membaca dashboard | Mengetahui definisi net sales, watermark dan N/A |
| Customer checkout | Memahami total, expiry quote dan pending order |

Catat tanggal, persona, device, consent, success/fail, waktu, error, friksi dan revisi prioritas. **Belum ada hasil uji pengguna pada dokumen ini.** Figma, wireframe dan mockup final menjadi artefak terpisah setelah desain nyata, bukan dianggap sudah dibuat oleh Markdown.
