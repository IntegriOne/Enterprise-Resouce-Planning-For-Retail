"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import rawStyles from "./verify-email.module.css";
import {
  fetchBrowserSession,
  formatAuthErrorMessage,
  submitResendValidationEmailToken,
  submitVerifyEmail,
  type EmailVerificationStatus,
} from "@/lib/graphql/iam-browser-auth";

const styles = new Proxy((rawStyles ?? {}) as Record<string, string>, {
  get(target, prop: string) {
    return target[prop] || prop;
  },
});

type VerifyState =
  | "checking" // Pemeriksaan awal token fragment di browser
  | "needs_verification" // Pengalihan pasca-login belum terverifikasi atau akses tanpa token
  | "ready" // Memiliki token, menunggu klik konfirmasi
  | "verifying" // Sedang memproses mutasi
  | "VERIFIED" // Berhasil diverifikasi
  | "EXPIRED" // Token kedaluwarsa
  | "ALREADY_USED" // Token sudah pernah digunakan
  | "INVALID" // Token tidak valid
  | "service_error"; // Error jaringan / 500 / CSRF

function Icon({
  name,
  size = 24,
}: {
  name: "check" | "alert" | "clock" | "help" | "shield" | "mail";
  size?: number;
}) {
  const common = {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 2,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true as const,
    focusable: false as const,
  };

  switch (name) {
    case "check":
      return (
        <svg {...common}>
          <path d="M20 6 9 17l-5-5" />
        </svg>
      );
    case "clock":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
      );
    case "alert":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      );
    case "help":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="10" />
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
          <line x1="12" y1="17" x2="12.01" y2="17" />
        </svg>
      );
    case "shield":
      return (
        <svg {...common}>
          <path d="M12 2 4.5 5v6c0 5.1 3.2 8.4 7.5 11 4.3-2.6 7.5-5.9 7.5-11V5L12 2Z" />
        </svg>
      );
    case "mail":
      return (
        <svg {...common}>
          <rect width="20" height="16" x="2" y="4" rx="2" />
          <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
        </svg>
      );
    default:
      return null;
  }
}

function BrandMark() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
      focusable="false"
    >
      <rect x="3" y="3" width="8" height="8" rx="2" fill="#F9FAFB" />
      <rect x="13" y="3" width="8" height="8" rx="2" fill="#DA414A" />
      <rect
        x="3"
        y="13"
        width="8"
        height="8"
        rx="2"
        fill="rgba(249,250,251,0.65)"
      />
      <rect x="13" y="13" width="8" height="8" rx="2" fill="#F9FAFB" />
    </svg>
  );
}

