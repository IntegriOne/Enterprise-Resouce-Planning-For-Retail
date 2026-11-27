"use client";

import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import gsap from "gsap";
import { useGSAP } from "@gsap/react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import {
  ScopeBadge,
  SensitiveBadge,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import { ScopePicker } from "@/components/identity/scope-picker";
import { fetchAllRolePermissions } from "@/lib/graphql/iam-dashboard-queries";
import {
  checkRoleCompatibility,
  extractUniqueDomains,
  groupPermissionsByDomain,
  resolveEffectivePermissions,
} from "@/lib/graphql/iam-role-access-helpers";
import type {
  IamPermission,
  IamRole,
  IamUser,
  ScopeType,
  UserRoleEdge,
} from "@/lib/graphql/iam-types";
import { QuickRoleBuilder } from "./quick-role-builder";

gsap.registerPlugin(useGSAP);

/**
 * Workbench penugasan Peran RBAC & inspeksi Izin Atomik pada Detail Pengguna.
 * Menampilkan alur 2-tahap (Permission -> Role -> User + Scope), kompatibilitas
 * otomatis dengan `user.kind`, pratinjau izin langsung, serta pembuatan peran kustom inline.
 */
export function UserRoleWorkbench({
  user,
  availableRoles,
  directRoles,
  effectiveRoles,
  isMutating,
  onAssignRole,
  onRequestRevokeRole,
  onRolesCatalogChanged,
}: {
  user: IamUser;
  availableRoles: IamRole[];
  directRoles: UserRoleEdge[];
  effectiveRoles: IamRole[];
  isMutating: boolean;
  onAssignRole: (payload: {
    roleId: string;
    scopeType: ScopeType;
    scopeId: string;
    expiresAt: string;
  }) => Promise<boolean>;
  onRequestRevokeRole: (edge: UserRoleEdge) => void;
  onRolesCatalogChanged: () => Promise<void>;
}) {
  const previewContainerRef = useRef<HTMLDivElement | null>(null);

  const [roleSearch, setRoleSearch] = useState("");
  const [domainFilter, setDomainFilter] = useState<string>("ALL");
  const [showQuickBuilder, setShowQuickBuilder] = useState(false);

  const [selectedRoleId, setSelectedRoleId] = useState<string>("");
  const [scopeType, setScopeType] = useState<ScopeType>("GLOBAL");
  const [scopeId, setScopeId] = useState("");
  const [expiresAt, setExpiresAt] = useState("");

  // Cache lokal per-sesi drawer untuk izin tiap peran agar tidak memicu query berulang.
  const [rolePermsCache, setRolePermsCache] = useState<
    Record<string, IamPermission[]>
  >({});
  const [loadingRoleIds, setLoadingRoleIds] = useState<Record<string, boolean>>(
    {},
  );
  const [rolePermErrors, setRolePermErrors] = useState<Record<string, string>>(
    {},
  );
  const [expandedAssignedRoleId, setExpandedAssignedRoleId] = useState<
    string | null
  >(null);
  const requestedRoleIdsRef = useRef<Set<string>>(new Set());

  const evaluatedRoles = useMemo(() => {
    return availableRoles
      .map((role) => ({
        role,
        compatibility: checkRoleCompatibility(role.code, user.kind),
      }))
      .sort((a, b) => {
        if (a.compatibility.compatible !== b.compatibility.compatible) {
          return a.compatibility.compatible ? -1 : 1;
        }
        return a.role.redirect_priority - b.role.redirect_priority;
      });
  }, [availableRoles, user.kind]);

  const domainChips = useMemo(
    () => extractUniqueDomains([], availableRoles),
    [availableRoles],
  );

  const filteredRoles = useMemo(() => {
    const q = roleSearch.trim().toLowerCase();
    return evaluatedRoles.filter(({ role }) => {
      const matchesDomain =
        domainFilter === "ALL" || role.management_domain === domainFilter;
      const matchesQuery =
        !q ||
        role.name.toLowerCase().includes(q) ||
        role.code.toLowerCase().includes(q) ||
        role.management_domain.toLowerCase().includes(q);
      return matchesDomain && matchesQuery;
    });
  }, [evaluatedRoles, domainFilter, roleSearch]);

  // Pastikan peran terpilih awal selalu mengarah ke peran pertama yang kompatibel.
  const effectiveSelectedRoleId = useMemo(() => {
    if (
      selectedRoleId &&
      availableRoles.some((r) => r._id === selectedRoleId)
    ) {
      return selectedRoleId;
    }
    const firstCompatible = evaluatedRoles.find(
      (item) => item.compatibility.compatible,
    );
    return firstCompatible?.role._id ?? availableRoles[0]?._id ?? "";
  }, [selectedRoleId, availableRoles, evaluatedRoles]);

  const selectedEntry = useMemo(
    () =>
      evaluatedRoles.find(
        (item) => item.role._id === effectiveSelectedRoleId,
      ) ?? null,
    [evaluatedRoles, effectiveSelectedRoleId],
  );

  // Sinkronisasi otomatis UUID target saat memilih scope SELF.
  const effectiveScopeId =
    scopeType === "SELF" ? user._id : scopeType === "GLOBAL" ? "" : scopeId;

  const ensureRolePermissionsLoaded = async (roleId: string, force = false) => {
    if (!roleId) return;
    if (!force && (rolePermsCache[roleId] || loadingRoleIds[roleId])) return;
    requestedRoleIdsRef.current.add(roleId);
    setLoadingRoleIds((prev) => ({ ...prev, [roleId]: true }));
    setRolePermErrors((prev) => {
      const next = { ...prev };
      delete next[roleId];
      return next;
    });
    try {
      const perms = await fetchAllRolePermissions(roleId);
      setRolePermsCache((prev) => ({ ...prev, [roleId]: perms }));
    } catch (err) {
      setRolePermErrors((prev) => ({
        ...prev,
        [roleId]:
          err instanceof Error ? err.message : "Gagal memuat izin peran.",
      }));
    } finally {
      setLoadingRoleIds((prev) => ({ ...prev, [roleId]: false }));
    }
  };

  // Muat pratinjau izin peran yang sedang dipilih dan seluruh peran efektif akun.
  useEffect(() => {
    let active = true;
    const targetIds = new Set<string>();
    if (effectiveSelectedRoleId) targetIds.add(effectiveSelectedRoleId);
    for (const r of effectiveRoles) {
      if (r._id) targetIds.add(r._id);
    }

    const missing = Array.from(targetIds).filter(
      (id) => !requestedRoleIdsRef.current.has(id),
    );
    if (missing.length === 0) return;

    for (const id of missing) {
      requestedRoleIdsRef.current.add(id);
    }

    async function preload() {
      setLoadingRoleIds((prev) => {
        const next = { ...prev };
        for (const id of missing) next[id] = true;
        return next;
      });
      await Promise.all(
        missing.map(async (id) => {
          try {
            const perms = await fetchAllRolePermissions(id);
            if (!active) return;
            setRolePermsCache((prev) => ({ ...prev, [id]: perms }));
          } catch (err) {
            if (!active) return;
            setRolePermErrors((prev) => ({
              ...prev,
              [id]:
                err instanceof Error ? err.message : "Gagal memuat izin peran.",
            }));
          } finally {
            if (active) {
              setLoadingRoleIds((prev) => ({ ...prev, [id]: false }));
            }
          }
        }),
      );
    }

    void preload();
    return () => {
      active = false;
    };
  }, [effectiveSelectedRoleId, effectiveRoles]);

  // Transisi lembut GPU-only saat mengganti pratinjau peran terpilih.
  useGSAP(
    () => {
      if (!previewContainerRef.current) return;
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.fromTo(
          previewContainerRef.current,
          { opacity: 0.65, y: 6 },
          { opacity: 1, y: 0, duration: 0.2, ease: "power2.out", clearProps: "transform" },
        );
      });
      return () => mm.revert();
    },
    { scope: previewContainerRef, dependencies: [effectiveSelectedRoleId] },
  );

  const selectedRolePerms = effectiveSelectedRoleId
    ? rolePermsCache[effectiveSelectedRoleId]
    : undefined;
  const selectedRoleGroupedPerms = useMemo(
    () => groupPermissionsByDomain(selectedRolePerms ?? []),
    [selectedRolePerms],
  );

  const effectivePermissions = useMemo(
    () => resolveEffectivePermissions(effectiveRoles, rolePermsCache),
    [effectiveRoles, rolePermsCache],
  );

  const effectivePermissionsByDomain = useMemo(() => {
    const map = new Map<string, typeof effectivePermissions>();
    for (const item of effectivePermissions) {
      const list = map.get(item.domain);
      if (list) list.push(item);
      else map.set(item.domain, [item]);
    }
    return Array.from(map.entries()).sort(([a], [b]) => a.localeCompare(b));
  }, [effectivePermissions]);

  const directAssignmentRoleIds = useMemo(
    () => new Set(directRoles.map((edge) => edge.node._id)),
    [directRoles],
  );

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!selectedEntry || !selectedEntry.compatibility.compatible || isMutating) {
      return;
    }
    const ok = await onAssignRole({
      roleId: selectedEntry.role._id,
      scopeType,
      scopeId: effectiveScopeId,
      expiresAt,
    });
    if (ok) {
      setScopeId("");
      setExpiresAt("");
    }
  }

  return (
    <div className="space-y-6">
      {/* Penjelasan Arsitektur 2-Tahap IAM */}
      <div className="rounded-xl border border-[#dde3e8] bg-[#f4f7f8] p-4 text-xs">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="flex flex-wrap items-center gap-1.5 font-semibold text-[#0b2439]">
            <span className={`rounded border border-[#dde3e8] bg-white px-2 py-0.5 ${styles.monoText}`}>
              1. Izin Atomik
            </span>
            <IdentityIcon name="chevronRight" size={13} className="text-[#5a6672]" />
            <span className={`rounded border border-[#dde3e8] bg-white px-2 py-0.5 ${styles.monoText}`}>
              2. Peran RBAC (Role)
            </span>
            <IdentityIcon name="chevronRight" size={13} className="text-[#5a6672]" />
            <span className={`rounded border border-[#133554]/25 bg-[#e9eef2] px-2 py-0.5 text-[#133554] ${styles.monoText}`}>
              3. Akun {user.kind} + Scope
            </span>
          </div>

          <div className="flex flex-wrap gap-1.5">
            <button
              type="button"
              onClick={() => setShowQuickBuilder((prev) => !prev)}
              className={`${styles.btnPrimary} ${styles.btnSmall}`}
            >
              <IdentityIcon name="plus" size={13} />
              <span>{showQuickBuilder ? "Tutup Builder" : "Buat Peran Kustom"}</span>
            </button>
            <Link
              href="/identity/permissions"
              className={`${styles.btnSecondary} ${styles.btnSmall}`}
            >
              <span>Katalog Izin</span>
            </Link>
          </div>
        </div>

        <p className="mt-2.5 leading-relaxed text-[#5a6672]">
          {user.kind === "SYSTEM" ? (
            <>
              Akun <strong className="text-[#0b2439]">SYSTEM</strong> hanya dapat menerima peran
              layanan berakhiran <code className={styles.monoText}>*_automation</code> (seperti{" "}
              <code className={styles.monoText}>iam_automation</code> atau{" "}
              <code className={styles.monoText}>ef_automation</code>).
            </>
          ) : (
            <>
              Izin bisnis tidak dipasang langsung ke pengguna, melainkan dibundel dalam{" "}
              <strong className="text-[#0b2439]">Peran (Role)</strong>. Peran{" "}
              <code className={styles.monoText}>ef_automation</code> hanya untuk akun{" "}
              <code className={styles.monoText}>SYSTEM</code>. Untuk memberi akses{" "}
              <code className={styles.monoText}>enterprise_foundation.*</code> ke akun{" "}
              <strong className="text-[#0b2439]">{user.kind}</strong>, pilih{" "}
              {user.kind === "EMPLOYEE" ? (
                <>
                  <code className={styles.monoText}>superadmin</code> atau{" "}
                </>
              ) : null}
              buat peran kustom via tombol <strong className="text-[#0b2439]">Buat Peran Kustom</strong>{" "}
              di atas.
            </>
          )}
        </p>
      </div>

      {/* Builder Peran Kustom Inline */}
      {showQuickBuilder ? (
        <QuickRoleBuilder
          defaultDomain={
            domainFilter !== "ALL" ? domainFilter : "enterprise_foundation"
          }
          onCancel={() => setShowQuickBuilder(false)}
          onCreatedAndPopulated={(createdRole, attachedPerms) => {
            setRolePermsCache((prev) => ({
              ...prev,
              [createdRole._id]: attachedPerms,
            }));
            setSelectedRoleId(createdRole._id);
            setShowQuickBuilder(false);
            void onRolesCatalogChanged();
          }}
        />
      ) : null}

      {/* Form Penetapan Peran dengan Pemilih Visual + Pratinjau Izin */}
      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="space-y-4 rounded-xl border border-[#dde3e8] bg-white p-4"
      >
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-[#dde3e8] pb-3">
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[#0b2439]">
              Tetapkan Peran ke {user.display_name} (`assignUserRole`)
            </h3>
            <p className="mt-0.5 text-[11px] text-[#5a6672]">
              Pilih peran yang kompatibel dengan kategori {user.kind}, tinjau izinnya, lalu atur scope.
            </p>
          </div>
          <Link
            href="/identity/roles"
            className="text-xs font-semibold text-[#133554] underline-offset-4 hover:underline"
          >
            Buka Halaman Peran
          </Link>
        </div>

        {/* Filter Domain & Pencarian Role */}
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-1.5">
            <button
              type="button"
              onClick={() => setDomainFilter("ALL")}
              className={`rounded-md border px-2.5 py-1 text-[11px] font-semibold transition-colors ${
                domainFilter === "ALL"
                  ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                  : "border-[#dde3e8] bg-[#f9fafb] text-[#5a6672] hover:text-[#0b2439]"
              }`}
            >
              Semua ({availableRoles.length})
            </button>
            {domainChips.map((d) => {
              const count = availableRoles.filter(
                (r) => r.management_domain === d,
              ).length;
              return (
                <button
                  key={d}
                  type="button"
                  onClick={() => setDomainFilter(d)}
                  className={`rounded-md border px-2.5 py-1 text-[11px] font-semibold transition-colors ${styles.monoText} ${
                    domainFilter === d
                      ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                      : "border-[#dde3e8] bg-[#f9fafb] text-[#5a6672] hover:text-[#0b2439]"
                  }`}
                >
                  {d} ({count})
                </button>
              );
            })}
          </div>

          <input
            type="text"
            value={roleSearch}
            onChange={(e) => setRoleSearch(e.target.value)}
            placeholder="Cari nama atau kode peran (mis. superadmin, ef_operator, iam_operator)..."
            aria-label="Cari peran RBAC"
            className={styles.fieldInput}
          />
        </div>

        {/* Daftar Pilihan Peran + Panel Pratinjau Izin */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-12">
          <div className="space-y-1.5 lg:col-span-6">
            <span className="block text-[11px] font-semibold uppercase tracking-wider text-[#5a6672]">
              Daftar Peran Aktif ({filteredRoles.length})
            </span>
            <div className="max-h-72 divide-y divide-[#dde3e8] overflow-y-auto rounded-xl border border-[#dde3e8]">
              {filteredRoles.length === 0 ? (
                <div className="space-y-2 p-4 text-xs text-[#5a6672]">
                  <p>Tidak ada peran yang cocok pada filter ini.</p>
                  <button
                    type="button"
                    onClick={() => setShowQuickBuilder(true)}
                    className={`${styles.btnSecondary} ${styles.btnSmall}`}
                  >
                    Buat Peran Kustom Sekarang
                  </button>
                </div>
              ) : (
                filteredRoles.map(({ role, compatibility }) => {
                  const isSelected = role._id === effectiveSelectedRoleId;
                  const alreadyAttached = directAssignmentRoleIds.has(role._id);
                  return (
                    <button
                      key={role.id}
                      type="button"
                      onClick={() => setSelectedRoleId(role._id)}
                      className={`flex w-full flex-col gap-1 px-3 py-2.5 text-left transition-colors ${
                        isSelected
                          ? "bg-[#e9eef2]"
                          : compatibility.compatible
                            ? "bg-white hover:bg-[#f9fafb]"
                            : "bg-[#f9fafb]/70 opacity-70 hover:bg-[#f4f7f8]"
                      }`}
                    >
                      <div className="flex flex-wrap items-center justify-between gap-1.5">
                        <span className="text-xs font-semibold text-[#0b2439]">
                          {role.name}
                        </span>
                        <div className="flex items-center gap-1">
                          {alreadyAttached ? (
                            <span className="rounded bg-[#ecfdf3] px-1.5 py-0.5 text-[10px] font-semibold text-[#067647]">
                              Terpasang
                            </span>
                          ) : null}
                          <span
                            className={`rounded border px-1.5 py-0.5 text-[10px] font-semibold ${styles.monoText} ${
                              compatibility.compatible
                                ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                                : "border-[#fecdca] bg-[#fef3f2] text-[#b42318]"
                            }`}
                          >
                            {compatibility.compatible
                              ? "KOMPATIBEL"
                              : compatibility.isSystemAutomation
                                ? "KHUSUS SYSTEM"
                                : "TIDAK KOMPATIBEL"}
                          </span>
                        </div>
                      </div>
                      <div className={`flex flex-wrap items-center gap-1.5 text-[11px] text-[#5a6672] ${styles.monoText}`}>
                        <span>{role.code}</span>
                        <span>·</span>
                        <span className="text-[#133554]">{role.management_domain || "global"}</span>
                        {role.is_system ? (
                          <>
                            <span>·</span>
                            <span>SYSTEM</span>
                          </>
                        ) : (
                          <>
                            <span>·</span>
                            <span>KUSTOM</span>
                          </>
                        )}
                      </div>
                    </button>
                  );
                })
              )}
            </div>
          </div>

          {/* Pratinjau Izin Atomik Peran Terpilih */}
          <div
            ref={previewContainerRef}
            className="flex flex-col rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-3 lg:col-span-6"
          >
            {selectedEntry ? (
              <>
                <div className="flex flex-wrap items-start justify-between gap-2 border-b border-[#dde3e8] pb-2.5">
                  <div className="min-w-0">
                    <p className={`text-[10px] font-semibold uppercase tracking-wider text-[#5a6672] ${styles.monoText}`}>
                      PRATINJAU IZIN PERAN
                    </p>
                    <p className="truncate text-xs font-bold text-[#0b2439]">
                      {selectedEntry.role.name}
                    </p>
                    <p className={`truncate text-[11px] text-[#5a6672] ${styles.monoText}`}>
                      {selectedEntry.role.code} · redirect:{" "}
                      {selectedEntry.role.default_redirect_path || "(none)"}
                    </p>
                  </div>
                  {!selectedEntry.role.is_system ? (
                    <Link
                      href={`/identity/roles?roleId=${encodeURIComponent(selectedEntry.role._id)}`}
                      className={`${styles.btnSecondary} ${styles.btnSmall} shrink-0`}
                    >
                      Edit Izin
                    </Link>
                  ) : null}
                </div>

                {!selectedEntry.compatibility.compatible ? (
                  <p className="mt-2.5 rounded-lg border border-[#fecdca] bg-[#fef3f2] px-3 py-2 text-xs text-[#b42318]">
                    {selectedEntry.compatibility.reason}
                  </p>
                ) : null}

                <div className="mt-2.5 flex-1 overflow-y-auto max-h-56 space-y-2.5">
                  {loadingRoleIds[selectedEntry.role._id] && !selectedRolePerms ? (
                    <div className="space-y-1.5">
                      <div className={`${styles.skeletonShimmer} h-7 w-full`} />
                      <div className={`${styles.skeletonShimmer} h-7 w-full`} />
                      <div className={`${styles.skeletonShimmer} h-7 w-full`} />
                    </div>
                  ) : rolePermErrors[selectedEntry.role._id] ? (
                    <div className="space-y-2 rounded-lg border border-[#fecdca] bg-[#fef3f2] p-3 text-xs text-[#b42318]">
                      <p>{rolePermErrors[selectedEntry.role._id]}</p>
                      <button
                        type="button"
                        onClick={() =>
                          void ensureRolePermissionsLoaded(
                            selectedEntry.role._id,
                            true,
                          )
                        }
                        className={`${styles.btnSecondary} ${styles.btnSmall}`}
                      >
                        Coba Lagi
                      </button>
                    </div>
                  ) : !selectedRolePerms || selectedRolePerms.length === 0 ? (
                    <div className="space-y-2 rounded-lg border border-dashed border-[#c5cfd8] bg-white p-3 text-xs text-[#5a6672]">
                      <p>Peran ini belum memiliki izin atomik yang terpasang.</p>
                      {!selectedEntry.role.is_system ? (
                        <Link
                          href={`/identity/roles?roleId=${encodeURIComponent(selectedEntry.role._id)}`}
                          className="inline-block font-semibold text-[#133554] underline"
                        >
                          Pasang izin di halaman Peran
                        </Link>
                      ) : null}
                    </div>
                  ) : (
                    selectedRoleGroupedPerms.map((group) => (
                      <div
                        key={group.domain}
                        className="rounded-lg border border-[#dde3e8] bg-white"
                      >
                        <div className="flex items-center justify-between border-b border-[#dde3e8] bg-[#f4f7f8] px-2.5 py-1.5 text-[11px]">
                          <span className={`font-semibold text-[#0b2439] ${styles.monoText}`}>
                            {group.domain} ({group.permissions.length})
                          </span>
                          {group.sensitiveCount > 0 ? (
                            <span className={`text-[10px] font-semibold text-[#b42318] ${styles.monoText}`}>
                              {group.sensitiveCount} sensitif
                            </span>
                          ) : null}
                        </div>
                        <ul className="divide-y divide-[#dde3e8]">
                          {group.permissions.map((perm) => (
                            <li
                              key={perm.id}
                              className="flex items-center justify-between gap-2 px-2.5 py-1.5 text-[11px]"
                            >
                              <div className="min-w-0">
                                <p className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                                  {perm.code}
                                </p>
                                <p className="truncate text-[#5a6672]">
                                  {perm.description}
                                </p>
                              </div>
                              <SensitiveBadge sensitive={perm.is_sensitive} />
                            </li>
                          ))}
                        </ul>
                      </div>
                    ))
                  )}
                </div>
              </>
            ) : (
              <p className="p-4 text-xs text-[#5a6672]">
                Pilih salah satu peran di samping untuk melihat daftar izin atomiknya.
              </p>
            )}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-3 lg:grid-cols-12">
          <div className="lg:col-span-8">
            <ScopePicker
              idPrefix="user-role-scope"
              scopeType={scopeType}
              scopeId={effectiveScopeId}
              selfUserId={user._id}
              allowSelf
              onScopeTypeChange={(next) => {
                setScopeType(next);
                if (next === "SELF") {
                  setScopeId(user._id);
                } else if (next === "GLOBAL") {
                  setScopeId("");
                }
              }}
              onScopeIdChange={setScopeId}
            />
          </div>

          <div className="flex flex-col justify-between gap-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-3.5 lg:col-span-4">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="assign-role-expiry" className="text-xs font-semibold text-[#0b2439]">
                Kedaluwarsa (Opsional)
              </label>
              <input
                id="assign-role-expiry"
                type="datetime-local"
                value={expiresAt}
                onChange={(e) => setExpiresAt(e.target.value)}
                className={styles.fieldInput}
              />
              <span className="text-[11px] text-[#5a6672]">
                Kosongkan untuk penugasan permanen hingga dicabut.
              </span>
            </div>

            <button
              type="submit"
              disabled={
                isMutating ||
                !selectedEntry ||
                !selectedEntry.compatibility.compatible ||
                (scopeType !== "GLOBAL" && !effectiveScopeId.trim())
              }
              className={`${styles.btnPrimary} w-full`}
            >
              {isMutating
                ? "Menyimpan..."
                : selectedEntry
                  ? `Tetapkan ${selectedEntry.role.code}`
                  : "Tetapkan Peran"}
            </button>
          </div>
        </div>
      </form>

      {/* Daftar Penugasan Langsung dengan Expand Izin */}
      <div className="space-y-2">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
          Penugasan Peran Langsung ({directRoles.length})
        </h3>
        <div className="divide-y divide-[#dde3e8] rounded-xl border border-[#dde3e8] bg-white">
          {directRoles.length === 0 ? (
            <p className="p-4 text-xs text-[#5a6672]">
              Belum ada peran yang dipasang langsung ke pengguna ini.
            </p>
          ) : (
            directRoles.map((edge) => {
              const isExpanded = expandedAssignedRoleId === edge.assignment_id;
              const rolePerms = rolePermsCache[edge.node._id] ?? [];
              return (
                <div key={edge.assignment_id} className="p-3.5 text-xs">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="space-y-1">
                      <p className="font-semibold text-[#0b2439]">
                        {edge.node.name}{" "}
                        <span className={`font-normal text-[#5a6672] ${styles.monoText}`}>
                          ({edge.node.code})
                        </span>
                      </p>
                      <div className="flex flex-wrap items-center gap-2">
                        <ScopeBadge scope={edge.scope} />
                        <span className={`rounded border border-[#dde3e8] bg-[#f9fafb] px-2 py-0.5 text-[11px] text-[#133554] ${styles.monoText}`}>
                          {edge.node.management_domain || "global"}
                        </span>
                        {edge.expires_at ? (
                          <span className="text-[#b54708]">
                            Berlaku s/d {formatDateTimeID(edge.expires_at)}
                          </span>
                        ) : null}
                      </div>
                    </div>
                    <div className="flex flex-wrap items-center gap-1.5">
                      <button
                        type="button"
                        onClick={() => {
                          if (!isExpanded) {
                            void ensureRolePermissionsLoaded(edge.node._id);
                          }
                          setExpandedAssignedRoleId(
                            isExpanded ? null : edge.assignment_id,
                          );
                        }}
                        className={`${styles.btnSecondary} ${styles.btnSmall}`}
                      >
                        {isExpanded
                          ? "Tutup Izin"
                          : `Lihat Izin (${rolePerms.length || "..."})`}
                      </button>
                      <button
                        type="button"
                        disabled={isMutating}
                        onClick={() => onRequestRevokeRole(edge)}
                        className={`${styles.btnDanger} ${styles.btnSmall}`}
                      >
                        Cabut
                      </button>
                    </div>
                  </div>

                  {isExpanded ? (
                    <div className="mt-3 rounded-lg border border-[#dde3e8] bg-[#f9fafb] p-2.5">
                      {loadingRoleIds[edge.node._id] && !rolePermsCache[edge.node._id] ? (
                        <div className={`${styles.skeletonShimmer} h-8 w-full`} />
                      ) : rolePerms.length === 0 ? (
                        <p className="text-[11px] text-[#5a6672]">
                          Tidak ada izin atomik terpasang pada peran ini.
                        </p>
                      ) : (
                        <div className="flex flex-wrap gap-1.5">
                          {rolePerms.map((p) => (
                            <span
                              key={p.id}
                              className={`inline-flex items-center gap-1 rounded border px-2 py-0.5 text-[11px] ${styles.monoText} ${
                                p.is_sensitive
                                  ? "border-[#fecdca] bg-[#fef3f2] text-[#b42318]"
                                  : "border-[#dde3e8] bg-white text-[#0b2439]"
                              }`}
                            >
                              {p.code}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  ) : null}
                </div>
              );
            })
          )}
        </div>
      </div>

      {/* Ringkasan Peran Efektif & Matriks Izin Efektif */}
      <div className="space-y-3 border-t border-[#dde3e8] pt-4">
        <div className="space-y-2">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
            Peran Efektif Termasuk Pewarisan Grup ({effectiveRoles.length})
          </h3>
          <div className="flex flex-wrap gap-2">
            {effectiveRoles.length === 0 ? (
              <p className="text-xs text-[#5a6672]">
                Belum memiliki peran efektif aktif.
              </p>
            ) : (
              effectiveRoles.map((r) => (
                <Link
                  key={r.id}
                  href={`/identity/roles?roleId=${encodeURIComponent(r._id)}`}
                  className="rounded-lg border border-[#dde3e8] bg-[#f4f7f8] px-2.5 py-1 text-xs font-medium text-[#0b2439] hover:border-[#133554]"
                >
                  {r.name}{" "}
                  <span className={`text-[#5a6672] ${styles.monoText}`}>
                    {r.code}
                  </span>
                </Link>
              ))
            )}
          </div>
        </div>

        <div className="space-y-2">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
            Izin Atomik Efektif Akun ({effectivePermissions.length})
          </h3>
          {effectivePermissionsByDomain.length === 0 ? (
            <p className="rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-3 text-xs text-[#5a6672]">
              Belum ada izin efektif yang terdeteksi dari peran pengguna ini.
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {effectivePermissionsByDomain.map(([domain, items]) => (
                <div
                  key={domain}
                  className="overflow-hidden rounded-xl border border-[#dde3e8] bg-white"
                >
                  <div className="border-b border-[#dde3e8] bg-[#f4f7f8] px-3 py-2 text-[11px] font-semibold text-[#0b2439]">
                    <span className={styles.monoText}>
                      {domain} ({items.length})
                    </span>
                  </div>
                  <ul className="max-h-44 divide-y divide-[#dde3e8] overflow-y-auto">
                    {items.map((item) => (
                      <li key={item.permission.id} className="px-3 py-2 text-[11px]">
                        <div className="flex items-center justify-between gap-2">
                          <span className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                            {item.permission.code}
                          </span>
                          <SensitiveBadge sensitive={item.permission.is_sensitive} />
                        </div>
                        <p className="mt-0.5 truncate text-[10px] text-[#5a6672]">
                          Dari:{" "}
                          {item.grantedByRoles.map((r) => r.code).join(", ")}
                        </p>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
