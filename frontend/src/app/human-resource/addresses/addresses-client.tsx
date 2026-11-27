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
  SlideOverDrawer,
  TableSkeleton,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  EfCodePill,

  EfToolbarCard,
} from "@/components/workspace/ef-ui-primitives";
import { formatEfError, hasEfPermission } from "@/lib/graphql/ef-helpers";
import {
  fetchAddress,
  fetchAddresses,
  fetchAddressesByIds,
  fetchAllAddresses,

  fetchIsAddressReferenced,
} from "@/lib/graphql/ef-structure-queries";
import { deleteAddress } from "@/lib/graphql/ef-structure-mutations";
import type { EfAddress } from "@/lib/graphql/ef-types";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { AddressFormModal } from "./address-form-modal";

export default function AddressesClient() {
  const { snapshot, csrfToken, pushToast, runDashboardAction } =
    useIdentityWorkspace();

  const canWrite = hasEfPermission(snapshot, "enterprise_foundation.addresses.write");

  const [items, setItems] = useState<EfAddress[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadErr, setLoadErr] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [countryFilter, setCountryFilter] = useState("");
  const [cityFilter, setCityFilter] = useState("");
  const [provinceFilter, setProvinceFilter] = useState("");
  const [postalFilter, setPostalFilter] = useState("");
  const [allMode, setAllMode] = useState(false);

  const [batchIdsInput, setBatchIdsInput] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<EfAddress | null>(null);
  const [selected, setSelected] = useState<EfAddress | null>(null);
  const [selectedReferenced, setSelectedReferenced] = useState<boolean | null>(
    null,
  );
  const [deleteTarget, setDeleteTarget] = useState<EfAddress | null>(null);

  const fetchCurrentList = useCallback(
    async (cursor?: { after?: string | null; before?: string | null }) => {
      return allMode
        ? await fetchAllAddresses(20, cursor?.after)
        : await fetchAddresses({
            query: query.trim() || undefined,
            country_code: countryFilter.trim() || undefined,
            city: cityFilter.trim() || undefined,
            province: provinceFilter.trim() || undefined,
            postal_code: postalFilter.trim() || undefined,
            first: 20,
            after: cursor?.after || null,
            before: cursor?.before || null,
          });
    },
    [allMode, cityFilter, countryFilter, postalFilter, provinceFilter, query],
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

  async function handleBatchLookup() {
    const ids = batchIdsInput
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    if (ids.length === 0) return;
    setLoading(true);
    try {
      const found = await fetchAddressesByIds(ids);
      setItems(found);
      setTotalCount(found.length);
      setPageInfo(null);
    } catch (err) {
      pushToast("error", "Lookup Batch Gagal", formatEfError(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleInspectAddress(addr: EfAddress) {
    setSelected(addr);
    setSelectedReferenced(null);
    try {
      const [fresh, ref] = await Promise.all([
        fetchAddress(addr.id),
        fetchIsAddressReferenced(addr.id),
      ]);
      if (fresh) setSelected(fresh);
      setSelectedReferenced(ref);
    } catch (err) {
      pushToast("error", "Gagal Memuat Detail", formatEfError(err));
    }
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const deleted = await runDashboardAction(
      () => deleteAddress(target.id, csrfToken),
      { stepUpReason: "Menghapus master alamat bersifat sensitif." },
    );
    if (deleted) {
      setDeleteTarget(null);
      if (selected?.id === target.id) setSelected(null);
      pushToast("success", "Alamat Dihapus", target.address_line_1);
      await loadList();
    }
  }

  return (
    <div className="space-y-5">
      <EfToolbarCard
        title="Master Alamat Geografis"
        description="Daftar alamat bersama (`addresses`, `allAddresses`, `findOrCreateAddress`, `isAddressReferenced`)."
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
                <span>Tambah Alamat</span>
              </button>
            ) : null}
          </>
        }
      >
        <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-6">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cari jalan / label..."
            aria-label="Cari alamat"
            className={`${styles.fieldInput} lg:col-span-2`}
          />
          <input
            type="text"
            value={cityFilter}
            onChange={(e) => setCityFilter(e.target.value)}
            placeholder="Kota"
            aria-label="Filter kota"
            className={styles.fieldInput}
          />
          <input
            type="text"
            value={provinceFilter}
            onChange={(e) => setProvinceFilter(e.target.value)}
            placeholder="Provinsi"
            aria-label="Filter provinsi"
            className={styles.fieldInput}
          />
          <input
            type="text"
            value={postalFilter}
            onChange={(e) => setPostalFilter(e.target.value)}
            placeholder="Kode Pos"
            aria-label="Filter kode pos"
            className={styles.fieldInput}
          />
          <div className="flex gap-2">
            <input
              type="text"
              value={countryFilter}
              onChange={(e) => setCountryFilter(e.target.value.toUpperCase())}
              maxLength={2}
              placeholder="ID"
              aria-label="Filter negara"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
            <button
              type="button"
              onClick={() => void loadList()}
              className={styles.btnPrimary}
            >
              Filter
            </button>
          </div>
        </div>

        <div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-[#dde3e8] pt-3">
          <label className="flex items-center gap-2 text-xs text-[#0b2439]">
            <input
              type="checkbox"
              checked={allMode}
              onChange={(e) => setAllMode(e.target.checked)}
            />
            <span>Gunakan query `allAddresses` tanpa filter</span>
          </label>

          <div className="flex flex-wrap items-center gap-2">
            <input
              type="text"
              value={batchIdsInput}
              onChange={(e) => setBatchIdsInput(e.target.value)}
              placeholder="Lookup `addressesByIds`: id1, id2"
              aria-label="Lookup batch ID alamat"
              className={`${styles.fieldInput} w-60 ${styles.monoText}`}
            />
            <button
              type="button"
              onClick={() => void handleBatchLookup()}
              className={`${styles.btnSecondary} ${styles.btnSmall}`}
            >
              Cari ID Batch
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
            title="Belum Ada Data Alamat"
            description="Tambahkan alamat baru atau ubah kata kunci filter Anda."
          />
        ) : (
          <>
            <div className={styles.tableWrap}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Label / Jalan</th>
                    <th>Kota &amp; Provinsi</th>
                    <th>Kode Pos</th>
                    <th>Negara</th>
                    <th className="text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((addr) => (
                    <tr key={addr.id}>
                      <td>
                        <button
                          type="button"
                          onClick={() => void handleInspectAddress(addr)}
                          className="text-left hover:underline"
                        >
                          <p className="font-semibold text-[#0b2439]">
                            {addr.label || addr.address_line_1}
                          </p>
                          <p className="text-xs text-[#5a6672]">
                            {addr.address_line_1}
                          </p>
                        </button>
                      </td>
                      <td>
                        <p className="text-sm text-[#17202a]">{addr.city}</p>
                        <p className="text-xs text-[#5a6672]">
                          {addr.province || "-"}
                        </p>
                      </td>
                      <td className={styles.monoText}>{addr.postal_code || "-"}</td>
                      <td>
                        <EfCodePill code={addr.country_code} />
                      </td>
                      <td>
                        <div className="flex justify-end gap-1.5">
                          <button
                            type="button"
                            onClick={() => void handleInspectAddress(addr)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Detail
                          </button>
                          {canWrite ? (
                            <>
                              <button
                                type="button"
                                onClick={() => {
                                  setEditing(addr);
                                  setFormOpen(true);
                                }}
                                className={`${styles.btnSecondary} ${styles.btnSmall}`}
                              >
                                Ubah
                              </button>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(addr)}
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
        <AddressFormModal
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
        title={selected?.label || selected?.address_line_1 || "Detail Alamat"}
        subtitle={selected?.city}
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <EfCodePill code={selected.country_code} />
              {selectedReferenced !== null ? (
                <span
                  className={`rounded-md border px-2 py-0.5 text-xs font-semibold ${
                    selectedReferenced
                      ? "border-[#fedf89] bg-[#fffaeb] text-[#b54708]"
                      : "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                  }`}
                >
                  {selectedReferenced
                    ? "Sedang Direferensikan Entitas Lain"
                    : "Belum Direferensikan (Aman Dihapus)"}
                </span>
              ) : null}

            </div>

            <dl className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-4 text-xs sm:grid-cols-2">
              <div>
                <dt className="font-semibold text-[#5a6672]">Baris 1</dt>
                <dd className="mt-0.5 text-[#0b2439]">{selected.address_line_1}</dd>
              </div>
              <div>
                <dt className="font-semibold text-[#5a6672]">Baris 2</dt>
                <dd className="mt-0.5 text-[#0b2439]">
                  {selected.address_line_2 || "-"}
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-[#5a6672]">Kota / Provinsi</dt>
                <dd className="mt-0.5 text-[#0b2439]">
                  {selected.city}, {selected.province || "-"}
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-[#5a6672]">Koordinat</dt>
                <dd className={`mt-0.5 text-[#0b2439] ${styles.monoText}`}>
                  {selected.latitude ?? "-"}, {selected.longitude ?? "-"}
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-[#5a6672]">Diperbarui</dt>
                <dd className="mt-0.5 text-[#0b2439]">
                  {formatDateTimeID(selected.updated_at)}
                </dd>
              </div>
            </dl>
          </div>
        ) : null}
      </SlideOverDrawer>

      <ConfirmModal
        open={Boolean(deleteTarget)}
        title="Hapus Master Alamat?"
        description={`Alamat "${deleteTarget?.address_line_1}" akan dihapus permanen bila tidak lagi direferensikan oleh badan hukum, lokasi, maupun mitra.`}
        confirmLabel="Hapus Alamat"
        onCancel={() => setDeleteTarget(null)}
        onConfirm={() => void handleDelete()}
      />
    </div>
  );
}
