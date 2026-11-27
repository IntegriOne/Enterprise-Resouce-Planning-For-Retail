"use client";

import { useCallback, useEffect, useState } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ConfirmModal,
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  TableSkeleton,
} from "@/components/identity/identity-primitives";
import {
  EfCodePill,
  EfStatusBadge,
  EfToolbarCard,
} from "@/components/workspace/ef-ui-primitives";
import {
  PARTY_KIND_OPTIONS,
  PARTY_ROLE_KIND_OPTIONS,
  PARTY_STATUS_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import { deleteParty } from "@/lib/graphql/ef-partner-mutations";
import {
  fetchAllParties,
  fetchParties,
  fetchPartiesByEmail,
  fetchPartiesByIds,
  fetchParty,
  fetchPartyByCode,
  fetchPartyByTaxId,
} from "@/lib/graphql/ef-partner-queries";
import type {
  EfParty,
  PartyKind,
  PartyRoleKind,
  PartyStatus,
} from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { PartyDetailDrawer } from "./party-detail-drawer";
import { PartyFormModal } from "./party-form-modal";

export default function PartnersClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(snapshot, "enterprise_foundation.parties.write");

  const [items, setItems] = useState<EfParty[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [kindFilter, setKindFilter] = useState<PartyKind | "">("");
  const [statusFilter, setStatusFilter] = useState<PartyStatus | "">("");
  const [roleFilter, setRoleFilter] = useState<PartyRoleKind | "">("");
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [allMode, setAllMode] = useState(false);

  const [lookupMode, setLookupMode] = useState<"CODE" | "NPWP" | "EMAIL" | "IDS">("CODE");
  const [lookupValue, setLookupValue] = useState("");

  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "profile" | "tax">("create");
  const [editing, setEditing] = useState<EfParty | null>(null);
  const [selected, setSelected] = useState<EfParty | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<EfParty | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllParties(20, cursor?.after)
        : await fetchParties({
            query: query.trim() || undefined,
            kind: kindFilter || null,
            status: statusFilter || null,
            role: roleFilter || null,
            include_deleted: includeDeleted,
            first: 20,
            after: cursor?.after || null,
            before: cursor?.before || null,
          });
    },
    [allMode, includeDeleted, kindFilter, query, roleFilter, statusFilter],
  );

  const loadList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      setLoading(true);
      setLoadErr(null);
      try {
        const res = await fetchCurrentList(cursor);
        setItems(res.nodes);
        setTotalCount(res.total_count);
        setPageInfo(res.pageInfo);
      } catch (err) {
        setLoadErr(formatEfError(err));
      } finally {
        setLoading(false);
      }
    },
    [fetchCurrentList],
  );

  useEffect(() => {
    let active = true;
    async function initList() {
      try {
        const res = await fetchCurrentList();
        if (!active) return;
        setItems(res.nodes);
        setTotalCount(res.total_count);
        setPageInfo(res.pageInfo);
        setLoadErr(null);
      } catch (err) {
        if (!active) return;
        setLoadErr(formatEfError(err));
      } finally {
        if (active) setLoading(false);
      }
    }
    void initList();
    return () => {
      active = false;
    };
  }, [fetchCurrentList]);

  async function handleExactLookup() {
    const val = lookupValue.trim();
    if (!val) return;
    setLoading(true);
    try {
      if (lookupMode === "CODE") {
        const one = await fetchPartyByCode(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "NPWP") {
        const one = await fetchPartyByTaxId(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "EMAIL") {
        const conn = await fetchPartiesByEmail(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else {
        const ids = val.split(",").map((s) => s.trim()).filter(Boolean);
        const list = await fetchPartiesByIds(ids);
        setItems(list);
        setTotalCount(list.length);
        setPageInfo(null);
      }
    } catch (err) {
      pushToast("error", "Lookup Mitra Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  async function openDetail(p: EfParty) {
    setSelected(p);
    try {
      const full = await fetchParty(p.id);
      if (full) setSelected(full);
    } catch (err) {
      pushToast("error", "Gagal Memuat Detail Mitra", formatEfError(err));
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const ok = await runDashboardAction(
      () => deleteParty(target.id, csrfToken),
      { stepUpReason: "Menghapus mitra bisnis bersifat sensitif." },
    );
    if (ok) {
      setDeleteTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "Mitra Dihapus", target.display_name);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Direktori Mitra (Party Core)"
        description="Pengelolaan Supplier, Vendor CMT, Ekspedisi & Customer B2B (`parties`, `allParties`, profil pajak, & verifikasi)."
        actions={
          <>
            <button
              type="button"
              onClick={() => void loadList()}
              className={styles.btnSecondary}
            >
              <IdentityIcon name="refresh" size={15} />
              <span>Segarkan</span>
            </button>
            {canWrite ? (
              <button
                type="button"
                onClick={() => {
                  setEditing(null);
                  setFormMode("create");
                  setFormOpen(true);
                }}
                className={styles.btnPrimary}
              >
                <IdentityIcon name="plus" size={15} />
                <span>Daftarkan Mitra</span>
              </button>
            ) : null}
          </>
        }
      >
        <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-5">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cari kode / nama / email mitra..."
            aria-label="Cari mitra"
            className={styles.fieldInput}
          />
          <select
            aria-label="Filter bentuk entitas mitra"
            value={kindFilter}
            onChange={(e) => setKindFilter(e.target.value as PartyKind | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Bentuk</option>
            {PARTY_KIND_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <select
            aria-label="Filter status mitra"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as PartyStatus | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Status</option>
            {PARTY_STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <select
            aria-label="Filter peran mitra"
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value as PartyRoleKind | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Peran</option>
            {PARTY_ROLE_KIND_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={() => void loadList()}
            className={styles.btnPrimary}
          >
            Terapkan Filter
          </button>
        </div>

        <div className="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-[#dde3e8] pt-3">
          <div className="flex flex-wrap gap-4 text-xs text-[#0b2439]">
            <label className="flex items-center gap-1.5">
              <input
                type="checkbox"
                checked={includeDeleted}
                onChange={(e) => setIncludeDeleted(e.target.checked)}
              />
              <span>Tampilkan terhapus</span>
            </label>
            <label className="flex items-center gap-1.5">
              <input
                type="checkbox"
                checked={allMode}
                onChange={(e) => setAllMode(e.target.checked)}
              />
              <span>Mode `allParties`</span>
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup mitra"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(e.target.value as "CODE" | "NPWP" | "EMAIL" | "IDS")
              }
              className={`${styles.fieldSelect} w-36`}
            >
              <option value="CODE">By Code</option>
              <option value="NPWP">By NPWP</option>
              <option value="EMAIL">By Email</option>
              <option value="IDS">By IDs</option>
            </select>
            <input
              type="text"
              value={lookupValue}
              onChange={(e) => setLookupValue(e.target.value)}
              placeholder="Nilai lookup..."
              aria-label="Nilai lookup mitra"
              className={`${styles.fieldInput} w-48 ${styles.monoText}`}
            />
            <button
              type="button"
              onClick={() => void handleExactLookup()}
              className={`${styles.btnSecondary} ${styles.btnSmall}`}
            >
              Lookup
            </button>
          </div>
        </div>
      </EfToolbarCard>

      {loadErr ? <ErrorPanel message={loadErr} onRetry={() => void loadList()} /> : null}

      <div className={styles.hairlinePanel}>
        {loading ? (
          <TableSkeleton rows={5} cols={5} />
        ) : items.length === 0 ? (
          <EmptyPanel
            title="Belum Ada Mitra Eksternal"
            description="Daftarkan vendor CMT, supplier bahan, atau rekanan B2B baru."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Mitra</th>
                    <th>Bentuk</th>
                    <th>Status</th>
                    <th>NPWP / PKP</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((p) => (
                    <tr key={p.id}>
                      <td>
                        <button
                          type="button"
                          onClick={() => void openDetail(p)}
                          className="text-left hover:underline"
                        >
                          <p className="font-semibold text-[#0b2439]">
                            {p.display_name}
                          </p>
                          <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                            {p.partner_code} · {p.email || "-"}
                          </p>
                        </button>
                      </td>
                      <td>
                        <EfCodePill code={p.kind} />
                      </td>
                      <td>
                        <EfStatusBadge label={p.status} />
                      </td>
                      <td className={`text-xs ${styles.monoText}`}>
                        <span>{p.tax_id_number || "-"}</span>
                        {p.is_pkp ? <span className="ml-1.5 text-[#067647]">PKP</span> : null}
                      </td>
                      <td>
                        <div className="flex flex-wrap justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => void openDetail(p)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            360° Detail
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => {
                                  setEditing(p);
                                  setFormMode("profile");
                                  setFormOpen(true);
                                }}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Profil
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  setEditing(p);
                                  setFormMode("tax");
                                  setFormOpen(true);
                                }}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Pajak
                              </button>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(p)}
                                className={`${styles.btnDanger} ${styles.btnSmall}`}
                              >
                                Hapus
                              </button>
                            </>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <RelayPaginationBar
              totalCount={totalCount}
              currentCount={items.length}
              pageInfo={pageInfo}
              isLoading={loading}
              onPrev={() => void loadList({ before: pageInfo?.startCursor })}
              onNext={() => void loadList({ after: pageInfo?.endCursor })}
            />
          </>
        )}
      </div>

      {formOpen ? (
        <PartyFormModal
          open={formOpen}
          editing={editing}
          mode={formMode}
          onClose={() => setFormOpen(false)}
          onSaved={() => {
            setFormOpen(false);
            void loadList();
          }}
        />
      ) : null}

      <PartyDetailDrawer
        party={selected}
        canWrite={canWrite}
        onClose={() => setSelected(null)}
        onMutated={() => {
          void loadList();
          if (selected) void openDetail(selected);
        }}
      />

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Hapus Mitra Eksternal?"
        description={`Mitra "${deleteTarget?.display_name}" akan di-soft-delete apabila tidak ada tautan aktif yang memblokir.`}
        confirmLabel="Hapus Mitra"
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDelete()}
      />
    </div>
  );
}
