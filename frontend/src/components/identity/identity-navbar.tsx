"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import styles from "./identity.module.css";
import { IdentityIcon } from "./identity-icons";
import { useIdentityWorkspace } from "./identity-context";
import { IDENTITY_NAV_ITEMS } from "./identity-sidebar";
import { WorkspaceSwitcher } from "@/components/workspace/workspace-switcher";

function resolveRouteLabel(pathname: string): { title: string; code: string } {
  if (pathname === "/identity/users" || pathname.startsWith("/identity/users/")) {
    if (pathname.startsWith("/identity/users/customers")) {
      return { title: "Customers", code: "CUSTOMER" };
    }
    if (pathname.startsWith("/identity/users/vendors")) {
      return { title: "Vendors", code: "VENDOR" };
    }
    if (pathname.startsWith("/identity/users/internals")) {
      return { title: "Internals", code: "INTERNAL" };
    }
    return { title: "Semua Pengguna", code: "USERS" };
  }

  const match = IDENTITY_NAV_ITEMS.find((item) =>
    item.exact
      ? pathname === item.href
      : pathname === item.href || pathname.startsWith(`${item.href}/`),
  );
  if (!match) {
    return { title: "Identitas & Akses", code: "IAM" };
  }
  const codeMap: Record<string, string> = {
    "/identity": "OVERVIEW",
    "/identity/users/customers": "CUSTOMER",
    "/identity/users/vendors": "VENDOR",
    "/identity/users/internals": "INTERNAL",
    "/identity/roles": "RBAC_ROLES",
    "/identity/groups": "GROUPS",
    "/identity/permissions": "CATALOG",
    "/identity/sessions": "SECURITY",
  };
  return { title: match.label, code: codeMap[match.href] ?? "IAM" };
}

