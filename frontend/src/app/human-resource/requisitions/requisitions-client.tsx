"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { employmentCategories, requisitionTypes, fetchReferenceLabels, type ReferenceLabels } from "./requisition-references";
import RequisitionForm, { type RequisitionDraft } from "./requisition-form";
import { RelayPaginationBar, SlideOverDrawer } from "@/components/identity/identity-primitives";
import styles from "@/components/identity/identity.module.css";
import { AuthRequestError, executeGraphQL, unwrapMutationPayload, type MutationResult } from "@/lib/graphql/iam-browser-auth";
import { isForbiddenError, isStepUpRequiredError } from "@/lib/graphql/iam-dashboard-queries";
import { fetchWorkerDisplayLabel } from "@/lib/graphql/hrm-worker-provisioning";
import type { PageInfo } from "@/lib/graphql/iam-types";

type Requisition = {
  id: string;
  requisition_number: string;
  job_id: string;
  organization_unit_reference_id: string;
  status: string;
  headcount: number;
  target_start_date: string | null;
};

type RequisitionConnection = {
  nodes: Requisition[];
  total_count: number;
  pageInfo: PageInfo;
};

type RequisitionPage = RequisitionConnection & {
  jobLabels: Map<string, string>;
  unitLabels: Map<string, string>;
};

type ListRequest = { number: string; after: string | null; before: string | null };
type RequisitionDetail = RequisitionDraft & { status: string; rejection_reason_code: string | null };
type Detail = { id: string; value: RequisitionDetail | null; manager: string; references: ReferenceLabels };

const requisitionActions = {
  submit: {
    label: "Ajukan draf",
    resultStatus: "submitted",
    description: "Setelah diajukan, draf masuk tahap persetujuan dan tidak dapat diedit.",
    success: "Requisition diajukan.",
    conflict: "Requisition tidak dapat diajukan. Periksa rincian dan pastikan Anda adalah pembuat draf.",
    query: `mutation SubmitRecruitmentDraft($id: ID!) {
      result: submitRequisition(id: $id) {
        was_successful errors { type message } requisition { status }
      }
    }`,
  },
  reject: {
    label: "Tolak pengajuan",
    resultStatus: "rejected",
    description: "Penolakan akan dicatat bersama kode alasan.",
    success: "Requisition ditolak.",
    conflict: "Pengajuan tidak dapat ditolak. Periksa status terbaru requisition.",
    query: `mutation RejectRecruitmentRequest($id: ID!, $reason: String!) {
      result: rejectRequisition(id: $id, reason_code: $reason) {
        was_successful errors { type message } requisition { status }
      }
    }`,
  },
  approve: {
    label: "Setujui requisition",
    resultStatus: "approved",
    description: "Mandat unit dan otorisasi anggaran diperiksa sebelum persetujuan dicatat.",
    success: "Requisition disetujui.",
    conflict: "Requisition belum dapat disetujui. Periksa status dan bukti anggaran.",
    query: `mutation ApproveRecruitmentRequest($id: ID!) {
      result: approveRequisition(id: $id) {
        was_successful errors { type message } requisition { status }
      }
    }`,
  },
  open: {
    label: "Buka requisition",
    resultStatus: "open",
    description: "Mandat peninjau dan otorisasi anggaran diperiksa ulang sebelum requisition dibuka.",
    success: "Requisition dibuka.",
    conflict: "Requisition belum dapat dibuka. Periksa status dan bukti anggaran.",
    query: `mutation OpenRecruitmentRequest($id: ID!) {
      result: openRequisition(id: $id) {
        was_successful errors { type message } requisition { status }
      }
    }`,
  },
  close: {
    label: "Tutup requisition",
    resultStatus: "closed",
    description: "Requisition akan ditutup setelah seluruh proses rekrutmen aktif selesai.",
    success: "Requisition ditutup.",
    conflict: "Requisition belum dapat ditutup. Periksa apakah masih ada posting atau lamaran aktif.",
    query: `mutation CloseRecruitmentRequest($id: ID!) {
      result: closeRequisition(id: $id) {
        was_successful errors { type message } requisition { status }
      }
    }`,
  },
  delete: {
    label: "Hapus draf",
    resultStatus: null,
    description: "Draf akan dihapus dari daftar requisition.",
    success: "Draf requisition dihapus.",
    conflict: "Draf tidak dapat dihapus. Periksa status dan data yang terkait.",
    query: `mutation DeleteRecruitmentDraft($id: ID!) {
      result: deleteRequisition(id: $id) { was_successful errors { type message } deleted }
    }`,
  },
} as const;
type RequisitionAction = keyof typeof requisitionActions;
type RequisitionActionResult = MutationResult & { deleted?: boolean | null; requisition?: { status: string } | null };

