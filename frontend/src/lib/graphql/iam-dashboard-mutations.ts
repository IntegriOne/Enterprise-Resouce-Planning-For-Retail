/**
 * Operasi Mutasi GraphQL terketik untuk Dashboard IAM (`/identity`).
 * Seluruh mutasi menyertakan X-CSRF-Token terbaru via cookie HttpOnly
 * serta membuka payload MutationResult secara konsisten.
 */

import {
  AuthRequestError,
  executeGraphQL,
  fetchBrowserSession,
  unwrapMutationPayload,
  type MutationResult,
} from "./iam-browser-auth.ts";
import type {
  AdminCreateUserInput,
  AssignGroupRoleFormInput,
  AssignUserRoleFormInput,
  CreateGroupFormInput,
  CreateRoleFormInput,
  GroupMemberEdge,
  GroupRoleEdge,
  IamGroup,
  IamRole,
  IamUser,
  IamUserIdentity,
  RolePermissionEdge,
  SessionEndReason,
  UpdateGroupFormInput,
  UpdateRoleFormInput,
  UserRoleEdge,
} from "./iam-types.ts";

const RESULT_FIELDS = "errors { type message } was_successful";

const USER_SUMMARY_FIELDS = `
  id _id kind status display_name avatar_url locale timezone
  last_login_at created_at updated_at deleted_at auth_epoch perm_revision
`;

const ROLE_FIELDS = `
  id _id code name description management_domain
  default_redirect_path redirect_priority is_system is_active created_at updated_at
`;

const GROUP_FIELDS = `
  id _id code name description is_active created_at updated_at
`;

const PERM_FIELDS = `
  id _id code resource action description is_sensitive created_at
`;

async function resolveCsrf(provided?: string | null): Promise<string | null> {
  if (provided) return provided;
  try {
    const sess = await fetchBrowserSession();
    return sess.csrfToken;
  } catch {
    return null;
  }
}

async function runMutation<TPayload extends Partial<MutationResult>>(
  mutation: string,
  variables: Record<string, unknown> | undefined,
  field: string,
  csrfToken?: string | null,
  fallbackMessage = "Mutasi IAM belum dapat diproses",
): Promise<TPayload> {
  const token = await resolveCsrf(csrfToken);
  const { data } = await executeGraphQL<Record<string, TPayload | null | undefined>>(
    mutation,
    variables,
    token,
  );
  return unwrapMutationPayload(data?.[field], 200, fallbackMessage);
}

const LOGOUT_MUTATION = `
  mutation IdentityUserLogout {
    userLogout { logged_out ${RESULT_FIELDS} }
  }
`;

export async function submitUserLogout(
  csrfToken?: string | null,
): Promise<boolean> {
  const res = await runMutation<Partial<MutationResult> & { logged_out?: boolean | null }>(
    LOGOUT_MUTATION,
    undefined,
    "userLogout",
    csrfToken,
    "Gagal mengakhiri sesi aktif",
  );
  return Boolean(res.logged_out);
}

// ============================================================================
// Manajemen User & Identitas
// ============================================================================

const CREATE_EMPLOYEE_USER_MUTATION = `
  mutation AdminCreateEmployeeUser($input: CreateEmployeeUserInput!) {
    createEmployeeUser(input: $input) {
      user { ${USER_SUMMARY_FIELDS} }
      ${RESULT_FIELDS}
    }
  }
`;

const CREATE_VENDOR_USER_MUTATION = `
  mutation AdminCreateVendorUser($input: CreateVendorUserInput!) {
    createVendorUser(input: $input) {
      user { ${USER_SUMMARY_FIELDS} }
      party_contact { id _id party_id iam_user_id full_name contact_role portal_access_status }
      ${RESULT_FIELDS}
    }
  }
`;

const CREATE_SUPPLIER_USER_MUTATION = `
  mutation AdminCreateSupplierUser($input: CreateSupplierUserInput!) {
    createSupplierUser(input: $input) {
      user { ${USER_SUMMARY_FIELDS} }
      party_contact { id _id party_id iam_user_id full_name contact_role portal_access_status }
      ${RESULT_FIELDS}
    }
  }
`;

const CREATE_CUSTOMER_USER_MUTATION = `
  mutation AdminCreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      user { ${USER_SUMMARY_FIELDS} }
      redirectLocation
      ${RESULT_FIELDS}
    }
  }
`;

