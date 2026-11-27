/**
 * Operasi Query GraphQL terketik untuk Dashboard IAM (`/identity`).
 * Query dipecah ringan agar menjaga batas AST depth dan complexity Gateway (<= 250)
 * serta tetap fail-safe apabila sebagian modul izin tidak dimiliki operator.
 */

import { AuthRequestError, executeGraphQL } from "./iam-browser-auth.ts";
import type {
  AccessSnapshot,
  GroupConnection,
  GroupMemberConnection,
  GroupRoleConnection,
  IamGroup,
  IamRole,
  IamUser,
  PermissionConnection,
  RelayCursorVars,
  RoleConnection,
  RolePermissionConnection,
  SessionConnection,
  UserConnection,
  UserIdentityConnection,
  UserKind,
  UserRoleConnection,
  UserStatus,
} from "./iam-types.ts";

const PAGE_INFO_FIELDS = "endCursor hasNextPage hasPreviousPage startCursor";

const USER_FIELDS = `
  id _id kind status display_name avatar_url locale timezone
  last_login_at created_at updated_at deleted_at auth_epoch perm_revision
`;

const IDENTITY_FIELDS = `
  id _id user_id kind value is_verified is_primary verified_at created_at updated_at
`;

const SESSION_FIELDS = `
  id _id user_id mfa_verified mfa_verified_at ip_address user_agent
  authenticated_at last_active_at idle_expires_at absolute_expires_at
  ended_at end_reason created_at generation
`;

const ROLE_FIELDS = `
  id _id code name description management_domain
  default_redirect_path redirect_priority is_system is_active created_at updated_at
`;

const PERMISSION_FIELDS = `
  id _id code resource action description is_sensitive created_at
`;

const GROUP_FIELDS = `
  id _id code name description is_active created_at updated_at
`;

const ME_AND_SNAPSHOT_QUERY = `
  query IdentityOperatorContext {
    me { ${USER_FIELDS} }
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
      permissions {
        permission_code
        is_sensitive
        scope { type id }
      }
      synced_at_unix
      fresh_until_unix
    }
  }
`;

const USERS_COUNT_QUERY = `query IdentityUsersCount { users(first: 1) { total_count } }`;
const ROLES_COUNT_QUERY = `query IdentityRolesCount { roles(first: 1, active_only: false) { total_count } }`;
const GROUPS_COUNT_QUERY = `query IdentityGroupsCount { groups(first: 1, active_only: false) { total_count } }`;
const PERMS_COUNT_QUERY = `query IdentityPermissionsCount { permissions(first: 1) { total_count } }`;

const USERS_LIST_QUERY = `
  query IdentityUsers(
    $after: String
    $before: String
    $first: Int
    $last: Int
    $query: String
    $kind: UserKind
    $status: UserStatus
    $include_deleted: Boolean
  ) {
    users(
      after: $after
      before: $before
      first: $first
      last: $last
      query: $query
      kind: $kind
      status: $status
      include_deleted: $include_deleted
    ) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${USER_FIELDS} }
    }
  }
`;

const USER_DETAIL_QUERY = `
  query IdentityUserDetail($id: ID!) {
    user(id: $id) { ${USER_FIELDS} }
  }
`;

const USER_IDENTITIES_QUERY = `
  query IdentityUserIdentities($user_id: ID!, $first: Int, $after: String) {
    userIdentities(user_id: $user_id, first: $first, after: $after) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${IDENTITY_FIELDS} }
    }
  }
`;

const USER_ROLES_QUERY = `
  query IdentityUserRoles($user_id: ID!, $include_expired: Boolean, $first: Int, $after: String) {
    userRoles(user_id: $user_id, include_expired: $include_expired, first: $first, after: $after) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      edges {
        cursor
        assignment_id
        user_id
        scope { type id }
        granted_by
        expires_at
        created_at
        node { ${ROLE_FIELDS} }
      }
    }
  }
`;

const USER_EFFECTIVE_ROLES_QUERY = `
  query IdentityUserEffectiveRoles($user_id: ID!, $first: Int) {
    userEffectiveRoles(user_id: $user_id, first: $first) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${ROLE_FIELDS} }
    }
  }
`;

const USER_GROUPS_QUERY = `
  query IdentityUserGroups($user_id: ID!, $first: Int) {
    userGroups(user_id: $user_id, first: $first) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${GROUP_FIELDS} }
    }
  }
`;

