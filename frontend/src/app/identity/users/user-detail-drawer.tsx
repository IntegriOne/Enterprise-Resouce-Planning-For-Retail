"use client";

import { useCallback, useEffect, useState } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ActiveBooleanBadge,
  ConfirmModal,
  ErrorPanel,
  SlideOverDrawer,
  UserKindBadge,
  UserStatusBadge,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  fetchAllRoles,
  fetchUserDetail,
  fetchUserEffectiveRoles,
  fetchUserGroups,
  fetchUserIdentities,
  fetchUserRoles,
  fetchUserSessions,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import {
  submitAssignUserRole,
  submitRemoveUserIdentity,
  submitRevokeAllUserSessions,
  submitRevokeSession,
  submitRevokeUserRole,
  submitSetPrimaryUserIdentity,
  submitUserLifecycleAction,
} from "@/lib/graphql/iam-dashboard-mutations";
import type {
  IamGroup,
  IamRole,
  IamSession,
  IamUser,
  IamUserIdentity,
  ScopeType,
  UserRoleEdge,
} from "@/lib/graphql/iam-types";
import {
  lookupPartyContactsByIamUserId,
  type PartnerContactIamLink,
} from "@/lib/graphql/ef-scope-lookup";
import { UserRoleWorkbench } from "./user-role-workbench";

export type DetailTab = "overview" | "roles" | "sessions";

export function UserDetailDrawer({
  userId,
  initialTab = "overview",
  onClose,
  onUpdated,
}: {
  userId: string | null;
  initialTab?: DetailTab;
  onClose: () => void;
  onUpdated: () => void;
}) {
  if (!userId) return null;
  return (
    <UserDetailInner
      key={`${userId}-${initialTab}`}
      userId={userId}
      initialTab={initialTab}
      onClose={onClose}
      onUpdated={onUpdated}
    />
  );
}