function buildPartnerCreateInput(input: AdminCreateUserInput) {
  const partyId = input.party_id?.trim() ?? "";
  if (!partyId) {
    throw new AuthRequestError(
      400,
      "ARGUMENT",
      "Target mitra (Party ID Enterprise Foundation) wajib dipilih untuk akun Vendor atau Supplier.",
    );
  }
  return {
    party_id: partyId,
    ...(input.contact_id?.trim() ? { contact_id: input.contact_id.trim() } : {}),
    ...(input.contact_role?.trim() ? { contact_role: input.contact_role.trim() } : {}),
    ...(input.job_title?.trim() ? { job_title: input.job_title.trim() } : {}),
    ...(input.phone?.trim() ? { phone: input.phone.trim() } : {}),
    ...(typeof input.is_primary_contact === "boolean"
      ? { is_primary_contact: input.is_primary_contact }
      : {}),
    full_name: input.full_name.trim(),
    email: input.email.trim(),
    username: input.username.trim().toLowerCase(),
    password: input.password,
  };
}

export async function submitAdminCreateUser(
  input: AdminCreateUserInput,
  csrfToken?: string | null,
): Promise<IamUser | null> {
  const baseInput = {
    full_name: input.full_name.trim(),
    email: input.email.trim(),
    username: input.username.trim().toLowerCase(),
    password: input.password,
  };

  switch (input.kind) {
    case "EMPLOYEE": {
      const workerId = input.worker_id?.trim();
      const requestID = input.provisioning_request_id?.trim();
      if (!workerId || !requestID) throw new Error("Pilih pengajuan akun dari HRM.");
      const empInput = {
        ...baseInput,
        worker_id: workerId,
        provisioning_request_id: requestID,
      };
      const payload = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        CREATE_EMPLOYEE_USER_MUTATION,
        { input: empInput },
        "createEmployeeUser",
        csrfToken,
      );
      return payload.user ?? null;
    }
    case "VENDOR": {
      const payload = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        CREATE_VENDOR_USER_MUTATION,
        { input: buildPartnerCreateInput(input) },
        "createVendorUser",
        csrfToken,
      );
      return payload.user ?? null;
    }
    case "SUPPLIER": {
      const payload = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        CREATE_SUPPLIER_USER_MUTATION,
        { input: buildPartnerCreateInput(input) },
        "createSupplierUser",
        csrfToken,
      );
      return payload.user ?? null;
    }
    case "CUSTOMER": {
      const payload = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        CREATE_CUSTOMER_USER_MUTATION,
        { input: baseInput },
        "createUser",
        csrfToken,
      );
      return payload.user ?? null;
    }
    default:
      throw new AuthRequestError(400, "ARGUMENT", "Kategori akun tidak didukung");
  }
}

export async function submitUserLifecycleAction(
  action: "activate" | "suspend" | "disable" | "softDelete",
  userId: string,
  csrfToken?: string | null,
): Promise<IamUser | boolean | null> {
  switch (action) {
    case "activate": {
      const p = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        `mutation ActivateUser($user_id: ID!) { activateUser(user_id: $user_id) { user { ${USER_SUMMARY_FIELDS} } ${RESULT_FIELDS} } }`,
        { user_id: userId },
        "activateUser",
        csrfToken,
      );
      return p.user ?? null;
    }
    case "suspend": {
      const p = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        `mutation SuspendUser($user_id: ID!) { suspendUser(user_id: $user_id) { user { ${USER_SUMMARY_FIELDS} } ${RESULT_FIELDS} } }`,
        { user_id: userId },
        "suspendUser",
        csrfToken,
      );
      return p.user ?? null;
    }
    case "disable": {
      const p = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
        `mutation DisableUser($user_id: ID!) { disableUser(user_id: $user_id) { user { ${USER_SUMMARY_FIELDS} } ${RESULT_FIELDS} } }`,
        { user_id: userId },
        "disableUser",
        csrfToken,
      );
      return p.user ?? null;
    }
    case "softDelete": {
      const p = await runMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
        `mutation SoftDeleteUser($user_id: ID!) { softDeleteUser(user_id: $user_id) { deleted ${RESULT_FIELDS} } }`,
        { user_id: userId },
        "softDeleteUser",
        csrfToken,
      );
      return Boolean(p.deleted);
    }
  }
}

