"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import styles from "@/components/identity/identity.module.css";
import {
  isForbiddenError,
  isStepUpRequiredError,
} from "@/lib/graphql/iam-dashboard-queries";
import {
  fetchWorkerAccountRequest,
  listProvisioningWorkers,
  submitWorkerAccountRequest,
  type HrmAccountRequest,
  type HrmWorker,
} from "@/lib/graphql/hrm-worker-provisioning";

const requestLabels: Record<string, string> = {
  pending: "Pengiriman ke IAM belum terkonfirmasi",
  verification_unavailable: "Pengajuan tersimpan. Status IAM belum dapat diperiksa.",
  submitted: "Menunggu proses IAM",
  applied: "Akun tertaut",
  rejected: "Pengajuan ditolak",
  failed: "Perlu ditindaklanjuti",
  cancelled: "Pengajuan dibatalkan",
};

const workerStatusLabels: Record<string, string> = {
  preboarding: "Persiapan kerja",
  active: "Aktif",
  leave_of_absence: "Cuti",
  suspended: "Diskors",
  notice_period: "Masa pemberitahuan",
  terminated: "Hubungan kerja berakhir",
  deceased: "Meninggal dunia",
};

const employmentDateFormat = new Intl.DateTimeFormat("id-ID", {
  dateStyle: "long",
  timeZone: "UTC",
});

function workerRequestError(error: unknown, fallback: string): string {
  if (isStepUpRequiredError(error)) {
    return "Lakukan verifikasi ulang untuk melanjutkan.";
  }
  if (isForbiddenError(error)) {
    return "Anda tidak memiliki izin untuk tindakan ini.";
  }
  return fallback;
}

