"use client";

import { useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { ORG_UNIT_KIND_OPTIONS, formatEfError } from "@/lib/graphql/ef-helpers";
import {
  createOrganizationUnit,
  updateOrganizationUnit,
} from "@/lib/graphql/ef-structure-mutations";
import {
  fetchAllLegalEntities,
  fetchAllLocations,
} from "@/lib/graphql/ef-structure-queries";
import type {
  EfLegalEntity,
  EfLocation,
  EfOrganizationUnit,
  OrgUnitKind,
} from "@/lib/graphql/ef-types";

export function OrgUnitFormModal({
  open,
  editing,
  onClose,
  onSaved,
}: {
  open: boolean;
  editing: EfOrganizationUnit | null;
  onClose: () => void;
  onSaved: (ou: EfOrganizationUnit) => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();

  const [legalEntities, setLegalEntities] = useState<EfLegalEntity[]>([]);
  const [locations, setLocations] = useState<EfLocation[]>([]);

  const [legalEntityId, setLegalEntityId] = useState(
    editing?.legal_entity_id || "",
  );
  const [parentUnitId, setParentUnitId] = useState(
    editing?.parent_unit_id || "",
  );
  const [primaryLocationId, setPrimaryLocationId] = useState(
    editing?.primary_location_id || "",
  );
  const [unitCode, setUnitCode] = useState(editing?.unit_code || "");
  const [name, setName] = useState(editing?.name || "");
  const [kind, setKind] = useState<OrgUnitKind>(editing?.kind || "DEPARTMENT");
  const [costCenterCode, setCostCenterCode] = useState(
    editing?.cost_center_code || "",
  );
  const [saving, setSaving] = useState(false);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    let active = true;
    async function loadRefs() {
      try {
        const [leConn, locConn] = await Promise.all([
          fetchAllLegalEntities(25),
          fetchAllLocations(25),
        ]);
        if (!active) return;
        setLegalEntities(leConn.nodes);
        setLocations(locConn.nodes);
        if (!editing && !legalEntityId && leConn.nodes[0]) {
          setLegalEntityId(leConn.nodes[0]._id);
        }
      } catch {
        // Fallback input UUID manual tetap tersedia.
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
    if (!name.trim() || (!editing && (!legalEntityId.trim() || !unitCode.trim()))) {
      setErrMsg("Badan hukum, kode unit, dan nama unit wajib diisi.");
      return;
    }

    setSaving(true);
    setErrMsg(null);
    try {
      if (editing) {
        const updated = await runDashboardAction(() =>
          updateOrganizationUnit(
            {
              id: editing.id,
              parent_unit_id: parentUnitId.trim() || null,
              primary_location_id: primaryLocationId.trim() || null,
              name: name.trim(),
              kind,
              cost_center_code: costCenterCode.trim() || null,
              preserve_head_person: true,
            },
            csrfToken,
          ),
        );
        if (updated) {
          pushToast("success", "Unit Organisasi Diperbarui", updated.name);
          onSaved(updated);
        }
      } else {
        const created = await runDashboardAction(() =>
          createOrganizationUnit(
            {
              legal_entity_id: legalEntityId.trim(),
              parent_unit_id: parentUnitId.trim() || null,
              primary_location_id: primaryLocationId.trim() || null,
              unit_code: unitCode.trim().toUpperCase(),
              name: name.trim(),
              kind,
              cost_center_code: costCenterCode.trim() || null,
              
            },
            csrfToken,
          ),
        );
        if (created) {
          pushToast("success", "Unit Organisasi Dibuat", created.name);
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
      aria-label={editing ? "Perbarui Unit Organisasi" : "Buat Unit Organisasi"}
    >
      <div className="max-h-[92dvh] w-full max-w-2xl overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-[#0b2439]">
              {editing ? "Perbarui Unit Organisasi" : "Tambah Unit Organisasi"}
            </h2>
            <p className="text-xs text-[#5a6672]">
              Unit aktif di bawah Badan Hukum aktif menjadi target scope `ORGANIZATION_UNIT`.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup modal unit organisasi"
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
              <label htmlFor="ou-le" className="block text-xs font-semibold text-[#0b2439]">
                Badan Hukum Pemilik *
              </label>
              {legalEntities.length > 0 && !editing ? (
                <select
                  id="ou-le"
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
                  id="ou-le"
                  type="text"
                  disabled={Boolean(editing)}
                  value={legalEntityId}
                  onChange={(e) => setLegalEntityId(e.target.value)}
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
              )}
            </div>

            <div>
              <label htmlFor="ou-code" className="block text-xs font-semibold text-[#0b2439]">
                Kode Unit *
              </label>
              <input
                id="ou-code"
                type="text"
                disabled={Boolean(editing)}
                required={!editing}
                value={unitCode}
                onChange={(e) => setUnitCode(e.target.value)}
                placeholder="DEPT-HR-01"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div className="sm:col-span-2">
              <label htmlFor="ou-name" className="block text-xs font-semibold text-[#0b2439]">
                Nama Unit *
              </label>
              <input
                id="ou-name"
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Departemen Sumber Daya Manusia"
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="ou-kind" className="block text-xs font-semibold text-[#0b2439]">
                Jenis *
              </label>
              <select
                id="ou-kind"
                value={kind}
                onChange={(e) => setKind(e.target.value as OrgUnitKind)}
                className={styles.fieldSelect}
              >
                {ORG_UNIT_KIND_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="ou-loc" className="block text-xs font-semibold text-[#0b2439]">
                Lokasi Utama (Opsional)
              </label>
              {locations.length > 0 ? (
                <select
                  id="ou-loc"
                  value={primaryLocationId}
                  onChange={(e) => setPrimaryLocationId(e.target.value)}
                  className={styles.fieldSelect}
                >
                  <option value="">-- Tanpa Lokasi Utama --</option>
                  {locations.map((loc) => (
                    <option key={loc.id} value={loc._id}>
                      {loc.location_code} — {loc.name}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  id="ou-loc"
                  type="text"
                  value={primaryLocationId}
                  onChange={(e) => setPrimaryLocationId(e.target.value)}
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
              )}
            </div>
            <div>
              <label htmlFor="ou-parent" className="block text-xs font-semibold text-[#0b2439]">
                Parent Unit ID (Opsional)
              </label>
              <input
                id="ou-parent"
                type="text"
                value={parentUnitId}
                onChange={(e) => setParentUnitId(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="ou-cc" className="block text-xs font-semibold text-[#0b2439]">
                Cost Center Code
              </label>
              <input
                id="ou-cc"
                type="text"
                value={costCenterCode}
                onChange={(e) => setCostCenterCode(e.target.value)}
                placeholder="CC-HR-100"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className={styles.btnSecondary}>
              Batal
            </button>
            <button type="submit" disabled={saving} className={styles.btnPrimary}>
              {saving ? "Menyimpan..." : editing ? "Simpan Perubahan" : "Buat Unit"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
