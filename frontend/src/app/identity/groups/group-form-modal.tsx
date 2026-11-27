"use client";

import { useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  submitCreateGroup,
  submitUpdateGroupDetails,
} from "@/lib/graphql/iam-dashboard-mutations";
import {
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import type { IamGroup } from "@/lib/graphql/iam-types";

export function GroupFormModal({
  open,
  editingGroup,
  onClose,
  onSaved,
}: {
  open: boolean;
  editingGroup: IamGroup | null;
  onClose: () => void;
  onSaved: (group: IamGroup | null) => void;
}) {
  if (!open) return null;
  return (
    <GroupFormInner
      key={editingGroup?._id ?? "new-group"}
      editingGroup={editingGroup}
      onClose={onClose}
      onSaved={onSaved}
    />
  );
}

function GroupFormInner({
  editingGroup,
  onClose,
  onSaved,
}: {
  editingGroup: IamGroup | null;
  onClose: () => void;
  onSaved: (group: IamGroup | null) => void;
}) {
  const { csrfToken, openStepUpModal, pushToast } = useIdentityWorkspace();

  const [code, setCode] = useState(editingGroup?.code ?? "");
  const [name, setName] = useState(editingGroup?.name ?? "");
  const [description, setDescription] = useState(editingGroup?.description ?? "");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function executeSave() {
    setIsSubmitting(true);
    setErrorMsg(null);
    try {
      if (editingGroup) {
        const updated = await submitUpdateGroupDetails(
          { id: editingGroup._id, name, description },
          csrfToken,
        );
        pushToast("success", "Grup Diperbarui", `${name.trim()} berhasil disimpan.`);
        onSaved(updated);
      } else {
        const created = await submitCreateGroup(
          { code, name, description },
          csrfToken,
        );
        pushToast("success", "Grup Dibuat", `${name.trim()} siap menampung anggota dan peran.`);
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
          "Perubahan struktur grup memerlukan verifikasi Step-Up MFA.",
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
    if (!editingGroup && !code.trim()) {
      setErrorMsg("Kode grup wajib diisi.");
      return;
    }
    if (!name.trim()) {
      setErrorMsg("Nama grup wajib diisi.");
      return;
    }
    void executeSave();
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label={editingGroup ? "Ubah Detail Grup" : "Buat Grup Baru"}
      onClick={(e) => {
        if (e.target === e.currentTarget && !isSubmitting) onClose();
      }}
    >
      <div className="w-full max-w-md rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3 border-b border-[#dde3e8] pb-4">
          <div>
            <h2 className="text-lg font-semibold text-[#0b2439]">
              {editingGroup ? "Ubah Detail Grup" : "Buat Grup Baru"}
            </h2>
            <p className="mt-0.5 text-xs text-[#5a6672]">
              Grup mewariskan peran berscope kepada seluruh anggota aktifnya.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            aria-label="Tutup modal grup"
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
          <div className="flex flex-col gap-1.5">
            <label htmlFor="group-code" className="text-xs font-semibold text-[#0b2439]">
              Kode Grup
            </label>
            <input
              id="group-code"
              type="text"
              value={code}
              disabled={Boolean(editingGroup) || isSubmitting}
              onChange={(e) => setCode(e.target.value)}
              placeholder="mis. hrm_admins_jakarta"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="group-name" className="text-xs font-semibold text-[#0b2439]">
              Nama Grup
            </label>
            <input
              id="group-name"
              type="text"
              value={name}
              disabled={isSubmitting}
              onChange={(e) => setName(e.target.value)}
              placeholder="Contoh: Tim SDM Regional Jakarta"
              className={styles.fieldInput}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="group-desc" className="text-xs font-semibold text-[#0b2439]">
              Deskripsi (Opsional)
            </label>
            <input
              id="group-desc"
              type="text"
              value={description}
              disabled={isSubmitting}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Catatan ruang lingkup anggota grup"
              className={styles.fieldInput}
            />
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
              {isSubmitting ? "Menyimpan..." : editingGroup ? "Simpan" : "Buat Grup"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
