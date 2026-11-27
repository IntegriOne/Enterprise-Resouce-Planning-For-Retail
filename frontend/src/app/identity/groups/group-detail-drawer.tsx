"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ActiveBooleanBadge,
  ConfirmModal,
  ErrorPanel,
  ScopeBadge,
  SlideOverDrawer,
  UserKindBadge,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import { ScopePicker } from "@/components/identity/scope-picker";
import {
  fetchAllRoles,
  fetchGroupDetail,
  fetchGroupMembers,
  fetchGroupRoles,
  fetchUsersConnection,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import { isAutomationRoleCode } from "@/lib/graphql/iam-role-access-helpers";
import {
  submitAddGroupMember,
  submitAssignGroupRole,
  submitGroupLifecycleAction,
  submitRemoveGroupMember,
  submitRevokeGroupRole,
} from "@/lib/graphql/iam-dashboard-mutations";
import type {
  GroupAssignableScopeType,
  GroupMemberEdge,
  GroupRoleEdge,
  IamGroup,
  IamRole,
  IamUser,
  ScopeType,
} from "@/lib/graphql/iam-types";

export function GroupDetailDrawer({
  groupId,
  onClose,
  onEditRequested,
  onUpdated,
}: {
  groupId: string | null;
  onClose: () => void;
  onEditRequested: (group: IamGroup) => void;
  onUpdated: () => void;
}) {
  if (!groupId) return null;
  return (
    <GroupDetailInner
      key={groupId}
      groupId={groupId}
      onClose={onClose}
      onEditRequested={onEditRequested}
      onUpdated={onUpdated}
    />
  );
}

function GroupDetailInner({
  groupId,
  onClose,
  onEditRequested,
  onUpdated,
}: {
  groupId: string;
  onClose: () => void;
  onEditRequested: (group: IamGroup) => void;
  onUpdated: () => void;
}) {
  const { csrfToken, pushToast, openStepUpModal } = useIdentityWorkspace();

  const [tab, setTab] = useState<"members" | "roles">("members");
  const [group, setGroup] = useState<IamGroup | null>(null);
  const [members, setMembers] = useState<GroupMemberEdge[]>([]);
  const [roles, setRoles] = useState<GroupRoleEdge[]>([]);
  const [availableUsers, setAvailableUsers] = useState<IamUser[]>([]);
  const [availableRoles, setAvailableRoles] = useState<IamRole[]>([]);
  const [userSearch, setUserSearch] = useState("");
  const [selectedRoleId, setSelectedRoleId] = useState("");
  const [scopeType, setScopeType] = useState<GroupAssignableScopeType>("GLOBAL");
  const [scopeId, setScopeId] = useState("");
  const [expiresAt, setExpiresAt] = useState("");

  const [isLoading, setIsLoading] = useState(true);
  const [isMutating, setIsMutating] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const loadGroupAll = useCallback(async (id: string, qUser = "") => {
    setIsLoading(true);
    setErrorMsg(null);
    try {
      const [detail, memConn, roleConn, usersConn, catRoles] = await Promise.all([
        fetchGroupDetail(id),
        fetchGroupMembers(id, { first: 20 }),
        fetchGroupRoles(id, false, 20),
        fetchUsersConnection({ first: 15, query: qUser || null }),
        fetchAllRoles({ active_only: true }),
      ]);
      const assignableRoles = catRoles.filter((r) => !isAutomationRoleCode(r.code));
      setGroup(detail);
      setMembers(memConn.edges);
      setRoles(roleConn.edges);
      setAvailableUsers(usersConn.nodes);
      setAvailableRoles(assignableRoles);
      setSelectedRoleId((prev) => prev || assignableRoles[0]?._id || "");
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
        const [detail, memConn, roleConn, usersConn, catRoles] = await Promise.all([
          fetchGroupDetail(groupId),
          fetchGroupMembers(groupId, { first: 20 }),
          fetchGroupRoles(groupId, false, 20),
          fetchUsersConnection({ first: 15, query: userSearch || null }),
          fetchAllRoles({ active_only: true }),
        ]);
        if (!active) return;
        const assignableRoles = catRoles.filter((r) => !isAutomationRoleCode(r.code));
        setGroup(detail);
        setMembers(memConn.edges);
        setRoles(roleConn.edges);
        setAvailableUsers(usersConn.nodes);
        setAvailableRoles(assignableRoles);
        setSelectedRoleId((prev) => prev || assignableRoles[0]?._id || "");
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
  }, [groupId, userSearch]);

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
      } else if (groupId) {
        await loadGroupAll(groupId, userSearch);
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
          } else if (groupId) {
            await loadGroupAll(groupId, userSearch);
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

  async function handleAssignRoleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!group || !selectedRoleId || isMutating) return;

    const isoExpiry = expiresAt
      ? new Date(expiresAt).toISOString()
      : undefined;

    await runSensitiveMutation(
      "Menetapkan peran pada grup membutuhkan verifikasi Step-Up MFA.",
      async () => {
        await submitAssignGroupRole(
          {
            group_id: group._id,
            role_id: selectedRoleId,
            scope_type: scopeType,
            scope_id: scopeType === "GLOBAL" ? null : scopeId,
            expires_at: isoExpiry || null,
          },
          csrfToken,
        );
        setScopeId("");
        setExpiresAt("");
      },
      "Peran berhasil dipasang ke grup",
    );
  }

  const memberIds = new Set(members.map((m) => m.node._id));

  return (
    <>
      <SlideOverDrawer
        open={Boolean(groupId)}
        title={group?.name || "Detail Grup"}
        subtitle={group ? `${group.code} · ${group._id}` : undefined}
        onClose={onClose}
      >
        {isLoading && !group ? (
          <div className="space-y-3">
            <div className={`${styles.skeletonShimmer} h-14 w-full`} />
            <div className={`${styles.skeletonShimmer} h-52 w-full`} />
          </div>
        ) : errorMsg ? (
          <ErrorPanel
            message={errorMsg}
            onRetry={() => {
              if (groupId) void loadGroupAll(groupId, userSearch);
            }}
          />
        ) : group ? (
          <div className="space-y-6">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#dde3e8] pb-4">
              <div className="flex items-center gap-2">
                <ActiveBooleanBadge active={group.is_active} />
                <div className="flex rounded-lg border border-[#dde3e8] bg-[#f4f7f8] p-1">
                  <button
                    type="button"
                    onClick={() => setTab("members")}
                    className={`rounded-md px-2.5 py-1 text-xs font-semibold ${
                      tab === "members"
                        ? "bg-white text-[#0b2439] shadow-xs"
                        : "text-[#5a6672]"
                    }`}
                  >
                    Anggota ({members.length})
                  </button>
                  <button
                    type="button"
                    onClick={() => setTab("roles")}
                    className={`rounded-md px-2.5 py-1 text-xs font-semibold ${
                      tab === "roles"
                        ? "bg-white text-[#0b2439] shadow-xs"
                        : "text-[#5a6672]"
                    }`}
                  >
                    Peran ({roles.length})
                  </button>
                </div>
              </div>

              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  disabled={isMutating}
                  onClick={() => onEditRequested(group)}
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
                      "Mengubah status aktif grup memerlukan Step-Up MFA.",
                      async () => {
                        await submitGroupLifecycleAction(
                          group.is_active ? "deactivate" : "activate",
                          group._id,
                          csrfToken,
                        );
                      },
                      group.is_active ? "Grup dinonaktifkan" : "Grup diaktifkan",
                    )
                  }
                  className={`${styles.btnSecondary} ${styles.btnSmall}`}
                >
                  {group.is_active ? "Nonaktifkan" : "Aktifkan"}
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
            </div>

            {tab === "members" ? (
              <div className="space-y-5">
                <div className="space-y-2">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
                    Anggota Saat Ini ({members.length})
                  </h3>
                  <div className="max-h-60 divide-y divide-[#dde3e8] overflow-y-auto rounded-xl border border-[#dde3e8]">
                    {members.length === 0 ? (
                      <p className="p-4 text-xs text-[#5a6672]">
                        Belum ada anggota dalam grup ini.
                      </p>
                    ) : (
                      members.map((edge) => (
                        <div
                          key={edge.node.id}
                          className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2.5 text-xs"
                        >
                          <div className="min-w-0">
                            <p className="truncate font-semibold text-[#0b2439]">
                              {edge.node.display_name}
                            </p>
                            <div className="mt-1 flex items-center gap-1.5">
                              <UserKindBadge kind={edge.node.kind} />
                              <span className={`text-[#5a6672] ${styles.monoText}`}>
                                {edge.node._id}
                              </span>
                            </div>
                          </div>
                          <button
                            type="button"
                            disabled={isMutating}
                            onClick={() =>
                              void runSensitiveMutation(
                                "Mengeluarkan anggota grup memerlukan Step-Up MFA.",
                                async () => {
                                  await submitRemoveGroupMember(
                                    group._id,
                                    edge.node._id,
                                    csrfToken,
                                  );
                                },
                                `${edge.node.display_name} dikeluarkan dari grup`,
                              )
                            }
                            className={`${styles.btnDanger} ${styles.btnSmall}`}
                          >
                            Keluarkan
                          </button>
                        </div>
                      ))
                    )}
                  </div>
                </div>

                <div className="space-y-2 border-t border-[#dde3e8] pt-4">
                  <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                    <h3 className="text-xs font-semibold uppercase tracking-wider text-[#0b2439]">
                      Tambah Anggota dari Direktori
                    </h3>
                    <input
                      type="text"
                      value={userSearch}
                      onChange={(e) => setUserSearch(e.target.value)}
                      placeholder="Cari pengguna..."
                      aria-label="Cari calon anggota grup"
                      className={`${styles.fieldInput} sm:max-w-[240px]`}
                    />
                  </div>
                  <div className="max-h-56 divide-y divide-[#dde3e8] overflow-y-auto rounded-xl border border-[#dde3e8] bg-[#f9fafb]">
                    {availableUsers.map((u) => {
                      const isMember = memberIds.has(u._id);
                      return (
                        <div
                          key={u.id}
                          className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2.5 text-xs"
                        >
                          <div className="min-w-0">
                            <p className="truncate font-semibold text-[#0b2439]">
                              {u.display_name}
                            </p>
                            <p className={`truncate text-[#5a6672] ${styles.monoText}`}>
                              {u.kind} · {u.status}
                            </p>
                          </div>
                          {isMember ? (
                            <span className="rounded bg-[#ecfdf3] px-2 py-1 text-[11px] font-semibold text-[#067647]">
                              Anggota
                            </span>
                          ) : (
                            <button
                              type="button"
                              disabled={isMutating}
                              onClick={() =>
                                void runSensitiveMutation(
                                  "Menambah anggota grup memerlukan Step-Up MFA.",
                                  async () => {
                                    await submitAddGroupMember(
                                      group._id,
                                      u._id,
                                      csrfToken,
                                    );
                                  },
                                  `${u.display_name} ditambahkan ke grup`,
                                )
                              }
                              className={`${styles.btnPrimary} ${styles.btnSmall}`}
                            >
                              Tambah
                            </button>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            ) : (
              <div className="space-y-5">
                <form
                  onSubmit={(e) => void handleAssignRoleSubmit(e)}
                  className="space-y-3 rounded-xl border border-[#dde3e8] bg-white p-4"
                >
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[#0b2439]">
                    Pasang Peran ke Grup (`assignGroupRole`)
                  </h3>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <div className="flex flex-col gap-1.5">
                      <label htmlFor="group-role-select" className="text-xs font-semibold text-[#0b2439]">
                        Pilih Peran
                      </label>
                      <select
                        id="group-role-select"
                        value={selectedRoleId}
                        onChange={(e) => setSelectedRoleId(e.target.value)}
                        className={styles.fieldSelect}
                      >
                        {availableRoles.map((r) => (
                          <option key={r.id} value={r._id}>
                            {r.name} ({r.code})
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <label htmlFor="group-role-expiry" className="text-xs font-semibold text-[#0b2439]">
                        Kedaluwarsa (Opsional)
                      </label>
                      <input
                        id="group-role-expiry"
                        type="datetime-local"
                        value={expiresAt}
                        onChange={(e) => setExpiresAt(e.target.value)}
                        className={styles.fieldInput}
                      />
                    </div>
                  </div>

                  <ScopePicker
                    idPrefix="group-role-scope"
                    scopeType={scopeType}
                    scopeId={scopeId}
                    allowSelf={false}
                    onScopeTypeChange={(next: ScopeType) => {
                      if (next !== "SELF") {
                        setScopeType(next as GroupAssignableScopeType);
                      }
                    }}
                    onScopeIdChange={setScopeId}
                  />

                  <div className="flex justify-end">
                    <button
                      type="submit"
                      disabled={isMutating || !selectedRoleId}
                      className={styles.btnPrimary}
                    >
                      {isMutating ? "Menyimpan..." : "Pasang Peran ke Grup"}
                    </button>
                  </div>
                </form>

                <div className="divide-y divide-[#dde3e8] rounded-xl border border-[#dde3e8]">
                  {roles.length === 0 ? (
                    <p className="p-4 text-xs text-[#5a6672]">
                      Belum ada peran terpasang pada grup ini.
                    </p>
                  ) : (
                    roles.map((edge) => (
                      <div
                        key={edge.assignment_id}
                        className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2.5 text-xs"
                      >
                        <div className="space-y-1">
                          <p className="font-semibold text-[#0b2439]">
                            {edge.node.name}{" "}
                            <span className={`font-normal text-[#5a6672] ${styles.monoText}`}>
                              ({edge.node.code})
                            </span>
                          </p>
                          <div className="flex flex-wrap items-center gap-2">
                            <ScopeBadge scope={edge.scope} />
                            {edge.expires_at ? (
                              <span className="text-[#b54708]">
                                s/d {formatDateTimeID(edge.expires_at)}
                              </span>
                            ) : null}
                          </div>
                        </div>
                        <button
                          type="button"
                          disabled={isMutating}
                          onClick={() =>
                            void runSensitiveMutation(
                              "Pencabutan peran dari grup memerlukan Step-Up MFA.",
                              async () => {
                                await submitRevokeGroupRole(edge.assignment_id, csrfToken);
                              },
                              "Peran dicabut dari grup",
                            )
                          }
                          className={`${styles.btnDanger} ${styles.btnSmall}`}
                        >
                          Cabut
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={confirmDeleteOpen}
        title="Hapus Grup Pengguna?"
        description={`Grup ${group?.name || ""} beserta ikatan keanggotaannya akan dihapus permanen.`}
        confirmLabel="Hapus Grup"
        isBusy={isMutating}
        onCancel={() => setConfirmDeleteOpen(false)}
        onConfirm={() => {
          if (!group) return;
          void runSensitiveMutation(
            "Menghapus grup memerlukan verifikasi Step-Up MFA.",
            async () => {
              await submitGroupLifecycleAction("delete", group._id, csrfToken);
            },
            "Grup berhasil dihapus",
            true,
          );
        }}
      />
    </>
  );
}
