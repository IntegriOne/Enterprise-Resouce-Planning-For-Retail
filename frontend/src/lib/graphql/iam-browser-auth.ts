/**
 * IAM Browser Authentication Client & Operations
 *
 * Mengonsumsi endpoint tunggal same-origin `/api/graphql` menggunakan cookie HttpOnly
 * dan header X-CSRF-Token sesuai standar keamanan IntegriOne.
 */

export const GRAPHQL_PATH = "/api/graphql";

export type EmailVerificationStatus =
  | "VERIFIED"
  | "EXPIRED"
  | "ALREADY_USED"
  | "INVALID";

export type CustomerEmailVerificationStatus = EmailVerificationStatus;

export type MutationErrorType =
  | "ARGUMENT"
  | "AUTHENTICATION"
  | "AUTHORIZATION"
  | "CONFLICT"
  | "INTERNAL_ERROR"
  | "NOT_FOUND"
  | "THROTTLE"
  | "UNAVAILABLE"
  | "UNPROCESSABLE_ENTITY";

export type MutationErrorItem = {
  type: MutationErrorType | string;
  message: string;
};

export type MutationResult = {
  errors: MutationErrorItem[];
  was_successful: boolean;
};

export type UserLoginPayload = Partial<MutationResult> & {
  status?:
    | "AUTHENTICATED"
    | "MFA_REQUIRED"
    | "EMAIL_VERIFICATION_REQUIRED"
    | "UNVERIFIED"
    | null;
  mfaChallenge?: string | null;
  mfaExpiresAt?: string | null;
  csrfToken?: string | null;
  redirectLocation?: string | null;
};

export type ValidateUserMfaPayload = Partial<MutationResult> & {
  csrfToken?: string | null;
  redirectLocation?: string | null;
};

export type CreateUserPayload = Partial<MutationResult> & {
  user?: {
    id: string;
    _id: string;
    kind: string;
    status: string;
    display_name: string;
  } | null;
  redirectLocation?: string | null;
};

export type VerifyEmailPayload = Partial<MutationResult> & {
  status?: EmailVerificationStatus | null;
  verified?: boolean | null;
};

export type VerifyCustomerEmailPayload = VerifyEmailPayload;

export type ResendValidationEmailPayload = Partial<MutationResult> & {
  is_resend?: boolean | null;
};

export type UserResetPasswordPayload = Partial<MutationResult> & {
  message?: string | null;
};

export type ResetPasswordPayload = Partial<MutationResult> & {
  reset?: boolean | null;
};

export type UserNamePayload = Partial<MutationResult> & {
  username?: string | null;
};

export type MyWorkspaceAccessPayload = {
  canAccessIdentity: boolean;
  defaultLocation: string;
};

export type GraphQLError = {
  message?: string;
  extensions?: {
    code?: string;
    type?: string;
    [key: string]: unknown;
  };
};

export type GraphQLResponse<T> = {
  data?: T | null;
  errors?: GraphQLError[];
};

const DEFAULT_AUTH_ERROR_MESSAGE = "authentication request failed";

export class AuthRequestError extends Error {
  readonly status: number;
  readonly code: string | undefined;
  readonly errorType: string | undefined;

  constructor(
    status: number,
    code: string | undefined,
    message?: string,
    errorType?: string,
  ) {
    super(message || DEFAULT_AUTH_ERROR_MESSAGE);
    this.name = "AuthRequestError";
    this.status = status;
    this.code = code;
    this.errorType = errorType;
  }
}

const MUTATION_RESULT_FIELDS = "errors { type message } was_successful";

const SESSION_QUERY =
  "query LoginBrowserSession { browserSession { authenticated principalId csrfToken } }";

const MY_WORKSPACE_ACCESS_QUERY =
  "query MyWorkspaceAccess { myWorkspaceAccess { canAccessIdentity defaultLocation } }";

const LOGIN_MUTATION = `mutation LoginWithPassword($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation ${MUTATION_RESULT_FIELDS} } }`;

const MFA_MUTATION = `mutation VerifyLoginMfa($input: ValidateUserMfaInput!) { validateUserMfa(input: $input) { csrfToken redirectLocation ${MUTATION_RESULT_FIELDS} } }`;