const USER_SESSIONS_QUERY = `
  query IdentityUserSessions($user_id: ID!, $active_only: Boolean, $first: Int, $after: String) {
    userSessions(user_id: $user_id, active_only: $active_only, first: $first, after: $after) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${SESSION_FIELDS} }
    }
  }
`;

const MY_SESSIONS_QUERY = `
  query IdentityMySessions(
    $after: String
    $before: String
    $first: Int
    $last: Int
    $active_only: Boolean
  ) {
    mySessions(
      after: $after
      before: $before
      first: $first
      last: $last
      active_only: $active_only
    ) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${SESSION_FIELDS} }
    }
  }
`;

const ROLES_LIST_QUERY = `
  query IdentityRoles(
    $after: String
    $before: String
    $first: Int
    $last: Int
    $query: String
    $management_domain: String
    $active_only: Boolean
  ) {
    roles(
      after: $after
      before: $before
      first: $first
      last: $last
      query: $query
      management_domain: $management_domain
      active_only: $active_only
    ) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${ROLE_FIELDS} }
    }
  }
`;

const ROLE_DETAIL_QUERY = `
  query IdentityRoleDetail($id: ID!) {
    role(id: $id) { ${ROLE_FIELDS} }
  }
`;

const ROLE_PERMISSIONS_QUERY = `
  query IdentityRolePermissions($role_id: ID!, $first: Int, $after: String, $last: Int, $before: String) {
    rolePermissions(role_id: $role_id, first: $first, after: $after, last: $last, before: $before) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${PERMISSION_FIELDS} }
    }
  }
`;

const PERMISSIONS_LIST_QUERY = `
  query IdentityPermissions(
    $after: String
    $before: String
    $first: Int
    $last: Int
    $query: String
    $resource: String
    $is_sensitive: Boolean
  ) {
    permissions(
      after: $after
      before: $before
      first: $first
      last: $last
      query: $query
      resource: $resource
      is_sensitive: $is_sensitive
    ) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${PERMISSION_FIELDS} }
    }
  }
`;

const GROUPS_LIST_QUERY = `
  query IdentityGroups(
    $after: String
    $before: String
    $first: Int
    $last: Int
    $query: String
    $active_only: Boolean
  ) {
    groups(
      after: $after
      before: $before
      first: $first
      last: $last
      query: $query
      active_only: $active_only
    ) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      nodes { ${GROUP_FIELDS} }
    }
  }
`;

const GROUP_DETAIL_QUERY = `
  query IdentityGroupDetail($id: ID!) {
    group(id: $id) { ${GROUP_FIELDS} }
  }
`;

const GROUP_MEMBERS_QUERY = `
  query IdentityGroupMembers($group_id: ID!, $first: Int, $after: String, $last: Int, $before: String) {
    groupMembers(group_id: $group_id, first: $first, after: $after, last: $last, before: $before) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      edges {
        cursor
        group_id
        created_at
        node { ${USER_FIELDS} }
      }
    }
  }
`;

const GROUP_ROLES_QUERY = `
  query IdentityGroupRoles($group_id: ID!, $include_expired: Boolean, $first: Int, $after: String) {
    groupRoles(group_id: $group_id, include_expired: $include_expired, first: $first, after: $after) {
      total_count
      pageInfo { ${PAGE_INFO_FIELDS} }
      edges {
        cursor
        assignment_id
        group_id
        scope { type id }
        granted_by
        expires_at
        created_at
        node { ${ROLE_FIELDS} }
      }
    }
  }
`;

const CHECK_USERNAME_QUERY = `
  query CheckUsernameAvailability($query: String!) {
    users(query: $query, first: 10, include_deleted: true) {
      nodes {
        id
        _id
        kind
        status
        display_name
        identities(first: 5) {
          nodes {
            kind
            value
            is_verified
          }
        }
      }
    }
  }
`;

export type UsernameAvailabilityResult = {
  available: boolean;
  normalizedUsername: string;
  existingUser?: {
    id: string;
    _id: string;
    kind: UserKind;
    status: UserStatus;
    display_name: string;
  };
};

