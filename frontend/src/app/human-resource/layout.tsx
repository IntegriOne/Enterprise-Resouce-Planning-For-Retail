import { Suspense, type ReactNode } from "react";
import type { Metadata } from "next";
import {
  EfWorkspaceServerGuard,
  WorkspaceGuardFallback,
} from "@/components/workspace/ef-server-guard";

export const metadata: Metadata = {
  title: "SDM dan Struktur Perusahaan — IntegriOne",
  description:
    "Pengelolaan pekerja, badan hukum, unit organisasi, dan lokasi IntegriOne.",
  robots: { index: false, follow: false },
};

export default function HumanResourceLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <Suspense fallback={<WorkspaceGuardFallback />}>
      <EfWorkspaceServerGuard domain="human-resource">
        {children}
      </EfWorkspaceServerGuard>
    </Suspense>
  );
}