const CREATE_USER_MUTATION = `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { user { id _id kind status display_name } redirectLocation ${MUTATION_RESULT_FIELDS} } }`;

const VERIFY_EMAIL_MUTATION = `mutation VerifyEmail($input: VerifyEmailInput!) { verifyEmail(input: $input) { status verified ${MUTATION_RESULT_FIELDS} } }`;

const RESEND_VALIDATION_EMAIL_MUTATION = `mutation ResendValidationEmailToken($input: ResendValidationEmailInput) { resendValidationEmailToken(input: $input) { is_resend ${MUTATION_RESULT_FIELDS} } }`;

const USER_RESET_PASSWORD_MUTATION = `mutation UserResetPassword($input: UserResetPasswordInput!) { userResetPassword(input: $input) { message ${MUTATION_RESULT_FIELDS} } }`;

const RESET_PASSWORD_MUTATION = `mutation ResetPassword($input: IAMResetPasswordInput!) { resetPassword(input: $input) { reset ${MUTATION_RESULT_FIELDS} } }`;

const CHANGE_USERNAME_MUTATION = `mutation ChangeUsername($input: UsernameInput!) { changeUsername(input: $input) { username ${MUTATION_RESULT_FIELDS} } }`;

function mapErrorTypeToStatus(errorType?: string): number {
  switch (errorType) {
    case "AUTHENTICATION":
      return 401;
    case "AUTHORIZATION":
      return 403;
    case "NOT_FOUND":
      return 404;
    case "CONFLICT":
      return 409;
    case "THROTTLE":
      return 429;
    case "UNAVAILABLE":
      return 503;
    case "INTERNAL_ERROR":
      return 500;
    default:
      return 400;
  }
}

// Memastikan payload MutationResult berhasil (was_successful=true dan errors kosong), atau melempar AuthRequestError terstruktur.
export function unwrapMutationPayload<T extends Partial<MutationResult>>(
  payload: T | null | undefined,
  responseStatus: number,
  fallbackMessage = "Permintaan belum dapat diproses",
): T {
  if (!payload) {
    throw new AuthRequestError(500, "INVALID_RESPONSE", fallbackMessage);
  }
  const hasErrors = Array.isArray(payload.errors) && payload.errors.length > 0;
  if (payload.was_successful === false || (payload.was_successful !== true && hasErrors) || hasErrors) {
    const firstErr = payload.errors?.[0];
    const errorType = firstErr?.type;
    const message = firstErr?.message || fallbackMessage;
    throw new AuthRequestError(
      errorType ? mapErrorTypeToStatus(errorType) : responseStatus || 400,
      errorType,
      message,
      errorType,
    );
  }
  if (payload.was_successful !== true) {
    throw new AuthRequestError(500, "INVALID_RESPONSE", fallbackMessage);
  }
  return payload;
}

export async function executeGraphQL<T>(
  query: string,
  variables: Record<string, unknown> | undefined,
  csrfToken?: string | null,
  customHeaders?: Record<string, string>,
): Promise<{ data: T; response: Response }> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...customHeaders,
  };
  if (csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }

  let response: Response;
  try {
    response = await fetch(GRAPHQL_PATH, {
      method: "POST",
      credentials: "include",
      cache: "no-store",
      headers,
      body: JSON.stringify({ query, variables }),
    });
  } catch {
    throw new AuthRequestError(0, "NETWORK", "Koneksi terputus");
  }

  let payload: GraphQLResponse<T>;
  try {
    payload = (await response.json()) as GraphQLResponse<T>;
  } catch {
    throw new AuthRequestError(
      response.status,
      "INVALID_RESPONSE",
      "Format respons tidak valid",
    );
  }

  if (!response.ok || payload.errors?.length || !payload.data) {
    const firstErr = payload.errors?.[0];
    const firstCode = firstErr?.extensions?.code;
    const firstType =
      typeof firstErr?.extensions?.type === "string"
        ? firstErr.extensions.type
        : undefined;
    const firstMessage = firstErr?.message;
    throw new AuthRequestError(response.status, firstCode, firstMessage, firstType);
  }

  return { data: payload.data, response };
}

