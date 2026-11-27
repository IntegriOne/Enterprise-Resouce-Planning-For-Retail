import { Suspense, type ReactNode } from "react";
import type { Metadata } from "next";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import {
  AuthRequestError,
  fetchMyWorkspaceAccessServer,
} from "@/lib/graphql/iam-browser-auth";
import { IdentityShell } from "@/components/identity/identity-shell";

export const metadata: Metadata = {
  title: "Identitas & Akses (IAM) — IntegriOne",
  description: "Panel tata kelola identitas, peran RBAC, grup, dan keamanan sesi IntegriOne.",
  robots: { index: false, follow: false },
};

async function IdentityShellGuard({ children }: { children: ReactNode }) {
  const cookieStore = await cookies();
  const allCookies = cookieStore.getAll();

  const hasSessionCookie = allCookies.some(
    (c) => c.name === "__Host-integrione_session" && c.value.length > 0,
  );
  if (!hasSessionCookie) {
    redirect("/login");
  }

  const cookieHeader = allCookies
    .map((c) => `${encodeURIComponent(c.name)}=${encodeURIComponent(c.value)}`)
    .join("; ");

  let canAccess = false;
  let shouldRedirectToLogin = false;

  try {
    const access = await fetchMyWorkspaceAccessServer(cookieHeader);
    canAccess = access.canAccessIdentity;
  } catch (err) {
    if (err instanceof AuthRequestError && err.code === "UNAUTHENTICATED") {
      shouldRedirectToLogin = true;
    } else {
      canAccess = false;
    }
  }

  if (shouldRedirectToLogin) {
    redirect("/login");
  }

  if (!canAccess) {
    return (
      <main className="flex min-h-[100dvh] items-center justify-center bg-[#f9fafb] px-5 py-10 text-[#17202a]">
        <section className="w-full max-w-md rounded-xl border border-[#dde3e8] bg-white p-8">
          <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl border border-[#fecdca] bg-[#fef3f2] text-[#b42318]">
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.75"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <circle cx="12" cy="12" r="9" />
              <path d="m4.9 4.9 14.2 14.2" />
            </svg>
          </div>
          <p className="font-mono text-xs font-semibold tracking-wider text-[#b42318]">
            FORBIDDEN · 403
          </p>
          <h1 className="mt-1 text-xl font-bold tracking-tight text-[#0b2439]">
            Akses Ruang Kerja IAM Ditolak
          </h1>
          <p className="mt-2 text-sm leading-relaxed text-[#5a6672]">
            Akun Anda aktif tetapi tidak memegang izin administrasi modul
            Identitas (IAM). Hubungi administrator berwenang bila akses ini
            diperlukan.
          </p>
          <div className="mt-6 flex flex-wrap gap-2.5">
            <Link
              href="/"
              className="inline-flex min-h-[42px] items-center justify-center rounded-lg bg-[#133554] px-4 text-sm font-semibold text-[#f9fafb] transition-colors hover:bg-[#0b2439]"
            >
              Kembali ke Beranda
            </Link>
            <Link
              href="/login"
              className="inline-flex min-h-[42px] items-center justify-center rounded-lg border border-[#dde3e8] bg-white px-4 text-sm font-semibold text-[#17202a] transition-colors hover:bg-[#f4f7f8]"
            >
              Halaman Masuk
            </Link>
          </div>
        </section>
      </main>
    );
  }

  return <IdentityShell>{children}</IdentityShell>;
}

function IdentityGuardFallback() {
  return (
    <div className="flex min-h-[100dvh] items-center justify-center bg-[#f9fafb] px-4">
      <div className="w-full max-w-sm space-y-3 rounded-xl border border-[#dde3e8] bg-white p-6">
        <div className="h-4 w-32 animate-pulse rounded bg-[#e9eef2]" />
        <div className="h-3 w-full animate-pulse rounded bg-[#f4f7f8]" />
        <p className="font-mono text-xs text-[#5a6672]">
          Memverifikasi sesi dan AccessSnapshot IAM...
        </p>
      </div>
    </div>
  );
}

export default function IdentityLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <Suspense fallback={<IdentityGuardFallback />}>
      <IdentityShellGuard>{children}</IdentityShellGuard>
    </Suspense>
  );
}
