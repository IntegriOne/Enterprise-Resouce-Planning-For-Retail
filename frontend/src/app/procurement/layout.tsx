import { Suspense, type ReactNode } from "react";
import type { Metadata } from "next";
import {
  EfWorkspaceServerGuard,
  WorkspaceGuardFallback,
} from "@/components/workspace/ef-server-guard";

export const metadata: Metadata = {
  title: "Mitra & Procurement (Enterprise Foundation) — IntegriOne",
  description:
    "Pengelolaan Master Mitra (Vendor CMT & Supplier), Peran Bisnis, PIC Portal IAM, dan Relasi Alamat.",
  robots: { index: false, follow: false },
};

export default function ProcurementLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <Suspense fallback={<WorkspaceGuardFallback />}>
      <EfWorkspaceServerGuard domain="procurement">
        {children}
      </EfWorkspaceServerGuard>
    </Suspense>
  );
}
