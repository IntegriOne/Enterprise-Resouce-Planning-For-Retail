"use client";

import { useState, type ReactNode } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { SlideOverDrawer } from "@/components/identity/identity-primitives";
import { statusBadgeTone } from "@/lib/graphql/ef-helpers";
import type { ScopeTargetDecision } from "@/lib/graphql/ef-types";

export function EfStatusBadge({
  label,
  toneSource,
}: {
  label: string;
  toneSource?: string;
}) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-md border px-2 py-0.5 text-[11px] font-semibold tracking-wide ${styles.monoText} ${statusBadgeTone(toneSource || label)}`}
    >
      {label}
    </span>
  );
}

export function EfCodePill({ code }: { code: string }) {
  return (
    <span
      className={`inline-flex items-center rounded-md border border-[#dde3e8] bg-[#f4f7f8] px-2 py-0.5 text-xs font-semibold text-[#0b2439] ${styles.monoText}`}
    >
      {code}
    </span>
  );
}

export function EfStatStrip({
  items,
}: {
  items: Array<{
    label: string;
    value: string | number;
    helper?: string;
    accent?: boolean;
  }>;
}) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {items.map((item) => (
        <div
          key={item.label}
          className={`${styles.hairlinePanel} flex flex-col justify-between gap-2 p-4`}
        >
          <span
            className={`text-[11px] font-semibold uppercase tracking-wider text-[#5a6672] ${styles.monoText}`}
          >
            {item.label}
          </span>
          <div className="flex items-baseline justify-between gap-2">
            <span
              className={`text-2xl font-bold tracking-tight ${
                item.accent ? "text-[#da414a]" : "text-[#0b2439]"
              } ${styles.monoText}`}
            >
              {item.value}
            </span>
            {item.helper ? (
              <span className="truncate text-xs text-[#5a6672]">
                {item.helper}
              </span>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}

export function EfScopeDecisionBox({
  decision,
  error,
  showIdentifiers = true,
}: {
  decision: ScopeTargetDecision | null;
  error?: string | null;
  showIdentifiers?: boolean;
}) {
  if (error) {
    return (
      <div
        role="alert"
        className="rounded-lg border border-[#fecdca] bg-[#fef3f2] p-3 text-xs text-[#b42318]"
      >
        <p className="font-semibold">Target Scope Ditolak / Tidak Layak</p>
        <p className="mt-1">{error}</p>
      </div>
    );
  }
  if (!decision) return null;

  return (
    <div
      className={`rounded-lg border p-3 text-xs ${
        decision.eligible
          ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
          : "border-[#fecdca] bg-[#fef3f2] text-[#b42318]"
      }`}
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="font-semibold">
          {decision.eligible
            ? "Target Scope Layak Digunakan IAM"
            : "Target Scope Tidak Layak"}
        </span>
        <span className={styles.monoText}>{decision.scope_kind}</span>
      </div>
      <div className={`mt-2 space-y-0.5 ${styles.monoText}`}>
        <p>Kode: {decision.code || "-"}</p>
        {showIdentifiers ? <p>Target ID: {decision.target_id}</p> : null}
        {showIdentifiers && decision.legal_entity_id ? (
          <p>Legal Entity: {decision.legal_entity_id}</p>
        ) : null}
      </div>
    </div>
  );
}

export function EfRelayNodeInspectorButton({
  gid,
  label = "Inspeksi Node",
  onInspect,
}: {
  gid: string;
  label?: string;
  onInspect: (gid: string) => Promise<Record<string, unknown> | null>;
}) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [payload, setPayload] = useState<Record<string, unknown> | null>(null);
  const [err, setErr] = useState<string | null>(null);

  async function handleOpen() {
    setOpen(true);
    setLoading(true);
    setErr(null);
    try {
      const node = await onInspect(gid);
      setPayload(node);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Gagal memuat Relay Node");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <button
        type="button"
        onClick={() => void handleOpen()}
        className={`${styles.btnSecondary} ${styles.btnSmall}`}
      >
        <IdentityIcon name="search" size={13} />
        <span>{label}</span>
      </button>

      <SlideOverDrawer
        open={open}
        title="Inspektor Relay Global Node"
        subtitle={gid}
        onClose={() => setOpen(false)}
      >
        {loading ? (
          <p className="text-sm text-[#5a6672]">Mengambil Query.node(id)...</p>
        ) : err ? (
          <div className="rounded-lg border border-[#fecdca] bg-[#fef3f2] p-3 text-xs text-[#b42318]">
            {err}
          </div>
        ) : (
          <pre
            className={`overflow-x-auto rounded-lg border border-[#dde3e8] bg-[#f9fafb] p-4 text-xs text-[#0b2439] ${styles.monoText}`}
          >
            {JSON.stringify(payload, null, 2)}
          </pre>
        )}
      </SlideOverDrawer>
    </>
  );
}

export function EfToolbarCard({
  title,
  description,
  actions,
  children,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
  children?: ReactNode;
}) {
  return (
    <section className={`${styles.hairlinePanel} p-4 sm:p-5`}>
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-[#0b2439]">{title}</h2>
          {description ? (
            <p className="mt-0.5 text-sm text-[#5a6672]">{description}</p>
          ) : null}
        </div>
        {actions ? (
          <div className="flex flex-wrap items-center gap-2">{actions}</div>
        ) : null}
      </div>
      {children ? <div className="mt-4">{children}</div> : null}
    </section>
  );
}
