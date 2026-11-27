"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { SensitiveBadge } from "@/components/identity/identity-primitives";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  fetchAllPermissionsCatalog,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import {
  submitAssignRolePermission,
  submitCreateRole,
} from "@/lib/graphql/iam-dashboard-mutations";
import {
  MANAGEMENT_DOMAIN_PRESETS,
  getPermissionDomain,
} from "@/lib/graphql/iam-role-access-helpers";
import type { IamPermission, IamRole } from "@/lib/graphql/iam-types";

/**
 * Panel inline untuk membuat Peran Kustom (`is_system: false`) sekaligus
 * memasang daftar izin atomik domain (mis. `enterprise_foundation.*`)
 * tanpa meninggalkan alur Detail Pengguna.
 */
export function QuickRoleBuilder({
  defaultDomain = "enterprise_foundation",
  onCancel,
  onCreatedAndPopulated,
}: {
  defaultDomain?: string;
  onCancel: () => void;
  onCreatedAndPopulated: (role: IamRole, attachedPermissions: IamPermission[]) => void;
}) {
  const { csrfToken, openStepUpModal, pushToast } = useIdentityWorkspace();

  const initialPreset =
    MANAGEMENT_DOMAIN_PRESETS.find((d) => d.value === defaultDomain) ??
    MANAGEMENT_DOMAIN_PRESETS[1];

  const [domain, setDomain] = useState(initialPreset.value);
  const [code, setCode] = useState(
    initialPreset.value === "enterprise_foundation" ? "ef_operator" : "",
  );
  const [name, setName] = useState(
    initialPreset.value === "enterprise_foundation"
      ? "Operator Enterprise Foundation"
      : "",
  );
  const [description, setDescription] = useState("");
  const [defaultRedirect, setDefaultRedirect] = useState(
    initialPreset.suggestedRedirect,
  );
  const [redirectPriority, setRedirectPriority] = useState("40");

  const [catalogPerms, setCatalogPerms] = useState<IamPermission[]>([]);
  const [isLoadingPerms, setIsLoadingPerms] = useState(true);
  const [permLoadError, setPermLoadError] = useState<string | null>(null);
  const [permSearch, setPermSearch] = useState("");
  const [selectedPermIds, setSelectedPermIds] = useState<Set<string>>(new Set());

  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    let active = true;
    async function loadCatalog() {
      setIsLoadingPerms(true);
      setPermLoadError(null);
      try {
        const list = await fetchAllPermissionsCatalog();
        if (!active) return;
        setCatalogPerms(list);
        // Pilih otomatis izin pada domain awal agar operator cukup meninjau atau menyesuaikan.
        const initialMatched = list
          .filter((p) => getPermissionDomain(p) === initialPreset.value)
          .map((p) => p._id);
        setSelectedPermIds(new Set(initialMatched));
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setPermLoadError(formatIamDashboardError(err));
      } finally {
        if (active) setIsLoadingPerms(false);
      }
    }
    void loadCatalog();
    return () => {
      active = false;
    };
  }, [initialPreset.value]);

  const filteredPermissions = useMemo(() => {
    const q = permSearch.trim().toLowerCase();
    return catalogPerms.filter((p) => {
      const pDomain = getPermissionDomain(p);
      const matchesDomain = !domain.trim() || pDomain === domain.trim();
      const matchesSearch =
        !q ||
        p.code.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q);
      return matchesDomain && matchesSearch;
    });
  }, [catalogPerms, domain, permSearch]);

  function handlePresetSelect(presetValue: string) {
    const preset = MANAGEMENT_DOMAIN_PRESETS.find((p) => p.value === presetValue);
    setDomain(presetValue);
    if (preset) {
      setDefaultRedirect(preset.suggestedRedirect);
      if (presetValue === "enterprise_foundation" && !code.trim()) {
        setCode("ef_operator");
        setName("Operator Enterprise Foundation");
      }
    }
    const matched = catalogPerms
      .filter((p) => getPermissionDomain(p) === presetValue)
      .map((p) => p._id);
    setSelectedPermIds(new Set(matched));
  }

  function togglePermission(permId: string) {
    setSelectedPermIds((prev) => {
      const next = new Set(prev);
      if (next.has(permId)) {
        next.delete(permId);
      } else {
        next.add(permId);
      }
      return next;
    });
  }

  function selectAllFiltered() {
    setSelectedPermIds((prev) => {
      const next = new Set(prev);
      for (const p of filteredPermissions) {
        next.add(p._id);
      }
      return next;
    });
  }

  function clearAllFiltered() {
    setSelectedPermIds((prev) => {
      const next = new Set(prev);
      for (const p of filteredPermissions) {
        next.delete(p._id);
      }
      return next;
    });
  }

  async function executeCreateAndAttach() {
    setIsSubmitting(true);
    setSubmitError(null);
    const priorityNum = Number.parseInt(redirectPriority, 10);

    let createdRole: IamRole | null = null;
    try {
      createdRole = await submitCreateRole(
        {
          code: code.trim(),
          name: name.trim(),
          description: description.trim() || null,
          management_domain: domain.trim(),
          default_redirect_path: defaultRedirect.trim(),
          redirect_priority: Number.isFinite(priorityNum) ? priorityNum : 40,
          is_system: false,
        },
        csrfToken,
      );

      if (!createdRole) {
        setSubmitError("Server IAM tidak mengembalikan objek peran yang dibuat.");
        return;
      }

      const targetPerms = catalogPerms.filter((p) => selectedPermIds.has(p._id));
      const attached: IamPermission[] = [];
      for (const perm of targetPerms) {
        await submitAssignRolePermission(createdRole._id, perm._id, csrfToken);
        attached.push(perm);
      }

      pushToast(
        "success",
        "Peran Kustom Siap Digunakan",
        `${createdRole.name} (${createdRole.code}) dibuat dengan ${attached.length} izin atomik.`,
      );
      onCreatedAndPopulated(createdRole, attached);
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      if (isStepUpRequiredError(err)) {
        const createdSoFar = createdRole;
        openStepUpModal(
          "Pembuatan peran kustom dan pemasangan izin memerlukan verifikasi Step-Up MFA.",
          async () => {
            if (!createdSoFar) {
              await executeCreateAndAttach();
              return;
            }
            const remaining = catalogPerms.filter((p) =>
              selectedPermIds.has(p._id),
            );
            for (const perm of remaining) {
              try {
                await submitAssignRolePermission(
                  createdSoFar._id,
                  perm._id,
                  csrfToken,
                );
              } catch {
                // Abaikan jika izin tertentu sudah terpasang sebelum step-up.
              }
            }
            onCreatedAndPopulated(createdSoFar, remaining);
          },
        );
        return;
      }
      if (createdRole) {
        setSubmitError(
          `Peran ${createdRole.code} sudah dibuat, tetapi pemasangan izin terhenti: ${formatIamDashboardError(err)}`,
        );
        onCreatedAndPopulated(createdRole, []);
        return;
      }
      setSubmitError(formatIamDashboardError(err));
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (isSubmitting) return;
    if (!code.trim() || !name.trim() || !domain.trim()) {
      setSubmitError("Kode peran, nama tampilan, dan domain management wajib diisi.");
      return;
    }
    void executeCreateAndAttach();
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4 rounded-xl border border-[#133554]/30 bg-[#f4f7f8] p-4"
    >
      <div className="flex flex-wrap items-start justify-between gap-2 border-b border-[#dde3e8] pb-3">
        <div>
          <p className={`text-[11px] font-semibold uppercase tracking-wider text-[#133554] ${styles.monoText}`}>
            QUICK ROLE BUILDER · PERAN KUSTOM NON-SISTEM
          </p>
          <h4 className="mt-0.5 text-sm font-semibold text-[#0b2439]">
            Buat Peran Baru &amp; Pilih Izin Atomik Sekaligus
          </h4>
          <p className="mt-0.5 text-xs text-[#5a6672]">
            Peran kustom langsung muncul di daftar dan siap dipasang ke akun pengguna ini.
          </p>
        </div>
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className={`${styles.btnSecondary} ${styles.btnSmall}`}
        >
          Tutup
        </button>
      </div>

      {submitError ? (
        <div
          role="alert"
          className="rounded-lg border border-[#fecdca] bg-[#fef3f2] px-3 py-2 text-xs text-[#b42318]"
        >
          {submitError}
        </div>
      ) : null}

      <div className="space-y-1.5">
        <span className="block text-xs font-semibold text-[#0b2439]">
          Pilih Domain Management
        </span>
        <div className="flex flex-wrap gap-1.5">
          {MANAGEMENT_DOMAIN_PRESETS.slice(0, 4).map((preset) => {
            const active = domain === preset.value;
            return (
              <button
                key={preset.value}
                type="button"
                disabled={isSubmitting}
                onClick={() => handlePresetSelect(preset.value)}
                className={`rounded-md border px-2.5 py-1 text-[11px] font-semibold transition-colors ${styles.monoText} ${
                  active
                    ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                    : "border-[#dde3e8] bg-white text-[#5a6672] hover:text-[#0b2439]"
                }`}
              >
                {preset.value}
              </button>
            );
          })}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="flex flex-col gap-1">
          <label htmlFor="quick-role-code" className="text-xs font-semibold text-[#0b2439]">
            Kode Peran
          </label>
          <input
            id="quick-role-code"
            type="text"
            value={code}
            disabled={isSubmitting}
            onChange={(e) => setCode(e.target.value)}
            placeholder="mis. ef_operator"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="quick-role-name" className="text-xs font-semibold text-[#0b2439]">
            Nama Peran
          </label>
          <input
            id="quick-role-name"
            type="text"
            value={name}
            disabled={isSubmitting}
            onChange={(e) => setName(e.target.value)}
            placeholder="Operator Enterprise Foundation"
            className={styles.fieldInput}
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="quick-role-domain" className="text-xs font-semibold text-[#0b2439]">
            Management Domain
          </label>
          <input
            id="quick-role-domain"
            type="text"
            value={domain}
            disabled={isSubmitting}
            onChange={(e) => setDomain(e.target.value)}
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
        </div>
        <div className="grid grid-cols-3 gap-2">
          <div className="col-span-2 flex flex-col gap-1">
            <label htmlFor="quick-role-redirect" className="text-xs font-semibold text-[#0b2439]">
              Default Redirect
            </label>
            <input
              id="quick-role-redirect"
              type="text"
              value={defaultRedirect}
              disabled={isSubmitting}
              onChange={(e) => setDefaultRedirect(e.target.value)}
              placeholder="/human-resource"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
          </div>
          <div className="flex flex-col gap-1">
            <label htmlFor="quick-role-prio" className="text-xs font-semibold text-[#0b2439]">
              Prioritas
            </label>
            <input
              id="quick-role-prio"
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
      </div>

      <div className="flex flex-col gap-1">
        <label htmlFor="quick-role-desc" className="text-xs font-semibold text-[#0b2439]">
          Deskripsi Singkat (Opsional)
        </label>
        <input
          id="quick-role-desc"
          type="text"
          value={description}
          disabled={isSubmitting}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Hak baca/kelola data master legal entity, unit, lokasi, dan mitra"
          className={styles.fieldInput}
        />
      </div>

      {/* Pemilih Permission Atomik */}
      <div className="space-y-2 rounded-lg border border-[#dde3e8] bg-white p-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <p className="text-xs font-semibold text-[#0b2439]">
              Pilih Izin Atomik ({selectedPermIds.size} dipilih)
            </p>
            <p className="text-[11px] text-[#5a6672]">
              Menampilkan izin domain <code className={styles.monoText}>{domain || "semua"}</code>.
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={selectAllFiltered}
              disabled={isSubmitting || filteredPermissions.length === 0}
              className={`${styles.btnSecondary} ${styles.btnSmall}`}
            >
              Pilih Semua
            </button>
            <button
              type="button"
              onClick={clearAllFiltered}
              disabled={isSubmitting || filteredPermissions.length === 0}
              className={`${styles.btnSecondary} ${styles.btnSmall}`}
            >
              Bersihkan
            </button>
          </div>
        </div>

        <input
          type="text"
          value={permSearch}
          onChange={(e) => setPermSearch(e.target.value)}
          placeholder="Saring kode izin (mis. parties, legal_entities)..."
          aria-label="Saring kode izin peran baru"
          className={styles.fieldInput}
        />

        {isLoadingPerms ? (
          <div className="space-y-1.5 py-2">
            <div className={`${styles.skeletonShimmer} h-8 w-full`} />
            <div className={`${styles.skeletonShimmer} h-8 w-full`} />
          </div>
        ) : permLoadError ? (
          <p className="text-xs text-[#b42318]">{permLoadError}</p>
        ) : filteredPermissions.length === 0 ? (
          <p className="py-3 text-xs text-[#5a6672]">
            Belum ada katalog izin terdaftar di server IAM untuk domain{" "}
            <code className={styles.monoText}>{domain || "(kosong)"}</code>. Anda masih dapat
            membuat kerangka peran lalu memasang izin dari katalog yang tersedia.
          </p>
        ) : (
          <div className="max-h-48 divide-y divide-[#dde3e8] overflow-y-auto rounded-lg border border-[#dde3e8]">
            {filteredPermissions.map((perm) => {
              const checked = selectedPermIds.has(perm._id);
              return (
                <label
                  key={perm.id}
                  className="flex cursor-pointer items-start justify-between gap-2.5 px-3 py-2 text-xs hover:bg-[#f9fafb]"
                >
                  <span className="flex items-start gap-2.5 min-w-0">
                    <input
                      type="checkbox"
                      checked={checked}
                      disabled={isSubmitting}
                      onChange={() => togglePermission(perm._id)}
                      className="mt-0.5 h-4 w-4 shrink-0 accent-[#133554]"
                    />
                    <span className="min-w-0">
                      <span className={`block truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                        {perm.code}
                      </span>
                      <span className="block truncate text-[11px] text-[#5a6672]">
                        {perm.description}
                      </span>
                    </span>
                  </span>
                  <SensitiveBadge sensitive={perm.is_sensitive} />
                </label>
              );
            })}
          </div>
        )}
      </div>

      <div className="flex flex-wrap items-center justify-end gap-2 pt-1">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className={styles.btnSecondary}
        >
          Batal
        </button>
        <button
          type="submit"
          disabled={isSubmitting || !code.trim() || !name.trim()}
          className={styles.btnPrimary}
        >
          <IdentityIcon name="plus" size={15} />
          <span>
            {isSubmitting
              ? "Membuat Peran & Memasang Izin..."
              : `Buat Peran (${selectedPermIds.size} Izin)`}
          </span>
        </button>
      </div>
    </form>
  );
}
