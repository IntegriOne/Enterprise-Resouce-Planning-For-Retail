"use client";

import { useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { LOCATION_KIND_OPTIONS, formatEfError } from "@/lib/graphql/ef-helpers";
import { createLocation, updateLocation } from "@/lib/graphql/ef-structure-mutations";
import {
  fetchAllAddresses,
  fetchAllLegalEntities,
} from "@/lib/graphql/ef-structure-queries";
import type {
  EfAddress,
  EfLegalEntity,
  EfLocation,
  LocationKind,
} from "@/lib/graphql/ef-types";

export function LocationFormModal({
  open,
  editing,
  onClose,
  onSaved,
}: {
  open: boolean;
  editing: EfLocation | null;
  onClose: () => void;
  onSaved: (loc: EfLocation) => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();

  const [legalEntities, setLegalEntities] = useState<EfLegalEntity[]>([]);
  const [addresses, setAddresses] = useState<EfAddress[]>([]);

  const [legalEntityId, setLegalEntityId] = useState(
    editing?.legal_entity_id || "",
  );
  const [parentLocationId, setParentLocationId] = useState(
    editing?.parent_location_id || "",
  );
  const [addressId, setAddressId] = useState(editing?.address_id || "");
  const [locationCode, setLocationCode] = useState(editing?.location_code || "");
  const [name, setName] = useState(editing?.name || "");
  const [kind, setKind] = useState<LocationKind>(editing?.kind || "WAREHOUSE");
  const [phone, setPhone] = useState(editing?.phone || "");
  const [timezone, setTimezone] = useState(editing?.timezone || "Asia/Jakarta");
  const [attrsJson, setAttrsJson] = useState(
    editing?.operational_attributes_json || "{}",
  );
  const [saving, setSaving] = useState(false);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    let active = true;
    async function loadRefs() {
      try {
        const [leConn, addrConn] = await Promise.all([
          fetchAllLegalEntities(25),
          fetchAllAddresses(25),
        ]);
        if (!active) return;
        setLegalEntities(leConn.nodes);
        setAddresses(addrConn.nodes);
        if (!editing && !legalEntityId && leConn.nodes[0]) {
          setLegalEntityId(leConn.nodes[0]._id);
        }
      } catch {
        // Opsional: user tetap dapat menempel UUID manual jika daftar baca terbatas.
      }
    }
    void loadRefs();
    return () => {
      active = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  if (!open) return null;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (saving) return;
    if (!name.trim() || !timezone.trim() || (!editing && (!legalEntityId.trim() || !locationCode.trim()))) {
      setErrMsg("Badan hukum, kode lokasi, nama, dan zona waktu wajib diisi.");
      return;
    }

    setSaving(true);
    setErrMsg(null);
    try {
      if (editing) {
        const updated = await runDashboardAction(() =>
          updateLocation(
            {
              id: editing.id,
              parent_location_id: parentLocationId.trim() || null,
              address_id: addressId.trim() || null,
              name: name.trim(),
              kind,
              phone: phone.trim() || null,
              timezone: timezone.trim(),
              operational_attributes_json: attrsJson.trim() || "{}",
            },
            csrfToken,
          ),
        );
        if (updated) {
          pushToast("success", "Lokasi Diperbarui", updated.name);
          onSaved(updated);
        }
      } else {
        const created = await runDashboardAction(() =>
          createLocation(
            {
              legal_entity_id: legalEntityId.trim(),
              parent_location_id: parentLocationId.trim() || null,
              address_id: addressId.trim() || null,
              location_code: locationCode.trim().toUpperCase(),
              name: name.trim(),
              kind,
              phone: phone.trim() || null,
              timezone: timezone.trim(),
              operational_attributes_json: attrsJson.trim() || "{}",
            },
            csrfToken,
          ),
        );
        if (created) {
          pushToast("success", "Lokasi Dibuat", created.name);
          onSaved(created);
        }
      }
    } catch (err) {
      setErrMsg(formatEfError(err));
    } finally {
      setSaving(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label={editing ? "Perbarui Lokasi" : "Tambah Lokasi"}
    >
      <div className="max-h-[92dvh] w-full max-w-2xl overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-[#0b2439]">
              {editing ? "Perbarui Lokasi / Fasilitas" : "Tambah Lokasi / Fasilitas"}
            </h2>
            <p className="text-xs text-[#5a6672]">
              Lokasi aktif menjadi target scope `FACILITY` (atau `Warehouse` khusus tipe Gudang).
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup modal lokasi"
            className={styles.iconBtn}
          >
            <IdentityIcon name="close" size={16} />
          </button>
        </div>

        {errMsg ? (
          <div
            role="alert"
            className="mt-3 rounded-lg border border-[#fecdca] bg-[#fef3f2] px-3 py-2 text-xs text-[#b42318]"
          >
            {errMsg}
          </div>
        ) : null}

        <form onSubmit={(e) => void handleSubmit(e)} className="mt-4 space-y-3">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="loc-le" className="block text-xs font-semibold text-[#0b2439]">
                Badan Hukum Pemilik *
              </label>
              {legalEntities.length > 0 && !editing ? (
                <select
                  id="loc-le"
                  value={legalEntityId}
                  onChange={(e) => setLegalEntityId(e.target.value)}
                  className={styles.fieldSelect}
                >
                  <option value="">-- Pilih Badan Hukum --</option>
                  {legalEntities.map((le) => (
                    <option key={le.id} value={le._id}>
                      {le.entity_code} — {le.legal_name}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  id="loc-le"
                  type="text"
                  disabled={Boolean(editing)}
                  value={legalEntityId}
                  onChange={(e) => setLegalEntityId(e.target.value)}
                  placeholder="UUID / GID Legal Entity"
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
              )}
            </div>

            <div>
              <label htmlFor="loc-code" className="block text-xs font-semibold text-[#0b2439]">
                Kode Lokasi *
              </label>
              <input
                id="loc-code"
                type="text"
                disabled={Boolean(editing)}
                required={!editing}
                value={locationCode}
                onChange={(e) => setLocationCode(e.target.value)}
                placeholder="WH-CKG-01"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div className="sm:col-span-2">
              <label htmlFor="loc-name" className="block text-xs font-semibold text-[#0b2439]">
                Nama Lokasi *
              </label>
              <input
                id="loc-name"
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Gudang Pusat Cikarang"
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="loc-kind" className="block text-xs font-semibold text-[#0b2439]">
                Jenis *
              </label>
              <select
                id="loc-kind"
                value={kind}
                onChange={(e) => setKind(e.target.value as LocationKind)}
                className={styles.fieldSelect}
              >
                {LOCATION_KIND_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="loc-addr" className="block text-xs font-semibold text-[#0b2439]">
                Master Alamat (Opsional)
              </label>
              {addresses.length > 0 ? (
                <select
                  id="loc-addr"
                  value={addressId}
                  onChange={(e) => setAddressId(e.target.value)}
                  className={styles.fieldSelect}
                >
                  <option value="">-- Tanpa Alamat Terkait --</option>
                  {addresses.map((a) => (
                    <option key={a.id} value={a._id}>
                      {a.label || a.address_line_1} ({a.city})
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  id="loc-addr"
                  type="text"
                  value={addressId}
                  onChange={(e) => setAddressId(e.target.value)}
                  placeholder="UUID / GID Address"
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
              )}
            </div>

            <div>
              <label htmlFor="loc-parent" className="block text-xs font-semibold text-[#0b2439]">
                Parent Location ID (Opsional)
              </label>
              <input
                id="loc-parent"
                type="text"
                value={parentLocationId}
                onChange={(e) => setParentLocationId(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="loc-phone" className="block text-xs font-semibold text-[#0b2439]">
                Telepon
              </label>
              <input
                id="loc-phone"
                type="text"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+6221..."
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="loc-tz" className="block text-xs font-semibold text-[#0b2439]">
                Zona Waktu (IANA) *
              </label>
              <input
                id="loc-tz"
                type="text"
                required
                value={timezone}
                onChange={(e) => setTimezone(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div>
            <label htmlFor="loc-attrs" className="block text-xs font-semibold text-[#0b2439]">
              Atribut Operasional JSON
            </label>
            <textarea
              id="loc-attrs"
              rows={2}
              value={attrsJson}
              onChange={(e) => setAttrsJson(e.target.value)}
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className={styles.btnSecondary}>
              Batal
            </button>
            <button type="submit" disabled={saving} className={styles.btnPrimary}>
              {saving ? "Menyimpan..." : editing ? "Simpan Perubahan" : "Buat Lokasi"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