// Memeriksa ketersediaan username melalui query GraphQL users(query) untuk operator terautentikasi.
export async function checkUsernameAvailabilityViaGraphQL(
  rawUsername: string,
): Promise<UsernameAvailabilityResult> {
  const normalizedUsername = rawUsername.trim().toLowerCase();
  if (!normalizedUsername) {
    return { available: false, normalizedUsername };
  }

  const { data } = await executeGraphQL<{
    users?: {
      nodes: Array<
        IamUser & {
          identities?: {
            nodes: Array<{ kind: string; value: string; is_verified: boolean }>;
          };
        }
      >;
    };
  }>(CHECK_USERNAME_QUERY, { query: normalizedUsername });

  const matchedUser = (data.users?.nodes ?? []).find((userNode) =>
    (userNode.identities?.nodes ?? []).some(
      (ident) =>
        ident.kind === "USERNAME" &&
        ident.value.trim().toLowerCase() === normalizedUsername,
    ),
  );

  if (matchedUser) {
    return {
      available: false,
      normalizedUsername,
      existingUser: {
        id: matchedUser.id,
        _id: matchedUser._id,
        kind: matchedUser.kind,
        status: matchedUser.status,
        display_name: matchedUser.display_name,
      },
    };
  }

  return { available: true, normalizedUsername };
}

export async function fetchOperatorContext(): Promise<{
  me: IamUser;
  snapshot: AccessSnapshot;
}> {
  const { data } = await executeGraphQL<{
    me?: IamUser | null;
    myAccessSnapshot?: AccessSnapshot | null;
  }>(ME_AND_SNAPSHOT_QUERY, undefined);

  if (!data.me || !data.myAccessSnapshot) {
    throw new AuthRequestError(
      500,
      "INVALID_RESPONSE",
      "Data profil operator atau snapshot akses tidak tersedia",
    );
  }
  return { me: data.me, snapshot: data.myAccessSnapshot };
}

export type OverviewCounts = {
  usersTotal: number | null;
  rolesTotal: number | null;
  groupsTotal: number | null;
  permissionsTotal: number | null;
};

async function safeCount(
  query: string,
  extractor: (data: Record<string, { total_count?: number } | undefined>) => number | undefined,
): Promise<number | null> {
  try {
    const { data } = await executeGraphQL<
      Record<string, { total_count?: number } | undefined>
    >(query, undefined);
    const val = extractor(data);
    return typeof val === "number" ? val : null;
  } catch (err) {
    if (isUnauthenticatedError(err)) {
      throw err;
    }
    return null;
  }
}

export async function fetchOverviewCounts(): Promise<OverviewCounts> {
  const [usersTotal, rolesTotal, groupsTotal, permissionsTotal] =
    await Promise.all([
      safeCount(USERS_COUNT_QUERY, (d) => d.users?.total_count),
      safeCount(ROLES_COUNT_QUERY, (d) => d.roles?.total_count),
      safeCount(GROUPS_COUNT_QUERY, (d) => d.groups?.total_count),
      safeCount(PERMS_COUNT_QUERY, (d) => d.permissions?.total_count),
    ]);

  return { usersTotal, rolesTotal, groupsTotal, permissionsTotal };
}

export type UsersFilterVars = RelayCursorVars & {
  query?: string | null;
  kind?: UserKind | null;
  status?: UserStatus | null;
  include_deleted?: boolean;
};

export async function fetchUsersConnection(
  vars: UsersFilterVars,
): Promise<UserConnection> {
  const cleanVars: Record<string, unknown> = {
    include_deleted: Boolean(vars.include_deleted),
  };
  if (vars.first) cleanVars.first = vars.first;
  if (vars.after) cleanVars.after = vars.after;
  if (vars.last) cleanVars.last = vars.last;
  if (vars.before) cleanVars.before = vars.before;
  if (vars.query?.trim()) cleanVars.query = vars.query.trim();
  if (vars.kind) cleanVars.kind = vars.kind;
  if (vars.status) cleanVars.status = vars.status;

  const { data } = await executeGraphQL<{ users: UserConnection }>(
    USERS_LIST_QUERY,
    cleanVars,
  );
  return data.users;
}

export async function fetchUserDetail(id: string): Promise<IamUser | null> {
  const { data } = await executeGraphQL<{ user: IamUser | null }>(
    USER_DETAIL_QUERY,
    { id },
  );
  return data.user;
}

export async function fetchUserIdentities(
  userId: string,
  first = 25,
  after?: string | null,
): Promise<UserIdentityConnection> {
  const { data } = await executeGraphQL<{
    userIdentities: UserIdentityConnection;
  }>(USER_IDENTITIES_QUERY, { user_id: userId, first, after: after || null });
  return data.userIdentities;
}

