"use client";

import { useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  createLegalEntity,
  updateLegalEntity,
} from "@/lib/graphql/ef-structure-mutations";
import { formatEfError } from "@/lib/graphql/ef-helpers";
import type { EfLegalEntity } from "@/lib/graphql/ef-types";

export function LegalEntityFormModal({
  open,
  editing,
  onClose,
  onSaved,
}: {
  open: boolean;
  editing: EfLegalEntity | null;
  onClose: () => void;
  onSaved: (entity: EfLegalEntity) => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();

  const [entityCode, setEntityCode] = useState(editing?.entity_code || "");
  const [legalName, setLegalName] = useState(editing?.legal_name || "");
  const [tradingName, setTradingName] = useState(editing?.trading_name || "");
  const [legalForm, setLegalForm] = useState(editing?.legal_form || "PT");
  const [countryCode, setCountryCode] = useState(
    editing?.jurisdiction_country_code || "ID",
  );
  const [parentId, setParentId] = useState(
    editing?.parent_legal_entity_id || "",
  );
  const [regAddressId, setRegAddressId] = useState(
    editing?.registered_address_id || "",
  );
  const [nib, setNib] = useState(editing?.nib || "");
  const [taxId, setTaxId] = useState(editing?.tax_id_number || "");
  const [isPkp, setIsPkp] = useState(Boolean(editing?.is_pkp));
  const [isPrimary, setIsPrimary] = useState(
    Boolean(editing?.is_primary_operating_entity),
  );
  const [fiscalMonth, setFiscalMonth] = useState(
    editing?.fiscal_year_start_month || 1,
  );
  const [saving, setSaving] = useState(false);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  if (!open) return null;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (saving) return;
    if (!legalName.trim() || !countryCode.trim() || (!editing && !entityCode.trim())) {
      setErrMsg("Kode entitas, nama legal, dan kode negara wajib diisi.");
      return;
    }

    setSaving(true);
    setErrMsg(null);
    try {
      if (editing) {
        const updated = await runDashboardAction(() =>
          updateLegalEntity(
            {
              id: editing.id,
              parent_legal_entity_id: parentId.trim() || null,
              legal_name: legalName.trim(),
              trading_name: tradingName.trim() || null,
              legal_form: legalForm.trim() || null,
              jurisdiction_country_code: countryCode.trim().toUpperCase(),
              nib: nib.trim() || null,
              tax_id_number: taxId.trim() || null,
              is_pkp: isPkp,
              registered_address_id: regAddressId.trim() || null,
              fiscal_year_start_month: Number(fiscalMonth) || 1,
            },
            csrfToken,
          ),
        );
        if (updated) {
          pushToast("success", "Badan Hukum Diperbarui", updated.legal_name);
          onSaved(updated);
        }
      } else {
        const created = await runDashboardAction(() =>
          createLegalEntity(
            {
              parent_legal_entity_id: parentId.trim() || null,
              entity_code: entityCode.trim().toUpperCase(),
              legal_name: legalName.trim(),
              trading_name: tradingName.trim() || null,
              legal_form: legalForm.trim() || null,
              is_primary_operating_entity: isPrimary,
              jurisdiction_country_code: countryCode.trim().toUpperCase(),
              nib: nib.trim() || null,
              tax_id_number: taxId.trim() || null,
              is_pkp: isPkp,
              registered_address_id: regAddressId.trim() || null,
              fiscal_year_start_month: Number(fiscalMonth) || 1,
            },
            csrfToken,
          ),
        );
        if (created) {
          pushToast("success", "Badan Hukum Dibuat", created.legal_name);
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
      aria-label={editing ? "Perbarui Badan Hukum" : "Buat Badan Hukum"}
    >
      <div className="max-h-[92dvh] w-full max-w-2xl overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-[#0b2439]">
              {editing ? "Perbarui Badan Hukum" : "Daftarkan Badan Hukum Baru"}
            </h2>
            <p className="text-xs text-[#5a6672]">
              Entitas badan hukum aktif menjadi target scope `LEGAL_ENTITY` di IAM.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup modal badan hukum"
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
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div>
              <label htmlFor="le-code" className="block text-xs font-semibold text-[#0b2439]">
                Kode Entitas *
              </label>
              <input
                id="le-code"
                type="text"
                required={!editing}
                disabled={Boolean(editing)}
                value={entityCode}
                onChange={(e) => setEntityCode(e.target.value)}
                placeholder="PT-IG-01"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
            <div className="sm:col-span-2">
              <label htmlFor="le-name" className="block text-xs font-semibold text-[#0b2439]">
                Nama Legal Resmi *
              </label>
              <input
                id="le-name"
                type="text"
                required
                value={legalName}
                onChange={(e) => setLegalName(e.target.value)}
                placeholder="PT IntegriOne Retail Indonesia"
                className={styles.fieldInput}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div>
              <label htmlFor="le-form" className="block text-xs font-semibold text-[#0b2439]">
                Bentuk Usaha
              </label>
              <input
                id="le-form"
                type="text"
                value={legalForm}
                onChange={(e) => setLegalForm(e.target.value)}
                placeholder="PT / CV / Yayasan"
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="le-trading" className="block text-xs font-semibold text-[#0b2439]">
                Nama Dagang
              </label>
              <input
                id="le-trading"
                type="text"
                value={tradingName}
                onChange={(e) => setTradingName(e.target.value)}
                placeholder="IntegriOne"
                className={styles.fieldInput}
              />
            </div>
            <div>
              <label htmlFor="le-country" className="block text-xs font-semibold text-[#0b2439]">
                Yurisdiksi (ISO 2) *
              </label>
              <input
                id="le-country"
                type="text"
                maxLength={2}
                required
                value={countryCode}
                onChange={(e) => setCountryCode(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div>
              <label htmlFor="le-nib" className="block text-xs font-semibold text-[#0b2439]">
                NIB
              </label>
              <input
                id="le-nib"
                type="text"
                value={nib}
                onChange={(e) => setNib(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
            <div>
              <label htmlFor="le-npwp" className="block text-xs font-semibold text-[#0b2439]">
                NPWP / Tax ID
              </label>
              <input
                id="le-npwp"
                type="text"
                value={taxId}
                onChange={(e) => setTaxId(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
            <div>
              <label htmlFor="le-fiscal" className="block text-xs font-semibold text-[#0b2439]">
                Awal Tahun Fiskal (1-12)
              </label>
              <input
                id="le-fiscal"
                type="number"
                min={1}
                max={12}
                value={fiscalMonth}
                onChange={(e) => setFiscalMonth(Number(e.target.value))}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="le-parent" className="block text-xs font-semibold text-[#0b2439]">
                Parent Legal Entity (UUID/GID Opsional)
              </label>
              <input
                id="le-parent"
                type="text"
                value={parentId}
                onChange={(e) => setParentId(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
            <div>
              <label htmlFor="le-addr" className="block text-xs font-semibold text-[#0b2439]">
                Registered Address ID (UUID/GID Opsional)
              </label>
              <input
                id="le-addr"
                type="text"
                value={regAddressId}
                onChange={(e) => setRegAddressId(e.target.value)}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
            </div>
          </div>

          <div className="flex flex-wrap gap-4 pt-1">
            <label className="flex items-center gap-2 text-xs text-[#0b2439]">
              <input
                type="checkbox"
                checked={isPkp}
                onChange={(e) => setIsPkp(e.target.checked)}
              />
              <span>Pengusaha Kena Pajak (PKP)</span>
            </label>
            {!editing ? (
              <label className="flex items-center gap-2 text-xs text-[#0b2439]">
                <input
                  type="checkbox"
                  checked={isPrimary}
                  onChange={(e) => setIsPrimary(e.target.checked)}
                />
                <span>Tetapkan sebagai Entitas Operasional Utama</span>
              </label>
            ) : null}
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className={styles.btnSecondary}>
              Batal
            </button>
            <button type="submit" disabled={saving} className={styles.btnPrimary}>
              {saving ? "Menyimpan..." : editing ? "Simpan Perubahan" : "Buat Badan Hukum"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
