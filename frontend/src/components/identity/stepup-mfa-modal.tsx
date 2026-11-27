"use client";

import { useState, type FormEvent } from "react";
import styles from "./identity.module.css";
import { IdentityIcon } from "./identity-icons";
import { useIdentityWorkspace } from "./identity-context";
import { submitVerifyMySessionMfa } from "@/lib/graphql/iam-dashboard-mutations";
import { formatIamDashboardError } from "@/lib/graphql/iam-dashboard-queries";

function StepUpMfaInner({
  reason,
  csrfToken,
  onClose,
  onSuccess,
}: {
  reason?: string;
  csrfToken: string | null;
  onClose: () => void;
  onSuccess: () => Promise<void>;
}) {
  const [totpCode, setTotpCode] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (isSubmitting) return;
    const trimmed = totpCode.trim();
    if (!trimmed) {
      setErrorMessage("Masukkan kode TOTP 6 digit dari aplikasi autentikator Anda.");
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      const ok = await submitVerifyMySessionMfa(trimmed, csrfToken);
      if (!ok) {
        setErrorMessage("Kode verifikasi tidak valid atau gagal diverifikasi.");
        return;
      }
      await onSuccess();
    } catch (err) {
      setErrorMessage(formatIamDashboardError(err));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label="Verifikasi Step-Up MFA"
    >
      <div className="w-full max-w-md rounded-xl border border-[#dde3e8] bg-white p-6 shadow-xl">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-2.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-[#e9eef2] text-[#133554]">
              <IdentityIcon name="shield" size={18} />
            </span>
            <div>
              <h2 className="text-base font-semibold text-[#0b2439]">
                Verifikasi Step-Up Sesi
              </h2>
              <p className={`text-[11px] text-[#5a6672] ${styles.monoText}`}>
                AUTH SENSITIVE GUARD
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup dialog step-up"
            className={styles.iconBtn}
          >
            <IdentityIcon name="close" size={16} />
          </button>
        </div>

        <p className="mt-3 text-sm leading-relaxed text-[#5a6672]">
          {reason ||
            "Aksi sensitif ini memerlukan konfirmasi ulang kode TOTP aktif pada sesi Anda. Bila TOTP belum aktif atau sesi melewati jendela 10 menit, aktifkan TOTP di Sesi & MFA atau masuk ulang."}
        </p>

        {errorMessage ? (
          <div
            role="alert"
            className="mt-3 rounded-lg border border-[#fecdca] bg-[#fef3f2] px-3 py-2 text-xs text-[#b42318]"
          >
            {errorMessage}
          </div>
        ) : null}

        <form onSubmit={(e) => void handleSubmit(e)} className="mt-4 space-y-4">
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="stepup-totp-code"
              className="text-xs font-semibold text-[#0b2439]"
            >
              Kode Autentikator (TOTP 6 Digit)
            </label>
            <input
              id="stepup-totp-code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              value={totpCode}
              onChange={(e) => setTotpCode(e.target.value)}
              placeholder="123456"
              disabled={isSubmitting}
              autoFocus
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
            <span className="text-[11px] text-[#5a6672]">
              Verifikasi berlaku selama jendela kebijakan re-autentikasi aktif.
            </span>
          </div>

          <div className="flex items-center justify-end gap-2.5">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className={styles.btnSecondary}
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className={styles.btnPrimary}
            >
              {isSubmitting ? "Memverifikasi..." : "Konfirmasi TOTP"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export function StepUpMfaModal() {
  const { stepUpRequest, closeStepUpModal, handleStepUpSuccess, csrfToken } =
    useIdentityWorkspace();

  if (!stepUpRequest) return null;

  return (
    <StepUpMfaInner
      reason={stepUpRequest.reason}
      csrfToken={csrfToken}
      onClose={closeStepUpModal}
      onSuccess={handleStepUpSuccess}
    />
  );
}
