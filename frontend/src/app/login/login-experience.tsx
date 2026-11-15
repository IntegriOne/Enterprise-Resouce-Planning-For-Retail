"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import rawStyles from "./login.module.css";
import {
  fetchBrowserSession,
  formatAuthErrorMessage,
  getPasswordStrength,
  submitResetPassword,
  submitUserLogin,
  submitUserResetPassword,
  submitValidateUserMfa,
} from "@/lib/graphql/iam-browser-auth";
import { validatePostLoginLocation } from "@/lib/auth/redirect-allowlist";

const styles = new Proxy((rawStyles ?? {}) as Record<string, string>, {
  get(target, prop: string) {
    return target[prop] || prop;
  },
});

type Step =
  | "credentials"
  | "mfa"
  | "reset-request"
  | "reset-confirm"
  | "reset-done"
  | "navigating";

type MfaMode = "totp" | "recovery";

function Icon({
  name,
  size = 18,
}: {
  name: "eye" | "eyeOff" | "shield" | "check" | "alert" | "arrowLeft";
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
    case "shield":
      return (
        <svg {...common}>
          <path d="M12 2 4.5 5v6c0 5.1 3.2 8.4 7.5 11 4.3-2.6 7.5-5.9 7.5-11V5L12 2Z" />
          <path d="m9.5 12 1.8 1.8 3.7-3.8" />
        </svg>
      );
    case "check":
      return (
        <svg {...common}>
          <path d="m5 12 4 4 10-10" />
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

export default function LoginExperience() {
  const [step, setStep] = useState<Step>("credentials");
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const [mfaMode, setMfaMode] = useState<MfaMode>("totp");
  const [mfaCode, setMfaCode] = useState("");
  const [mfaChallenge, setMfaChallenge] = useState<string | null>(null);

  const [resetIdentifier, setResetIdentifier] = useState("");
  const [resetToken, setResetToken] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [showNewPassword, setShowNewPassword] = useState(false);

  const [csrfToken, setCsrfToken] = useState<string | null>(null);
  const [isInitializing, setIsInitializing] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [identifierError, setIdentifierError] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [statusNotice, setStatusNotice] = useState<string | null>(null);

  const submitLockRef = useRef(false);

  useEffect(() => {
    const controller = new AbortController();
    let active = true;

    async function bootstrapSession() {
      try {
        const session = await fetchBrowserSession(controller.signal);
        if (!active) return;
        setCsrfToken(session.csrfToken);
        if (session.authenticated) {
          // Jika sesi sudah aktif, arahkan ke rute aman default
          setStep("navigating");
          window.location.replace("/identity");
        }
      } catch (err) {
        if (!active || (err instanceof Error && err.name === "AbortError")) {
          return;
        }
        setStatusNotice(
          "Koneksi awal ke layanan autentikasi tertunda. Anda tetap dapat mencoba masuk.",
        );
      } finally {
        if (active) {
          setIsInitializing(false);
        }
      }
    }

    void bootstrapSession();
    return () => {
      active = false;
      controller.abort();
    };
  }, []);

  async function ensureCsrfToken(): Promise<string | null> {
    if (csrfToken) return csrfToken;
    try {
      const session = await fetchBrowserSession();
      const fresh = session.csrfToken;
      setCsrfToken(fresh);
      return fresh;
    } catch {
      return null;
    }
  }

  function handleFinalNavigation(
    locationHeader: string | null,
    redirectLocationPayload?: string | null,
  ) {
    // Utamakan header Location HTTP 200, fallback ke redirectLocation dari payload GraphQL
    const candidate = locationHeader || redirectLocationPayload || "/";
    const validatedDestination = validatePostLoginLocation(candidate);

    if (validatedDestination) {
      setStep("navigating");
      window.location.replace(validatedDestination);
    } else {
      // Tolak jika tujuan tidak lolos allowlist
      setErrorMessage(
        "Tujuan pengalihan tidak diizinkan. Silakan muat ulang halaman.",
      );
    }
  }

  async function handleLoginSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting) return;

    const trimmedIdentifier = identifier.trim();
    setIdentifier(trimmedIdentifier);
    setErrorMessage(null);
    setIdentifierError(null);
    setPasswordError(null);

    let hasValidationError = false;
    if (!trimmedIdentifier) {
      setIdentifierError("Masukkan email atau username Anda.");
      hasValidationError = true;
    }
    if (!password) {
      setPasswordError("Masukkan kata sandi Anda.");
      hasValidationError = true;
    }
    if (hasValidationError) return;

    submitLockRef.current = true;
    setIsSubmitting(true);

    try {
      const activeCsrf = await ensureCsrfToken();
      const { payload, locationHeader } = await submitUserLogin(
        trimmedIdentifier,
        password,
        activeCsrf,
      );

      if (payload.csrfToken) {
        setCsrfToken(payload.csrfToken);
      }

      // Kosongkan password dari memori komponen
      setPassword("");

      if (payload.status === "MFA_REQUIRED" && payload.mfaChallenge) {
        setMfaChallenge(payload.mfaChallenge);
        setMfaCode("");
        setStep("mfa");
        return;
      }

      if (payload.status === "AUTHENTICATED") {
        handleFinalNavigation(locationHeader, payload.redirectLocation);
        return;
      }

      setErrorMessage("Respons autentikasi tidak dikenal. Silakan coba lagi.");
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "login"));
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

  async function handleMfaSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting || !mfaChallenge) return;

    const trimmedCode = mfaCode.trim();
    setErrorMessage(null);
    if (!trimmedCode) {
      setErrorMessage(
        mfaMode === "totp"
          ? "Masukkan kode 6 digit dari aplikasi autentikator."
          : "Masukkan kode pemulihan akun Anda.",
      );
      return;
    }

    submitLockRef.current = true;
    setIsSubmitting(true);

    try {
      const { payload, locationHeader } = await submitValidateUserMfa(
        mfaChallenge,
        trimmedCode,
        mfaMode,
        csrfToken,
      );

      if (payload.csrfToken) {
        setCsrfToken(payload.csrfToken);
      }
      setMfaChallenge(null);
      setMfaCode("");

      handleFinalNavigation(locationHeader, payload.redirectLocation);
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "mfa"));
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  async function handleResetRequestSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting) return;

    const trimmed = resetIdentifier.trim();
    setResetIdentifier(trimmed);
    setErrorMessage(null);

    if (!trimmed) {
      setErrorMessage("Masukkan alamat email akun Anda yang terverifikasi.");
      return;
    }

    submitLockRef.current = true;
    setIsSubmitting(true);

    try {
      const activeCsrf = await ensureCsrfToken();
      const resetPayload = await submitUserResetPassword(trimmed, activeCsrf);
      if (resetPayload.message) {
        setStatusNotice(resetPayload.message);
      }
      setResetToken("");
      setNewPassword("");
      setStep("reset-confirm");
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "reset-request"));
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  async function handleResetConfirmSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting) return;

    const trimmedToken = resetToken.trim();
    setErrorMessage(null);

    if (!trimmedToken || !newPassword) {
      setErrorMessage("Lengkapi token pemulihan dan kata sandi baru Anda.");
      return;
    }

    submitLockRef.current = true;
    setIsSubmitting(true);

    try {
      const activeCsrf = await ensureCsrfToken();
      await submitResetPassword(trimmedToken, newPassword, activeCsrf);
      setResetToken("");
      setNewPassword("");
      setStep("reset-done");
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "reset-confirm"));
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  function openForgotPassword() {
    setErrorMessage(null);
    setStatusNotice(null);
    setIdentifierError(null);
    setPasswordError(null);
    const trimmed = identifier.trim();
    setResetIdentifier(trimmed.includes("@") ? trimmed : "");
    setStep("reset-request");
  }

  async function returnToLogin() {
    setErrorMessage(null);
    setStatusNotice(null);
    setIdentifierError(null);
    setPasswordError(null);
    setMfaChallenge(null);
    setMfaCode("");
    setStep("credentials");
    try {
      const session = await fetchBrowserSession();
      setCsrfToken(session.csrfToken);
    } catch {
      // Abaikan jika tertunda
    }
  }

  const passwordStrength = getPasswordStrength(newPassword);

  return (
    <div className={styles.enterpriseShell}>
      {/* Sticky Navigation Header (68px) */}
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
            <span className={styles.navContextPill}>
              <span className={styles.navContextDot} aria-hidden="true" />
              <span>Satu Akses Terpadu</span>
            </span>
          </div>
        </div>
      </header>

      {/* Main Two-Column Editorial Stage */}
      <div className={styles.mainStage}>
        <div className={styles.splitContainer}>
          {/* Left Editorial Brand Panel */}
          <aside className={styles.brandPanel} aria-label="Identitas IntegriOne">
            <div className={styles.brandGridOverlay} aria-hidden="true" />
            <div className={styles.brandGrainOverlay} aria-hidden="true" />

            <div className={styles.editorialTop}>
              <div className={styles.editorialKickerRow}>
                <span className={styles.coralRule} aria-hidden="true" />
                <span className={styles.editorialKicker}>INTEGRIONE</span>
              </div>
              <span className={styles.editorialSubKicker}>
                IDENTITY &amp; GATEWAY PLATFORM
              </span>
            </div>

            <div className={styles.editorialCenter}>
              <h1 className={styles.editorialHeadline}>
                <span className={styles.editorialHeadlineLine}>
                  Satu akun.
                </span>
                <span className={styles.editorialHeadlineItalic}>
                  Beragam kemudahan.
                </span>
              </h1>
              <p className={styles.editorialLead}>
                Akses terpadu untuk autentikasi pengguna, otorisasi peran, dan
                kontrol sesi dalam arsitektur IAM serta Gateway yang aman.
              </p>

              <div className={styles.erpDiagram} aria-hidden="true">
                <div className={styles.erpDiagramHeader}>
                  <span>Arsitektur Akses Terpadu</span>
                  <span>Sesi Opaque Terpusat</span>
                </div>
                <div className={styles.erpModuleRow}>
                  <div className={styles.erpModuleCard}>
                    <span className={styles.erpModuleMarker} />
                    <span className={styles.erpModuleLabel}>
                      Autentikasi &amp; MFA
                    </span>
                  </div>
                  <div className={styles.erpModuleCard}>
                    <span
                      className={`${styles.erpModuleMarker} ${styles.erpModuleMarkerNavy}`}
                    />
                    <span className={styles.erpModuleLabel}>
                      Otorisasi &amp; Kebijakan
                    </span>
                  </div>
                  <div className={styles.erpModuleCard}>
                    <span
                      className={`${styles.erpModuleMarker} ${styles.erpModuleMarkerNavy}`}
                    />
                    <span className={styles.erpModuleLabel}>
                      Manajemen Sesi
                    </span>
                  </div>
                  <div className={styles.erpModuleCard}>
                    <span className={styles.erpModuleMarker} />
                    <span className={styles.erpModuleLabel}>
                      Gateway GraphQL
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div className={styles.editorialBottom}>
              <div className={styles.fadingRule} aria-hidden="true" />
              <div className={styles.editorialFooterRow}>
                <p className={styles.editorialQuote}>
                  Satu pintu masuk. Pengalaman terbaik untuk semua.
                </p>
                <span className={styles.editorialIndexTag}>
                  01 / UNIFIED ACCESS
                </span>
              </div>
            </div>
          </aside>

          {/* Right Sign-In Form Panel */}
          <main className={styles.authPanel}>
            <div className={styles.authContainer}>
              {step === "credentials" && (
                <>
                  <header className={styles.formHeader}>
                    <div className={styles.formHeaderText}>
                      <div className={styles.formEyebrow}>
                        <span
                          className={styles.formEyebrowAccent}
                          aria-hidden="true"
                        />
                        <span>MASUK KE AKUN</span>
                      </div>
                      <h2 className={styles.formHeading}>
                        Selamat datang kembali.
                      </h2>
                      <p className={styles.formSubtitle}>
                        Masuk untuk melanjutkan aktivitas Anda di IntegriOne.
                      </p>
                    </div>
                    <span className={styles.oversizedIndex} aria-hidden="true">
                      01
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

                  {!errorMessage && statusNotice && (
                    <div
                      role="status"
                      className={`${styles.alertBox} ${styles.alertWarning}`}
                    >
                      <span className={styles.alertIcon}>
                        <Icon name="alert" size={18} />
                      </span>
                      <span>{statusNotice}</span>
                    </div>
                  )}

                  <form
                    onSubmit={handleLoginSubmit}
                    className={styles.formStack}
                    noValidate
                  >
                    <div className={styles.fieldGroup}>
                      <label
                        htmlFor="login-identifier"
                        className={styles.fieldLabel}
                      >
                        Email atau username
                      </label>
                      <div className={styles.inputWrap}>
                        <input
                          id="login-identifier"
                          name="identifier"
                          type="text"
                          autoComplete="username"
                          autoCapitalize="none"
                          autoCorrect="off"
                          spellCheck={false}
                          value={identifier}
                          onChange={(e) => {
                            setIdentifier(e.target.value);
                            if (identifierError) setIdentifierError(null);
                          }}
                          placeholder="Masukkan email atau username"
                          aria-invalid={Boolean(identifierError)}
                          aria-describedby={
                            identifierError ? "identifier-error" : undefined
                          }
                          disabled={isSubmitting}
                          className={`${styles.inputField} ${
                            identifierError ? styles.inputFieldError : ""
                          }`}
                        />
                      </div>
                      {identifierError && (
                        <p
                          id="identifier-error"
                          className={styles.fieldErrorText}
                        >
                          {identifierError}
                        </p>
                      )}
                    </div>

                    <div className={styles.fieldGroup}>
                      <div className={styles.labelRow}>
                        <label
                          htmlFor="login-password"
                          className={styles.fieldLabel}
                        >
                          Kata sandi
                        </label>
                        <button
                          type="button"
                          onClick={openForgotPassword}
                          className={styles.linkUnderlineBtn}
                        >
                          Lupa kata sandi?
                        </button>
                      </div>
                      <div className={styles.inputWrap}>
                        <input
                          id="login-password"
                          name="password"
                          type={showPassword ? "text" : "password"}
                          autoComplete="current-password"
                          value={password}
                          onChange={(e) => {
                            setPassword(e.target.value);
                            if (passwordError) setPasswordError(null);
                          }}
                          placeholder="Masukkan kata sandi"
                          aria-invalid={Boolean(passwordError)}
                          aria-describedby={
                            passwordError ? "password-error" : undefined
                          }
                          disabled={isSubmitting}
                          className={`${styles.inputField} ${styles.inputFieldWithToggle} ${
                            passwordError ? styles.inputFieldError : ""
                          }`}
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
                      {passwordError && (
                        <p
                          id="password-error"
                          className={styles.fieldErrorText}
                        >
                          {passwordError}
                        </p>
                      )}
                    </div>

                    <div className={styles.sessionNote}>
                      <span
                        className={styles.sessionNoteDot}
                        aria-hidden="true"
                      />
                      <span>
                        Sesi Anda dikelola secara aman oleh IntegriOne.
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
                          <span>Memverifikasi...</span>
                        </>
                      ) : isInitializing ? (
                        <span>Menyiapkan sesi...</span>
                      ) : (
                        <span>Masuk</span>
                      )}
                    </button>
                  </form>

                  <div className={styles.registerPrompt}>
                    <span>Belum memiliki akun?</span>{" "}
                    <Link href="/register" className={styles.registerLink}>
                      Daftar akun baru
                    </Link>
                  </div>
                </>
              )}

              {step === "mfa" && (
                <>
                  <div className={styles.backNavRow}>
                    <button
                      type="button"
                      onClick={() => void returnToLogin()}
                      className={styles.backLinkButton}
                    >
                      <Icon name="arrowLeft" size={16} />
                      <span>Kembali ke form masuk</span>
                    </button>
                  </div>

                  <header className={styles.formHeader}>
                    <div className={styles.formHeaderText}>
                      <div className={styles.formEyebrow}>
                        <Icon name="shield" size={14} />
                        <span>VERIFIKASI KEAMANAN</span>
                      </div>
                      <h2 className={styles.formHeading}>Satu langkah lagi.</h2>
                      <p className={styles.formSubtitle}>
                        {mfaMode === "totp"
                          ? "Masukkan kode 6 digit dari aplikasi autentikator Anda."
                          : "Masukkan salah satu kode pemulihan cadangan akun Anda."}
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
                    onSubmit={handleMfaSubmit}
                    className={styles.formStack}
                    noValidate
                  >
                    <div className={styles.fieldGroup}>
                      <label htmlFor="mfa-code" className={styles.fieldLabel}>
                        {mfaMode === "totp"
                          ? "Kode autentikator"
                          : "Kode pemulihan"}
                      </label>
                      <input
                        id="mfa-code"
                        name="mfaCode"
                        type="text"
                        inputMode={mfaMode === "totp" ? "numeric" : "text"}
                        autoComplete="one-time-code"
                        value={mfaCode}
                        onChange={(e) => setMfaCode(e.target.value)}
                        placeholder={
                          mfaMode === "totp" ? "Contoh: 123456" : "Masukkan kode cadangan"
                        }
                        disabled={isSubmitting}
                        className={styles.inputField}
                        autoFocus
                      />
                    </div>

                    <div className={styles.mfaToggleRow}>
                      <button
                        type="button"
                        onClick={() => {
                          setErrorMessage(null);
                          setMfaCode("");
                          setMfaMode((prev) =>
                            prev === "totp" ? "recovery" : "totp",
                          );
                        }}
                        className={styles.linkUnderlineBtn}
                      >
                        {mfaMode === "totp"
                          ? "Gunakan kode pemulihan cadangan"
                          : "Gunakan kode autentikator (TOTP)"}
                      </button>
                    </div>

                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className={styles.primaryButton}
                    >
                      {isSubmitting ? (
                        <>
                          <span className={styles.spinner} aria-hidden="true" />
                          <span>Memverifikasi...</span>
                        </>
                      ) : (
                        <span>Verifikasi &amp; Masuk</span>
                      )}
                    </button>
                  </form>
                </>
              )}

              {step === "reset-request" && (
                <>
                  <div className={styles.backNavRow}>
                    <button
                      type="button"
                      onClick={() => void returnToLogin()}
                      className={styles.backLinkButton}
                    >
                      <Icon name="arrowLeft" size={16} />
                      <span>Kembali ke form masuk</span>
                    </button>
                  </div>

                  <header className={styles.formHeader}>
                    <div className={styles.formHeaderText}>
                      <div className={styles.formEyebrow}>
                        <Icon name="shield" size={14} />
                        <span>PEMULIHAN KATA SANDI</span>
                      </div>
                      <h2 className={styles.formHeading}>
                        Atur ulang kata sandi.
                      </h2>
                      <p className={styles.formSubtitle}>
                        Masukkan alamat email terverifikasi akun Anda. Tautan
                        atau token pemulihan hanya dikirim ke email yang sudah
                        terverifikasi.
                      </p>
                    </div>
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
                    onSubmit={handleResetRequestSubmit}
                    className={styles.formStack}
                    noValidate
                  >
                    <div className={styles.fieldGroup}>
                      <label
                        htmlFor="reset-identifier"
                        className={styles.fieldLabel}
                      >
                        Alamat email terdaftar
                      </label>
                      <input
                        id="reset-identifier"
                        type="email"
                        value={resetIdentifier}
                        onChange={(e) => setResetIdentifier(e.target.value)}
                        placeholder="Masukkan alamat email terverifikasi"
                        disabled={isSubmitting}
                        className={styles.inputField}
                        autoFocus
                      />
                    </div>

                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className={styles.primaryButton}
                    >
                      {isSubmitting ? (
                        <>
                          <span className={styles.spinner} aria-hidden="true" />
                          <span>Mengirimkan permintaan...</span>
                        </>
                      ) : (
                        <span>Kirim Permintaan Pemulihan</span>
                      )}
                    </button>
                  </form>
                </>
              )}

              {step === "reset-confirm" && (
                <>
                  <div className={styles.backNavRow}>
                    <button
                      type="button"
                      onClick={() => void returnToLogin()}
                      className={styles.backLinkButton}
                    >
                      <Icon name="arrowLeft" size={16} />
                      <span>Kembali ke form masuk</span>
                    </button>
                  </div>

                  <header className={styles.formHeader}>
                    <div className={styles.formHeaderText}>
                      <div className={styles.formEyebrow}>
                        <Icon name="check" size={14} />
                        <span>KATA SANDI BARU</span>
                      </div>
                      <h2 className={styles.formHeading}>
                        Masukkan kata sandi baru.
                      </h2>
                      <p className={styles.formSubtitle}>
                        {statusNotice ||
                          "Masukkan token yang diterima serta kata sandi baru Anda."}
                      </p>
                    </div>
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
                    onSubmit={handleResetConfirmSubmit}
                    className={styles.formStack}
                    noValidate
                  >
                    <div className={styles.fieldGroup}>
                      <label htmlFor="reset-token" className={styles.fieldLabel}>
                        Token pemulihan
                      </label>
                      <input
                        id="reset-token"
                        type="text"
                        value={resetToken}
                        onChange={(e) => setResetToken(e.target.value)}
                        placeholder="Tempel token pemulihan dari email"
                        disabled={isSubmitting}
                        className={styles.inputField}
                        autoFocus
                      />
                    </div>

                    <div className={styles.fieldGroup}>
                      <label
                        htmlFor="new-password"
                        className={styles.fieldLabel}
                      >
                        Kata sandi baru
                      </label>
                      <div className={styles.inputWrap}>
                        <input
                          id="new-password"
                          type={showNewPassword ? "text" : "password"}
                          value={newPassword}
                          onChange={(e) => setNewPassword(e.target.value)}
                          placeholder="Minimal 10 karakter"
                          disabled={isSubmitting}
                          className={`${styles.inputField} ${styles.inputFieldWithToggle}`}
                        />
                        <button
                          type="button"
                          onClick={() => setShowNewPassword((prev) => !prev)}
                          aria-label={
                            showNewPassword
                              ? "Sembunyikan kata sandi"
                              : "Tampilkan kata sandi"
                          }
                          className={styles.passwordToggle}
                        >
                          <Icon
                            name={showNewPassword ? "eyeOff" : "eye"}
                            size={18}
                          />
                        </button>
                      </div>

                      {newPassword && (
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

                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className={styles.primaryButton}
                    >
                      {isSubmitting ? (
                        <>
                          <span className={styles.spinner} aria-hidden="true" />
                          <span>Menyimpan kata sandi...</span>
                        </>
                      ) : (
                        <span>Simpan Kata Sandi Baru</span>
                      )}
                    </button>
                  </form>
                </>
              )}

              {step === "reset-done" && (
                <div className={styles.successCard}>
                  <div className={styles.successIconWrap}>
                    <Icon name="check" size={28} />
                  </div>
                  <h2 className={styles.formHeading}>Kata sandi diperbarui.</h2>
                  <p className={styles.formSubtitle}>
                    Kata sandi akun Anda berhasil diperbarui. Silakan masuk
                    kembali menggunakan kredensial baru.
                  </p>
                  <button
                    type="button"
                    onClick={() => void returnToLogin()}
                    className={styles.primaryButton}
                  >
                    Masuk Sekarang
                  </button>
                </div>
              )}

              {step === "navigating" && (
                <div className={styles.successCard}>
                  <div className={styles.spinner} aria-hidden="true" />
                  <h2 className={styles.formHeading}>Autentikasi Berhasil.</h2>
                  <p className={styles.formSubtitle}>
                    Mengarahkan Anda ke ruang kerja IntegriOne...
                  </p>
                </div>
              )}
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