export async function fetchBrowserSession(signal?: AbortSignal): Promise<{
  authenticated: boolean;
  principalId?: string | null;
  csrfToken: string | null;
}> {
  const url = GRAPHQL_PATH + "?query=" + encodeURIComponent(SESSION_QUERY);
  let response: Response;
  try {
    response = await fetch(url, {
      method: "GET",
      credentials: "include",
      cache: "no-store",
      signal,
    });
  } catch {
    throw new AuthRequestError(0, "NETWORK", "Koneksi sesi terputus");
  }

  if (!response.ok) {
    throw new AuthRequestError(response.status, "BOOTSTRAP");
  }

  const payload = (await response.json()) as GraphQLResponse<{
    browserSession?: {
      authenticated: boolean;
      principalId?: string | null;
      csrfToken?: string | null;
    } | null;
  }>;

  if (payload.errors?.length || !payload.data?.browserSession) {
    throw new AuthRequestError(response.status, "BOOTSTRAP");
  }

  return {
    authenticated: payload.data.browserSession.authenticated,
    principalId: payload.data.browserSession.principalId,
    csrfToken:
      payload.data.browserSession.csrfToken ??
      response.headers.get("X-CSRF-Token") ??
      null,
  };
}

export async function submitUserLogin(
  identifier: string,
  password: string,
  csrfToken: string | null,
): Promise<{ payload: UserLoginPayload; locationHeader: string | null }> {
  const { data, response } = await executeGraphQL<{
    userLogin?: UserLoginPayload | null;
  }>(LOGIN_MUTATION, { input: { identifier, password } }, csrfToken);

  return {
    payload: unwrapMutationPayload(data.userLogin, response.status, "Login gagal"),
    locationHeader: response.headers.get("Location"),
  };
}

// Mendeteksi payload userLogin ketika akun sudah terdaftar namun email belum diverifikasi.
export function isUnverifiedLoginResponse(
  payload: Pick<UserLoginPayload, "status" | "redirectLocation"> | null | undefined,
  locationHeader?: string | null,
): boolean {
  const target = (locationHeader || payload?.redirectLocation || "").trim();
  if (target === "/verify-email") {
    return true;
  }
  return (
    payload?.status === "EMAIL_VERIFICATION_REQUIRED" ||
    payload?.status === "UNVERIFIED"
  );
}

// Mendeteksi error userLogin ketika server menyatakan email/identitas belum terverifikasi.
export function isUnverifiedLoginError(error: unknown): boolean {
  if (!(error instanceof AuthRequestError)) return false;
  const code = (error.code || "").toUpperCase();
  const errType = (error.errorType || "").toUpperCase();
  if (
    code === "EMAIL_NOT_VERIFIED" ||
    code === "IDENTITY_NOT_VERIFIED" ||
    code === "EMAIL_VERIFICATION_REQUIRED" ||
    errType === "EMAIL_NOT_VERIFIED" ||
    errType === "IDENTITY_NOT_VERIFIED"
  ) {
    return true;
  }
  const msg = (error.message || "").toLowerCase();
  return (
    msg.includes("identity_not_verified") ||
    msg.includes("email not verified") ||
    msg.includes("not verified") ||
    msg.includes("belum terverifikasi")
  );
}

export async function submitValidateUserMfa(
  challenge: string,
  code: string,
  mode: "totp" | "recovery",
  csrfToken: string | null,
): Promise<{ payload: ValidateUserMfaPayload; locationHeader: string | null }> {
  const input =
    mode === "totp"
      ? { challenge, totpCode: code }
      : { challenge, recoveryCode: code };

  const { data, response } = await executeGraphQL<{
    validateUserMfa?: ValidateUserMfaPayload | null;
  }>(MFA_MUTATION, { input }, csrfToken);

  return {
    payload: unwrapMutationPayload(
      data.validateUserMfa,
      response.status,
      "Verifikasi MFA gagal",
    ),
    locationHeader: response.headers.get("Location"),
  };
}

