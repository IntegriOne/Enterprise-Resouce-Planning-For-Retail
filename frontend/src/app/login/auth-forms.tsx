"use client";

import type { FormEvent, RefObject } from "react";
import { Icon } from "./icons";
import {
  getPasswordChecklist,
  getPasswordStrength,
} from "@/lib/graphql/iam-browser-auth";

export type AuthStep =
  | "credentials"
  | "register"
  | "register-pending"
  | "mfa"
  | "reset-request"
  | "reset-confirm"
  | "reset-done"
  | "resend-verification"
  | "navigating";

export type MfaMode = "totp" | "recovery";

export type RegisterFieldErrors = {
  fullName?: string;
  email?: string;
  username?: string;
  password?: string;
};

type StylesMap = Record<string, string>;

function AlertBanner({
  styles,
  errorMessage,
  statusNotice,
}: {
  styles: StylesMap;
  errorMessage: string | null;
  statusNotice: string | null;
}) {
  if (errorMessage) {
    return (
      <div
        role="alert"
        aria-live="assertive"
        className={`${styles.alertBox} ${styles.alertError}`}
      >
        <span className={styles.alertIcon}>
          <Icon name="alert" size={18} />
        </span>
        <span>{errorMessage}</span>
      </div>
    );
  }
  if (statusNotice) {
    return (
      <div
        role="status"
        aria-live="polite"
        className={`${styles.alertBox} ${styles.alertWarning}`}
      >
        <span className={styles.alertIcon}>
          <Icon name="alert" size={18} />
        </span>
        <span>{statusNotice}</span>
      </div>
    );
  }
  return null;
}