export async function submitSetPrimaryUserIdentity(
  userId: string,
  identityId: string,
  csrfToken?: string | null,
): Promise<IamUserIdentity | null> {
  const p = await runMutation<
    Partial<MutationResult> & { user_identity?: IamUserIdentity | null }
  >(
    `mutation SetPrimaryUserIdentity($user_id: ID!, $identity_id: ID!) {
      setPrimaryUserIdentity(user_id: $user_id, identity_id: $identity_id) {
        user_identity { id _id user_id kind value is_verified is_primary verified_at created_at updated_at }
        ${RESULT_FIELDS}
      }
    }`,
    { user_id: userId, identity_id: identityId },
    "setPrimaryUserIdentity",
    csrfToken,
  );
  return p.user_identity ?? null;
}

export async function submitRemoveUserIdentity(
  userId: string,
  identityId: string,
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { removed?: boolean | null }>(
    `mutation RemoveUserIdentity($user_id: ID!, $identity_id: ID!) {
      removeUserIdentity(user_id: $user_id, identity_id: $identity_id) {
        removed
        ${RESULT_FIELDS}
      }
    }`,
    { user_id: userId, identity_id: identityId },
    "removeUserIdentity",
    csrfToken,
  );
  return Boolean(p.removed);
}

// ============================================================================
// Manajemen Sesi
// ============================================================================

export async function submitRevokeSession(
  sessionId: string,
  reason: SessionEndReason = "ADMIN_REVOKED",
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { revoked?: boolean | null }>(
    `mutation RevokeSession($session_id: ID!, $reason: SessionEndReason) {
      revokeSession(session_id: $session_id, reason: $reason) {
        revoked
        ${RESULT_FIELDS}
      }
    }`,
    { session_id: sessionId, reason },
    "revokeSession",
    csrfToken,
  );
  return Boolean(p.revoked);
}

export async function submitRevokeAllUserSessions(
  userId: string,
  reason: SessionEndReason = "ADMIN_REVOKED",
  csrfToken?: string | null,
): Promise<number> {
  const p = await runMutation<
    Partial<MutationResult> & { revoked_count?: number | null }
  >(
    `mutation RevokeAllUserSessions($user_id: ID!, $reason: SessionEndReason) {
      revokeAllUserSessions(user_id: $user_id, reason: $reason) {
        revoked_count
        ${RESULT_FIELDS}
      }
    }`,
    { user_id: userId, reason },
    "revokeAllUserSessions",
    csrfToken,
  );
  return p.revoked_count ?? 0;
}

// ============================================================================
// Manajemen Role & Permission RBAC
// ============================================================================

export async function submitCreateRole(
  input: CreateRoleFormInput,
  csrfToken?: string | null,
): Promise<IamRole | null> {
  const p = await runMutation<Partial<MutationResult> & { role?: IamRole | null }>(
    `mutation CreateRole($input: CreateRoleInput!) {
      createRole(input: $input) {
        role { ${ROLE_FIELDS} }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        code: input.code.trim(),
        name: input.name.trim(),
        description: input.description?.trim() || null,
        management_domain: input.management_domain.trim(),
        default_redirect_path: input.default_redirect_path.trim(),
        redirect_priority: input.redirect_priority ?? 50,
        is_system: Boolean(input.is_system),
      },
    },
    "createRole",
    csrfToken,
  );
  return p.role ?? null;
}

export async function submitUpdateRoleDetails(
  input: UpdateRoleFormInput,
  csrfToken?: string | null,
): Promise<IamRole | null> {
  const p = await runMutation<Partial<MutationResult> & { role?: IamRole | null }>(
    `mutation UpdateRoleDetails($input: UpdateRoleDetailsInput!) {
      updateRoleDetails(input: $input) {
        role { ${ROLE_FIELDS} }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        id: input.id,
        name: input.name.trim(),
        description: input.description?.trim() || null,
        management_domain: input.management_domain.trim(),
        default_redirect_path: input.default_redirect_path.trim(),
        redirect_priority: input.redirect_priority ?? 50,
      },
    },
    "updateRoleDetails",
    csrfToken,
  );
  return p.role ?? null;
}

export async function submitRoleLifecycleAction(
  action: "activate" | "deactivate" | "delete",
  roleId: string,
  csrfToken?: string | null,
): Promise<IamRole | boolean | null> {
  switch (action) {
    case "activate": {
      const p = await runMutation<Partial<MutationResult> & { role?: IamRole | null }>(
        `mutation ActivateRole($id: ID!) { activateRole(id: $id) { role { ${ROLE_FIELDS} } ${RESULT_FIELDS} } }`,
        { id: roleId },
        "activateRole",
        csrfToken,
      );
      return p.role ?? null;
    }
    case "deactivate": {
      const p = await runMutation<Partial<MutationResult> & { role?: IamRole | null }>(
        `mutation DeactivateRole($id: ID!) { deactivateRole(id: $id) { role { ${ROLE_FIELDS} } ${RESULT_FIELDS} } }`,
        { id: roleId },
        "deactivateRole",
        csrfToken,
      );
      return p.role ?? null;
    }
    case "delete": {
      const p = await runMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
        `mutation DeleteRole($id: ID!) { deleteRole(id: $id) { deleted ${RESULT_FIELDS} } }`,
        { id: roleId },
        "deleteRole",
        csrfToken,
      );
      return Boolean(p.deleted);
    }
  }
}