export async function fetchUserRoles(
  userId: string,
  includeExpired = false,
  first = 20,
  after?: string | null,
): Promise<UserRoleConnection> {
  const { data } = await executeGraphQL<{ userRoles: UserRoleConnection }>(
    USER_ROLES_QUERY,
    {
      user_id: userId,
      include_expired: includeExpired,
      first,
      after: after || null,
    },
  );
  return {
    ...data.userRoles,
    nodes: data.userRoles.nodes ?? [],
  };
}

export async function fetchUserEffectiveRoles(
  userId: string,
  first = 30,
): Promise<RoleConnection> {
  const { data } = await executeGraphQL<{
    userEffectiveRoles: RoleConnection;
  }>(USER_EFFECTIVE_ROLES_QUERY, { user_id: userId, first });
  return data.userEffectiveRoles;
}

export async function fetchUserGroups(
  userId: string,
  first = 30,
): Promise<GroupConnection> {
  const { data } = await executeGraphQL<{ userGroups: GroupConnection }>(
    USER_GROUPS_QUERY,
    { user_id: userId, first },
  );
  return data.userGroups;
}

export async function fetchUserSessions(
  userId: string,
  activeOnly = true,
  first = 25,
  after?: string | null,
): Promise<SessionConnection> {
  const { data } = await executeGraphQL<{ userSessions: SessionConnection }>(
    USER_SESSIONS_QUERY,
    {
      user_id: userId,
      active_only: activeOnly,
      first,
      after: after || null,
    },
  );
  return data.userSessions;
}

export async function fetchMySessions(
  vars: RelayCursorVars & { active_only?: boolean },
): Promise<SessionConnection> {
  const { data } = await executeGraphQL<{ mySessions: SessionConnection }>(
    MY_SESSIONS_QUERY,
    {
      first: vars.first ?? null,
      after: vars.after ?? null,
      last: vars.last ?? null,
      before: vars.before ?? null,
      active_only: vars.active_only ?? true,
    },
  );
  return data.mySessions;
}

export type RolesFilterVars = RelayCursorVars & {
  query?: string | null;
  management_domain?: string | null;
  active_only?: boolean;
};

export async function fetchRolesConnection(
  vars: RolesFilterVars,
): Promise<RoleConnection> {
  const cleanVars: Record<string, unknown> = {
    active_only: Boolean(vars.active_only),
  };
  if (vars.first) cleanVars.first = vars.first;
  if (vars.after) cleanVars.after = vars.after;
  if (vars.last) cleanVars.last = vars.last;
  if (vars.before) cleanVars.before = vars.before;
  if (vars.query?.trim()) cleanVars.query = vars.query.trim();
  if (vars.management_domain?.trim()) {
    cleanVars.management_domain = vars.management_domain.trim();
  }

  const { data } = await executeGraphQL<{ roles: RoleConnection }>(
    ROLES_LIST_QUERY,
    cleanVars,
  );
  return data.roles;
}

export async function fetchRoleDetail(id: string): Promise<IamRole | null> {
  const { data } = await executeGraphQL<{ role: IamRole | null }>(
    ROLE_DETAIL_QUERY,
    { id },
  );
  return data.role;
}

export async function fetchRolePermissions(
  roleId: string,
  vars: RelayCursorVars = { first: 40 },
): Promise<RolePermissionConnection> {
  const { data } = await executeGraphQL<{
    rolePermissions: RolePermissionConnection;
  }>(ROLE_PERMISSIONS_QUERY, {
    role_id: roleId,
    first: vars.first ?? null,
    after: vars.after ?? null,
    last: vars.last ?? null,
    before: vars.before ?? null,
  });
  return {
    ...data.rolePermissions,
    edges: data.rolePermissions.edges ?? [],
  };
}

// Mengambil seluruh peran dengan paginasi bertahap agar tetap di bawah batas kompleksitas Gateway.
export async function fetchAllRoles(
  vars: Omit<RolesFilterVars, "first" | "after" | "last" | "before"> = { active_only: true },
): Promise<IamRole[]> {
  const collected: IamRole[] = [];
  let after: string | null = null;
  for (let page = 0; page < 5; page += 1) {
    const conn = await fetchRolesConnection({
      ...vars,
      first: 40,
      after,
    });
    collected.push(...conn.nodes);
    if (!conn.pageInfo.hasNextPage || !conn.pageInfo.endCursor) {
      break;
    }
    after = conn.pageInfo.endCursor;
  }
  return collected;
}