function PasswordChecklistView({
  styles,
  password,
  forbiddenTerms,
}: {
  styles: StylesMap;
  password: string;
  forbiddenTerms: string[];
}) {
  const checklist = getPasswordChecklist(password, forbiddenTerms);
  const strength = getPasswordStrength(password);

  if (!password) return null;

  const items = [
    { ok: checklist.minLength && checklist.maxLength, text: "12–128 karakter" },
    { ok: checklist.hasUpperAndLower, text: "Huruf besar & huruf kecil" },
    { ok: checklist.hasDigit && checklist.hasSymbol, text: "Angka & simbol" },
    { ok: checklist.noIdentityTerms, text: "Tanpa bagian nama/email/username" },
  ];

  return (
    <div className={styles.strengthContainer}>
      <div className={styles.strengthHeaderRow}>
        <span
          className={styles.strengthLabel}
          style={{ color: strength.color }}
        >
          Kekuatan: {strength.label}
        </span>
      </div>
      <div className={styles.strengthBarTrack}>
        <div
          className={styles.strengthBarFill}
          style={{
            width: `${strength.score}%`,
            backgroundColor: strength.color,
          }}
        />
      </div>
      <ul className={styles.policyChecklist} aria-label="Syarat kata sandi">
        {items.map((item) => (
          <li
            key={item.text}
            className={`${styles.policyCheckItem} ${
              item.ok ? styles.policyCheckOk : ""
            }`}
          >
            <Icon name="check" size={13} />
            <span>{item.text}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function CredentialsForm({
  styles,
  identifierInputRef,
  identifier,
  password,
  showPassword,
  identifierError,
  passwordError,
  errorMessage,
  statusNotice,
  isSubmitting,
  isInitializing,
  onIdentifierChange,
  onPasswordChange,
  onTogglePassword,
  onSubmit,
  onOpenForgotPassword,
  onOpenRegister,
  onOpenResendVerification,
}: {
  styles: StylesMap;
  identifierInputRef: RefObject<HTMLInputElement | null>;
  identifier: string;
  password: string;
  showPassword: boolean;
  identifierError: string | null;
  passwordError: string | null;
  errorMessage: string | null;
  statusNotice: string | null;
  isSubmitting: boolean;
  isInitializing: boolean;
  onIdentifierChange: (val: string) => void;
  onPasswordChange: (val: string) => void;
  onTogglePassword: () => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onOpenForgotPassword: () => void;
  onOpenRegister: () => void;
  onOpenResendVerification: () => void;
}) {
  return (
    <>
      <header className={styles.formHeader}>
        <div className={styles.formHeaderText}>
          <div className={styles.formEyebrow}>
            <span className={styles.formEyebrowAccent} aria-hidden="true" />
            <span>MASUK KE AKUN</span>
          </div>
          <h2 className={styles.formHeading}>Selamat datang kembali.</h2>
          <p className={styles.formSubtitle}>
            Masuk menggunakan email atau username akun IntegriOne Anda.
          </p>
        </div>
      </header>

      {isInitializing && (
        <div
          className={styles.skeletonBanner}
          aria-live="polite"
          aria-label="Menyiapkan sesi"
        >
          <span className={styles.skeletonBar} />
          <span className={styles.skeletonBarShort} />
        </div>
      )}

      <AlertBanner
        styles={styles}
        errorMessage={errorMessage}
        statusNotice={statusNotice}
      />

      <form onSubmit={onSubmit} className={styles.formStack} noValidate>
        <div className={styles.fieldGroup}>
          <label htmlFor="login-identifier" className={styles.fieldLabel}>
            Email atau username
          </label>
          <div className={styles.inputWrap}>
            <input
              ref={identifierInputRef}
              id="login-identifier"
              name="identifier"
              type="text"
              autoComplete="username"
              autoCapitalize="none"
              autoCorrect="off"
              spellCheck={false}
              value={identifier}
              onChange={(e) => onIdentifierChange(e.target.value)}
              placeholder="nama@domain.com atau username"
              aria-invalid={Boolean(identifierError)}
              aria-describedby={identifierError ? "identifier-error" : undefined}
              disabled={isSubmitting}
              className={`${styles.inputField} ${
                identifierError ? styles.inputFieldError : ""
              }`}
            />
          </div>
          {identifierError && (
            <p id="identifier-error" className={styles.fieldErrorText}>
              {identifierError}
            </p>
          )}
        </div>

        <div className={styles.fieldGroup}>
          <div className={styles.labelRow}>
            <label htmlFor="login-password" className={styles.fieldLabel}>
              Kata sandi
            </label>
            <button
              type="button"
              onClick={onOpenForgotPassword}
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
              onChange={(e) => onPasswordChange(e.target.value)}
              placeholder="Masukkan kata sandi"
              aria-invalid={Boolean(passwordError)}
              aria-describedby={passwordError ? "password-error" : undefined}
              disabled={isSubmitting}
              className={`${styles.inputField} ${styles.inputFieldWithToggle} ${
                passwordError ? styles.inputFieldError : ""
              }`}
            />
            <button
              type="button"
              onClick={onTogglePassword}
              aria-label={
                showPassword ? "Sembunyikan kata sandi" : "Tampilkan kata sandi"
              }
              aria-pressed={showPassword}
              disabled={isSubmitting}
              className={styles.passwordToggle}
            >
              <Icon name={showPassword ? "eyeOff" : "eye"} size={18} />
            </button>
          </div>
          {passwordError && (
            <p id="password-error" className={styles.fieldErrorText}>
              {passwordError}
            </p>
          )}
        </div>

        <div className={styles.sessionNote}>
          <span className={styles.sessionNoteDot} aria-hidden="true" />
          <span>
            Sesi browser dilindungi enkripsi dan validasi keamanan otomatis.
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
            <>
              <span>Masuk ke Akun</span>
              <Icon name="arrowRight" size={17} />
            </>
          )}
        </button>
      </form>

      <div className={styles.authAuxiliaryRow}>
        <button
          type="button"
          onClick={onOpenResendVerification}
          className={styles.linkUnderlineBtn}
        >
          Kirim ulang email verifikasi
        </button>
      </div>

      <div className={styles.registerPrompt}>
        <span>Belum memiliki akun?</span>{" "}
        <button
          type="button"
          onClick={onOpenRegister}
          className={styles.registerInlineAction}
        >
          Daftar akun baru
        </button>
      </div>
    </>
  );
}

export function RegisterForm({
  styles,
  fullNameInputRef,
  fullName,
  email,
  username,
  password,
  showPassword,
  fieldErrors,
  errorMessage,
  statusNotice,
  isSubmitting,
  isInitializing,
  onFullNameChange,
  onEmailChange,
  onUsernameChange,
  onPasswordChange,
  onTogglePassword,
  onSubmit,
  onOpenLogin,
}: {
  styles: StylesMap;
  fullNameInputRef: RefObject<HTMLInputElement | null>;
  fullName: string;
  email: string;
  username: string;
  password: string;
  showPassword: boolean;
  fieldErrors: RegisterFieldErrors;
  errorMessage: string | null;
  statusNotice: string | null;
  isSubmitting: boolean;
  isInitializing: boolean;
  onFullNameChange: (val: string) => void;
  onEmailChange: (val: string) => void;
  onUsernameChange: (val: string) => void;
  onPasswordChange: (val: string) => void;
  onTogglePassword: () => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onOpenLogin: () => void;
}) {
  return (
    <>
      <header className={styles.formHeader}>
        <div className={styles.formHeaderText}>
          <div className={styles.formEyebrow}>
            <span className={styles.formEyebrowAccent} aria-hidden="true" />
            <span>PENDAFTARAN AKUN</span>
          </div>
          <h2 className={styles.formHeading}>Buat akun baru.</h2>
          <p className={styles.formSubtitle}>
            Daftar akun IntegriOne untuk pengalaman belanja dan pelacakan
            pesanan terhubung.
          </p>
        </div>
      </header>

      <AlertBanner
        styles={styles}
        errorMessage={errorMessage}
        statusNotice={statusNotice}
      />

      <form onSubmit={onSubmit} className={styles.formStack} noValidate>
        <div className={styles.fieldGroup}>
          <label htmlFor="reg-full-name" className={styles.fieldLabel}>
            Nama lengkap
          </label>
          <input
            ref={fullNameInputRef}
            id="reg-full-name"
            name="fullName"
            type="text"
            autoComplete="name"
            value={fullName}
            onChange={(e) => onFullNameChange(e.target.value)}
            placeholder="Contoh: Budi Santoso"
            aria-invalid={Boolean(fieldErrors.fullName)}
            aria-describedby={
              fieldErrors.fullName ? "reg-fullname-error" : undefined
            }
            disabled={isSubmitting}
            className={`${styles.inputField} ${
              fieldErrors.fullName ? styles.inputFieldError : ""
            }`}
          />
          {fieldErrors.fullName && (
            <p id="reg-fullname-error" className={styles.fieldErrorText}>
              {fieldErrors.fullName}
            </p>
          )}
        </div>

        <div className={styles.fieldGridTwo}>
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
              onChange={(e) => onEmailChange(e.target.value)}
              placeholder="nama@domain.com"
              aria-invalid={Boolean(fieldErrors.email)}
              aria-describedby={fieldErrors.email ? "reg-email-error" : undefined}
              disabled={isSubmitting}
              className={`${styles.inputField} ${
                fieldErrors.email ? styles.inputFieldError : ""
              }`}
            />
            {fieldErrors.email && (
              <p id="reg-email-error" className={styles.fieldErrorText}>
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
              autoCapitalize="none"
              autoCorrect="off"
              spellCheck={false}
              value={username}
              onChange={(e) => onUsernameChange(e.target.value)}
              placeholder="budi.santoso"
              aria-invalid={Boolean(fieldErrors.username)}
              aria-describedby={
                fieldErrors.username ? "reg-username-error" : undefined
              }
              disabled={isSubmitting}
              className={`${styles.inputField} ${
                fieldErrors.username ? styles.inputFieldError : ""
              }`}
            />
            {fieldErrors.username && (
              <p id="reg-username-error" className={styles.fieldErrorText}>
                {fieldErrors.username}
              </p>
            )}
          </div>
        </div>

        <div className={styles.fieldGroup}>
          <label htmlFor="reg-password" className={styles.fieldLabel}>
            Kata sandi
          </label>
          <div className={styles.inputWrap}>
            <input
              id="reg-password"
              name="password"
              type={showPassword ? "text" : "password"}
              autoComplete="new-password"
              value={password}
              onChange={(e) => onPasswordChange(e.target.value)}
              placeholder="Min. 12 karakter (Aa1!)"
              aria-invalid={Boolean(fieldErrors.password)}
              aria-describedby={
                fieldErrors.password ? "reg-password-error" : "reg-password-hint"
              }
              disabled={isSubmitting}
              className={`${styles.inputField} ${styles.inputFieldWithToggle} ${
                fieldErrors.password ? styles.inputFieldError : ""
              }`}
            />
            <button
              type="button"
              onClick={onTogglePassword}
              aria-label={
                showPassword ? "Sembunyikan kata sandi" : "Tampilkan kata sandi"
              }
              aria-pressed={showPassword}
              disabled={isSubmitting}
              className={styles.passwordToggle}
            >
              <Icon name={showPassword ? "eyeOff" : "eye"} size={18} />
            </button>
          </div>
          {fieldErrors.password ? (
            <p id="reg-password-error" className={styles.fieldErrorText}>
              {fieldErrors.password}
            </p>
          ) : (
            <p id="reg-password-hint" className={styles.fieldHintText}>
              Minimal 12 karakter: huruf besar, huruf kecil, angka, dan simbol.
            </p>
          )}
          <PasswordChecklistView
            styles={styles}
            password={password}
            forbiddenTerms={[fullName, email, username]}
          />
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
            <>
              <span>Daftar Akun</span>
              <Icon name="arrowRight" size={17} />
            </>
          )}
        </button>
      </form>

      <div className={styles.registerPrompt}>
        <span>Sudah memiliki akun?</span>{" "}
        <button
          type="button"
          onClick={onOpenLogin}
          className={styles.registerInlineAction}
        >
          Masuk sekarang
        </button>
      </div>
    </>
  );
}

export function RegisterPendingView({
  styles,
  registeredEmail,
  resendMessage,
  isResending,
  onResend,
  onBackToLogin,
}: {
  styles: StylesMap;
  registeredEmail: string;
  resendMessage: string | null;
  isResending: boolean;
  onResend: () => void;
  onBackToLogin: () => void;
}) {
  return (
    <div className={styles.successCard}>
      <div className={styles.successIconWrap}>
        <Icon name="mail" size={26} />
      </div>
      <h2 className={styles.formHeading}>Periksa email Anda.</h2>
      <p className={styles.formSubtitle}>
        Akun terdaftar atas <strong>{registeredEmail}</strong> dalam status
        menunggu aktivasi. Konfirmasikan tautan verifikasi sebelum masuk.
      </p>

      {resendMessage && (
        <div
          role="status"
          aria-live="polite"
          className={`${styles.alertBox} ${styles.alertSuccess}`}
        >
          <span className={styles.alertIcon}>
            <Icon name="check" size={18} />
          </span>
          <span>{resendMessage}</span>
        </div>
      )}

      <div className={styles.actionStack}>
        <button
          type="button"
          onClick={onBackToLogin}
          className={styles.primaryButton}
        >
          <span>Kembali ke Halaman Masuk</span>
          <Icon name="arrowRight" size={17} />
        </button>

        <button
          type="button"
          onClick={onResend}
          disabled={isResending}
          className={styles.secondaryButton}
        >
          {isResending ? "Mengirim ulang..." : "Kirim Ulang Email Verifikasi"}
        </button>
      </div>
    </div>
  );
}

export function MfaForm({
  styles,
  mfaMode,
  mfaCode,
  errorMessage,
  isSubmitting,
  onCodeChange,
  onToggleMode,
  onSubmit,
  onBackToLogin,
}: {
  styles: StylesMap;
  mfaMode: MfaMode;
  mfaCode: string;
  errorMessage: string | null;
  isSubmitting: boolean;
  onCodeChange: (val: string) => void;
  onToggleMode: () => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onBackToLogin: () => void;
}) {
  return (
    <>
      <div className={styles.backNavRow}>
        <button
          type="button"
          onClick={onBackToLogin}
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
            <span>VERIFIKASI DUA LANGKAH</span>
          </div>
          <h2 className={styles.formHeading}>Satu langkah lagi.</h2>
          <p className={styles.formSubtitle}>
            {mfaMode === "totp"
              ? "Masukkan kode 6 digit dari aplikasi autentikator Anda."
              : "Masukkan salah satu kode pemulihan cadangan akun Anda."}
          </p>
        </div>
      </header>

      <AlertBanner styles={styles} errorMessage={errorMessage} statusNotice={null} />

      <form onSubmit={onSubmit} className={styles.formStack} noValidate>
        <div className={styles.fieldGroup}>
          <label htmlFor="mfa-code" className={styles.fieldLabel}>
            {mfaMode === "totp" ? "Kode autentikator" : "Kode pemulihan"}
          </label>
          <input
            id="mfa-code"
            name="mfaCode"
            type="text"
            inputMode={mfaMode === "totp" ? "numeric" : "text"}
            autoComplete="one-time-code"
            value={mfaCode}
            onChange={(e) => onCodeChange(e.target.value)}
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
            onClick={onToggleMode}
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
  );
}

export function ResetRequestForm({
  styles,
  resetIdentifier,
  errorMessage,
  isSubmitting,
  onIdentifierChange,
  onSubmit,
  onBackToLogin,
}: {
  styles: StylesMap;
  resetIdentifier: string;
  errorMessage: string | null;
  isSubmitting: boolean;
  onIdentifierChange: (val: string) => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onBackToLogin: () => void;
}) {
  return (
    <>
      <div className={styles.backNavRow}>
        <button
          type="button"
          onClick={onBackToLogin}
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
          <h2 className={styles.formHeading}>Atur ulang kata sandi.</h2>
          <p className={styles.formSubtitle}>
            Masukkan email terverifikasi Anda untuk menerima token pemulihan.
          </p>
        </div>
      </header>

      <AlertBanner styles={styles} errorMessage={errorMessage} statusNotice={null} />

      <form onSubmit={onSubmit} className={styles.formStack} noValidate>
        <div className={styles.fieldGroup}>
          <label htmlFor="reset-identifier" className={styles.fieldLabel}>
            Alamat email terdaftar
          </label>
          <input
            id="reset-identifier"
            type="email"
            value={resetIdentifier}
            onChange={(e) => onIdentifierChange(e.target.value)}
            placeholder="nama@domain.com"
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
  );
}

export function ResetConfirmForm({
  styles,
  resetToken,
  newPassword,
  showNewPassword,
  statusNotice,
  errorMessage,
  isSubmitting,
  onTokenChange,
  onNewPasswordChange,
  onToggleNewPassword,
  onSubmit,
  onBackToLogin,
}: {
  styles: StylesMap;
  resetToken: string;
  newPassword: string;
  showNewPassword: boolean;
  statusNotice: string | null;
  errorMessage: string | null;
  isSubmitting: boolean;
  onTokenChange: (val: string) => void;
  onNewPasswordChange: (val: string) => void;
  onToggleNewPassword: () => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onBackToLogin: () => void;
}) {
  return (
    <>
      <div className={styles.backNavRow}>
        <button
          type="button"
          onClick={onBackToLogin}
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
          <h2 className={styles.formHeading}>Masukkan kata sandi baru.</h2>
          <p className={styles.formSubtitle}>
            {statusNotice ||
              "Masukkan token yang diterima serta kata sandi baru Anda."}
          </p>
        </div>
      </header>

      <AlertBanner styles={styles} errorMessage={errorMessage} statusNotice={null} />

      <form onSubmit={onSubmit} className={styles.formStack} noValidate>
        <div className={styles.fieldGroup}>
          <label htmlFor="reset-token" className={styles.fieldLabel}>
            Token pemulihan
          </label>
          <input
            id="reset-token"
            type="text"
            value={resetToken}
            onChange={(e) => onTokenChange(e.target.value)}
            placeholder="Tempel token pemulihan dari email"
            disabled={isSubmitting}
            className={styles.inputField}
            autoFocus
          />
        </div>

        <div className={styles.fieldGroup}>
          <label htmlFor="new-password" className={styles.fieldLabel}>
            Kata sandi baru
          </label>
          <div className={styles.inputWrap}>
            <input
              id="new-password"
              type={showNewPassword ? "text" : "password"}
              autoComplete="new-password"
              value={newPassword}
              onChange={(e) => onNewPasswordChange(e.target.value)}
              placeholder="Min. 12 karakter (Aa1!)"
              disabled={isSubmitting}
              className={`${styles.inputField} ${styles.inputFieldWithToggle}`}
            />
            <button
              type="button"
              onClick={onToggleNewPassword}
              aria-label={
                showNewPassword
                  ? "Sembunyikan kata sandi"
                  : "Tampilkan kata sandi"
              }
              className={styles.passwordToggle}
            >
              <Icon name={showNewPassword ? "eyeOff" : "eye"} size={18} />
            </button>
          </div>
          <PasswordChecklistView
            styles={styles}
            password={newPassword}
            forbiddenTerms={[]}
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
              <span>Menyimpan kata sandi...</span>
            </>
          ) : (
            <span>Simpan Kata Sandi Baru</span>
          )}
        </button>
      </form>
    </>
  );
}

export function ResendVerificationForm({
  styles,
  email,
  errorMessage,
  statusNotice,
  isSubmitting,
  onEmailChange,
  onSubmit,
  onBackToLogin,
}: {
  styles: StylesMap;
  email: string;
  errorMessage: string | null;
  statusNotice: string | null;
  isSubmitting: boolean;
  onEmailChange: (val: string) => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onBackToLogin: () => void;
}) {
  return (
    <>
      <div className={styles.backNavRow}>
        <button
          type="button"
          onClick={onBackToLogin}
          className={styles.backLinkButton}
        >
          <Icon name="arrowLeft" size={16} />
          <span>Kembali ke form masuk</span>
        </button>
      </div>

      <header className={styles.formHeader}>
        <div className={styles.formHeaderText}>
          <div className={styles.formEyebrow}>
            <Icon name="mail" size={14} />
            <span>VERIFIKASI EMAIL</span>
          </div>
          <h2 className={styles.formHeading}>Kirim ulang tautan.</h2>
          <p className={styles.formSubtitle}>
            Masukkan alamat email yang digunakan saat mendaftar. Token verifikasi
            sebelumnya akan digantikan.
          </p>
        </div>
      </header>

      <AlertBanner
        styles={styles}
        errorMessage={errorMessage}
        statusNotice={statusNotice}
      />

      <form onSubmit={onSubmit} className={styles.formStack} noValidate>
        <div className={styles.fieldGroup}>
          <label htmlFor="resend-email" className={styles.fieldLabel}>
            Alamat email pendaftaran
          </label>
          <input
            id="resend-email"
            type="email"
            value={email}
            onChange={(e) => onEmailChange(e.target.value)}
            placeholder="nama@domain.com"
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
              <span>Mengirim tautan baru...</span>
            </>
          ) : (
            <span>Kirim Ulang Tautan Verifikasi</span>
          )}
        </button>
      </form>
    </>
  );
}

export function StatusStepView({
  styles,
  step,
  onBackToLogin,
}: {
  styles: StylesMap;
  step: "reset-done" | "navigating";
  onBackToLogin: () => void;
}) {
  if (step === "reset-done") {
    return (
      <div className={styles.successCard}>
        <div className={styles.successIconWrap}>
          <Icon name="check" size={28} />
        </div>
        <h2 className={styles.formHeading}>Kata sandi diperbarui.</h2>
        <p className={styles.formSubtitle}>
          Kata sandi akun Anda berhasil diperbarui. Silakan masuk kembali
          menggunakan kredensial baru.
        </p>
        <button
          type="button"
          onClick={onBackToLogin}
          className={styles.primaryButton}
        >
          Masuk Sekarang
        </button>
      </div>
    );
  }

  return (
    <div className={styles.successCard}>
      <div className={styles.navigatingSpinner} aria-hidden="true" />
      <h2 className={styles.formHeading}>Autentikasi Berhasil.</h2>
      <p className={styles.formSubtitle}>
        Mengarahkan Anda ke halaman IntegriOne...
      </p>
    </div>
  );
}