const pageSize = 20;
const statusLabels: Record<string, string> = {
  draft: "Draf",
  submitted: "Menunggu persetujuan",
  approved: "Disetujui",
  rejected: "Ditolak",
  open: "Dibuka",
  closed: "Ditutup",
};
const targetDateFormat = new Intl.DateTimeFormat("id-ID", { dateStyle: "long", timeZone: "UTC" });

async function fetchRequisitions(request: ListRequest): Promise<RequisitionPage> {
  const { data } = await executeGraphQL<{ requisitions: RequisitionConnection }>(
    `query RecruitmentRequisitions($number: String, $after: String, $before: String, $first: Int, $last: Int) {
      requisitions(requisition_number: $number, after: $after, before: $before, first: $first, last: $last) {
        total_count pageInfo { startCursor endCursor hasPreviousPage hasNextPage }
        nodes { id requisition_number job_id organization_unit_reference_id status headcount target_start_date }
      }
    }`,
    {
      number: request.number || null, after: request.after, before: request.before,
      first: request.before ? null : pageSize, last: request.before ? pageSize : null,
    },
  );
  const page = data.requisitions;
  const jobLabels = new Map<string, string>();
  const unitLabels = new Map<string, string>();
  if (page.nodes.length === 0) return { ...page, jobLabels, unitLabels };

  const jobIDs = [...new Set(page.nodes.map((row) => row.job_id))];
  const unitIDs = [...new Set(page.nodes.map((row) => row.organization_unit_reference_id))];
  const [jobs, units] = await Promise.allSettled([
    executeGraphQL<{ jobsByIds: { _id: string; job_code: string; job_title: string }[] }>(
      "query RequisitionJobs($ids: [ID!]!) { jobsByIds(ids: $ids) { _id job_code job_title } }",
      { ids: jobIDs },
    ),
    executeGraphQL<{ organizationUnitsByIds: { _id: string; unit_code: string; name: string }[] }>(
      "query RequisitionUnits($ids: [ID!]!) { organizationUnitsByIds(ids: $ids) { _id unit_code name } }",
      { ids: unitIDs },
    ),
  ]);
  if (jobs.status === "fulfilled") {
    for (const job of jobs.value.data.jobsByIds) jobLabels.set(job._id, `${job.job_code} · ${job.job_title}`);
  }
  if (units.status === "fulfilled") {
    for (const unit of units.value.data.organizationUnitsByIds) unitLabels.set(unit._id, `${unit.unit_code} · ${unit.name}`);
  }
  return { ...page, jobLabels, unitLabels };
}

async function fetchDetail(id: string): Promise<Detail> {
  const { data } = await executeGraphQL<{
    requisition: RequisitionDetail | null;
  }>(
    `query RequisitionHiringManager($id: ID!) {
      requisition(id: $id) {
        id requisition_number status job_id organization_unit_reference_id hiring_manager_worker_id
        requisition_type employment_category headcount target_start_date
        position_id work_location_reference_id workforce_plan_line_id rejection_reason_code
      }
    }`, { id });
  const workerID = data.requisition?.hiring_manager_worker_id;
  const [manager, references] = await Promise.all([
    workerID ? fetchWorkerDisplayLabel(workerID).catch(() => null) : null,
    data.requisition ? fetchReferenceLabels(data.requisition) : {},
  ]);
  return { id, references, value: data.requisition, manager: manager ?? "Data hiring manager tidak tersedia" };
}

function formatTargetDate(value: string | null) {
  if (!value) return "Belum ditentukan";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "Belum ditentukan" : targetDateFormat.format(date);
}