export function IdentityNavbar({
  sidebarCollapsed,
  onToggleCollapse,
  onOpenMobileMenu,
}: {
  sidebarCollapsed: boolean;
  onToggleCollapse: () => void;
  onOpenMobileMenu: () => void;
}) {
  const pathname = usePathname();
  const {
    me,
    snapshot,
    isLoggingOut,
    openCommandPalette,
    openStepUpModal,
    logout,
  } = useIdentityWorkspace();

  const [menuOpen, setMenuOpen] = useState(false);
  const menuWrapRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      if (
        menuWrapRef.current &&
        !menuWrapRef.current.contains(e.target as Node)
      ) {
        setMenuOpen(false);
      }
    }
    function handleEscape(e: KeyboardEvent) {
      if (e.key === "Escape") setMenuOpen(false);
    }
    document.addEventListener("mousedown", handleOutsideClick);
    window.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      window.removeEventListener("keydown", handleEscape);
    };
  }, []);

  const routeMeta = resolveRouteLabel(pathname);
  const initials = (me?.display_name || "IA")
    .split(/\s+/)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() || "")
    .join("");

  return (
    <header className={styles.topNavbar}>
      <div className={styles.navbarLeft}>
        <button
          type="button"
          onClick={onOpenMobileMenu}
          aria-label="Buka navigasi samping"
          className={`${styles.iconBtn} lg:hidden`}
        >
          <IdentityIcon name="menu" size={18} />
        </button>

        <button
          type="button"
          onClick={onToggleCollapse}
          aria-label={
            sidebarCollapsed ? "Perluas sidebar IAM" : "Ciutkan sidebar IAM"
          }
          title={sidebarCollapsed ? "Perluas sidebar" : "Ciutkan sidebar"}
          className={`${styles.iconBtn} hidden lg:inline-flex`}
        >
          <IdentityIcon name="panelCollapse" size={17} />
        </button>

        <div className="flex min-w-0 items-center gap-2.5">
          <span
            className={`hidden rounded border border-[#dde3e8] bg-white px-1.5 py-0.5 text-[10px] font-semibold tracking-wider text-[#133554] sm:inline-block ${styles.monoText}`}
          >
            {routeMeta.code}
          </span>
          <span className="truncate text-sm font-semibold text-[#0b2439]">
            {routeMeta.title}
          </span>
        </div>
      </div>

      <div className={styles.navbarRight}>
        <WorkspaceSwitcher variant="navbar" />
        <button
          type="button"
          onClick={openCommandPalette}
          className={styles.commandTriggerBtn}
          aria-label="Buka navigasi cepat IAM (Ctrl+K)"
        >
          <IdentityIcon name="search" size={15} />
          <span className="hidden md:inline">Navigasi Cepat</span>
          <kbd className={`${styles.kbdShortcut} hidden sm:inline`}>Ctrl K</kbd>
        </button>

        <button
          type="button"
          onClick={() =>
            openStepUpModal(
              "Konfirmasi ulang kode TOTP sesi aktif untuk membuka aksi sensitif IAM.",
            )
          }
          className={styles.stepUpPill}
          title="Verifikasi Step-Up MFA"
        >
          <IdentityIcon
            name={snapshot?.mfa_enabled ? "shield" : "lock"}
            size={14}
          />
          <span className="hidden sm:inline">
            {snapshot?.mfa_enabled ? "MFA Aktif" : "Step-Up"}
          </span>
        </button>

        <div ref={menuWrapRef} className="relative">
          <button
            type="button"
            onClick={() => setMenuOpen((prev) => !prev)}
            aria-expanded={menuOpen}
            aria-haspopup="menu"
            className="flex min-h-[40px] items-center gap-2.5 rounded-xl border border-[#dde3e8] bg-white px-2.5 py-1 text-left transition-colors hover:border-[#133554]"
          >
            <span
              className={`flex h-7 w-7 items-center justify-center rounded-lg bg-[#133554] text-xs font-semibold text-[#f9fafb] ${styles.monoText}`}
            >
              {initials || "IA"}
            </span>
            <span className="hidden min-w-0 flex-col md:flex">
              <span className="max-w-[140px] truncate text-xs font-semibold text-[#0b2439]">
                {me?.display_name || "Operator IAM"}
              </span>
              <span className={`text-[10px] text-[#5a6672] ${styles.monoText}`}>
                {me?.kind || "LOADING"}
              </span>
            </span>
          </button>

          {menuOpen ? (
            <div
              role="menu"
              className="absolute right-0 mt-2 w-64 rounded-xl border border-[#dde3e8] bg-white p-2 shadow-xl"
            >
              <div className="border-b border-[#dde3e8] px-3 py-2.5">
                <p className="truncate text-sm font-semibold text-[#0b2439]">
                  {me?.display_name || "Operator IAM"}
                </p>
                <p
                  className={`mt-0.5 truncate text-[11px] text-[#5a6672] ${styles.monoText}`}
                >
                  {me?._id || snapshot?.principal_id || "-"}
                </p>
              </div>

              <div className="py-1.5">
                <Link
                  href="/identity/sessions"
                  role="menuitem"
                  onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-xs font-medium text-[#17202a] hover:bg-[#f4f7f8]"
                >
                  <IdentityIcon name="sessions" size={15} />
                  <span>Sesi Aktif &amp; Keamanan Akun</span>
                </Link>
                <button
                  type="button"
                  role="menuitem"
                  onClick={() => {
                    setMenuOpen(false);
                    openStepUpModal(
                      "Verifikasi TOTP untuk operasi sensitif 10 menit ke depan.",
                    );
                  }}
                  className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left text-xs font-medium text-[#17202a] hover:bg-[#f4f7f8]"
                >
                  <IdentityIcon name="shield" size={15} />
                  <span>Verifikasi Step-Up MFA</span>
                </button>
              </div>

              <div className="border-t border-[#dde3e8] pt-1.5">
                <button
                  type="button"
                  role="menuitem"
                  disabled={isLoggingOut}
                  onClick={() => void logout()}
                  className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-left text-xs font-semibold text-[#b42318] hover:bg-[#fef3f2]"
                >
                  <IdentityIcon name="logout" size={15} />
                  <span>{isLoggingOut ? "Mengakhiri sesi..." : "Keluar Sesi"}</span>
                </button>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </header>
  );
}
