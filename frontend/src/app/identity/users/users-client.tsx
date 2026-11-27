"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "next/navigation";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import {
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  TableSkeleton,
  UserKindBadge,
  UserStatusBadge,
  formatDateTimeID,
} from "@/components/identity/identity-primitives";
import {
  fetchUsersConnection,
  formatIamDashboardError,
  isUnauthenticatedError,
  type UsersFilterVars,
} from "@/lib/graphql/iam-dashboard-queries";
import type {
  ProvisionUserKind,
  UserConnection,
  UserKind,
  UserStatus,
} from "@/lib/graphql/iam-types";
import { UserCreateModal } from "./user-create-modal";
import { UserDetailDrawer, type DetailTab } from "./user-detail-drawer";

export type UserSegment = "all" | "customers" | "vendors" | "internals";

type SegmentKindTab = {
  kind: UserKind;
  label: string;
};

type SegmentConfig = {
  kicker: string;
  title: string;
  subtitle: string;
  defaultKindFilter: UserKind | "";
  tabs: SegmentKindTab[] | null;
  allowedCreateKinds: ProvisionUserKind[];
};

const SEGMENT_CONFIGS: Record<UserSegment, SegmentConfig> = {
  all: {
    kicker: "DIRECTORY · SEMUA KATEGORI",
    title: "Direktori Pengguna (`users`)",
    subtitle:
      "Kelola seluruh akun lintas kategori (Customer, Vendor, Supplier, Employee, System), identitas login, dan peran berscope.",
    defaultKindFilter: "",
    tabs: null,
    allowedCreateKinds: ["EMPLOYEE", "VENDOR", "SUPPLIER", "CUSTOMER"],
  },
  customers: {
    kicker: "USER KIND · CUSTOMER",
    title: "Direktori Customers",
    subtitle:
      "Daftar akun pelanggan Web Commerce (kind: CUSTOMER), verifikasi identitas, dan siklus hidup status.",
    defaultKindFilter: "CUSTOMER",
    tabs: [{ kind: "CUSTOMER", label: "Customer" }],
    allowedCreateKinds: ["CUSTOMER"],
  },
  vendors: {
    kicker: "USER KIND · VENDOR & SUPPLIER",
    title: "Direktori Vendors & Mitra",
    subtitle:
      "Daftar akun mitra jasa eksternal (VENDOR) dan pemasok bahan/barang (SUPPLIER).",
    defaultKindFilter: "VENDOR",
    tabs: [
      { kind: "VENDOR", label: "Vendor (Jasa / CMT)" },
      { kind: "SUPPLIER", label: "Supplier (Pemasok)" },
    ],
    allowedCreateKinds: ["VENDOR", "SUPPLIER"],
  },
  internals: {
    kicker: "USER KIND · EMPLOYEE & SYSTEM",
    title: "Direktori Internals",
    subtitle:
      "Daftar akun pegawai/operator internal (EMPLOYEE) serta akun layanan non-interaktif (SYSTEM).",
    defaultKindFilter: "EMPLOYEE",
    tabs: [
      { kind: "EMPLOYEE", label: "Employee (Internal)" },
      { kind: "SYSTEM", label: "System (Layanan)" },
    ],
    allowedCreateKinds: ["EMPLOYEE"],
  },
};

function resolveInitialModalState(
  segment: UserSegment,
  action: string | null,
): { open: boolean; kind: ProvisionUserKind } {
  if (action === "create-employee") {
    return { open: true, kind: "EMPLOYEE" };
  }
  if (action === "create-vendor") {
    return { open: true, kind: "VENDOR" };
  }
  if (action === "create-supplier") {
    return { open: true, kind: "SUPPLIER" };
  }
  if (action === "create-customer") {
    return { open: true, kind: "CUSTOMER" };
  }
  if (action === "create") {
    const defaultKind = SEGMENT_CONFIGS[segment].allowedCreateKinds[0] ?? "EMPLOYEE";
    return { open: true, kind: defaultKind };
  }
  return {
    open: false,
    kind: SEGMENT_CONFIGS[segment].allowedCreateKinds[0] ?? "EMPLOYEE",
  };
}

