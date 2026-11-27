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
  LEGAL_ENTITY_STATUS_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import { deleteLegalEntity } from "@/lib/graphql/ef-structure-mutations";
import {
  fetchAllLegalEntities,
  fetchLegalEntities,
  fetchLegalEntitiesByIds,
  fetchLegalEntitiesByNib,
  fetchLegalEntity,
  fetchLegalEntityByCode,
  fetchLegalEntityByTaxId,
} from "@/lib/graphql/ef-structure-queries";
import type { EfLegalEntity, LegalEntityStatus } from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { LegalEntityDetailDrawer } from "./legal-entity-detail-drawer";
import { LegalEntityFormModal } from "./legal-entity-form-modal";

export default function LegalEntitiesClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(
    snapshot,
    "enterprise_foundation.legal_entities.write",
  );

  const [items, setItems] = useState<EfLegalEntity[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<LegalEntityStatus | "">("");
  const [primaryOnly, setPrimaryOnly] = useState(false);
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [allMode, setAllMode] = useState(false);

  const [lookupMode, setLookupMode] = useState<"CODE" | "NPWP" | "NIB" | "IDS">("CODE");
  const [lookupValue, setLookupValue] = useState("");

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<EfLegalEntity | null>(null);
  const [selected, setSelected] = useState<EfLegalEntity | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<EfLegalEntity | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllLegalEntities(20, cursor?.after)
        : await fetchLegalEntities({
            query: query.trim() || undefined,
            status: statusFilter || null,
            primary_only: primaryOnly ? true : null,
            include_deleted: includeDeleted,
            first: 20,
            after: cursor?.after || null,
            before: cursor?.before || null,
          });
    },
    [allMode, includeDeleted, primaryOnly, query, statusFilter],
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
    setLoadErr(null);
    try {
      if (lookupMode === "CODE") {
        const one = await fetchLegalEntityByCode(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "NPWP") {
        const one = await fetchLegalEntityByTaxId(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "NIB") {
        const conn = await fetchLegalEntitiesByNib(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else {
        const ids = val.split(",").map((s) => s.trim()).filter(Boolean);
        const list = await fetchLegalEntitiesByIds(ids);
        setItems(list);
        setTotalCount(list.length);
        setPageInfo(null);
      }
    } catch (err) {
      pushToast("error", "Lookup Badan Hukum Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  async function openDetail(le: EfLegalEntity) {
    setSelected(le);
    try {
      const full = await fetchLegalEntity(le.id);
      if (full) setSelected(full);
    } catch (err) {
      pushToast("error", "Gagal Memuat Detail", formatEfError(err));
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const ok = await runDashboardAction(
      () => deleteLegalEntity(target.id, csrfToken),
      { stepUpReason: "Menghapus badan hukum adalah operasi sensitif." },
    );
    if (ok) {
      setDeleteTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "Badan Hukum Dihapus", target.legal_name);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Direktori Badan Hukum Perusahaan"
        description="Manajemen entitas hukum (`legalEntities`, lookup kode/NPWP/NIB, aktivasi & scope IAM)."
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
                  setFormOpen(true);
                }}
                className={styles.btnPrimary}
              >
                <IdentityIcon name="plus" size={15} />
                <span>Tambah Badan Hukum</span>
              </button>
            ) : null}
          </>
        }
      >
        <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-4">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cari kode / nama legal..."
            aria-label="Cari badan hukum"
            className={`${styles.fieldInput} lg:col-span-2`}
          />
          <select
            aria-label="Filter status badan hukum"
            value={statusFilter}
            onChange={(e) =>
              setStatusFilter(e.target.value as LegalEntityStatus | "")
            }
            className={styles.fieldSelect}
          >
            <option value="">Semua Status</option>
            {LEGAL_ENTITY_STATUS_OPTIONS.map((opt) => (
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
                checked={primaryOnly}
                onChange={(e) => setPrimaryOnly(e.target.checked)}
              />
              <span>Utama saja</span>
            </label>
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
              <span>Mode `allLegalEntities`</span>
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup badan hukum"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(e.target.value as "CODE" | "NPWP" | "NIB" | "IDS")
              }
              className={`${styles.fieldSelect} w-36`}
            >
              <option value="CODE">By Code</option>
              <option value="NPWP">By NPWP</option>
              <option value="NIB">By NIB</option>
              <option value="IDS">By IDs</option>
            </select>
            <input
              type="text"
              value={lookupValue}
              onChange={(e) => setLookupValue(e.target.value)}
              placeholder="Nilai lookup..."
              aria-label="Nilai lookup badan hukum"
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
            title="Belum Ada Badan Hukum"
            description="Buat entitas legal perusahaan baru untuk mengaktifkan struktur organisasi & scope IAM."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Entitas</th>
                    <th>Status</th>
                    <th>NPWP / NIB</th>
                    <th>PKP / Utama</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((le) => (
                    <tr key={le.id}>
                      <td>
                        <button
                          type="button"
                          onClick={() => void openDetail(le)}
                          className="text-left hover:underline"
                        >
                          <p className="font-semibold text-[#0b2439]">
                            {le.legal_name}
                          </p>
                          <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                            {le.entity_code} · {le.jurisdiction_country_code}
                          </p>
                        </button>
                      </td>
                      <td>
                        <EfStatusBadge label={le.status} />
                      </td>
                      <td className={`text-xs ${styles.monoText}`}>
                        <p>{le.tax_id_number || "-"}</p>
                        <p className="text-[#5a6672]">{le.nib || "-"}</p>
                      </td>
                      <td>
                        <div className="flex flex-wrap gap-1">
                          {le.is_primary_operating_entity ? (
                            <EfStatusBadge label="UTAMA" toneSource="ACTIVE" />
                          ) : null}
                          {le.is_pkp ? <EfCodePill code="PKP" /> : null}
                        </div>
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => void openDetail(le)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Detail
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => {
                                  setEditing(le);
                                  setFormOpen(true);
                                }}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Ubah
                              </button>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(le)}
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
        <LegalEntityFormModal
          open={formOpen}
          editing={editing}
          onClose={() => setFormOpen(false)}
          onSaved={() => {
            setFormOpen(false);
            void loadList();
          }}
        />
      ) : null}

      <LegalEntityDetailDrawer
        entity={selected}
        canWrite={canWrite}
        onClose={() => setSelected(null)}
        onMutated={() => {
          void loadList();
          if (selected) void openDetail(selected);
        }}
      />

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Hapus Badan Hukum?"
        description={`Entitas "${deleteTarget?.legal_name}" akan di-soft-delete apabila tidak lagi memiliki unit organisasi atau lokasi dependen.`}
        confirmLabel="Hapus Badan Hukum"
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDelete()}
      />
    </div>
  );
}
