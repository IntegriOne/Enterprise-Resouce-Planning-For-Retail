"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import styles from "@/components/identity/identity.module.css";
import { AuthRequestError, executeGraphQL } from "@/lib/graphql/iam-browser-auth";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { isForbiddenError } from "@/lib/graphql/iam-dashboard-queries";

type Payslip = {
  payroll_result_id: string;
  period_code: string;
  period_start: string;
  period_end: string;
  published_at: string;
  scheduled_pay_date: string;
  payment_status: string;
};

type PayslipPage = {
  nodes: Payslip[];
  total_count: number;
  pageInfo: PageInfo;
};

const pageSize = 20;
const paymentLabels: Record<string, string> = {
  scheduled: "Terjadwal",
  processing: "Dalam proses",
  paid: "Dibayar",
  failed: "Pembayaran gagal",
};

const dateFormat = new Intl.DateTimeFormat("id-ID", {
  dateStyle: "long",
  timeZone: "UTC",
});

const publicationTimeFormat = new Intl.DateTimeFormat("id-ID", {
  dateStyle: "long",
  timeStyle: "short",
  timeZone: "UTC",
});

function formatDate(value: string, formatter = dateFormat) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "Belum tercatat" : formatter.format(date);
}

export default function PayslipsClient() {
  const [request, setRequest] = useState({ after: null as string | null, before: null as string | null });
  const [page, setPage] = useState<PayslipPage | null>(null);
  const [failure, setFailure] = useState<"session" | "access" | "load" | null>(null);

  const [pdf, setPDF] = useState<{ url: string; period: string } | null>(null);
  const [pdfPending, setPDFPending] = useState(false);
  const [pdfError, setPDFError] = useState<string | null>(null);
  const openURL = useRef<string | null>(null);
  const mounted = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      if (openURL.current) URL.revokeObjectURL(openURL.current);
    };
  }, []);

  useEffect(() => {
    let active = true;
    void executeGraphQL<{ myPayslips: PayslipPage }>(
      `query EmployeePayslips($first: Int, $last: Int, $after: String, $before: String) {
        myPayslips(first: $first, last: $last, after: $after, before: $before) {
          total_count
          pageInfo { startCursor endCursor hasPreviousPage hasNextPage }
          nodes {
            payroll_result_id period_code period_start period_end
            published_at scheduled_pay_date payment_status
          }
        }
      }`,
      { first: request.before ? null : pageSize, last: request.before ? pageSize : null, after: request.after, before: request.before },
    ).then(
      ({ data }) => { if (active) setPage(data.myPayslips); },
      (error: unknown) => {
        if (!active) return;
        if (error instanceof AuthRequestError &&
            (error.status === 401 || error.code === "UNAUTHENTICATED")) {
          setFailure("session");
        } else {
          setFailure(isForbiddenError(error) ? "access" : "load");
        }
      },
    );
    return () => { active = false; };
  }, [request]);

  function closePDF() {
    if (openURL.current) URL.revokeObjectURL(openURL.current);
    openURL.current = null;
    setPDF(null);
    setPDFError(null);
  }

  async function preparePDF(slip: Payslip) {
    if (pdfPending) return;
    closePDF();
    setPDFPending(true);
    try {
      const { data } = await executeGraphQL<{
        myPayslipPdf: { content_base64: string } | null;
      }>(
        `query EmployeePayslipPDF($result: ID!) {
          myPayslipPdf(payroll_result_id: $result) { content_base64 }
        }`,
        { result: slip.payroll_result_id },
      );
      if (!mounted.current) return;
      const encoded = data.myPayslipPdf?.content_base64;
      if (!encoded) throw new Error("Missing payslip PDF");
      const content = Uint8Array.from(atob(encoded), (byte) => byte.charCodeAt(0));
      const url = URL.createObjectURL(new Blob([content], { type: "application/pdf" }));
      openURL.current = url;
      setPDF({ url, period: slip.period_code || formatDate(slip.period_start) });
    } catch (error) {
      if (mounted.current) {
        setPDFError(isForbiddenError(error)
          ? "Anda tidak memiliki akses ke PDF slip ini."
          : "PDF slip gaji belum dapat dibuka. Coba lagi.");
      }
    } finally {
      if (mounted.current) setPDFPending(false);
    }
  }

  function moveTo(after: string | null, before: string | null) {
    closePDF();
    setPage(null);
    setFailure(null);
    setRequest({ after, before });
  }

  return (
    <main className="mx-auto max-w-5xl space-y-5 px-4 py-10">
      <h1 className="text-2xl font-bold text-[#0b2439]">Slip Gaji Saya</h1>
      {!page && !failure ? <p role="status">Memuat slip gaji…</p> : null}
      {failure === "session" ? (
        <p role="alert">Sesi berakhir. <Link href="/login">Masuk kembali</Link>.</p>
      ) : null}
      {failure === "access" ? <p role="alert">Anda tidak memiliki akses ke slip gaji.</p> : null}
      {failure === "load" ? (
        <div className="flex items-center gap-2">
          <p role="alert">Daftar slip gaji gagal dimuat.</p>
          <button type="button" className={styles.btnSecondary} onClick={() => moveTo(request.after, request.before)}>
            Coba lagi
          </button>
        </div>
      ) : null}
      {page && page.nodes.length === 0 ? <p>Belum ada slip gaji pada halaman ini.</p> : null}
      {page && page.nodes.length > 0 ? (
        <div className="overflow-x-auto rounded-xl border border-[#dde3e8]">
          <table className={styles.dataTable}>
            <thead><tr><th>Periode</th><th>Waktu terbit (UTC)</th><th>Tanggal bayar terjadwal</th><th>Status pembayaran</th><th>PDF</th></tr></thead>
            <tbody>{page.nodes.map((slip) => (
              <tr key={slip.payroll_result_id}>
                <td>
                  {slip.period_code ? <span className="block font-medium">{slip.period_code}</span> : null}
                  {formatDate(slip.period_start)} – {formatDate(slip.period_end)}
                </td>
                <td><time dateTime={slip.published_at}>{formatDate(slip.published_at, publicationTimeFormat)}</time></td>
                <td>{formatDate(slip.scheduled_pay_date)}</td>
                <td>{paymentLabels[slip.payment_status] ?? "Status belum dikenal"}</td>
                <td><button type="button" className={styles.btnSecondary} disabled={pdfPending}
                  onClick={() => void preparePDF(slip)}>Siapkan PDF</button></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : null}
      {pdfPending ? <p role="status">Menyiapkan PDF…</p> : null}
      {pdfError ? <p role="alert">{pdfError}</p> : null}
      {pdf ? (
        <div className="flex flex-wrap items-center gap-3 rounded-xl border border-[#dde3e8] p-4">
          <span>Slip gaji · {pdf.period}</span>
          <a href={pdf.url} target="_blank" rel="noopener noreferrer" className={styles.btnPrimary}>
            Buka PDF untuk dicetak
          </a>
          <button type="button" className={styles.btnSecondary} onClick={closePDF}>Tutup</button>
        </div>
      ) : null}
      {page && (page.pageInfo.hasPreviousPage || page.pageInfo.hasNextPage) ? (
        <nav aria-label="Halaman slip gaji" className="flex gap-2">
          <button type="button" className={styles.btnSecondary} disabled={pdfPending || (!page.pageInfo.hasPreviousPage || !page.pageInfo.startCursor)}
            onClick={() => moveTo(null, page.pageInfo.startCursor ?? null)}>Sebelumnya</button>
          <button type="button" className={styles.btnSecondary} disabled={pdfPending || (!page.pageInfo.hasNextPage || !page.pageInfo.endCursor)}
            onClick={() => moveTo(page.pageInfo.endCursor ?? null, null)}>Berikutnya</button>
        </nav>
      ) : null}
    </main>
  );
}
