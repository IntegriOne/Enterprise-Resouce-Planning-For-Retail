"use client";

import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { fetchOrganizationUnits } from "@/lib/graphql/ef-structure-queries";
import { listProvisioningWorkers, type HrmWorker } from "@/lib/graphql/hrm-worker-provisioning";
import { executeGraphQL, unwrapMutationPayload, type MutationResult } from "@/lib/graphql/iam-browser-auth";
import { isForbiddenError } from "@/lib/graphql/iam-dashboard-queries";

import { employmentCategories, requisitionTypes, fetchPositionOptions, fetchUnitLegalEntity, fetchLocationOptions, fetchPlanOptions, fetchLineOptions,
  type Option, type OptionsPage, type ReferenceLabels } from "./requisition-references";

async function fetchJobs(after: string | null): Promise<OptionsPage> {
  const { data } = await executeGraphQL<{
    jobs: {
      nodes: { _id: string; job_code: string; job_title: string; is_active: boolean }[];
      pageInfo: { hasNextPage: boolean; endCursor: string | null };
    };
  }>(`query RequisitionJobOptions($after: String) {
    jobs(first: 20, after: $after) {
      nodes { _id job_code job_title is_active }
      pageInfo { hasNextPage endCursor }
    }
  }`, { after });
  const { nodes, pageInfo } = data.jobs;
  return {
    items: nodes.filter((job) => job.is_active).map((job) => ({
      id: job._id,
      label: `${job.job_code} · ${job.job_title}`,
    })),
    next: pageInfo.hasNextPage ? pageInfo.endCursor ?? null : null,
  };
}

async function fetchUnits(after: string | null): Promise<OptionsPage> {
  const { nodes, pageInfo } = await fetchOrganizationUnits({ active_only: true, after });
  return {
    items: nodes.map((unit) => ({ id: unit._id, label: `${unit.unit_code} · ${unit.name}` })),
    next: pageInfo.hasNextPage ? pageInfo.endCursor ?? null : null,
  };
}

function ReferenceSelect({ name, label, fetchOptions, current, optional, onChange }: {
  name: string;
  label: string;
  current?: Option;
  optional?: boolean;
  onChange?: (id: string) => void;
  fetchOptions: (after: string | null) => Promise<OptionsPage>;
}) {
  const [options, setOptions] = useState<OptionsPage>({ items: [], next: null });
  const [request, setRequest] = useState<{ after: string | null }>({ after: null });
  const [finished, setFinished] = useState<typeof request | null>(null);
  const [failed, setFailed] = useState(false);
  const pending = finished !== request;

  useEffect(() => {
    let active = true;
    void fetchOptions(request.after).then((result) => {
      if (!active) return;
      setOptions((previous) => ({
        items: request.after ? [...previous.items, ...result.items] : result.items,
        next: result.next,
      }));
      setFailed(false);
      setFinished(request);
    }, () => {
      if (!active) return;
      setFailed(true);
      setFinished(request);
    });
    return () => { active = false; };
  }, [fetchOptions, request]);

  return (
    <div className="space-y-2">
      <label className="flex flex-col gap-1">{label}
        <select name={name} className={styles.fieldSelect} defaultValue={current?.id ?? ""} required={!optional} onChange={(event) => onChange?.(event.target.value)}>
          <option value="">{optional ? "Tanpa" : "Pilih"} {label.toLowerCase()}</option>
          {current ? <option value={current.id}>{current.label}</option> : null}
          {options.items.filter((item) => item.id !== current?.id).map((item) => <option key={item.id} value={item.id}>{item.label}</option>)}
        </select>
      </label>
      {pending ? <p role="status">Memuat pilihan…</p> : null}
      {!pending && failed ? (
        <div className="flex flex-wrap items-center gap-2">
          <p role="alert">Pilihan gagal dimuat. Periksa akses Anda.</p>
          <button type="button" className={styles.btnSecondary} onClick={() => setRequest({ ...request })}>Coba lagi</button>
        </div>
      ) : null}
      {!pending && !failed && options.next ? (
        <button type="button" className={styles.btnSecondary} onClick={() => setRequest({ after: options.next })}>
          Muat {label.toLowerCase()} berikutnya
        </button>
      ) : null}
      {!pending && !failed && options.items.length === 0 && !options.next ? (
        <p>Tidak ada pilihan aktif.</p>
      ) : null}
    </div>
  );
}