export async function submitAssignRolePermission(
  roleId: string,
  permissionId: string,
  csrfToken?: string | null,
): Promise<RolePermissionEdge | null> {
  const p = await runMutation<
    Partial<MutationResult> & { role_permission_edge?: RolePermissionEdge | null }
  >(
    `mutation AssignRolePermission($role_id: ID!, $permission_id: ID!) {
      assignRolePermission(role_id: $role_id, permission_id: $permission_id) {
        role_permission_edge { cursor role_id created_at node { ${PERM_FIELDS} } }
        ${RESULT_FIELDS}
      }
    }`,
    { role_id: roleId, permission_id: permissionId },
    "assignRolePermission",
    csrfToken,
  );
  return p.role_permission_edge ?? null;
}

export async function submitRevokeRolePermission(
  roleId: string,
  permissionId: string,
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { revoked?: boolean | null }>(
    `mutation RevokeRolePermission($role_id: ID!, $permission_id: ID!) {
      revokeRolePermission(role_id: $role_id, permission_id: $permission_id) {
        revoked
        ${RESULT_FIELDS}
      }
    }`,
    { role_id: roleId, permission_id: permissionId },
    "revokeRolePermission",
    csrfToken,
  );
  return Boolean(p.revoked);
}

export async function submitAssignUserRole(
  input: AssignUserRoleFormInput,
  csrfToken?: string | null,
): Promise<UserRoleEdge | null> {
  if (input.scope_type !== "GLOBAL" && !input.scope_id?.trim()) {
    throw new AuthRequestError(
      400,
      "ARGUMENT",
      "Scope ID wajib diisi untuk jenis scope selain GLOBAL",
    );
  }
  const p = await runMutation<
    Partial<MutationResult> & { user_role_edge?: UserRoleEdge | null }
  >(
    `mutation AssignUserRole($input: AssignUserRoleInput!) {
      assignUserRole(input: $input) {
        user_role_edge {
          cursor
          assignment_id
          user_id
          scope { type id }
          granted_by
          expires_at
          created_at
          node { ${ROLE_FIELDS} }
        }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        user_id: input.user_id,
        role_id: input.role_id,
        scope_type: input.scope_type,
        scope_id: input.scope_type === "GLOBAL" ? null : input.scope_id?.trim(),
        expires_at: input.expires_at?.trim() || null,
      },
    },
    "assignUserRole",
    csrfToken,
  );
  return p.user_role_edge ?? null;
}

export async function submitRevokeUserRole(
  assignmentId: string,
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { revoked?: boolean | null }>(
    `mutation RevokeUserRole($assignment_id: ID!) {
      revokeUserRole(assignment_id: $assignment_id) {
        revoked
        ${RESULT_FIELDS}
      }
    }`,
    { assignment_id: assignmentId },
    "revokeUserRole",
    csrfToken,
  );
  return Boolean(p.revoked);
}

// ============================================================================
// Manajemen Group & Pewarisan Peran
// ============================================================================

export async function submitCreateGroup(
  input: CreateGroupFormInput,
  csrfToken?: string | null,
): Promise<IamGroup | null> {
  const p = await runMutation<Partial<MutationResult> & { group?: IamGroup | null }>(
    `mutation CreateGroup($input: CreateGroupInput!) {
      createGroup(input: $input) {
        group { ${GROUP_FIELDS} }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        code: input.code.trim(),
        name: input.name.trim(),
        description: input.description?.trim() || null,
      },
    },
    "createGroup",
    csrfToken,
  );
  return p.group ?? null;
}