export async function submitCreateUser(
  fullName: string,
  email: string,
  username: string,
  password: string,
  csrfToken: string | null,
): Promise<{ payload: CreateUserPayload; locationHeader: string | null }> {
  const { data, response } = await executeGraphQL<{
    createUser?: CreateUserPayload | null;
  }>(
    CREATE_USER_MUTATION,
    { input: { full_name: fullName, email, username, password } },
    csrfToken,
  );

  return {
    payload: unwrapMutationPayload(
      data.createUser,
      response.status,
      "Pendaftaran akun gagal",
    ),
    locationHeader: response.headers.get("Location"),
  };
}

export async function submitVerifyEmail(
  token: string,
  csrfToken: string | null,
): Promise<VerifyEmailPayload> {
  const { data, response } = await executeGraphQL<{
    verifyEmail?: VerifyEmailPayload | null;
  }>(VERIFY_EMAIL_MUTATION, { input: { token } }, csrfToken);

  return unwrapMutationPayload(
    data.verifyEmail,
    response.status,
    "Verifikasi email gagal",
  );
}

export const submitVerifyCustomerEmail = submitVerifyEmail;

export async function submitResendValidationEmailToken(
  email: string | undefined,
  csrfToken: string | null,
): Promise<ResendValidationEmailPayload> {
  const input = email ? { email } : undefined;
  const { data, response } = await executeGraphQL<{
    resendValidationEmailToken?: ResendValidationEmailPayload | null;
  }>(RESEND_VALIDATION_EMAIL_MUTATION, input ? { input } : {}, csrfToken);

  return unwrapMutationPayload(
    data.resendValidationEmailToken,
    response.status,
    "Pengiriman ulang tautan verifikasi gagal",
  );
}

export async function submitUserResetPassword(
  email: string,
  csrfToken: string | null,
): Promise<UserResetPasswordPayload> {
  const { data, response } = await executeGraphQL<{
    userResetPassword?: UserResetPasswordPayload | null;
  }>(USER_RESET_PASSWORD_MUTATION, { input: { email } }, csrfToken);

  return unwrapMutationPayload(
    data.userResetPassword,
    response.status,
    "Permintaan reset kata sandi gagal",
  );
}

export async function submitResetPassword(
  token: string,
  newPassword: string,
  csrfToken: string | null,
): Promise<boolean> {
  const { data, response } = await executeGraphQL<{
    resetPassword?: ResetPasswordPayload | null;
  }>(RESET_PASSWORD_MUTATION, { input: { token, newPassword } }, csrfToken);

  const payload = unwrapMutationPayload(
    data.resetPassword,
    response.status,
    "Pembaruan kata sandi gagal",
  );
  return Boolean(payload.reset);
}

export async function submitChangeUsername(
  username: string,
  csrfToken: string | null,
): Promise<UserNamePayload> {
  const { data, response } = await executeGraphQL<{
    changeUsername?: UserNamePayload | null;
  }>(CHANGE_USERNAME_MUTATION, { input: { username } }, csrfToken);

  return unwrapMutationPayload(
    data.changeUsername,
    response.status,
    "Pembaruan username gagal",
  );
}

export async function fetchMyWorkspaceAccessBrowser(
  signal?: AbortSignal,
): Promise<MyWorkspaceAccessPayload> {
  const url = GRAPHQL_PATH + "?query=" + encodeURIComponent(MY_WORKSPACE_ACCESS_QUERY);
  let response: Response;
  try {
    response = await fetch(url, {
      method: "GET",
      credentials: "include",
      cache: "no-store",
      signal,
    });
  } catch {
    throw new AuthRequestError(0, "NETWORK", "Gagal memeriksa akses ruang kerja");
  }

  let payload: GraphQLResponse<{
    myWorkspaceAccess?: MyWorkspaceAccessPayload | null;
  }>;
  try {
    payload = (await response.json()) as GraphQLResponse<{
      myWorkspaceAccess?: MyWorkspaceAccessPayload | null;
    }>;
  } catch {
    throw new AuthRequestError(
      response.status,
      "INVALID_RESPONSE",
      "Format respons akses tidak valid",
    );
  }

  if (!response.ok || payload.errors?.length || !payload.data?.myWorkspaceAccess) {
    const firstErr = payload.errors?.[0];
    const code = firstErr?.extensions?.code;
    const errorType =
      typeof firstErr?.extensions?.type === "string"
        ? firstErr.extensions.type
        : undefined;
    throw new AuthRequestError(
      response.status,
      code,
      firstErr?.message,
      errorType,
    );
  }

  return payload.data.myWorkspaceAccess;
}

