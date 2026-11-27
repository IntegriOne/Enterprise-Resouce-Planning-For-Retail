"use client";

import { useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  submitCreateRole,
  submitUpdateRoleDetails,
} from "@/lib/graphql/iam-dashboard-mutations";
import {
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import { MANAGEMENT_DOMAIN_PRESETS } from "@/lib/graphql/iam-role-access-helpers";
import type { IamRole } from "@/lib/graphql/iam-types";

const REDIRECT_SUGGESTIONS = [
  "/identity",
  "/human-resource",
  "/procurement",
  "/warehouse",
  "/finance",
  "/accounting",
  "/product",
  "/supplier",
  "/vendor",
  "/commerce",
  "/account",
  "/profile",
  "/",
  "",
];

export function RoleFormModal({
  open,
  editingRole,
  initialDomain,
  onClose,
  onSaved,
}: {
  open: boolean;
  editingRole: IamRole | null;
  initialDomain?: string;
  onClose: () => void;
  onSaved: (role: IamRole | null) => void;
}) {
  if (!open) return null;
  return (
    <RoleFormInner
      key={editingRole?._id ?? `new-role-${initialDomain || "iam"}`}
      editingRole={editingRole}
      initialDomain={initialDomain}
      onClose={onClose}
      onSaved={onSaved}
    />
  );
}

function RoleFormInner({
  editingRole,
  initialDomain,
  onClose,
  onSaved,
}: {
  editingRole: IamRole | null;
  initialDomain?: string;
  onClose: () => void;
  onSaved: (role: IamRole | null) => void;
}) {
  const { csrfToken, openStepUpModal, pushToast } = useIdentityWorkspace();

  const resolvedInitialDomain =
    editingRole?.management_domain || initialDomain || "iam";
  const initialPreset = MANAGEMENT_DOMAIN_PRESETS.find(
    (p) => p.value === resolvedInitialDomain,
  );

  const [code, setCode] = useState(
    editingRole?.code ??
      (resolvedInitialDomain === "enterprise_foundation" ? "ef_operator" : ""),
  );
  const [name, setName] = useState(
    editingRole?.name ??
      (resolvedInitialDomain === "enterprise_foundation"
        ? "Operator Enterprise Foundation"
        : ""),
  );
  const [description, setDescription] = useState(editingRole?.description ?? "");
  const [managementDomain, setManagementDomain] = useState(resolvedInitialDomain);
  const [defaultRedirectPath, setDefaultRedirectPath] = useState(
    editingRole
      ? editingRole.default_redirect_path || ""
      : initialPreset?.suggestedRedirect ?? "/identity",
  );
  const [redirectPriority, setRedirectPriority] = useState(
    String(editingRole?.redirect_priority ?? 50),
  );
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function executeSave() {
    setIsSubmitting(true);
    setErrorMsg(null);
    const priorityNum = Number.parseInt(redirectPriority, 10);

    try {
      if (editingRole) {
        const updated = await submitUpdateRoleDetails(
          {
            id: editingRole._id,
            name,
            description,
            management_domain: managementDomain,
            default_redirect_path: defaultRedirectPath,
            redirect_priority: Number.isFinite(priorityNum) ? priorityNum : 50,
          },
          csrfToken,
        );
        pushToast("success", "Detail Peran Diperbarui", `${name.trim()} berhasil disimpan.`);
        onSaved(updated);
      } else {
        const created = await submitCreateRole(
          {
            code,
            name,
            description,
            management_domain: managementDomain,
            default_redirect_path: defaultRedirectPath,
            redirect_priority: Number.isFinite(priorityNum) ? priorityNum : 50,
          },
          csrfToken,
        );
        pushToast(
          "success",
          "Peran RBAC Dibuat",
          `${name.trim()} dibuat. Lanjutkan memilih izin atomiknya di panel detail.`,
        );
        onSaved(created);
      }
      onClose();
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      if (isStepUpRequiredError(err)) {
        openStepUpModal(
          "Perubahan struktur peran RBAC memerlukan verifikasi Step-Up MFA.",
          async () => {
            await executeSave();
          },
        );
        return;
      }
      setErrorMsg(formatIamDashboardError(err));
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (isSubmitting) return;
    if (!editingRole && !code.trim()) {
      setErrorMsg("Kode peran wajib diisi.");
      return;
    }
    if (!name.trim()) {
      setErrorMsg("Nama peran wajib diisi.");
      return;
    }
    void executeSave();
  }

  function applyDomainPreset(presetValue: string) {
    setManagementDomain(presetValue);
    const preset = MANAGEMENT_DOMAIN_PRESETS.find((p) => p.value === presetValue);
    if (preset && !editingRole) {
      setDefaultRedirectPath(preset.suggestedRedirect);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label={editingRole ? "Ubah Detail Peran" : "Buat Peran RBAC Baru"}
      onClick={(e) => {
        if (e.target === e.currentTarget && !isSubmitting) onClose();
      }}
    >
      <div className="max-h-[92dvh] w-full max-w-lg overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3 border-b border-[#dde3e8] pb-4">
          <div>
            <h2 className="text-lg font-semibold text-[#0b2439]">
              {editingRole ? "Ubah Detail Peran RBAC" : "Buat Peran RBAC Baru"}
            </h2>
            <p className="mt-0.5 text-xs text-[#5a6672]">
              Tetapkan domain management dan prioritas tujuan pasca-login, lalu
              pasang izin atomiknya di panel detail peran.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            aria-label="Tutup modal peran"
            className={styles.iconBtn}
          >
            <IdentityIcon name="close" size={16} />
          </button>
        </div>

        {errorMsg ? (
          <div
            role="alert"
            className="mt-4 rounded-lg border border-[#fecdca] bg-[#fef3f2] px-3 py-2 text-xs text-[#b42318]"
          >
            {errorMsg}
          </div>
        ) : null}

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          {/* Preset Domain Cepat */}
          <div className="space-y-1.5">
            <span className="block text-xs font-semibold text-[#0b2439]">
              Preset Domain Management
            </span>
            <div className="flex flex-wrap gap-1.5">
              {MANAGEMENT_DOMAIN_PRESETS.slice(0, 6).map((preset) => {
                const selected = managementDomain === preset.value;
                return (
                  <button
                    key={preset.value}
                    type="button"
                    disabled={isSubmitting}
                    onClick={() => applyDomainPreset(preset.value)}
                    className={`rounded-md border px-2 py-1 text-[11px] font-semibold transition-colors ${styles.monoText} ${
                      selected
                        ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                        : "border-[#dde3e8] bg-[#f9fafb] text-[#5a6672] hover:text-[#0b2439]"
                    }`}
                  >
                    {preset.value}
                  </button>
                );
              })}
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="role-code" className="text-xs font-semibold text-[#0b2439]">
                Kode Peran
              </label>
              <input
                id="role-code"
                type="text"
                value={code}
                disabled={Boolean(editingRole) || isSubmitting}
                onChange={(e) => setCode(e.target.value)}
                placeholder="mis. ef_operator"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="role-domain" className="text-xs font-semibold text-[#0b2439]">
                Management Domain
              </label>
              <input
                id="role-domain"
                type="text"
                value={managementDomain}
                disabled={isSubmitting}
                onChange={(e) => setManagementDomain(e.target.value)}
                placeholder="iam / enterprise_foundation / human_resource"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="role-name" className="text-xs font-semibold text-[#0b2439]">
              Nama Tampilan Peran
            </label>
            <input
              id="role-name"
              type="text"
              value={name}
              disabled={isSubmitting}
              onChange={(e) => setName(e.target.value)}
              placeholder="Contoh: Operator Enterprise Foundation"
              className={styles.fieldInput}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="role-desc" className="text-xs font-semibold text-[#0b2439]">
              Deskripsi (Opsional)
            </label>
            <input
              id="role-desc"
              type="text"
              value={description}
              disabled={isSubmitting}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Ringkasan tanggung jawab peran"
              className={styles.fieldInput}
            />
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="role-redirect" className="text-xs font-semibold text-[#0b2439]">
                Default Redirect Path
              </label>
              <input
                id="role-redirect"
                type="text"
                list="iam-redirect-suggestions"
                value={defaultRedirectPath}
                disabled={isSubmitting}
                onChange={(e) => setDefaultRedirectPath(e.target.value)}
                placeholder="/identity"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
              <datalist id="iam-redirect-suggestions">
                {REDIRECT_SUGGESTIONS.map((p) => (
                  <option key={p || "empty"} value={p} />
                ))}
              </datalist>
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="role-priority" className="text-xs font-semibold text-[#0b2439]">
                Redirect Priority (Angka Kecil Menang)
              </label>
              <input
                id="role-priority"
                type="number"
                min={1}
                max={32767}
                value={redirectPriority}
                disabled={isSubmitting}
                onChange={(e) => setRedirectPriority(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="flex items-center justify-end gap-2.5 border-t border-[#dde3e8] pt-4">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className={styles.btnSecondary}
            >
              Batal
            </button>
            <button type="submit" disabled={isSubmitting} className={styles.btnPrimary}>
              {isSubmitting ? "Menyimpan..." : editingRole ? "Simpan Perubahan" : "Buat Peran"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
