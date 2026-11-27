"use client";

import { useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { ConfirmModal } from "@/components/identity/identity-primitives";
import {
  submitChangeUsername,
  validateIamPassword,
  validateIamUsername,
} from "@/lib/graphql/iam-browser-auth";
import {
  submitChangeUserPassword,
  submitConfirmMyTotp,
  submitDisableMyTotp,
  submitEnrollMyTotp,
  submitRegenerateMyRecoveryCodes,
  submitUpdateMyProfile,
} from "@/lib/graphql/iam-dashboard-mutations";
import {
  checkUsernameAvailabilityViaGraphQL,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
  type UsernameAvailabilityResult,
} from "@/lib/graphql/iam-dashboard-queries";
import type { IamUser } from "@/lib/graphql/iam-types";

export function OperatorSecurityPanel() {
  const { me } = useIdentityWorkspace();
  return <OperatorSecurityInner key={me?._id ?? "anon"} initialMe={me} />;
}

function OperatorSecurityInner({ initialMe }: { initialMe: IamUser | null }) {
  const {
    me,
    snapshot,
    csrfToken,
    pushToast,
    openStepUpModal,
    refreshOperatorContext,
  } = useIdentityWorkspace();

  const activeMe = me ?? initialMe;
  const [displayName, setDisplayName] = useState(activeMe?.display_name ?? "");
  const [locale, setLocale] = useState(activeMe?.locale || "id-ID");
  const [timezone, setTimezone] = useState(activeMe?.timezone || "Asia/Jakarta");
  const [newUsername, setNewUsername] = useState("");
  const [usernameChecking, setUsernameChecking] = useState(false);
  const [usernameStatus, setUsernameStatus] =
    useState<UsernameAvailabilityResult | null>(null);
  const [usernameConflictMsg, setUsernameConflictMsg] = useState<string | null>(
    null,
  );

  // TOTP enrollment & recovery codes
  const [otpAuthUri, setOtpAuthUri] = useState<string | null>(null);
  const [confirmTotpCode, setConfirmTotpCode] = useState("");
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
  const [confirmDisableMfaOpen, setConfirmDisableMfaOpen] = useState(false);

  // Change password
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [isBusy, setIsBusy] = useState(false);

  useEffect(() => {
    const normalized = newUsername.trim().toLowerCase();
    if (!normalized || validateIamUsername(normalized) !== null) return;

    let active = true;
    const timer = window.setTimeout(async () => {
      if (!active) return;
      setUsernameChecking(true);
      try {
        const res = await checkUsernameAvailabilityViaGraphQL(normalized);
        if (!active) return;
        setUsernameStatus(res);
        if (!res.available && res.existingUser) {
          if (res.existingUser._id === activeMe?._id) {
            setUsernameConflictMsg("Ini adalah username aktif Anda saat ini.");
          } else {
            setUsernameConflictMsg(
              `Username '${res.normalizedUsername}' sudah digunakan oleh ${res.existingUser.display_name} (${res.existingUser.kind}).`,
            );
          }
        } else {
          setUsernameConflictMsg(null);
        }
      } catch {
        // Abaikan jika operator tidak memiliki izin baca direktori users
      } finally {
        if (active) setUsernameChecking(false);
      }
    }, 450);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [newUsername, activeMe?._id]);

  async function handleUsernameBlurCheck() {
    const normalized = newUsername.trim().toLowerCase();
    if (!normalized || validateIamUsername(normalized) !== null) return;
    setUsernameChecking(true);
    try {
      const res = await checkUsernameAvailabilityViaGraphQL(normalized);
      setUsernameStatus(res);
      if (!res.available && res.existingUser) {
        setUsernameConflictMsg(
          res.existingUser._id === activeMe?._id
            ? "Ini adalah username aktif Anda saat ini."
            : `Username '${res.normalizedUsername}' sudah digunakan oleh ${res.existingUser.display_name} (${res.existingUser.kind}).`,
        );
      } else {
        setUsernameConflictMsg(null);
      }
    } catch {
      // Abaikan jika tidak tersedia
    } finally {
      setUsernameChecking(false);
    }
  }

  async function runSecurityOp(
    reason: string,
    op: () => Promise<void>,
    successTitle: string,
  ) {
    setIsBusy(true);
    try {
      await op();
      pushToast("success", successTitle);
      await refreshOperatorContext();
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      if (isStepUpRequiredError(err)) {
        openStepUpModal(reason, async () => {
          await op();
          pushToast("success", successTitle);
          await refreshOperatorContext();
        });
        return;
      }
      pushToast("error", "Pembaruan Gagal", formatIamDashboardError(err));
    } finally {
      setIsBusy(false);
    }
  }

  function handleProfileSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!displayName.trim() || isBusy) {
      pushToast("error", "Validasi", "Nama tampilan wajib diisi.");
      return;
    }
    void runSecurityOp(
      "Memperbarui profil operator.",
      async () => {
        await submitUpdateMyProfile(
          { display_name: displayName, locale, timezone },
          csrfToken,
        );
      },
      "Profil operator diperbarui",
    );
  }

  function handleUsernameSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (isBusy) return;
    const err = validateIamUsername(newUsername);
    if (err) {
      pushToast("error", "Username Tidak Valid", err);
      return;
    }
    void runSecurityOp(
      "Pengubahan username memerlukan konfirmasi sesi.",
      async () => {
        await submitChangeUsername(newUsername, csrfToken);
        setNewUsername("");
      },
      "Username berhasil diperbarui",
    );
  }

  function handleConfirmTotp(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!confirmTotpCode.trim() || isBusy) return;
    void runSecurityOp(
      "Konfirmasi faktor TOTP baru.",
      async () => {
        const res = await submitConfirmMyTotp(confirmTotpCode, csrfToken);
        setRecoveryCodes(res.recoveryCodes);
        setOtpAuthUri(null);
        setConfirmTotpCode("");
      },
      "MFA TOTP berhasil diaktifkan",
    );
  }

  function handleChangePassword(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!currentPassword || isBusy) {
      pushToast("error", "Validasi", "Masukkan kata sandi saat ini.");
      return;
    }
    const err = validateIamPassword(newPassword, [me?.display_name || ""]);
    if (err) {
      pushToast("error", "Kata Sandi Baru Lemah", err);
      return;
    }
    void runSecurityOp(
      "Penggantian kata sandi akun memerlukan verifikasi sensitif.",
      async () => {
        const res = await submitChangeUserPassword(
          currentPassword,
          newPassword,
          csrfToken,
        );
        setCurrentPassword("");
        setNewPassword("");
        if (res.sessionsRevoked > 0) {
          pushToast(
            "info",
            "Sesi Lama Dicabut",
            `${res.sessionsRevoked} sesi lain telah dihentikan.`,
          );
        }
      },
      "Kata sandi berhasil diganti",
    );
  }

  return (
    <div className="space-y-5">
      {/* Panel MFA & Step-Up */}
      <section className={styles.hairlinePanel}>
        <div className="flex items-center justify-between gap-3 border-b border-[#dde3e8] px-5 py-4">
          <div>
            <h2 className="text-sm font-semibold text-[#0b2439]">
              Autentikasi Dua Langkah (TOTP)
            </h2>
            <p className="text-xs text-[#5a6672]">
              Diperlukan untuk mutasi sensitif &amp; persetujuan RBAC
            </p>
          </div>
          <span
            className={`rounded-md border px-2 py-0.5 text-[11px] font-semibold ${styles.monoText} ${
              snapshot?.mfa_enabled
                ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                : "border-[#fedf89] bg-[#fffaeb] text-[#b54708]"
            }`}
          >
            {snapshot?.mfa_enabled ? "MFA AKTIF" : "BELUM AKTIF"}
          </span>
        </div>

        <div className="space-y-4 p-5 text-xs">
          {!snapshot?.mfa_enabled ? (
            <div className="space-y-3">
              <p className="text-[#5a6672]">
                Daftarkan aplikasi autentikator (TOTP) agar dapat melakukan
                verifikasi step-up pada tindakan sensitif.
              </p>
              {!otpAuthUri ? (
                <button
                  type="button"
                  disabled={isBusy}
                  onClick={() =>
                    void runSecurityOp(
                      "Mendaftarkan faktor TOTP baru.",
                      async () => {
                        const res = await submitEnrollMyTotp(csrfToken);
                        setOtpAuthUri(res.otpAuthUri);
                      },
                      "Secret TOTP siap dipindai",
                    )
                  }
                  className={styles.btnPrimary}
                >
                  <IdentityIcon name="shield" size={15} />
                  <span>Mulai Aktivasi TOTP</span>
                </button>
              ) : (
                <form onSubmit={handleConfirmTotp} className="space-y-3">
                  <div className="rounded-lg border border-[#dde3e8] bg-[#f4f7f8] p-3">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-semibold text-[#0b2439]">
                        URI OTPAuth (Impor ke Authenticator)
                      </span>
                      <button
                        type="button"
                        onClick={() => {
                          void navigator.clipboard.writeText(otpAuthUri);
                          pushToast("info", "URI Disalin");
                        }}
                        className="font-semibold text-[#133554] hover:underline"
                      >
                        Salin URI
                      </button>
                    </div>
                    <p className={`mt-1.5 break-all text-[11px] text-[#5a6672] ${styles.monoText}`}>
                      {otpAuthUri}
                    </p>
                  </div>

                  <div className="flex gap-2">
                    <input
                      type="text"
                      inputMode="numeric"
                      value={confirmTotpCode}
                      onChange={(e) => setConfirmTotpCode(e.target.value)}
                      placeholder="Kode 6 digit konfirmasi"
                      className={`${styles.fieldInput} ${styles.monoText}`}
                    />
                    <button
                      type="submit"
                      disabled={isBusy}
                      className={`${styles.btnPrimary} shrink-0`}
                    >
                      Konfirmasi
                    </button>
                  </div>
                </form>
              )}
            </div>
          ) : (
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                disabled={isBusy}
                onClick={() =>
                  openStepUpModal(
                    "Konfirmasi kode TOTP sesi ini untuk menjaga jendela sensitif.",
                  )
                }
                className={`${styles.btnPrimary} ${styles.btnSmall}`}
              >
                <IdentityIcon name="shield" size={14} />
                <span>Verifikasi Step-Up Sesi</span>
              </button>
              <button
                type="button"
                disabled={isBusy}
                onClick={() =>
                  void runSecurityOp(
                    "Membuat ulang kode pemulihan membutuhkan verifikasi Step-Up.",
                    async () => {
                      const codes = await submitRegenerateMyRecoveryCodes(csrfToken);
                      setRecoveryCodes(codes);
                    },
                    "Kode pemulihan cadangan baru diterbitkan",
                  )
                }
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Regenerasi Recovery Codes
              </button>
              <button
                type="button"
                disabled={isBusy}
                onClick={() => setConfirmDisableMfaOpen(true)}
                className={`${styles.btnDanger} ${styles.btnSmall}`}
              >
                Nonaktifkan MFA
              </button>
            </div>
          )}

          {recoveryCodes && recoveryCodes.length > 0 ? (
            <div className="rounded-xl border border-[#abefc6] bg-[#ecfdf3] p-3.5">
              <p className="font-semibold text-[#067647]">
                Simpan Kode Pemulihan Cadangan Ini (Sekali Tampil)
              </p>
              <div className={`mt-2 grid grid-cols-2 gap-1.5 text-[11px] font-semibold text-[#0b2439] ${styles.monoText}`}>
                {recoveryCodes.map((code) => (
                  <span key={code} className="rounded bg-white px-2 py-1">
                    {code}
                  </span>
                ))}
              </div>
            </div>
          ) : null}
        </div>
      </section>

      {/* Panel Profil & Username */}
      <section className={styles.hairlinePanel}>
        <div className="border-b border-[#dde3e8] px-5 py-4">
          <h2 className="text-sm font-semibold text-[#0b2439]">
            Profil &amp; Username Operator
          </h2>
        </div>
        <div className="space-y-5 p-5">
          <form onSubmit={handleProfileSubmit} className="space-y-3">
            <div className="flex flex-col gap-1">
              <label htmlFor="op-display-name" className="text-xs font-semibold text-[#0b2439]">
                Nama Tampilan
              </label>
              <input
                id="op-display-name"
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                className={styles.fieldInput}
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1">
                <label htmlFor="op-locale" className="text-xs font-semibold text-[#0b2439]">
                  Locale
                </label>
                <input
                  id="op-locale"
                  type="text"
                  value={locale}
                  onChange={(e) => setLocale(e.target.value)}
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
              </div>
              <div className="flex flex-col gap-1">
                <label htmlFor="op-tz" className="text-xs font-semibold text-[#0b2439]">
                  Timezone
                </label>
                <input
                  id="op-tz"
                  type="text"
                  value={timezone}
                  onChange={(e) => setTimezone(e.target.value)}
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isBusy}
              className={`${styles.btnSecondary} ${styles.btnSmall}`}
            >
              Simpan Profil
            </button>
          </form>

          <form
            onSubmit={handleUsernameSubmit}
            className="space-y-2 border-t border-[#dde3e8] pt-4"
          >
            <label htmlFor="op-username" className="block text-xs font-semibold text-[#0b2439]">
              Ganti Username (`changeUsername`)
            </label>
            <div className="flex gap-2">
              <input
                id="op-username"
                type="text"
                value={newUsername}
                onChange={(e) => {
                  setNewUsername(e.target.value);
                  setUsernameStatus(null);
                  setUsernameConflictMsg(null);
                }}
                onBlur={() => void handleUsernameBlurCheck()}
                placeholder="username.baru"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
              <button
                type="submit"
                disabled={isBusy}
                className={`${styles.btnSecondary} ${styles.btnSmall} shrink-0`}
              >
                Perbarui
              </button>
            </div>
            {usernameChecking ? (
              <p className="text-[11px] text-[#5a6672]">
                Memeriksa ketersediaan username via GraphQL...
              </p>
            ) : usernameConflictMsg ? (
              <p className="text-xs text-[#b42318]">{usernameConflictMsg}</p>
            ) : usernameStatus?.available &&
              usernameStatus.normalizedUsername ===
                newUsername.trim().toLowerCase() ? (
              <p className="text-[11px] font-medium text-[#067647]">
                Username `{usernameStatus.normalizedUsername}` tersedia.
              </p>
            ) : null}
          </form>
        </div>
      </section>

      {/* Panel Ganti Kata Sandi */}
      <section className={styles.hairlinePanel}>
        <div className="border-b border-[#dde3e8] px-5 py-4">
          <h2 className="text-sm font-semibold text-[#0b2439]">
            Pembaruan Kata Sandi Pribadi
          </h2>
        </div>
        <form onSubmit={handleChangePassword} className="space-y-3 p-5">
          <div className="flex flex-col gap-1">
            <label htmlFor="op-cur-pass" className="text-xs font-semibold text-[#0b2439]">
              Kata Sandi Saat Ini
            </label>
            <input
              id="op-cur-pass"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className={styles.fieldInput}
            />
          </div>
          <div className="flex flex-col gap-1">
            <label htmlFor="op-new-pass" className="text-xs font-semibold text-[#0b2439]">
              Kata Sandi Baru (12–128 karakter)
            </label>
            <input
              id="op-new-pass"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className={styles.fieldInput}
            />
          </div>
          <button
            type="submit"
            disabled={isBusy}
            className={`${styles.btnSecondary} ${styles.btnSmall}`}
          >
            Ganti Kata Sandi
          </button>
        </form>
      </section>

      <ConfirmModal
        open={confirmDisableMfaOpen}
        title="Nonaktifkan MFA TOTP?"
        description="Aksi sensitif akan ditolak sampai faktor TOTP baru didaftarkan bila peran mewajibkan MFA."
        confirmLabel="Nonaktifkan MFA"
        isBusy={isBusy}
        onCancel={() => setConfirmDisableMfaOpen(false)}
        onConfirm={() => {
          setConfirmDisableMfaOpen(false);
          void runSecurityOp(
            "Menonaktifkan TOTP memerlukan verifikasi Step-Up.",
            async () => {
              await submitDisableMyTotp(csrfToken);
            },
            "Faktor TOTP dinonaktifkan",
          );
        }}
      />
    </div>
  );
}