const MY_ACCESS_SNAPSHOT_SERVER_QUERY = `
  query MyAccessSnapshotServer {
    myAccessSnapshot {
      principal_id
      user_kind
      status
      mfa_enabled
      auth_epoch
      perm_revision
      recovery_epoch
      can_access_identity
      default_redirect_path
      allowed_redirect_paths
      roles { id code name priority default_redirect_path }
      permissions { permission_code is_sensitive scope { type id } }
      synced_at_unix
      fresh_until_unix
    }
  }
`;

/**
 * Server Component Guard helper untuk mengevaluasi AccessSnapshot lengkap (peran & izin domain)
 * sebelum merender ruang kerja terlindungi (/human-resource dan /procurement).
 */
export async function fetchMyAccessSnapshotServer(
  cookieHeader: string,
): Promise<{
  principal_id: string;
  user_kind: string;
  status: string;
  can_access_identity: boolean;
  default_redirect_path: string;
  allowed_redirect_paths: string[];
  roles: Array<{ id: string; code: string; name: string; priority: number; default_redirect_path: string }>;
  permissions: Array<{ permission_code: string; is_sensitive: boolean; scope: { type: string; id?: string | null } }>;
}> {
  const baseUrl =
    process.env.GATEWAY_INTERNAL_URL ||
    (process.env.NODE_ENV === "production"
      ? "http://gateway-1:8080"
      : "https://erp.localhost:8443");

  let response: Response;
  try {
    response = await fetch(`${baseUrl}/api/graphql`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Cookie: cookieHeader,
      },
      cache: "no-store",
      body: JSON.stringify({ query: MY_ACCESS_SNAPSHOT_SERVER_QUERY }),
    });
  } catch {
    throw new AuthRequestError(0, "NETWORK", "Gagal menghubungi server otorisasi IAM");
  }

  if (!response.ok) {
    throw new AuthRequestError(response.status, "UNAVAILABLE");
  }

  const payload = (await response.json()) as GraphQLResponse<{
    myAccessSnapshot?: {
      principal_id: string;
      user_kind: string;
      status: string;
      can_access_identity: boolean;
      default_redirect_path: string;
      allowed_redirect_paths: string[];
      roles: Array<{ id: string; code: string; name: string; priority: number; default_redirect_path: string }>;
      permissions: Array<{ permission_code: string; is_sensitive: boolean; scope: { type: string; id?: string | null } }>;
    } | null;
  }>;

  if (payload.errors?.length || !payload.data?.myAccessSnapshot) {
    const firstErr = payload.errors?.[0];
    const code = firstErr?.extensions?.code;
    const errorType =
      typeof firstErr?.extensions?.type === "string"
        ? firstErr.extensions.type
        : undefined;
    throw new AuthRequestError(
      response.status,
      code,
      firstErr?.message,
      errorType,
    );
  }

  return payload.data.myAccessSnapshot;
}

/**
 * Server Component Guard helper untuk query myWorkspaceAccess
 * Dijalankan di lingkungan server-side Next.js.
 */
export async function fetchMyWorkspaceAccessServer(
  cookieHeader: string,
): Promise<MyWorkspaceAccessPayload> {
  // Dalam container Next.js, Gateway dapat diakses via Caddy ingress https://caddy atau host port internal.
  // Gunakan GATEWAY_INTERNAL_URL bila ada, atau fallback ke host internal Caddy / localhost.
  const baseUrl =
    process.env.GATEWAY_INTERNAL_URL ||
    (process.env.NODE_ENV === "production"
      ? "http://gateway-1:8080"
      : "https://erp.localhost:8443");

  const endpoint = `${baseUrl}/api/graphql`;

  let response: Response;
  try {
    response = await fetch(endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Cookie: cookieHeader,
      },
      cache: "no-store",
      body: JSON.stringify({ query: MY_WORKSPACE_ACCESS_QUERY }),
    });
  } catch {
    throw new AuthRequestError(0, "NETWORK", "Gagal menghubungi server IAM");
  }

  if (!response.ok) {
    throw new AuthRequestError(response.status, "UNAVAILABLE");
  }

  const payload = (await response.json()) as GraphQLResponse<{
    myWorkspaceAccess?: MyWorkspaceAccessPayload | null;
  }>;

  if (payload.errors?.length || !payload.data?.myWorkspaceAccess) {
    const firstErr = payload.errors?.[0];
    const code = firstErr?.extensions?.code;
    const errorType =
      typeof firstErr?.extensions?.type === "string"
        ? firstErr.extensions.type
        : undefined;
    throw new AuthRequestError(
      response.status,
      code,
      firstErr?.message,
      errorType,
    );
  }

  return payload.data.myWorkspaceAccess;
}

