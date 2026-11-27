"use client";

import { useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { PARTY_KIND_OPTIONS, formatEfError } from "@/lib/graphql/ef-helpers";
import {
  createParty,
  updatePartyProfile,
  updatePartyTaxProfile,
} from "@/lib/graphql/ef-partner-mutations";
import type { EfParty, PartyKind } from "@/lib/graphql/ef-types";

export function PartyFormModal({
  open,
  editing,
  mode = "profile",
  onClose,
  onSaved,
}: {
  open: boolean;
  editing: EfParty | null;
  mode?: "create" | "profile" | "tax";
  onClose: () => void;
  onSaved: (party: EfParty) => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();

  const [partnerCode, setPartnerCode] = useState(editing?.partner_code || "");
  const [kind, setKind] = useState<PartyKind>(editing?.kind || "ORGANIZATION");
  const [displayName, setDisplayName] = useState(editing?.display_name || "");
  const [legalName, setLegalName] = useState(editing?.legal_name || "");
  const [email, setEmail] = useState(editing?.email || "");
  const [phone, setPhone] = useState(editing?.phone || "");
  const [websiteUrl, setWebsiteUrl] = useState(editing?.website_url || "");
  const [preferredLanguage, setPreferredLanguage] = useState(
    editing?.preferred_language || "id",
  );
  const [notes, setNotes] = useState(editing?.notes || "");

  const [taxIdNumber, setTaxIdNumber] = useState(editing?.tax_id_number || "");
  const [taxRegisteredName, setTaxRegisteredName] = useState(
    editing?.tax_registered_name || "",
  );
  const [isPkp, setIsPkp] = useState(Boolean(editing?.is_pkp));
  const [isTaxExempt, setIsTaxExempt] = useState(Boolean(editing?.is_tax_exempt));
  const [taxExemptionReason, setTaxExemptionReason] = useState(
    editing?.tax_exemption_reason || "",
  );
  const [saving, setSaving] = useState(false);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  if (!open) return null;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (saving) return;
    if (mode !== "tax" && !displayName.trim()) {
      setErrMsg("Nama tampilan mitra wajib diisi.");
      return;
    }
    if (!editing && !partnerCode.trim()) {
      setErrMsg("Kode mitra unik wajib diisi.");
      return;
    }

    setSaving(true);
    setErrMsg(null);
    try {
      if (!editing) {
        const created = await runDashboardAction(() =>
          createParty(
            {
              partner_code: partnerCode.trim().toUpperCase(),
              kind,
              display_name: displayName.trim(),
              legal_name: legalName.trim() || null,
              email: email.trim() || null,
              phone: phone.trim() || null,
              website_url: websiteUrl.trim() || null,
              tax_id_number: taxIdNumber.trim() || null,
              tax_registered_name: taxRegisteredName.trim() || null,
              is_pkp: isPkp,
              is_tax_exempt: isTaxExempt,
              tax_exemption_reason: taxExemptionReason.trim() || null,
              preferred_language: preferredLanguage.trim() || "id",
              notes: notes.trim() || null,
            },
            csrfToken,
          ),
        );
        if (created) {
          pushToast("success", "Mitra Dibuat", created.display_name);
          onSaved(created);
        }
      } else if (mode === "tax") {
        const updated = await runDashboardAction(() =>
          updatePartyTaxProfile(
            {
              id: editing.id,
              tax_id_number: taxIdNumber.trim() || null,
              tax_registered_name: taxRegisteredName.trim() || null,
              is_pkp: isPkp,
              is_tax_exempt: isTaxExempt,
              tax_exemption_reason: taxExemptionReason.trim() || null,
              tax_verified_at: new Date().toISOString(),
            },
            csrfToken,
          ),
        );
        if (updated) {
          pushToast("success", "Profil Pajak Mitra Diperbarui", updated.display_name);
          onSaved(updated);
        }
      } else {
        const updated = await runDashboardAction(() =>
          updatePartyProfile(
            {
              id: editing.id,
              kind,
              display_name: displayName.trim(),
              legal_name: legalName.trim() || null,
              email: email.trim() || null,
              phone: phone.trim() || null,
              website_url: websiteUrl.trim() || null,
              preferred_language: preferredLanguage.trim() || "id",
              notes: notes.trim() || null,
            },
            csrfToken,
          ),
        );
        if (updated) {
          pushToast("success", "Profil Mitra Diperbarui", updated.display_name);
          onSaved(updated);
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
      aria-label={editing ? "Perbarui Mitra" : "Daftarkan Mitra Baru"}
    >
      <div className="max-h-[92dvh] w-full max-w-2xl overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-[#0b2439]">
              {!editing
                ? "Daftarkan Mitra Eksternal"
                : mode === "tax"
                  ? "Perbarui Profil Pajak Mitra"
                  : "Perbarui Profil Mitra"}
            </h2>
            <p className="text-xs text-[#5a6672]">
              Master mitra Supplier, Vendor CMT, Ekspedisi, &amp; B2B.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup modal mitra"
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
          {mode !== "tax" ? (
            <>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <div>
                  <label htmlFor="prt-code" className="block text-xs font-semibold text-[#0b2439]">
                    Kode Mitra *
                  </label>
                  <input
                    id="prt-code"
                    type="text"
                    required={!editing}
                    disabled={Boolean(editing)}
                    value={partnerCode}
                    onChange={(e) => setPartnerCode(e.target.value)}
                    placeholder="VND-CMT-01"
                    className={`${styles.fieldInput} ${styles.monoText}`}
                  />
                </div>
                <div>
                  <label htmlFor="prt-kind" className="block text-xs font-semibold text-[#0b2439]">
                    Bentuk Entitas *
                  </label>
                  <select
                    id="prt-kind"
                    value={kind}
                    onChange={(e) => setKind(e.target.value as PartyKind)}
                    className={styles.fieldSelect}
                  >
                    {PARTY_KIND_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label htmlFor="prt-lang" className="block text-xs font-semibold text-[#0b2439]">
                    Bahasa
                  </label>
                  <input
                    id="prt-lang"
                    type="text"
                    value={preferredLanguage}
                    onChange={(e) => setPreferredLanguage(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label htmlFor="prt-display" className="block text-xs font-semibold text-[#0b2439]">
                    Nama Tampilan *
                  </label>
                  <input
                    id="prt-display"
                    type="text"
                    required
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    placeholder="CV Jahit Karya Mandiri"
                    className={styles.fieldInput}
                  />
                </div>
                <div>
                  <label htmlFor="prt-legal" className="block text-xs font-semibold text-[#0b2439]">
                    Nama Legal Lengkap
                  </label>
                  <input
                    id="prt-legal"
                    type="text"
                    value={legalName}
                    onChange={(e) => setLegalName(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <div>
                  <label htmlFor="prt-email" className="block text-xs font-semibold text-[#0b2439]">
                    Email Bisnis
                  </label>
                  <input
                    id="prt-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
                <div>
                  <label htmlFor="prt-phone" className="block text-xs font-semibold text-[#0b2439]">
                    Telepon
                  </label>
                  <input
                    id="prt-phone"
                    type="text"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
                <div>
                  <label htmlFor="prt-web" className="block text-xs font-semibold text-[#0b2439]">
                    Situs Web
                  </label>
                  <input
                    id="prt-web"
                    type="url"
                    value={websiteUrl}
                    onChange={(e) => setWebsiteUrl(e.target.value)}
                    placeholder="https://..."
                    className={styles.fieldInput}
                  />
                </div>
              </div>

              <div>
                <label htmlFor="prt-notes" className="block text-xs font-semibold text-[#0b2439]">
                  Catatan
                </label>
                <textarea
                  id="prt-notes"
                  rows={2}
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className={styles.fieldInput}
                />
              </div>
            </>
          ) : null}

          {!editing || mode === "tax" ? (
            <div className="space-y-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-3.5">
              <p className="text-xs font-semibold text-[#0b2439]">Profil Pajak</p>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label htmlFor="prt-npwp" className="block text-xs font-semibold text-[#0b2439]">
                    NPWP / Tax ID
                  </label>
                  <input
                    id="prt-npwp"
                    type="text"
                    value={taxIdNumber}
                    onChange={(e) => setTaxIdNumber(e.target.value)}
                    className={`${styles.fieldInput} ${styles.monoText}`}
                  />
                </div>
                <div>
                  <label htmlFor="prt-taxname" className="block text-xs font-semibold text-[#0b2439]">
                    Nama Terdaftar Pajak
                  </label>
                  <input
                    id="prt-taxname"
                    type="text"
                    value={taxRegisteredName}
                    onChange={(e) => setTaxRegisteredName(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
              </div>
              <div className="flex flex-wrap gap-4">
                <label className="flex items-center gap-2 text-xs text-[#0b2439]">
                  <input
                    type="checkbox"
                    checked={isPkp}
                    onChange={(e) => setIsPkp(e.target.checked)}
                  />
                  <span>PKP</span>
                </label>
                <label className="flex items-center gap-2 text-xs text-[#0b2439]">
                  <input
                    type="checkbox"
                    checked={isTaxExempt}
                    onChange={(e) => setIsTaxExempt(e.target.checked)}
                  />
                  <span>Bebas Pajak (Tax Exempt)</span>
                </label>
              </div>
              {isTaxExempt ? (
                <div>
                  <label htmlFor="prt-exempt" className="block text-xs font-semibold text-[#0b2439]">
                    Alasan Pengecualian Pajak
                  </label>
                  <input
                    id="prt-exempt"
                    type="text"
                    value={taxExemptionReason}
                    onChange={(e) => setTaxExemptionReason(e.target.value)}
                    className={styles.fieldInput}
                  />
                </div>
              ) : null}
            </div>
          ) : null}

          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={onClose} className={styles.btnSecondary}>
              Batal
            </button>
            <button type="submit" disabled={saving} className={styles.btnPrimary}>
              {saving ? "Menyimpan..." : editing ? "Simpan" : "Daftarkan Mitra"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
