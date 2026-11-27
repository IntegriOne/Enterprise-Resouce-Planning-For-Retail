"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import gsap from "gsap";
import { useGSAP } from "@gsap/react";
import styles from "@/components/identity/identity.module.css";
import {
  IdentityProvider,
  useIdentityWorkspace,
  type WorkspaceDomain,
} from "@/components/identity/identity-context";
import {
  IdentityBrandMark,
  IdentityIcon,
  type IdentityIconName,
} from "@/components/identity/identity-icons";
import { StepUpMfaModal } from "@/components/identity/stepup-mfa-modal";
import { formatUnixTimeID } from "@/components/identity/identity-primitives";
import { WorkspaceSwitcher } from "./workspace-switcher";

gsap.registerPlugin(useGSAP);

export type WorkspaceNavItem = {
  href: string;
  label: string;
  description: string;
  icon: IdentityIconName;
  group: string;
  requiredGlobalPermission?: string;
};

export const HR_EF_NAV_ITEMS: WorkspaceNavItem[] = [
  {
    href: "/human-resource",
    label: "Ringkasan Struktur",
    description: "Badan hukum, unit organisasi, lokasi, dan alamat perusahaan",
    icon: "overview",
    group: "Enterprise Foundation",
  },
  {
    href: "/human-resource/workers",
    label: "Pekerja",
    description: "Daftar pekerja dan pengajuan akun kerja",
    icon: "groups",
    group: "SDM",
  },
  {
    href: "/human-resource/requisitions",
    label: "Requisition",
    description: "Daftar kebutuhan tenaga kerja dan status pengajuan",
    icon: "groups",
    group: "SDM",
    requiredGlobalPermission: "human_resource.requisitions.read",
  },
  {
    href: "/human-resource/candidates",
    label: "Kandidat",
    description: "Daftar dan pencarian kandidat rekrutmen",
    icon: "users",
    group: "SDM",
    requiredGlobalPermission: "human_resource.candidates.read",
  },
  {
    href: "/human-resource/legal-entities",
    label: "Badan Hukum",
    description: "PT/CV induk, cabang entitas hukum, NPWP, NIB & status aktif",
    icon: "building",
    group: "Struktur Organisasi",
  },
  {
    href: "/human-resource/organization-units",
    label: "Unit Organisasi",
    description: "Divisi, departemen, tim, pusat biaya & kepala unit",
    icon: "groups",
    group: "Struktur Organisasi",
  },
  {
    href: "/human-resource/locations",
    label: "Lokasi & Fasilitas",
    description: "Gudang, toko ritel, HQ, cabang & validasi scope fasilitas",
    icon: "shield",
    group: "Lokasi & Alamat",
  },
  {
    href: "/human-resource/addresses",
    label: "Master Alamat",
    description: "Geografi & alamat terpusat lintas entitas & lokasi",
    icon: "key",
    group: "Lokasi & Alamat",
  },
];

export const PROCUREMENT_EF_NAV_ITEMS: WorkspaceNavItem[] = [
  {
    href: "/procurement",
    label: "Ringkasan Mitra",
    description: "Ringkasan rekanan, evaluator akses portal & validator scope",
    icon: "overview",
    group: "Partner Governance",
  },
  {
    href: "/procurement/partners",
    label: "Direktori Mitra",
    description: "Vendor CMT, supplier, B2B, profil pajak PKP & status aktivasi",
    icon: "building",
    group: "Master Mitra & Peran",
  },
  {
    href: "/procurement/roles",
    label: "Peran Bisnis",
    description: "Penetapan SUPPLIER, VENDOR, CARRIER & validasi scope party",
    icon: "roles",
    group: "Master Mitra & Peran",
  },
  {
    href: "/procurement/contacts",
    label: "PIC & Portal IAM",
    description: "Kontak perwakilan mitra, tautan akun IAM & aktivasi portal",
    icon: "users",
    group: "PIC & Akses Portal",
  },
  {
    href: "/procurement/addresses",
    label: "Alamat Mitra",
    description: "Relasi alamat penagihan, pengiriman, HQ & kantor terdaftar",
    icon: "badge",
    group: "PIC & Akses Portal",
  },
];

