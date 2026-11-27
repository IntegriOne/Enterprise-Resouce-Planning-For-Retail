"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "next/navigation";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import {
  ActiveBooleanBadge,
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  TableSkeleton,
} from "@/components/identity/identity-primitives";
import {
  fetchRolesConnection,
  formatIamDashboardError,
  isUnauthenticatedError,
  type RolesFilterVars,
} from "@/lib/graphql/iam-dashboard-queries";
import { MANAGEMENT_DOMAIN_PRESETS } from "@/lib/graphql/iam-role-access-helpers";
import type { IamRole, RoleConnection } from "@/lib/graphql/iam-types";
import { RoleFormModal } from "./role-form-modal";
import { RoleDetailDrawer } from "./role-detail-drawer";

export default function RolesClient() {
  const searchParams = useSearchParams();
  const initialAction = searchParams.get("action");
  const initialDomain = searchParams.get("domain") || undefined;
  const initialRoleId = searchParams.get("roleId");

  const [searchInput, setSearchInput] = useState("");
  const [appliedQuery, setAppliedQuery] = useState("");
  const [domainFilter, setDomainFilter] = useState("");
  const [activeOnly, setActiveOnly] = useState(false);

  const [conn, setConn] = useState<RoleConnection | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [formOpen, setFormOpen] = useState(initialAction === "create");
  const [editingRole, setEditingRole] = useState<IamRole | null>(null);
  const [activeRoleId, setActiveRoleId] = useState<string | null>(initialRoleId);

  const loadRoles = useCallback(
    async (cursorVars: Partial<RolesFilterVars> = { first: 20 }) => {
      setIsLoading(true);
      setErrorMsg(null);
      try {
        const res = await fetchRolesConnection({
          query: appliedQuery || null,
          management_domain: domainFilter || null,
          active_only: activeOnly,
          ...cursorVars,
        });
        setConn(res);
      } catch (err) {
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setErrorMsg(formatIamDashboardError(err));
      } finally {
        setIsLoading(false);
      }
    },
    [appliedQuery, domainFilter, activeOnly],
  );

  useEffect(() => {
    let active = true;
    async function init() {
      try {
        const res = await fetchRolesConnection({
          query: appliedQuery || null,
          management_domain: domainFilter || null,
          active_only: activeOnly,
          first: 20,
        });
        if (!active) return;
        setConn(res);
        setErrorMsg(null);
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setErrorMsg(formatIamDashboardError(err));
      } finally {
        if (active) setIsLoading(false);
      }
    }
    void init();
    return () => {
      active = false;
    };
  }, [appliedQuery, domainFilter, activeOnly]);

  function handleSearchSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAppliedQuery(searchInput.trim());
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col justify-between gap-4 border-b border-[#dde3e8] pb-5 sm:flex-row sm:items-end">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-[#0b2439]">
            Peran RBAC &amp; Hak Akses (`roles`)
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-[#5a6672]">
            Kelola peran fungsional, pemetaan izin atomik, serta prioritas
            redirect pasca-autentikasi. Peran sistem (<code className={styles.monoText}>is_system</code>)
            dikelola otomatis oleh katalog startup; buat peran kustom baru untuk
            mengatur izin domain sesuai kebutuhan operasional.
          </p>
        </div>

        <button
          type="button"
          onClick={() => {
            setEditingRole(null);
            setFormOpen(true);
          }}
          className={styles.btnPrimary}
        >
          <IdentityIcon name="plus" size={16} />
          <span>Buat Peran Baru</span>
        </button>
      </div>

      {/* Filter Cepat Domain */}
      <div className="flex flex-wrap items-center gap-1.5">
        <button
          type="button"
          onClick={() => setDomainFilter("")}
          className={`rounded-md border px-2.5 py-1 text-xs font-semibold transition-colors ${
            domainFilter === ""
              ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
              : "border-[#dde3e8] bg-white text-[#5a6672] hover:text-[#0b2439]"
          }`}
        >
          Semua Domain
        </button>
        {MANAGEMENT_DOMAIN_PRESETS.slice(0, 5).map((preset) => (
          <button
            key={preset.value}
            type="button"
            onClick={() => setDomainFilter(preset.value)}
            className={`rounded-md border px-2.5 py-1 text-xs font-semibold transition-colors ${styles.monoText} ${
              domainFilter === preset.value
                ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                : "border-[#dde3e8] bg-white text-[#5a6672] hover:text-[#0b2439]"
            }`}
          >
            {preset.value}
          </button>
        ))}
      </div>

      <form
        onSubmit={handleSearchSubmit}
        className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-white p-4 md:grid-cols-12"
      >
        <div className="md:col-span-6">
          <label htmlFor="roles-search" className="mb-1 block text-xs font-semibold text-[#0b2439]">
            Cari Nama / Kode Peran
          </label>
          <div className="flex gap-2">
            <input
              id="roles-search"
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Contoh: superadmin, iam_operator, ef_operator..."
              className={styles.fieldInput}
            />
            <button type="submit" className={`${styles.btnSecondary} shrink-0`}>
              Cari
            </button>
          </div>
        </div>

        <div className="md:col-span-4">
          <label htmlFor="roles-domain" className="mb-1 block text-xs font-semibold text-[#0b2439]">
            Filter Management Domain
          </label>
          <input
            id="roles-domain"
            type="text"
            value={domainFilter}
            onChange={(e) => setDomainFilter(e.target.value)}
            placeholder="iam / enterprise_foundation / human_resource"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
        </div>

        <div className="flex items-end md:col-span-2">
          <label className="flex min-h-[42px] w-full cursor-pointer items-center gap-2 rounded-lg border border-[#dde3e8] bg-[#f9fafb] px-3 text-xs font-medium text-[#0b2439]">
            <input
              type="checkbox"
              checked={activeOnly}
              onChange={(e) => setActiveOnly(e.target.checked)}
              className="h-4 w-4 accent-[#133554]"
            />
            <span>Hanya Aktif</span>
          </label>
        </div>
      </form>

      {errorMsg ? (
        <ErrorPanel
          message={errorMsg}
          onRetry={() => void loadRoles({ first: 20 })}
        />
      ) : (
        <div className={`${styles.hairlinePanel} overflow-hidden`}>
          {isLoading && !conn ? (
            <TableSkeleton rows={6} cols={5} />
          ) : !conn || conn.nodes.length === 0 ? (
            <EmptyPanel
              title="Tidak Ada Peran Ditemukan"
              description="Belum ada peran RBAC yang sesuai dengan filter saat ini. Buat peran kustom baru jika domain tersebut belum memiliki peran operasional manusia."
              action={
                <button
                  type="button"
                  onClick={() => {
                    setEditingRole(null);
                    setFormOpen(true);
                  }}
                  className={styles.btnPrimary}
                >
                  Buat Peran Baru
                </button>
              }
            />
          ) : (
            <>
              <div className={styles.tableWrap}>
                <table className={styles.dataTable}>
                  <thead>
                    <tr>
                      <th>Peran &amp; Kode</th>
                      <th>Domain</th>
                      <th>Default Redirect</th>
                      <th>Status</th>
                      <th className="text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {conn.nodes.map((r) => (
                      <tr key={r.id}>
                        <td>
                          <p className="font-semibold text-[#0b2439]">{r.name}</p>
                          <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                            {r.code} {r.is_system ? "· SYSTEM" : "· KUSTOM"}
                          </p>
                        </td>
                        <td className={`text-xs text-[#133554] ${styles.monoText}`}>
                          {r.management_domain || "-"}
                        </td>
                        <td className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                          {r.default_redirect_path || "(none)"} · P{r.redirect_priority}
                        </td>
                        <td>
                          <ActiveBooleanBadge active={r.is_active} />
                        </td>
                        <td className="text-right">
                          <button
                            type="button"
                            onClick={() => setActiveRoleId(r._id)}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            Izin &amp; Detail
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <RelayPaginationBar
                totalCount={conn.total_count}
                currentCount={conn.nodes.length}
                pageInfo={conn.pageInfo}
                isLoading={isLoading}
                onPrev={() =>
                  void loadRoles({
                    last: 20,
                    before: conn.pageInfo.startCursor || undefined,
                  })
                }
                onNext={() =>
                  void loadRoles({
                    first: 20,
                    after: conn.pageInfo.endCursor || undefined,
                  })
                }
              />
            </>
          )}
        </div>
      )}

      <RoleFormModal
        open={formOpen}
        editingRole={editingRole}
        initialDomain={initialDomain}
        onClose={() => setFormOpen(false)}
        onSaved={(saved) => {
          void loadRoles({ first: 20 });
          if (saved?._id) setActiveRoleId(saved._id);
        }}
      />

      <RoleDetailDrawer
        roleId={activeRoleId}
        onClose={() => setActiveRoleId(null)}
        onEditRequested={(role) => {
          setEditingRole(role);
          setFormOpen(true);
        }}
        onUpdated={() => void loadRoles({ first: 20 })}
      />
    </div>
  );
}
