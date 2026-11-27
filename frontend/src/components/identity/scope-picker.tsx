"use client";

import { useState } from "react";
import styles from "./identity.module.css";
import { IdentityIcon } from "./identity-icons";
import {
  lookupEnterpriseScopeTargets,
  type ScopeTargetOption,
} from "@/lib/graphql/ef-scope-lookup";
import { getScopeVerificationHint } from "@/lib/graphql/iam-role-access-helpers";
import type { ScopeType } from "@/lib/graphql/iam-types";

const BASE_SCOPE_OPTIONS: Array<{ value: ScopeType; label: string; hint: string }> = [
  { value: "GLOBAL", label: "GLOBAL", hint: "Berlaku untuk seluruh perusahaan" },
  { value: "SELF", label: "SELF", hint: "Khusus kepemilikan akun sendiri" },
  { value: "LEGAL_ENTITY", label: "LEGAL_ENTITY", hint: "UUID entitas hukum tertentu" },
  {
    value: "ORGANIZATION_UNIT",
    label: "ORGANIZATION_UNIT",
    hint: "UUID divisi / departemen organisasi",
  },
  { value: "FACILITY", label: "FACILITY", hint: "UUID lokasi / gudang / toko" },
  { value: "SUPPLIER_PARTY", label: "SUPPLIER_PARTY", hint: "UUID party pemasok bahan" },
  { value: "VENDOR_PARTY", label: "VENDOR_PARTY", hint: "UUID party vendor CMT / jasa" },
];

export function ScopePicker({
  scopeType,
  scopeId,
  selfUserId,
  allowSelf = true,
  onScopeTypeChange,
  onScopeIdChange,
  idPrefix = "scope",
}: {
  scopeType: ScopeType;
  scopeId: string;
  selfUserId?: string;
  allowSelf?: boolean;
  onScopeTypeChange: (next: ScopeType) => void;
  onScopeIdChange: (next: string) => void;
  idPrefix?: string;
}) {
  const [searchTerm, setSearchTerm] = useState("");
  const [lookupResults, setLookupResults] = useState<ScopeTargetOption[]>([]);
  const [lookupNotice, setLookupNotice] = useState<string | null>(null);
  const [isSearching, setIsSearching] = useState(false);

  const visibleScopeTypes = allowSelf
    ? BASE_SCOPE_OPTIONS
    : BASE_SCOPE_OPTIONS.filter((opt) => opt.value !== "SELF");

  const supportsEfLookup =
    scopeType === "LEGAL_ENTITY" ||
    scopeType === "ORGANIZATION_UNIT" ||
    scopeType === "FACILITY" ||
    scopeType === "SUPPLIER_PARTY" ||
    scopeType === "VENDOR_PARTY";

  const isAutoBoundSelf = scopeType === "SELF" && Boolean(selfUserId);
  const verificationHint = getScopeVerificationHint(scopeType);

  async function handleLookup() {
    if (!supportsEfLookup || isSearching) return;
    setIsSearching(true);
    setLookupNotice(null);
    const res = await lookupEnterpriseScopeTargets(scopeType, searchTerm);
    setLookupResults(res.options);
    if (res.unavailableReason) {
      setLookupNotice(res.unavailableReason);
    } else if (res.options.length === 0) {
      setLookupNotice("Tidak ada target aktif yang cocok di Enterprise Foundation.");
    }
    setIsSearching(false);
  }

  return (
    <div className="space-y-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-3.5">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor={`${idPrefix}-type`}
            className="text-xs font-semibold text-[#0b2439]"
          >
            Jenis Scope
          </label>
          <select
            id={`${idPrefix}-type`}
            value={scopeType}
            onChange={(e) => {
              const next = e.target.value as ScopeType;
              onScopeTypeChange(next);
              setLookupResults([]);
              setLookupNotice(null);
              if (next === "GLOBAL") {
                onScopeIdChange("");
              } else if (next === "SELF" && selfUserId) {
                onScopeIdChange(selfUserId);
              }
            }}
            className={styles.fieldSelect}
          >
            {visibleScopeTypes.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <span className="text-[11px] text-[#5a6672]">
            {visibleScopeTypes.find((item) => item.value === scopeType)?.hint}
          </span>
        </div>

        <div className="flex flex-col gap-1.5">
          <label
            htmlFor={`${idPrefix}-id`}
            className="text-xs font-semibold text-[#0b2439]"
          >
            Scope ID (UUID / GID)
          </label>
          <input
            id={`${idPrefix}-id`}
            type="text"
            value={isAutoBoundSelf ? selfUserId || scopeId : scopeId}
            disabled={scopeType === "GLOBAL" || isAutoBoundSelf}
            onChange={(e) => onScopeIdChange(e.target.value)}
            placeholder={
              scopeType === "GLOBAL"
                ? "NULL otomatis untuk GLOBAL"
                : "Contoh: 00000000-0000-4000-8000-..."
            }
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
          <span className="text-[11px] text-[#5a6672]">
            {scopeType === "GLOBAL"
              ? "Scope GLOBAL tidak menerima target ID."
              : isAutoBoundSelf
                ? "Otomatis mengikat UUID akun target ini."
                : "Wajib UUID valid pemilik domain."}
          </span>
        </div>
      </div>

      {/* Catatan Kesiapan Verifikasi Scope Server IAM */}
      <div
        className={`flex items-start gap-2 rounded-lg border px-2.5 py-2 text-[11px] ${
          verificationHint.supportedNow
            ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
            : "border-[#fedf89] bg-[#fffaeb] text-[#b54708]"
        }`}
      >
        <span className={`shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold ${styles.monoText}`}>
          {verificationHint.badgeLabel}
        </span>
        <span className="leading-relaxed">{verificationHint.message}</span>
      </div>

      {supportsEfLookup ? (
        <div className="space-y-2 border-t border-[#dde3e8] pt-2.5">
          <div className="flex flex-col gap-2 sm:flex-row">
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Cari kode / nama di Enterprise Foundation (opsional)..."
              aria-label="Cari target Enterprise Foundation"
              className={styles.fieldInput}
            />
            <button
              type="button"
              onClick={() => void handleLookup()}
              disabled={isSearching}
              className={`${styles.btnSecondary} shrink-0`}
            >
              <IdentityIcon name="search" size={15} />
              <span>{isSearching ? "Mencari..." : "Cari Target"}</span>
            </button>
          </div>

          {lookupNotice ? (
            <p className="text-xs text-[#5a6672]">{lookupNotice}</p>
          ) : null}

          {lookupResults.length > 0 ? (
            <div className="max-h-40 divide-y divide-[#dde3e8] overflow-y-auto rounded-lg border border-[#dde3e8] bg-white">
              {lookupResults.map((item) => {
                const chosen = item.rawId === scopeId;
                return (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => onScopeIdChange(item.rawId)}
                    className={`flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-xs ${
                      chosen ? "bg-[#e9eef2]" : "hover:bg-[#f4f7f8]"
                    }`}
                  >
                    <div className="min-w-0">
                      <p className="truncate font-semibold text-[#0b2439]">
                        {item.label}
                      </p>
                      <p className={`truncate text-[#5a6672] ${styles.monoText}`}>
                        {item.code} · {item.rawId}
                      </p>
                    </div>
                    <span
                      className={`shrink-0 rounded border px-1.5 py-0.5 font-medium ${
                        chosen
                          ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                          : "border-[#dde3e8] bg-[#f4f7f8] text-[#133554]"
                      }`}
                    >
                      {chosen ? "Dipilih" : "Pilih"}
                    </span>
                  </button>
                );
              })}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
