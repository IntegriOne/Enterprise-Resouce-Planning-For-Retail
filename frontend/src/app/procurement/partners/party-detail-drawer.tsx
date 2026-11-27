"use client";

import { useState } from "react";
import Link from "next/link";
import styles from "@/components/identity/identity.module.css";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ActiveBooleanBadge,
  SlideOverDrawer,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  EfCodePill,
  EfRelayNodeInspectorButton,
  EfStatusBadge,
} from "@/components/workspace/ef-ui-primitives";
import { PARTY_ROLE_KIND_OPTIONS } from "@/lib/graphql/ef-helpers";
import {
  assignPartyRole,
  deactivateParty,
  suspendParty,
  verifyAndActivateParty,
} from "@/lib/graphql/ef-partner-mutations";
import { fetchEfPartnerNode } from "@/lib/graphql/ef-partner-queries";
import type { EfParty, PartyRoleKind } from "@/lib/graphql/ef-types";

export function PartyDetailDrawer({
  party,
  canWrite,
  onClose,
  onMutated,
}: {
  party: EfParty | null;
  canWrite: boolean;
  onClose: () => void;
  onMutated: () => void;
}) {
  const { csrfToken, pushToast, runDashboardAction } = useIdentityWorkspace();
  const [newRole, setNewRole] = useState<PartyRoleKind>("VENDOR");
  const [roleNote, setRoleNote] = useState("");

  if (!party) return null;
  const activeParty = party;

  async function handleStatusAction(
    label: string,
    fn: () => Promise<EfParty | null>,
  ) {
    const updated = await runDashboardAction(fn);
    if (updated) {
      pushToast("success", label, updated.display_name);
      onMutated();
    }
  }

  async function handleAssignRole() {
    const added = await runDashboardAction(() =>
      assignPartyRole(activeParty.id, newRole, roleNote.trim() || null, csrfToken),
    );
    if (added) {
      setRoleNote("");
      pushToast("success", "Peran Mitra Ditambahkan", added.role);
      onMutated();
    }
  }

  const roles = activeParty.roles?.nodes ?? [];
  const contacts = activeParty.contacts?.nodes ?? [];
  const addresses = activeParty.addresses?.nodes ?? [];

  return (
    <SlideOverDrawer
      open={Boolean(party)}
      title={activeParty.display_name}
      subtitle={`${activeParty.partner_code} · ${activeParty._id}`}
      onClose={onClose}
    >
      <div className="space-y-5">
        <div className="flex flex-wrap items-center gap-2">
          <EfCodePill code={activeParty.kind} />
          <EfStatusBadge label={activeParty.status} />
          {activeParty.is_pkp ? <EfCodePill code="PKP" /> : null}
          <EfRelayNodeInspectorButton
            gid={activeParty.id}
            onInspect={fetchEfPartnerNode}
          />
        </div>

        <dl className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-[#f9fafb] p-4 text-xs sm:grid-cols-2">
          <div>
            <dt className="font-semibold text-[#5a6672]">Nama Legal</dt>
            <dd className="mt-0.5 text-[#0b2439]">{activeParty.legal_name || "-"}</dd>
          </div>
          <div>
            <dt className="font-semibold text-[#5a6672]">Email / Telepon</dt>
            <dd className="mt-0.5 text-[#0b2439]">
              {activeParty.email || "-"} · {activeParty.phone || "-"}
            </dd>
          </div>
          <div>
            <dt className="font-semibold text-[#5a6672]">NPWP</dt>
            <dd className={`mt-0.5 text-[#0b2439] ${styles.monoText}`}>
              {activeParty.tax_id_number || "-"}
            </dd>
          </div>
          <div>
            <dt className="font-semibold text-[#5a6672]">Diperbarui</dt>
            <dd className="mt-0.5 text-[#0b2439]">
              {formatDateTimeID(activeParty.updated_at)}
            </dd>
          </div>
        </dl>

        {canWrite ? (
          <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
            <p className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
              Siklus Hidup Verifikasi Mitra
            </p>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() =>
                  void handleStatusAction("Mitra Diverifikasi & Aktif", () =>
                    verifyAndActivateParty(activeParty.id, csrfToken),
                  )
                }
                className={`${styles.btnPrimary} ${styles.btnSmall}`}
              >
                Verifikasi &amp; Aktifkan
              </button>
              <button
                type="button"
                onClick={() =>
                  void handleStatusAction("Mitra Ditangguhkan", () =>
                    suspendParty(activeParty.id, csrfToken),
                  )
                }
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Suspend
              </button>
              <button
                type="button"
                onClick={() =>
                  void handleStatusAction("Mitra Dinonaktifkan", () =>
                    deactivateParty(activeParty.id, csrfToken),
                  )
                }
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                Nonaktifkan
              </button>
            </div>
          </div>
        ) : null}

        <div className="space-y-3 rounded-xl border border-[#dde3e8] p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
              Peran Bisnis ({roles.length})
            </h3>
            <Link href="/procurement/roles" className="text-xs font-semibold text-[#133554] hover:underline">
              Kelola Penuh →
            </Link>
          </div>
          <div className="space-y-1.5">
            {roles.map((r) => (
              <div
                key={r.id}
                className="flex items-center justify-between rounded-lg border border-[#dde3e8] px-3 py-2 text-xs"
              >
                <div>
                  <EfCodePill code={r.role} />
                  {r.reference_note ? (
                    <span className="ml-2 text-[#5a6672]">{r.reference_note}</span>
                  ) : null}
                </div>
                <ActiveBooleanBadge active={r.is_active} />
              </div>
            ))}
          </div>
          {canWrite ? (
            <div className="grid grid-cols-1 gap-2 pt-2 sm:grid-cols-3">
              <select
                aria-label="Tambah peran mitra"
                value={newRole}
                onChange={(e) => setNewRole(e.target.value as PartyRoleKind)}
                className={styles.fieldSelect}
              >
                {PARTY_ROLE_KIND_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <input
                type="text"
                value={roleNote}
                onChange={(e) => setRoleNote(e.target.value)}
                placeholder="Catatan referensi..."
                className={styles.fieldInput}
              />
              <button
                type="button"
                onClick={() => void handleAssignRole()}
                className={`${styles.btnPrimary} ${styles.btnSmall}`}
              >
                Tambah Peran
              </button>
            </div>
          ) : null}
        </div>

        <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
              Kontak PIC &amp; Portal ({contacts.length})
            </h3>
            <Link href="/procurement/contacts" className="text-xs font-semibold text-[#133554] hover:underline">
              Kelola PIC →
            </Link>
          </div>
          {contacts.length === 0 ? (
            <p className="text-xs text-[#5a6672]">Belum ada PIC terdaftar.</p>
          ) : (
            contacts.map((c) => (
              <div
                key={c.id}
                className="flex items-center justify-between rounded-lg border border-[#dde3e8] px-3 py-2 text-xs"
              >
                <div>
                  <p className="font-semibold text-[#0b2439]">{c.full_name}</p>
                  <p className="text-[#5a6672]">
                    {c.contact_role} · {c.email || "-"}
                  </p>
                </div>
                <EfStatusBadge label={c.portal_access_status} />
              </div>
            ))
          )}
        </div>

        <div className="space-y-2 rounded-xl border border-[#dde3e8] p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-[#5a6672]">
              Alamat Mitra ({addresses.length})
            </h3>
            <Link href="/procurement/addresses" className="text-xs font-semibold text-[#133554] hover:underline">
              Kelola Alamat →
            </Link>
          </div>
          {addresses.length === 0 ? (
            <p className="text-xs text-[#5a6672]">Belum ada tautan alamat.</p>
          ) : (
            addresses.map((pa) => (
              <div
                key={pa.id}
                className="flex items-center justify-between rounded-lg border border-[#dde3e8] px-3 py-2 text-xs"
              >
                <div>
                  <EfCodePill code={pa.usage} />
                  <span className="ml-2 text-[#0b2439]">
                    {pa.address?.address_line_1 || pa.address_id}
                  </span>
                </div>
                {pa.is_primary ? (
                  <EfStatusBadge label="PRIMARY" toneSource="ACTIVE" />
                ) : null}
              </div>
            ))
          )}
        </div>
      </div>
    </SlideOverDrawer>
  );
}
