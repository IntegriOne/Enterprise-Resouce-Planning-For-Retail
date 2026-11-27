"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import styles from "./identity.module.css";
import {
  IdentityBrandMark,
  IdentityIcon,
  type IdentityIconName,
} from "./identity-icons";
import { useIdentityWorkspace } from "./identity-context";
import { formatUnixTimeID } from "./identity-primitives";
import { WorkspaceSwitcher } from "@/components/workspace/workspace-switcher";

type NavItem = {
  href: string;
  label: string;
  icon: IdentityIconName;
  badge?: string;
  exact?: boolean;
};

const ACCESS_ITEMS: NavItem[] = [
  { href: "/identity", label: "Ringkasan", icon: "overview", exact: true },
  { href: "/identity/roles", label: "Peran RBAC", icon: "roles" },
  { href: "/identity/groups", label: "Grup", icon: "groups" },
  { href: "/identity/permissions", label: "Katalog Izin", icon: "key" },
];

const USER_DIRECTORY_ITEMS: NavItem[] = [
  {
    href: "/identity/users/customers",
    label: "Customers",
    icon: "users",
    badge: "CUSTOMER",
  },
  {
    href: "/identity/users/vendors",
    label: "Vendors",
    icon: "building",
    badge: "VENDOR",
  },
  {
    href: "/identity/users/internals",
    label: "Internals",
    icon: "badge",
    badge: "INTERNAL",
  },
];

const SECURITY_ITEMS: NavItem[] = [
  { href: "/identity/sessions", label: "Sesi & Keamanan", icon: "sessions" },
];

export const IDENTITY_NAV_ITEMS: NavItem[] = [
  ...ACCESS_ITEMS,
  ...USER_DIRECTORY_ITEMS,
  ...SECURITY_ITEMS,
];

export function IdentitySidebar({
  collapsed,
  mobileOpen,
  onCloseMobile,
}: {
  collapsed: boolean;
  mobileOpen: boolean;
  onCloseMobile: () => void;
}) {
  const pathname = usePathname();
  const { snapshot, isLoadingContext } = useIdentityWorkspace();

  function isActive(item: NavItem): boolean {
    if (item.exact) {
      return pathname === item.href;
    }
    return pathname === item.href || pathname.startsWith(`${item.href}/`);
  }

  return (
    <>
      {mobileOpen ? (
        <div
          className={styles.sidebarBackdrop}
          onClick={onCloseMobile}
          aria-hidden="true"
        />
      ) : null}

      <aside
        aria-label="Navigasi Modul Identitas"
        className={`${styles.sidebar} ${collapsed ? styles.sidebarCollapsed : ""} ${
          mobileOpen ? styles.sidebarOpen : ""
        }`}
      >
        <div className={styles.sidebarHeader}>
          <Link
            href="/identity"
            onClick={onCloseMobile}
            className={styles.brandLockup}
          >
            <div className={styles.brandIconTile}>
              <IdentityBrandMark />
            </div>
            {!collapsed ? (
              <div className={styles.brandMeta}>
                <span className={styles.brandTitle}>IntegriOne</span>
                <span className={styles.brandSub}>IAM Control</span>
              </div>
            ) : null}
          </Link>

          <button
            type="button"
            onClick={onCloseMobile}
            aria-label="Tutup menu samping"
            className={`${styles.iconBtn} lg:hidden`}
          >
            <IdentityIcon name="close" size={17} />
          </button>
        </div>

        <WorkspaceSwitcher compact={collapsed} variant="sidebar" />

        <nav className={styles.sidebarNavScroll}>
          <div>
            {!collapsed ? (
              <p className={styles.navGroupLabel}>Tata Kelola Akses</p>
            ) : null}
            <div className={styles.navList}>
              {ACCESS_ITEMS.map((item) => {
                const active = isActive(item);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={onCloseMobile}
                    title={collapsed ? item.label : undefined}
                    aria-current={active ? "page" : undefined}
                    className={`${styles.navItem} ${active ? styles.navItemActive : ""}`}
                  >
                    <IdentityIcon name={item.icon} size={18} />
                    {!collapsed ? <span>{item.label}</span> : null}
                  </Link>
                );
              })}
            </div>
          </div>

          <div>
            {!collapsed ? (
              <p className={styles.navGroupLabel}>Pengguna</p>
            ) : null}
            <div className={styles.navList}>
              {USER_DIRECTORY_ITEMS.map((item) => {
                const active = isActive(item);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={onCloseMobile}
                    title={
                      collapsed
                        ? `${item.label}${item.badge ? ` (${item.badge})` : ""}`
                        : undefined
                    }
                    aria-current={active ? "page" : undefined}
                    className={`${styles.navItem} ${active ? styles.navItemActive : ""}`}
                  >
                    <IdentityIcon name={item.icon} size={18} />
                    {!collapsed ? (
                      <>
                        <span>{item.label}</span>
                        {item.badge ? (
                          <span className={styles.navItemBadge}>{item.badge}</span>
                        ) : null}
                      </>
                    ) : null}
                  </Link>
                );
              })}
            </div>
          </div>

          <div>
            {!collapsed ? (
              <p className={styles.navGroupLabel}>Audit &amp; Keamanan</p>
            ) : null}
            <div className={styles.navList}>
              {SECURITY_ITEMS.map((item) => {
                const active = isActive(item);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={onCloseMobile}
                    title={collapsed ? item.label : undefined}
                    aria-current={active ? "page" : undefined}
                    className={`${styles.navItem} ${active ? styles.navItemActive : ""}`}
                  >
                    <IdentityIcon name={item.icon} size={18} />
                    {!collapsed ? <span>{item.label}</span> : null}
                  </Link>
                );
              })}
            </div>
          </div>
        </nav>

        <div className={styles.sidebarFooter}>
          {collapsed ? (
            <div
              className="flex justify-center py-1"
              title={
                snapshot
                  ? `Epoch #${snapshot.auth_epoch} / Perm #${snapshot.perm_revision}`
                  : "Menyinkronkan snapshot"
              }
            >
              <span
                className={`${styles.liveDot} ${
                  !snapshot ? styles.liveDotWarn : ""
                }`}
              />
            </div>
          ) : (
            <div className={styles.snapshotCard}>
              <div className={styles.snapshotRow}>
                <span className="flex items-center gap-1.5 font-semibold text-[#0b2439]">
                  <span
                    className={`${styles.liveDot} ${
                      isLoadingContext ? styles.liveDotWarn : ""
                    }`}
                  />
                  <span>AccessSnapshot</span>
                </span>
                <span className={`text-[#5a6672] ${styles.monoText}`}>
                  {snapshot ? formatUnixTimeID(snapshot.fresh_until_unix) : "--:--"}
                </span>
              </div>
              <div className={`${styles.snapshotRow} ${styles.monoText} text-[#5a6672]`}>
                <span>auth_epoch</span>
                <span className="font-semibold text-[#0b2439]">
                  {snapshot ? `v${snapshot.auth_epoch}` : "..."}
                </span>
              </div>
              <div className={`${styles.snapshotRow} ${styles.monoText} text-[#5a6672]`}>
                <span>perm_revision</span>
                <span className="font-semibold text-[#0b2439]">
                  {snapshot ? `r${snapshot.perm_revision}` : "..."}
                </span>
              </div>
            </div>
          )}
        </div>
      </aside>
    </>
  );
}
