"use client";

import { useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  createAddress,
  findOrCreateAddress,
  updateAddress,
} from "@/lib/graphql/ef-structure-mutations";
import { formatEfError } from "@/lib/graphql/ef-helpers";
import type { EfAddress } from "@/lib/graphql/ef-types";

export function AddressFormModal({
  open,
  editing,
  onClose,
  onSaved,
}: {
  open: boolean;
  editing: EfAddress | null;
  onClose: () => void;
  onSaved: (address: EfAddress) => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();

  const [label, setLabel] = useState(editing?.label || "");
  const [line1, setLine1] = useState(editing?.address_line_1 || "");
  const [line2, setLine2] = useState(editing?.address_line_2 || "");
  const [city, setCity] = useState(editing?.city || "");
  const [province, setProvince] = useState(editing?.province || "");
  const [postalCode, setPostalCode] = useState(editing?.postal_code || "");
  const [countryCode, setCountryCode] = useState(editing?.country_code || "ID");
  const [formatted, setFormatted] = useState(editing?.formatted_address || "");
  const [latitude, setLatitude] = useState(
    editing?.latitude != null ? String(editing.latitude) : "",
  );
  const [longitude, setLongitude] = useState(
    editing?.longitude != null ? String(editing.longitude) : "",
  );
  const [deduplicateMode, setDeduplicateMode] = useState(false);
  const [saving, setSaving] = useState(false);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  if (!open) return null;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (saving) return;
    if (!line1.trim() || !city.trim() || !countryCode.trim()) {
      setErrMsg("Alamat baris 1, kota, dan kode negara ISO 2-huruf wajib diisi.");
      return;
    }

    const latVal = latitude.trim() ? Number(latitude.trim()) : null;
    const lngVal = longitude.trim() ? Number(longitude.trim()) : null;
    if (
      (latVal !== null && Number.isNaN(latVal)) ||
      (lngVal !== null && Number.isNaN(lngVal))
    ) {
      setErrMsg("Latitude/Longitude harus berupa angka valid.");
      return;
    }

    setSaving(true);
    setErrMsg(null);
    const input = {
      label: label.trim() || null,
      address_line_1: line1.trim(),
      address_line_2: line2.trim() || null,
      city: city.trim(),
      province: province.trim() || null,
      postal_code: postalCode.trim() || null,
      country_code: countryCode.trim().toUpperCase(),
      formatted_address: formatted.trim() || null,
      latitude: latVal,
      longitude: lngVal,
    };

    try {
      if (editing) {
        const updated = await runDashboardAction(() =>
          updateAddress({ id: editing.id, ...input }, csrfToken),
        );
        if (updated) {
          pushToast("success", "Alamat Diperbarui", updated.city);
          onSaved(updated);
        }
      } else if (deduplicateMode) {
        const res = await runDashboardAction(() =>
          findOrCreateAddress(input, csrfToken),
        );
        if (res?.address) {
          pushToast(
            "success",
            res.created ? "Alamat Baru Dibuat" : "Alamat Identik Ditemukan",
            `${res.address.address_line_1} (${res.address.city})`,
          );
          onSaved(res.address);
        }
      } else {
        const created = await runDashboardAction(() =>
          createAddress(input, csrfToken),
        );
        if (created) {
          pushToast("success", "Alamat Baru Dibuat", created.city);
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
      aria-label={editing ? "Perbarui Alamat" : "Tambah Master Alamat"}
    >
      <div className="max-h-[92dvh] w-full max-w-xl overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-[#0b2439]">
              {editing ? "Perbarui Master Alamat" : "Tambah Master Alamat"}
            </h2>
            <p className="text-xs text-[#5a6672]">
              Dipakai bersama oleh Badan Hukum, Lokasi Fisik, dan Mitra Party.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup modal alamat"
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
              <label htmlFor="addr-label" className="block text-xs font-semibold text-[#0b2439]">
                Label
              </label>
              <input
                id="addr-label"
                type="text"
                value={label}
                onChange={(e) => setLabel(e.target.value)}
                placeholder="Kantor Pusat / Gudang Cikarang"
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="addr-country" className="block text-xs font-semibold text-[#0b2439]">
                Kode Negara (ISO 2) *
              </label>
              <input
                id="addr-country"
                type="text"
                maxLength={2}
                required
                value={countryCode}
                onChange={(e) => setCountryCode(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div>
            <label htmlFor="addr-line1" className="block text-xs font-semibold text-[#0b2439]">
              Alamat Baris 1 *
            </label>
            <input
              id="addr-line1"
              type="text"
              required
              value={line1}
              onChange={(e) => setLine1(e.target.value)}
              placeholder="Jl. Jenderal Sudirman Kav. 52-53"
              className={styles.fieldInput}
            />
          </div>

          <div>
            <label htmlFor="addr-line2" className="block text-xs font-semibold text-[#0b2439]">
              Alamat Baris 2
            </label>
            <input
              id="addr-line2"
              type="text"
              value={line2}
              onChange={(e) => setLine2(e.target.value)}
              placeholder="Gedung / Blok / Unit"
              className={styles.fieldInput}
            />
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div>
              <label htmlFor="addr-city" className="block text-xs font-semibold text-[#0b2439]">
                Kota / Kabupaten *
              </label>
              <input
                id="addr-city"
                type="text"
                required
                value={city}
                onChange={(e) => setCity(e.target.value)}
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="addr-prov" className="block text-xs font-semibold text-[#0b2439]">
                Provinsi
              </label>
              <input
                id="addr-prov"
                type="text"
                value={province}
                onChange={(e) => setProvince(e.target.value)}
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="addr-postal" className="block text-xs font-semibold text-[#0b2439]">
                Kode Pos
              </label>
              <input
                id="addr-postal"
                type="text"
                value={postalCode}
                onChange={(e) => setPostalCode(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div>
            <label htmlFor="addr-formatted" className="block text-xs font-semibold text-[#0b2439]">
              Alamat Lengkap Terformat
            </label>
            <input
              id="addr-formatted"
              type="text"
              value={formatted}
              onChange={(e) => setFormatted(e.target.value)}
              className={styles.fieldInput}
            />
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="addr-lat" className="block text-xs font-semibold text-[#0b2439]">
                Latitude
              </label>
              <input
                id="addr-lat"
                type="text"
                value={latitude}
                onChange={(e) => setLatitude(e.target.value)}
                placeholder="-6.2088"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
            <div>
              <label htmlFor="addr-lng" className="block text-xs font-semibold text-[#0b2439]">
                Longitude
              </label>
              <input
                id="addr-lng"
                type="text"
                value={longitude}
                onChange={(e) => setLongitude(e.target.value)}
                placeholder="106.8456"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          {!editing ? (
            <label className="flex items-center gap-2 rounded-lg border border-[#dde3e8] bg-[#f9fafb] p-3 text-xs text-[#0b2439]">
              <input
                type="checkbox"
                checked={deduplicateMode}
                onChange={(e) => setDeduplicateMode(e.target.checked)}
              />
              <span>
                Gunakan <code className={styles.monoText}>findOrCreateAddress</code>{" "}
                (deduplikasi otomatis bila alamat identik sudah ada)
              </span>
            </label>
          ) : null}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className={styles.btnSecondary}>
              Batal
            </button>
            <button type="submit" disabled={saving} className={styles.btnPrimary}>
              {saving ? "Menyimpan..." : editing ? "Simpan Perubahan" : "Simpan Alamat"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
