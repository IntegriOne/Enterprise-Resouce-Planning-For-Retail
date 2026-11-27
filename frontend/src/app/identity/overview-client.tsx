"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  ErrorPanel,
  ScopeBadge,
  SensitiveBadge,
  UserKindBadge,
  UserStatusBadge,
  formatDateTimeID,
  formatUnixTimeID,
} from "@/components/identity/identity-primitives";
import {
  fetchMySessions,
  fetchOverviewCounts,
  formatIamDashboardError,
  isUnauthenticatedError,
  type OverviewCounts,
} from "@/lib/graphql/iam-dashboard-queries";
import type { IamSession } from "@/lib/graphql/iam-types";

export default function IdentityOverviewClient() {
  const {
    me,
    snapshot,
    isLoadingContext,
    contextError,
    refreshOperatorContext,
    openStepUpModal,
  } = useIdentityWorkspace();

  const [counts, setCounts] = useState<OverviewCounts | null>(null);
  const [recentSessions, setRecentSessions] = useState<IamSession[]>([]);
  const [isLoadingStats, setIsLoadingStats] = useState(true);
  const [statsError, setStatsError] = useState<string | null>(null);

  const loadStats = useCallback(async () => {
    setIsLoadingStats(true);
    setStatsError(null);
    try {
      const [overviewCounts, sessionsConn] = await Promise.all([
        fetchOverviewCounts(),
        fetchMySessions({ first: 5, active_only: true }),
      ]);
      setCounts(overviewCounts);
      setRecentSessions(sessionsConn.nodes);
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      setStatsError(formatIamDashboardError(err));
    } finally {
      setIsLoadingStats(false);
    }
  }, []);

  useEffect(() => {
    let active = true;
    async function initStats() {
      try {
        const [overviewCounts, sessionsConn] = await Promise.all([
          fetchOverviewCounts(),
          fetchMySessions({ first: 5, active_only: true }),
        ]);
        if (!active) return;
        setCounts(overviewCounts);
        setRecentSessions(sessionsConn.nodes);
        setStatsError(null);
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setStatsError(formatIamDashboardError(err));
      } finally {
        if (active) setIsLoadingStats(false);
      }
    }
    void initStats();
    return () => {
      active = false;
    };
  }, []);

  if (contextError) {
    return (
      <ErrorPanel
        title="Gagal Memuat Konteks Operator"
        message={contextError}
        onRetry={() => void refreshOperatorContext()}
      />
    );
  }

  return (
    <div className="space-y-6">
      {/* Header Asimetris Dashboard */}
      <div className="flex flex-col justify-between gap-4 border-b border-[#dde3e8] pb-5 md:flex-row md:items-end">
        <div>
          <p className={`text-xs font-semibold tracking-wider text-[#133554] ${styles.monoText}`}>
            IDENTITY &amp; ACCESS MANAGEMENT
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight text-[#0b2439] sm:text-3xl">
            Tata Kelola Identitas &amp; Otorisasi
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-[#5a6672]">
            Pusat pengendalian akun terpadu, peran RBAC berscope, grup, dan
            pemantauan sesi opaque aktif.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          <button
            type="button"
            onClick={() => {
              void refreshOperatorContext();
              void loadStats();
            }}
            className={styles.btnSecondary}
          >
            <IdentityIcon name="refresh" size={15} />
            <span>Segarkan Data</span>
          </button>
          <Link
            href="/identity/users/vendors?action=create-vendor"
            className={styles.btnSecondary}
          >
            <IdentityIcon name="plus" size={15} />
            <span>Create Vendor</span>
          </Link>
          <Link
            href="/identity/users/internals?action=create-employee"
            className={styles.btnPrimary}
          >
            <IdentityIcon name="plus" size={15} />
            <span>Create Employee</span>
          </Link>
        </div>
      </div>

      {/* Strip Metrik Hairline Tanpa Dinding Kartu Tebal */}
      <section
        aria-label="Ringkasan jumlah entitas IAM"
        className="grid grid-cols-2 divide-x divide-y divide-[#dde3e8] overflow-hidden rounded-xl border border-[#dde3e8] bg-white lg:grid-cols-4 lg:divide-y-0"
      >
        {[
          {
            label: "Total Akun Pengguna",
            value: counts?.usersTotal,
            href: "/identity/users",
            sub: "Direktori lintas kategori",
          },
          {
            label: "Peran RBAC Terdaftar",
            value: counts?.rolesTotal,
            href: "/identity/roles",
            sub: "Termasuk peran sistem",
          },
          {
            label: "Grup Identitas",
            value: counts?.groupsTotal,
            href: "/identity/groups",
            sub: "Pewarisan peran berscope",
          },
          {
            label: "Katalog Izin Atomik",
            value: counts?.permissionsTotal,
            href: "/identity/permissions",
            sub: "Guard RBAC–ABAC Gateway",
          },
        ].map((metric) => (
          <Link
            key={metric.label}
            href={metric.href}
            className="group flex flex-col justify-between gap-2 p-5 transition-colors hover:bg-[#f4f7f8]"
          >
            <span className="text-xs font-medium text-[#5a6672]">
              {metric.label}
            </span>
            <div className="flex items-baseline justify-between gap-2">
              {isLoadingStats ? (
                <div className={`${styles.skeletonShimmer} h-8 w-16`} />
              ) : (
                <span
                  className={`text-2xl font-bold tracking-tight text-[#0b2439] sm:text-3xl ${styles.monoText}`}
                >
                  {metric.value !== null && metric.value !== undefined
                    ? metric.value
                    : "N/A"}
                </span>
              )}
              <span className="text-xs font-semibold text-[#133554] opacity-0 transition-opacity group-hover:opacity-100">
                Buka →
              </span>
            </div>
            <span className="text-[11px] text-[#788491]">{metric.sub}</span>
          </Link>
        ))}
      </section>

      {statsError ? (
        <ErrorPanel
          title="Sebagian Metrik IAM Tertunda"
          message={statsError}
          onRetry={() => void loadStats()}
        />
      ) : null}

      {/* Grid Asimetris 1.35fr : 1fr */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Kolom Kiri (7/12): Snapshot Otorisasi Operator */}
        <section className="space-y-5 lg:col-span-7">
          <div className={styles.hairlinePanel}>
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#dde3e8] px-5 py-4">
              <div>
                <h2 className="text-sm font-semibold text-[#0b2439]">
                  AccessSnapshot Otorisasi Aktif
                </h2>
                <p className="text-xs text-[#5a6672]">
                  Proyeksi keamanan server (TTL segar ≤ 30 detik, fail-closed)
                </p>
              </div>
              {snapshot ? (
                <div className="flex items-center gap-2">
                  <UserKindBadge kind={snapshot.user_kind} />
                  <UserStatusBadge status={snapshot.status} />
                </div>
              ) : null}
            </div>

            {isLoadingContext || !snapshot || !me ? (
              <div className="space-y-3 p-5">
                <div className={`${styles.skeletonShimmer} h-5 w-48`} />
                <div className={`${styles.skeletonShimmer} h-24 w-full`} />
              </div>
            ) : (
              <div className="divide-y divide-[#dde3e8]">
                <dl
                  className={`grid grid-cols-2 gap-4 px-5 py-4 text-xs sm:grid-cols-4 ${styles.monoText}`}
                >
                  <div>
                    <dt className="text-[#788491]">auth_epoch</dt>
                    <dd className="mt-1 text-sm font-semibold text-[#0b2439]">
                      #{snapshot.auth_epoch}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-[#788491]">perm_revision</dt>
                    <dd className="mt-1 text-sm font-semibold text-[#0b2439]">
                      #{snapshot.perm_revision}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-[#788491]">recovery_epoch</dt>
                    <dd className="mt-1 text-sm font-semibold text-[#0b2439]">
                      #{snapshot.recovery_epoch}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-[#788491]">fresh_until</dt>
                    <dd className="mt-1 text-sm font-semibold text-[#067647]">
                      {formatUnixTimeID(snapshot.fresh_until_unix)}
                    </dd>
                  </div>
                </dl>

                <div className="px-5 py-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="text-xs font-semibold text-[#0b2439]">
                      Peran Aktif Operator ({snapshot.roles.length})
                    </span>
                    <span className={`text-[11px] text-[#5a6672] ${styles.monoText}`}>
                      Default redirect: {snapshot.default_redirect_path}
                    </span>
                  </div>
                  <div className="mt-2.5 flex flex-wrap gap-2">
                    {snapshot.roles.map((role) => (
                      <span
                        key={role.id}
                        className="inline-flex items-center gap-2 rounded-lg border border-[#dde3e8] bg-[#f4f7f8] px-2.5 py-1 text-xs"
                      >
                        <strong className="text-[#0b2439]">{role.name}</strong>
                        <span className={`text-[#5a6672] ${styles.monoText}`}>
                          {role.code} · P{role.priority}
                        </span>
                      </span>
                    ))}
                  </div>
                </div>

                <div className="px-5 py-4">
                  <p className="mb-2.5 text-xs font-semibold text-[#0b2439]">
                    Rincian Izin Terscope ({snapshot.permissions.length})
                  </p>
                  <div className="max-h-64 divide-y divide-[#dde3e8] overflow-y-auto rounded-lg border border-[#dde3e8]">
                    {snapshot.permissions.map((perm, idx) => (
                      <div
                        key={`${perm.permission_code}-${perm.scope.type}-${idx}`}
                        className="flex flex-wrap items-center justify-between gap-2 px-3 py-2 text-xs"
                      >
                        <span className={`font-medium text-[#0b2439] ${styles.monoText}`}>
                          {perm.permission_code}
                        </span>
                        <div className="flex items-center gap-2">
                          <ScopeBadge scope={perm.scope} />
                          <SensitiveBadge sensitive={perm.is_sensitive} />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        </section>

        {/* Kolom Kanan (5/12): Sesi Saya & Pintasan Operasional */}
        <section className="space-y-5 lg:col-span-5">
          <div className={styles.hairlinePanel}>
            <div className="flex items-center justify-between gap-3 border-b border-[#dde3e8] px-5 py-4">
              <div>
                <h2 className="text-sm font-semibold text-[#0b2439]">
                  Aksi Cepat Tata Kelola
                </h2>
                <p className="text-xs text-[#5a6672]">
                  Operasi harian administrator IAM
                </p>
              </div>
              <button
                type="button"
                onClick={() =>
                  openStepUpModal(
                    "Aktifkan verifikasi TOTP pada sesi ini sebelum menjalankan aksi sensitif.",
                  )
                }
                className={`${styles.btnSecondary} ${styles.btnSmall}`}
              >
                <IdentityIcon name="shield" size={14} />
                <span>Step-Up MFA</span>
              </button>
            </div>

            <div className="divide-y divide-[#dde3e8]">
              {[
                {
                  href: "/identity/users/customers",
                  title: "Direktori Customers (CUSTOMER)",
                  desc: "Tinjau akun pelanggan Web Commerce, status verifikasi, dan peran self-service.",
                },
                {
                  href: "/identity/users/vendors",
                  title: "Direktori Vendors & Mitra (VENDOR / SUPPLIER)",
                  desc: "Provision akun vendor (createVendorUser) atau supplier, serta atur scope party.",
                },
                {
                  href: "/identity/users/internals",
                  title: "Direktori Internals (EMPLOYEE / SYSTEM)",
                  desc: "Provision karyawan (createEmployeeUser), akun sistem, suspend, dan penetapan peran.",
                },
                {
                  href: "/identity/roles",
                  title: "Desain Peran RBAC & Redirect",
                  desc: "Kelola izin per peran, priority redirect, dan domain manajemen.",
                },
                {
                  href: "/identity/groups",
                  title: "Grup & Pewarisan Peran Berscope",
                  desc: "Kelompokkan operator serta tetapkan peran berscope perusahaan/fasilitas.",
                },
                {
                  href: "/identity/sessions",
                  title: "Keamanan Akun, TOTP & Pencabutan Sesi",
                  desc: "Kelola MFA TOTP pribadi serta cabut sesi mencurigakan secara instan.",
                },
              ].map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="flex items-center justify-between gap-3 px-5 py-3.5 transition-colors hover:bg-[#f4f7f8]"
                >
                  <div className="min-w-0">
                    <p className="text-sm font-semibold text-[#0b2439]">
                      {link.title}
                    </p>
                    <p className="mt-0.5 text-xs text-[#5a6672]">{link.desc}</p>
                  </div>
                  <IdentityIcon
                    name="chevronRight"
                    size={16}
                    className="shrink-0 text-[#5a6672]"
                  />
                </Link>
              ))}
            </div>
          </div>

          <div className={styles.hairlinePanel}>
            <div className="flex items-center justify-between gap-3 border-b border-[#dde3e8] px-5 py-4">
              <div>
                <h2 className="text-sm font-semibold text-[#0b2439]">
                  Sesi Aktif Operator
                </h2>
                <p className="text-xs text-[#5a6672]">
                  {recentSessions.length} sesi aktif terkini
                </p>
              </div>
              <Link
                href="/identity/sessions"
                className="text-xs font-semibold text-[#133554] hover:underline"
              >
                Kelola Semua →
              </Link>
            </div>

            {isLoadingStats ? (
              <div className="space-y-2 p-5">
                <div className={`${styles.skeletonShimmer} h-10 w-full`} />
                <div className={`${styles.skeletonShimmer} h-10 w-full`} />
              </div>
            ) : recentSessions.length === 0 ? (
              <p className="px-5 py-6 text-xs text-[#5a6672]">
                Tidak ada sesi aktif tambahan tercatat.
              </p>
            ) : (
              <div className="divide-y divide-[#dde3e8]">
                {recentSessions.map((sess) => (
                  <div
                    key={sess.id}
                    className="flex flex-wrap items-center justify-between gap-2 px-5 py-3 text-xs"
                  >
                    <div className="min-w-0">
                      <p className={`truncate font-semibold text-[#0b2439] ${styles.monoText}`}>
                        {sess.ip_address || "IP internal"} · gen {sess.generation}
                      </p>
                      <p className="mt-0.5 truncate text-[#5a6672]">
                        Batas idle: {formatDateTimeID(sess.idle_expires_at)}
                      </p>
                    </div>
                    <span
                      className={`rounded-md border px-2 py-0.5 text-[11px] font-semibold ${styles.monoText} ${
                        sess.mfa_verified
                          ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                          : "border-[#dde3e8] bg-[#f4f7f8] text-[#5a6672]"
                      }`}
                    >
                      {sess.mfa_verified ? "MFA OK" : "STD"}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  );
}
