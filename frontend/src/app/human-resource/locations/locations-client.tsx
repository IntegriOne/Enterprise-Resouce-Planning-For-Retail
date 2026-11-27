"use client";

import { useCallback, useEffect, useState } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ActiveBooleanBadge,
  ConfirmModal,
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  SlideOverDrawer,
  TableSkeleton,
} from "@/components/identity/identity-primitives";
import {
  EfCodePill,

  EfScopeDecisionBox,
  EfToolbarCard,
} from "@/components/workspace/ef-ui-primitives";
import {
  LOCATION_KIND_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import {
  activateLocation,
  deactivateLocation,
  deleteLocation,
} from "@/lib/graphql/ef-structure-mutations";
import {
  fetchAllLocations,

  fetchLocation,
  fetchLocationByCode,
  fetchLocationChildren,
  fetchLocations,
  fetchLocationsByAddressId,
  fetchLocationsByIds,
  fetchLocationsByLegalEntityAndKind,
  validateFacilityScope,
  validateWarehouseScope,
} from "@/lib/graphql/ef-structure-queries";
import type {
  EfLocation,
  LocationKind,
  ScopeTargetDecision,
} from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { LocationFormModal } from "./location-form-modal";

export default function LocationsClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(
    snapshot,
    "enterprise_foundation.locations.write",
  );

  const [items, setItems] = useState<EfLocation[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [legalEntityFilter, setLegalEntityFilter] = useState("");
  const [kindFilter, setKindFilter] = useState<LocationKind | "">("");
  const [activeOnly, setActiveOnly] = useState(false);
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [allMode, setAllMode] = useState(false);

  const [lookupMode, setLookupMode] = useState<
    "CODE" | "IDS" | "ADDRESS_ID" | "ENTITY_KIND"
  >("CODE");
  const [lookupValue, setLookupValue] = useState("");

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<EfLocation | null>(null);
  const [selected, setSelected] = useState<EfLocation | null>(null);
  const [children, setChildren] = useState<EfLocation[] | null>(null);
  const [decision, setDecision] = useState<ScopeTargetDecision | null>(null);
  const [decisionErr, setDecisionErr] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<EfLocation | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllLocations(20, cursor?.after)
        : await fetchLocations({
            query: query.trim() || undefined,
            legal_entity_id: legalEntityFilter.trim() || null,
            kind: kindFilter || null,
            active_only: activeOnly,
            include_deleted: includeDeleted,
            first: 20,
            after: cursor?.after || null,
            before: cursor?.before || null,
          });
    },
    [activeOnly, allMode, includeDeleted, kindFilter, legalEntityFilter, query],
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
    if (!val && lookupMode !== "ENTITY_KIND") return;
    setLoading(true);
    try {
      if (lookupMode === "CODE") {
        const one = await fetchLocationByCode(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "IDS") {
        const ids = val.split(",").map((s) => s.trim()).filter(Boolean);
        const list = await fetchLocationsByIds(ids);
        setItems(list);
        setTotalCount(list.length);
        setPageInfo(null);
      } else if (lookupMode === "ADDRESS_ID") {
        const conn = await fetchLocationsByAddressId(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else {
        const entity = val || legalEntityFilter.trim();
        const k = (kindFilter || "WAREHOUSE") as LocationKind;
        const conn = await fetchLocationsByLegalEntityAndKind(entity, k, activeOnly, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      }
    } catch (err) {
      pushToast("error", "Lookup Lokasi Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  async function openDetail(loc: EfLocation) {
    setSelected(loc);
    setChildren(null);
    setDecision(null);
    setDecisionErr(null);
    try {
      const [full, ch] = await Promise.all([
        fetchLocation(loc.id),
        fetchLocationChildren(loc.id, 15),
      ]);
      if (full) setSelected(full);
      setChildren(ch.nodes);
    } catch (err) {
      pushToast("error", "Gagal Memuat Detail Lokasi", formatEfError(err));
    }
  }

  async function handleToggleActive(loc: EfLocation) {
    const updated = await runDashboardAction(() =>
      loc.is_active
        ? deactivateLocation(loc.id, csrfToken)
        : activateLocation(loc.id, csrfToken),
    );
    if (updated) {
      pushToast(
        "success",
        updated.is_active ? "Lokasi Diaktifkan" : "Lokasi Dinonaktifkan",
        updated.name,
      );
      await loadList();
      if (selected?.id === updated.id) setSelected(updated);
    }
  }

  async function handleValidate(mode: "FACILITY" | "WAREHOUSE") {
    if (!selected) return;
    setDecision(null);
    setDecisionErr(null);
    try {
      const res =
        mode === "FACILITY"
          ? await validateFacilityScope(selected.id)
          : await validateWarehouseScope(selected.id);
      setDecision(res);
    } catch (err) {
      setDecisionErr(formatEfError(err));
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const ok = await runDashboardAction(
      () => deleteLocation(target.id, csrfToken),
      { stepUpReason: "Menghapus lokasi fisik bersifat sensitif." },
    );
    if (ok) {
      setDeleteTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "Lokasi Dihapus", target.name);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Lokasi Fisik, Gudang & Toko"
        description="Pengelolaan fasilitas (`locations`, lookup kode/alamat/entity+kind, serta validasi `FACILITY`/`WAREHOUSE`)."
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
                <span>Tambah Lokasi</span>
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
            placeholder="Cari kode / nama lokasi..."
            aria-label="Cari lokasi"
            className={styles.fieldInput}
          />
          <input
            type="text"
            value={legalEntityFilter}
            onChange={(e) => setLegalEntityFilter(e.target.value)}
            placeholder="Filter Legal Entity ID..."
            aria-label="Filter ID badan hukum"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <select
            aria-label="Filter jenis lokasi"
            value={kindFilter}
            onChange={(e) => setKindFilter(e.target.value as LocationKind | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Jenis</option>
            {LOCATION_KIND_OPTIONS.map((opt) => (
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
                checked={activeOnly}
                onChange={(e) => setActiveOnly(e.target.checked)}
              />
              <span>Aktif saja</span>
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
              <span>Mode `allLocations`</span>
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup lokasi"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(
                  e.target.value as "CODE" | "IDS" | "ADDRESS_ID" | "ENTITY_KIND",
                )
              }
              className={`${styles.fieldSelect} w-40`}
            >
              <option value="CODE">By Code</option>
              <option value="IDS">By IDs</option>
              <option value="ADDRESS_ID">By Address ID</option>
              <option value="ENTITY_KIND">By Entity + Kind</option>
            </select>
            <input
              type="text"
              value={lookupValue}
              onChange={(e) => setLookupValue(e.target.value)}
              placeholder="Nilai / UUID lookup..."
              aria-label="Nilai lookup lokasi"
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
            title="Belum Ada Lokasi"
            description="Tambahkan lokasi operasional (Gudang, Toko, HQ, Cabang) di bawah Badan Hukum aktif."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Lokasi</th>
                    <th>Jenis</th>
                    <th>Zona Waktu</th>
                    <th>Status</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((loc) => (
                    <tr key={loc.id}>
                      <td>
                        <button
                          type="button"
                          onClick={() => void openDetail(loc)}
                          className="text-left hover:underline"
                        >
                          <p className="font-semibold text-[#0b2439]">{loc.name}</p>
                          <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                            {loc.location_code}
                          </p>
                        </button>
                      </td>
                      <td>
                        <EfCodePill code={loc.kind} />
                      </td>
                      <td className={`text-xs ${styles.monoText}`}>{loc.timezone}</td>
                      <td>
                        <ActiveBooleanBadge active={loc.is_active} />
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => void openDetail(loc)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Detail
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => void handleToggleActive(loc)}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                {loc.is_active ? "Nonaktifkan" : "Aktifkan"}
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  setEditing(loc);
                                  setFormOpen(true);
                                }}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Ubah
                              </button>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(loc)}
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
        <LocationFormModal
          open={formOpen}
          editing={editing}
          onClose={() => setFormOpen(false)}
          onSaved={() => {
            setFormOpen(false);
            void loadList();
          }}
        />
      ) : null}

      <SlideOverDrawer
        open={Boolean(selected)}
        title={selected?.name || "Detail Lokasi"}
        subtitle={selected?.location_code}
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <EfCodePill code={selected.kind} />
              <ActiveBooleanBadge active={selected.is_active} />

            </div>

            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => void handleValidate("FACILITY")}
                className={`${styles.btnPrimary} ${styles.btnSmall}`}
              >
                Validasi Scope FACILITY
              </button>
              <button
                type="button"
                onClick={() => void handleValidate("WAREHOUSE")}
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Validasi Scope WAREHOUSE
              </button>
            </div>

            <EfScopeDecisionBox decision={decision} error={decisionErr} showIdentifiers={false} />

            {children ? (
              <div className="rounded-xl border border-[#dde3e8] p-4">
                <p className="text-xs font-semibold text-[#0b2439]">
                  Lokasi Anak ({children.length})
                </p>
                <div className="mt-2 space-y-1">
                  {children.map((c) => (
                    <p key={c.id} className="text-xs text-[#5a6672]">
                      <strong className={styles.monoText}>{c.location_code}</strong>{" "}
                      — {c.name} ({c.kind})
                    </p>
                  ))}
                  {children.length === 0 ? (
                    <p className="text-xs text-[#5a6672]">Tidak ada sub-lokasi.</p>
                  ) : null}
                </div>
              </div>
            ) : null}
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Hapus Lokasi?"
        description={`Lokasi "${deleteTarget?.name}" akan di-soft-delete apabila tidak masih direferensikan oleh unit organisasi atau sub-lokasi aktif.`}
        confirmLabel="Hapus Lokasi"
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDelete()}
      />
    </div>
  );
}
