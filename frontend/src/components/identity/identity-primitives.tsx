"use client";

import { useEffect, type ReactNode } from "react";
import styles from "./identity.module.css";
import { IdentityIcon } from "./identity-icons";
import type { PageInfo, Scope, UserKind, UserStatus } from "@/lib/graphql/iam-types";

export function formatDateTimeID(iso?: string | null): string {
  if (!iso) return "Belum ada";
  const dt = new Date(iso);
  if (Number.isNaN(dt.getTime())) return iso;
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(dt);
}

export function formatUnixTimeID(unixSec?: number | null): string {
  if (!unixSec || unixSec <= 0) return "N/A";
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(unixSec * 1000));
}

export function compactId(raw?: string | null): string {
  if (!raw) return "-";
  if (raw.length <= 14) return raw;
  return `${raw.slice(0, 8)}...${raw.slice(-4)}`;
}

export function UserStatusBadge({ status }: { status: UserStatus | string }) {
  const normalized = status.toUpperCase();
  const palette: Record<string, string> = {
    ACTIVE: "border-[#abefc6] bg-[#ecfdf3] text-[#067647]",
    PENDING: "border-[#fedf89] bg-[#fffaeb] text-[#b54708]",
    SUSPENDED: "border-[#fecdca] bg-[#fef3f2] text-[#b42318]",
    DISABLED: "border-[#dde3e8] bg-[#f4f7f8] text-[#5a6672]",
  };
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-[11px] font-semibold tracking-wide ${styles.monoText} ${palette[normalized] ?? palette.DISABLED}`}
    >
      <span
        aria-hidden="true"
        className={`h-1.5 w-1.5 rounded-full ${
          normalized === "ACTIVE"
            ? "bg-[#067647]"
            : normalized === "PENDING"
              ? "bg-[#d97706]"
              : normalized === "SUSPENDED"
                ? "bg-[#da414a]"
                : "bg-[#788491]"
        }`}
      />
      {normalized}
    </span>
  );
}

export function UserKindBadge({ kind }: { kind: UserKind | string }) {
  const labels: Record<string, string> = {
    EMPLOYEE: "EMPLOYEE",
    VENDOR: "VENDOR",
    SUPPLIER: "SUPPLIER",
    CUSTOMER: "CUSTOMER",
    SYSTEM: "SYSTEM",
  };
  return (
    <span
      className={`inline-flex items-center rounded-md border border-[#dde3e8] bg-[#f4f7f8] px-2 py-0.5 text-[11px] font-medium text-[#133554] ${styles.monoText}`}
    >
      {labels[kind] ?? kind}
    </span>
  );
}

export function ActiveBooleanBadge({
  active,
  trueText = "AKTIF",
  falseText = "NONAKTIF",
}: {
  active: boolean;
  trueText?: string;
  falseText?: string;
}) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-[11px] font-semibold ${styles.monoText} ${
        active
          ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
          : "border-[#dde3e8] bg-[#f4f7f8] text-[#5a6672]"
      }`}
    >
      {active ? trueText : falseText}
    </span>
  );
}

export function SensitiveBadge({ sensitive }: { sensitive: boolean }) {
  if (!sensitive) {
    return (
      <span
        className={`inline-flex items-center rounded-md border border-[#dde3e8] bg-[#f9fafb] px-2 py-0.5 text-[11px] font-medium text-[#5a6672] ${styles.monoText}`}
      >
        STANDAR
      </span>
    );
  }
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-md border border-[#fecdca] bg-[#fef3f2] px-2 py-0.5 text-[11px] font-semibold text-[#b42318] ${styles.monoText}`}
    >
      <IdentityIcon name="lock" size={11} />
      SENSITIF
    </span>
  );
}

export function ScopeBadge({ scope }: { scope: Scope }) {
  return (
    <span
      className={`inline-flex flex-wrap items-center gap-1.5 rounded-md border border-[#dde3e8] bg-[#f4f7f8] px-2 py-0.5 text-[11px] text-[#0b2439] ${styles.monoText}`}
    >
      <span className="font-semibold">{scope.type}</span>
      {scope.id ? (
        <span className="text-[#5a6672]">{compactId(scope.id)}</span>
      ) : null}
    </span>
  );
}