export default function UsersClient({
  segment = "all",
}: {
  segment?: UserSegment;
}) {
  const searchParams = useSearchParams();
  const config = SEGMENT_CONFIGS[segment];
  const initialModal = resolveInitialModalState(
    segment,
    searchParams.get("action"),
  );

  const [searchInput, setSearchInput] = useState("");
  const [appliedQuery, setAppliedQuery] = useState("");
  const [kindFilter, setKindFilter] = useState<UserKind | "">(
    config.defaultKindFilter,
  );
  const [statusFilter, setStatusFilter] = useState<UserStatus | "">("");
  const [includeDeleted, setIncludeDeleted] = useState(false);

  const [conn, setConn] = useState<UserConnection | null>(null);
  const [tabCounts, setTabCounts] = useState<Partial<Record<UserKind, number>>>(
    {},
  );
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [isCreateOpen, setIsCreateOpen] = useState(initialModal.open);
  const [createKind, setCreateKind] = useState<ProvisionUserKind>(
    initialModal.kind,
  );
  const [createModalNonce, setCreateModalNonce] = useState(0);
  const [activeUserId, setActiveUserId] = useState<string | null>(null);
  const [activeUserTab, setActiveUserTab] = useState<DetailTab>("overview");

  const effectiveKind: UserKind | null = kindFilter || null;

  const refreshSegmentCounts = useCallback(async () => {
    if (!config.tabs || config.tabs.length <= 1) return;
    try {
      const results = await Promise.all(
        config.tabs.map(async (tab) => {
          const res = await fetchUsersConnection({
            kind: tab.kind,
            include_deleted: includeDeleted,
            first: 1,
          });
          return [tab.kind, res.total_count] as const;
        }),
      );
      setTabCounts(Object.fromEntries(results));
    } catch {
      // Abaikan apabila count sekunder tertunda; query utama tetap menangani error
    }
  }, [config.tabs, includeDeleted]);

  const loadUsers = useCallback(
    async (
      cursorVars: Partial<UsersFilterVars> = { first: 20 },
      overrideKind?: UserKind | null,
    ) => {
      setIsLoading(true);
      setErrorMsg(null);
      try {
        const targetKind =
          overrideKind !== undefined ? overrideKind : effectiveKind;
        const res = await fetchUsersConnection({
          query: appliedQuery || null,
          kind: targetKind,
          status: statusFilter || null,
          include_deleted: includeDeleted,
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
    [appliedQuery, effectiveKind, statusFilter, includeDeleted],
  );

  useEffect(() => {
    let active = true;
    async function init() {
      try {
        const [res] = await Promise.all([
          fetchUsersConnection({
            query: appliedQuery || null,
            kind: effectiveKind,
            status: statusFilter || null,
            include_deleted: includeDeleted,
            first: 20,
          }),
          refreshSegmentCounts(),
        ]);
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
  }, [
    appliedQuery,
    effectiveKind,
    statusFilter,
    includeDeleted,
    refreshSegmentCounts,
  ]);

  function openCreateModalFor(targetKind: ProvisionUserKind) {
    setCreateKind(targetKind);
    setCreateModalNonce((n) => n + 1);
    setIsCreateOpen(true);
  }

  function handleSearchSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAppliedQuery(searchInput.trim());
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col justify-between gap-4 border-b border-[#dde3e8] pb-5 lg:flex-row lg:items-end">
        <div>
          <p
            className={`text-xs font-semibold tracking-wider text-[#133554] ${styles.monoText}`}
          >
            {config.kicker}
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight text-[#0b2439]">
            {config.title}
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-[#5a6672]">
            {config.subtitle}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          {(segment === "all" || segment === "internals") && (
            <button
              type="button"
              onClick={() => openCreateModalFor("EMPLOYEE")}
              className={styles.btnPrimary}
            >
              <IdentityIcon name="plus" size={16} />
              <span>Create Employee</span>
            </button>
          )}

          {(segment === "all" || segment === "vendors") && (
            <button
              type="button"
              onClick={() => openCreateModalFor("VENDOR")}
              className={
                segment === "vendors" ? styles.btnPrimary : styles.btnSecondary
              }
            >
              <IdentityIcon name="plus" size={16} />
              <span>Create Vendor</span>
            </button>
          )}

          {segment === "vendors" && (
            <button
              type="button"
              onClick={() => openCreateModalFor("SUPPLIER")}
              className={styles.btnSecondary}
            >
              <IdentityIcon name="plus" size={16} />
              <span>Create Supplier</span>
            </button>
          )}

          {segment === "customers" && (
            <button
              type="button"
              onClick={() => openCreateModalFor("CUSTOMER")}
              className={styles.btnPrimary}
            >
              <IdentityIcon name="plus" size={16} />
              <span>Create Customer</span>
            </button>
          )}
        </div>
      </div>

      {/* Sub-tab kategori kind khusus segmen multi-kind (Vendors & Internals) */}
      {config.tabs ? (
        <div
          className="flex flex-wrap items-center gap-2"
          role="tablist"
          aria-label="Filter kategori kind pengguna"
        >
          {config.tabs.map((tab) => {
            const active = kindFilter === tab.kind;
            const count = active
              ? conn?.total_count
              : tabCounts[tab.kind];
            return (
              <button
                key={tab.kind}
                type="button"
                role="tab"
                aria-selected={active}
                disabled={config.tabs!.length <= 1}
                onClick={() => setKindFilter(tab.kind)}
                className={`inline-flex items-center gap-2 rounded-lg border px-3.5 py-2 text-xs font-semibold transition-colors ${
                  active
                    ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                    : "border-[#dde3e8] bg-white text-[#5a6672] hover:bg-[#f4f7f8] hover:text-[#0b2439]"
                }`}
              >
                <span>{tab.label}</span>
                <span
                  className={`rounded px-1.5 py-0.5 text-[10px] ${styles.monoText} ${
                    active
                      ? "bg-white/15 text-[#f9fafb]"
                      : "border border-[#dde3e8] bg-[#f4f7f8] text-[#133554]"
                  }`}
                >
                  {tab.kind}
                  {typeof count === "number" ? ` · ${count}` : ""}
                </span>
              </button>
            );
          })}
        </div>
      ) : null}

      {/* Bar Filter & Pencarian */}
      <form
        onSubmit={handleSearchSubmit}
        className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-white p-4 md:grid-cols-12"
      >
        <div className={segment === "all" ? "md:col-span-5" : "md:col-span-6"}>
          <label
            htmlFor="users-search"
            className="mb-1 block text-xs font-semibold text-[#0b2439]"
          >
            Cari Nama atau Identitas
          </label>
          <div className="flex gap-2">
            <input
              id="users-search"
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Nama, email, atau username..."
              className={styles.fieldInput}
            />
            <button type="submit" className={`${styles.btnSecondary} shrink-0`}>
              Cari
            </button>
          </div>
        </div>

        {segment === "all" ? (
          <div className="md:col-span-3">
            <label
              htmlFor="users-kind-filter"
              className="mb-1 block text-xs font-semibold text-[#0b2439]"
            >
              Kategori (`UserKind`)
            </label>
            <select
              id="users-kind-filter"
              value={kindFilter}
              onChange={(e) => setKindFilter(e.target.value as UserKind | "")}
              className={styles.fieldSelect}
            >
              <option value="">Semua Kategori</option>
              <option value="EMPLOYEE">EMPLOYEE</option>
              <option value="VENDOR">VENDOR</option>
              <option value="SUPPLIER">SUPPLIER</option>
              <option value="CUSTOMER">CUSTOMER</option>
              <option value="SYSTEM">SYSTEM</option>
            </select>
          </div>
        ) : null}

        <div className={segment === "all" ? "md:col-span-2" : "md:col-span-3"}>
          <label
            htmlFor="users-status-filter"
            className="mb-1 block text-xs font-semibold text-[#0b2439]"
          >
            Status
          </label>
          <select
            id="users-status-filter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as UserStatus | "")}
            className={styles.fieldSelect}
          >
            <option value="">Semua Status</option>
            <option value="ACTIVE">ACTIVE</option>
            <option value="PENDING">PENDING</option>
            <option value="SUSPENDED">SUSPENDED</option>
            <option value="DISABLED">DISABLED</option>
          </select>
        </div>

        <div
          className={
            segment === "all"
              ? "flex items-end md:col-span-2"
              : "flex items-end md:col-span-3"
          }
        >
          <label className="flex min-h-[42px] w-full cursor-pointer items-center gap-2 rounded-lg border border-[#dde3e8] bg-[#f9fafb] px-3 text-xs font-medium text-[#0b2439]">
            <input
              type="checkbox"
              checked={includeDeleted}
              onChange={(e) => setIncludeDeleted(e.target.checked)}
              className="h-4 w-4 accent-[#133554]"
            />
            <span>Sertakan Dihapus</span>
          </label>
        </div>
      </form>

      {errorMsg ? (
        <ErrorPanel
          message={errorMsg}
          onRetry={() => void loadUsers({ first: 20 })}
        />
      ) : (
        <div className={`${styles.hairlinePanel} overflow-hidden`}>
          {isLoading && !conn ? (
            <TableSkeleton rows={6} cols={5} />
          ) : !conn || conn.nodes.length === 0 ? (
            <EmptyPanel
              title="Tidak Ada Pengguna Ditemukan"
              description="Sesuaikan kata kunci pencarian atau filter kategori/status untuk menampilkan daftar akun."
              action={
                <button
                  type="button"
                  onClick={() => {
                    setSearchInput("");
                    setAppliedQuery("");
                    setKindFilter(config.defaultKindFilter);
                    setStatusFilter("");
                    setIncludeDeleted(false);
                  }}
                  className={styles.btnSecondary}
                >
                  Reset Filter
                </button>
              }
            />
          ) : (
            <>
              <div className={styles.tableWrap}>
                <table className={styles.dataTable}>
                  <thead>
                    <tr>
                      <th>Pengguna</th>
                      <th>Kategori</th>
                      <th>Status</th>
                      <th>Login Terakhir</th>
                      <th className="text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {conn.nodes.map((u) => (
                      <tr key={u.id}>
                        <td>
                          <button
                            type="button"
                            onClick={() => {
                              setActiveUserTab("overview");
                              setActiveUserId(u._id);
                            }}
                            className="text-left group"
                          >
                            <p className="font-semibold text-[#0b2439] group-hover:text-[#133554] group-hover:underline">
                              {u.display_name}
                            </p>
                            <p
                              className={`text-xs text-[#5a6672] ${styles.monoText}`}
                            >
                              {u._id}
                            </p>
                          </button>
                        </td>
                        <td>
                          <UserKindBadge kind={u.kind} />
                        </td>
                        <td>
                          <UserStatusBadge status={u.status} />
                        </td>
                        <td className="text-xs text-[#5a6672]">
                          {formatDateTimeID(u.last_login_at)}
                        </td>
                        <td className="text-right">
                          <div className="inline-flex items-center justify-end gap-1.5">
                            <button
                              type="button"
                              onClick={() => {
                                setActiveUserTab("roles");
                                setActiveUserId(u._id);
                              }}
                              className={`${styles.btnSecondary} ${styles.btnSmall}`}
                            >
                              <IdentityIcon name="roles" size={13} />
                              <span>Peran</span>
                            </button>
                            <button
                              type="button"
                              onClick={() => {
                                setActiveUserTab("overview");
                                setActiveUserId(u._id);
                              }}
                              className={`${styles.btnSecondary} ${styles.btnSmall}`}
                            >
                              Kelola
                            </button>
                          </div>
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
                  void loadUsers({
                    last: 20,
                    before: conn.pageInfo.startCursor || undefined,
                  })
                }
                onNext={() =>
                  void loadUsers({
                    first: 20,
                    after: conn.pageInfo.endCursor || undefined,
                  })
                }
              />
            </>
          )}
        </div>
      )}

      {isCreateOpen ? (
        <UserCreateModal
          key={`${createKind}-${createModalNonce}`}
          open
          initialKind={createKind}
          allowedKinds={config.allowedCreateKinds}
          onClose={() => setIsCreateOpen(false)}
          onCreated={(created) => {
            const nextKind =
              created &&
              segment !== "all" &&
              config.tabs?.some((t) => t.kind === created.kind)
                ? created.kind
                : effectiveKind;
            if (created && nextKind && nextKind !== kindFilter) {
              setKindFilter(nextKind);
            }
            void loadUsers({ first: 20 }, nextKind);
            void refreshSegmentCounts();
            if (created?._id) {
              setActiveUserTab("roles");
              setActiveUserId(created._id);
            }
          }}
        />
      ) : null}

      <UserDetailDrawer
        userId={activeUserId}
        initialTab={activeUserTab}
        onClose={() => setActiveUserId(null)}
        onUpdated={() => {
          void loadUsers({ first: 20 });
          void refreshSegmentCounts();
        }}
      />
    </div>
  );
}
