"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import gsap from "gsap";
import { useGSAP } from "@gsap/react";
import styles from "./identity.module.css";
import { IdentityProvider, useIdentityWorkspace } from "./identity-context";
import { IDENTITY_NAV_ITEMS, IdentitySidebar } from "./identity-sidebar";
import { IdentityNavbar } from "./identity-navbar";
import { StepUpMfaModal } from "./stepup-mfa-modal";
import { IdentityIcon } from "./identity-icons";

gsap.registerPlugin(useGSAP);

function CommandPaletteInner({ onClose }: { onClose: () => void }) {
  const router = useRouter();
  const [query, setQuery] = useState("");

  useEffect(() => {
    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", handleEsc);
    return () => window.removeEventListener("keydown", handleEsc);
  }, [onClose]);

  const filtered = IDENTITY_NAV_ITEMS.filter(
    (item) =>
      item.label.toLowerCase().includes(query.trim().toLowerCase()) ||
      item.href.toLowerCase().includes(query.trim().toLowerCase()),
  );

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-[#0b2439]/40 p-4 pt-20 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label="Navigasi Cepat IAM"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-full max-w-lg overflow-hidden rounded-xl border border-[#dde3e8] bg-white shadow-2xl">
        <div className="flex items-center gap-2.5 border-b border-[#dde3e8] px-4 py-3">
          <IdentityIcon name="search" size={17} className="text-[#5a6672]" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Ketik halaman IAM: Pengguna, Peran, Grup, Izin, Sesi..."
            aria-label="Cari navigasi IAM"
            autoFocus
            className="w-full bg-transparent text-sm text-[#0b2439] placeholder-[#788491] focus:outline-none"
          />
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup pencarian cepat"
            className="rounded px-1.5 py-0.5 text-xs text-[#5a6672] hover:bg-[#f4f7f8]"
          >
            Esc
          </button>
        </div>

        <div className="max-h-72 divide-y divide-[#dde3e8] overflow-y-auto p-1.5">
          {filtered.length === 0 ? (
            <p className="px-3 py-6 text-center text-xs text-[#5a6672]">
              Tidak ada rute yang cocok.
            </p>
          ) : (
            filtered.map((item) => (
              <button
                key={item.href}
                type="button"
                onClick={() => {
                  onClose();
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
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function CommandPaletteModal() {
  const { isCommandOpen, closeCommandPalette } = useIdentityWorkspace();
  if (!isCommandOpen) return null;
  return <CommandPaletteInner onClose={closeCommandPalette} />;
}

function ToastViewport() {
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

function IdentityShellFrame({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const contentRef = useRef<HTMLElement>(null);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  useGSAP(
    () => {
      const mm = gsap.matchMedia();
      mm.add("(prefers-reduced-motion: no-preference)", () => {
        gsap.fromTo(
          "[data-iam-page-body]",
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

  return (
    <div className={styles.workspaceRoot}>
      <IdentitySidebar
        collapsed={sidebarCollapsed}
        mobileOpen={mobileOpen}
        onCloseMobile={() => setMobileOpen(false)}
      />

      <div className={styles.mainColumn}>
        <IdentityNavbar
          sidebarCollapsed={sidebarCollapsed}
          onToggleCollapse={() => setSidebarCollapsed((prev) => !prev)}
          onOpenMobileMenu={() => setMobileOpen(true)}
        />

        <main ref={contentRef} className={styles.pageContent}>
          <div data-iam-page-body>{children}</div>
        </main>
      </div>

      <CommandPaletteModal />
      <StepUpMfaModal />
      <ToastViewport />
    </div>
  );
}

export function IdentityShell({ children }: { children: ReactNode }) {
  return (
    <IdentityProvider>
      <IdentityShellFrame>{children}</IdentityShellFrame>
    </IdentityProvider>
  );
}