function getWorkspaceConfig(domain: Exclude<WorkspaceDomain, "identity">) {
  if (domain === "human-resource") {
    return {
      title: "HR & Struktur EF",
      sub: "Enterprise Structure",
      items: HR_EF_NAV_ITEMS,
    };
  }
  return {
    title: "Mitra & Procurement",
    sub: "Vendor & Supplier EF",
    items: PROCUREMENT_EF_NAV_ITEMS,
  };
}

function WorkspaceToastViewport() {
  const { toasts, dismissToast } = useIdentityWorkspace();
  if (toasts.length === 0) return null;
  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-full max-w-sm flex-col gap-2 px-4 sm:px-0"
    >
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`pointer-events-auto flex items-start justify-between gap-3 rounded-xl border p-3.5 shadow-lg ${
            t.kind === "error"
              ? "border-[#fecdca] bg-[#fef3f2] text-[#b42318]"
              : t.kind === "success"
                ? "border-[#abefc6] bg-[#ecfdf3] text-[#067647]"
                : "border-[#dde3e8] bg-white text-[#0b2439]"
          }`}
        >
          <div className="min-w-0">
            <p className="text-xs font-semibold">{t.title}</p>
            {t.detail ? (
              <p className="mt-0.5 text-xs opacity-90">{t.detail}</p>
            ) : null}
          </div>
          <button
            type="button"
            onClick={() => dismissToast(t.id)}
            aria-label="Tutup notifikasi"
            className="shrink-0 opacity-70 hover:opacity-100"
          >
            <IdentityIcon name="close" size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}