// Mengambil seluruh izin terpasang pada satu peran sambil menjaga kompleksitas tiap kueri (first: 40).
export async function fetchAllRolePermissions(
  roleId: string,
): Promise<RolePermissionConnection["nodes"]> {
  const collected: RolePermissionConnection["nodes"] = [];
  let after: string | null = null;
  for (let page = 0; page < 6; page += 1) {
    const conn = await fetchRolePermissions(roleId, { first: 40, after });
    collected.push(...conn.nodes);
    if (!conn.pageInfo.hasNextPage || !conn.pageInfo.endCursor) {
      break;
    }
    after = conn.pageInfo.endCursor;
  }
  return collected;
}

// Mengambil katalog izin terfilter secara aman di bawah batas kompleksitas Gateway (first: 40).
export async function fetchAllPermissionsCatalog(
  vars: Omit<PermissionsFilterVars, "first" | "after" | "last" | "before"> = {},
): Promise<PermissionConnection["nodes"]> {
  const collected: PermissionConnection["nodes"] = [];
  let after: string | null = null;
  for (let page = 0; page < 6; page += 1) {
    const conn = await fetchPermissionsConnection({
      ...vars,
      first: 40,
      after,
    });
    collected.push(...conn.nodes);
    if (!conn.pageInfo.hasNextPage || !conn.pageInfo.endCursor) {
      break;
    }
    after = conn.pageInfo.endCursor;
  }
  return collected;
}

export type PermissionsFilterVars = RelayCursorVars & {
  query?: string | null;
  resource?: string | null;
  is_sensitive?: boolean | null;
};

export async function fetchPermissionsConnection(
  vars: PermissionsFilterVars,
): Promise<PermissionConnection> {
  const cleanVars: Record<string, unknown> = {};
  if (vars.first) cleanVars.first = vars.first;
  if (vars.after) cleanVars.after = vars.after;
  if (vars.last) cleanVars.last = vars.last;
  if (vars.before) cleanVars.before = vars.before;
  if (vars.query?.trim()) cleanVars.query = vars.query.trim();
  if (vars.resource?.trim()) cleanVars.resource = vars.resource.trim();
  if (typeof vars.is_sensitive === "boolean") {
    cleanVars.is_sensitive = vars.is_sensitive;
  }

  const { data } = await executeGraphQL<{ permissions: PermissionConnection }>(
    PERMISSIONS_LIST_QUERY,
    cleanVars,
  );
  return data.permissions;
}

export type GroupsFilterVars = RelayCursorVars & {
  query?: string | null;
  active_only?: boolean;
};

export async function fetchGroupsConnection(
  vars: GroupsFilterVars,
): Promise<GroupConnection> {
  const cleanVars: Record<string, unknown> = {
    active_only: Boolean(vars.active_only),
  };
  if (vars.first) cleanVars.first = vars.first;
  if (vars.after) cleanVars.after = vars.after;
  if (vars.last) cleanVars.last = vars.last;
  if (vars.before) cleanVars.before = vars.before;
  if (vars.query?.trim()) cleanVars.query = vars.query.trim();

  const { data } = await executeGraphQL<{ groups: GroupConnection }>(
    GROUPS_LIST_QUERY,
    cleanVars,
  );
  return data.groups;
}

export async function fetchGroupDetail(id: string): Promise<IamGroup | null> {
  const { data } = await executeGraphQL<{ group: IamGroup | null }>(
    GROUP_DETAIL_QUERY,
    { id },
  );
  return data.group;
}

export async function fetchGroupMembers(
  groupId: string,
  vars: RelayCursorVars = { first: 20 },
): Promise<GroupMemberConnection> {
  const { data } = await executeGraphQL<{
    groupMembers: GroupMemberConnection;
  }>(GROUP_MEMBERS_QUERY, {
    group_id: groupId,
    first: vars.first ?? null,
    after: vars.after ?? null,
    last: vars.last ?? null,
    before: vars.before ?? null,
  });
  return {
    ...data.groupMembers,
    nodes: data.groupMembers.nodes ?? [],
  };
}

export async function fetchGroupRoles(
  groupId: string,
  includeExpired = false,
  first = 20,
  after?: string | null,
): Promise<GroupRoleConnection> {
  const { data } = await executeGraphQL<{ groupRoles: GroupRoleConnection }>(
    GROUP_ROLES_QUERY,
    {
      group_id: groupId,
      include_expired: includeExpired,
      first,
      after: after || null,
    },
  );
  return {
    ...data.groupRoles,
    nodes: data.groupRoles.nodes ?? [],
  };
}

