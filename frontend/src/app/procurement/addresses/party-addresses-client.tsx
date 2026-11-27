"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
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
  EfStatusBadge,
  EfToolbarCard,
} from "@/components/workspace/ef-ui-primitives";
import {
  ADDRESS_USAGE_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import {
  linkPartyAddress,
  setPrimaryPartyAddressLink,
  unlinkPartyAddress,
  updatePartyAddressLink,
} from "@/lib/graphql/ef-partner-mutations";
import {
  fetchAllParties,
  fetchAllPartyAddressLinks,
  fetchEfPartnerNode,
  fetchPartyAddressLink,
  fetchPartyAddressLinks,
  fetchPartyAddressLinksByAddressId,
  fetchPartyAddressLinksByParty,
  fetchPartyAddressLinksByPartyAndUsage,
  fetchPrimaryPartyAddressLink,
} from "@/lib/graphql/ef-partner-queries";
import { fetchAllAddresses } from "@/lib/graphql/ef-structure-queries";
import type {
  AddressUsageKind,
  EfAddress,
  EfParty,
  EfPartyAddress,
} from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";

export default function PartyAddressesClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(snapshot, "enterprise_foundation.parties.write");

  const [items, setItems] = useState<EfPartyAddress[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [parties, setParties] = useState<EfParty[]>([]);
  const [addresses, setAddresses] = useState<EfAddress[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [partyFilter, setPartyFilter] = useState("");
  const [addrFilter, setAddrFilter] = useState("");
  const [usageFilter, setUsageFilter] = useState<AddressUsageKind | "">("");
  const [primaryOnly, setPrimaryOnly] = useState(false);
  const [allMode, setAllMode] = useState(false);

  const [lookupMode, setLookupMode] = useState<
    "BY_PARTY" | "PARTY_AND_USAGE" | "PRIMARY_USAGE" | "BY_ADDR" | "BY_ID"
  >("BY_PARTY");
  const [lookupValue, setLookupValue] = useState("");
  const [lookupUsage, setLookupUsage] = useState<AddressUsageKind>("BILLING");

  // Form link alamat mitra
  const [linkPartyId, setLinkPartyId] = useState("");
  const [linkAddressId, setLinkAddressId] = useState("");
  const [linkUsage, setLinkUsage] = useState<AddressUsageKind>("BILLING");
  const [linkPrimary, setLinkPrimary] = useState(false);

  const [selected, setSelected] = useState<EfPartyAddress | null>(null);
  const [editUsage, setEditUsage] = useState<AddressUsageKind>("BILLING");
  const [editPrimary, setEditPrimary] = useState(false);
  const [unlinkTarget, setUnlinkTarget] = useState<EfPartyAddress | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllPartyAddressLinks(20, cursor?.after)
        : await fetchPartyAddressLinks({
            party_id: partyFilter.trim() || null,
            address_id: addrFilter.trim() || null,
            usage: usageFilter || null,
            primary_only: primaryOnly ? true : null,
            first: 20,
            after: cursor?.after || null,
            before: cursor?.before || null,
          });
    },
    [addrFilter, allMode, partyFilter, primaryOnly, usageFilter],
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
    void Promise.all([fetchAllParties(25), fetchAllAddresses(25)])
      .then(([pConn, aConn]) => {
        if (!active) return;
        setParties(pConn.nodes);
        setAddresses(aConn.nodes);
        const firstPartyId = pConn.nodes[0]?._id;
        if (firstPartyId) {
          setLinkPartyId((prev) => prev || firstPartyId);
        }
        const firstAddrId = aConn.nodes[0]?._id;
        if (firstAddrId) {
          setLinkAddressId((prev) => prev || firstAddrId);
        }
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, []);

  async function handleExactLookup() {
    const val = lookupValue.trim();
    if (!val) return;
    setLoading(true);
    try {
      if (lookupMode === "BY_ID") {
        const one = await fetchPartyAddressLink(val);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "PRIMARY_USAGE") {
        const one = await fetchPrimaryPartyAddressLink(val, lookupUsage);
        setItems(one ? [one] : []);
        setTotalCount(one ? 1 : 0);
        setPageInfo(null);
      } else if (lookupMode === "PARTY_AND_USAGE") {
        const conn = await fetchPartyAddressLinksByPartyAndUsage(
          val,
          lookupUsage,
          20,
        );
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else if (lookupMode === "BY_ADDR") {
        const conn = await fetchPartyAddressLinksByAddressId(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      } else {
        const conn = await fetchPartyAddressLinksByParty(val, 20);
        setItems(conn.nodes);
        setTotalCount(conn.total_count);
        setPageInfo(conn.pageInfo);
      }
    } catch (err) {
      pushToast("error", "Lookup Alamat Mitra Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleLink(e: FormEvent) {
    e.preventDefault();
    if (!linkPartyId.trim() || !linkAddressId.trim()) return;
    const created = await runDashboardAction(() =>
      linkPartyAddress(
        {
          party_id: linkPartyId.trim(),
          address_id: linkAddressId.trim(),
          usage: linkUsage,
          is_primary: linkPrimary,
          replace_primary: true,
        },
        csrfToken,
      ),
    );
    if (created) {
      pushToast("success", "Alamat Ditautkan", created.usage);
      await loadList();
    }
  }

  async function handleUpdateLink() {
    if (!selected) return;
    const updated = await runDashboardAction(() =>
      updatePartyAddressLink(
        {
          id: selected.id,
          usage: editUsage,
          is_primary: editPrimary,
          replace_primary: true,
        },
        csrfToken,
      ),
    );
    if (updated) {
      setSelected(updated);
      pushToast("success", "Tautan Diperbarui", updated.usage);
      await loadList();
    }
  }

  async function handleSetPrimary(link: EfPartyAddress) {
    const updated = await runDashboardAction(() =>
      setPrimaryPartyAddressLink(link.id, csrfToken),
    );
    if (updated) {
      pushToast("success", "Alamat Utama Ditetapkan", updated.usage);
      await loadList();
      if (selected?.id === updated.id) setSelected(updated);
    }
  }

  async function handleUnlink() {
    if (!unlinkTarget) return;
    const target = unlinkTarget;
    const ok = await runDashboardAction(() =>
      unlinkPartyAddress(target.id, csrfToken),
    );
    if (ok) {
      setUnlinkTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "Tautan Alamat Dilepas", target.usage);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Relasi Alamat Mitra (`PartyAddress`)"
        description="Penautan alamat billing, shipping, HQ, & kantor terdaftar milik mitra eksternal."
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
            value={partyFilter}
            onChange={(e) => setPartyFilter(e.target.value)}
            placeholder="Filter Party ID..."
            aria-label="Filter ID party tautan alamat"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <input
            type="text"
            value={addrFilter}
            onChange={(e) => setAddrFilter(e.target.value)}
            placeholder="Filter Address ID..."
            aria-label="Filter ID master alamat"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <select
            aria-label="Filter peruntukan alamat"
            value={usageFilter}
            onChange={(e) => setUsageFilter(e.target.value as AddressUsageKind | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Peruntukan</option>
            {ADDRESS_USAGE_OPTIONS.map((opt) => (
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
                checked={allMode}
                onChange={(e) => setAllMode(e.target.checked)}
              />
              <span>Mode `allPartyAddressLinks`</span>
            </label>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <select
              aria-label="Metode lookup alamat mitra"
              value={lookupMode}
              onChange={(e) =>
                setLookupMode(
                  e.target.value as
                    | "BY_PARTY"
                    | "PARTY_AND_USAGE"
                    | "PRIMARY_USAGE"
                    | "BY_ADDR"
                    | "BY_ID",
                )
              }
              className={`${styles.fieldSelect} w-44`}
            >
              <option value="BY_PARTY">By Party ID</option>
              <option value="PARTY_AND_USAGE">By Party + Usage</option>
              <option value="PRIMARY_USAGE">Primary Usage</option>
              <option value="BY_ADDR">By Address ID</option>
              <option value="BY_ID">By Link ID</option>
            </select>
            <input
              type="text"
              value={lookupValue}
              onChange={(e) => setLookupValue(e.target.value)}
              placeholder="Party / Address / Link ID..."
              aria-label="Nilai lookup tautan alamat"
              className={`${styles.fieldInput} w-44 ${styles.monoText}`}
            />
            {lookupMode === "PARTY_AND_USAGE" || lookupMode === "PRIMARY_USAGE" ? (
              <select
                aria-label="Peruntukan lookup"
                value={lookupUsage}
                onChange={(e) =>
                  setLookupUsage(e.target.value as AddressUsageKind)
                }
                className={`${styles.fieldSelect} w-36`}
              >
                {ADDRESS_USAGE_OPTIONS.map((opt) => (
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
          onSubmit={(e) => void handleLink(e)}
          className={`${styles.hairlinePanel} grid grid-cols-1 gap-2.5 p-4 sm:grid-cols-5`}
        >
          {parties.length > 0 ? (
            <select
              aria-label="Mitra tautan alamat"
              value={linkPartyId}
              onChange={(e) => setLinkPartyId(e.target.value)}
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
              value={linkPartyId}
              onChange={(e) => setLinkPartyId(e.target.value)}
              placeholder="Party ID"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
          )}

          {addresses.length > 0 ? (
            <select
              aria-label="Master alamat ditautkan"
              value={linkAddressId}
              onChange={(e) => setLinkAddressId(e.target.value)}
              className={styles.fieldSelect}
            >
              {addresses.map((a) => (
                <option key={a.id} value={a._id}>
                  {a.label || a.address_line_1} ({a.city})
                </option>
              ))}
            </select>
          ) : (
            <input
              type="text"
              value={linkAddressId}
              onChange={(e) => setLinkAddressId(e.target.value)}
              placeholder="Address ID"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
          )}

          <select
            aria-label="Peruntukan alamat baru"
            value={linkUsage}
            onChange={(e) => setLinkUsage(e.target.value as AddressUsageKind)}
            className={styles.fieldSelect}
          >
            {ADDRESS_USAGE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>

          <label className="flex items-center gap-2 text-xs text-[#0b2439]">
            <input
              type="checkbox"
              checked={linkPrimary}
              onChange={(e) => setLinkPrimary(e.target.checked)}
            />
            <span>Utama per peruntukan</span>
          </label>

          <button type="submit" className={styles.btnPrimary}>
            Tautkan Alamat
          </button>
        </form>
      ) : null}

      {loadErr ? <ErrorPanel message={loadErr} onRetry={() => void loadList()} /> : null}

      <div className={styles.hairlinePanel}>
        {loading ? (
          <TableSkeleton rows={5} cols={4} />
        ) : items.length === 0 ? (
          <EmptyPanel
            title="Belum Ada Relasi Alamat Mitra"
            description="Tautkan alamat master ke mitra sesuai kebutuhan billing/shipping/registered office."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Peruntukan</th>
                    <th>Alamat</th>
                    <th>Primary</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((pa) => (
                    <tr key={pa.id}>
                      <td>
                        <EfCodePill code={pa.usage} />
                      </td>
                      <td>
                        <p className="text-sm font-semibold text-[#0b2439]">
                          {pa.address?.address_line_1 || pa.address_id}
                        </p>
                        <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                          {pa.address ? `${pa.address.city}, ${pa.address.country_code}` : pa.party_id}
                        </p>
                      </td>
                      <td>
                        {pa.is_primary ? (
                          <EfStatusBadge label="UTAMA" toneSource="ACTIVE" />
                        ) : (
                          <span className="text-xs text-[#5a6672]">Sekunder</span>
                        )}
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => {
                              setSelected(pa);
                              setEditUsage(pa.usage);
                              setEditPrimary(pa.is_primary);
                            }}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Kelola
                          </button>
                          {canWrite ? (
                            <>
                              {!pa.is_primary ? (
                                <button
                                  type="button"
                                  onClick={() => void handleSetPrimary(pa)}
                                  className={`${styles.btnSecondary} ${styles.btnSmall}`}
                                >
                                  Set Utama
                                </button>
                              ) : null}
                              <button
                                type="button"
                                onClick={() => setUnlinkTarget(pa)}
                                className={`${styles.btnDanger} ${styles.btnSmall}`}
                              >
                                Lepas
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
        title={selected ? `Relasi ${selected.usage}` : "Detail Tautan Alamat"}
        subtitle={selected?.id}
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <div className="space-y-4">
            <EfRelayNodeInspectorButton
              gid={selected.id}
              onInspect={fetchEfPartnerNode}
            />
            {canWrite ? (
              <div className="space-y-3 rounded-xl border border-[#dde3e8] p-4">
                <p className="text-xs font-semibold text-[#0b2439]">
                  Perbarui Relasi (`updatePartyAddressLink`)
                </p>
                <select
                  aria-label="Ubah peruntukan alamat"
                  value={editUsage}
                  onChange={(e) =>
                    setEditUsage(e.target.value as AddressUsageKind)
                  }
                  className={styles.fieldSelect}
                >
                  {ADDRESS_USAGE_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <label className="flex items-center gap-2 text-xs text-[#0b2439]">
                  <input
                    type="checkbox"
                    checked={editPrimary}
                    onChange={(e) => setEditPrimary(e.target.checked)}
                  />
                  <span>Utama</span>
                </label>
                <button
                  type="button"
                  onClick={() => void handleUpdateLink()}
                  className={`${styles.btnPrimary} ${styles.btnSmall}`}
                >
                  Simpan Perubahan
                </button>
              </div>
            ) : null}
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(unlinkTarget)}
        title="Lepas Tautan Alamat Mitra?"
        description="Relasi alamat akan dilepas dari mitra terkait tanpa menghapus master alamatnya."
        confirmLabel="Lepas Tautan"
        onCancel={() => setUnlinkTarget(null)}
        onConfirm={() => void handleUnlink()}
      />
    </div>
  );
}
