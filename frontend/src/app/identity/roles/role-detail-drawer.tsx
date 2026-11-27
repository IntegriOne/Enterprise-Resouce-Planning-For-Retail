"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ActiveBooleanBadge,
  ConfirmModal,
  ErrorPanel,
  SensitiveBadge,
  SlideOverDrawer,
} from "@/components/identity/identity-primitives";
import {
  fetchAllPermissionsCatalog,
  fetchAllRolePermissions,
  fetchRoleDetail,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import {
  submitAssignRolePermission,
  submitRevokeRolePermission,
  submitRoleLifecycleAction,
} from "@/lib/graphql/iam-dashboard-mutations";
import {
  extractUniqueDomains,
  getPermissionDomain,
} from "@/lib/graphql/iam-role-access-helpers";
import type { IamPermission, IamRole } from "@/lib/graphql/iam-types";

export function RoleDetailDrawer({
  roleId,
  onClose,
  onEditRequested,
  onUpdated,
}: {
  roleId: string | null;
  onClose: () => void;
  onEditRequested: (role: IamRole) => void;
  onUpdated: () => void;
}) {
  if (!roleId) return null;
  return (
    <RoleDetailInner
      key={roleId}
      roleId={roleId}
      onClose={onClose}
      onEditRequested={onEditRequested}
      onUpdated={onUpdated}
    />
  );
}

function RoleDetailInner({
  roleId,
  onClose,
  onEditRequested,
  onUpdated,
}: {
  roleId: string;
  onClose: () => void;
  onEditRequested: (role: IamRole) => void;
  onUpdated: () => void;
}) {
  const { csrfToken, pushToast, openStepUpModal } = useIdentityWorkspace();

  const [role, setRole] = useState<IamRole | null>(null);
  const [assignedPerms, setAssignedPerms] = useState<IamPermission[]>([]);
  const [catalogPerms, setCatalogPerms] = useState<IamPermission[]>([]);
  const [permQuery, setPermQuery] = useState("");
  const [permDomainFilter, setPermDomainFilter] = useState<string>("ALL");
  const [isLoading, setIsLoading] = useState(true);
  const [isMutating, setIsMutating] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const loadRole = useCallback(async (id: string) => {
    setIsLoading(true);
    setErrorMsg(null);
    try {
      const [detail, rolePerms, catPerms] = await Promise.all([
        fetchRoleDetail(id),
        fetchAllRolePermissions(id),
        fetchAllPermissionsCatalog(),
      ]);
      setRole(detail);
      setAssignedPerms(rolePerms);
      setCatalogPerms(catPerms);
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      setErrorMsg(formatIamDashboardError(err));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    let active = true;
    async function init() {
      try {
        const [detail, rolePerms, catPerms] = await Promise.all([
          fetchRoleDetail(roleId),
          fetchAllRolePermissions(roleId),
          fetchAllPermissionsCatalog(),
        ]);
        if (!active) return;
        setRole(detail);
        setAssignedPerms(rolePerms);
        setCatalogPerms(catPerms);
        if (detail?.management_domain) {
          const hasMatchingDomain = catPerms.some(
            (p) => getPermissionDomain(p) === detail.management_domain,
          );
          if (hasMatchingDomain && !detail.is_system) {
            setPermDomainFilter(detail.management_domain);
          }
        }
        setErrorMsg(null);
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setErrorMsg(formatIamDashboardError(err));
      } finally {
        if (active) setIsLoading(false);
      }
    }
    void init();
    return () => {
      active = false;
    };
  }, [roleId]);

  async function runSensitiveMutation(
    reason: string,
    op: () => Promise<void>,
    successMsg: string,
    closeAfter = false,
  ) {
    setIsMutating(true);
    try {
      await op();
      pushToast("success", successMsg);
      if (closeAfter) {
        onClose();
      } else if (roleId) {
        await loadRole(roleId);
      }
      onUpdated();
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      if (isStepUpRequiredError(err)) {
        openStepUpModal(reason, async () => {
          await op();
          pushToast("success", successMsg);
          if (closeAfter) {
            onClose();
          } else if (roleId) {
            await loadRole(roleId);
          }
          onUpdated();
        });
        return;
      }
      pushToast("error", "Aksi Gagal", formatIamDashboardError(err));
    } finally {
      setIsMutating(false);
      setConfirmDeleteOpen(false);
    }
  }

  const assignedSet = useMemo(
    () => new Set(assignedPerms.map((p) => p._id)),
    [assignedPerms],
  );

  const catalogDomains = useMemo(
    () => extractUniqueDomains(catalogPerms),
    [catalogPerms],
  );

  const filteredCatalog = useMemo(() => {
    const q = permQuery.trim().toLowerCase();
    return catalogPerms.filter((p) => {
      const matchesDomain =
        permDomainFilter === "ALL" || getPermissionDomain(p) === permDomainFilter;
      const matchesQ =
        !q ||
        p.code.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q);
      return matchesDomain && matchesQ;
    });
  }, [catalogPerms, permDomainFilter, permQuery]);

  const unattachedFiltered = useMemo(
    () => filteredCatalog.filter((p) => !assignedSet.has(p._id)),
    [filteredCatalog, assignedSet],
  );

  async function handleAttachAllFiltered() {
    if (!role || role.is_system || unattachedFiltered.length === 0 || isMutating) {
      return;
    }
    const targets = [...unattachedFiltered];
    await runSensitiveMutation(
      `Memasang ${targets.length} izin atomik ke peran ${role.code} memerlukan Step-Up MFA.`,
      async () => {
        for (const perm of targets) {
          await submitAssignRolePermission(role._id, perm._id, csrfToken);
        }
      },
      `${targets.length} izin berhasil ditambahkan ke ${role.code}`,
    );
  }

  return (
    <>
      <SlideOverDrawer
        open={Boolean(roleId)}
        title={role?.name || "Detail Peran RBAC"}
        subtitle={role ? `${role.code} · domain ${role.management_domain}` : undefined}
        onClose={onClose}
      >
        {isLoading && !role ? (
          <div className="space-y-3">
            <div className={`${styles.skeletonShimmer} h-14 w-full`} />
            <div className={`${styles.skeletonShimmer} h-52 w-full`} />
          </div>
        ) : errorMsg ? (
          <ErrorPanel
            message={errorMsg}
            onRetry={() => {
              if (roleId) void loadRole(roleId);
            }}
          />
        ) : role ? (
          <div className="space-y-6">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#dde3e8] pb-4">
              <div className="flex flex-wrap items-center gap-2">
                <ActiveBooleanBadge active={role.is_active} />
                {role.is_system ? (
                  <span className={`rounded-md border border-[#dde3e8] bg-[#e9eef2] px-2 py-0.5 text-[11px] font-semibold text-[#133554] ${styles.monoText}`}>
                    SYSTEM
                  </span>
                ) : (
                  <span className={`rounded-md border border-[#abefc6] bg-[#ecfdf3] px-2 py-0.5 text-[11px] font-semibold text-[#067647] ${styles.monoText}`}>
                    KUSTOM
                  </span>
                )}
                <span className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                  redirect: {role.default_redirect_path || "(none)"} · P{role.redirect_priority}
                </span>
              </div>

              {!role.is_system ? (
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    disabled={isMutating}
                    onClick={() => onEditRequested(role)}
                    className={`${styles.btnSecondary} ${styles.btnSmall}`}
                  >
                    <IdentityIcon name="edit" size={13} />
                    <span>Ubah</span>
                  </button>
                  <button
                    type="button"
                    disabled={isMutating}
                    onClick={() =>
                      void runSensitiveMutation(
                        "Mengubah status aktif peran RBAC membutuhkan verifikasi sensitif.",
                        async () => {
                          await submitRoleLifecycleAction(
                            role.is_active ? "deactivate" : "activate",
                            role._id,
                            csrfToken,
                          );
                        },
                        role.is_active ? "Peran dinonaktifkan" : "Peran diaktifkan",
                      )
                    }
                    className={`${styles.btnSecondary} ${styles.btnSmall}`}
                  >
                    {role.is_active ? "Nonaktifkan" : "Aktifkan"}
                  </button>
                  <button
                    type="button"
                    disabled={isMutating}
                    onClick={() => setConfirmDeleteOpen(true)}
                    className={`${styles.btnDanger} ${styles.btnSmall}`}
                  >
                    Hapus
                  </button>
                </div>
              ) : null}
            </div>

            {role.description ? (
              <p className="text-sm text-[#5a6672]">{role.description}</p>
            ) : null}

            {role.is_system ? (
              <div className="rounded-xl border border-[#dde3e8] bg-[#f4f7f8] p-3.5 text-xs text-[#5a6672]">
                <strong className="text-[#0b2439]">Peran Bawaan Sistem (Immutable):</strong>{" "}
                Izin pada peran ini disinkronkan otoritatif dari manifest katalog{" "}
                <code className={styles.monoText}>backend/shared/rbac/catalogs/</code> dan tidak
                dapat diubah atau dihapus melalui UI. Untuk kombinasi izin baru, buat{" "}
                <strong className="text-[#0b2439]">Peran Kustom</strong>.
              </div>
            ) : null}

            {/* Daftar Permission Terpasang */}
            <div className="space-y-2.5">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                Izin Terpasang pada Peran ({assignedPerms.length})
              </h3>
              <div className="max-h-60 divide-y divide-[#dde3e8] overflow-y-auto rounded-xl border border-[#dde3e8]">
                {assignedPerms.length === 0 ? (
                  <p className="p-4 text-xs text-[#5a6672]">
                    Belum ada izin terpasang pada peran ini. Pilih izin dari katalog di bawah.
                  </p>
                ) : (
                  assignedPerms.map((p) => (
                    <div
                      key={p.id}
                      className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2.5 text-xs"
                    >
                      <div className="min-w-0">
                        <p className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                          {p.code}
                        </p>
                        <p className="truncate text-[#5a6672]">{p.description}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <SensitiveBadge sensitive={p.is_sensitive} />
                        {!role.is_system ? (
                          <button
                            type="button"
                            disabled={isMutating}
                            onClick={() =>
                              void runSensitiveMutation(
                                "Pencabutan izin dari peran memerlukan Step-Up MFA.",
                                async () => {
                                  await submitRevokeRolePermission(
                                    role._id,
                                    p._id,
                                    csrfToken,
                                  );
                                },
                                `Izin ${p.code} dicabut`,
                              )
                            }
                            className={`${styles.btnDanger} ${styles.btnSmall}`}
                          >
                            Cabut
                          </button>
                        ) : null}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* Pemasang Izin dari Katalog (khusus peran non-sistem) */}
            {!role.is_system ? (
              <div className="space-y-3 border-t border-[#dde3e8] pt-4">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[#0b2439]">
                    Tambah Izin dari Katalog ({unattachedFiltered.length} tersedia)
                  </h3>
                  {unattachedFiltered.length > 0 ? (
                    <button
                      type="button"
                      disabled={isMutating}
                      onClick={() => void handleAttachAllFiltered()}
                      className={`${styles.btnPrimary} ${styles.btnSmall}`}
                    >
                      <IdentityIcon name="plus" size={13} />
                      <span>Pasang Semua ({unattachedFiltered.length})</span>
                    </button>
                  ) : null}
                </div>

                {/* Filter Domain + Pencarian */}
                <div className="flex flex-wrap gap-1.5">
                  <button
                    type="button"
                    onClick={() => setPermDomainFilter("ALL")}
                    className={`rounded-md border px-2 py-1 text-[11px] font-semibold ${
                      permDomainFilter === "ALL"
                        ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                        : "border-[#dde3e8] bg-white text-[#5a6672] hover:text-[#0b2439]"
                    }`}
                  >
                    Semua ({catalogPerms.length})
                  </button>
                  {catalogDomains.map((d) => (
                    <button
                      key={d}
                      type="button"
                      onClick={() => setPermDomainFilter(d)}
                      className={`rounded-md border px-2 py-1 text-[11px] font-semibold ${styles.monoText} ${
                        permDomainFilter === d
                          ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                          : "border-[#dde3e8] bg-white text-[#5a6672] hover:text-[#0b2439]"
                      }`}
                    >
                      {d}
                    </button>
                  ))}
                </div>

                <input
                  type="text"
                  value={permQuery}
                  onChange={(e) => setPermQuery(e.target.value)}
                  placeholder="Filter kode izin (mis. enterprise_foundation.parties)..."
                  aria-label="Filter kode izin"
                  className={styles.fieldInput}
                />

                <div className="max-h-64 divide-y divide-[#dde3e8] overflow-y-auto rounded-xl border border-[#dde3e8] bg-[#f9fafb]">
                  {filteredCatalog.length === 0 ? (
                    <p className="p-4 text-xs text-[#5a6672]">
                      Tidak ada izin yang cocok dengan filter ini.
                    </p>
                  ) : (
                    filteredCatalog.map((p) => {
                      const alreadyAttached = assignedSet.has(p._id);
                      return (
                        <div
                          key={p.id}
                          className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2.5 text-xs"
                        >
                          <div className="min-w-0">
                            <p className={`truncate font-medium text-[#0b2439] ${styles.monoText}`}>
                              {p.code}
                            </p>
                            <p className="truncate text-[#5a6672]">{p.description}</p>
                          </div>
                          <div className="flex items-center gap-2">
                            <SensitiveBadge sensitive={p.is_sensitive} />
                            {alreadyAttached ? (
                              <span className="rounded bg-[#ecfdf3] px-2 py-1 text-[11px] font-semibold text-[#067647]">
                                Terpasang
                              </span>
                            ) : (
                              <button
                                type="button"
                                disabled={isMutating}
                                onClick={() =>
                                  void runSensitiveMutation(
                                    "Penambahan izin pada peran memerlukan Step-Up MFA.",
                                    async () => {
                                      await submitAssignRolePermission(
                                        role._id,
                                        p._id,
                                        csrfToken,
                                      );
                                    },
                                    `Izin ${p.code} ditambahkan`,
                                  )
                                }
                                className={`${styles.btnPrimary} ${styles.btnSmall}`}
                              >
                                Pasang
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            ) : null}
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={confirmDeleteOpen}
        title="Hapus Peran RBAC?"
        description={`Peran ${role?.name || ""} akan dihapus dari sistem RBAC. Pastikan tidak sedang dipakai oleh akun aktif.`}
        confirmLabel="Hapus Peran"
        isBusy={isMutating}
        onCancel={() => setConfirmDeleteOpen(false)}
        onConfirm={() => {
          if (!role) return;
          void runSensitiveMutation(
            "Menghapus peran RBAC memerlukan verifikasi Step-Up MFA.",
            async () => {
              await submitRoleLifecycleAction("delete", role._id, csrfToken);
            },
            "Peran RBAC dihapus",
            true,
          );
        }}
      />
    </>
  );
}
