"use client";

import Link from "next/link";
import { useIdentityWorkspace, type WorkspaceDomain } from "@/components/identity/identity-context";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import {
  hasAnyEfPartnerAccess,
  hasAnyEfStructureAccess,
} from "@/lib/graphql/ef-helpers";

type WorkspaceEntry = {
  id: WorkspaceDomain;
  href: string;
  label: string;
  shortLabel: string;
  subtitle: string;
  icon: "shield" | "building" | "badge";
};

const WORKSPACE_ENTRIES: WorkspaceEntry[] = [
  {
    id: "identity",
    href: "/identity",
    label: "IAM & Keamanan",
    shortLabel: "IAM",
    subtitle: "Pengguna, Peran & Izin",
    icon: "shield",
  },
  {
    id: "human-resource",
    href: "/human-resource",
    label: "HR & Struktur EF",
    shortLabel: "HR & Struktur",
    subtitle: "Badan Hukum, Unit, Lokasi",
    icon: "building",
  },
  {
    id: "procurement",
    href: "/procurement",
    label: "Mitra & Procurement",
    shortLabel: "Procurement",
    subtitle: "Party, Vendor CMT, Portal",
    icon: "badge",
  },
];

export function WorkspaceSwitcher({
  compact = false,
  variant = "sidebar",
}: {
  compact?: boolean;
  variant?: "sidebar" | "navbar";
}) {
  const { snapshot, workspaceDomain } = useIdentityWorkspace();

  const available = WORKSPACE_ENTRIES.filter((entry) => {
    if (!snapshot) return entry.id === workspaceDomain;
    if (entry.id === "identity") return snapshot.can_access_identity;
    if (entry.id === "human-resource") return hasAnyEfStructureAccess(snapshot);
    return hasAnyEfPartnerAccess(snapshot);
  });

  if (available.length <= 1 && compact) {
    return null;
  }

  if (variant === "navbar") {
    return (
      <nav
        aria-label="Pindah modul kerja"
        className="hidden items-center gap-1 rounded-lg border border-[#dde3e8] bg-white p-1 md:inline-flex"
      >
        {available.map((entry) => {
          const active = entry.id === workspaceDomain;
          return (
            <Link
              key={entry.id}
              href={entry.href}
              aria-current={active ? "page" : undefined}
              className={`inline-flex min-h-8 items-center gap-1.5 rounded-md px-2.5 text-xs font-semibold transition-colors ${
                active
                  ? "bg-[#133554] text-[#f9fafb]"
                  : "text-[#5a6672] hover:bg-[#e9eef2] hover:text-[#0b2439]"
              }`}
            >
              <IdentityIcon name={entry.icon} size={13} />
              <span>{entry.shortLabel}</span>
            </Link>
          );
        })}
      </nav>
    );
  }

  if (compact) {
    return (
      <div className="border-b border-[#dde3e8] px-2 py-2.5">
        <div className="flex flex-col gap-1.5">
          {available.map((entry) => {
            const active = entry.id === workspaceDomain;
            return (
              <Link
                key={entry.id}
                href={entry.href}
                title={entry.label}
                aria-label={entry.label}
                className={`flex h-9 w-full items-center justify-center rounded-lg border transition-colors ${
                  active
                    ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                    : "border-transparent bg-[#f4f7f8] text-[#5a6672] hover:border-[#dde3e8] hover:text-[#0b2439]"
                }`}
              >
                <IdentityIcon name={entry.icon} size={15} />
              </Link>
            );
          })}
        </div>
      </div>
    );
  }

  return (
    <div className="border-b border-[#dde3e8] bg-[#f9fafb] px-3 py-3">
      <p className={`px-1 pb-2 text-[10px] font-semibold uppercase tracking-wider text-[#788491] ${styles.monoText}`}>
        Workspace Aktif
      </p>
      <div className="grid grid-cols-1 gap-1.5">
        {available.map((entry) => {
          const active = entry.id === workspaceDomain;
          return (
            <Link
              key={entry.id}
              href={entry.href}
              aria-current={active ? "page" : undefined}
              className={`flex items-center gap-2.5 rounded-lg border px-2.5 py-2 text-left transition-colors ${
                active
                  ? "border-[#133554] bg-[#133554] text-[#f9fafb]"
                  : "border-[#dde3e8] bg-white text-[#17202a] hover:bg-[#e9eef2]"
              }`}
            >
              <IdentityIcon name={entry.icon} size={15} />
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-semibold leading-tight">
                  {entry.label}
                </p>
                <p
                  className={`truncate text-[10px] leading-tight ${
                    active ? "text-[#e9eef2]/85" : "text-[#5a6672]"
                  }`}
                >
                  {entry.subtitle}
                </p>
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
