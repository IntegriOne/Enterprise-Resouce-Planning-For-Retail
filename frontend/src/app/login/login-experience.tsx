"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import gsap from "gsap";
import { useGSAP } from "@gsap/react";
import rawStyles from "./login.module.css";
import { BrandMark } from "./icons";
import {
  CredentialsForm,
  MfaForm,
  RegisterForm,
  RegisterPendingView,
  ResendVerificationForm,
  ResetConfirmForm,
  ResetRequestForm,
  StatusStepView,
  type AuthStep,
  type MfaMode,
  type RegisterFieldErrors,
} from "./auth-forms";
import { LoginFootbar } from "./login-footbar";
import {
  AuthRequestError,
  fetchBrowserSession,
  fetchMyWorkspaceAccessBrowser,
  formatAuthErrorMessage,
  isUnverifiedLoginError,
  isUnverifiedLoginResponse,
  submitCreateUser,
  submitResendValidationEmailToken,
  submitResetPassword,
  submitUserLogin,
  submitUserResetPassword,
  submitValidateUserMfa,
  validateIamPassword,
  validateIamUsername,
} from "@/lib/graphql/iam-browser-auth";
import { validatePostLoginLocation } from "@/lib/auth/redirect-allowlist";

gsap.registerPlugin(useGSAP);

const styles = new Proxy((rawStyles ?? {}) as Record<string, string>, {
  get(target, prop: string) {
    return target[prop] || prop;
  },
});