function retainedOption(id: string | null | undefined, option: Option | undefined, fallback: string) {
  return id ? option ?? { id, label: fallback } : undefined;
}

export type RequisitionDraft = {
  id: string;
  requisition_number: string;
  position_id: string | null;
  work_location_reference_id: string | null;
  workforce_plan_line_id: string | null;
  hiring_manager_worker_id: string | null;
  job_id: string;
  organization_unit_reference_id: string;
  requisition_type: string;
  employment_category: string;
  headcount: number;
  target_start_date: string | null;
};

type SaveResult = MutationResult & { requisition: { requisition_number: string } | null };

export default function RequisitionForm({ csrfToken, onSaved, onCancel, draft }: {
  csrfToken: string | null;
  onSaved: (number: string) => void;
  draft?: { value: RequisitionDraft; job: string; unit: string; manager: string; references: ReferenceLabels };
  onCancel: () => void;
}) {
  const [workerNumber, setWorkerNumber] = useState("");
  const [manager, setManager] = useState<HrmWorker | null>(null);
  const [lookingUp, setLookingUp] = useState(false);
  const [lookupError, setLookupError] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const current = draft?.value;
  const references = draft?.references;
  const initialPlan = references?.plan?.id ?? "";
  const [plan, setPlan] = useState(initialPlan);
  const [job, setJob] = useState(current?.job_id ?? "");
  const [unit, setUnit] = useState(current?.organization_unit_reference_id ?? "");
  const [category, setCategory] = useState(current?.employment_category ?? "employee");
  const [unitScope, setUnitScope] = useState<{ unit: string; legalEntity: string | null } | null>(null);
  const legalEntity = unitScope?.unit === unit ? unitScope.legalEntity : null;
  const scopePending = unit !== "" && unitScope?.unit !== unit;
  const managerMismatch = !!manager && !!legalEntity &&
    manager.primary_legal_entity_reference_id !== legalEntity;
  const loadLocations = useCallback((after: string | null) => fetchLocationOptions(legalEntity, after), [legalEntity]);
  const loadPlans = useCallback((after: string | null) => fetchPlanOptions(legalEntity, after), [legalEntity]);
  const loadPositions = useCallback((after: string | null) => fetchPositionOptions({ job, unit }, after), [job, unit]);
  const loadLines = useCallback((after: string | null) => fetchLineOptions(plan, { job, unit, category }, after),
    [plan, job, unit, category]);
  const needsManager = !manager && (!draft || workerNumber.trim() !== "");
  const attempt = useRef<{ fields: string; key: string } | null>(null);

  useEffect(() => {
    if (!unit) return;
    let active = true;
    void fetchUnitLegalEntity(unit).then(
      (resolved) => { if (active) setUnitScope({ unit, legalEntity: resolved }); },
      () => { if (active) setUnitScope({ unit, legalEntity: null }); },
    );
    return () => { active = false; };
  }, [unit]);

  async function findManager() {
    const number = workerNumber.trim();
    if (!number || lookingUp || saving) return;
    setLookingUp(true);
    setLookupError("");
    setManager(null);
    try {
      const { items } = await listProvisioningWorkers(number, null, null);
      const worker = items.find((item) => item.status === "active");
      if (worker) setManager(worker);
      else setLookupError("Pekerja aktif dengan nomor itu tidak ditemukan.");
    } catch {
      setLookupError("Data pekerja gagal dimuat. Periksa nomor pekerja dan izin Anda.");
    } finally {
      setLookingUp(false);
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (needsManager || managerMismatch || scopePending || !csrfToken || lookingUp || saving) return;
    const fields = new FormData(event.currentTarget);
    const number = current?.requisition_number ?? String(fields.get("requisition_number") ?? "").trim();
    const input = {
      job_id: fields.get("job_id"),
      organization_unit_reference_id: fields.get("organization_unit_reference_id"),
      hiring_manager_worker_id: manager?.id,
      requisition_type: fields.get("requisition_type"),
      employment_category: fields.get("employment_category"),
      headcount: Number(fields.get("headcount")),
      target_start_date: fields.get("target_start_date") || null,
      position_id: fields.get("position_id") || null,
      work_location_reference_id: fields.get("work_location_reference_id") || null,
      workforce_plan_line_id: fields.get("workforce_plan_line_id") || null,
    };
    if (!number) {
      setSaveError("Isi nomor requisition.");
      return;
    }
    const saveFields = current
      ? Object.fromEntries(Object.entries(input).filter(([key, value]) =>
        value !== undefined && value !== current[key as keyof RequisitionDraft]))
      : { ...input, requisition_number: number };
    if (Object.keys(saveFields).length === 0) {
      setSaveError("Belum ada perubahan untuk disimpan.");
      return;
    }
    setSaving(true);
    setSaveError("");
    try {
      let result: SaveResult;
      if (current) {
        const { data, response } = await executeGraphQL<{ updateRequisition: SaveResult }>(
          `mutation EditRequisitionDraft($id: ID!, $input: UpdateRequisitionInput!) {
            updateRequisition(id: $id, input: $input) {
              was_successful errors { type message } requisition { requisition_number }
            }
          }`, { id: current.id, input: saveFields }, csrfToken);
        result = unwrapMutationPayload(data.updateRequisition, response.status);
      } else {
        const signature = JSON.stringify(saveFields);
        if (attempt.current?.fields !== signature) {
          attempt.current = { fields: signature, key: crypto.randomUUID() };
        }
        const { data, response } = await executeGraphQL<{ createRequisition: SaveResult }>(
          `mutation SaveRequisitionDraft($input: CreateRequisitionInput!) {
            createRequisition(input: $input) {
              was_successful errors { type message } requisition { requisition_number }
            }
          }`, { input: { ...saveFields, idempotency_key: attempt.current.key } }, csrfToken);
        result = unwrapMutationPayload(data.createRequisition, response.status);
      }
      if (!result.requisition) throw new Error("Invalid requisition response");
      onSaved(result.requisition.requisition_number);
    } catch (failure) {
      setSaveError(isForbiddenError(failure)
        ? "Anda tidak memiliki izin menyimpan requisition."
        : "Draf gagal disimpan. Periksa nomor, jabatan, unit, hiring manager, dan tanggal target.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="space-y-4 rounded-xl border border-[#dde3e8] bg-white p-4" onSubmit={(event) => void submit(event)}>
      <h2 className="font-semibold">{current ? "Edit draf requisition" : "Buat draf requisition"}</h2>
      <fieldset disabled={saving} className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1">Nomor requisition
          <input name="requisition_number" className={styles.fieldInput} maxLength={100} defaultValue={current?.requisition_number} readOnly={Boolean(current)} required />
        </label>
        <label className="flex flex-col gap-1">Jenis kebutuhan
          <select name="requisition_type" className={styles.fieldSelect} defaultValue={current?.requisition_type ?? "new_headcount"}>
            {Object.entries(requisitionTypes).map(([key, label]) => <option key={key} value={key}>{label}</option>)}
          </select>
        </label>
        <ReferenceSelect name="job_id" label="Jabatan" fetchOptions={fetchJobs} onChange={setJob} current={draft && { id: draft.value.job_id, label: draft.job }} />
        <ReferenceSelect name="organization_unit_reference_id" label="Unit organisasi" fetchOptions={fetchUnits} onChange={(id) => {
          setUnit(id);
          setPlan(initialPlan);
        }} current={draft && { id: draft.value.organization_unit_reference_id, label: draft.unit }} />
        <label className="flex flex-col gap-1">Kategori pekerja
          <select name="employment_category" className={styles.fieldSelect} value={category} onChange={(event) => setCategory(event.target.value)}>
            {Object.entries(employmentCategories).map(([key, label]) => <option key={key} value={key}>{label}</option>)}
          </select>
        </label>
        <label className="flex flex-col gap-1">Jumlah kebutuhan
          <input name="headcount" type="number" min={1} max={2147483647} defaultValue={current?.headcount ?? 1} className={styles.fieldInput} required />
        </label>
        <label className="flex flex-col gap-1">Tanggal target mulai
          <input name="target_start_date" type="date" defaultValue={current?.target_start_date ?? ""} className={styles.fieldInput} />
        </label>
      </fieldset>
      <fieldset disabled={saving} className="grid gap-3 sm:grid-cols-2">
        <legend className="font-medium">Referensi opsional</legend>
        <p className="text-sm sm:col-span-2">
          Pilih jabatan dan unit untuk melihat posisi, lokasi kerja, dan lini yang sesuai. Periksa kembali pilihan referensi bila rincian berubah.
        </p>
        <ReferenceSelect key={`position-${job}-${unit}`} name="position_id" label="Posisi" optional fetchOptions={loadPositions}
          current={retainedOption(current?.position_id, references?.position, "Posisi saat ini")} />
        <ReferenceSelect key={`location-${unit}`} name="work_location_reference_id" label="Lokasi kerja" optional fetchOptions={loadLocations}
          current={retainedOption(current?.work_location_reference_id, references?.location, "Lokasi saat ini")} />
        <ReferenceSelect key={`plan-${unit}`} name="workforce_plan" label="Rencana tenaga kerja" optional fetchOptions={loadPlans}
          current={references?.plan} onChange={setPlan} />
        {plan ? (
          <ReferenceSelect key={`line-${plan}-${job}-${unit}-${category}`} name="workforce_plan_line_id" label="Lini rencana" optional fetchOptions={loadLines}
            current={plan === initialPlan ? references?.line : undefined} />
        ) : current?.workforce_plan_line_id && !initialPlan ? (
          <label className="flex flex-col gap-1">Lini rencana
            <select name="workforce_plan_line_id" className={styles.fieldSelect} defaultValue={current.workforce_plan_line_id}>
              <option value="">Tanpa lini rencana</option>
              <option value={current.workforce_plan_line_id}>Lini saat ini</option>
            </select>
          </label>
        ) : null}
      </fieldset>
      <fieldset disabled={lookingUp || saving} className="space-y-2">
        <legend className="font-medium">Hiring manager</legend>
        <div className="flex flex-wrap items-end gap-2">
          <label className="flex flex-col gap-1">Nomor pekerja
            <input value={workerNumber} className={styles.fieldInput} onChange={(event) => {
              setWorkerNumber(event.target.value);
              setManager(null);
              setLookupError("");
            }} />
          </label>
          <button type="button" className={styles.btnSecondary} disabled={!workerNumber.trim()}
            onClick={() => void findManager()}>Cari pekerja</button>
        </div>
        {draft && !manager ? <p>Hiring manager saat ini · {draft.manager}</p> : null}
        {lookingUp ? <p role="status">Mencari pekerja…</p> : null}
        {manager ? <p role="status">{manager.worker_name ?? "Pekerja"} · {manager.worker_number}</p> : null}
        {lookupError ? <p role="alert">{lookupError}</p> : null}
        {managerMismatch ? <p role="alert">Badan hukum hiring manager berbeda dari unit organisasi yang dipilih.</p> : null}
      </fieldset>
      {saveError ? <p role="alert">{saveError}</p> : null}
      <div className="flex gap-2">
        <button type="submit" className={styles.btnPrimary} disabled={!csrfToken || needsManager || managerMismatch || scopePending || lookingUp || saving}>
          {saving ? "Menyimpan…" : "Simpan draf"}
        </button>
        <button type="button" className={styles.btnSecondary} disabled={saving || lookingUp} onClick={onCancel}>Batal</button>
      </div>
    </form>
  );
}