function UserDetailInner({
  userId,
  initialTab,
  onClose,
  onUpdated,
}: {
  userId: string;
  initialTab: DetailTab;
  onClose: () => void;
  onUpdated: () => void;
}) {
  const { csrfToken, pushToast, openStepUpModal } = useIdentityWorkspace();

  const [tab, setTab] = useState<DetailTab>(initialTab);
  const [user, setUser] = useState<IamUser | null>(null);
  const [identities, setIdentities] = useState<IamUserIdentity[]>([]);
  const [directRoles, setDirectRoles] = useState<UserRoleEdge[]>([]);
  const [effectiveRoles, setEffectiveRoles] = useState<IamRole[]>([]);
  const [groups, setGroups] = useState<IamGroup[]>([]);
  const [sessions, setSessions] = useState<IamSession[]>([]);
  const [availableRoles, setAvailableRoles] = useState<IamRole[]>([]);
  const [partnerContacts, setPartnerContacts] = useState<PartnerContactIamLink[]>([]);
  const [partnerUnavailableReason, setPartnerUnavailableReason] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [isMutating, setIsMutating] = useState(false);

  const [confirmState, setConfirmState] = useState<{
    title: string;
    description: string;
    confirmLabel: string;
    run: () => Promise<void>;
  } | null>(null);

  // Memuat detail IAM dan tautan PartyContact Enterprise Foundation secara terisolasi.
  const applyDrawerPayload = useCallback(async (id: string, isCancelled?: () => boolean) => {
    const [
      detail,
      idConn,
      roleConn,
      effConn,
      grpConn,
      sessConn,
      catalogRoles,
    ] = await Promise.all([
      fetchUserDetail(id),
      fetchUserIdentities(id, 20),
      fetchUserRoles(id, false, 20),
      fetchUserEffectiveRoles(id, 20),
      fetchUserGroups(id, 20),
      fetchUserSessions(id, false, 20),
      fetchAllRoles({ active_only: true }),
    ]);

    let efContacts: PartnerContactIamLink[] = [];
    let efUnavailable: string | null = null;
    if (detail && (detail.kind === "VENDOR" || detail.kind === "SUPPLIER")) {
      const efRes = await lookupPartyContactsByIamUserId(detail._id);
      efContacts = efRes.contacts;
      efUnavailable = efRes.unavailableReason ?? null;
    }

    if (isCancelled?.()) return;

    setUser(detail);
    setIdentities(idConn.nodes);
    setDirectRoles(roleConn.edges);
    setEffectiveRoles(effConn.nodes);
    setGroups(grpConn.nodes);
    setSessions(sessConn.nodes);
    setAvailableRoles(catalogRoles);
    setPartnerContacts(efContacts);
    setPartnerUnavailableReason(efUnavailable);
  }, []);

  const loadUserAll = useCallback(async (id: string) => {
    setIsLoading(true);
    setErrorMsg(null);
    try {
      await applyDrawerPayload(id);
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      setErrorMsg(formatIamDashboardError(err));
    } finally {
      setIsLoading(false);
    }
  }, [applyDrawerPayload]);

  useEffect(() => {
    let active = true;
    async function init() {
      try {
        await applyDrawerPayload(userId, () => !active);
        if (active) setErrorMsg(null);
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
  }, [applyDrawerPayload, userId]);

  async function runSensitiveMutation(
    reason: string,
    op: () => Promise<void>,
    successTitle: string,
  ): Promise<boolean> {
    setIsMutating(true);
    try {
      await op();
      pushToast("success", successTitle);
      if (userId) await loadUserAll(userId);
      onUpdated();
      return true;
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return false;
      }
      if (isStepUpRequiredError(err)) {
        openStepUpModal(reason, async () => {
          await op();
          pushToast("success", successTitle);
          if (userId) await loadUserAll(userId);
          onUpdated();
        });
        return false;
      }
      pushToast("error", "Aksi Gagal", formatIamDashboardError(err));
      return false;
    } finally {
      setIsMutating(false);
      setConfirmState(null);
    }
  }

  async function handleAssignRole(payload: {
    roleId: string;
    scopeType: ScopeType;
    scopeId: string;
    expiresAt: string;
  }): Promise<boolean> {
    if (!user || !payload.roleId || isMutating) return false;
    const isoExpiry = payload.expiresAt
      ? new Date(payload.expiresAt).toISOString()
      : undefined;

    return runSensitiveMutation(
      "Penugasan peran ke akun pengguna memerlukan verifikasi Step-Up MFA.",
      async () => {
        await submitAssignUserRole(
          {
            user_id: user._id,
            role_id: payload.roleId,
            scope_type: payload.scopeType,
            scope_id: payload.scopeType === "GLOBAL" ? null : payload.scopeId,
            expires_at: isoExpiry || null,
          },
          csrfToken,
        );
      },
      "Peran berhasil ditambahkan ke akun pengguna",
    );
  }

  return (
    <>
      <SlideOverDrawer
        open={Boolean(userId)}
        title={user?.display_name || "Detail Pengguna"}
        onClose={onClose}
      >
        {isLoading && !user ? (
          <div className="space-y-3">
            <div className={`${styles.skeletonShimmer} h-14 w-full`} />
            <div className={`${styles.skeletonShimmer} h-44 w-full`} />
          </div>
        ) : errorMsg ? (
          <ErrorPanel
            message={errorMsg}
            onRetry={() => {
              if (userId) void loadUserAll(userId);
            }}
          />
        ) : user ? (
          <div className="space-y-6">
            {/* Bar Status & Tab */}
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#dde3e8] pb-4">
              <div className="flex flex-wrap items-center gap-2">
                <UserKindBadge kind={user.kind} />
                <UserStatusBadge status={user.status} />
                <span className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                  perm v{user.perm_revision}
                </span>
              </div>

              <div className="flex flex-wrap gap-1 rounded-lg border border-[#dde3e8] bg-[#f4f7f8] p-1">
                {(
                  [
                    { id: "overview", label: "Profil & Identitas" },
                    { id: "roles", label: `Peran & Izin (${directRoles.length})` },
                    { id: "sessions", label: `Grup & Sesi (${sessions.length})` },
                  ] as const
                ).map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setTab(item.id)}
                    className={`rounded-md px-2.5 py-1.5 text-xs font-semibold transition-colors ${
                      tab === item.id
                        ? "bg-white text-[#0b2439] shadow-xs"
                        : "text-[#5a6672] hover:text-[#0b2439]"
                    }`}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
            </div>

            {tab === "overview" ? (
              <div className="space-y-6">
                <dl className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-4 text-xs sm:grid-cols-2">
                  <div>
                    <dt className="text-[#788491]">Login Terakhir</dt>
                    <dd className="mt-0.5 font-medium text-[#0b2439]">
                      {formatDateTimeID(user.last_login_at)}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-[#788491]">Locale / Timezone</dt>
                    <dd className={`mt-0.5 font-medium text-[#0b2439] ${styles.monoText}`}>
                      {user.locale} · {user.timezone}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-[#788491]">Dibuat</dt>
                    <dd className="mt-0.5 font-medium text-[#0b2439]">
                      {formatDateTimeID(user.created_at)}
                    </dd>
                  </div>
                </dl>

                {user.kind === "VENDOR" || user.kind === "SUPPLIER" ? (
                  <div className="space-y-2.5">
                    <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                      Tautan Mitra Enterprise Foundation (`party_contacts.iam_user_id`)
                    </h3>
                    {partnerUnavailableReason ? (
                      <p className="rounded-xl border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-xs text-amber-900">
                        {partnerUnavailableReason}
                      </p>
                    ) : partnerContacts.length === 0 ? (
                      <p className="rounded-xl border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-xs text-amber-900">
                        Akun mitra ini belum memiliki kontak penanggung jawab.
                      </p>
                    ) : (
                      <div className="divide-y divide-[#dde3e8] rounded-xl border border-[#dde3e8] bg-[#f9fafb]">
                        {partnerContacts.map((pc) => (
                          <div key={pc.id} className="space-y-1.5 p-3.5 text-xs">
                            <div className="flex flex-wrap items-center justify-between gap-2">
                              <p className="font-semibold text-[#0b2439]">
                                {pc.fullName}
                                {pc.isPrimary ? " (Utama)" : ""}
                              </p>
                              <span
                                className={`rounded px-2 py-0.5 text-[11px] font-semibold ${styles.monoText} ${
                                  pc.portalAccessStatus === "ACTIVE"
                                    ? "bg-emerald-50 text-emerald-800"
                                    : pc.portalAccessStatus === "PENDING_LINK"
                                      ? "bg-sky-50 text-sky-900"
                                      : "bg-slate-100 text-slate-700"
                                }`}
                              >
                                {pc.portalAccessStatus}
                              </span>
                            </div>
                            <p className={`text-[11px] text-[#5a6672] ${styles.monoText}`}>
                              role: {pc.contactRole}
                              {pc.jobTitle ? ` · ${pc.jobTitle}` : ""}
                            </p>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ) : null}

                {/* Kontrol Siklus Hidup */}
                <div className="space-y-2.5">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                    Aksi Siklus Hidup Akun
                  </h3>
                  <div className="flex flex-wrap gap-2">
                    <button
                      type="button"
                      onClick={() => setTab("roles")}
                      className={`${styles.btnSecondary} ${styles.btnSmall}`}
                    >
                      <IdentityIcon name="roles" size={14} />
                      <span>Kelola Peran &amp; Izin ({directRoles.length})</span>
                    </button>

                    {user.status !== "ACTIVE" ? (
                      <button
                        type="button"
                        disabled={isMutating}
                        onClick={() =>
                          setConfirmState({
                            title: "Aktifkan Akun?",
                            description: `Mengubah status ${user.display_name} ke ACTIVE agar dapat diautentikasi.`,
                            confirmLabel: "Aktifkan Akun",
                            run: async () => {
                              await submitUserLifecycleAction("activate", user._id, csrfToken);
                            },
                          })
                        }
                        className={`${styles.btnPrimary} ${styles.btnSmall}`}
                      >
                        Aktifkan
                      </button>
                    ) : null}

                    {user.status !== "SUSPENDED" ? (
                      <button
                        type="button"
                        disabled={isMutating}
                        onClick={() =>
                          setConfirmState({
                            title: "Tangguhkan Akun Sementara?",
                            description: `Menangguhkan ${user.display_name} akan memblokir sesi dan otorisasi saat ini.`,
                            confirmLabel: "Tangguhkan",
                            run: async () => {
                              await submitUserLifecycleAction("suspend", user._id, csrfToken);
                            },
                          })
                        }
                        className={`${styles.btnSecondary} ${styles.btnSmall}`}
                      >
                        Tangguhkan
                      </button>
                    ) : null}

                    {user.status !== "DISABLED" ? (
                      <button
                        type="button"
                        disabled={isMutating}
                        onClick={() =>
                          setConfirmState({
                            title: "Nonaktifkan Akun Permanen?",
                            description: `Akun ${user.display_name} akan dinonaktifkan dan kehilangan akses masuk.`,
                            confirmLabel: "Nonaktifkan",
                            run: async () => {
                              await submitUserLifecycleAction("disable", user._id, csrfToken);
                            },
                          })
                        }
                        className={`${styles.btnDanger} ${styles.btnSmall}`}
                      >
                        Nonaktifkan
                      </button>
                    ) : null}

                    {!user.deleted_at ? (
                      <button
                        type="button"
                        disabled={isMutating}
                        onClick={() =>
                          setConfirmState({
                            title: "Hapus Lunak (Soft Delete) Akun?",
                            description: `Menandai ${user.display_name} sebagai dihapus lunak beserta pencabutan seluruh sesi.`,
                            confirmLabel: "Hapus Lunak",
                            run: async () => {
                              await submitUserLifecycleAction("softDelete", user._id, csrfToken);
                            },
                          })
                        }
                        className={`${styles.btnDanger} ${styles.btnSmall}`}
                      >
                        Hapus Lunak
                      </button>
                    ) : null}
                  </div>
                </div>

                {/* Daftar Identitas Login */}
                <div className="space-y-2.5">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                    Identitas Login Terdaftar ({identities.length})
                  </h3>
                  <div className="divide-y divide-[#dde3e8] rounded-xl border border-[#dde3e8]">
                    {identities.length === 0 ? (
                      <p className="p-4 text-xs text-[#5a6672]">
                        Tidak ada identitas login ditemukan.
                      </p>
                    ) : (
                      identities.map((idItem) => (
                        <div
                          key={idItem.id}
                          className="flex flex-wrap items-center justify-between gap-2 p-3 text-xs"
                        >
                          <div className="min-w-0">
                            <p className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                              {idItem.value}
                            </p>
                            <div className="mt-1 flex flex-wrap items-center gap-1.5 text-[11px] text-[#5a6672]">
                              <span className={styles.monoText}>{idItem.kind}</span>
                              <span>·</span>
                              <ActiveBooleanBadge
                                active={idItem.is_verified}
                                trueText="VERIFIED"
                                falseText="UNVERIFIED"
                              />
                              {idItem.is_primary ? (
                                <span className="rounded bg-[#e9eef2] px-1.5 py-0.5 font-semibold text-[#133554]">
                                  PRIMARY
                                </span>
                              ) : null}
                            </div>
                          </div>

                          <div className="flex items-center gap-1.5">
                            {!idItem.is_primary && idItem.kind === "EMAIL" ? (
                              <button
                                type="button"
                                disabled={isMutating}
                                onClick={() =>
                                  void runSensitiveMutation(
                                    "Mengubah identitas utama memerlukan otorisasi sensitif.",
                                    async () => {
                                      await submitSetPrimaryUserIdentity(
                                        user._id,
                                        idItem._id,
                                        csrfToken,
                                      );
                                    },
                                    "Identitas utama diperbarui",
                                  )
                                }
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Jadikan Utama
                              </button>
                            ) : null}
                            {!idItem.is_primary ? (
                              <button
                                type="button"
                                disabled={isMutating}
                                onClick={() =>
                                  setConfirmState({
                                    title: "Hapus Identitas Login?",
                                    description: `Identitas ${idItem.value} akan dihapus dari akun ini.`,
                                    confirmLabel: "Hapus Identitas",
                                    run: async () => {
                                      await submitRemoveUserIdentity(
                                        user._id,
                                        idItem._id,
                                        csrfToken,
                                      );
                                    },
                                  })
                                }
                                aria-label={`Hapus identitas ${idItem.value}`}
                                className={`${styles.btnDanger} ${styles.btnSmall}`}
                              >
                                <IdentityIcon name="trash" size={13} />
                              </button>
                            ) : null}
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>
            ) : null}

            {tab === "roles" ? (
              <UserRoleWorkbench
                user={user}
                availableRoles={availableRoles}
                directRoles={directRoles}
                effectiveRoles={effectiveRoles}
                isMutating={isMutating}
                onAssignRole={handleAssignRole}
                onRequestRevokeRole={(edge) =>
                  setConfirmState({
                    title: "Cabut Penugasan Peran?",
                    description: `Peran ${edge.node.name} pada scope ${edge.scope.type} akan dicabut dari ${user.display_name}.`,
                    confirmLabel: "Cabut Peran",
                    run: async () => {
                      await submitRevokeUserRole(edge.assignment_id, csrfToken);
                    },
                  })
                }
                onRolesCatalogChanged={async () => {
                  if (userId) await loadUserAll(userId);
                }}
              />
            ) : null}

            {tab === "sessions" ? (
              <div className="space-y-6">
                <div className="space-y-2">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                    Keanggotaan Grup ({groups.length})
                  </h3>
                  {groups.length === 0 ? (
                    <p className="rounded-xl border border-[#dde3e8] p-4 text-xs text-[#5a6672]">
                      Tidak tergabung dalam grup mana pun.
                    </p>
                  ) : (
                    <div className="flex flex-wrap gap-2">
                      {groups.map((g) => (
                        <span
                          key={g.id}
                          className="inline-flex items-center gap-2 rounded-lg border border-[#dde3e8] bg-[#f4f7f8] px-3 py-1.5 text-xs"
                        >
                          <strong className="text-[#0b2439]">{g.name}</strong>
                          <span className={`text-[#5a6672] ${styles.monoText}`}>
                            {g.code}
                          </span>
                        </span>
                      ))}
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                      Riwayat &amp; Sesi Aktif ({sessions.length})
                    </h3>
                    {sessions.some((s) => !s.ended_at) ? (
                      <button
                        type="button"
                        disabled={isMutating}
                        onClick={() =>
                          setConfirmState({
                            title: "Cabut Seluruh Sesi Pengguna?",
                            description: `Semua sesi aktif milik ${user.display_name} akan langsung dihentikan.`,
                            confirmLabel: "Cabut Semua Sesi",
                            run: async () => {
                              await submitRevokeAllUserSessions(
                                user._id,
                                "ADMIN_REVOKED",
                                csrfToken,
                              );
                            },
                          })
                        }
                        className={`${styles.btnDanger} ${styles.btnSmall}`}
                      >
                        Cabut Semua Sesi
                      </button>
                    ) : null}
                  </div>

                  <div className="divide-y divide-[#dde3e8] rounded-xl border border-[#dde3e8]">
                    {sessions.length === 0 ? (
                      <p className="p-4 text-xs text-[#5a6672]">
                        Belum ada sesi tercatat untuk akun ini.
                      </p>
                    ) : (
                      sessions.map((s) => (
                        <div
                          key={s.id}
                          className="flex flex-wrap items-center justify-between gap-2 p-3 text-xs"
                        >
                          <div className="min-w-0 space-y-0.5">
                            <p className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                              {s.ip_address || "-"} · gen {s.generation}{" "}
                              {s.ended_at ? `· ${s.end_reason || "ENDED"}` : "· AKTIF"}
                            </p>
                            <p className="truncate text-[#5a6672]">{s.user_agent}</p>
                            <p className="text-[11px] text-[#788491]">
                              Login {formatDateTimeID(s.authenticated_at)} · Idle{" "}
                              {formatDateTimeID(s.idle_expires_at)}
                            </p>
                          </div>
                          {!s.ended_at ? (
                            <button
                              type="button"
                              disabled={isMutating}
                              onClick={() =>
                                setConfirmState({
                                  title: "Cabut Sesi Ini?",
                                  description: "Pemegang sesi harus masuk ulang setelah sesi dicabut.",
                                  confirmLabel: "Cabut Sesi",
                                  run: async () => {
                                    await submitRevokeSession(
                                      s._id,
                                      "ADMIN_REVOKED",
                                      csrfToken,
                                    );
                                  },
                                })
                              }
                              className={`${styles.btnDanger} ${styles.btnSmall}`}
                            >
                              Cabut
                            </button>
                          ) : null}
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>
            ) : null}
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(confirmState)}
        title={confirmState?.title || ""}
        description={confirmState?.description || ""}
        confirmLabel={confirmState?.confirmLabel || "Konfirmasi"}
        isBusy={isMutating}
        onCancel={() => setConfirmState(null)}
        onConfirm={() => {
          if (confirmState) {
            void runSensitiveMutation(
              "Aksi sensitif ini memerlukan verifikasi ulang TOTP sesi.",
              confirmState.run,
              "Perubahan berhasil diterapkan",
            );
          }
        }}
      />
    </>
  );
}