function WorkspaceCommandModal({ items }: { items: WorkspaceNavItem[] }) {
  const { isCommandOpen, closeCommandPalette } = useIdentityWorkspace();
  const router = useRouter();
  const [query, setQuery] = useState("");

  useEffect(() => {
    if (!isCommandOpen) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") closeCommandPalette();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [isCommandOpen, closeCommandPalette]);

  if (!isCommandOpen) return null;

  const filtered = items.filter(
    (item) =>
      item.label.toLowerCase().includes(query.trim().toLowerCase()) ||
      item.href.toLowerCase().includes(query.trim().toLowerCase()),
  );

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-[#0b2439]/40 p-4 pt-20 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label="Navigasi Cepat Enterprise Foundation"
      onClick={(e) => {
        if (e.target === e.currentTarget) closeCommandPalette();
      }}
    >
      <div className="w-full max-w-lg overflow-hidden rounded-xl border border-[#dde3e8] bg-white shadow-2xl">
        <div className="flex items-center gap-2.5 border-b border-[#dde3e8] px-4 py-3">
          <IdentityIcon name="search" size={17} className="text-[#5a6672]" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cari modul Struktur / Mitra..."
            aria-label="Cari modul"
            autoFocus
            className="w-full bg-transparent text-sm text-[#0b2439] placeholder-[#788491] focus:outline-none"
          />
          <button
            type="button"
            onClick={closeCommandPalette}
            className="rounded px-1.5 py-0.5 text-xs text-[#5a6672] hover:bg-[#f4f7f8]"
          >
            Esc
          </button>
        </div>
        <div className="max-h-72 divide-y divide-[#dde3e8] overflow-y-auto p-1.5">
          {filtered.map((item) => (
            <button
              key={item.href}
              type="button"
              onClick={() => {
                closeCommandPalette();
                router.push(item.href);
              }}
              className="flex w-full items-center justify-between gap-3 rounded-lg px-3 py-2.5 text-left text-sm text-[#17202a] hover:bg-[#e9eef2]"
            >
              <span className="flex items-center gap-2.5 font-medium">
                <IdentityIcon name={item.icon} size={16} />
                <span>{item.label}</span>
              </span>
              <span className={`text-xs text-[#5a6672] ${styles.monoText}`}>
                {item.href}
              </span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

function WorkspaceShellFrame({
  domain,
  children,
}: {
  domain: Exclude<WorkspaceDomain, "identity">;
  children: ReactNode;
}) {
  const pathname = usePathname();
  const contentRef = useRef<HTMLElement>(null);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const {
    me,
    snapshot,
    accessDenied,
    isLoadingContext,
    openCommandPalette,
    openStepUpModal,
    logout,
    isLoggingOut,
  } = useIdentityWorkspace();

  const cfg = getWorkspaceConfig(domain);

  useGSAP(
    () => {
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.fromTo(
          "[data-ef-page-body]",
          { opacity: 0.72, y: 6 },
          {
            opacity: 1,
            y: 0,
            duration: 0.2,
            ease: "power2.out",
            clearProps: "transform,opacity",
          },
        );
      });
      return () => mm.revert();
    },
    { scope: contentRef, dependencies: [pathname] },
  );

  const visibleItems = cfg.items.filter((item) =>
    !item.requiredGlobalPermission || snapshot?.permissions.some((grant) =>
      grant.permission_code === item.requiredGlobalPermission && grant.scope.type === "GLOBAL",
    ),
  );
  const groups = Array.from(new Set(visibleItems.map((item) => item.group)));
  const activeItem =
    visibleItems.find((item) => item.href === pathname) || cfg.items[0];
  const primaryRoleCode = snapshot?.roles[0]?.code || "EF_ROLE";

  return (
    <div className={styles.workspaceRoot}>
      {mobileOpen ? (
        <div
          className={styles.sidebarBackdrop}
          onClick={() => setMobileOpen(false)}
          aria-hidden="true"
        />
      ) : null}

      <aside
        aria-label={`Navigasi ${cfg.title}`}
        className={`${styles.sidebar} ${
          sidebarCollapsed ? styles.sidebarCollapsed : ""
        } ${mobileOpen ? styles.sidebarOpen : ""}`}
      >
        <div className={styles.sidebarHeader}>
          <Link
            href={cfg.items[0]?.href || "/human-resource"}
            className={styles.brandLockup}
            onClick={() => setMobileOpen(false)}
          >
            <div className={styles.brandIconTile}>
              <IdentityBrandMark />
            </div>
            {!sidebarCollapsed ? (
              <div className={styles.brandMeta}>
                <span className={styles.brandTitle}>IntegriOne</span>
                <span className={styles.brandSub}>{cfg.sub}</span>
              </div>
            ) : null}
          </Link>

          <button
            type="button"
            onClick={() => setMobileOpen(false)}
            aria-label="Tutup navigasi"
            className={`${styles.iconBtn} lg:hidden`}
          >
            <IdentityIcon name="close" size={16} />
          </button>
        </div>

        <WorkspaceSwitcher compact={sidebarCollapsed} variant="sidebar" />

        <nav className={styles.sidebarNavScroll}>
          {groups.map((grp) => (
            <div key={grp}>
              {!sidebarCollapsed ? (
                <p className={styles.navGroupLabel}>{grp}</p>
              ) : null}
              <div className={styles.navList}>
                {visibleItems
                  .filter((item) => item.group === grp)
                  .map((item) => {
                    const active = pathname === item.href;
                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        onClick={() => setMobileOpen(false)}
                        title={sidebarCollapsed ? item.label : undefined}
                        aria-current={active ? "page" : undefined}
                        className={`${styles.navItem} ${
                          active ? styles.navItemActive : ""
                        }`}
                      >
                        <IdentityIcon name={item.icon} size={17} />
                        {!sidebarCollapsed ? (
                          <span className="truncate">{item.label}</span>
                        ) : null}
                      </Link>
                    );
                  })}
              </div>
            </div>
          ))}
        </nav>

        <div className={styles.sidebarFooter}>
          {!sidebarCollapsed ? (
            <div className={styles.snapshotCard}>
              <div className={styles.snapshotRow}>
                <span className="flex items-center gap-2 font-semibold text-[#0b2439]">
                  <span
                    className={`${styles.liveDot} ${
                      !snapshot ? styles.liveDotWarn : ""
                    }`}
                  />
                  Snapshot Aktif
                </span>
                <span className={`text-[11px] text-[#5a6672] ${styles.monoText}`}>
                  v{snapshot?.perm_revision ?? 0}
                </span>
              </div>
              <div className={`${styles.snapshotRow} text-[#5a6672]`}>
                <span>Peran Utama</span>
                <span className={`font-medium text-[#133554] ${styles.monoText}`}>
                  {primaryRoleCode}
                </span>
              </div>
              <div className={`${styles.snapshotRow} text-[#5a6672]`}>
                <span>Sinkronisasi</span>
                <span className={styles.monoText}>
                  {formatUnixTimeID(snapshot?.synced_at_unix)}
                </span>
              </div>
            </div>
          ) : null}
        </div>
      </aside>

      <div className={styles.mainColumn}>
        <header className={styles.topNavbar}>
          <div className={styles.navbarLeft}>
            <button
              type="button"
              onClick={() => setMobileOpen(true)}
              aria-label="Buka navigasi samping"
              className={`${styles.iconBtn} lg:hidden`}
            >
              <IdentityIcon name="menu" size={18} />
            </button>
            <button
              type="button"
              onClick={() => setSidebarCollapsed((p) => !p)}
              aria-label="Alihkan lebar sidebar"
              className={`${styles.iconBtn} hidden lg:inline-flex`}
            >
              <IdentityIcon name="panelCollapse" size={18} />
            </button>
            <div className="min-w-0">
              <p className={`text-[11px] font-semibold uppercase tracking-wider text-[#5a6672] ${styles.monoText}`}>
                {cfg.title}
              </p>
              <h1 className="truncate text-base font-bold text-[#0b2439] sm:text-lg">
                {activeItem?.label || cfg.title}
              </h1>
            </div>
          </div>

          <div className={styles.navbarRight}>
            <WorkspaceSwitcher variant="navbar" />
            <button
              type="button"
              onClick={openCommandPalette}
              className={styles.commandTriggerBtn}
              aria-label="Buka navigasi cepat (Ctrl+K)"
            >
              <IdentityIcon name="search" size={15} />
              <span className="hidden sm:inline">Cari</span>
              <span className={`${styles.kbdShortcut} hidden md:inline`}>
                ⌘K
              </span>
            </button>
            <button
              type="button"
              onClick={() =>
                openStepUpModal(
                  "Verifikasi kode TOTP aktif untuk menjalankan 12 mutasi sensitif Enterprise Foundation.",
                )
              }
              className={styles.stepUpPill}
            >
              <IdentityIcon name="shield" size={14} />
              <span className="hidden sm:inline">Step-Up</span>
            </button>
            {me ? (
              <span className="hidden max-w-[140px] truncate text-xs font-semibold text-[#0b2439] xl:inline">
                {me.display_name}
              </span>
            ) : null}
            <button
              type="button"
              onClick={() => void logout()}
              disabled={isLoggingOut}
              aria-label="Keluar sesi"
              className={styles.iconBtn}
            >
              <IdentityIcon name="logout" size={16} />
            </button>
          </div>
        </header>

        <main ref={contentRef} className={styles.pageContent}>
          <div data-ef-page-body>
            {isLoadingContext ? (
              <div className={`${styles.hairlinePanel} space-y-3 p-6`}>
                <div className={`${styles.skeletonShimmer} h-5 w-48`} />
                <div className={`${styles.skeletonShimmer} h-4 w-72`} />
                <div className={`${styles.skeletonShimmer} h-32 w-full`} />
              </div>
            ) : accessDenied ? (
              <section className={`${styles.hairlinePanel} max-w-xl p-6`}>
                <p className={`text-xs font-semibold text-[#b42318] ${styles.monoText}`}>
                  403 FORBIDDEN — AKSES TERBATAS
                </p>
                <h2 className="mt-2 text-lg font-bold text-[#0b2439]">
                  Izin Enterprise Foundation Diperlukan
                </h2>
                <p className="mt-2 text-sm text-[#5a6672]">
                  Akun Anda tidak memegang peran{" "}
                  <code className={styles.monoText}>superadmin</code> atau izin{" "}
                  <code className={styles.monoText}>enterprise_foundation.*</code>{" "}
                  yang disyaratkan untuk mengakses workspace ini.
                </p>
                <div className="mt-4 flex flex-wrap gap-2">
                  <Link href="/login" className={styles.btnPrimary}>
                    Masuk Akun Lain
                  </Link>
                  <Link href="/" className={styles.btnSecondary}>
                    Kembali ke Beranda
                  </Link>
                </div>
              </section>
            ) : (
              children
            )}
          </div>
        </main>
      </div>

      <WorkspaceCommandModal items={visibleItems} />
      <StepUpMfaModal />
      <WorkspaceToastViewport />
    </div>
  );
}

export function WorkspaceShell({
  domain,
  children,
}: {
  domain: Exclude<WorkspaceDomain, "identity">;
  children: ReactNode;
}) {
  return (
    <IdentityProvider workspaceDomain={domain}>
      <WorkspaceShellFrame domain={domain}>{children}</WorkspaceShellFrame>
    </IdentityProvider>
  );
}