export async function submitUpdateGroupDetails(
  input: UpdateGroupFormInput,
  csrfToken?: string | null,
): Promise<IamGroup | null> {
  const p = await runMutation<Partial<MutationResult> & { group?: IamGroup | null }>(
    `mutation UpdateGroupDetails($input: UpdateGroupDetailsInput!) {
      updateGroupDetails(input: $input) {
        group { ${GROUP_FIELDS} }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        id: input.id,
        name: input.name.trim(),
        description: input.description?.trim() || null,
      },
    },
    "updateGroupDetails",
    csrfToken,
  );
  return p.group ?? null;
}

export async function submitGroupLifecycleAction(
  action: "activate" | "deactivate" | "delete",
  groupId: string,
  csrfToken?: string | null,
): Promise<IamGroup | boolean | null> {
  switch (action) {
    case "activate": {
      const p = await runMutation<Partial<MutationResult> & { group?: IamGroup | null }>(
        `mutation ActivateGroup($id: ID!) { activateGroup(id: $id) { group { ${GROUP_FIELDS} } ${RESULT_FIELDS} } }`,
        { id: groupId },
        "activateGroup",
        csrfToken,
      );
      return p.group ?? null;
    }
    case "deactivate": {
      const p = await runMutation<Partial<MutationResult> & { group?: IamGroup | null }>(
        `mutation DeactivateGroup($id: ID!) { deactivateGroup(id: $id) { group { ${GROUP_FIELDS} } ${RESULT_FIELDS} } }`,
        { id: groupId },
        "deactivateGroup",
        csrfToken,
      );
      return p.group ?? null;
    }
    case "delete": {
      const p = await runMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
        `mutation DeleteGroup($id: ID!) { deleteGroup(id: $id) { deleted ${RESULT_FIELDS} } }`,
        { id: groupId },
        "deleteGroup",
        csrfToken,
      );
      return Boolean(p.deleted);
    }
  }
}

export async function submitAddGroupMember(
  groupId: string,
  userId: string,
  csrfToken?: string | null,
): Promise<GroupMemberEdge | null> {
  const p = await runMutation<
    Partial<MutationResult> & { group_member_edge?: GroupMemberEdge | null }
  >(
    `mutation AddGroupMember($group_id: ID!, $user_id: ID!) {
      addGroupMember(group_id: $group_id, user_id: $user_id) {
        group_member_edge { cursor group_id created_at node { ${USER_SUMMARY_FIELDS} } }
        ${RESULT_FIELDS}
      }
    }`,
    { group_id: groupId, user_id: userId },
    "addGroupMember",
    csrfToken,
  );
  return p.group_member_edge ?? null;
}

export async function submitRemoveGroupMember(
  groupId: string,
  userId: string,
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { removed?: boolean | null }>(
    `mutation RemoveGroupMember($group_id: ID!, $user_id: ID!) {
      removeGroupMember(group_id: $group_id, user_id: $user_id) {
        removed
        ${RESULT_FIELDS}
      }
    }`,
    { group_id: groupId, user_id: userId },
    "removeGroupMember",
    csrfToken,
  );
  return Boolean(p.removed);
}

export async function submitAssignGroupRole(
  input: AssignGroupRoleFormInput,
  csrfToken?: string | null,
): Promise<GroupRoleEdge | null> {
  if ((input.scope_type as string) === "SELF") {
    throw new AuthRequestError(
      400,
      "ARGUMENT",
      "Scope SELF tidak diizinkan pada penugasan peran grup",
    );
  }
  if (input.scope_type !== "GLOBAL" && !input.scope_id?.trim()) {
    throw new AuthRequestError(
      400,
      "ARGUMENT",
      "Scope ID wajib diisi untuk jenis scope selain GLOBAL",
    );
  }
  const p = await runMutation<
    Partial<MutationResult> & { group_role_edge?: GroupRoleEdge | null }
  >(
    `mutation AssignGroupRole($input: AssignGroupRoleInput!) {
      assignGroupRole(input: $input) {
        group_role_edge {
          cursor
          assignment_id
          group_id
          scope { type id }
          granted_by
          expires_at
          created_at
          node { ${ROLE_FIELDS} }
        }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        group_id: input.group_id,
        role_id: input.role_id,
        scope_type: input.scope_type,
        scope_id: input.scope_type === "GLOBAL" ? null : input.scope_id?.trim(),
        expires_at: input.expires_at?.trim() || null,
      },
    },
    "assignGroupRole",
    csrfToken,
  );
  return p.group_role_edge ?? null;
}

export async function submitRevokeGroupRole(
  assignmentId: string,
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { revoked?: boolean | null }>(
    `mutation RevokeGroupRole($assignment_id: ID!) {
      revokeGroupRole(assignment_id: $assignment_id) {
        revoked
        ${RESULT_FIELDS}
      }
    }`,
    { assignment_id: assignmentId },
    "revokeGroupRole",
    csrfToken,
  );
  return Boolean(p.revoked);
}

