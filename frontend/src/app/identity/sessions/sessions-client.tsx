"use client";

import { useCallback, useEffect, useState } from "react";
import styles from "@/components/identity/identity.module.css";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ConfirmModal,
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  TableSkeleton,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  fetchMySessions,
  fetchUserSessions,
  fetchUsersConnection,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import {
  submitRevokeAllUserSessions,
  submitRevokeSession,
} from "@/lib/graphql/iam-dashboard-mutations";
import type { IamUser, RelayCursorVars, SessionConnection } from "@/lib/graphql/iam-types";
import { OperatorSecurityPanel } from "./operator-security-panel";

export default function SessionsClient() {
  const { csrfToken, pushToast, openStepUpModal } = useIdentityWorkspace();

  const [activeOnly, setActiveOnly] = useState(true);
  const [myConn, setMyConn] = useState<SessionConnection | null>(null);
  const [isLoadingMy, setIsLoadingMy] = useState(true);
  const [myError, setMyError] = useState<string | null>(null);

  // Inspektor sesi akun pengguna lain
  const [directoryUsers, setDirectoryUsers] = useState<IamUser[]>([]);
  const [selectedUserId, setSelectedUserId] = useState("");
  const [targetConn, setTargetConn] = useState<SessionConnection | null>(null);
  const [isLoadingTarget, setIsLoadingTarget] = useState(false);

  const [isRevoking, setIsRevoking] = useState(false);
  const [confirmState, setConfirmState] = useState<{
    title: string;
    description: string;
    confirmLabel: string;
    run: () => Promise<void>;
  } | null>(null);

  const loadMySessions = useCallback(
    async (cursorVars: RelayCursorVars = { first: 15 }) => {
      setIsLoadingMy(true);
      setMyError(null);
      try {
        const res = await fetchMySessions({
          active_only: activeOnly,
          ...cursorVars,
        });
        setMyConn(res);
      } catch (err) {
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setMyError(formatIamDashboardError(err));
      } finally {
        setIsLoadingMy(false);
      }
    },
    [activeOnly],
  );

  const loadTargetSessions = useCallback(
    async (userId: string) => {
      if (!userId) return;
      setIsLoadingTarget(true);
      try {
        const res = await fetchUserSessions(userId, activeOnly, 15);
        setTargetConn(res);
      } catch {
        setTargetConn(null);
      } finally {
        setIsLoadingTarget(false);
      }
    },
    [activeOnly],
  );

  useEffect(() => {
    let active = true;
    async function initMySessions() {
      try {
        const res = await fetchMySessions({ active_only: activeOnly, first: 15 });
        if (!active) return;
        setMyConn(res);
        setMyError(null);
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setMyError(formatIamDashboardError(err));
      } finally {
        if (active) setIsLoadingMy(false);
      }
    }
    void initMySessions();
    return () => {
      active = false;
    };
  }, [activeOnly]);

  useEffect(() => {
    let active = true;
    async function initDirectory() {
      try {
        const users = await fetchUsersConnection({ first: 30 });
        if (!active) return;
        setDirectoryUsers(users.nodes);
        setSelectedUserId((prev) => prev || users.nodes[0]?._id || "");
      } catch {
        // Abaikan bila operator hanya memegang akses self tanpa iam.users.read
      }
    }
    void initDirectory();
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!selectedUserId) return;
    let active = true;
    async function initTargetSessions() {
      try {
        const res = await fetchUserSessions(selectedUserId, activeOnly, 15);
        if (!active) return;
        setTargetConn(res);
      } catch {
        if (active) setTargetConn(null);
      } finally {
        if (active) setIsLoadingTarget(false);
      }
    }
    void initTargetSessions();
    return () => {
      active = false;
    };
  }, [selectedUserId, activeOnly]);

  async function runRevoke(
    reason: string,
    op: () => Promise<void>,
    successMsg: string,
  ) {
    setIsRevoking(true);
    try {
      await op();
      pushToast("success", successMsg);
      await loadMySessions({ first: 15 });
      if (selectedUserId) await loadTargetSessions(selectedUserId);
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      if (isStepUpRequiredError(err)) {
        openStepUpModal(reason, async () => {
          await op();
          pushToast("success", successMsg);
          await loadMySessions({ first: 15 });
          if (selectedUserId) await loadTargetSessions(selectedUserId);
        });
        return;
      }
      pushToast("error", "Gagal Mencabut Sesi", formatIamDashboardError(err));
    } finally {
      setIsRevoking(false);
      setConfirmState(null);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 border-b border-[#dde3e8] pb-5 sm:flex-row sm:items-end">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-[#0b2439]">
            Sesi Opaque &amp; Keamanan Akun
          </h1>
          <p className="mt-1 text-sm text-[#5a6672]">
            Pantau sesi aktif (idle 30 menit / absolut 8 jam), cabut akses,
            serta kelola MFA TOTP dan profil operator.
          </p>
        </div>

        <label className="flex min-h-[40px] cursor-pointer items-center gap-2 rounded-lg border border-[#dde3e8] bg-white px-3 text-xs font-medium text-[#0b2439]">
          <input
            type="checkbox"
            checked={activeOnly}
            onChange={(e) => setActiveOnly(e.target.checked)}
            className="h-4 w-4 accent-[#133554]"
          />
          <span>Hanya Sesi Aktif</span>
        </label>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Kolom Kiri (7/12): Sesi Pribadi & Sesi Target User */}
        <div className="space-y-6 lg:col-span-7">
          <section className={`${styles.hairlinePanel} overflow-hidden`}>
            <div className="border-b border-[#dde3e8] px-5 py-4">
              <h2 className="text-sm font-semibold text-[#0b2439]">
                Sesi Saya (`mySessions`)
              </h2>
              <p className="text-xs text-[#5a6672]">
                Daftar sesi pada perangkat yang pernah atau sedang terautentikasi.
              </p>
            </div>

            {myError ? (
              <div className="p-4">
                <ErrorPanel
                  message={myError}
                  onRetry={() => void loadMySessions({ first: 15 })}
                />
              </div>
            ) : isLoadingMy && !myConn ? (
              <TableSkeleton rows={4} cols={4} />
            ) : !myConn || myConn.nodes.length === 0 ? (
              <EmptyPanel
                title="Tidak Ada Sesi Tercatat"
                description="Belum ada entri sesi yang cocok dengan filter status saat ini."
              />
            ) : (
              <>
                <div className={styles.tableWrap}>
                  <table className={styles.dataTable}>
                    <thead>
                      <tr>
                        <th>IP &amp; Perangkat</th>
                        <th>MFA</th>
                        <th>Batas Idle / Absolut</th>
                        <th className="text-right">Aksi</th>
                      </tr>
                    </thead>
                    <tbody>
                      {myConn.nodes.map((s) => (
                        <tr key={s.id}>
                          <td>
                            <p className={`font-semibold text-[#0b2439] ${styles.monoText}`}>
                              {s.ip_address || "-"} · gen {s.generation}
                            </p>
                            <p className="max-w-xs truncate text-xs text-[#5a6672]">
                              {s.user_agent}
                            </p>
                          </td>
                          <td>
                            <span
                              className={`rounded border px-2 py-0.5 text-[11px] font-semibold ${styles.monoText} ${
                                s.mfa_verified
                                  ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                                  : "border-[#dde3e8] bg-[#f4f7f8] text-[#5a6672]"
                              }`}
                            >
                              {s.mfa_verified ? "MFA" : "STD"}
                            </span>
                          </td>
                          <td className="text-xs text-[#5a6672]">
                            <p>Idle: {formatDateTimeID(s.idle_expires_at)}</p>
                            <p>Absolut: {formatDateTimeID(s.absolute_expires_at)}</p>
                          </td>
                          <td className="text-right">
                            {!s.ended_at ? (
                              <button
                                type="button"
                                disabled={isRevoking}
                                onClick={() =>
                                  setConfirmState({
                                    title: "Cabut Sesi Ini?",
                                    description:
                                      "Bila Anda mencabut sesi perangkat saat ini, Anda akan diminta login ulang.",
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
                            ) : (
                              <span className={`text-[11px] text-[#788491] ${styles.monoText}`}>
                                {s.end_reason || "ENDED"}
                              </span>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <RelayPaginationBar
                  totalCount={myConn.total_count}
                  currentCount={myConn.nodes.length}
                  pageInfo={myConn.pageInfo}
                  isLoading={isLoadingMy}
                  onPrev={() =>
                    void loadMySessions({
                      last: 15,
                      before: myConn.pageInfo.startCursor || undefined,
                    })
                  }
                  onNext={() =>
                    void loadMySessions({
                      first: 15,
                      after: myConn.pageInfo.endCursor || undefined,
                    })
                  }
                />
              </>
            )}
          </section>

          {/* Inspektor Sesi Pengguna Lain */}
          {directoryUsers.length > 0 ? (
            <section className={`${styles.hairlinePanel} overflow-hidden`}>
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#dde3e8] px-5 py-4">
                <div>
                  <h2 className="text-sm font-semibold text-[#0b2439]">
                    Inspektor Sesi Pengguna (`userSessions`)
                  </h2>
                  <p className="text-xs text-[#5a6672]">
                    Pantau dan cabut sesi milik akun pengguna lain secara jarak jauh.
                  </p>
                </div>

                {targetConn && targetConn.nodes.some((s) => !s.ended_at) ? (
                  <button
                    type="button"
                    disabled={isRevoking}
                    onClick={() =>
                      setConfirmState({
                        title: "Cabut Semua Sesi Pengguna Terpilih?",
                        description:
                          "Seluruh sesi aktif pada akun yang dipilih akan langsung dihentikan.",
                        confirmLabel: "Cabut Semua",
                        run: async () => {
                          await submitRevokeAllUserSessions(
                            selectedUserId,
                            "ADMIN_REVOKED",
                            csrfToken,
                          );
                        },
                      })
                    }
                    className={`${styles.btnDanger} ${styles.btnSmall}`}
                  >
                    Cabut Semua
                  </button>
                ) : null}
              </div>

              <div className="border-b border-[#dde3e8] bg-[#f9fafb] px-5 py-3">
                <label htmlFor="session-user-select" className="mb-1 block text-xs font-semibold text-[#0b2439]">
                  Pilih Pengguna
                </label>
                <select
                  id="session-user-select"
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(e.target.value)}
                  className={styles.fieldSelect}
                >
                  {directoryUsers.map((u) => (
                    <option key={u.id} value={u._id}>
                      {u.display_name} ({u.kind} · {u.status})
                    </option>
                  ))}
                </select>
              </div>

              {isLoadingTarget ? (
                <TableSkeleton rows={3} cols={3} />
              ) : !targetConn || targetConn.nodes.length === 0 ? (
                <EmptyPanel
                  title="Belum Ada Sesi Tercatat"
                  description="Pengguna yang dipilih tidak memiliki sesi aktif atau riwayat sesi pada filter ini."
                />
              ) : (
                <div className="divide-y divide-[#dde3e8]">
                  {targetConn.nodes.map((s) => (
                    <div
                      key={s.id}
                      className="flex flex-wrap items-center justify-between gap-3 px-5 py-3 text-xs"
                    >
                      <div className="min-w-0">
                        <p className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                          {s.ip_address || "-"} · gen {s.generation}
                        </p>
                        <p className="mt-0.5 truncate text-[#5a6672]">{s.user_agent}</p>
                        <p className="mt-0.5 text-[11px] text-[#788491]">
                          Masuk {formatDateTimeID(s.authenticated_at)}
                        </p>
                      </div>
                      {!s.ended_at ? (
                        <button
                          type="button"
                          disabled={isRevoking}
                          onClick={() =>
                            setConfirmState({
                              title: "Cabut Sesi Pengguna?",
                              description: "Sesi aktif akun ini akan langsung berakhir.",
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
                      ) : (
                        <span className={`text-[11px] text-[#788491] ${styles.monoText}`}>
                          {s.end_reason || "ENDED"}
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </section>
          ) : null}
        </div>

        {/* Kolom Kanan (5/12): Panel Keamanan & MFA Operator */}
        <div className="lg:col-span-5">
          <OperatorSecurityPanel />
        </div>
      </div>

      <ConfirmModal
        open={Boolean(confirmState)}
        title={confirmState?.title || ""}
        description={confirmState?.description || ""}
        confirmLabel={confirmState?.confirmLabel || "Cabut"}
        isBusy={isRevoking}
        onCancel={() => setConfirmState(null)}
        onConfirm={() => {
          if (confirmState) {
            void runRevoke(
              "Pencabutan sesi memerlukan verifikasi ulang TOTP sesi Anda.",
              confirmState.run,
              "Sesi berhasil dicabut",
            );
          }
        }}
      />
    </div>
  );
}
