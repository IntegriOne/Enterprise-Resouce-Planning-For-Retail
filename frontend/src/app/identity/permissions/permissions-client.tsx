"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  EmptyPanel,
  ErrorPanel,
  RelayPaginationBar,
  SensitiveBadge,
  TableSkeleton,
} from "@/components/identity/identity-primitives";
import {
  fetchPermissionsConnection,
  formatIamDashboardError,
  isUnauthenticatedError,
  type PermissionsFilterVars,
} from "@/lib/graphql/iam-dashboard-queries";
import type { PermissionConnection } from "@/lib/graphql/iam-types";

export default function PermissionsClient() {
  const { pushToast } = useIdentityWorkspace();

  const [searchInput, setSearchInput] = useState("");
  const [appliedQuery, setAppliedQuery] = useState("");
  const [resourceFilter, setResourceFilter] = useState("");
  const [sensitiveFilter, setSensitiveFilter] = useState<"ALL" | "SENSITIVE" | "STANDARD">("ALL");

  const [conn, setConn] = useState<PermissionConnection | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const loadPermissions = useCallback(
    async (cursorVars: Partial<PermissionsFilterVars> = { first: 25 }) => {
      setIsLoading(true);
      setErrorMsg(null);
      try {
        const res = await fetchPermissionsConnection({
          query: appliedQuery || null,
          resource: resourceFilter.trim() || null,
          is_sensitive:
            sensitiveFilter === "SENSITIVE"
              ? true
              : sensitiveFilter === "STANDARD"
                ? false
                : null,
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
    [appliedQuery, resourceFilter, sensitiveFilter],
  );

  useEffect(() => {
    let active = true;
    async function init() {
      try {
        const res = await fetchPermissionsConnection({
          query: appliedQuery || null,
          resource: resourceFilter.trim() || null,
          is_sensitive:
            sensitiveFilter === "SENSITIVE"
              ? true
              : sensitiveFilter === "STANDARD"
                ? false
                : null,
          first: 25,
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
  }, [appliedQuery, resourceFilter, sensitiveFilter]);

  function handleSearchSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setAppliedQuery(searchInput.trim());
  }

  async function handleCopyCode(code: string) {
    try {
      await navigator.clipboard.writeText(code);
      pushToast("info", "Kode Disalin", code);
    } catch {
      pushToast("error", "Gagal Menyalin", "Browser memblokir akses clipboard.");
    }
  }

  return (
    <div className="space-y-5">
      <div className="border-b border-[#dde3e8] pb-5">
        <h1 className="text-2xl font-bold tracking-tight text-[#0b2439]">
          Katalog Izin Atomik (`permissions`)
        </h1>
        <p className="mt-1 text-sm text-[#5a6672]">
          Daftar izin otoritatif dari server IAM beserta penanda operasi
          sensitif yang mewajibkan re-autentikasi MFA.
        </p>
      </div>

      <form
        onSubmit={handleSearchSubmit}
        className="grid grid-cols-1 gap-3 rounded-xl border border-[#dde3e8] bg-white p-4 md:grid-cols-12"
      >
        <div className="md:col-span-5">
          <label htmlFor="perm-search" className="mb-1 block text-xs font-semibold text-[#0b2439]">
            Cari Kode atau Deskripsi
          </label>
          <div className="flex gap-2">
            <input
              id="perm-search"
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="Contoh: iam.roles.write, users.read..."
              className={styles.fieldInput}
            />
            <button type="submit" className={`${styles.btnSecondary} shrink-0`}>
              Cari
            </button>
          </div>
        </div>

        <div className="md:col-span-4">
          <label htmlFor="perm-resource" className="mb-1 block text-xs font-semibold text-[#0b2439]">
            Filter Resource
          </label>
          <input
            id="perm-resource"
            type="text"
            value={resourceFilter}
            onChange={(e) => setResourceFilter(e.target.value)}
            placeholder="users / roles / groups / sessions"
            className={`${styles.fieldInput} ${styles.monoText}`}
          />
        </div>

        <div className="md:col-span-3">
          <label htmlFor="perm-sensitive" className="mb-1 block text-xs font-semibold text-[#0b2439]">
            Tingkat Sensitivitas
          </label>
          <select
            id="perm-sensitive"
            value={sensitiveFilter}
            onChange={(e) =>
              setSensitiveFilter(e.target.value as "ALL" | "SENSITIVE" | "STANDARD")
            }
            className={styles.fieldSelect}
          >
            <option value="ALL">Semua Izin</option>
            <option value="SENSITIVE">Sensitif (Wajib Step-Up)</option>
            <option value="STANDARD">Standar</option>
          </select>
        </div>
      </form>

      {errorMsg ? (
        <ErrorPanel
          message={errorMsg}
          onRetry={() => void loadPermissions({ first: 25 })}
        />
      ) : (
        <div className={`${styles.hairlinePanel} overflow-hidden`}>
          {isLoading && !conn ? (
            <TableSkeleton rows={8} cols={4} />
          ) : !conn || conn.nodes.length === 0 ? (
            <EmptyPanel
              title="Tidak Ada Izin Ditemukan"
              description="Ubah kata kunci pencarian atau filter resource untuk menelusuri katalog izin."
            />
          ) : (
            <>
              <div className={styles.tableWrap}>
                <table className={styles.dataTable}>
                  <thead>
                    <tr>
                      <th>Kode Izin</th>
                      <th>Resource / Aksi</th>
                      <th>Deskripsi</th>
                      <th>Klasifikasi</th>
                      <th className="text-right">Salin</th>
                    </tr>
                  </thead>
                  <tbody>
                    {conn.nodes.map((p) => (
                      <tr key={p.id}>
                        <td className={`font-semibold text-[#0b2439] ${styles.monoText}`}>
                          {p.code}
                        </td>
                        <td className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                          {p.resource} · {p.action}
                        </td>
                        <td className="max-w-md text-xs text-[#5a6672]">
                          {p.description || "-"}
                        </td>
                        <td>
                          <SensitiveBadge sensitive={p.is_sensitive} />
                        </td>
                        <td className="text-right">
                          <button
                            type="button"
                            onClick={() => void handleCopyCode(p.code)}
                            aria-label={`Salin kode izin ${p.code}`}
                            className={`${styles.btnSecondary} ${styles.btnSmall}`}
                          >
                            <IdentityIcon name="copy" size={13} />
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
                  void loadPermissions({
                    last: 25,
                    before: conn.pageInfo.startCursor || undefined,
                  })
                }
                onNext={() =>
                  void loadPermissions({
                    first: 25,
                    after: conn.pageInfo.endCursor || undefined,
                  })
                }
              />
            </>
          )}
        </div>
      )}
    </div>
  );
}