export default function WorkersClient() {
  const { csrfToken, snapshot } = useIdentityWorkspace();
  const canSubmit = Boolean(snapshot?.permissions?.some(
    (grant) => grant.permission_code === "human_resource.access_requests.create" && grant.scope.type === "GLOBAL",
  ));
  const [numberInput, setNumberInput] = useState("");
  const [search, setSearch] = useState("");
  const [cursor, setCursor] = useState<{ after: string | null; before: string | null }>({ after: null, before: null });
  const [page, setPage] = useState<Awaited<ReturnType<typeof listProvisioningWorkers>> | null>(null);
  const [listError, setListError] = useState<string | null>(null);
  const [listRevision, setListRevision] = useState(0);
  const [selected, setSelected] = useState<HrmWorker | null>(null);
  const [accountRequest, setAccountRequest] = useState<HrmAccountRequest | null>(null);
  const [statusLoaded, setStatusLoaded] = useState(false);
  const [statusError, setStatusError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [revision, setRevision] = useState(0);

  useEffect(() => {
    let active = true;
    void listProvisioningWorkers(search, cursor.after, cursor.before).then(
      (result) => { if (active) setPage(result); },
      (err: unknown) => { if (active) setListError(workerRequestError(err, "Daftar pekerja gagal dimuat.")); },
    );
    return () => { active = false; };
  }, [search, cursor, listRevision]);

  useEffect(() => {
    if (!selected) return;
    let active = true;
    void fetchWorkerAccountRequest(selected.id).then(
      (result) => {
        if (active) {
          setAccountRequest(result);
          setStatusLoaded(true);
        }
      },
      (err: unknown) => { if (active) setStatusError(workerRequestError(err, "Status pengajuan gagal dimuat. Perbarui status untuk mencoba lagi.")); },
    );
    return () => { active = false; };
  }, [selected, revision]);

  function selectWorker(worker: HrmWorker) {
    setSelected(worker);
    setAccountRequest(null);
    setStatusLoaded(false);
    setStatusError(null);
  }

  function refreshStatus() {
    setStatusLoaded(false);
    setStatusError(null);
    setRevision((value) => value + 1);
  }

  function searchWorkers(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const next = numberInput.trim();
    if (next === search && !cursor.after && !cursor.before) return;
    setPage(null);
    setListError(null);
    setSearch(next);
    setCursor({ after: null, before: null });
  }

  async function submitRequest() {
    if (!selected || !csrfToken || submitting) return;
    setSubmitting(true);
    setStatusError(null);
    try {
      const { request, deliveryFailed } = await submitWorkerAccountRequest(selected.id, csrfToken);
      setAccountRequest(request);
      if (deliveryFailed) {
        setStatusError("Pengajuan tersimpan. Perbarui status atau kirim ulang ke IAM.");
      }
    } catch (err) {
      setStatusError(workerRequestError(err, "Pengajuan akun gagal dikirim. Periksa status sebelum mencoba lagi."));
      // Baca lagi karena HRM dapat menyimpan pengajuan sebelum pengiriman ke IAM gagal.
      try {
        setAccountRequest(await fetchWorkerAccountRequest(selected.id));
      } catch {
        setStatusLoaded(false);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="space-y-5">
      <div>
        <h1 className="text-2xl font-bold text-[#0b2439]">Pekerja</h1>
        <p className="mt-1 text-sm text-[#5a6672]">Cari pekerja dan kelola pengajuan akun kerja ke IAM.</p>
      </div>
      <form className="flex flex-wrap items-end gap-2" onSubmit={searchWorkers}>
        <label className="flex flex-col gap-1 text-xs font-semibold" htmlFor="worker-number-search">
          Nomor pekerja
          <input id="worker-number-search" className={styles.fieldInput} value={numberInput}
            onChange={(event) => setNumberInput(event.target.value)} />
        </label>
        <button type="submit" className={styles.btnSecondary}>Cari pekerja</button>
      </form>
      {listError ? (
        <div className="flex items-center gap-2">
          <p role="alert">{listError}</p>
          <button type="button" className={styles.btnSecondary} onClick={() => {
            setListError(null);
            setListRevision((value) => value + 1);
          }}>Muat ulang daftar</button>
        </div>
      ) : null}
      {!page && !listError ? <p role="status">Memuat daftar pekerja…</p> : null}
      {page && page.items.length === 0 ? <p>Tidak ada pekerja yang cocok.</p> : null}
      {page && page.items.length > 0 ? (
        <div className="overflow-x-auto rounded-xl border border-[#dde3e8]">
          <table className={styles.dataTable}>
            <thead><tr><th>Nomor pekerja</th><th>Nama pekerja</th><th>Badan hukum</th><th>Status kerja</th><th>Tanggal masuk</th><th>Aksi</th></tr></thead>
            <tbody>{page.items.map((worker) => (
              <tr key={worker.id}>
                <td>{worker.worker_number}</td>
                <td>{worker.worker_name || "Nama tidak tersedia"}</td>
                <td>{worker.legal_entity_name || "Nama badan hukum tidak tersedia"}</td>
                <td>{workerStatusLabels[worker.status] ?? "Status belum dikenal"}</td>
                <td>{worker.hire_date ? employmentDateFormat.format(new Date(worker.hire_date)) : "Belum tercatat"}</td>
                <td><button type="button" className={styles.btnSecondary} disabled={submitting}
                  onClick={() => selectWorker(worker)}>Lihat pengajuan</button></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : null}
      {page && (page.pageInfo.hasPreviousPage || page.pageInfo.hasNextPage) ? (
        <div className="flex gap-2">
          <button type="button" className={styles.btnSecondary} disabled={!page.pageInfo.hasPreviousPage || !page.pageInfo.startCursor}
            onClick={() => { setPage(null); setCursor({ before: page.pageInfo.startCursor ?? null, after: null }); }}>Sebelumnya</button>
          <button type="button" className={styles.btnSecondary} disabled={!page.pageInfo.hasNextPage || !page.pageInfo.endCursor}
            onClick={() => { setPage(null); setCursor({ after: page.pageInfo.endCursor ?? null, before: null }); }}>Berikutnya</button>
        </div>
      ) : null}
      {selected ? (
        <div className="space-y-3 rounded-xl border border-[#dde3e8] bg-white p-4">
          <h2 className="font-semibold text-[#0b2439]">{selected.worker_name || "Pengajuan akun pekerja"} · {selected.worker_number}</h2>
          {selected.legal_entity_name ? <p className="text-sm">{selected.legal_entity_name}</p> : null}
          {statusError ? <p role="alert" className="text-sm">{statusError}</p> : null}
          {!statusLoaded && !statusError ? <p role="status">Memeriksa status pengajuan…</p> : null}
          {statusLoaded ? (
            <p role="status" className="text-sm">
              {accountRequest ? requestLabels[accountRequest.status] ?? "Status perlu diperiksa" : "Belum ada pengajuan akun"}
              {accountRequest ? ` · ${new Date(accountRequest.requested_at).toLocaleString("id-ID")}` : ""}
            </p>
          ) : null}
          <div className="flex flex-wrap gap-2">
            {canSubmit && statusLoaded && selected.status === "active" &&
              (!accountRequest || accountRequest.status === "pending") ? (
              <button type="button" className={styles.btnPrimary} disabled={!csrfToken || submitting}
                onClick={() => void submitRequest()}>
                {submitting ? "Mengirim…" : accountRequest ? "Kirim ulang ke IAM" : "Ajukan akun ke IAM"}
              </button>
            ) : null}
            <button type="button" className={styles.btnSecondary} disabled={submitting}
              onClick={refreshStatus}>Perbarui status</button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