export function formatAuthErrorMessage(
  error: unknown,
  context: "login" | "mfa" | "register" | "verify" | "reset-request" | "reset-confirm",
): string {
  if (error instanceof AuthRequestError) {
    const errKind = error.errorType || error.code;
    const rawMsg = error.message?.trim();
    const hasServerMessage =
      Boolean(rawMsg) && rawMsg !== DEFAULT_AUTH_ERROR_MESSAGE;

    if (error.code === "NETWORK") {
      return "Koneksi ke layanan terputus. Silakan periksa jaringan Anda dan coba lagi.";
    }
    if (error.status === 0 || error.status >= 500 || errKind === "UNAVAILABLE") {
      return "Layanan sedang tidak tersedia. Silakan coba beberapa saat lagi.";
    }
    if (errKind === "INTERNAL_ERROR") {
      return "Terjadi kendala internal pada layanan. Silakan coba beberapa saat lagi.";
    }
    if (errKind === "FORBIDDEN" || error.status === 403 || errKind === "AUTHORIZATION") {
      if (hasServerMessage && rawMsg?.toLowerCase().includes("csrf")) {
        return "Sesi keamanan formulir telah diperbarui. Silakan kirim kembali permintaan Anda.";
      }
      if (hasServerMessage) {
        return `${rawMsg}`;
      }
      return "Sesi keamanan telah berakhir atau token kedaluwarsa. Silakan mulai kembali.";
    }
    if (
      errKind === "TOO_MANY_REQUESTS" ||
      errKind === "THROTTLE" ||
      error.status === 429
    ) {
      return hasServerMessage
        ? `Terlalu banyak percobaan (${rawMsg}). Silakan tunggu beberapa saat sebelum mencoba kembali.`
        : "Terlalu banyak percobaan. Silakan tunggu beberapa saat sebelum mencoba kembali.";
    }
    if (errKind === "INVALID_CREDENTIALS") {
      return "Email, username, atau kata sandi tidak sesuai. Jika baru mendaftar, pastikan email Anda telah diverifikasi.";
    }
    if (errKind === "AUTHENTICATION") {
      if (context === "login") {
        return hasServerMessage
          ? `Email, username, atau kata sandi tidak sesuai. Pastikan email Anda telah diverifikasi (${rawMsg}).`
          : "Email, username, atau kata sandi tidak sesuai. Jika baru mendaftar, pastikan email Anda telah diverifikasi.";
      }
      if (hasServerMessage) {
        return rawMsg!;
      }
    }
    if (errKind === "INVALID_ARGUMENT" || errKind === "ARGUMENT") {
      return hasServerMessage
        ? rawMsg!
        : "Data yang dimasukkan belum lengkap atau format tidak sesuai.";
    }
    if (errKind === "CONFLICT" || errKind === "UNPROCESSABLE_ENTITY" || errKind === "NOT_FOUND") {
      if (hasServerMessage) {
        return rawMsg!;
      }
    }
    if (hasServerMessage) {
      return rawMsg!;
    }
  }

  switch (context) {
    case "mfa":
      return "Kode verifikasi tidak sesuai atau sudah kedaluwarsa.";
    case "register":
      return "Pendaftaran akun belum dapat diproses. Silakan periksa kembali data Anda.";
    case "verify":
      return "Verifikasi email tidak dapat diproses saat ini.";
    case "reset-request":
      return "Permintaan pemulihan belum dapat diproses. Silakan periksa kembali identitas akun Anda.";
    case "reset-confirm":
      return "Token pemulihan tidak sesuai atau kata sandi baru belum memenuhi syarat keamanan.";
    default:
      return "Email, username, atau kata sandi tidak sesuai. Silakan periksa kembali.";
  }
}