export function isUnauthenticatedError(err: unknown): boolean {
  if (!(err instanceof AuthRequestError)) return false;
  return (
    err.status === 401 ||
    err.code === "UNAUTHENTICATED" ||
    err.errorType === "AUTHENTICATION"
  );
}

export function isForbiddenError(err: unknown): boolean {
  if (!(err instanceof AuthRequestError)) return false;
  return (
    err.status === 403 ||
    err.code === "FORBIDDEN" ||
    err.code === "PERMISSION_DENIED" ||
    err.errorType === "AUTHORIZATION"
  );
}

export function isStepUpRequiredError(err: unknown): boolean {
  if (!(err instanceof AuthRequestError)) return false;
  const msg = (err.message || "").toLowerCase();
  return (
    (err.errorType === "AUTHORIZATION" ||
      err.code === "FORBIDDEN" ||
      err.status === 403) &&
    (msg.includes("mfa step-up") ||
      msg.includes("employee mfa authentication required") ||
      msg.includes("recent authentication within 10 minutes") ||
      msg.includes("sensitive action"))
  );
}

export function formatIamDashboardError(err: unknown): string {
  if (err instanceof AuthRequestError) {
    if (err.code === "NETWORK") {
      return "Koneksi ke layanan IAM terputus. Periksa jaringan lalu coba lagi.";
    }
    if (err.code === "UNAVAILABLE" || err.status >= 500) {
      return "Layanan otorisasi IAM sedang tidak tersedia sementara (fail-closed).";
    }
    if (isStepUpRequiredError(err)) {
      return "Operasi sensitif ini membutuhkan verifikasi ulang MFA atau autentikasi 10 menit terakhir.";
    }
    if (isForbiddenError(err)) {
      return err.message || "Anda tidak memiliki izin atau scope untuk tindakan ini.";
    }
    if (err.message) {
      return err.message;
    }
  }
  if (err instanceof Error && err.message) {
    return err.message;
  }
  return "Permintaan ke layanan IAM gagal diproses.";
}

export type EmployeeProvisioningOption = {
  id: string;
  worker_id: string;
  worker_number: string;
  legal_entity_id: string;
  requested_at: string;
  status: string;
};

export async function fetchEmployeeProvisioningRequests(after: string | null) {
  const { data } = await executeGraphQL<{
    employeeProvisioningRequests: {
      nodes: EmployeeProvisioningOption[];
      pageInfo: { endCursor: string | null; hasNextPage: boolean };
    };
  }>(
    `query EmployeeProvisioningRequests($after: String) {
      employeeProvisioningRequests(first: 20, after: $after) {
        nodes { id worker_id worker_number legal_entity_id requested_at status }
        pageInfo { endCursor hasNextPage }
      }
    }`,
    { after },
  );
  return data.employeeProvisioningRequests;
}

export async function fetchEmployeeProvisioningDetails(
  workerID: string,
  legalEntityID: string,
) {
  const [workerResult, entityResult] = await Promise.allSettled([
    executeGraphQL<{ worker: { person_party_reference_id?: string | null } | null }>(
      `query ProvisioningWorker($id: ID!) { worker(id: $id) { person_party_reference_id } }`,
      { id: workerID },
    ),
    executeGraphQL<{ legalEntity: { legal_name: string; entity_code: string } | null }>(
      `query ProvisioningLegalEntity($id: ID!) { legalEntity(id: $id) { legal_name entity_code } }`,
      { id: legalEntityID },
    ),
  ]);
  const partyID = workerResult.status === "fulfilled"
    ? workerResult.value.data.worker?.person_party_reference_id
    : null;
  const legalEntity = entityResult.status === "fulfilled"
    ? entityResult.value.data.legalEntity
    : null;
  let workerName: string | null = null;
  if (partyID) {
    const [partyResult] = await Promise.allSettled([
      executeGraphQL<{ party: { display_name: string } | null }>(
        `query ProvisioningPerson($id: ID!) { party(id: $id) { display_name } }`,
        { id: partyID },
      ),
    ]);
    if (partyResult.status === "fulfilled") {
      workerName = partyResult.value.data.party?.display_name ?? null;
    }
  }
  return { workerName, legalEntity };
}