// ============================================================================
// Keamanan Mandiri Operator & Step-Up MFA
// ============================================================================

export async function submitVerifyMySessionMfa(
  totpCode: string,
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { verified?: boolean | null }>(
    `mutation VerifyMySessionMfa($totp_code: String!) {
      verifyMySessionMfa(totp_code: $totp_code) {
        verified
        ${RESULT_FIELDS}
      }
    }`,
    { totp_code: totpCode.trim() },
    "verifyMySessionMfa",
    csrfToken,
  );
  return Boolean(p.verified);
}

export async function submitEnrollMyTotp(
  csrfToken?: string | null,
): Promise<{ factorId: string | null; otpAuthUri: string | null }> {
  const p = await runMutation<
    Partial<MutationResult> & { factor_id?: string | null; otp_auth_uri?: string | null }
  >(
    `mutation EnrollMyTotp {
      enrollMyTotp {
        factor_id
        otp_auth_uri
        ${RESULT_FIELDS}
      }
    }`,
    undefined,
    "enrollMyTotp",
    csrfToken,
  );
  return { factorId: p.factor_id ?? null, otpAuthUri: p.otp_auth_uri ?? null };
}

export async function submitConfirmMyTotp(
  code: string,
  csrfToken?: string | null,
): Promise<{ enabled: boolean; recoveryCodes: string[] }> {
  const p = await runMutation<
    Partial<MutationResult> & {
      enabled?: boolean | null;
      recovery_codes?: string[] | null;
    }
  >(
    `mutation ConfirmMyTotp($input: ConfirmTotpInput!) {
      confirmMyTotp(input: $input) {
        enabled
        recovery_codes
        ${RESULT_FIELDS}
      }
    }`,
    { input: { code: code.trim() } },
    "confirmMyTotp",
    csrfToken,
  );
  return {
    enabled: Boolean(p.enabled),
    recoveryCodes: p.recovery_codes ?? [],
  };
}

export async function submitDisableMyTotp(
  csrfToken?: string | null,
): Promise<boolean> {
  const p = await runMutation<Partial<MutationResult> & { disabled?: boolean | null }>(
    `mutation DisableMyTotp {
      disableMyTotp {
        disabled
        ${RESULT_FIELDS}
      }
    }`,
    undefined,
    "disableMyTotp",
    csrfToken,
  );
  return Boolean(p.disabled);
}

export async function submitRegenerateMyRecoveryCodes(
  csrfToken?: string | null,
): Promise<string[]> {
  const p = await runMutation<
    Partial<MutationResult> & { recovery_codes?: string[] | null }
  >(
    `mutation RegenerateMyRecoveryCodes {
      regenerateMyRecoveryCodes {
        recovery_codes
        ${RESULT_FIELDS}
      }
    }`,
    undefined,
    "regenerateMyRecoveryCodes",
    csrfToken,
  );
  return p.recovery_codes ?? [];
}

export async function submitUpdateMyProfile(
  input: { display_name: string; locale?: string | null; timezone?: string | null },
  csrfToken?: string | null,
): Promise<IamUser | null> {
  const p = await runMutation<Partial<MutationResult> & { user?: IamUser | null }>(
    `mutation UpdateMyProfile($input: UpdateMyProfileInput!) {
      updateMyProfile(input: $input) {
        user { ${USER_SUMMARY_FIELDS} }
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        display_name: input.display_name.trim(),
        locale: input.locale?.trim() || null,
        timezone: input.timezone?.trim() || null,
      },
    },
    "updateMyProfile",
    csrfToken,
  );
  return p.user ?? null;
}

export async function submitChangeUserPassword(
  currentPassword: string,
  newPassword: string,
  csrfToken?: string | null,
): Promise<{ changed: boolean; sessionsRevoked: number }> {
  const p = await runMutation<
    Partial<MutationResult> & {
      changed?: boolean | null;
      sessions_revoked?: number | null;
    }
  >(
    `mutation ChangeUserPassword($input: ChangeUserPasswordInput!) {
      changeUserPassword(input: $input) {
        changed
        sessions_revoked
        ${RESULT_FIELDS}
      }
    }`,
    {
      input: {
        current_password: currentPassword,
        new_password: newPassword,
      },
    },
    "changeUserPassword",
    csrfToken,
  );
  return {
    changed: Boolean(p.changed),
    sessionsRevoked: p.sessions_revoked ?? 0,
  };
}
