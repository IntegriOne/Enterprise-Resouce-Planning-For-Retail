"use client";

import { useState } from "react";
import styles from "@/components/identity/identity.module.css";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  SlideOverDrawer,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  EfCodePill,

  EfScopeDecisionBox,
  EfStatusBadge,
} from "@/components/workspace/ef-ui-primitives";
import { formatEfError } from "@/lib/graphql/ef-helpers";
import {
  activateLegalEntity,
  dissolveLegalEntity,
  setPrimaryOperatingLegalEntity,
  suspendLegalEntity,
} from "@/lib/graphql/ef-structure-mutations";
import {

  fetchLegalEntityChildren,
  validateLegalEntityScopeTarget,
} from "@/lib/graphql/ef-structure-queries";
import type { EfLegalEntity, ScopeTargetDecision } from "@/lib/graphql/ef-types";

export function LegalEntityDetailDrawer({
  entity,
  canWrite,
  onClose,
  onMutated,
}: {
  entity: EfLegalEntity | null;
  canWrite: boolean;
  onClose: () => void;
  onMutated: () => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();
  const [decision, setDecision] = useState<ScopeTargetDecision | null>(null);
  const [decisionErr, setDecisionErr] = useState<string | null>(null);
  const [children, setChildren] = useState<EfLegalEntity[] | null>(null);

  if (!entity) return null;
  const activeEntity = entity;

  async function runLifecycle(
    label: string,
    fn: () => Promise<EfLegalEntity | null>,
    sensitive = false,
  ) {
    const updated = await runDashboardAction(
      fn,
      sensitive
        ? { stepUpReason: `Aksi "${label}" pada badan hukum bersifat sensitif.` }
        : undefined,
    );
    if (updated) {
      pushToast("success", label, updated.legal_name);
      onMutated();
    }
  }

  async function handleValidateScope() {
    setDecision(null);
    setDecisionErr(null);
    try {
      const res = await validateLegalEntityScopeTarget(activeEntity.id);
      setDecision(res);
    } catch (err) {
      setDecisionErr(formatEfError(err));
    }
  }

  async function handleLoadChildren() {
    try {
      const conn = await fetchLegalEntityChildren(activeEntity.id, 20);
      setChildren(conn.nodes);
    } catch (err) {
      pushToast("error", "Gagal Memuat Entitas Anak", formatEfError(err));
    }
  }

  return (
    <SlideOverDrawer
      open={Boolean(entity)}
      title={activeEntity.legal_name}
      subtitle={activeEntity.entity_code}
      onClose={onClose}
    >
      <div className="space-y-5">
        <div className="flex flex-wrap items-center gap-2">
          <EfCodePill code={activeEntity.entity_code} />
          <EfStatusBadge label={activeEntity.status} />
          {activeEntity.is_primary_operating_entity ? (
            <EfStatusBadge label="PRIMARY OPERATING" toneSource="ACTIVE" />
          ) : null}
          {activeEntity.is_pkp ? <EfCodePill code="PKP" /> : null}

        </div>

        <dl className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-4 text-xs sm:grid-cols-2">
          <div>
            <dt className="font-semibold text-[#5a6672]">Nama Dagang</dt>
            <dd className="mt-0.5 text-[#0b2439]">
              {activeEntity.trading_name || "-"}
            </dd>
          </div>
          <div>
            <dt className="font-semibold text-[#5a6672]">NPWP / NIB</dt>
            <dd className={`mt-0.5 text-[#0b2439] ${styles.monoText}`}>
              {activeEntity.tax_id_number || "-"} / {activeEntity.nib || "-"}
            </dd>
          </div>
          <div>
            <dt className="font-semibold text-[#5a6672]">Yurisdiksi / Fiskal</dt>
            <dd className="mt-0.5 text-[#0b2439]">
              {activeEntity.jurisdiction_country_code} (Bulan {activeEntity.fiscal_year_start_month})
            </dd>
          </div>
          <div>
            <dt className="font-semibold text-[#5a6672]">Diperbarui</dt>
            <dd className="mt-0.5 text-[#0b2439]">
              {formatDateTimeID(activeEntity.updated_at)}
            </dd>
          </div>
        </dl>

        {canWrite ? (
          <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
              Kontrol Siklus Hidup Badan Hukum
            </h3>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() =>
                  void runLifecycle("Badan Hukum Diaktifkan", () =>
                    activateLegalEntity(activeEntity.id, csrfToken),
                  )
                }
                className={`${styles.btnPrimary} ${styles.btnSmall}`}
              >
                Aktifkan
              </button>
              <button
                type="button"
                onClick={() =>
                  void runLifecycle(
                    "Tetapkan Entitas Utama",
                    () =>
                      setPrimaryOperatingLegalEntity(activeEntity.id, csrfToken),
                    true,
                  )
                }
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Set Utama (Sensitif)
              </button>
              <button
                type="button"
                onClick={() =>
                  void runLifecycle("Badan Hukum Ditangguhkan", () =>
                    suspendLegalEntity(activeEntity.id, csrfToken),
                  )
                }
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Suspend
              </button>
              <button
                type="button"
                onClick={() =>
                  void runLifecycle(
                    "Likuidasi / Dissolve Entitas",
                    () =>
                      dissolveLegalEntity(
                        activeEntity.id,
                        new Date().toISOString(),
                        csrfToken,
                      ),
                    true,
                  )
                }
                className={`${styles.btnDanger} ${styles.btnSmall}`}
              >
                Dissolve (Sensitif)
              </button>
            </div>
          </div>
        ) : null}

        <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
              Validasi Target Scope IAM &amp; Anak Entitas
            </h3>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => void handleValidateScope()}
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Validasi Scope IAM
              </button>
              <button
                type="button"
                onClick={() => void handleLoadChildren()}
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Muat Children
              </button>
            </div>
          </div>

          <EfScopeDecisionBox decision={decision} error={decisionErr} />

          {children ? (
            <div className="space-y-1.5 pt-2">
              <p className="text-xs font-semibold text-[#0b2439]">
                Entitas Anak ({children.length})
              </p>
              {children.length === 0 ? (
                <p className="text-xs text-[#5a6672]">Tidak ada entitas anak.</p>
              ) : (
                children.map((ch) => (
                  <div
                    key={ch.id}
                    className="flex items-center justify-between rounded-lg border border-[#dde3e8] px-3 py-2 text-xs"
                  >
                    <span>
                      <strong className={styles.monoText}>{ch.entity_code}</strong>{" "}
                      {ch.legal_name}
                    </span>
                    <EfStatusBadge label={ch.status} />
                  </div>
                ))
              )}
            </div>
          ) : null}
        </div>
      </div>
    </SlideOverDrawer>
  );
}
