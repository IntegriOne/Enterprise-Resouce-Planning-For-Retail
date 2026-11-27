"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import styles from "@/components/identity/identity.module.css";
import { AuthRequestError, executeGraphQL, unwrapMutationPayload, type MutationResult } from "@/lib/graphql/iam-browser-auth";
import CreateCandidateForm from "./create-candidate-form";
import type { PageInfo } from "@/lib/graphql/iam-types";
import { isForbiddenError, isStepUpRequiredError } from "@/lib/graphql/iam-dashboard-queries";

type Candidate = {
  id: string;
  candidate_number: string;
  preferred_name: string | null;
  candidate_status: string;
};

type CandidatePage = {
  total_count: number;
  pageInfo: PageInfo;
  nodes: Candidate[];
};

const pageSize = 20;

export default function CandidatesClient() {
  const { csrfToken, snapshot, openStepUpModal } = useIdentityWorkspace();
  const hasGlobalPermission = (code: string) => snapshot?.permissions.some((grant) =>
    grant.permission_code === code && grant.scope.type === "GLOBAL",
  ) ?? false;
  const canUpdate = hasGlobalPermission("human_resource.candidates.update");
  const canDelete = hasGlobalPermission("human_resource.candidates.delete");
  const canCreate = hasGlobalPermission("human_resource.candidates.create");
  const canReadParties = hasGlobalPermission("enterprise_foundation.parties.read");
  const [creating, setCreating] = useState(false);
  const [deleting, setDeleting] = useState<Candidate | null>(null);
  const [editing, setEditing] = useState<Candidate | null>(null);
  const [preferredName, setPreferredName] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveNotice, setSaveNotice] = useState<string | null>(null);
  const [numberInput, setNumberInput] = useState("");
  const [request, setRequest] = useState({ number: "", after: null as string | null, before: null as string | null });
  const [page, setPage] = useState<CandidatePage | null>(null);
  const [error, setError] = useState<"access" | "load" | null>(null);

  useEffect(() => {
    let active = true;
    void executeGraphQL<{ candidates: CandidatePage }>(
      `query RecruitmentCandidates($number: String, $first: Int, $last: Int, $after: String, $before: String) {
        candidates(candidate_number: $number, first: $first, last: $last, after: $after, before: $before) {
          total_count
          pageInfo { startCursor endCursor hasPreviousPage hasNextPage }
          nodes { id candidate_number preferred_name candidate_status }
        }
      }`,
      { number: request.number || null, first: request.before ? null : pageSize, last: request.before ? pageSize : null, after: request.after, before: request.before },
    ).then(
      ({ data }) => { if (active) setPage(data.candidates); },
      (failure: unknown) => { if (active) setError(isForbiddenError(failure) ? "access" : "load"); },
    );
    return () => { active = false; };
  }, [request]);

  function loadCandidates(number: string, after: string | null = null, before: string | null = null) {
    setPage(null);
    setEditing(null);
    setDeleting(null);
    setSaveNotice(null);
    setError(null);
    setRequest({ number, after, before });
  }

  function searchCandidates(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    loadCandidates(numberInput.trim());
  }

  function editCandidate(candidate: Candidate) {
    setDeleting(null);
    setEditing(candidate);
    setPreferredName(candidate.preferred_name ?? "");
    setSaveError(null);
    setSaveNotice(null);
  }

  async function saveCandidate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editing || !csrfToken || !canUpdate || saving) return;

    setSaving(true);
    setSaveError(null);
    try {
      const { data, response } = await executeGraphQL<{
        updateCandidate: MutationResult & { candidate: Candidate | null };
      }>(
        `mutation CorrectCandidateName($id: ID!, $name: String) {
          updateCandidate(id: $id, input: { preferred_name: $name }) {
            was_successful
            errors { type message }
            candidate { id candidate_number preferred_name candidate_status }
          }
        }`,
        { id: editing.id, name: preferredName.trim() || null },
        csrfToken,
      );
      const updated = unwrapMutationPayload(data.updateCandidate, response.status).candidate;
      if (!updated || updated.id !== editing.id) throw new Error("Invalid candidate response");
      setPage((current) => current && {
        ...current,
        nodes: current.nodes.map((candidate) => candidate.id === updated.id ? updated : candidate),
      });
      setEditing(null);
      setSaveNotice("Nama pilihan diperbarui.");
    } catch (failure) {
      setSaveError(isForbiddenError(failure)
        ? "Anda tidak memiliki izin mengubah kandidat."
        : "Nama pilihan gagal disimpan. Coba lagi.");
    } finally {
      setSaving(false);
    }
  }

  function confirmDeletion(candidate: Candidate) {
    setEditing(null);
    setDeleting(candidate);
    setSaveError(null);
    setSaveNotice(null);
  }

  async function deleteCandidate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!deleting || !csrfToken || !canDelete || saving) return;

    setSaving(true);
    setSaveError(null);
    try {
      const { data, response } = await executeGraphQL<{
        deleteCandidate: MutationResult & { deleted: boolean | null };
      }>(
        `mutation RemoveCandidate($id: ID!) {
          deleteCandidate(id: $id) { deleted was_successful errors { type message } }
        }`,
        { id: deleting.id },
        csrfToken,
      );
      if (!unwrapMutationPayload(data.deleteCandidate, response.status).deleted) {
        throw new Error("Invalid candidate deletion response");
      }
      if (page?.nodes.length === 1) {
        loadCandidates(request.number);
      } else {
        loadCandidates(request.number, request.after, request.before);
      }
      setSaveNotice("Kandidat dihapus.");
    } catch (failure) {
      if (isStepUpRequiredError(failure)) {
        setSaveError("Setelah verifikasi, pilih Hapus kandidat untuk melanjutkan.");
        openStepUpModal("Verifikasi diperlukan untuk menghapus kandidat.");
      } else if (failure instanceof AuthRequestError && failure.errorType === "CONFLICT") {
        setSaveError("Kandidat tidak dapat dihapus. Periksa status kandidat dan lamaran terkait.");
      } else {
        setSaveError(isForbiddenError(failure)
          ? "Anda tidak memiliki izin menghapus kandidat."
          : "Kandidat gagal dihapus. Coba lagi.");
      }
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-2xl font-bold text-[#0b2439]">Kandidat</h1>
        {canCreate && !creating ? (
          <button type="button" className={styles.btnPrimary} disabled={saving} onClick={() => {
            setEditing(null);
            setDeleting(null);
            setSaveNotice(null);
            setCreating(true);
          }}>Tambah kandidat</button>
        ) : null}
      </div>
      {canCreate && creating ? (
        <CreateCandidateForm csrfToken={csrfToken} canReadParties={canReadParties}
          onCancel={() => setCreating(false)} onCreated={(number) => {
            setCreating(false);
            setNumberInput(number);
            loadCandidates(number);
            setSaveNotice("Kandidat ditambahkan.");
          }} />
      ) : null}
      <form className="flex flex-wrap items-end gap-2" onSubmit={searchCandidates}>
        <label className="flex flex-col gap-1 text-xs font-semibold">
          Nomor kandidat
          <input type="search" className={styles.fieldInput} value={numberInput} maxLength={120}
            onChange={(event) => setNumberInput(event.target.value)} />
        </label>
        <button type="submit" className={styles.btnSecondary} disabled={saving}>Cari kandidat</button>
      </form>
      {!page && !error ? <p role="status">Memuat daftar kandidat…</p> : null}
      {error === "access" ? <p role="alert">Anda tidak memiliki izin membaca daftar kandidat.</p> : null}
      {error === "load" ? (
        <div className="flex items-center gap-2">
          <p role="alert">Daftar kandidat gagal dimuat.</p>
          <button type="button" className={styles.btnSecondary}
            onClick={() => loadCandidates(request.number, request.after, request.before)}>Coba lagi</button>
        </div>
      ) : null}
      {page && page.nodes.length === 0 ? <p>Tidak ada kandidat yang cocok.</p> : null}
      {page && page.nodes.length > 0 ? (
        <div className="overflow-x-auto rounded-xl border border-[#dde3e8]">
          <table className={styles.dataTable}>
            <thead><tr><th>Nomor kandidat</th><th>Nama pilihan</th><th>Status</th>{canUpdate || canDelete ? <th>Aksi</th> : null}</tr></thead>
            <tbody>{page.nodes.map((candidate) => (
              <tr key={candidate.id}>
                <td>{candidate.candidate_number}</td>
                <td>{candidate.preferred_name || "Belum tercatat"}</td>
                <td>{candidate.candidate_status === "active" ? "Aktif" : "Status belum dikenal"}</td>
                {canUpdate || canDelete ? (
                  <td>
                    {candidate.candidate_status === "active" ? (
                      <div className="flex flex-wrap gap-2">
                        {canUpdate ? <button type="button" className={styles.btnSecondary} disabled={saving || creating}
                          onClick={() => editCandidate(candidate)}>Ubah nama</button> : null}
                        {canDelete ? <button type="button" className={styles.btnSecondary} disabled={saving || creating}
                          onClick={() => confirmDeletion(candidate)}>Hapus</button> : null}
                      </div>
                    ) : null}
                  </td>
                ) : null}
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : null}
      {saveNotice ? <p role="status">{saveNotice}</p> : null}
      {canUpdate && editing ? (
        <form className="space-y-3 rounded-xl border border-[#dde3e8] bg-white p-4" onSubmit={(event) => void saveCandidate(event)}>
          <h2 className="font-semibold">Koreksi nama pilihan · {editing.candidate_number}</h2>
          <label className="flex flex-col gap-1 text-sm font-medium">
            Nama pilihan
            <input className={styles.fieldInput} value={preferredName} maxLength={180} disabled={saving}
              onChange={(event) => setPreferredName(event.target.value)} />
          </label>
          {saveError ? <p role="alert">{saveError}</p> : null}
          <div className="flex gap-2">
            <button type="submit" className={styles.btnPrimary} disabled={saving || !csrfToken}>
              {saving ? "Menyimpan…" : "Simpan nama"}
            </button>
            <button type="button" className={styles.btnSecondary} disabled={saving} onClick={() => setEditing(null)}>
              Batal
            </button>
          </div>
        </form>
      ) : null}
      {canDelete && deleting ? (
        <form className="space-y-3 rounded-xl border border-[#dde3e8] bg-white p-4" onSubmit={(event) => void deleteCandidate(event)}>
          <h2 className="font-semibold">Hapus kandidat {deleting.candidate_number}?</h2>
          <p>Kandidat yang sudah memiliki lamaran tidak dapat dihapus.</p>
          {saveError ? <p role="alert">{saveError}</p> : null}
          <div className="flex gap-2">
            <button type="submit" className={styles.btnPrimary} disabled={saving || !csrfToken}>
              {saving ? "Menghapus…" : "Hapus kandidat"}
            </button>
            <button type="button" className={styles.btnSecondary} disabled={saving} onClick={() => setDeleting(null)}>
              Batal
            </button>
          </div>
        </form>
      ) : null}
      {page && (page.pageInfo.hasPreviousPage || page.pageInfo.hasNextPage) ? (
        <nav aria-label="Halaman kandidat" className="flex gap-2">
          <button type="button" className={styles.btnSecondary} disabled={saving || (!page.pageInfo.hasPreviousPage || !page.pageInfo.startCursor)}
            onClick={() => loadCandidates(request.number, null, page.pageInfo.startCursor)}>Sebelumnya</button>
          <button type="button" className={styles.btnSecondary} disabled={saving || (!page.pageInfo.hasNextPage || !page.pageInfo.endCursor)}
            onClick={() => loadCandidates(request.number, page.pageInfo.endCursor)}>Berikutnya</button>
        </nav>
      ) : null}
    </section>
  );
}
