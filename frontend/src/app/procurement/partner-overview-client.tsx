"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { ErrorPanel } from "@/components/identity/identity-primitives";
import {
  EfCodePill,
  EfScopeDecisionBox,
  EfStatStrip,
} from "@/components/workspace/ef-ui-primitives";
import {
  PARTY_ROLE_KIND_OPTIONS,
  formatEfError,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import {
  evaluateRepresentativePortalAccess,
  fetchAllParties,
  fetchAllPartyAddressLinks,
  fetchAllPartyContacts,
  fetchAllPartyRoles,
  fetchEfPartnerNode,
  validatePartyRoleScope,
} from "@/lib/graphql/ef-partner-queries";
import type {
  PartyRoleKind,
  PortalAccessEvaluation,
  ScopeTargetDecision,
} from "@/lib/graphql/ef-types";

export default function PartnerOverviewClient() {
  const { snapshot, pushToast } = useIdentityWorkspace();
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [counts, setCounts] = useState({
    parties: 0,
    roles: 0,
    contacts: 0,
    addresses: 0,
  });

  const [evalPartyId, setEvalPartyId] = useState("");
  const [evalUserId, setEvalUserId] = useState("");
  const [evalRole, setEvalRole] = useState<PartyRoleKind>("VENDOR");
  const [evalResult, setEvalResult] = useState<PortalAccessEvaluation | null>(
    null,
  );
  const [evalError, setEvalError] = useState<string | null>(null);

  const [scopePartyId, setScopePartyId] = useState("");
  const [scopeRole, setScopeRole] = useState<PartyRoleKind>("SUPPLIER");
  const [decision, setDecision] = useState<ScopeTargetDecision | null>(null);
  const [scopeError, setScopeError] = useState<string | null>(null);

  const [nodeInput, setNodeInput] = useState("");
  const [nodeResult, setNodeResult] = useState<Record<string, unknown> | null>(
    null,
  );

  const fetchOverviewData = useCallback(async () => {
    const canReadParties = hasEfPermission(
      snapshot,
      "enterprise_foundation.parties.read",
    );
    if (!canReadParties) {
      return null;
    }
    const [pRes, rRes, cRes, aRes] = await Promise.all([
      fetchAllParties(5),
      fetchAllPartyRoles(5),
      fetchAllPartyContacts(5),
      fetchAllPartyAddressLinks(5),
    ]);
    return { pRes, rRes, cRes, aRes };
  }, [snapshot]);

  const loadOverview = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const data = await fetchOverviewData();
      if (!data) {
        setCounts({ parties: 0, roles: 0, contacts: 0, addresses: 0 });
        return;
      }
      const { pRes, rRes, cRes, aRes } = data;
      setCounts({
        parties: pRes.total_count,
        roles: rRes.total_count,
        contacts: cRes.total_count,
        addresses: aRes.total_count,
      });
      if (pRes.nodes[0]) {
        const firstPartyId = pRes.nodes[0].id;
        setEvalPartyId((prev) => prev || firstPartyId);
        setScopePartyId((prev) => prev || firstPartyId);
      }
      if (snapshot?.principal_id) {
        const principalId = snapshot.principal_id;
        setEvalUserId((prev) => prev || principalId);
      }
    } catch (err) {
      setLoadError(formatEfError(err));
    } finally {
      setLoading(false);
    }
  }, [fetchOverviewData, snapshot]);

  useEffect(() => {
    let active = true;
    async function initOverview() {
      try {
        const data = await fetchOverviewData();
        if (!active) return;
        if (!data) {
          setCounts({ parties: 0, roles: 0, contacts: 0, addresses: 0 });
          setLoadError(null);
          return;
        }
        const { pRes, rRes, cRes, aRes } = data;
        setCounts({
          parties: pRes.total_count,
          roles: rRes.total_count,
          contacts: cRes.total_count,
          addresses: aRes.total_count,
        });
        if (pRes.nodes[0]) {
          const firstPartyId = pRes.nodes[0].id;
          setEvalPartyId((prev) => prev || firstPartyId);
          setScopePartyId((prev) => prev || firstPartyId);
        }
        if (snapshot?.principal_id) {
          const principalId = snapshot.principal_id;
          setEvalUserId((prev) => prev || principalId);
        }
        setLoadError(null);
      } catch (err) {
        if (!active) return;
        setLoadError(formatEfError(err));
      } finally {
        if (active) setLoading(false);
      }
    }
    void initOverview();
    return () => {
      active = false;
    };
  }, [fetchOverviewData, snapshot]);

  async function handleEvaluatePortal() {
    if (!evalPartyId.trim() || !evalUserId.trim()) {
      setEvalError("Masukkan Party ID dan IAM User ID.");
      return;
    }
    setEvalResult(null);
    setEvalError(null);
    try {
      const res = await evaluateRepresentativePortalAccess(
        evalPartyId.trim(),
        evalUserId.trim(),
        evalRole,
      );
      setEvalResult(res);
    } catch (err) {
      setEvalError(formatEfError(err));
    }
  }

  async function handleValidateScope() {
    if (!scopePartyId.trim()) {
      setScopeError("Masukkan Party ID untuk divalidasi.");
      return;
    }
    setDecision(null);
    setScopeError(null);
    try {
      const res = await validatePartyRoleScope(scopePartyId.trim(), scopeRole);
      setDecision(res);
    } catch (err) {
      setScopeError(formatEfError(err));
    }
  }

  async function handleInspectNode() {
    if (!nodeInput.trim()) return;
    try {
      const node = await fetchEfPartnerNode(nodeInput.trim());
      setNodeResult(node);
      pushToast("info", "Node Mitra Dimuat", nodeInput.trim());
    } catch (err) {
      pushToast("error", "Inspeksi Node Gagal", formatEfError(err));
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-[#0b2439]">
            Master Mitra &amp; Rekanan — Procurement
          </h2>
          <p className="mt-0.5 text-sm text-[#5a6672]">
            Kelola supplier bahan baku, vendor CMT, peran bisnis, PIC portal IAM,
            dan relasi alamat mitra IntegriOne.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void loadOverview()}
          disabled={loading}
          className={styles.btnSecondary}
        >
          <IdentityIcon name="refresh" size={15} />
          <span>{loading ? "Memuat..." : "Muat Ulang"}</span>
        </button>
      </div>

      {loadError ? (
        <ErrorPanel message={loadError} onRetry={() => void loadOverview()} />
      ) : null}

      <EfStatStrip
        items={[
          { label: "Total Mitra (Party)", value: counts.parties, helper: "Supplier & Vendor" },
          { label: "Peran Bisnis Aktif", value: counts.roles, helper: "PartyRole" },
          { label: "Kontak PIC & Portal", value: counts.contacts, helper: "PartyContact IAM" },
          { label: "Tautan Alamat Mitra", value: counts.addresses, helper: "PartyAddress" },
        ]}
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <section className={`${styles.hairlinePanel} space-y-4 p-5`}>
          <div>
            <h3 className="text-base font-semibold text-[#0b2439]">
              Evaluator Akses Portal Perwakilan (`evaluateRepresentativePortalAccess`)
            </h3>
            <p className="text-xs text-[#5a6672]">
              Memeriksa kelayakan user IAM mewakili mitra pada peran tertentu.
            </p>
          </div>

          <div className="space-y-2.5">
            <input
              type="text"
              value={evalPartyId}
              onChange={(e) => setEvalPartyId(e.target.value)}
              placeholder="Party ID (UUID / gid://integrione/party/...)"
              aria-label="Party ID evaluator portal"
              className={`${styles.fieldInput} ${styles.monoText}`}
            />
            <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
              <input
                type="text"
                value={evalUserId}
                onChange={(e) => setEvalUserId(e.target.value)}
                placeholder="IAM User ID"
                aria-label="IAM User ID evaluator portal"
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
              <select
                aria-label="Peran wajib evaluator"
                value={evalRole}
                onChange={(e) => setEvalRole(e.target.value as PartyRoleKind)}
                className={styles.fieldSelect}
              >
                {PARTY_ROLE_KIND_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
            <button
              type="button"
              onClick={() => void handleEvaluatePortal()}
              className={styles.btnPrimary}
            >
              Evaluasi Hak Portal
            </button>
          </div>

          {evalError ? (
            <div className="rounded-lg border border-[#fecdca] bg-[#fef3f2] p-3 text-xs text-[#b42318]">
              {evalError}
            </div>
          ) : null}

          {evalResult ? (
            <div
              className={`rounded-lg border p-3 text-xs ${
                evalResult.allowed
                  ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                  : "border-[#fecdca] bg-[#fef3f2] text-[#b42318]"
              }`}
            >
              <p className="font-semibold">
                {evalResult.allowed ? "Akses Portal Diizinkan" : "Akses Portal Ditolak"}
              </p>
              <div className={`mt-1.5 flex flex-wrap gap-1 ${styles.monoText}`}>
                {evalResult.active_roles.map((r) => (
                  <EfCodePill key={r} code={r} />
                ))}
              </div>
            </div>
          ) : null}

          <div className="grid grid-cols-2 gap-2 pt-1">
            <Link href="/procurement/partners" className={styles.btnPrimary}>
              Direktori Mitra
            </Link>
            <Link href="/procurement/roles" className={styles.btnSecondary}>
              Peran Bisnis
            </Link>
            <Link href="/procurement/contacts" className={styles.btnSecondary}>
              PIC &amp; Portal IAM
            </Link>
            <Link href="/procurement/addresses" className={styles.btnSecondary}>
              Alamat Mitra
            </Link>
          </div>
        </section>

        <section className={`${styles.hairlinePanel} space-y-4 p-5`}>
          <div>
            <h3 className="text-base font-semibold text-[#0b2439]">
              Validator Target Scope Mitra &amp; Relay Node
            </h3>
            <p className="text-xs text-[#5a6672]">
              Menguji `validatePartyRoleScope` untuk `SUPPLIER_PARTY` &amp; `VENDOR_PARTY`.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-3">
            <input
              type="text"
              value={scopePartyId}
              onChange={(e) => setScopePartyId(e.target.value)}
              placeholder="Party ID..."
              aria-label="Party ID validasi scope"
              className={`${styles.fieldInput} sm:col-span-2 ${styles.monoText}`}
            />
            <select
              aria-label="Required role validator"
              value={scopeRole}
              onChange={(e) => setScopeRole(e.target.value as PartyRoleKind)}
              className={styles.fieldSelect}
            >
              <option value="SUPPLIER">SUPPLIER</option>
              <option value="VENDOR">VENDOR</option>
              <option value="CARRIER">CARRIER</option>
            </select>
          </div>

          <button
            type="button"
            onClick={() => void handleValidateScope()}
            className={styles.btnPrimary}
          >
            Validasi Party Scope
          </button>

          <EfScopeDecisionBox decision={decision} error={scopeError} />

          <div className="border-t border-[#dde3e8] pt-3">
            <label
              htmlFor="ef-partner-node-gid"
              className="block text-xs font-semibold text-[#0b2439]"
            >
              Inspektor Global Relay `Query.node(id)`
            </label>
            <div className="mt-1.5 flex gap-2">
              <input
                id="ef-partner-node-gid"
                type="text"
                value={nodeInput}
                onChange={(e) => setNodeInput(e.target.value)}
                placeholder="gid://integrione/party/..."
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
              <button
                type="button"
                onClick={() => void handleInspectNode()}
                className={styles.btnSecondary}
              >
                Query Node
              </button>
            </div>
            {nodeResult ? (
              <pre
                className={`mt-2 max-h-40 overflow-auto rounded-lg border border-[#dde3e8] bg-[#f9fafb] p-3 text-xs ${styles.monoText}`}
              >
                {JSON.stringify(nodeResult, null, 2)}
              </pre>
            ) : null}
          </div>
        </section>
      </div>
    </div>
  );
}
