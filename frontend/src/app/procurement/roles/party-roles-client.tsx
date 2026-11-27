"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
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
  EfRelayNodeInspectorButton,
  EfScopeDecisionBox,
  EfToolbarCard,
} from "@/components/workspace/ef-ui-primitives";
import {
  PARTY_ROLE_KIND_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import {
  activatePartyRole,
  assignPartyRole,
  deactivatePartyRole,
  removePartyRole,
  updatePartyRoleNote,
} from "@/lib/graphql/ef-partner-mutations";
import {
  fetchAllParties,
  fetchAllPartyRoles,
  fetchAllPartyRolesByKind,
  fetchEfPartnerNode,
  fetchPartyRole,
  fetchPartyRoleByPartyAndKind,
  fetchPartyRoles,
  fetchPartyRolesByKind,
  fetchPartyRolesByParty,
  validatePartyRoleScope,
} from "@/lib/graphql/ef-partner-queries";
import type {
  EfParty,
  EfPartyRole,
  PartyRoleKind,
  ScopeTargetDecision,
} from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";

export default function PartyRolesClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(snapshot, "enterprise_foundation.parties.write");

  const [items, setItems] = useState<EfPartyRole[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [parties, setParties] = useState<EfParty[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [partyFilter, setPartyFilter] = useState("");
  const [roleFilter, setRoleFilter] = useState<PartyRoleKind | "">("");
  const [activeOnly, setActiveOnly] = useState(false);
  const [listMode, setListMode] = useState<
    "FILTER" | "ALL" | "KIND" | "ALL_KIND"
  >("FILTER");

  const [lookupMode, setLookupMode] = useState<"BY_PARTY" | "PARTY_AND_KIND" | "BY_ID">("BY_PARTY");
  const [lookupPartyId, setLookupPartyId] = useState("");
  const [lookupKind, setLookupKind] = useState<PartyRoleKind>("SUPPLIER");

  const [assignPartyId, setAssignPartyId] = useState("");
  const [assignKind, setAssignKind] = useState<PartyRoleKind>("VENDOR");
  const [assignNote, setAssignNote] = useState("");

  const [selected, setSelected] = useState<EfPartyRole | null>(null);
  const [editNote, setEditNote] = useState("");
  const [decision, setDecision] = useState<ScopeTargetDecision | null>(null);
  const [decisionErr, setDecisionErr] = useState<string | null>(null);
  const [removeTarget, setRemoveTarget] = useState<EfPartyRole | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return listMode === "ALL"
        ? await fetchAllPartyRoles(20, cursor?.after)
        : listMode === "KIND"
          ? await fetchPartyRolesByKind(roleFilter || "SUPPLIER", activeOnly, 20)
          : listMode === "ALL_KIND"
            ? await fetchAllPartyRolesByKind(
                roleFilter || "SUPPLIER",
                activeOnly,
                20,
              )
            : await fetchPartyRoles({
                query: query.trim() || undefined,
                party_id: partyFilter.trim() || null,
                role: roleFilter || null,
                active_only: activeOnly,
                first: 20,
                after: cursor?.after || null,
                before: cursor?.before || null,
              });
    },
    [activeOnly, listMode, partyFilter, query, roleFilter],
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

  useEffect(() => {
    let active = true;
    void fetchAllParties(25)
      .then((res) => {
        if (!active) return;
        setParties(res.nodes);
        const firstPartyId = res.nodes[0]?._id;
        if (firstPartyId) {
          setAssignPartyId((prev) => prev || firstPartyId);
        }
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, []);

  async function handleExactLookup() {
    const val = lookupPartyId.trim();
    if (!val) return;
    setLoading(true);
    try {
      if (lookupMode === "BY_ID") {
        const one = await fetchPartyRole(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "PARTY_AND_KIND") {
        const one = await fetchPartyRoleByPartyAndKind(val, lookupKind);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else {
        const conn = await fetchPartyRolesByParty(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      }
    } catch (err) {
      pushToast("error", "Lookup Peran Mitra Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleAssign(e: FormEvent) {
    e.preventDefault();
    if (!assignPartyId.trim()) return;
    const created = await runDashboardAction(() =>
      assignPartyRole(
        assignPartyId.trim(),
        assignKind,
        assignNote.trim() || null,
        csrfToken,
      ),
    );
    if (created) {
      setAssignNote("");
      pushToast("success", "Peran Mitra Ditetapkan", created.role);
      await loadList();
    }
  }

  async function handleToggle(role: EfPartyRole) {
    const updated = await runDashboardAction(() =>
      role.is_active
        ? deactivatePartyRole(role.id, csrfToken)
        : activatePartyRole(role.id, csrfToken),
    );
    if (updated) {
      pushToast(
        "success",
        updated.is_active ? "Peran Diaktifkan" : "Peran Dinonaktifkan",
        updated.role,
      );
      await loadList();
      if (selected?.id === updated.id) setSelected(updated);
    }
  }

  async function handleUpdateNote() {
    if (!selected) return;
    const updated = await runDashboardAction(() =>
      updatePartyRoleNote(selected.id, editNote.trim() || null, csrfToken),
    );
    if (updated) {
      setSelected(updated);
      pushToast("success", "Catatan Peran Diperbarui", updated.role);
      await loadList();
    }
  }

  async function handleValidatePartyScope(role: EfPartyRole) {
    setDecision(null);
    setDecisionErr(null);
    try {
      const res = await validatePartyRoleScope(role.party_id, role.role);
      setDecision(res);
    } catch (err) {
      setDecisionErr(formatEfError(err));
    }
  }

  async function handleRemove() {
    if (!removeTarget) return;
    const target = removeTarget;
    const ok = await runDashboardAction(
      () => removePartyRole(target.id, csrfToken),
      { stepUpReason: "Menghapus peran bisnis mitra bersifat sensitif." },
    );
    if (ok) {
      setRemoveTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "Peran Mitra Dihapus", target.role);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Peran Bisnis Mitra (`PartyRole`)"
        description="Penetapan peran `SUPPLIER`, `VENDOR`, `CUSTOMER_B2B`, `CARRIER`, `FINANCIAL_INSTITUTION` beserta validasi target scope."
        actions={
          <button
            type="button"
            onClick={() => void loadList()}
            className={styles.btnSecondary}
          >
            <IdentityIcon name="refresh" size={15} />
            <span>Segarkan</span>
          </button>
        }
      >
        <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-4">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cari catatan / party..."
            aria-label="Cari peran mitra"
            className={styles.fieldInput}
          />
          <input
            type="text"
            value={partyFilter}
            onChange={(e) => setPartyFilter(e.target.value)}
            placeholder="Filter Party ID..."
            aria-label="Filter ID party"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <select
            aria-label="Filter jenis peran"
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value as PartyRoleKind | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Jenis Peran</option>
            {PARTY_ROLE_KIND_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <select
            aria-label="Mode query PartyRole"
            value={listMode}
            onChange={(e) =>
              setListMode(
                e.target.value as "FILTER" | "ALL" | "KIND" | "ALL_KIND",
              )
            }
            className={styles.fieldSelect}
          >
            <option value="FILTER">Query `partyRoles`</option>
            <option value="ALL">Query `allPartyRoles`</option>
            <option value="KIND">Query `partyRolesByKind`</option>
            <option value="ALL_KIND">Query `allPartyRolesByKind`</option>
          </select>
        </div>

        <div className="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-[#dde3e8] pt-3">
          <label className="flex items-center gap-1.5 text-xs text-[#0b2439]">
            <input
              type="checkbox"
              checked={activeOnly}
              onChange={(e) => setActiveOnly(e.target.checked)}
            />
            <span>Aktif saja</span>
          </label>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup PartyRole"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(
                  e.target.value as "BY_PARTY" | "PARTY_AND_KIND" | "BY_ID",
                )
              }
              className={`${styles.fieldSelect} w-44`}
            >
              <option value="BY_PARTY">By Party</option>
              <option value="PARTY_AND_KIND">By Party + Kind</option>
              <option value="BY_ID">By Role ID</option>
            </select>
            <input
              type="text"
              value={lookupPartyId}
              onChange={(e) => setLookupPartyId(e.target.value)}
              placeholder="Party ID / Role ID..."
              aria-label="ID lookup peran mitra"
              className={`${styles.fieldInput} w-48 ${styles.monoText}`}
            />
            {lookupMode === "PARTY_AND_KIND" ? (
              <select
                aria-label="Jenis peran lookup"
                value={lookupKind}
                onChange={(e) => setLookupKind(e.target.value as PartyRoleKind)}
                className={`${styles.fieldSelect} w-36`}
              >
                {PARTY_ROLE_KIND_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.value}
                  </option>
                ))}
              </select>
            ) : null}
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

      {canWrite ? (
        <form
          onSubmit={(e) => void handleAssign(e)}
          className={`${styles.hairlinePanel} grid grid-cols-1 gap-2.5 p-4 sm:grid-cols-4`}
        >
          {parties.length > 0 ? (
            <select
              aria-label="Pilih mitra untuk ditetapkan peran"
              value={assignPartyId}
              onChange={(e) => setAssignPartyId(e.target.value)}
              className={styles.fieldSelect}
            >
              {parties.map((p) => (
                <option key={p.id} value={p._id}>
                  {p.partner_code} — {p.display_name}
                </option>
              ))}
            </select>
          ) : (
            <input
              type="text"
              value={assignPartyId}
              onChange={(e) => setAssignPartyId(e.target.value)}
              placeholder="Party UUID/GID"
              aria-label="Party UUID untuk peran baru"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
          )}
          <select
            aria-label="Pilih peran baru mitra"
            value={assignKind}
            onChange={(e) => setAssignKind(e.target.value as PartyRoleKind)}
            className={styles.fieldSelect}
          >
            {PARTY_ROLE_KIND_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <input
            type="text"
            value={assignNote}
            onChange={(e) => setAssignNote(e.target.value)}
            placeholder="Catatan kontrak / kategori (opsional)"
            aria-label="Catatan peran mitra"
            className={styles.fieldInput}
          />
          <button type="submit" className={styles.btnPrimary}>
            Tetapkan Peran Mitra
          </button>
        </form>
      ) : null}

      {loadErr ? <ErrorPanel message={loadErr} onRetry={() => void loadList()} /> : null}

      <div className={styles.hairlinePanel}>
        {loading ? (
          <TableSkeleton rows={5} cols={4} />
        ) : items.length === 0 ? (
          <EmptyPanel
            title="Belum Ada Peran Mitra"
            description="Tetapkan peran SUPPLIER atau VENDOR kepada mitra aktif di atas."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Peran</th>
                    <th>Party ID</th>
                    <th>Catatan</th>
                    <th>Status</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((r) => (
                    <tr key={r.id}>
                      <td>
                        <EfCodePill code={r.role} />
                      </td>
                      <td className={`text-xs ${styles.monoText}`}>{r.party_id}</td>
                      <td className="text-xs text-[#5a6672]">
                        {r.reference_note || "-"}
                      </td>
                      <td>
                        <ActiveBooleanBadge active={r.is_active} />
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => {
                              setSelected(r);
                              setEditNote(r.reference_note || "");
                              setDecision(null);
                              setDecisionErr(null);
                            }}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Kelola
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => void handleToggle(r)}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                {r.is_active ? "Nonaktifkan" : "Aktifkan"}
                              </button>
                              <button
                                type="button"
                                onClick={() => setRemoveTarget(r)}
                                className={`${styles.btnDanger} ${styles.btnSmall}`}
                              >
                                Cabut
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

      <SlideOverDrawer
        open={Boolean(selected)}
        title={selected ? `Peran ${selected.role}` : "Detail Peran Mitra"}
        subtitle={selected?.id}
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <EfCodePill code={selected.role} />
              <ActiveBooleanBadge active={selected.is_active} />
              <EfRelayNodeInspectorButton
                gid={selected.id}
                onInspect={fetchEfPartnerNode}
              />
            </div>

            {canWrite ? (
              <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
                <label
                  htmlFor="pr-note-edit"
                  className="block text-xs font-semibold text-[#0b2439]"
                >
                  Perbarui Catatan (`updatePartyRoleNote`)
                </label>
                <div className="flex gap-2">
                  <input
                    id="pr-note-edit"
                    type="text"
                    value={editNote}
                    onChange={(e) => setEditNote(e.target.value)}
                    className={styles.fieldInput}
                  />
                  <button
                    type="button"
                    onClick={() => void handleUpdateNote()}
                    className={`${styles.btnPrimary} ${styles.btnSmall}`}
                  >
                    Simpan
                  </button>
                </div>
              </div>
            ) : null}

            <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
              <div className="flex items-center justify-between gap-2">
                <h3 className="text-xs font-semibold text-[#0b2439]">
                  Validasi Scope (`validatePartyRoleScope`)
                </h3>
                <button
                  type="button"
                  onClick={() => void handleValidatePartyScope(selected)}
                  className={`${styles.btnSecondary} ${styles.btnSmall}`}
                >
                  Validasi
                </button>
              </div>
              <EfScopeDecisionBox decision={decision} error={decisionErr} />
            </div>
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(removeTarget)}
        title="Cabut Peran Mitra?"
        description={`Peran ${removeTarget?.role} akan dicabut dari party terkait.`}
        confirmLabel="Cabut Peran"
        onCancel={() => setRemoveTarget(null)}
        onConfirm={() => void handleRemove()}
      />
    </div>
  );
}