export default function VerifyEmailExperience() {
  const [state, setState] = useState<VerifyState>("checking");
  const [csrfToken, setCsrfToken] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [resendEmail, setResendEmail] = useState("");
  const [resendLoading, setResendLoading] = useState(false);
  const [resendNotice, setResendNotice] = useState<string | null>(null);
  const [resendError, setResendError] = useState<string | null>(null);

  // Simpan token sementara hanya di ref memori komponen (tidak di localStorage/sessionStorage/cookie)
  const tokenRef = useRef<string | null>(null);
  const submitLockRef = useRef(false);

  useEffect(() => {
    if (typeof window === "undefined") return;

    let timer: ReturnType<typeof setTimeout> | null = null;

    const consumeHashFragment = (fromHashChange = false) => {
      const hash = window.location.hash;
      const hasHashFragment = Boolean(hash && hash.length > 1);
      if (fromHashChange && !hasHashFragment) {
        return;
      }

      let extractedToken: string | null = null;
      if (hash && hash.startsWith("#")) {
        const params = new URLSearchParams(hash.slice(1));
        const rawToken = params.get("token");
        if (rawToken && rawToken.trim()) {
          extractedToken = rawToken.trim();
        }
      }

      // SEGERA bersihkan fragment dari URL browser sebelum request atau tindakan apa pun
      // untuk mencegah kebocoran token ke log browser / history / referer.
      if (window.history && window.history.replaceState && hasHashFragment) {
        window.history.replaceState(
          null,
          "",
          window.location.pathname + window.location.search,
        );
      }

      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        if (!extractedToken) {
          setState(hasHashFragment ? "INVALID" : "needs_verification");
        } else {
          tokenRef.current = extractedToken;
          setState("ready");
        }
      }, 0);
    };

    consumeHashFragment(false);
    const onHashChange = () => consumeHashFragment(true);
    window.addEventListener("hashchange", onHashChange);

    // Bootstrap pre-auth session untuk CSRF token
    void fetchBrowserSession()
      .then((session) => {
        setCsrfToken(session.csrfToken);
      })
      .catch(() => {
        // Abaikan sementara jika tertunda
      });

    return () => {
      if (timer) clearTimeout(timer);
      window.removeEventListener("hashchange", onHashChange);
    };
  }, []);

  async function handleResendSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (resendLoading) return;

    const trimmedEmail = resendEmail.trim();
    setResendEmail(trimmedEmail);
    setResendError(null);
    setResendNotice(null);

    if (!trimmedEmail || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      setResendError("Masukkan alamat email yang valid untuk menerima tautan baru.");
      return;
    }

    setResendLoading(true);
    try {
      let activeCsrf = csrfToken;
      if (!activeCsrf) {
        try {
          const session = await fetchBrowserSession();
          activeCsrf = session.csrfToken;
          setCsrfToken(activeCsrf);
        } catch {
          // Lanjutkan jika bootstrap belum mengembalikan token eksplisit
        }
      }

      await submitResendValidationEmailToken(trimmedEmail, activeCsrf);
      setResendNotice(
        "Jika alamat email terdaftar dan belum terverifikasi, tautan verifikasi baru telah dikirim ke kotak masuk Anda.",
      );
    } catch (err) {
      setResendError(formatAuthErrorMessage(err, "verify"));
    } finally {
      setResendLoading(false);
    }
  }

  async function handleConfirmVerification() {
    if (submitLockRef.current || state === "verifying") return;

    const token = tokenRef.current;
    if (!token) {
      setState("INVALID");
      return;
    }

    submitLockRef.current = true;
    setState("verifying");
    setErrorMessage(null);

    try {
      let activeCsrf = csrfToken;
      if (!activeCsrf) {
        try {
          const session = await fetchBrowserSession();
          activeCsrf = session.csrfToken;
          setCsrfToken(activeCsrf);
        } catch {
          // Lanjutkan jika pre-auth tidak mengembalikan CSRF eksplisit
        }
      }

      const payload = await submitVerifyEmail(token, activeCsrf);

      // Kosongkan token dari ref memori setelah respons didapat
      tokenRef.current = null;

      // 4 outcome views yang berbeda
      const statusMap: Record<EmailVerificationStatus, VerifyState> = {
        VERIFIED: "VERIFIED",
        EXPIRED: "EXPIRED",
        ALREADY_USED: "ALREADY_USED",
        INVALID: "INVALID",
      };

      if (payload.status && payload.status in statusMap) {
        setState(statusMap[payload.status]);
      } else {
        setState("INVALID");
      }
    } catch (err) {
      // Jangan ubah menjadi INVALID atau VERIFIED saat terjadi network / 500 error
      setErrorMessage(formatAuthErrorMessage(err, "verify"));
      setState("service_error");
    } finally {
      submitLockRef.current = false;
    }
  }

  return (
    <div className={styles.pageShell}>
      {/* Header */}
      <header className={styles.topNav}>
        <div className={styles.topNavInner}>
          <Link href="/" className={styles.navBrandLockup}>
            <div className={styles.navBrandMark}>
              <BrandMark />
            </div>
            <div className={styles.navBrandText}>
              <span className={styles.navBrandName}>IntegriOne</span>
              <span className={styles.navBrandTag}>Koleksi &amp; Layanan Akun</span>
            </div>
          </Link>
          <div className={styles.navRight}>
            <Link href="/login" className={styles.navLink}>
              Masuk ke Akun
            </Link>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className={styles.mainContainer}>
        <div className={styles.card}>
          {state === "checking" && (
            <div className={styles.cardContent}>
              <div className={styles.spinner} aria-hidden="true" />
              <h1 className={styles.heading}>Menyiapkan Halaman Verifikasi...</h1>
              <p className={styles.description}>
                Mohon tunggu sejenak, kami sedang memeriksa tautan verifikasi
                Anda.
              </p>
            </div>
          )}

          {state === "needs_verification" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconWarning}`}>
                <Icon name="mail" size={32} />
              </div>
              <h1 className={styles.heading}>Verifikasi Email Diperlukan</h1>
              <p className={styles.description}>
                Akun Anda belum dapat masuk sebelum alamat email terverifikasi.
                Silakan buka tautan verifikasi pada kotak masuk email Anda atau
                minta pengiriman ulang tautan di bawah ini.
              </p>

              {resendNotice && (
                <div
                  role="status"
                  aria-live="polite"
                  className={styles.feedbackBannerSuccess}
                >
                  {resendNotice}
                </div>
              )}

              {resendError && (
                <div role="alert" className={styles.feedbackBannerError}>
                  {resendError}
                </div>
              )}

              <form
                onSubmit={handleResendSubmit}
                className={styles.resendForm}
                noValidate
              >
                <div className={styles.fieldGroup}>
                  <label htmlFor="verify-resend-email" className={styles.fieldLabel}>
                    Alamat email akun
                  </label>
                  <input
                    id="verify-resend-email"
                    name="email"
                    type="email"
                    autoComplete="email"
                    value={resendEmail}
                    onChange={(e) => setResendEmail(e.target.value)}
                    placeholder="nama@domain.com"
                    disabled={resendLoading}
                    className={styles.inputField}
                  />
                </div>
                <button
                  type="submit"
                  disabled={resendLoading}
                  className={styles.primaryBtn}
                >
                  {resendLoading
                    ? "Mengirim Tautan..."
                    : "Kirim Ulang Tautan Verifikasi"}
                </button>
              </form>

              <Link href="/login" className={styles.secondaryBtn}>
                Kembali ke Halaman Masuk
              </Link>
            </div>
          )}

          {state === "ready" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconInfo}`}>
                <Icon name="mail" size={32} />
              </div>
              <h1 className={styles.heading}>Konfirmasi Verifikasi Email</h1>
              <p className={styles.description}>
                Tekan tombol di bawah ini untuk memverifikasi alamat email akun
                Anda dan mengaktifkan akses masuk ke IntegriOne.
              </p>
              <button
                type="button"
                onClick={handleConfirmVerification}
                className={styles.primaryBtn}
              >
                Verifikasi Email Sekarang
              </button>
            </div>
          )}

          {state === "verifying" && (
            <div className={styles.cardContent}>
              <div className={styles.spinner} aria-hidden="true" />
              <h1 className={styles.heading}>Memverifikasi Email...</h1>
              <p className={styles.description}>
                Mohon tunggu sejenak, kami sedang memvalidasi token verifikasi
                Anda.
              </p>
            </div>
          )}

          {state === "VERIFIED" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconSuccess}`}>
                <Icon name="check" size={32} />
              </div>
              <h1 className={styles.heading}>Email Berhasil Diverifikasi!</h1>
              <p className={styles.description}>
                Alamat email Anda telah berhasil diverifikasi. Akun Anda kini
                aktif dan siap digunakan di IntegriOne.
              </p>
              <Link href="/login" className={styles.primaryBtn}>
                Masuk ke Akun Anda
              </Link>
            </div>
          )}

          {state === "EXPIRED" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconWarning}`}>
                <Icon name="clock" size={32} />
              </div>
              <h1 className={styles.heading}>Tautan Kedaluwarsa</h1>
              <p className={styles.description}>
                Masa berlaku tautan verifikasi ini telah berakhir demi keamanan
                akun Anda. Silakan daftar ulang atau minta tautan baru melalui
                halaman pendaftaran.
              </p>
              <div className={styles.btnRow}>
                <button
                  type="button"
                  onClick={() => {
                    setResendError(null);
                    setResendNotice(null);
                    setState("needs_verification");
                  }}
                  className={styles.primaryBtn}
                >
                  Kirim Ulang Tautan Verifikasi
                </button>
                <Link href="/login?mode=register" className={styles.secondaryBtn}>
                  Daftar Ulang
                </Link>
                <Link href="/login" className={styles.secondaryBtn}>
                  Ke Halaman Masuk
                </Link>
              </div>
            </div>
          )}

          {state === "ALREADY_USED" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconInfo}`}>
                <Icon name="check" size={32} />
              </div>
              <h1 className={styles.heading}>Tautan Sudah Digunakan</h1>
              <p className={styles.description}>
                Tautan verifikasi ini sudah pernah digunakan sebelumnya. Akun
                Anda telah aktif. Anda dapat langsung masuk menggunakan email
                dan kata sandi Anda.
              </p>
              <Link href="/login" className={styles.primaryBtn}>
                Masuk Sekarang
              </Link>
            </div>
          )}

          {state === "INVALID" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconDanger}`}>
                <Icon name="alert" size={32} />
              </div>
              <h1 className={styles.heading}>Tautan Tidak Valid</h1>
              <p className={styles.description}>
                Tautan verifikasi tidak lengkap, rusak, atau format tidak
                dikenali. Pastikan Anda membuka tautan utuh yang tercantum pada
                email Anda.
              </p>
              <Link href="/login" className={styles.primaryBtn}>
                Kembali ke Halaman Masuk
              </Link>
            </div>
          )}

          {state === "service_error" && (
            <div className={styles.cardContent}>
              <div className={`${styles.iconWrap} ${styles.iconDanger}`}>
                <Icon name="alert" size={32} />
              </div>
              <h1 className={styles.heading}>Layanan Sedang Terganggu</h1>
              <p className={styles.description}>
                {errorMessage ||
                  "Terjadi kendala saat menghubungkan ke layanan verifikasi. Silakan coba kembali beberapa saat lagi."}
              </p>
              <div className={styles.btnRow}>
                <button
                  type="button"
                  onClick={handleConfirmVerification}
                  className={styles.primaryBtn}
                >
                  Coba Lagi
                </button>
                <Link href="/login" className={styles.secondaryBtn}>
                  Ke Halaman Masuk
                </Link>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
