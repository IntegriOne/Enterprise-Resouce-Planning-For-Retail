"use client";

import Link from "next/link";
import { BrandMark, Icon } from "./icons";

type StylesMap = Record<string, string>;

export function LoginFootbar({
  styles,
  isRegisterStep,
  onToggleAuthView,
  onOpenResetPassword,
}: {
  styles: StylesMap;
  isRegisterStep: boolean;
  onToggleAuthView: () => void;
  onOpenResetPassword: () => void;
}) {
  function handleScrollToTop() {
    if (typeof window === "undefined") return;
    const reduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    window.scrollTo({
      top: 0,
      behavior: reduced ? "auto" : "smooth",
    });
  }

  return (
    <footer className={styles.footbar}>
      <div className={styles.footbarInner}>
        <div className={styles.footbarGrid}>
          {/* Kolom identitas brand & ringkasan keamanan sesi */}
          <div className={styles.footbarBrandCol}>
            <div className={styles.footbarBrandHeader}>
              <span className={styles.footbarBrandMark} aria-hidden="true">
                <BrandMark />
              </span>
              <div className={styles.footbarBrandTitleWrap}>
                <span className={styles.footbarBrand}>IntegriOne</span>
                <span className={styles.footbarBrandKicker}>
                  RUANG TERPADU RITEL
                </span>
              </div>
            </div>

            <p className={styles.footbarDesc}>
              Satu akses aman untuk berbelanja, memantau pesanan, dan mengelola
              akun IntegriOne Anda.
            </p>

            <div className={styles.footbarAssurance}>
              <span className={styles.footbarAssuranceIcon} aria-hidden="true">
                <Icon name="shield" size={16} />
              </span>
              <span>Sesi dilindungi cookie terenkripsi dan proteksi CSRF</span>
            </div>
          </div>

          {/* Kolom navigasi akses akun */}
          <div className={styles.footbarCol}>
            <p className={styles.footbarHeading}>Akses &amp; Akun</p>
            <nav
              aria-label="Navigasi kaki halaman"
              className={styles.footbarNav}
            >
              <Link href="/" className={styles.footbarLink}>
                <span>Beranda Toko</span>
                <span className={styles.footbarLinkArrow} aria-hidden="true">
                  <Icon name="arrowRight" size={14} />
                </span>
              </Link>
              <button
                type="button"
                onClick={onToggleAuthView}
                className={styles.footbarLink}
              >
                <span>{isRegisterStep ? "Masuk ke Akun" : "Daftar Akun"}</span>
                <span className={styles.footbarLinkArrow} aria-hidden="true">
                  <Icon name="arrowRight" size={14} />
                </span>
              </button>
              <button
                type="button"
                onClick={onOpenResetPassword}
                className={styles.footbarLink}
              >
                <span>Lupa Kata Sandi</span>
                <span className={styles.footbarLinkArrow} aria-hidden="true">
                  <Icon name="arrowRight" size={14} />
                </span>
              </button>
            </nav>
          </div>

          {/* Kolom ringkasan standar keamanan halaman masuk */}
          <div className={styles.footbarCol}>
            <p className={styles.footbarHeading}>Keamanan Sesi</p>
            <dl className={styles.footbarSpecList}>
              <div className={styles.footbarSpecRow}>
                <dt className={styles.footbarSpecTerm}>Akses Halaman</dt>
                <dd className={styles.footbarSpecVal}>
                  Satu pintu masuk akun
                </dd>
              </div>
              <div className={styles.footbarSpecRow}>
                <dt className={styles.footbarSpecTerm}>Masa Sesi</dt>
                <dd className={styles.footbarSpecValMono}>
                  30m diam · maks. 8j
                </dd>
              </div>
              <div className={styles.footbarSpecRow}>
                <dt className={styles.footbarSpecTerm}>Autentikasi</dt>
                <dd className={styles.footbarSpecVal}>
                  Email / username + TOTP 2FA
                </dd>
              </div>
            </dl>
          </div>
        </div>

        {/* Baris penutup hak cipta & navigasi kembali ke atas */}
        <div className={styles.footbarMeta}>
          <div className={styles.footbarCopyrightWrap}>
            <span className={styles.footbarStatusDot} aria-hidden="true" />
            <p className={styles.footbarCopyright}>
              &copy; 2026 IntegriOne. Hak cipta dilindungi.
            </p>
          </div>

          <button
            type="button"
            onClick={handleScrollToTop}
            className={styles.footbarTopBtn}
          >
            Kembali ke atas
          </button>
        </div>
      </div>
    </footer>
  );
}