export default function LoginExperience() {
  const rootRef = useRef<HTMLDivElement>(null);
  const authCardColumnRef = useRef<HTMLDivElement>(null);
  const identifierInputRef = useRef<HTMLInputElement>(null);
  const fullNameInputRef = useRef<HTMLInputElement>(null);
  const submitLockRef = useRef(false);

  const [step, setStep] = useState<AuthStep>("credentials");

  // Credentials state
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [identifierError, setIdentifierError] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);

  // Pendaftaran akun baru state
  const [regFullName, setRegFullName] = useState("");
  const [regEmail, setRegEmail] = useState("");
  const [regUsername, setRegUsername] = useState("");
  const [regPassword, setRegPassword] = useState("");
  const [showRegPassword, setShowRegPassword] = useState(false);
  const [regFieldErrors, setRegFieldErrors] = useState<RegisterFieldErrors>({});
  const [pendingEmail, setPendingEmail] = useState<string | null>(null);
  const [isResending, setIsResending] = useState(false);
  const [resendEmailInput, setResendEmailInput] = useState("");
  const [resendFeedback, setResendFeedback] = useState<string | null>(null);

  // MFA state
  const [mfaMode, setMfaMode] = useState<MfaMode>("totp");
  const [mfaCode, setMfaCode] = useState("");
  const [mfaChallenge, setMfaChallenge] = useState<string | null>(null);

  // Reset password state
  const [resetIdentifier, setResetIdentifier] = useState("");
  const [resetToken, setResetToken] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [showNewPassword, setShowNewPassword] = useState(false);

  // Session & feedback state
  const [csrfToken, setCsrfToken] = useState<string | null>(null);
  const [isInitializing, setIsInitializing] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [statusNotice, setStatusNotice] = useState<string | null>(null);

  useGSAP(
    () => {
      const mm = gsap.matchMedia();

      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.from("[data-hero-stagger]", {
          y: 16,
          opacity: 0,
          duration: 0.65,
          stagger: 0.08,
          ease: "power3.out",
          clearProps: "transform,opacity",
        });
      });

      return () => mm.revert();
    },
    { scope: rootRef },
  );

  useGSAP(
    () => {
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.fromTo(
          "[data-auth-step]",
          { opacity: 0.85, y: 8 },
          {
            opacity: 1,
            y: 0,
            duration: 0.24,
            ease: "power2.out",
            clearProps: "transform,opacity",
          },
        );
      });
      return () => mm.revert();
    },
    { scope: rootRef, dependencies: [step] },
  );

  useEffect(() => {
    const controller = new AbortController();
    let active = true;

    async function bootstrapSession() {
      try {
        const session = await fetchBrowserSession(controller.signal);
        if (!active) return;
        setCsrfToken(session.csrfToken);

        if (session.authenticated) {
          try {
            const access = await fetchMyWorkspaceAccessBrowser(
              controller.signal,
            );
            if (!active) return;
            const safeTarget =
              validatePostLoginLocation(access.defaultLocation) || "/";
            setStep("navigating");
            window.location.replace(safeTarget);
            return;
          } catch (accessErr) {
            if (!active) return;
            if (
              !(
                accessErr instanceof AuthRequestError &&
                accessErr.code === "UNAUTHENTICATED"
              )
            ) {
              setStep("navigating");
              window.location.replace("/");
              return;
            }
          }
        } else {
          const searchParams = new URLSearchParams(window.location.search);
          if (searchParams.get("mode") === "register") {
            setStep("register");
          }
        }
      } catch (err) {
        if (!active || (err instanceof Error && err.name === "AbortError")) {
          return;
        }
        const searchParams = new URLSearchParams(window.location.search);
        if (searchParams.get("mode") === "register") {
          setStep("register");
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
      setCsrfToken(session.csrfToken);
      return session.csrfToken;
    } catch {
      return null;
    }
  }

  async function refreshCsrfSilently() {
    try {
      const refreshed = await fetchBrowserSession();
      setCsrfToken(refreshed.csrfToken);
    } catch {
      // Abaikan bila server belum merespons
    }
  }

  function clearFeedback() {
    setErrorMessage(null);
    setStatusNotice(null);
    setIdentifierError(null);
    setPasswordError(null);
    setRegFieldErrors({});
    setResendFeedback(null);
  }

  function switchToStep(nextStep: AuthStep) {
    clearFeedback();
    setStep(nextStep);
  }

  function scrollToAuthFocus(mode: "credentials" | "register") {
    switchToStep(mode);
    authCardColumnRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "center",
    });
    window.setTimeout(() => {
      if (mode === "register") {
        fullNameInputRef.current?.focus();
      } else {
        identifierInputRef.current?.focus();
      }
    }, 220);
  }

  function handleFinalNavigation(
    locationHeader: string | null,
    redirectLocationPayload?: string | null,
  ) {
    const candidate = locationHeader || redirectLocationPayload || "/";
    const validatedDestination = validatePostLoginLocation(candidate);

    if (validatedDestination) {
      setStep("navigating");
      window.location.replace(validatedDestination);
    } else {
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
    setStatusNotice(null);
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

      setPassword("");

      if (payload.status === "MFA_REQUIRED" && payload.mfaChallenge) {
        setMfaChallenge(payload.mfaChallenge);
        setMfaCode("");
        setStep("mfa");
        return;
      }

      if (isUnverifiedLoginResponse(payload, locationHeader)) {
        setStep("navigating");
        window.location.replace("/verify-email");
        return;
      }

      if (payload.status === "AUTHENTICATED") {
        handleFinalNavigation(locationHeader, payload.redirectLocation);
        return;
      }

      setErrorMessage("Respons autentikasi tidak dikenal. Silakan coba lagi.");
    } catch (err) {
      if (isUnverifiedLoginError(err)) {
        setStep("navigating");
        window.location.replace("/verify-email");
        return;
      }
      setErrorMessage(formatAuthErrorMessage(err, "login"));
      await refreshCsrfSilently();
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  async function handleRegisterSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting) return;

    const trimmedName = regFullName.trim();
    const trimmedEmail = regEmail.trim();
    const trimmedUsername = regUsername.trim().toLowerCase();
    setRegFullName(trimmedName);
    setRegEmail(trimmedEmail);
    setRegUsername(trimmedUsername);
    setErrorMessage(null);
    setStatusNotice(null);

    const errors: RegisterFieldErrors = {};
    if (!trimmedName) {
      errors.fullName = "Masukkan nama lengkap Anda.";
    }
    if (!trimmedEmail) {
      errors.email = "Masukkan alamat email Anda.";
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      errors.email = "Format alamat email tidak valid.";
    }
    const usernameErr = validateIamUsername(trimmedUsername);
    if (usernameErr) {
      errors.username = usernameErr;
    }
    const passwordErr = validateIamPassword(regPassword, [
      trimmedName,
      trimmedEmail,
      trimmedUsername,
    ]);
    if (passwordErr) {
      errors.password = passwordErr;
    }

    if (Object.keys(errors).length > 0) {
      setRegFieldErrors(errors);
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
        regPassword,
        activeCsrf,
      );

      setRegPassword("");

      const redirectTarget =
        locationHeader?.trim() || payload.redirectLocation?.trim();
      if (redirectTarget === "/login") {
        setStep("credentials");
        setStatusNotice(
          "Permintaan pendaftaran diterima. Jika email sudah terdaftar sebelumnya, silakan langsung masuk atau atur ulang kata sandi.",
        );
        return;
      }

      setPendingEmail(trimmedEmail);
      setStep("register-pending");
    } catch (err) {
      const formatted = formatAuthErrorMessage(err, "register");
      const lower = formatted.toLowerCase();
      if (lower.includes("username")) {
        setRegFieldErrors((prev) => ({
          ...prev,
          username: formatted.includes("taken")
            ? "Username tersebut sudah digunakan. Silakan pilih username lain."
            : formatted,
        }));
      }
      setErrorMessage(formatted);
      await refreshCsrfSilently();
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  async function handleResendForPendingRegistration() {
    if (!pendingEmail || isResending) return;
    setIsResending(true);
    setResendFeedback(null);
    try {
      const activeCsrf = await ensureCsrfToken();
      await submitResendValidationEmailToken(pendingEmail, activeCsrf);
      setResendFeedback(
        "Tautan verifikasi baru telah dikirim. Tautan sebelumnya otomatis kedaluwarsa.",
      );
    } catch (err) {
      setResendFeedback(formatAuthErrorMessage(err, "verify"));
    } finally {
      setIsResending(false);
    }
  }

  async function handleResendVerificationSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (submitLockRef.current || isSubmitting) return;

    const trimmedEmail = resendEmailInput.trim();
    setResendEmailInput(trimmedEmail);
    setErrorMessage(null);
    setStatusNotice(null);

    if (!trimmedEmail || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      setErrorMessage("Masukkan alamat email yang valid.");
      return;
    }

    submitLockRef.current = true;
    setIsSubmitting(true);

    try {
      const activeCsrf = await ensureCsrfToken();
      await submitResendValidationEmailToken(trimmedEmail, activeCsrf);
      setPendingEmail(trimmedEmail);
      setResendFeedback(
        "Jika alamat email terdaftar dan belum terverifikasi, tautan verifikasi baru telah dikirim.",
      );
      setStep("register-pending");
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "verify"));
      await refreshCsrfSilently();
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
      await refreshCsrfSilently();
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
      await submitUserResetPassword(trimmed, activeCsrf);
      setStatusNotice(
        "Jika email terdaftar dan sudah terverifikasi, instruksi pengaturan ulang kata sandi telah dikirim.",
      );
      setResetToken("");
      setNewPassword("");
      setStep("reset-confirm");
    } catch (err) {
      setErrorMessage(formatAuthErrorMessage(err, "reset-request"));
      await refreshCsrfSilently();
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

    if (!trimmedToken) {
      setErrorMessage("Masukkan token pemulihan dari email Anda.");
      return;
    }
    const newPasswordErr = validateIamPassword(newPassword, [resetIdentifier]);
    if (newPasswordErr) {
      setErrorMessage(newPasswordErr);
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
      await refreshCsrfSilently();
    } finally {
      submitLockRef.current = false;
      setIsSubmitting(false);
    }
  }

  function openForgotPassword() {
    const trimmed = identifier.trim();
    setResetIdentifier(trimmed.includes("@") ? trimmed : "");
    switchToStep("reset-request");
  }

  async function returnToLogin() {
    setMfaChallenge(null);
    setMfaCode("");
    switchToStep("credentials");
    await refreshCsrfSilently();
  }

  const showPrimaryTabs = step === "credentials" || step === "register";

  return (
    <div ref={rootRef} className={styles.enterpriseShell}>
      {/* Sticky Translucent Navigation Header */}
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

          <nav aria-label="Navigasi utama" className={styles.navRight}>
            <Link href="/" className={styles.navTextLink}>
              Beranda
            </Link>
            <button
              type="button"
              onClick={() =>
                scrollToAuthFocus(
                  step === "register" ? "credentials" : "register",
                )
              }
              className={styles.navActionPill}
            >
              {step === "register" ? "Masuk ke Akun" : "Daftar Akun"}
            </button>
          </nav>
        </div>
      </header>

      <main className="flex-1 overflow-x-hidden w-full max-w-full flex items-center">
        {/* Bagian 2: Login Section */}
        <section className={`${styles.heroSection} w-full`}>
          <div className={styles.sectionInner}>
            <div className={styles.heroGrid}>
              <aside
                className={styles.heroEditorialColumn}
                aria-label="Pengantar IntegriOne"
              >
                <div className={styles.heroGridLines} aria-hidden="true" />

                <div className={styles.heroTopMeta} data-hero-stagger>
                  <span className={styles.coralRule} aria-hidden="true" />
                  <span className={styles.heroBrandKicker}>
                    INTEGRIONE RETAIL HOUSE
                  </span>
                </div>

                <div className={`${styles.heroCopyBlock} max-w-5xl`} data-hero-stagger>
                  <h1 className={styles.heroDisplayTitle}>
                    Masuk ke akun{" "}
                    <span
                      className={styles.heroInlinePill}
                      style={{
                        backgroundImage:
                          "url('https://picsum.photos/seed/integrione-atelier-pill/640/360')",
                      }}
                      aria-hidden="true"
                    />
                    <span className={styles.heroItalicWord}>IntegriOne</span>{" "}
                    Anda.
                  </h1>
                  <p className={styles.heroLead}>
                    Masuk menggunakan email atau username terdaftar untuk
                    mengakses akun Anda, atau daftarkan akun baru dalam beberapa
                    langkah singkat.
                  </p>
                </div>

                <div className={styles.heroVisualFloatWrap} data-hero-stagger>
                  <div
                    className={`${styles.heroArtworkFrame} group`}
                    aria-hidden="true"
                  >
                    <div
                      className={`${styles.heroArtworkMedia} group-hover:scale-105 transition-transform duration-700 ease-out`}
                      style={{
                        backgroundImage:
                          "url('https://picsum.photos/seed/integrione-atelier-hero/1200/800')",
                      }}
                    />
                  </div>
                  <div className={styles.heroArtworkCaption}>
                    <p className={styles.heroCaptionTitle}>
                      Keamanan Akun Terjaga
                    </p>
                    <p className={styles.heroCaptionDesc}>
                      Pemulihan kata sandi, verifikasi email, dan autentikasi
                      dua langkah tersedia langsung di dalam halaman ini.
                    </p>
                  </div>
                </div>
              </aside>

              <div ref={authCardColumnRef} className={styles.authCardColumn}>
                <div className={styles.authCardSurface}>
                  <div data-auth-step className={styles.authContainer}>
                    {showPrimaryTabs && (
                      <div
                        className={styles.authModeSwitch}
                        role="tablist"
                        aria-label="Mode formulir akun"
                      >
                        <button
                          type="button"
                          role="tab"
                          aria-selected={step === "credentials"}
                          onClick={() => switchToStep("credentials")}
                          className={`${styles.authModeTab} ${
                            step === "credentials" ? styles.authModeTabActive : ""
                          }`}
                        >
                          Masuk
                        </button>
                        <button
                          type="button"
                          role="tab"
                          aria-selected={step === "register"}
                          onClick={() => switchToStep("register")}
                          className={`${styles.authModeTab} ${
                            step === "register" ? styles.authModeTabActive : ""
                          }`}
                        >
                          Daftar Akun
                        </button>
                      </div>
                    )}

                    {step === "credentials" && (
                      <CredentialsForm
                        styles={styles}
                        identifierInputRef={identifierInputRef}
                        identifier={identifier}
                        password={password}
                        showPassword={showPassword}
                        identifierError={identifierError}
                        passwordError={passwordError}
                        errorMessage={errorMessage}
                        statusNotice={statusNotice}
                        isSubmitting={isSubmitting}
                        isInitializing={isInitializing}
                        onIdentifierChange={(val) => {
                          setIdentifier(val);
                          if (identifierError) setIdentifierError(null);
                        }}
                        onPasswordChange={(val) => {
                          setPassword(val);
                          if (passwordError) setPasswordError(null);
                        }}
                        onTogglePassword={() => setShowPassword((prev) => !prev)}
                        onSubmit={handleLoginSubmit}
                        onOpenForgotPassword={openForgotPassword}
                        onOpenRegister={() => switchToStep("register")}
                        onOpenResendVerification={() => {
                          const trimmed = identifier.trim();
                          setResendEmailInput(trimmed.includes("@") ? trimmed : "");
                          switchToStep("resend-verification");
                        }}
                      />
                    )}

                    {step === "register" && (
                      <RegisterForm
                        styles={styles}
                        fullNameInputRef={fullNameInputRef}
                        fullName={regFullName}
                        email={regEmail}
                        username={regUsername}
                        password={regPassword}
                        showPassword={showRegPassword}
                        fieldErrors={regFieldErrors}
                        errorMessage={errorMessage}
                        statusNotice={statusNotice}
                        isSubmitting={isSubmitting}
                        isInitializing={isInitializing}
                        onFullNameChange={(val) => {
                          setRegFullName(val);
                          if (regFieldErrors.fullName) {
                            setRegFieldErrors((prev) => ({
                              ...prev,
                              fullName: undefined,
                            }));
                          }
                        }}
                        onEmailChange={(val) => {
                          setRegEmail(val);
                          if (regFieldErrors.email) {
                            setRegFieldErrors((prev) => ({
                              ...prev,
                              email: undefined,
                            }));
                          }
                        }}
                        onUsernameChange={(val) => {
                          setRegUsername(val);
                          if (regFieldErrors.username) {
                            setRegFieldErrors((prev) => ({
                              ...prev,
                              username: undefined,
                            }));
                          }
                        }}
                        onPasswordChange={(val) => {
                          setRegPassword(val);
                          if (regFieldErrors.password) {
                            setRegFieldErrors((prev) => ({
                              ...prev,
                              password: undefined,
                            }));
                          }
                        }}
                        onTogglePassword={() =>
                          setShowRegPassword((prev) => !prev)
                        }
                        onSubmit={handleRegisterSubmit}
                        onOpenLogin={() => switchToStep("credentials")}
                      />
                    )}

                    {step === "register-pending" && pendingEmail && (
                      <RegisterPendingView
                        styles={styles}
                        registeredEmail={pendingEmail}
                        resendMessage={resendFeedback}
                        isResending={isResending}
                        onResend={() => void handleResendForPendingRegistration()}
                        onBackToLogin={() => void returnToLogin()}
                      />
                    )}

                    {step === "mfa" && (
                      <MfaForm
                        styles={styles}
                        mfaMode={mfaMode}
                        mfaCode={mfaCode}
                        errorMessage={errorMessage}
                        isSubmitting={isSubmitting}
                        onCodeChange={setMfaCode}
                        onToggleMode={() => {
                          setErrorMessage(null);
                          setMfaCode("");
                          setMfaMode((prev) =>
                            prev === "totp" ? "recovery" : "totp",
                          );
                        }}
                        onSubmit={handleMfaSubmit}
                        onBackToLogin={() => void returnToLogin()}
                      />
                    )}

                    {step === "reset-request" && (
                      <ResetRequestForm
                        styles={styles}
                        resetIdentifier={resetIdentifier}
                        errorMessage={errorMessage}
                        isSubmitting={isSubmitting}
                        onIdentifierChange={setResetIdentifier}
                        onSubmit={handleResetRequestSubmit}
                        onBackToLogin={() => void returnToLogin()}
                      />
                    )}

                    {step === "reset-confirm" && (
                      <ResetConfirmForm
                        styles={styles}
                        resetToken={resetToken}
                        newPassword={newPassword}
                        showNewPassword={showNewPassword}
                        statusNotice={statusNotice}
                        errorMessage={errorMessage}
                        isSubmitting={isSubmitting}
                        onTokenChange={setResetToken}
                        onNewPasswordChange={setNewPassword}
                        onToggleNewPassword={() =>
                          setShowNewPassword((prev) => !prev)
                        }
                        onSubmit={handleResetConfirmSubmit}
                        onBackToLogin={() => void returnToLogin()}
                      />
                    )}

                    {step === "resend-verification" && (
                      <ResendVerificationForm
                        styles={styles}
                        email={resendEmailInput}
                        errorMessage={errorMessage}
                        statusNotice={statusNotice}
                        isSubmitting={isSubmitting}
                        onEmailChange={setResendEmailInput}
                        onSubmit={handleResendVerificationSubmit}
                        onBackToLogin={() => void returnToLogin()}
                      />
                    )}

                    {(step === "reset-done" || step === "navigating") && (
                      <StatusStepView
                        styles={styles}
                        step={step}
                        onBackToLogin={() => void returnToLogin()}
                      />
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>

      {/* Bagian 3: Footer Editorial */}
      <LoginFootbar
        styles={styles}
        isRegisterStep={step === "register"}
        onToggleAuthView={() =>
          scrollToAuthFocus(step === "register" ? "credentials" : "register")
        }
        onOpenResetPassword={() => {
          openForgotPassword();
          authCardColumnRef.current?.scrollIntoView({
            behavior: "smooth",
            block: "center",
          });
        }}
      />
    </div>
  );
}