export function TableSkeleton({
  rows = 6,
  cols = 5,
}: {
  rows?: number;
  cols?: number;
}) {
  return (
    <div className={styles.tableWrap} aria-busy="true" aria-label="Memuat data">
      <table className={styles.dataTable}>
        <thead>
          <tr>
            {Array.from({ length: cols }).map((_, idx) => (
              <th key={idx}>
                <div className={`${styles.skeletonShimmer} h-3 w-20`} />
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rows }).map((_, rIdx) => (
            <tr key={rIdx}>
              {Array.from({ length: cols }).map((__, cIdx) => (
                <td key={cIdx}>
                  <div
                    className={`${styles.skeletonShimmer} h-4 ${
                      cIdx === 0 ? "w-36" : cIdx === cols - 1 ? "w-16" : "w-24"
                    }`}
                  />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function EmptyPanel({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 px-6 py-12 text-center">
      <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-[#dde3e8] bg-[#f4f7f8] text-[#133554]">
        <IdentityIcon name="search" size={20} />
      </div>
      <div className="max-w-md space-y-1">
        <p className="text-sm font-semibold text-[#0b2439]">{title}</p>
        <p className="text-sm text-[#5a6672]">{description}</p>
      </div>
      {action ? <div className="pt-1">{action}</div> : null}
    </div>
  );
}

export function ErrorPanel({
  title = "Data Belum Dapat Dimuat",
  message,
  onRetry,
}: {
  title?: string;
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div
      role="alert"
      className="flex flex-col items-start justify-between gap-4 rounded-xl border border-[#fecdca] bg-[#fef3f2] p-5 sm:flex-row sm:items-center"
    >
      <div className="flex items-start gap-3">
        <span className="mt-0.5 text-[#da414a]">
          <IdentityIcon name="alert" size={19} />
        </span>
        <div>
          <p className="text-sm font-semibold text-[#b42318]">{title}</p>
          <p className="mt-0.5 text-sm text-[#7a271a]">{message}</p>
        </div>
      </div>
      {onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className={`${styles.btnSecondary} ${styles.btnSmall} shrink-0`}
        >
          <IdentityIcon name="refresh" size={14} />
          <span>Muat Ulang</span>
        </button>
      ) : null}
    </div>
  );
}

export function RelayPaginationBar({
  totalCount,
  currentCount,
  pageInfo,
  isLoading,
  onPrev,
  onNext,
}: {
  totalCount: number;
  currentCount: number;
  pageInfo?: PageInfo | null;
  isLoading: boolean;
  onPrev: () => void;
  onNext: () => void;
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-3 border-t border-[#dde3e8] bg-[#f9fafb] px-4 py-3 text-xs text-[#5a6672]">
      <span className={styles.monoText}>
        Menampilkan <strong className="text-[#0b2439]">{currentCount}</strong>{" "}
        dari <strong className="text-[#0b2439]">{totalCount}</strong> entri
      </span>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onPrev}
          disabled={isLoading || !pageInfo?.hasPreviousPage}
          className={`${styles.btnSecondary} ${styles.btnSmall}`}
        >
          <IdentityIcon name="chevronLeft" size={14} />
          <span>Sebelumnya</span>
        </button>
        <button
          type="button"
          onClick={onNext}
          disabled={isLoading || !pageInfo?.hasNextPage}
          className={`${styles.btnSecondary} ${styles.btnSmall}`}
        >
          <span>Berikutnya</span>
          <IdentityIcon name="chevronRight" size={14} />
        </button>
      </div>
    </div>
  );
}

export function SlideOverDrawer({
  open,
  title,
  subtitle,
  onClose,
  children,
}: {
  open: boolean;
  title: string;
  subtitle?: string;
  onClose: () => void;
  children: ReactNode;
}) {
  useEffect(() => {
    if (!open) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex justify-end bg-[#0b2439]/35 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label={title}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="flex h-full w-full max-w-3xl flex-col border-l border-[#dde3e8] bg-white shadow-[-16px_0_40px_rgba(11,36,57,0.12)]">
        <div className="flex items-start justify-between gap-4 border-b border-[#dde3e8] bg-[#f4f7f8] px-6 py-4">
          <div className="min-w-0">
            <h2 className="truncate text-lg font-semibold tracking-tight text-[#0b2439]">
              {title}
            </h2>
            {subtitle ? (
              <p className={`mt-0.5 truncate text-xs text-[#5a6672] ${styles.monoText}`}>
                {subtitle}
              </p>
            ) : null}
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup panel detail"
            className={styles.iconBtn}
          >
            <IdentityIcon name="close" size={18} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-6">{children}</div>
      </div>
    </div>
  );
}

export function ConfirmModal({
  open,
  title,
  description,
  confirmLabel,
  danger = true,
  isBusy = false,
  onCancel,
  onConfirm,
}: {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  danger?: boolean;
  isBusy?: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  useEffect(() => {
    if (!open) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape" && !isBusy) onCancel();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, isBusy, onCancel]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
      role="alertdialog"
      aria-modal="true"
      aria-label={title}
    >
      <div className="w-full max-w-md rounded-xl border border-[#dde3e8] bg-white p-6 shadow-xl">
        <h3 className="text-base font-semibold text-[#0b2439]">{title}</h3>
        <p className="mt-2 text-sm leading-relaxed text-[#5a6672]">{description}</p>
        <div className="mt-6 flex items-center justify-end gap-2.5">
          <button
            type="button"
            onClick={onCancel}
            disabled={isBusy}
            className={styles.btnSecondary}
          >
            Batal
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isBusy}
            className={danger ? styles.btnDanger : styles.btnPrimary}
          >
            {isBusy ? "Memproses..." : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
