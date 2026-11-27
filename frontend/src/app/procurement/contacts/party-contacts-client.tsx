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
  EfRelayNodeInspectorButton,
  EfStatusBadge,
  EfToolbarCard,
} from "@/components/workspace/ef-ui-primitives";
import {
  PORTAL_ACCESS_STATUS_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import {
  activatePartyContact,
  activatePartyPortalAccess,
  addPartyContact,
  deactivatePartyContact,
  deletePartyContact,
  requestPartyPortalLink,
  revokePartyPortalAccess,
  setPrimaryPartyContact,
  unlinkPartyPortalUser,
  updatePartyContact,
} from "@/lib/graphql/ef-partner-mutations";
import {
  evaluateRepresentativePortalAccess,
  fetchAllParties,
  fetchAllPartyContacts,
  fetchEfPartnerNode,
  fetchPartyContact,
  fetchPartyContacts,
  fetchPartyContactsByEmail,
  fetchPartyContactsByIamUserId,
  fetchPartyContactsByParty,
  fetchPrimaryPartyContact,
} from "@/lib/graphql/ef-partner-queries";
import type {
  EfParty,
  EfPartyContact,
  PartyRoleKind,
  PortalAccessEvaluation,
  PortalAccessStatus,
} from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";

export default function PartyContactsClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(snapshot, "enterprise_foundation.parties.write");

  const [items, setItems] = useState<EfPartyContact[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [parties, setParties] = useState<EfParty[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [partyFilter, setPartyFilter] = useState("");
  const [portalFilter, setPortalFilter] = useState<PortalAccessStatus | "">("");
  const [allMode, setAllMode] = useState(false);

  const [lookupMode, setLookupMode] = useState<
    "BY_PARTY" | "PRIMARY_BY_PARTY" | "BY_IAM_USER" | "BY_EMAIL" | "BY_ID"
  >("BY_EMAIL");
  const [lookupVal, setLookupVal] = useState("");

  // Form tambah/ubah kontak PIC
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<EfPartyContact | null>(null);
  const [cPartyId, setCPartyId] = useState("");
  const [cName, setCName] = useState("");
  const [cRole, setCRole] = useState("ACCOUNT_MANAGER");
  const [cJobTitle, setCJobTitle] = useState("");
  const [cEmail, setCEmail] = useState("");
  const [cPhone, setCPhone] = useState("");
  const [cPrimary, setCPrimary] = useState(false);

  // Drawer portal IAM & lifecycle kontak
  const [selected, setSelected] = useState<EfPartyContact | null>(null);
  const [iamUserInput, setIamUserInput] = useState("");
  const [evalRole, setEvalRole] = useState<PartyRoleKind>("VENDOR");
  const [portalEval, setPortalEval] = useState<PortalAccessEvaluation | null>(
    null,
  );
  const [deleteTarget, setDeleteTarget] = useState<EfPartyContact | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllPartyContacts(20, cursor?.after)
        : await fetchPartyContacts({
            query: query.trim() || undefined,
            party_id: partyFilter.trim() || null,
            portal_access_status: portalFilter || null,
            first: 20,
            after: cursor?.after || null,
            before: cursor?.before || null,
          });
    },
    [allMode, partyFilter, portalFilter, query],
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
          setCPartyId((prev) => prev || firstPartyId);
        }
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, []);

  async function handleExactLookup() {
    const val = lookupVal.trim();
    if (!val) return;
    setLoading(true);
    try {
      if (lookupMode === "BY_ID") {
        const one = await fetchPartyContact(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "PRIMARY_BY_PARTY") {
        const one = await fetchPrimaryPartyContact(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "BY_PARTY") {
        const conn = await fetchPartyContactsByParty(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else if (lookupMode === "BY_IAM_USER") {
        const conn = await fetchPartyContactsByIamUserId(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else {
        const conn = await fetchPartyContactsByEmail(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      }
    } catch (err) {
      pushToast("error", "Lookup PIC Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  function openContactForm(existing: EfPartyContact | null) {
    setEditing(existing);
    setCPartyId(existing?.party_id || parties[0]?._id || "");
    setCName(existing?.full_name || "");
    setCRole(existing?.contact_role || "ACCOUNT_MANAGER");
    setCJobTitle(existing?.job_title || "");
    setCEmail(existing?.email || "");
    setCPhone(existing?.phone || "");
    setCPrimary(Boolean(existing?.is_primary));
    setModalOpen(true);
  }

  async function handleSaveContact(e: FormEvent) {
    e.preventDefault();
    if (!cName.trim() || !cRole.trim()) return;
    if (editing) {
      const updated = await runDashboardAction(() =>
        updatePartyContact(
          {
            id: editing.id,
            full_name: cName.trim(),
            contact_role: cRole.trim(),
            job_title: cJobTitle.trim() || null,
            email: cEmail.trim() || null,
            phone: cPhone.trim() || null,
            is_primary: cPrimary,
          },
          csrfToken,
        ),
      );
      if (updated) {
        setModalOpen(false);
        pushToast("success", "PIC Diperbarui", updated.full_name);
        await loadList();
      }
    } else if (cPartyId.trim()) {
      const added = await runDashboardAction(() =>
        addPartyContact(
          {
            party_id: cPartyId.trim(),
            full_name: cName.trim(),
            contact_role: cRole.trim(),
            job_title: cJobTitle.trim() || null,
            email: cEmail.trim() || null,
            phone: cPhone.trim() || null,
            is_primary: cPrimary,
          },
          csrfToken,
        ),
      );
      if (added) {
        setModalOpen(false);
        pushToast("success", "PIC Ditambahkan", added.full_name);
        await loadList();
      }
    }
  }

  async function runPortalAction(
    label: string,
    fn: () => Promise<EfPartyContact | null>,
    sensitive = false,
  ) {
    const updated = await runDashboardAction(
      fn,
      sensitive ? { stepUpReason: `Aksi "${label}" bersifat sensitif.` } : undefined,
    );
    if (updated) {
      setSelected(updated);
      setIamUserInput(updated.iam_user_id || "");
      pushToast("success", label, updated.full_name);
      await loadList();
    }
  }

  async function handleEvalPortal() {
    if (!selected || !iamUserInput.trim()) return;
    try {
      const res = await evaluateRepresentativePortalAccess(
        selected.party_id,
        iamUserInput.trim(),
        evalRole,
      );
      setPortalEval(res);
    } catch (err) {
      pushToast("error", "Evaluasi Portal Gagal", formatEfError(err));
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const ok = await runDashboardAction(
      () => deletePartyContact(target.id, csrfToken),
      { stepUpReason: "Menghapus kontak PIC mitra bersifat sensitif." },
    );
    if (ok) {
      setDeleteTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "PIC Dihapus", target.full_name);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Kontak PIC & Akses Portal IAM Mitra"
        description="Manajemen PIC (`partyContacts`, `setPrimaryPartyContact`, `request/activate/revoke/unlink` Portal IAM)."
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
                onClick={() => openContactForm(null)}
                className={styles.btnPrimary}
              >
                <IdentityIcon name="plus" size={15} />
                <span>Tambah PIC Mitra</span>
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
            placeholder="Cari nama / email / peran..."
            aria-label="Cari kontak PIC"
            className={styles.fieldInput}
          />
          <input
            type="text"
            value={partyFilter}
            onChange={(e) => setPartyFilter(e.target.value)}
            placeholder="Filter Party ID..."
            aria-label="Filter Party ID PIC"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <select
            aria-label="Filter status portal PIC"
            value={portalFilter}
            onChange={(e) =>
              setPortalFilter(e.target.value as PortalAccessStatus | "")
            }
            className={styles.fieldSelect}
          >
            <option value="">Semua Status Portal</option>
            {PORTAL_ACCESS_STATUS_OPTIONS.map((opt) => (
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
          <label className="flex items-center gap-1.5 text-xs text-[#0b2439]">
            <input
              type="checkbox"
              checked={allMode}
              onChange={(e) => setAllMode(e.target.checked)}
            />
            <span>Mode `allPartyContacts`</span>
          </label>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup PIC"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(
                  e.target.value as
                    | "BY_PARTY"
                    | "PRIMARY_BY_PARTY"
                    | "BY_IAM_USER"
                    | "BY_EMAIL"
                    | "BY_ID",
                )
              }
              className={`${styles.fieldSelect} w-44`}
            >
              <option value="BY_EMAIL">By Email</option>
              <option value="BY_IAM_USER">By IAM User ID</option>
              <option value="BY_PARTY">By Party ID</option>
              <option value="PRIMARY_BY_PARTY">Primary Contact</option>
              <option value="BY_ID">By Contact ID</option>
            </select>
            <input
              type="text"
              value={lookupVal}
              onChange={(e) => setLookupVal(e.target.value)}
              placeholder="Nilai lookup..."
              aria-label="Nilai lookup PIC"
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
            title="Belum Ada PIC Mitra"
            description="Tambahkan PIC perwakilan mitra lalu hubungkan ke akun IAM bila memerlukan akses portal."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>PIC &amp; Kontak</th>
                    <th>Peran PIC</th>
                    <th>Status Portal IAM</th>
                    <th>Status</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((c) => (
                    <tr key={c.id}>
                      <td>
                        <p className="font-semibold text-[#0b2439]">
                          {c.full_name}{" "}
                          {c.is_primary ? (
                            <span className="ml-1 rounded bg-[#e9eef2] px-1.5 py-0.5 text-[10px] text-[#133554]">
                              UTAMA
                            </span>
                          ) : null}
                        </p>
                        <p className="text-xs text-[#5a6672]">
                          {c.email || "-"} · {c.phone || "-"}
                        </p>
                      </td>
                      <td className="text-xs text-[#0b2439]">{c.contact_role}</td>
                      <td>
                        <EfStatusBadge label={c.portal_access_status} />
                      </td>
                      <td>
                        <ActiveBooleanBadge active={c.is_active} />
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => {
                              setSelected(c);
                              setIamUserInput(c.iam_user_id || "");
                              setPortalEval(null);
                            }}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Portal &amp; Aksi
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => openContactForm(c)}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Ubah
                              </button>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(c)}
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

      {modalOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
          role="dialog"
          aria-modal="true"
          aria-label="Form Kontak PIC"
        >
          <div className="w-full max-w-lg rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
            <h2 className="text-lg font-bold text-[#0b2439]">
              {editing ? "Perbarui PIC Mitra" : "Tambah PIC Mitra"}
            </h2>
            <form onSubmit={(e) => void handleSaveContact(e)} className="mt-4 space-y-3">
              {!editing ? (
                <div>
                  <label htmlFor="pc-party" className="block text-xs font-semibold text-[#0b2439]">
                    Mitra (Party) *
                  </label>
                  {parties.length > 0 ? (
                    <select
                      id="pc-party"
                      value={cPartyId}
                      onChange={(e) => setCPartyId(e.target.value)}
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
                      id="pc-party"
                      type="text"
                      required
                      value={cPartyId}
                      onChange={(e) => setCPartyId(e.target.value)}
                      className={`${styles.fieldInput} ${styles.monoText}`}
                    />
                  )}
                </div>
              ) : null}
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label htmlFor="pc-name" className="block text-xs font-semibold text-[#0b2439]">
                    Nama Lengkap *
                  </label>
                  <input
                    id="pc-name"
                    type="text"
                    required
                    value={cName}
                    onChange={(e) => setCName(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
                <div>
                  <label htmlFor="pc-role" className="block text-xs font-semibold text-[#0b2439]">
                    Peran Kontak *
                  </label>
                  <input
                    id="pc-role"
                    type="text"
                    required
                    value={cRole}
                    onChange={(e) => setCRole(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
              </div>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label htmlFor="pc-email" className="block text-xs font-semibold text-[#0b2439]">
                    Email
                  </label>
                  <input
                    id="pc-email"
                    type="email"
                    value={cEmail}
                    onChange={(e) => setCEmail(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
                <div>
                  <label htmlFor="pc-phone" className="block text-xs font-semibold text-[#0b2439]">
                    Telepon
                  </label>
                  <input
                    id="pc-phone"
                    type="text"
                    value={cPhone}
                    onChange={(e) => setCPhone(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
              </div>
              <div>
                <label htmlFor="pc-job" className="block text-xs font-semibold text-[#0b2439]">
                  Jabatan
                </label>
                <input
                  id="pc-job"
                  type="text"
                  value={cJobTitle}
                  onChange={(e) => setCJobTitle(e.target.value)}
                  className={styles.fieldInput}
                />
              </div>
              <label className="flex items-center gap-2 text-xs text-[#0b2439]">
                <input
                  type="checkbox"
                  checked={cPrimary}
                  onChange={(e) => setCPrimary(e.target.checked)}
                />
                <span>Tetapkan sebagai PIC Utama</span>
              </label>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setModalOpen(false)}
                  className={styles.btnSecondary}
                >
                  Batal
                </button>
                <button type="submit" className={styles.btnPrimary}>
                  Simpan PIC
                </button>
              </div>
            </form>
          </div>
        </div>
      ) : null}

      <SlideOverDrawer
        open={Boolean(selected)}
        title={selected?.full_name || "Kelola Portal IAM PIC"}
        subtitle={selected?.id}
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <div className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <EfStatusBadge label={selected.portal_access_status} />
              <ActiveBooleanBadge active={selected.is_active} />
              <EfRelayNodeInspectorButton
                gid={selected.id}
                onInspect={fetchEfPartnerNode}
              />
            </div>

            {canWrite ? (
              <div className="space-y-3 rounded-xl border border-[#dde3e8] p-4">
                <label
                  htmlFor="pc-iam-uid"
                  className="block text-xs font-semibold text-[#0b2439]"
                >
                  IAM User UUID/GID Tautan Portal
                </label>
                <input
                  id="pc-iam-uid"
                  type="text"
                  value={iamUserInput}
                  onChange={(e) => setIamUserInput(e.target.value)}
                  placeholder="UUID / gid://integrione/user/..."
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() =>
                      void runPortalAction("Tautan Portal Diminta", () =>
                        requestPartyPortalLink(
                          selected.id,
                          iamUserInput.trim(),
                          csrfToken,
                        ),
                      )
                    }
                    className={`${styles.btnSecondary} ${styles.btnSmall}`}
                  >
                    Request Link
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      void runPortalAction(
                        "Akses Portal Diaktifkan",
                        () =>
                          activatePartyPortalAccess(
                            selected.id,
                            iamUserInput.trim(),
                            csrfToken,
                          ),
                        true,
                      )
                    }
                    className={`${styles.btnPrimary} ${styles.btnSmall}`}
                  >
                    Activate Portal (Sensitif)
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      void runPortalAction(
                        "Akses Portal Dicabut",
                        () => revokePartyPortalAccess(selected.id, csrfToken),
                        true,
                      )
                    }
                    className={`${styles.btnDanger} ${styles.btnSmall}`}
                  >
                    Revoke Portal (Sensitif)
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      void runPortalAction(
                        "Tautan IAM Dilepas",
                        () => unlinkPartyPortalUser(selected.id, csrfToken),
                        true,
                      )
                    }
                    className={`${styles.btnDanger} ${styles.btnSmall}`}
                  >
                    Unlink User (Sensitif)
                  </button>
                </div>

                <div className="flex flex-wrap gap-2 border-t border-[#dde3e8] pt-3">
                  <button
                    type="button"
                    onClick={() =>
                      void runPortalAction("PIC Dijadikan Utama", () =>
                        setPrimaryPartyContact(
                          selected.party_id,
                          selected.id,
                          csrfToken,
                        ),
                      )
                    }
                    className={`${styles.btnSecondary} ${styles.btnSmall}`}
                  >
                    Set Utama
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      void runPortalAction(
                        selected.is_active
                          ? "PIC Dinonaktifkan"
                          : "PIC Diaktifkan",
                        () =>
                          selected.is_active
                            ? deactivatePartyContact(selected.id, csrfToken)
                            : activatePartyContact(selected.id, csrfToken),
                      )
                    }
                    className={`${styles.btnSecondary} ${styles.btnSmall}`}
                  >
                    {selected.is_active ? "Nonaktifkan PIC" : "Aktifkan PIC"}
                  </button>
                </div>
              </div>
            ) : null}

            <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
              <div className="flex flex-wrap items-center gap-2">
                <select
                  aria-label="Required role evaluasi portal PIC"
                  value={evalRole}
                  onChange={(e) => setEvalRole(e.target.value as PartyRoleKind)}
                  className={`${styles.fieldSelect} w-40`}
                >
                  <option value="VENDOR">VENDOR</option>
                  <option value="SUPPLIER">SUPPLIER</option>
                </select>
                <button
                  type="button"
                  onClick={() => void handleEvalPortal()}
                  className={`${styles.btnSecondary} ${styles.btnSmall}`}
                >
                  Evaluasi Portal (`evaluateRepresentativePortalAccess`)
                </button>
              </div>
              {portalEval ? (
                <pre className={`rounded bg-[#f9fafb] p-2.5 text-xs ${styles.monoText}`}>
                  {JSON.stringify(portalEval, null, 2)}
                </pre>
              ) : null}
            </div>
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Hapus PIC Mitra?"
        description={`Kontak "${deleteTarget?.full_name}" akan di-soft-delete.`}
        confirmLabel="Hapus PIC"
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDelete()}
      />
    </div>
  );
}
