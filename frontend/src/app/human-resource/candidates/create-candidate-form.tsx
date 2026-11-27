"use client";

import { useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { executeGraphQL, unwrapMutationPayload, type MutationResult } from "@/lib/graphql/iam-browser-auth";
import { isForbiddenError } from "@/lib/graphql/iam-dashboard-queries";

type Person = { _id: string; display_name: string; kind: string; status: string };
type Contact = { _id: string; full_name: string };
type ContactPage = { nodes: Contact[]; pageInfo: { hasNextPage: boolean; endCursor: string | null } };

export default function CreateCandidateForm({ csrfToken, canReadParties, onCreated, onCancel }: {
  csrfToken: string | null;
  canReadParties: boolean;
  onCreated: (number: string) => void;
  onCancel: () => void;
}) {
  const [partyCode, setPartyCode] = useState("");
  const [person, setPerson] = useState<Person | null>(null);
  const [contacts, setContacts] = useState<ContactPage | null>(null);
  const [contactID, setContactID] = useState("");
  const [lookupBusy, setLookupBusy] = useState(false);
  const [lookupError, setLookupError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  async function loadContacts(party: Person, after: string | null) {
    const { data } = await executeGraphQL<{ partyContacts: ContactPage }>(
      `query CandidateContacts($party: ID!, $after: String) {
        partyContacts(party_id: $party, active_only: true, first: 20, after: $after) {
          nodes { _id full_name }
          pageInfo { hasNextPage endCursor }
        }
      }`,
      { party: party._id, after },
    );
    setContacts((previous) => ({
      ...data.partyContacts,
      nodes: [...(after ? previous?.nodes ?? [] : []), ...data.partyContacts.nodes],
    }));
  }

  async function findPerson() {
    const code = partyCode.trim();
    if (!code || lookupBusy || saving) return;
    setLookupBusy(true);
    setLookupError(null);
    setPerson(null);
    setContacts(null);
    setContactID("");
    try {
      const { data } = await executeGraphQL<{ partyByCode: Person | null }>(
        `query CandidatePerson($code: String!) {
          partyByCode(partner_code: $code) { _id display_name kind status }
        }`,
        { code },
      );
      const party = data.partyByCode;
      if (!party || party.kind !== "INDIVIDUAL" || party.status !== "ACTIVE") {
        setLookupError("Individu aktif tidak ditemukan.");
        return;
      }
      setPerson(party);
      await loadContacts(party, null);
    } catch {
      setLookupError("Data individu atau kontak gagal dimuat. Periksa kode dan izin Anda.");
    } finally {
      setLookupBusy(false);
    }
  }

  async function loadMoreContacts() {
    if (!person || !contacts?.pageInfo.endCursor || lookupBusy || saving) return;
    setLookupBusy(true);
    setLookupError(null);
    try {
      await loadContacts(person, contacts.pageInfo.endCursor);
    } catch {
      setLookupError("Kontak berikutnya gagal dimuat. Coba lagi.");
    } finally {
      setLookupBusy(false);
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!csrfToken || saving || lookupBusy) return;
    const fields = new FormData(event.currentTarget);
    const number = String(fields.get("candidate_number") ?? "").trim();
    const privacyVersion = String(fields.get("privacy_notice_version") ?? "").trim();
    const acknowledgedAt = new Date(String(fields.get("notice_acknowledged_at") ?? ""));
    if (!number || !privacyVersion || Number.isNaN(acknowledgedAt.getTime())) {
      setSaveError("Lengkapi nomor kandidat, versi pemberitahuan, dan waktu pengakuan.");
      return;
    }
    setSaving(true);
    setSaveError(null);
    try {
      const { data, response } = await executeGraphQL<{
        createCandidate: MutationResult & { candidate: { candidate_number: string } | null };
      }>(
        `mutation RegisterCandidate($input: CreateCandidateInput!) {
          createCandidate(input: $input) {
            was_successful
            errors { type message }
            candidate { candidate_number }
          }
        }`,
        { input: {
          candidate_number: number,
          preferred_name: String(fields.get("preferred_name") ?? "").trim() || null,
          privacy_notice_version: privacyVersion,
          notice_acknowledged_at: acknowledgedAt.toISOString(),
          data_retention_until: fields.get("data_retention_until"),
          person_party_reference_id: person?._id ?? null,
          contact_reference_id: contactID || null,
        } },
        csrfToken,
      );
      const candidate = unwrapMutationPayload(data.createCandidate, response.status).candidate;
      if (!candidate) throw new Error("Invalid candidate response");
      onCreated(candidate.candidate_number);
    } catch (error) {
      setSaveError(isForbiddenError(error)
        ? "Anda tidak memiliki izin menambah kandidat."
        : "Kandidat gagal disimpan. Periksa nomor, referensi, dan tanggal lalu coba lagi.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form className="space-y-4 rounded-xl border border-[#dde3e8] bg-white p-4" onSubmit={(event) => void submit(event)}>
      <h2 className="font-semibold">Tambah kandidat</h2>
      <fieldset disabled={saving} className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1">Nomor kandidat
          <input name="candidate_number" className={styles.fieldInput} maxLength={120} required />
        </label>
        <label className="flex flex-col gap-1">Nama pilihan
          <input name="preferred_name" className={styles.fieldInput} maxLength={180} />
        </label>
        <label className="flex flex-col gap-1">Versi pemberitahuan privasi
          <input name="privacy_notice_version" className={styles.fieldInput} required />
        </label>
        <label className="flex flex-col gap-1">Waktu pengakuan privasi (waktu lokal)
          <input name="notice_acknowledged_at" type="datetime-local" className={styles.fieldInput} required />
        </label>
        <label className="flex flex-col gap-1">Batas retensi data
          <input name="data_retention_until" type="date" className={styles.fieldInput} required />
        </label>
      </fieldset>
      {canReadParties ? (
        <fieldset disabled={saving || lookupBusy} className="space-y-2">
          <legend className="font-medium">Individu dan kontak (opsional)</legend>
          <div className="flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1">Kode individu
              <input className={styles.fieldInput} value={partyCode} onChange={(event) => {
                setPartyCode(event.target.value);
                setPerson(null);
                setContacts(null);
                setContactID("");
                setLookupError(null);
              }} />
            </label>
            <button type="button" className={styles.btnSecondary} disabled={!partyCode.trim()}
              onClick={() => void findPerson()}>Cari individu</button>
          </div>
          {person ? <p>Individu terpilih · {person.display_name}</p> : null}
          {contacts && contacts.nodes.length > 0 ? (
            <label className="flex flex-col gap-1">Kontak
              <select className={styles.fieldSelect} value={contactID} onChange={(event) => setContactID(event.target.value)}>
                <option value="">Tanpa kontak</option>
                {contacts.nodes.map((contact) => <option key={contact._id} value={contact._id}>{contact.full_name}</option>)}
              </select>
            </label>
          ) : null}
          {contacts?.pageInfo.hasNextPage ? (
            <button type="button" className={styles.btnSecondary} onClick={() => void loadMoreContacts()}>Muat kontak berikutnya</button>
          ) : null}
          {lookupError ? <p role="alert">{lookupError}</p> : null}
        </fieldset>
      ) : null}
      {saveError ? <p role="alert">{saveError}</p> : null}
      <div className="flex gap-2">
        <button type="submit" className={styles.btnPrimary} disabled={!csrfToken || saving || lookupBusy || (partyCode.trim() !== "" && !person)}>
          {saving ? "Menyimpan…" : "Simpan kandidat"}
        </button>
        <button type="button" className={styles.btnSecondary} disabled={saving || lookupBusy} onClick={onCancel}>Batal</button>
      </div>
    </form>
  );
}
