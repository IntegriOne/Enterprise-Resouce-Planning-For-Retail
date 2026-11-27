"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import {
  ActiveBooleanBadge,
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  TableSkeleton,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  fetchGroupsConnection,
  formatIamDashboardError,
  isUnauthenticatedError,
  type GroupsFilterVars,
} from "@/lib/graphql/iam-dashboard-queries";
import type { GroupConnection, IamGroup } from "@/lib/graphql/iam-types";
import { GroupFormModal } from "./group-form-modal";
import { GroupDetailDrawer } from "./group-detail-drawer";

export default function GroupsClient() {
  const [searchInput, setSearchInput] = useState("");
  const [appliedQuery, setAppliedQuery] = useState("");
  const [activeOnly, setActiveOnly] = useState(false);

  const [conn, setConn] = useState<GroupConnection | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [formOpen, setFormOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<IamGroup | null>(null);
  const [activeGroupId, setActiveGroupId] = useState<string | null>(null);

  const loadGroups = useCallback(
    async (cursorVars: Partial<GroupsFilterVars> = { first: 20 }) => {
      setIsLoading(true);
      setErrorMsg(null);
      try {
        const res = await fetchGroupsConnection({
          query: appliedQuery || null,
          active_only: activeOnly,
          ...cursorVars,
        });
        setConn(res);
      } catch (err) {
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setErrorMsg(formatIamDashboardError(err));
      } finally {
        setIsLoading(false);
      }
    },
    [appliedQuery, activeOnly],
  );

  useEffect(() => {
    let active = true;
    async function init() {
      try {
        const res = await fetchGroupsConnection({
          query: appliedQuery || null,
          active_only: activeOnly,
          first: 20,
        });
        if (!active) return;
        setConn(res);
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
  }, [appliedQuery, activeOnly]);

  function handleSearchSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAppliedQuery(searchInput.trim());
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col justify-between gap-4 border-b border-[#dde3e8] pb-5 sm:flex-row sm:items-end">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-[#0b2439]">
            Grup &amp; Pewarisan Peran (`groups`)
          </h1>
          <p className="mt-1 text-sm text-[#5a6672]">
            Kelompokkan pengguna dan tetapkan peran berscope perusahaan atau
            fasilitas tanpa scope `SELF`.
          </p>
        </div>

        <button
          type="button"
          onClick={() => {
            setEditingGroup(null);
            setFormOpen(true);
          }}
          className={styles.btnPrimary}
        >
          <IdentityIcon name="plus" size={16} />
          <span>Buat Grup Baru</span>
        </button>
      </div>

      <form
        onSubmit={handleSearchSubmit}
        className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-white p-4 sm:grid-cols-12"
      >
        <div className="sm:col-span-9">
          <label htmlFor="groups-search" className="mb-1 block text-xs font-semibold text-[#0b2439]">
            Cari Kode atau Nama Grup
          </label>
          <div className="flex gap-2">
            <input
              id="groups-search"
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Contoh: finance_approvers, hrm_staff..."
              className={styles.fieldInput}
            />
            <button type="submit" className={`${styles.btnSecondary} shrink-0`}>
              Cari
            </button>
          </div>
        </div>

        <div className="flex items-end sm:col-span-3">
          <label className="flex min-h-[42px] w-full cursor-pointer items-center gap-2 rounded-lg border border-[#dde3e8] bg-[#f9fafb] px-3 text-xs font-medium text-[#0b2439]">
            <input
              type="checkbox"
              checked={activeOnly}
              onChange={(e) => setActiveOnly(e.target.checked)}
              className="h-4 w-4 accent-[#133554]"
            />
            <span>Hanya Aktif</span>
          </label>
        </div>
      </form>

      {errorMsg ? (
        <ErrorPanel
          message={errorMsg}
          onRetry={() => void loadGroups({ first: 20 })}
        />
      ) : (
        <div className={`${styles.hairlinePanel} overflow-hidden`}>
          {isLoading && !conn ? (
            <TableSkeleton rows={5} cols={4} />
          ) : !conn || conn.nodes.length === 0 ? (
            <EmptyPanel
              title="Belum Ada Grup Ditemukan"
              description="Buat grup pertama atau ubah filter pencarian untuk melihat daftar grup."
            />
          ) : (
            <>
              <div className={styles.tableWrap}>
                <table className={styles.dataTable}>
                  <thead>
                    <tr>
                      <th>Nama &amp; Kode Grup</th>
                      <th>Deskripsi</th>
                      <th>Status</th>
                      <th>Diperbarui</th>
                      <th className="text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {conn.nodes.map((g) => (
                      <tr key={g.id}>
                        <td>
                          <p className="font-semibold text-[#0b2439]">{g.name}</p>
                          <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                            {g.code}
                          </p>
                        </td>
                        <td className="max-w-sm truncate text-xs text-[#5a6672]">
                          {g.description || "-"}
                        </td>
                        <td>
                          <ActiveBooleanBadge active={g.is_active} />
                        </td>
                        <td className="text-xs text-[#5a6672]">
                          {formatDateTimeID(g.updated_at)}
                        </td>
                        <td className="text-right">
                          <button
                            type="button"
                            onClick={() => setActiveGroupId(g._id)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Anggota &amp; Peran
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <RelayPaginationBar
                totalCount={conn.total_count}
                currentCount={conn.nodes.length}
                pageInfo={conn.pageInfo}
                isLoading={isLoading}
                onPrev={() =>
                  void loadGroups({
                    last: 20,
                    before: conn.pageInfo.startCursor || undefined,
                  })
                }
                onNext={() =>
                  void loadGroups({
                    first: 20,
                    after: conn.pageInfo.endCursor || undefined,
                  })
                }
              />
            </>
          )}
        </div>
      )}

      <GroupFormModal
        open={formOpen}
        editingGroup={editingGroup}
        onClose={() => setFormOpen(false)}
        onSaved={(saved) => {
          void loadGroups({ first: 20 });
          if (saved?._id) setActiveGroupId(saved._id);
        }}
      />

      <GroupDetailDrawer
        groupId={activeGroupId}
        onClose={() => setActiveGroupId(null)}
        onEditRequested={(grp) => {
          setEditingGroup(grp);
          setFormOpen(true);
        }}
        onUpdated={() => void loadGroups({ first: 20 })}
      />
    </div>
  );
}