export default function RequisitionsClient() {
  const { csrfToken, snapshot, openStepUpModal } = useIdentityWorkspace();
  const hasPermission = (action: string) => snapshot?.permissions.some((grant) =>
    grant.permission_code === `human_resource.requisitions.${action}` && grant.scope.type === "GLOBAL",
  ) ?? false;
  const canCreate = hasPermission("create");
  const canEdit = hasPermission("update");
  const canSubmit = hasPermission("submit");
  const canDelete = hasPermission("delete");
  const canReject = hasPermission("reject");
  const canClose = hasPermission("close");
  const canOpen = hasPermission("open");
  const [rejectionReason, setRejectionReason] = useState("");
  const [action, setAction] = useState<RequisitionAction | null>(null);
  const [actionError, setActionError] = useState("");
  const [notice, setNotice] = useState("");
  const [acting, setActing] = useState(false);
  const [editing, setEditing] = useState(false);
  const [creating, setCreating] = useState(false);
  const [input, setInput] = useState("");
  const [request, setRequest] = useState<ListRequest>({ number: "", after: null, before: null });
  const [page, setPage] = useState<RequisitionPage | null>(null);
  const [error, setError] = useState<"access" | "load" | null>(null);
  const [selected, setSelected] = useState<Requisition | null>(null);
  const [detail, setDetail] = useState<Detail | null>(null);
  const selectedID = selected?.id;
  const selectedDetail = detail && detail.id === selectedID ? detail : null;
  const details = selectedDetail?.value;
  const draft = details?.status === "draft" ? details : null;
  const canApprove = !!details && (snapshot?.permissions.some((grant) =>
    grant.permission_code === "human_resource.requisitions.approve" &&
    grant.scope.type === "ORGANIZATION_UNIT" &&
    grant.scope.id?.replace(/^gid:\/\/integrione\/organization-unit\//, "").toLowerCase() ===
      details.organization_unit_reference_id.toLowerCase(),
  ) ?? false);

  useEffect(() => {
    let active = true;
    void fetchRequisitions(request).then(
      (result) => { if (active) setPage(result); },
      (failure: unknown) => { if (active) setError(isForbiddenError(failure) ? "access" : "load"); },
    );
    return () => { active = false; };
  }, [request]);

  useEffect(() => {
    if (!selectedID) return;
    let active = true;
    void fetchDetail(selectedID).then(
      (result) => { if (active) setDetail(result); },
      () => { if (active) setDetail({ id: selectedID, manager: "Data hiring manager tidak tersedia", value: null, references: {} }); },
    );
    return () => { active = false; };
  }, [selectedID]);

  function load(next: ListRequest) {
    setPage(null);
    setError(null);
    setSelected(null);
    setEditing(false);
    setAction(null);
    setNotice("");
    setRequest(next);
  }

  async function runRequisitionAction() {
    if (!action || !selected || !csrfToken || acting || !(action === "approve" ? canApprove : hasPermission(action))) return;
    const command = requisitionActions[action];
    const reason = rejectionReason.trim();
    if (action === "reject" && (!reason || new TextEncoder().encode(reason).length > 80)) {
      setActionError("Isi kode alasan penolakan dengan maksimal 80 byte.");
      return;
    }
    setActing(true);
    setActionError("");
    try {
      const { data, response } = await executeGraphQL<{ result: RequisitionActionResult }>(
        command.query, action === "reject" ? { id: selected.id, reason } : { id: selected.id }, csrfToken,
      );
      const result = unwrapMutationPayload(data.result, response.status);
      if (command.resultStatus === null ? !result.deleted : result.requisition?.status !== command.resultStatus) {
        throw new Error("Invalid requisition action response");
      }
      load({ number: request.number, after: null, before: null });
      setNotice(command.success);
    } catch (failure) {
      if (isStepUpRequiredError(failure)) {
        setActionError(`Setelah verifikasi, pilih ${command.label} untuk melanjutkan.`);
        openStepUpModal(`Verifikasi diperlukan untuk tindakan ${command.label.toLowerCase()}.`);
      } else if ((action === "approve" || action === "open") &&
        failure instanceof AuthRequestError && failure.errorType === "UNAVAILABLE") {
        setActionError("Mandat atau bukti anggaran belum dapat diverifikasi. Coba lagi nanti.");
      } else if (failure instanceof AuthRequestError && failure.errorType === "CONFLICT") {
        setActionError(command.conflict);
      } else {
        setActionError(isForbiddenError(failure)
          ? action === "reject" || action === "approve"
            ? "Tindakan ini memerlukan peninjau berwenang yang berbeda dari pembuat dan pengaju."
            : "Anda tidak memiliki izin untuk tindakan ini."
          : "Tindakan gagal. Coba lagi.");
      }
    } finally {
      setActing(false);
    }
  }

  function search(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    load({ number: input.trim(), after: null, before: null });
  }

  function jobLabel(row: Pick<Requisition, "job_id">) {
    return page?.jobLabels.get(row.job_id) ?? "Jabatan tidak tersedia";
  }

  function unitLabel(row: Pick<Requisition, "organization_unit_reference_id">) {
    return page?.unitLabels.get(row.organization_unit_reference_id) ?? "Unit tidak tersedia";
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-xl font-semibold">Requisition</h1>
        {canCreate && !creating ? (
          <button type="button" className={styles.btnPrimary} onClick={() => setCreating(true)}>Buat draf</button>
        ) : null}
      </div>
      {canCreate && creating ? (
        <RequisitionForm csrfToken={csrfToken} onCancel={() => setCreating(false)} onSaved={(number) => {
          setCreating(false);
          setInput(number);
          load({ number, after: null, before: null });
        }} />
      ) : null}
      <form className="flex flex-wrap items-end gap-2" onSubmit={search}>
        <label className="flex flex-col gap-1 text-xs font-semibold">
          Nomor requisition
          <input type="search" className={styles.fieldInput} value={input} maxLength={100}
            onChange={(event) => setInput(event.target.value)} />
        </label>
        <button type="submit" className={styles.btnSecondary}>Cari requisition</button>
      </form>
      {notice ? <p role="status">{notice}</p> : null}
      {!page && !error ? <p role="status">Memuat daftar requisition…</p> : null}
      {error === "access" ? <p role="alert">Anda tidak memiliki izin membaca daftar requisition.</p> : null}
      {error === "load" ? (
        <div className="flex items-center gap-2">
          <p role="alert">Daftar requisition gagal dimuat.</p>
          <button type="button" className={styles.btnSecondary} onClick={() => load({ ...request })}>Coba lagi</button>
        </div>
      ) : null}
      {page ? (
        <div className="overflow-x-auto rounded-xl border border-[#dde3e8]">
          {page.nodes.length ? (
            <table className={styles.dataTable}>
              <thead>
                <tr><th>Nomor</th><th>Jabatan</th><th>Unit</th><th>Kebutuhan</th><th>Status</th><th>Target mulai</th></tr>
              </thead>
              <tbody>
                {page.nodes.map((row) => (
                  <tr key={row.id}>
                    <td>
                      <button type="button" className="font-semibold underline" onClick={() => {
                        setAction(null);
                        setEditing(false);
                        setDetail(null);
                        setSelected(row);
                      }}>{row.requisition_number}</button>
                    </td>
                    <td>{jobLabel(row)}</td>
                    <td>{unitLabel(row)}</td>
                    <td>{row.headcount}</td>
                    <td>{statusLabels[row.status] ?? "Status tidak tersedia"}</td>
                    <td>{formatTargetDate(row.target_start_date)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : <p className="p-4">Tidak ada requisition yang cocok.</p>}
          {page.total_count > 0 ? (
            <RelayPaginationBar totalCount={page.total_count} currentCount={page.nodes.length} pageInfo={page.pageInfo}
              isLoading={false}
              onPrev={() => {
                if (page.pageInfo.startCursor) load({ ...request, after: null, before: page.pageInfo.startCursor });
              }}
              onNext={() => {
                if (page.pageInfo.endCursor) load({ ...request, after: page.pageInfo.endCursor, before: null });
              }} />
          ) : null}
        </div>
      ) : null}
      <SlideOverDrawer open={Boolean(selected)} title={selected?.requisition_number ?? "Detail requisition"}
        onClose={() => { if (!acting) setSelected(null); }}>
        {selectedDetail && draft && canEdit && editing ? (
          <RequisitionForm csrfToken={csrfToken}
            draft={{ value: draft, job: jobLabel(draft), unit: unitLabel(draft), manager: selectedDetail.manager, references: selectedDetail.references }}
            onCancel={() => setEditing(false)}
            onSaved={(number) => {
              setInput(number);
              load({ number, after: null, before: null });
            }} />
        ) : selected ? (
          <div className="space-y-4">
            {draft || ((canApprove || canReject) && details?.status === "submitted") ||
              (canOpen && details?.status === "approved") || (canClose && details?.status === "open") ? (
              action ? (
                <form className="space-y-3 rounded-lg border border-[#dde3e8] p-4" onSubmit={(event) => {
                  event.preventDefault();
                  void runRequisitionAction();
                }}>
                  <h3 className="font-semibold">{requisitionActions[action].label} {selected.requisition_number}?</h3>
                  <p>{requisitionActions[action].description}</p>
                  {action === "reject" ? (
                    <label className="flex flex-col gap-1">Kode alasan penolakan
                      <input className={styles.fieldInput} value={rejectionReason} maxLength={80} required disabled={acting}
                        onChange={(event) => setRejectionReason(event.target.value)} />
                    </label>
                  ) : null}
                  {actionError ? <p role="alert">{actionError}</p> : null}
                  <div className="flex flex-wrap gap-2">
                    <button type="submit" className={styles.btnPrimary} disabled={!csrfToken || acting}>
                      {acting ? "Memproses…" : requisitionActions[action].label}
                    </button>
                    <button type="button" className={styles.btnSecondary} disabled={acting} onClick={() => setAction(null)}>Batal</button>
                  </div>
                </form>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {draft && canEdit ? <button type="button" className={styles.btnSecondary} onClick={() => setEditing(true)}>Edit draf</button> : null}
                  {draft && canSubmit ? <button type="button" className={styles.btnPrimary} onClick={() => {
                    setActionError("");
                    setAction("submit");
                  }}>Ajukan</button> : null}
                  {canApprove && details?.status === "submitted" ? (
                    <button type="button" className={styles.btnPrimary} onClick={() => {
                      setActionError("");
                      setAction("approve");
                    }}>Setujui requisition</button>
                  ) : null}
                  {canOpen && details?.status === "approved" ? (
                    <button type="button" className={styles.btnPrimary} onClick={() => {
                      setActionError("");
                      setAction("open");
                    }}>Buka requisition</button>
                  ) : null}
                  {canReject && details?.status === "submitted" ? (
                    <button type="button" className={styles.btnSecondary} onClick={() => {
                      setActionError("");
                      setRejectionReason("");
                      setAction("reject");
                    }}>Tolak</button>
                  ) : null}
                  {canClose && details?.status === "open" ? (
                    <button type="button" className={styles.btnSecondary} onClick={() => {
                      setActionError("");
                      setAction("close");
                    }}>Tutup requisition</button>
                  ) : null}
                  {draft && canDelete ? <button type="button" className={styles.btnSecondary} onClick={() => {
                    setActionError("");
                    setAction("delete");
                  }}>Hapus</button> : null}
                </div>
              )
            ) : null}
            <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div><dt className="font-semibold">Jabatan</dt><dd>{jobLabel(details ?? selected)}</dd></div>
              <div><dt className="font-semibold">Unit</dt><dd>{unitLabel(details ?? selected)}</dd></div>
              <div><dt className="font-semibold">Hiring manager</dt><dd role="status">
                {selectedDetail?.manager ?? "Memuat data hiring manager…"}
              </dd></div>
              <div><dt className="font-semibold">Status</dt><dd>{statusLabels[details?.status ?? selected.status] ?? "Status tidak tersedia"}</dd></div>
              <div><dt className="font-semibold">Jumlah kebutuhan</dt><dd>{details?.headcount ?? selected.headcount}</dd></div>
              <div><dt className="font-semibold">Target mulai</dt><dd>{formatTargetDate(details ? details.target_start_date : selected.target_start_date)}</dd></div>
              {details ? (
                <>
                  <div><dt className="font-semibold">Jenis kebutuhan</dt><dd>{requisitionTypes[details.requisition_type] ?? "Jenis belum dikenal"}</dd></div>
                  <div><dt className="font-semibold">Kategori pekerja</dt><dd>{employmentCategories[details.employment_category] ?? "Kategori belum dikenal"}</dd></div>
                  {details.position_id ? (
                    <div><dt className="font-semibold">Posisi</dt><dd>{selectedDetail?.references.position?.label ?? "Posisi tidak tersedia"}</dd></div>
                  ) : null}
                  {details.work_location_reference_id ? (
                    <div><dt className="font-semibold">Lokasi kerja</dt><dd>{selectedDetail?.references.location?.label ?? "Lokasi tidak tersedia"}</dd></div>
                  ) : null}
                  {details.workforce_plan_line_id ? (
                    <div><dt className="font-semibold">Lini rencana tenaga kerja</dt><dd>
                      {[selectedDetail?.references.plan?.label, selectedDetail?.references.line?.label].filter(Boolean).join(" · ") || "Lini tidak tersedia"}
                    </dd></div>
                  ) : null}
                  {details.rejection_reason_code ? (
                    <div><dt className="font-semibold">Kode alasan penolakan</dt><dd>{details.rejection_reason_code}</dd></div>
                  ) : null}
                </>
              ) : null}
            </dl>
          </div>
        ) : null}
      </SlideOverDrawer>
    </section>
  );
}