export const IAM_MIN_PASSWORD_LENGTH = 12;
export const IAM_MAX_PASSWORD_LENGTH = 128;
export const IAM_USERNAME_PATTERN = /^[a-z0-9][a-z0-9._-]{1,62}[a-z0-9]$/;

export type PasswordChecklist = {
  minLength: boolean;
  maxLength: boolean;
  hasUpperAndLower: boolean;
  hasDigit: boolean;
  hasSymbol: boolean;
  noIdentityTerms: boolean;
};

export function validateIamUsername(raw: string): string | null {
  const normalized = raw.trim().toLowerCase();
  if (!normalized) {
    return "Masukkan username Anda.";
  }
  if (!IAM_USERNAME_PATTERN.test(normalized)) {
    return "Username 3-64 karakter: diawali & diakhiri huruf kecil/angka, boleh memakai titik, garis bawah, atau tanda hubung.";
  }
  return null;
}

function containsForbiddenIdentityTerms(
  password: string,
  forbiddenTerms: string[],
): boolean {
  const lower = password.toLowerCase();
  for (const rawTerm of forbiddenTerms) {
    const term = rawTerm.trim().toLowerCase();
    if (term.length >= 3 && lower.includes(term)) {
      return true;
    }
    const tokens = term.split(/[@._\-\s]+/);
    for (const tok of tokens) {
      if (tok.length >= 3 && lower.includes(tok)) {
        return true;
      }
    }
  }
  return false;
}

export function getPasswordChecklist(
  password: string,
  forbiddenTerms: string[] = [],
): PasswordChecklist {
  const runeCount = Array.from(password).length;
  return {
    minLength: runeCount >= IAM_MIN_PASSWORD_LENGTH,
    maxLength: runeCount <= IAM_MAX_PASSWORD_LENGTH,
    hasUpperAndLower: /[A-Z]/.test(password) && /[a-z]/.test(password),
    hasDigit: /\d/.test(password),
    hasSymbol: /[^A-Za-z0-9\s]/.test(password),
    noIdentityTerms: !containsForbiddenIdentityTerms(password, forbiddenTerms),
  };
}

export function validateIamPassword(
  password: string,
  forbiddenTerms: string[] = [],
): string | null {
  if (!password) {
    return "Masukkan kata sandi akun Anda.";
  }
  if (password.includes("\u0000")) {
    return "Kata sandi mengandung karakter tidak valid.";
  }
  const checks = getPasswordChecklist(password, forbiddenTerms);
  if (!checks.minLength) {
    return `Kata sandi minimal ${IAM_MIN_PASSWORD_LENGTH} karakter.`;
  }
  if (!checks.maxLength) {
    return `Kata sandi maksimal ${IAM_MAX_PASSWORD_LENGTH} karakter.`;
  }
  if (!checks.hasUpperAndLower || !checks.hasDigit || !checks.hasSymbol) {
    return "Kata sandi wajib memuat huruf besar, huruf kecil, angka, dan simbol.";
  }
  if (!checks.noIdentityTerms) {
    return "Kata sandi tidak boleh memuat bagian dari nama, email, atau username Anda.";
  }
  return null;
}

export function getPasswordStrength(value: string): {
  score: number;
  label: string;
  color: string;
} {
  if (!value) {
    return { score: 0, label: "", color: "#dde3e8" };
  }
  let score = 0;
  if (Array.from(value).length >= IAM_MIN_PASSWORD_LENGTH) score += 1;
  if (/[A-Z]/.test(value) && /[a-z]/.test(value)) score += 1;
  if (/\d/.test(value)) score += 1;
  if (/[^A-Za-z0-9\s]/.test(value)) score += 1;

  if (score <= 1) {
    return { score: 25, label: "Lemah", color: "#da414a" };
  }
  if (score === 2) {
    return { score: 50, label: "Cukup", color: "#d97706" };
  }
  if (score === 3) {
    return { score: 75, label: "Baik", color: "#133554" };
  }
  return { score: 100, label: "Kuat", color: "#067647" };
}
