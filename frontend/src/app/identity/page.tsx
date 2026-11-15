import { Suspense } from "react";
import type { Metadata } from "next";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import {
  fetchMyWorkspaceAccessServer,
  AuthRequestError,
} from "@/lib/graphql/iam-browser-auth";

export const metadata: Metadata = {
  title: "Identitas IntegriOne",
  description: "Manajemen Akses & Identitas Terpusat IntegriOne.",
  robots: { index: false, follow: false },
};

async function IdentityGuard() {
  const cookieStore = await cookies();
  const allCookies = cookieStore.getAll();

  // Jika tidak ada cookie sesi sama sekali, arahkan ke /login
  if (!allCookies || allCookies.length === 0) {
    redirect("/login");
  }

  // Format header Cookie untuk dikirim ke internal Gateway GraphQL
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
      // Fail-closed jika error atau IAM unavailable
      canAccess = false;
    }
  }

  // Panggil redirect di luar blok try-catch agar error NEXT_REDIRECT tidak tertelan
  if (shouldRedirectToLogin) {
    redirect("/login");
  }

  if (!canAccess) {
    return (
      <main
        style={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#f9fafb",
          fontFamily: "var(--font-source-sans-3, sans-serif)",
          padding: "20px",
        }}
      >
        <div
          style={{
            maxWidth: "460px",
            width: "100%",
            backgroundColor: "#ffffff",
            padding: "36px 32px",
            borderRadius: "16px",
            border: "1px solid #dde3e8",
            textAlign: "center",
            boxShadow: "0 4px 16px rgba(11, 36, 57, 0.06)",
          }}
        >
          <div
            style={{
              width: "48px",
              height: "48px",
              borderRadius: "50%",
              backgroundColor: "#fef2f2",
              color: "#dc2626",
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              marginBottom: "16px",
            }}
          >
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="4.93" y1="4.93" x2="19.07" y2="19.07" />
            </svg>
          </div>
          <h1
            style={{
              fontFamily: "var(--font-fraunces, serif)",
              fontSize: "1.5rem",
              color: "#0b2439",
              margin: "0 0 8px",
            }}
          >
            Akses Terbatas (403)
          </h1>
          <p
            style={{
              fontSize: "0.9375rem",
              color: "#66717c",
              lineHeight: 1.5,
              margin: "0 0 24px",
            }}
          >
            Akun Anda tidak memiliki izin untuk mengakses ruang kerja Manajemen
            Identitas (IAM). Ruang kerja ini hanya dapat diakses oleh Super Admin
            yang berwenang.
          </p>
          <Link
            href="/"
            style={{
              display: "inline-block",
              padding: "12px 24px",
              backgroundColor: "#133554",
              color: "#ffffff",
              borderRadius: "8px",
              textDecoration: "none",
              fontWeight: 600,
              fontSize: "0.875rem",
            }}
          >
            Kembali ke Beranda
          </Link>
        </div>
      </main>
    );
  }

  // Authorized Admin Workspace — Blank / Minimal semantik aman sesuai instruksi
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        backgroundColor: "#f9fafb",
        fontFamily: "var(--font-source-sans-3, sans-serif)",
        color: "#17202a",
      }}
    >
      <header
        style={{
          height: "64px",
          borderBottom: "1px solid #dde3e8",
          backgroundColor: "#ffffff",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "0 24px",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
          <span
            style={{
              fontFamily: "var(--font-fraunces, serif)",
              fontSize: "1.25rem",
              fontWeight: 600,
              color: "#0b2439",
            }}
          >
            IntegriOne
          </span>
          <span
            style={{
              fontSize: "0.75rem",
              fontWeight: 600,
              backgroundColor: "#e9eef2",
              color: "#133554",
              padding: "3px 8px",
              borderRadius: "6px",
              letterSpacing: "0.06em",
              textTransform: "uppercase",
            }}
          >
            IAM Workspace
          </span>
        </div>
        <div>
          <Link
            href="/login"
            style={{
              fontSize: "0.875rem",
              color: "#66717c",
              textDecoration: "none",
            }}
          >
            Keluar Sesi
          </Link>
        </div>
      </header>

      <main
        style={{
          flex: 1,
          padding: "32px 24px",
          maxWidth: "1200px",
          margin: "0 auto",
          width: "100%",
        }}
      >
        <div
          style={{
            borderBottom: "1px solid #dde3e8",
            paddingBottom: "16px",
            marginBottom: "24px",
          }}
        >
          <h1
            style={{
              fontFamily: "var(--font-fraunces, serif)",
              fontSize: "1.75rem",
              color: "#0b2439",
              margin: "0 0 6px",
            }}
          >
            Identitas IntegriOne
          </h1>
          <p style={{ fontSize: "0.9375rem", color: "#66717c", margin: 0 }}>
            Ruang kerja manajemen identitas terpusat (IAM).
          </p>
        </div>

        <div
          style={{
            backgroundColor: "#ffffff",
            border: "1px solid #dde3e8",
            borderRadius: "12px",
            padding: "40px",
            textAlign: "center",
            color: "#66717c",
          }}
        >
          <p style={{ margin: 0, fontSize: "0.9375rem" }}>
            Ruang kerja IAM terautentikasi dan terlindungi server-side.
          </p>
        </div>
      </main>
    </div>
  );
}

function IdentityFallback() {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "#f9fafb",
        fontFamily: "var(--font-source-sans-3, sans-serif)",
        color: "#66717c",
      }}
    >
      <p style={{ margin: 0, fontSize: "0.9375rem" }}>
        Memeriksa izin akses ruang kerja...
      </p>
    </div>
  );
}

export default function IdentityPage() {
  return (
    <Suspense fallback={<IdentityFallback />}>
      <IdentityGuard />
    </Suspense>
  );
}
