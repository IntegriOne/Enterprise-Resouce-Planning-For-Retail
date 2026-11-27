"use client";

import { useEffect, useState, type FormEvent } from "react";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import {
  AuthRequestError,
  getPasswordStrength,
  validateIamPassword,
  validateIamUsername,
} from "@/lib/graphql/iam-browser-auth";
import { submitAdminCreateUser } from "@/lib/graphql/iam-dashboard-mutations";
import {
  checkUsernameAvailabilityViaGraphQL,
  fetchEmployeeProvisioningDetails,
  fetchEmployeeProvisioningRequests,
  formatIamDashboardError,
  isForbiddenError,
  isStepUpRequiredError,
  isUnauthenticatedError,
  type EmployeeProvisioningOption,
  type UsernameAvailabilityResult,
} from "@/lib/graphql/iam-dashboard-queries";
import {
  lookupEnterpriseScopeTargets,
  lookupUnlinkedPartyContacts,
  type PartyContactProvisionOption,
  type ScopeTargetOption,
} from "@/lib/graphql/ef-scope-lookup";
import type { IamUser, ProvisionUserKind } from "@/lib/graphql/iam-types";

const KIND_OPTIONS: Array<{
  value: ProvisionUserKind;
  shortLabel: string;
  label: string;
  mutationName: string;
  ctaLabel: string;
  loadingLabel: string;
  hint: string;
}> = [
  {
    value: "EMPLOYEE",
    shortLabel: "Employee",
    label: "EMPLOYEE — Pegawai Internal",
    mutationName: "createEmployeeUser",
    ctaLabel: "Create Employee",
    loadingLabel: "Membuat Employee...",
    hint: "Pilih pengajuan dari HRM untuk membuat dan menautkan akun pegawai.",
  },
  {
    value: "VENDOR",
    shortLabel: "Vendor",
    label: "VENDOR — Mitra Jasa / CMT",
    mutationName: "createVendorUser",
    ctaLabel: "Create Vendor",
    loadingLabel: "Membuat Vendor...",
    hint: "Memanggil mutasi GraphQL createVendorUser sekaligus menautkan FK party_contacts.iam_user_id di Enterprise Foundation.",
  },
  {
    value: "SUPPLIER",
    shortLabel: "Supplier",
    label: "SUPPLIER — Pemasok Bahan Mentah",
    mutationName: "createSupplierUser",
    ctaLabel: "Create Supplier",
    loadingLabel: "Membuat Supplier...",
    hint: "Memanggil mutasi GraphQL createSupplierUser sekaligus menautkan FK party_contacts.iam_user_id di Enterprise Foundation.",
  },
  {
    value: "CUSTOMER",
    shortLabel: "Customer",
    label: "CUSTOMER — Pelanggan Web",
    mutationName: "createUser",
    ctaLabel: "Create Customer",
    loadingLabel: "Membuat Customer...",
    hint: "Memanggil mutasi GraphQL createUser untuk kategori CUSTOMER (status awal PENDING).",
  },
];

