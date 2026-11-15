"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import rawStyles from "./register.module.css";
import {
  fetchBrowserSession,
  formatAuthErrorMessage,
  getPasswordStrength,
  submitCreateUser,
  submitResendValidationEmailToken,
} from "@/lib/graphql/iam-browser-auth";

const styles = new Proxy((rawStyles ?? {}) as Record<string, string>, {
  get(target, prop: string) {
    return target[prop] || prop;
  },
});

function Icon({
  name,
  size = 18,
}: {
  name: "eye" | "eyeOff" | "check" | "alert" | "mail" | "arrowLeft";
  size?: number;
}) {
  const common = {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true as const,
    focusable: false as const,
  };

  switch (name) {
    case "eye":
      return (
        <svg {...common}>
          <path d="M2 12s3.6-6 10-6 10 6 10 6-3.6 6-10 6-10-6-10-6Z" />
          <circle cx="12" cy="12" r="2.5" />
        </svg>
      );
    case "eyeOff":
      return (
        <svg {...common}>
          <path d="m3 3 18 18M10.6 6.1A10.9 10.9 0 0 1 12 6c6.4 0 10 6 10 6a15.4 15.4 0 0 1-3.1 3.5M6.5 6.5C3.6 8.4 2 12 2 12s3.6 6 10 6c1.6 0 3-.4 4.2-1" />
          <path d="M10.6 10.6a2 2 0 0 0 2.8 2.8" />
        </svg>
      );
    case "check":
      return (
        <svg {...common}>
          <path d="m5 12 4 4 10-10" />
        </svg>
      );
    case "mail":
      return (
        <svg {...common}>
          <rect width="20" height="16" x="2" y="4" rx="2" />
          <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
        </svg>
      );
    case "alert":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="9" />
          <path d="M12 8v4" />
          <path d="M12 16h.01" />
        </svg>
      );
    case "arrowLeft":
      return (
        <svg {...common}>
          <path d="M19 12H5m6-6-6 6 6 6" />
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

export default function RegisterExperience() {
  const router = useRouter();
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const [csrfToken, setCsrfToken] = useState<string | null>(null);
  const [isInitializing, setIsInitializing] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<{
    displayName?: string;
    email?: string;
    username?: string;
    password?: string;
  }>({});

  const [registeredEmail, setRegisteredEmail] = useState<string | null>(null);
  const [isResending, setIsResending] = useState(false);
  const [resendMessage, setResendMessage] = useState<string | null>(null);
  const submitLockRef = useRef(false);

  useEffect(() => {
    const controller = new AbortController();
    let active = true;

    async function bootstrap() {
      try {
        const session = await fetchBrowserSession(controller.signal);
        if (active) {
          setCsrfToken(session.csrfToken);
        }
      } catch {
        // Fallback jika layanan tertunda
      } finally {
        if (active) {
          setIsInitializing(false);
        }
      }
    }

    void bootstrap();
    return () => {
      active = false;
      controller.abort();
    };
  }, []);

  async function ensureCsrfToken(): Promise<string | null> {
    if (csrfToken) return csrfToken;
    try {
      const session = await fetchBrowserSession();
      setCsrfToken(session.csrfToken);
      return session.csrfToken;
    } catch {
      return null;
    }
  }

  async function handleRegisterSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting) return;

    const trimmedName = displayName.trim();
    const trimmedEmail = email.trim();
    const trimmedUsername = username.trim().toLowerCase();
    setErrorMessage(null);
    setFieldErrors({});

    const errors: {
      displayName?: string;
      email?: string;
      username?: string;
      password?: string;
    } = {};
    if (!trimmedName) {
      errors.displayName = "Masukkan nama lengkap Anda.";
    }
    if (!trimmedEmail) {
      errors.email = "Masukkan alamat email Anda.";
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      errors.email = "Format alamat email tidak valid.";
    }
    if (!trimmedUsername) {
      errors.username = "Masukkan username Anda.";
    } else if (!/^[a-z0-9][a-z0-9._-]{1,62}[a-z0-9]$/.test(trimmedUsername)) {
      errors.username =
        "Username minimal 3 karakter (huruf kecil, angka, titik, garis bawah, atau tanda hubung).";
    }
    if (!password) {
      errors.password = "Masukkan kata sandi akun Anda.";
    } else if (password.length < 10) {
      errors.password = "Kata sandi minimal 10 karakter.";
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }

    submitLockRef.current = true;
    setIsSubmitting(true);

    try {
      const activeCsrf = await ensureCsrfToken();
      const { payload, locationHeader } = await submitCreateUser(
        trimmedName,
        trimmedEmail,
        trimmedUsername,
        password,
        activeCsrf,
      );

      // Kosongkan password sensitif dari memori komponen
      setPassword("");

      // Alihkan ke halaman login jika server mengembalikan redirect /login untuk mencegah enumerasi email
      const redirectTarget =
        locationHeader?.trim() || payload.redirectLocation?.trim();
      if (redirectTarget === "/login") {
        router.push("/login");
        return;
      }

      setRegisteredEmail(trimmedEmail);
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "register"));
      try {
        const refreshed = await fetchBrowserSession();
        setCsrfToken(refreshed.csrfToken);
      } catch {
        // Abaikan kegagalan refresh
      }
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  async function handleResendVerification() {
    if (!registeredEmail || isResending) return;
    setIsResending(true);
    setResendMessage(null);
    try {
      const activeCsrf = await ensureCsrfToken();
      await submitResendValidationEmailToken(registeredEmail, activeCsrf);
      setResendMessage(
        "Tautan verifikasi baru telah dikirim. Token sebelumnya otomatis tidak berlaku.",
      );
    } catch (err) {
      setResendMessage(formatAuthErrorMessage(err, "verify"));
    } finally {
      setIsResending(false);
    }
  }

  const passwordStrength = getPasswordStrength(password);

  return (
    <div className={styles.enterpriseShell}>
      {/* Sticky Header */}
      <header className={styles.topNav}>
        <div className={styles.topNavInner}>
          <Link href="/" className={styles.navBrandLockup}>
            <div className={styles.navBrandMark}>
              <BrandMark />
            </div>
            <div className={styles.navBrandText}>
              <span className={styles.navBrandName}>IntegriOne</span>
              <span className={styles.navBrandTag}>Identity &amp; Gateway</span>
            </div>
          </Link>

          <div className={styles.navRight}>
            <Link href="/login" className={styles.loginLinkButton}>
              Sudah punya akun? Masuk
            </Link>
          </div>
        </div>
      </header>

      {/* Main Two-Column Stage */}
      <div className={styles.mainStage}>
        <div className={styles.splitContainer}>
          {/* Brand Aside */}
          <aside className={styles.brandPanel} aria-label="Identitas Pendaftaran">
            <div className={styles.brandGridOverlay} aria-hidden="true" />
            <div className={styles.brandGrainOverlay} aria-hidden="true" />

            <div className={styles.editorialTop}>
              <div className={styles.editorialKickerRow}>
                <span className={styles.coralRule} aria-hidden="true" />
                <span className={styles.editorialKicker}>INTEGRIONE</span>
              </div>
              <span className={styles.editorialSubKicker}>
                UNIFIED ACCOUNT ENROLLMENT
              </span>
            </div>

            <div className={styles.editorialCenter}>
              <h1 className={styles.editorialHeadline}>
                <span className={styles.editorialHeadlineLine}>
                  Bergabung bersama
                </span>
                <span className={styles.editorialHeadlineItalic}>
                  IntegriOne.
                </span>
              </h1>
              <p className={styles.editorialLead}>
                Daftarkan akun untuk akses aman ke portal identitas IntegriOne
                serta layanan Gateway terpadu.
              </p>
            </div>

            <div className={styles.editorialBottom}>
              <div className={styles.fadingRule} aria-hidden="true" />
              <div className={styles.editorialFooterRow}>
                <p className={styles.editorialQuote}>
                  Akses akun aman dengan sesi terproteksi.
                </p>
                <span className={styles.editorialIndexTag}>
                  02 / REGISTRATION
                </span>
              </div>
            </div>
          </aside>

          {/* Form Main Panel */}
          <main className={styles.authPanel}>
            <div className={styles.authContainer}>
              {registeredEmail ? (
                /* Sukses - Verifikasi Email Pending Card */
                <div className={styles.successCard}>
                  <div className={styles.successIconWrap}>
                    <Icon name="mail" size={32} />
                  </div>
                  <h2 className={styles.formHeading}>Periksa email Anda.</h2>
                  <p className={styles.formSubtitle}>
                    Kami telah mengirimkan tautan verifikasi ke{" "}
                    <strong>{registeredEmail}</strong>.
                  </p>
                  <div className={styles.noticeBox}>
                    <p className={styles.noticeText}>
                      Akun Anda belum aktif. Buka email Anda dan klik tombol
                      verifikasi untuk mengaktifkan akun sebelum masuk.
                    </p>
                  </div>
                  {resendMessage && (
                    <div className={styles.noticeBox}>
                      <p className={styles.noticeText}>{resendMessage}</p>
                    </div>
                  )}
                  <button
                    type="button"
                    onClick={() => void handleResendVerification()}
                    disabled={isResending}
                    className={styles.loginLinkButton}
                  >
                    {isResending
                      ? "Mengirim ulang..."
                      : "Kirim ulang tautan verifikasi"}
                  </button>
                  <Link href="/login" className={styles.primaryButton}>
                    Kembali ke Halaman Masuk
                  </Link>
                </div>
              ) : (
                /* Form Pendaftaran */
                <>
                  <header className={styles.formHeader}>
                    <div className={styles.formHeaderText}>
                      <div className={styles.formEyebrow}>
                        <span
                          className={styles.formEyebrowAccent}
                          aria-hidden="true"
                        />
                        <span>PENDAFTARAN AKUN</span>
                      </div>
                      <h2 className={styles.formHeading}>Buat akun baru.</h2>
                      <p className={styles.formSubtitle}>
                        Lengkapi data di bawah ini untuk memulai.
                      </p>
                    </div>
                    <span className={styles.oversizedIndex} aria-hidden="true">
                      02
                    </span>
                  </header>

                  {errorMessage && (
                    <div
                      role="alert"
                      className={`${styles.alertBox} ${styles.alertError}`}
                    >
                      <span className={styles.alertIcon}>
                        <Icon name="alert" size={18} />
                      </span>
                      <span>{errorMessage}</span>
                    </div>
                  )}

                  <form
                    onSubmit={handleRegisterSubmit}
                    className={styles.formStack}
                    noValidate
                  >
                    <div className={styles.fieldGroup}>
                      <label htmlFor="reg-name" className={styles.fieldLabel}>
                        Nama lengkap
                      </label>
                      <input
                        id="reg-name"
                        name="name"
                        type="text"
                        autoComplete="name"
                        value={displayName}
                        onChange={(e) => {
                          setDisplayName(e.target.value);
                          if (fieldErrors.displayName) {
                            setFieldErrors((prev) => ({
                              ...prev,
                              displayName: undefined,
                            }));
                          }
                        }}
                        placeholder="Contoh: Budi Santoso"
                        disabled={isSubmitting}
                        className={`${styles.inputField} ${
                          fieldErrors.displayName ? styles.inputFieldError : ""
                        }`}
                        aria-invalid={Boolean(fieldErrors.displayName)}
                        aria-describedby={
                          fieldErrors.displayName ? "name-error" : undefined
                        }
                      />
                      {fieldErrors.displayName && (
                        <p id="name-error" className={styles.fieldErrorText}>
                          {fieldErrors.displayName}
                        </p>
                      )}
                    </div>

                    <div className={styles.fieldGroup}>
                      <label htmlFor="reg-email" className={styles.fieldLabel}>
                        Alamat email
                      </label>
                      <input
                        id="reg-email"
                        name="email"
                        type="email"
                        autoComplete="email"
                        value={email}
                        onChange={(e) => {
                          setEmail(e.target.value);
                          if (fieldErrors.email) {
                            setFieldErrors((prev) => ({
                              ...prev,
                              email: undefined,
                            }));
                          }
                        }}
                        placeholder="nama@domain.com"
                        disabled={isSubmitting}
                        className={`${styles.inputField} ${
                          fieldErrors.email ? styles.inputFieldError : ""
                        }`}
                        aria-invalid={Boolean(fieldErrors.email)}
                        aria-describedby={
                          fieldErrors.email ? "email-error" : undefined
                        }
                      />
                      {fieldErrors.email && (
                        <p id="email-error" className={styles.fieldErrorText}>
                          {fieldErrors.email}
                        </p>
                      )}
                    </div>

                    <div className={styles.fieldGroup}>
                      <label htmlFor="reg-username" className={styles.fieldLabel}>
                        Username
                      </label>
                      <input
                        id="reg-username"
                        name="username"
                        type="text"
                        autoComplete="username"
                        value={username}
                        onChange={(e) => {
                          setUsername(e.target.value);
                          if (fieldErrors.username) {
                            setFieldErrors((prev) => ({
                              ...prev,
                              username: undefined,
                            }));
                          }
                        }}
                        placeholder="contoh: budi.santoso"
                        disabled={isSubmitting}
                        className={`${styles.inputField} ${
                          fieldErrors.username ? styles.inputFieldError : ""
                        }`}
                        aria-invalid={Boolean(fieldErrors.username)}
                        aria-describedby={
                          fieldErrors.username ? "username-error" : undefined
                        }
                      />
                      {fieldErrors.username && (
                        <p id="username-error" className={styles.fieldErrorText}>
                          {fieldErrors.username}
                        </p>
                      )}
                    </div>

                    <div className={styles.fieldGroup}>
                      <label
                        htmlFor="reg-password"
                        className={styles.fieldLabel}
                      >
                        Kata sandi
                      </label>
                      <div className={styles.inputWrap}>
                        <input
                          id="reg-password"
                          name="password"
                          type={showPassword ? "text" : "password"}
                          autoComplete="new-password"
                          value={password}
                          onChange={(e) => {
                            setPassword(e.target.value);
                            if (fieldErrors.password) {
                              setFieldErrors((prev) => ({
                                ...prev,
                                password: undefined,
                              }));
                            }
                          }}
                          placeholder="Minimal 10 karakter"
                          disabled={isSubmitting}
                          className={`${styles.inputField} ${styles.inputFieldWithToggle} ${
                            fieldErrors.password ? styles.inputFieldError : ""
                          }`}
                          aria-invalid={Boolean(fieldErrors.password)}
                          aria-describedby={
                            fieldErrors.password ? "password-error" : undefined
                          }
                        />
                        <button
                          type="button"
                          onClick={() => setShowPassword((prev) => !prev)}
                          aria-label={
                            showPassword
                              ? "Sembunyikan kata sandi"
                              : "Tampilkan kata sandi"
                          }
                          aria-pressed={showPassword}
                          disabled={isSubmitting}
                          className={styles.passwordToggle}
                        >
                          <Icon
                            name={showPassword ? "eyeOff" : "eye"}
                            size={18}
                          />
                        </button>
                      </div>
                      {fieldErrors.password && (
                        <p
                          id="password-error"
                          className={styles.fieldErrorText}
                        >
                          {fieldErrors.password}
                        </p>
                      )}

                      {password && (
                        <div className={styles.strengthContainer}>
                          <div className={styles.strengthBarBg}>
                            <div
                              className={styles.strengthBarFill}
                              style={{
                                width: `${passwordStrength.score}%`,
                                backgroundColor: passwordStrength.color,
                              }}
                            />
                          </div>
                          <span
                            className={styles.strengthLabel}
                            style={{ color: passwordStrength.color }}
                          >
                            Kekuatan: {passwordStrength.label}
                          </span>
                        </div>
                      )}
                    </div>

                    <div className={styles.sessionNote}>
                      <span
                        className={styles.sessionNoteDot}
                        aria-hidden="true"
                      />
                      <span>
                        Dengan mendaftar, Anda menyetujui kebijakan privasi dan
                        ketentuan layanan IntegriOne.
                      </span>
                    </div>

                    <button
                      type="submit"
                      disabled={isSubmitting || isInitializing}
                      className={styles.primaryButton}
                    >
                      {isSubmitting ? (
                        <>
                          <span className={styles.spinner} aria-hidden="true" />
                          <span>Mendaftarkan akun...</span>
                        </>
                      ) : isInitializing ? (
                        <span>Menyiapkan sesi...</span>
                      ) : (
                        <span>Daftar Akun</span>
                      )}
                    </button>
                  </form>

                  <div className={styles.loginPrompt}>
                    <span>Sudah memiliki akun?</span>{" "}
                    <Link href="/login" className={styles.loginLink}>
                      Masuk di sini
                    </Link>
                  </div>
                </>
              )}
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
