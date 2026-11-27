"use client";

import { useCallback, useEffect, useRef, useState } from "react";
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
  ORG_UNIT_KIND_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import {
  activateOrganizationUnit,
  assignOrganizationUnitHeadPerson,
  deactivateOrganizationUnit,
  deleteOrganizationUnit,
} from "@/lib/graphql/ef-structure-mutations";
import {
  fetchAllOrganizationUnits,

  fetchOrganizationUnit,
  fetchOrganizationUnitByEntityAndCode,
  fetchOrganizationUnitChildren,
  fetchOrganizationUnits,
  fetchOrganizationUnitsByHeadPersonId,
  fetchOrganizationUnitsByIds,
  fetchOrganizationUnitsByPrimaryLocationId,
  validateOrganizationUnitScopeTarget,
} from "@/lib/graphql/ef-structure-queries";
import type {
  EfOrganizationUnit,
  OrgUnitKind,
  ScopeTargetDecision,
} from "@/lib/graphql/ef-types";
import { fetchWorkerDisplayLabel, listProvisioningWorkers } from "@/lib/graphql/hrm-worker-provisioning";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { OrgUnitFormModal } from "./org-unit-form-modal";

export default function OrgUnitsClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(
    snapshot,
    "enterprise_foundation.organization_units.write",
  );

  const [items, setItems] = useState<EfOrganizationUnit[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [legalEntityFilter, setLegalEntityFilter] = useState("");
  const [kindFilter, setKindFilter] = useState<OrgUnitKind | "">("");
  const [activeOnly, setActiveOnly] = useState(false);
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [allMode, setAllMode] = useState(false);

  const [lookupMode, setLookupMode] = useState<
    "ENTITY_CODE" | "IDS" | "HEAD_PERSON" | "PRIMARY_LOCATION"
  >("IDS");
  const [lookupValue, setLookupValue] = useState("");

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<EfOrganizationUnit | null>(null);
  const [selected, setSelected] = useState<EfOrganizationUnit | null>(null);
  const detailRequest = useRef(0);
  const [children, setChildren] = useState<EfOrganizationUnit[] | null>(null);
  const [headInput, setHeadInput] = useState("");
  const [headSaving, setHeadSaving] = useState(false);
  const [decision, setDecision] = useState<ScopeTargetDecision | null>(null);
  const [decisionErr, setDecisionErr] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<EfOrganizationUnit | null>(
    null,
  );

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllOrganizationUnits(20, cursor?.after)
        : await fetchOrganizationUnits({
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
    if (!val) return;
    setLoading(true);
    try {
      if (lookupMode === "ENTITY_CODE") {
        const one = await fetchOrganizationUnitByEntityAndCode(
          legalEntityFilter.trim(),
          val,
        );
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "IDS") {
        const ids = val.split(",").map((s) => s.trim()).filter(Boolean);
        const list = await fetchOrganizationUnitsByIds(ids);
        setItems(list);
        setTotalCount(list.length);
        setPageInfo(null);
      } else if (lookupMode === "HEAD_PERSON") {
        const worker = await findWorkerByNumber(val);
        const conn = await fetchOrganizationUnitsByHeadPersonId(worker.id, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else {
        const conn = await fetchOrganizationUnitsByPrimaryLocationId(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      }
    } catch (err) {
      pushToast("error", "Lookup Unit Organisasi Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  const headID = selected?.head_person_id;
  const [headLabel, setHeadLabel] = useState<{ workerID: string; label: string | null } | null>(null);

  useEffect(() => {
    if (!headID) return;
    let active = true;
    void fetchWorkerDisplayLabel(headID).then(
      (label) => { if (active) setHeadLabel({ workerID: headID, label }); },
      () => { if (active) setHeadLabel({ workerID: headID, label: null }); },
    );
    return () => { active = false; };
  }, [headID]);

  async function openDetail(ou: EfOrganizationUnit) {
    const request = ++detailRequest.current;
    setSelected(ou);
    setHeadInput("");
    setChildren(null);
    setDecision(null);
    setDecisionErr(null);
    try {
      const [full, ch] = await Promise.all([
        fetchOrganizationUnit(ou.id),
        fetchOrganizationUnitChildren(ou.id, 15),
      ]);
      if (request !== detailRequest.current) return;
      if (full) setSelected((current) => current === ou ? full : current);
      setChildren(ch.nodes);
    } catch (err) {
      if (request === detailRequest.current) {
        pushToast("error", "Gagal Memuat Detail Unit", formatEfError(err));
      }
    }
  }

  function closeDetail() {
    detailRequest.current++;
    setSelected(null);
  }

  async function findWorkerByNumber(number: string) {
    const result = await listProvisioningWorkers(number, null, null);
    if (result.items.length !== 1 || result.total_count !== 1) {
      throw new Error("Nomor pekerja tidak ditemukan atau tidak unik.");
    }
    return result.items[0];
  }

  async function handleAssignHead(clear = false) {
    const number = headInput.trim();
    if (!selected || headSaving || (!clear && !number)) return;

    const request = detailRequest.current;
    setHeadSaving(true);
    try {
      const worker = clear ? null : await findWorkerByNumber(number);
      const updated = await runDashboardAction(() =>
        assignOrganizationUnitHeadPerson(selected.id, worker?.id ?? null, csrfToken),
      );
      if (updated) {
        setSelected((current) => current?.id === selected.id ? updated : current);
        if (request === detailRequest.current) setHeadInput("");
        pushToast("success", "Kepala Unit Diperbarui", updated.name);
        await loadList();
      }
    } catch (err) {
      pushToast("error", "Gagal Menetapkan Kepala Unit", formatEfError(err));
    } finally {
      setHeadSaving(false);
    }
  }

  async function handleToggleActive(ou: EfOrganizationUnit) {
    const updated = await runDashboardAction(() =>
      ou.is_active
        ? deactivateOrganizationUnit(ou.id, csrfToken)
        : activateOrganizationUnit(ou.id, csrfToken),
    );
    if (updated) {
      pushToast(
        "success",
        updated.is_active ? "Unit Diaktifkan" : "Unit Dinonaktifkan",
        updated.name,
      );
      await loadList();
      setSelected((current) => current?.id === updated.id ? updated : current);
    }
  }

  async function handleValidateScope() {
    if (!selected) return;
    const request = detailRequest.current;
    setDecision(null);
    setDecisionErr(null);
    try {
      const res = await validateOrganizationUnitScopeTarget(selected.id);
      if (request === detailRequest.current) setDecision(res);
    } catch (err) {
      if (request === detailRequest.current) setDecisionErr(formatEfError(err));
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const ok = await runDashboardAction(
      () => deleteOrganizationUnit(target.id, csrfToken),
      { stepUpReason: "Menghapus unit organisasi adalah operasi sensitif." },
    );
    if (ok) {
      setDeleteTarget(null);
      setSelected((current) => current?.id === target.id ? null : current);
      pushToast("success", "Unit Organisasi Dihapus", target.name);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Hierarki Unit Organisasi"
        description="Manajemen divisi/departemen/tim/cabang (`organizationUnits`, kepala unit, serta validasi `ORGANIZATION_UNIT`)."
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
                <span>Tambah Unit</span>
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
            placeholder="Cari kode / nama unit..."
            aria-label="Cari unit organisasi"
            className={styles.fieldInput}
          />
          <input
            type="text"
            value={legalEntityFilter}
            onChange={(e) => setLegalEntityFilter(e.target.value)}
            placeholder="Filter Legal Entity ID..."
            aria-label="Filter ID badan hukum unit"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <select
            aria-label="Filter jenis unit organisasi"
            value={kindFilter}
            onChange={(e) => setKindFilter(e.target.value as OrgUnitKind | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Jenis</option>
            {ORG_UNIT_KIND_OPTIONS.map((opt) => (
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
              <span>Mode `allOrganizationUnits`</span>
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup unit organisasi"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(
                  e.target.value as
                    | "ENTITY_CODE"
                    | "IDS"
                    | "HEAD_PERSON"
                    | "PRIMARY_LOCATION",
                )
              }
              className={`${styles.fieldSelect} w-44`}
            >
              <option value="IDS">By IDs</option>
              <option value="ENTITY_CODE">By Entity + Unit Code</option>
              <option value="HEAD_PERSON">Nomor Pekerja Kepala Unit</option>
              <option value="PRIMARY_LOCATION">By Primary Location</option>
            </select>
            <input
              type="text"
              value={lookupValue}
              onChange={(e) => setLookupValue(e.target.value)}
              placeholder="Nilai lookup..."
              aria-label="Nilai lookup unit organisasi"
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
            title="Belum Ada Unit Organisasi"
            description="Buat unit organisasi (divisi/departemen/tim) di bawah Badan Hukum aktif."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Unit</th>
                    <th>Jenis</th>
                    <th>Cost Center</th>
                    <th>Status</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((ou) => (
                    <tr key={ou.id}>
                      <td>
                        <button
                          type="button"
                          onClick={() => void openDetail(ou)}
                          className="text-left hover:underline"
                        >
                          <p className="font-semibold text-[#0b2439]">{ou.name}</p>
                          <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                            {ou.unit_code}
                          </p>
                        </button>
                      </td>
                      <td>
                        <EfCodePill code={ou.kind} />
                      </td>
                      <td className={`text-xs ${styles.monoText}`}>
                        {ou.cost_center_code || "-"}
                      </td>
                      <td>
                        <ActiveBooleanBadge active={ou.is_active} />
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => void openDetail(ou)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Detail
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => void handleToggleActive(ou)}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                {ou.is_active ? "Nonaktifkan" : "Aktifkan"}
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  setEditing(ou);
                                  setFormOpen(true);
                                }}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Ubah
                              </button>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(ou)}
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
        <OrgUnitFormModal
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
        title={selected?.name || "Detail Unit Organisasi"}
        subtitle={selected?.unit_code}
        onClose={closeDetail}
      >
        {selected ? (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <EfCodePill code={selected.kind} />
              <ActiveBooleanBadge active={selected.is_active} />

            </div>

            <div className="text-sm">
              <span className="font-semibold">Kepala unit</span>
              <p>
                {!headID
                  ? "Belum ditetapkan"
                  : headLabel?.workerID === headID
                    ? headLabel.label ?? "Data kepala unit tidak tersedia"
                    : "Memuat data kepala unit…"}
              </p>
            </div>

            {canWrite ? (
              <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
                <label
                  htmlFor="ou-head-assign"
                  className="block text-xs font-semibold text-[#0b2439]"
                >
                  Nomor pekerja kepala unit
                </label>
                <div className="flex gap-2">
                  <input
                    id="ou-head-assign"
                    type="text"
                    value={headInput}
                    onChange={(e) => setHeadInput(e.target.value)}
                    placeholder="Contoh W-001"
                    className={`${styles.fieldInput} ${styles.monoText}`}
                  />
                  <button
                    type="button"
                    onClick={() => void handleAssignHead()}
                    disabled={headSaving || !headInput.trim()}
                    className={`${styles.btnPrimary} ${styles.btnSmall}`}
                  >
                    Tetapkan
                  </button>
                  {selected.head_person_id ? (
                    <button type="button" onClick={() => void handleAssignHead(true)}
                      disabled={headSaving} className={`${styles.btnSecondary} ${styles.btnSmall}`}>
                      Lepas kepala unit
                    </button>
                  ) : null}
                </div>
              </div>
            ) : null}

            <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
              <div className="flex items-center justify-between gap-2">
                <h3 className="text-xs font-semibold text-[#0b2439]">
                  Validasi Target Scope ORGANIZATION_UNIT
                </h3>
                <button
                  type="button"
                  onClick={() => void handleValidateScope()}
                  className={`${styles.btnSecondary} ${styles.btnSmall}`}
                >
                  Validasi Scope
                </button>
              </div>
              <EfScopeDecisionBox decision={decision} error={decisionErr} showIdentifiers={false} />
            </div>

            {children ? (
              <div className="rounded-xl border border-[#dde3e8] p-4">
                <p className="text-xs font-semibold text-[#0b2439]">
                  Sub-Unit Anak ({children.length})
                </p>
                <div className="mt-2 space-y-1">
                  {children.map((ch) => (
                    <p key={ch.id} className="text-xs text-[#5a6672]">
                      <strong className={styles.monoText}>{ch.unit_code}</strong>{" "}
                      — {ch.name}
                    </p>
                  ))}
                  {children.length === 0 ? (
                    <p className="text-xs text-[#5a6672]">Tidak ada sub-unit.</p>
                  ) : null}
                </div>
              </div>
            ) : null}
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Hapus Unit Organisasi?"
        description={`Unit "${deleteTarget?.name}" akan di-soft-delete apabila tidak memiliki sub-unit aktif.`}
        confirmLabel="Hapus Unit"
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDelete()}
      />
    </div>
  );
}