export function UserCreateModal({
  open,
  initialKind = "EMPLOYEE",
  allowedKinds,
  onClose,
  onCreated,
}: {
  open: boolean;
  initialKind?: ProvisionUserKind;
  allowedKinds?: ProvisionUserKind[];
  onClose: () => void;
  onCreated: (user: IamUser | null) => void;
}) {
  const { csrfToken, openStepUpModal, pushToast } = useIdentityWorkspace();

  const selectableOptions = allowedKinds?.length
    ? KIND_OPTIONS.filter((opt) => allowedKinds.includes(opt.value))
    : KIND_OPTIONS;

  const [kind, setKind] = useState<ProvisionUserKind>(initialKind);
  const [partyId, setPartyId] = useState("");
  const [partySearch, setPartySearch] = useState("");
  const [partyOptions, setPartyOptions] = useState<ScopeTargetOption[]>([]);
  const [partyLookupNote, setPartyLookupNote] = useState<string | null>(null);
  const [isLoadingParties, setIsLoadingParties] = useState(false);
  const [contactId, setContactId] = useState("");
  const [contactOptions, setContactOptions] = useState<PartyContactProvisionOption[]>([]);
  const [contactLookupNote, setContactLookupNote] = useState<string | null>(null);
  const [isLoadingContacts, setIsLoadingContacts] = useState(false);
  const [contactRole, setContactRole] = useState("portal_pic");
  const [jobTitle, setJobTitle] = useState("");
  const [phone, setPhone] = useState("");
  const [isPrimaryContact, setIsPrimaryContact] = useState(false);
  const [provisioningRequestID, setProvisioningRequestID] = useState("");
  const [provisioningOptions, setProvisioningOptions] = useState<EmployeeProvisioningOption[]>([]);
  const [queueCursor, setQueueCursor] = useState<string | null>(null);
  const [nextQueueCursor, setNextQueueCursor] = useState<string | null>(null);
  const [queueLoading, setQueueLoading] = useState(false);
  const [queueError, setQueueError] = useState<string | null>(null);
  const [queueNeedsMFA, setQueueNeedsMFA] = useState(false);
  const [queueRefresh, setQueueRefresh] = useState(0);
  const selectedRequest = provisioningOptions.find((request) => request.id === provisioningRequestID);
  const isResumingEmployee = kind === "EMPLOYEE" && selectedRequest?.status === "processing";

  const workerID = selectedRequest?.worker_id;
  const legalEntityID = selectedRequest?.legal_entity_id;
  const [provisioningDetails, setProvisioningDetails] = useState<{
    requestID: string;
    workerName: string | null;
    legalEntity: { legal_name: string; entity_code: string } | null;
  } | null>(null);
  const selectedDetails = provisioningDetails?.requestID === provisioningRequestID ? provisioningDetails : null;
  const [enteredFullName, setFullName] = useState<string | null>(null);
  const fullName = enteredFullName ?? (kind === "EMPLOYEE" ? selectedDetails?.workerName ?? "" : "");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isCheckingUsername, setIsCheckingUsername] = useState(false);
  const [usernameCheck, setUsernameCheck] =
    useState<UsernameAvailabilityResult | null>(null);

  const isPartnerKind = kind === "VENDOR" || kind === "SUPPLIER";

  useEffect(() => {
    if (!open || kind !== "EMPLOYEE") return;
    let active = true;
    const timer = window.setTimeout(async () => {
      setQueueLoading(true);
      setQueueError(null);
      setQueueNeedsMFA(false);
      try {
        const queue = await fetchEmployeeProvisioningRequests(queueCursor);
        if (!active) return;
        setProvisioningOptions((previous) =>
          queueCursor ? [...previous, ...queue.nodes] : queue.nodes,
        );
        setNextQueueCursor(queue.pageInfo.hasNextPage ? queue.pageInfo.endCursor : null);
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        const needsMFA = isStepUpRequiredError(err);
        setQueueNeedsMFA(needsMFA);
        setQueueError(needsMFA
          ? "Verifikasi ulang untuk membuka antrean pengajuan akun."
          : isForbiddenError(err)
            ? "Anda tidak memiliki izin membuka antrean pengajuan akun."
            : "Antrean pengajuan akun gagal dimuat.");
      } finally {
        if (active) setQueueLoading(false);
      }
    }, 0);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [open, kind, queueCursor, queueRefresh]);

  useEffect(() => {
    if (!open || kind !== "EMPLOYEE" || !workerID || !legalEntityID) return;
    let active = true;
    void fetchEmployeeProvisioningDetails(workerID, legalEntityID).then((details) => {
      if (active) setProvisioningDetails({ requestID: provisioningRequestID, ...details });
    });
    return () => { active = false; };
  }, [open, kind, workerID, legalEntityID, provisioningRequestID]);

  function reloadProvisioningQueue() {
    if (provisioningRequestID !== "") setFullName(null);
    setProvisioningRequestID("");
    setProvisioningOptions([]);
    setNextQueueCursor(null);
    setQueueCursor(null);
    setQueueRefresh((previous) => previous + 1);
  }

  useEffect(() => {
    if (!open || !isPartnerKind) return;
    let active = true;
    const scopeType = kind === "VENDOR" ? "VENDOR_PARTY" : "SUPPLIER_PARTY";
    const timer = window.setTimeout(async () => {
      setIsLoadingParties(true);
      const res = await lookupEnterpriseScopeTargets(scopeType, partySearch);
      if (!active) return;
      setPartyOptions(res.options);
      setPartyLookupNote(res.unavailableReason ?? null);
      setIsLoadingParties(false);
    }, 220);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [open, isPartnerKind, kind, partySearch]);

  useEffect(() => {
    if (!open || !isPartnerKind || !partyId.trim()) {
      return;
    }
    let active = true;
    const timer = window.setTimeout(async () => {
      setIsLoadingContacts(true);
      const res = await lookupUnlinkedPartyContacts(partyId);
      if (!active) return;
      setContactOptions(res.contacts);
      setContactLookupNote(res.unavailableReason ?? null);
      setIsLoadingContacts(false);
    }, 180);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [open, isPartnerKind, partyId]);

  const visibleContactOptions = partyId.trim() ? contactOptions : [];
  const visibleContactLookupNote = partyId.trim() ? contactLookupNote : null;

  useEffect(() => {
    if (!open) return;
    const normalized = username.trim().toLowerCase();
    if (!normalized || validateIamUsername(normalized) !== null) {
      return;
    }

    let active = true;
    const timer = window.setTimeout(async () => {
      if (!active) return;
      setIsCheckingUsername(true);
      try {
        const result = await checkUsernameAvailabilityViaGraphQL(normalized);
        if (!active) return;
        setUsernameCheck(result);
        if (!result.available && result.existingUser) {
          setFieldErrors((prev) => ({
            ...prev,
            username: `Username '${result.normalizedUsername}' sudah digunakan oleh ${result.existingUser?.display_name} (${result.existingUser?.kind} · ${result.existingUser?.status}).`,
          }));
        } else {
          setFieldErrors((prev) => {
            const rest = { ...prev };
            delete rest.username;
            return rest;
          });
        }
      } catch {
        // Jangan memblokir input apabila cek availability tertunda
      } finally {
        if (active) setIsCheckingUsername(false);
      }
    }, 450);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [open, username]);

  if (!open) return null;

  async function runImmediateUsernameCheck(raw: string) {
    const normalized = raw.trim().toLowerCase();
    if (!normalized || validateIamUsername(normalized) !== null) return;
    setIsCheckingUsername(true);
    try {
      const result = await checkUsernameAvailabilityViaGraphQL(normalized);
      setUsernameCheck(result);
      if (!result.available && result.existingUser) {
        setFieldErrors((prev) => ({
          ...prev,
          username: `Username '${result.normalizedUsername}' sudah digunakan oleh ${result.existingUser?.display_name} (${result.existingUser?.kind} · ${result.existingUser?.status}).`,
        }));
      }
    } catch {
      // Abaikan bila gagal jaringan
    } finally {
      setIsCheckingUsername(false);
    }
  }

  const strength = getPasswordStrength(password);

  async function executeCreate() {
    setIsSubmitting(true);
    setSubmitError(null);
    try {
      const created = await submitAdminCreateUser(
        {
          kind,
          full_name: fullName,
          email,
          username,
          password,
          ...(isPartnerKind
            ? {
                party_id: partyId,
                ...(contactId.trim() ? { contact_id: contactId.trim() } : {}),
                ...(!contactId.trim() && contactRole.trim()
                  ? { contact_role: contactRole.trim() }
                  : {}),
                ...(!contactId.trim() && jobTitle.trim()
                  ? { job_title: jobTitle.trim() }
                  : {}),
                ...(!contactId.trim() && phone.trim() ? { phone: phone.trim() } : {}),
                ...(!contactId.trim() ? { is_primary_contact: isPrimaryContact } : {}),
              }
            : {}),
          ...(kind === "EMPLOYEE" && selectedRequest
            ? { worker_id: selectedRequest.worker_id, provisioning_request_id: selectedRequest.id }
            : {}),
        },
        csrfToken,
      );
      pushToast(
        "success",
        isResumingEmployee ? "Akun berhasil ditautkan" : "Akun berhasil dibuat",
        isResumingEmployee
          ? "Penautan pengajuan akun dari HRM sudah selesai."
          : `${fullName.trim()} berada pada status menunggu verifikasi email.`,
      );
      setPartyId("");
      setContactId("");
      setContactRole("portal_pic");
      setJobTitle("");
      setPhone("");
      setIsPrimaryContact(false);
      reloadProvisioningQueue();
      setFullName(null);
      setEmail("");
      setUsername("");
      setPassword("");
      setFieldErrors({});
      onCreated(created);
      onClose();
    } catch (err) {
      if (isUnauthenticatedError(err)) {
        window.location.replace("/login");
        return;
      }
      if (isStepUpRequiredError(err)) {
        openStepUpModal(
          "Pembuatan akun terkelola memerlukan verifikasi Step-Up MFA atau sesi aktif 10 menit terakhir.",
          async () => {
            await executeCreate();
          },
        );
        return;
      }
      if (kind !== "EMPLOYEE") {
        setSubmitError(formatIamDashboardError(err));
      } else if (isForbiddenError(err)) {
        setSubmitError("Anda tidak memiliki izin membuat atau menautkan akun pekerja.");
      } else if (err instanceof AuthRequestError && err.errorType === "CONFLICT") {
        setSubmitError("Pengajuan belum bisa diproses. Periksa status pengajuan serta email dan nama pengguna.");
      } else {
        setSubmitError("Akun pekerja gagal diproses. Periksa data pengajuan lalu coba lagi.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (isSubmitting) return;

    const errs: Record<string, string> = {};
    if (isPartnerKind && !partyId.trim()) {
      errs.partyId = `Mitra ${kind === "VENDOR" ? "Vendor" : "Supplier"} Enterprise Foundation wajib dipilih atau diisi.`;
    }
    if (kind === "EMPLOYEE" && !selectedRequest) {
      errs.provisioningRequest = "Pilih pengajuan akun dari HRM.";
    }
    if (!fullName.trim()) {
      errs.fullName = "Nama lengkap wajib diisi.";
    }
    if (!email.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      errs.email = "Format email tidak valid.";
    }
    const uErr = validateIamUsername(username);
    if (uErr) errs.username = uErr;

    const pErr = validateIamPassword(password, [fullName, email, username]);
    if (pErr) errs.password = pErr;

    setFieldErrors(errs);
    if (Object.keys(errs).length > 0) return;

    void executeCreate();
  }

  const activeOption =
    KIND_OPTIONS.find((k) => k.value === kind) ?? KIND_OPTIONS[0];
  const submitLabel = isResumingEmployee ? "Lanjutkan penautan" : activeOption.ctaLabel;
  const submittingLabel = isResumingEmployee ? "Menautkan akun…" : activeOption.loadingLabel;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0b2439]/45 p-4 backdrop-blur-[2px]"
      role="dialog"
      aria-modal="true"
      aria-label={isResumingEmployee ? "Lanjutkan penautan akun" : activeOption.ctaLabel}
      onClick={(e) => {
        if (e.target === e.currentTarget && !isSubmitting) onClose();
      }}
    >
      <div className="max-h-[92dvh] w-full max-w-lg overflow-y-auto rounded-xl border border-[#dde3e8] bg-white p-6 shadow-2xl">
        <div className="flex items-start justify-between gap-3 border-b border-[#dde3e8] pb-4">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-lg font-semibold text-[#0b2439]">
                {isResumingEmployee ? "Lanjutkan penautan akun" : activeOption.ctaLabel}
              </h2>
              <span
                className={`rounded-md border border-[#dde3e8] bg-[#f4f7f8] px-2 py-0.5 text-[11px] font-semibold text-[#133554] ${styles.monoText}`}
              >
                {activeOption.mutationName}
              </span>
            </div>
            <p className="mt-1 text-xs text-[#5a6672]">
              {isResumingEmployee
                ? "Selesaikan penautan akun yang sebelumnya diproses."
                : "Buat akun dengan email utama dan nama pengguna."}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            aria-label="Tutup modal buat akun"
            className={styles.iconBtn}
          >
            <IdentityIcon name="close" size={16} />
          </button>
        </div>

        {submitError ? (
          <div
            role="alert"
            className="mt-4 rounded-lg border border-[#fecdca] bg-[#fef3f2] px-3 py-2 text-xs text-[#b42318]"
          >
            {submitError}
          </div>
        ) : null}

        <form onSubmit={handleSubmit} className="mt-4 space-y-4" noValidate>
          <div className="flex flex-col gap-1.5">
            <label htmlFor="create-user-kind" className="text-xs font-semibold text-[#0b2439]">
              Kategori Akun (`UserKind`)
            </label>
            {selectableOptions.length > 1 ? (
              <div className="grid grid-cols-2 gap-2">
                {selectableOptions.map((opt) => {
                  const selected = kind === opt.value;
                  return (
                    <button
                      key={opt.value}
                      type="button"
                      disabled={isSubmitting}
                      onClick={() => {
                        setKind(opt.value);
                        setPartyId("");
                        setContactId("");
                        reloadProvisioningQueue();
                      }}
                      className={`flex items-center justify-between rounded-lg border px-3 py-2 text-left text-xs font-semibold transition-colors ${
                        selected
                          ? "border-[#133554] bg-[#e9eef2] text-[#0b2439]"
                          : "border-[#dde3e8] bg-white text-[#5a6672] hover:bg-[#f4f7f8]"
                      }`}
                    >
                      <span>{opt.shortLabel}</span>
                      <span className={`text-[10px] opacity-80 ${styles.monoText}`}>
                        {opt.value}
                      </span>
                    </button>
                  );
                })}
              </div>
            ) : null}
            <select
              id="create-user-kind"
              value={kind}
              onChange={(e) => {
                setKind(e.target.value as ProvisionUserKind);
                setPartyId("");
                setContactId("");
                reloadProvisioningQueue();
              }}
              disabled={isSubmitting || selectableOptions.length <= 1}
              className={styles.fieldSelect}
            >
              {selectableOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <span className="text-[11px] text-[#5a6672]">
              {activeOption.hint}
            </span>
          </div>

          {isPartnerKind ? (
            <div className="space-y-3 rounded-lg border border-[#dde3e8] bg-[#f8fafb] p-3">
              <div className="flex flex-col gap-1.5">
                <label htmlFor="create-user-party-search" className="text-xs font-semibold text-[#0b2439]">
                  Mitra {kind === "VENDOR" ? "Vendor" : "Supplier"} (`Enterprise Foundation Party`)
                </label>
                <input
                  id="create-user-party-search"
                  type="text"
                  value={partySearch}
                  onChange={(e) => setPartySearch(e.target.value)}
                  placeholder={`Cari kode atau nama mitra ${kind.toLowerCase()}...`}
                  disabled={isSubmitting}
                  className={styles.fieldInput}
                />
                <select
                  id="create-user-party-select"
                  aria-label="Pilih mitra Enterprise Foundation"
                  value={partyId}
                  onChange={(e) => {
                    setPartyId(e.target.value);
                    setContactId("");
                    setFieldErrors((prev) => {
                      const rest = { ...prev };
                      delete rest.partyId;
                      return rest;
                    });
                  }}
                  disabled={isSubmitting || isLoadingParties}
                  className={styles.fieldSelect}
                >
                  <option value="">
                    {isLoadingParties
                      ? "Memuat daftar mitra..."
                      : partyOptions.length > 0
                        ? "-- Pilih mitra dari Enterprise Foundation --"
                        : "-- Tidak ada mitra cocok; isi Party ID manual di bawah --"}
                  </option>
                  {partyOptions.map((opt) => (
                    <option key={opt.id} value={opt.id}>
                      {opt.code} · {opt.label} ({opt.secondary})
                    </option>
                  ))}
                </select>
                <input
                  id="create-user-party-id"
                  type="text"
                  value={partyOptions.some((option) => option.id === partyId) ? "" : partyId.replace(/^gid:\/\/integrione\/party\//, "")}
                  onChange={(e) => {
                    setPartyId(e.target.value);
                    setContactId("");
                  }}
                  placeholder="Referensi mitra jika pencarian tidak tersedia"
                  disabled={isSubmitting}
                  className={`${styles.fieldInput} ${styles.monoText}`}
                />
                {partyLookupNote ? (
                  <p className="text-[11px] text-[#5a6672]">{partyLookupNote}</p>
                ) : null}
                {fieldErrors.partyId ? (
                  <p className="text-xs text-[#b42318]">{fieldErrors.partyId}</p>
                ) : null}
              </div>

              <div className="flex flex-col gap-1.5">
                <label htmlFor="create-user-contact-select" className="text-xs font-semibold text-[#0b2439]">
                  Penautan Kontak PIC (`party_contacts.iam_user_id`)
                </label>
                <select
                  id="create-user-contact-select"
                  value={contactId}
                  onChange={(e) => {
                    const nextContactId = e.target.value;
                    setContactId(nextContactId);
                    const matched = visibleContactOptions.find((c) => c.id === nextContactId);
                    if (matched) {
                      if (!fullName.trim()) setFullName(matched.fullName);
                      if (!email.trim() && matched.email) setEmail(matched.email);
                    }
                  }}
                  disabled={isSubmitting || isLoadingContacts || !partyId.trim()}
                  className={styles.fieldSelect}
                >
                  <option value="">
                    Buat kontak PIC baru otomatis (status portal: PENDING_LINK)
                  </option>
                  {visibleContactOptions.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.fullName} ({c.contactRole}{c.email ? ` · ${c.email}` : ""})
                    </option>
                  ))}
                </select>
                {visibleContactLookupNote ? (
                  <p className="text-[11px] text-[#5a6672]">{visibleContactLookupNote}</p>
                ) : null}
              </div>

              {!contactId ? (
                <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
                  <div className="flex flex-col gap-1">
                    <label htmlFor="create-user-contact-role" className="text-[11px] font-semibold text-[#0b2439]">
                      Peran PIC
                    </label>
                    <input
                      id="create-user-contact-role"
                      type="text"
                      value={contactRole}
                      onChange={(e) => setContactRole(e.target.value)}
                      placeholder="portal_pic"
                      disabled={isSubmitting}
                      className={styles.fieldInput}
                    />
                  </div>
                  <div className="flex flex-col gap-1">
                    <label htmlFor="create-user-job-title" className="text-[11px] font-semibold text-[#0b2439]">
                      Jabatan (Opsional)
                    </label>
                    <input
                      id="create-user-job-title"
                      type="text"
                      value={jobTitle}
                      onChange={(e) => setJobTitle(e.target.value)}
                      placeholder="Finance / Purchasing PIC"
                      disabled={isSubmitting}
                      className={styles.fieldInput}
                    />
                  </div>
                  <div className="flex flex-col gap-1">
                    <label htmlFor="create-user-phone" className="text-[11px] font-semibold text-[#0b2439]">
                      Telepon PIC (Opsional)
                    </label>
                    <input
                      id="create-user-phone"
                      type="text"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder="+62..."
                      disabled={isSubmitting}
                      className={styles.fieldInput}
                    />
                  </div>
                  <label className="mt-4 flex items-center gap-2 text-xs text-[#0b2439]">
                    <input
                      type="checkbox"
                      checked={isPrimaryContact}
                      onChange={(e) => setIsPrimaryContact(e.target.checked)}
                      disabled={isSubmitting}
                    />
                    Jadikan kontak PIC utama
                  </label>
                </div>
              ) : null}
            </div>
          ) : null}

          {kind === "EMPLOYEE" ? (
            <div className="flex flex-col gap-1.5">
              <label htmlFor="create-user-provisioning-request" className="text-xs font-semibold text-[#0b2439]">
                Pengajuan akun dari HRM
              </label>
              <select
                id="create-user-provisioning-request"
                value={provisioningRequestID}
                onChange={(event) => {
                  if (provisioningRequestID !== "") setFullName(null);
                  setProvisioningRequestID(event.target.value);
                }}
                aria-invalid={Boolean(fieldErrors.provisioningRequest)}
                aria-describedby={fieldErrors.provisioningRequest ? "provisioning-request-error" : undefined}
                disabled={isSubmitting || (queueLoading && provisioningOptions.length === 0)}
                className={styles.fieldSelect}
              >
                <option value="">Pilih pengajuan</option>
                {provisioningOptions.map((request) => (
                  <option key={request.id} value={request.id}>
                    {request.worker_number} ·{" "}
                    {new Date(request.requested_at).toLocaleString("id-ID")} ·{" "}
                    {request.status === "processing" ? "Sedang diproses" : "Menunggu"}
                  </option>
                ))}
              </select>
              {selectedRequest ? (
                <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 rounded-lg bg-[#f4f7f8] p-3 text-xs">
                  <dt className="font-semibold">Nomor pekerja</dt>
                  <dd>{selectedRequest.worker_number}</dd>
                  {selectedDetails?.workerName ? (
                    <>
                      <dt className="font-semibold">Nama pekerja</dt>
                      <dd>{selectedDetails.workerName}</dd>
                    </>
                  ) : null}
                  {selectedDetails?.legalEntity ? (
                    <>
                      <dt className="font-semibold">Badan hukum</dt>
                      <dd>{selectedDetails.legalEntity.legal_name} ({selectedDetails.legalEntity.entity_code})</dd>
                    </>
                  ) : null}
                  <dt className="font-semibold">Diajukan</dt>
                  <dd>{new Date(selectedRequest.requested_at).toLocaleString("id-ID")}</dd>
                </dl>
              ) : null}
              {isResumingEmployee ? (
                <p role="status" className="text-xs text-[#5a6672]">
                  Gunakan kembali data pengajuan akun. Operator yang memulai pengajuan di IAM dapat menyelesaikan penautan akun yang sudah dibuat.
                </p>
              ) : null}
              {fieldErrors.provisioningRequest ? (
                <span id="provisioning-request-error" className="text-xs text-red-700">
                  {fieldErrors.provisioningRequest}
                </span>
              ) : null}
              {queueLoading ? <p role="status" className="text-xs text-[#5a6672]">Memuat pengajuan…</p> : null}
              {queueError ? (
                <div className="flex items-center gap-2 text-xs">
                  <span role="alert">{queueError}</span>
                  <button type="button" className="font-semibold underline" onClick={() => {
                    if (queueNeedsMFA) {
                      openStepUpModal("Verifikasi MFA untuk membuka antrean pengajuan akun.", reloadProvisioningQueue);
                    } else {
                      reloadProvisioningQueue();
                    }
                  }}>
                    {queueNeedsMFA ? "Verifikasi MFA" : "Coba lagi"}
                  </button>
                </div>
              ) : null}
              {!queueLoading && !queueError && provisioningOptions.length === 0 ? (
                <p className="text-xs text-[#5a6672]">Belum ada pengajuan akun HRM yang terbuka.</p>
              ) : null}
              {nextQueueCursor ? (
                <button type="button" className="self-start text-xs font-semibold underline"
                  disabled={queueLoading || isSubmitting} onClick={() => setQueueCursor(nextQueueCursor)}>
                  Muat pengajuan berikutnya
                </button>
              ) : null}
            </div>
          ) : null}

          <div className="flex flex-col gap-1.5">
            <label htmlFor="create-user-name" className="text-xs font-semibold text-[#0b2439]">
              Nama Tampilan Lengkap
            </label>
            <input
              id="create-user-name"
              type="text"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              placeholder="Contoh: Ardiansyah Putra"
              disabled={isSubmitting}
              className={styles.fieldInput}
            />
            {fieldErrors.fullName ? (
              <p className="text-xs text-[#b42318]">{fieldErrors.fullName}</p>
            ) : (
              <span className="text-[11px] text-[#5a6672]">
                Dipetakan ke field `display_name` entitas `users`.
              </span>
            )}
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="create-user-email" className="text-xs font-semibold text-[#0b2439]">
                Email Utama
              </label>
              <input
                id="create-user-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="nama@integrione.id"
                disabled={isSubmitting}
                className={styles.fieldInput}
              />
              {fieldErrors.email ? (
                <p className="text-xs text-[#b42318]">{fieldErrors.email}</p>
              ) : null}
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="create-user-username" className="text-xs font-semibold text-[#0b2439]">
                Username Login
              </label>
              <input
                id="create-user-username"
                type="text"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  setUsernameCheck(null);
                  if (fieldErrors.username) {
                    setFieldErrors((prev) => {
                      const rest = { ...prev };
                      delete rest.username;
                      return rest;
                    });
                  }
                }}
                onBlur={() => void runImmediateUsernameCheck(username)}
                placeholder="ardiansyah.p"
                disabled={isSubmitting}
                className={`${styles.fieldInput} ${styles.monoText}`}
              />
              {isCheckingUsername ? (
                <p className="text-[11px] text-[#5a6672]">
                  Memeriksa ketersediaan username via GraphQL...
                </p>
              ) : fieldErrors.username ? (
                <p className="text-xs text-[#b42318]">{fieldErrors.username}</p>
              ) : usernameCheck?.available &&
                usernameCheck.normalizedUsername ===
                  username.trim().toLowerCase() ? (
                <p className="text-[11px] font-medium text-[#067647]">
                  Username `{usernameCheck.normalizedUsername}` tersedia.
                </p>
              ) : null}
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="create-user-password" className="text-xs font-semibold text-[#0b2439]">
              Kata Sandi Awal
            </label>
            <div className="relative">
              <input
                id="create-user-password"
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Min. 12 karakter (Aa1!)"
                disabled={isSubmitting}
                className={`${styles.fieldInput} pr-11`}
              />
              <button
                type="button"
                onClick={() => setShowPassword((prev) => !prev)}
                aria-label={showPassword ? "Sembunyikan sandi" : "Tampilkan sandi"}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1.5 text-[#5a6672] hover:bg-[#f4f7f8]"
              >
                <IdentityIcon name={showPassword ? "unlock" : "lock"} size={15} />
              </button>
            </div>
            {fieldErrors.password ? (
              <p className="text-xs text-[#b42318]">{fieldErrors.password}</p>
            ) : (
              <span className="text-[11px] text-[#5a6672]">
                12–128 karakter, wajib huruf besar/kecil, angka, simbol, dan tanpa unsur nama/email.
              </span>
            )}
            {password ? (
              <div className="mt-1 space-y-1">
                <div className="h-1.5 w-full overflow-hidden rounded-full bg-[#e9eef2]">
                  <div
                    className="h-full transition-all"
                    style={{
                      width: `${strength.score}%`,
                      backgroundColor: strength.color,
                    }}
                  />
                </div>
                <p className="text-[11px] font-semibold" style={{ color: strength.color }}>
                  Kekuatan sandi: {strength.label}
                </p>
              </div>
            ) : null}
          </div>

          <div className="flex items-center justify-end gap-2.5 border-t border-[#dde3e8] pt-4">
            <button
              type="button"
              onClick={onClose}
              disabled={isSubmitting}
              className={styles.btnSecondary}
            >
              Batal
            </button>
            <button type="submit" disabled={isSubmitting} className={styles.btnPrimary}>
              {isSubmitting ? submittingLabel : submitLabel}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
