import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import {
  AuthRequestError,
  fetchMyAccessSnapshotServer,
} from "@/lib/graphql/iam-browser-auth";
import { WorkspaceShell } from "./workspace-shell";
import type { WorkspaceDomain } from "@/components/identity/identity-context";

type GuardDomain = Exclude<WorkspaceDomain, "identity">;

function snapshotAllowsEfDomain(
  domain: GuardDomain,
  snap: Awaited<ReturnType<typeof fetchMyAccessSnapshotServer>>,
): boolean {
  const isSuperAdmin = (snap.roles ?? []).some((r) => {
    const code = (r.code || "").trim().toLowerCase();
    return code === "superadmin" || code === "super_admin";
  });
  if (isSuperAdmin) return true;

  const perms = new Set((snap.permissions ?? []).map((p) => p.permission_code));
  if (domain === "human-resource") {
    return (
      perms.has("human_resource.workers.read") ||
      snap.permissions.some((grant) =>
        grant.permission_code === "human_resource.candidates.read" && grant.scope.type === "GLOBAL",
      ) ||
      perms.has("enterprise_foundation.legal_entities.read") ||
      perms.has("enterprise_foundation.legal_entities.write") ||
      perms.has("enterprise_foundation.organization_units.read") ||
      perms.has("enterprise_foundation.organization_units.write") ||
      perms.has("enterprise_foundation.locations.read") ||
      perms.has("enterprise_foundation.locations.write") ||
      perms.has("enterprise_foundation.addresses.read") ||
      perms.has("enterprise_foundation.addresses.write")
    );
  }

  return (
    perms.has("enterprise_foundation.parties.read") ||
    perms.has("enterprise_foundation.parties.write") ||
    perms.has("enterprise_foundation.addresses.read") ||
    perms.has("enterprise_foundation.addresses.write")
  );
}

export function WorkspaceGuardFallback() {
  return (
    <div className="flex min-h-[100dvh] items-center justify-center bg-[#f9fafb] px-4">
      <div className="w-full max-w-sm space-y-3 rounded-xl border border-[#dde3e8] bg-white p-6">
        <div className="h-4 w-40 animate-pulse rounded bg-[#e9eef2]" />
        <div className="h-3 w-full animate-pulse rounded bg-[#f4f7f8]" />
        <div className="h-3 w-4/5 animate-pulse rounded bg-[#f4f7f8]" />
      </div>
    </div>
  );
}

export async function EfWorkspaceServerGuard({
  domain,
  children,
}: {
  domain: GuardDomain;
  children: ReactNode;
}) {
  const cookieStore = await cookies();
  const allCookies = cookieStore.getAll();
  const hasSessionCookie = allCookies.some(
    (c) => c.name === "__Host-integrione_session" && c.value.length > 0,
  );
  if (!hasSessionCookie) {
    redirect("/login");
  }

  const cookieHeader = allCookies.map((c) => `${c.name}=${c.value}`).join("; ");
  let snapshot: Awaited<ReturnType<typeof fetchMyAccessSnapshotServer>>;
  try {
    snapshot = await fetchMyAccessSnapshotServer(cookieHeader);
  } catch (err) {
    if (
      err instanceof AuthRequestError &&
      (err.status === 401 ||
        err.code === "UNAUTHENTICATED" ||
        err.errorType === "AUTHENTICATION")
    ) {
      redirect("/login");
    }

    return (
      <main className="flex min-h-[100dvh] items-center justify-center bg-[#f9fafb] px-4 py-12">
        <section className="w-full max-w-lg rounded-2xl border border-[#dde3e8] bg-white p-8 shadow-sm">
          <p className="font-mono text-xs font-semibold uppercase tracking-widest text-[#da414a]">
            503 SERVICE UNAVAILABLE
          </p>
          <h1 className="mt-2 text-2xl font-bold tracking-tight text-[#0b2439]">
            Otorisasi IAM Sedang Tidak Tersedia
          </h1>
          <p className="mt-3 text-sm leading-relaxed text-[#5a6672]">
            Workspace Enterprise Foundation gagal memverifikasi izin sesi Anda
            secara fail-closed. Muat ulang halaman atau coba beberapa saat lagi.
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href={domain === "human-resource" ? "/human-resource" : "/procurement"}
              className="inline-flex min-h-10 items-center justify-center rounded-lg bg-[#133554] px-4 text-sm font-semibold text-[#f9fafb]"
            >
              Coba Lagi
            </Link>
            <Link
              href="/login"
              className="inline-flex min-h-10 items-center justify-center rounded-lg border border-[#dde3e8] bg-white px-4 text-sm font-semibold text-[#0b2439]"
            >
              Halaman Masuk
            </Link>
          </div>
        </section>
      </main>
    );
  }

  if (!snapshotAllowsEfDomain(domain, snapshot)) {
    return (
      <main className="flex min-h-[100dvh] items-center justify-center bg-[#f9fafb] px-4 py-12">
        <section className="w-full max-w-lg rounded-2xl border border-[#dde3e8] bg-white p-8 shadow-sm">
          <p className="font-mono text-xs font-semibold uppercase tracking-widest text-[#da414a]">
            403 FORBIDDEN — AKSES TERBATAS
          </p>
          <h1 className="mt-2 text-2xl font-bold tracking-tight text-[#0b2439]">
            Izin Enterprise Foundation Diperlukan
          </h1>
          <p className="mt-3 text-sm leading-relaxed text-[#5a6672]">
            Sesi aktif Anda tidak memiliki peran atau izin{" "}
            <code className="font-mono">enterprise_foundation.*</code> yang
            disyaratkan untuk membuka area{" "}
            {domain === "human-resource"
              ? "Struktur Perusahaan (/human-resource)"
              : "Mitra & Procurement (/procurement)"}
            .
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href="/"
              className="inline-flex min-h-10 items-center justify-center rounded-lg bg-[#133554] px-4 text-sm font-semibold text-[#f9fafb]"
            >
              Kembali ke Beranda
            </Link>
            <Link
              href="/login"
              className="inline-flex min-h-10 items-center justify-center rounded-lg border border-[#dde3e8] bg-white px-4 text-sm font-semibold text-[#0b2439]"
            >
              Masuk Akun Lain
            </Link>
          </div>
        </section>
      </main>
    );
  }

  return <WorkspaceShell domain={domain}>{children}</WorkspaceShell>;
}
