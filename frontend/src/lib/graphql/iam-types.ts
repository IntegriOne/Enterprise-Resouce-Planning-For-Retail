/**
 * Definisi tipe TypeScript untuk antarmuka Dashboard IAM (`/identity`).
 * Seluruh tipe selaras dengan kontrak GraphQL Gateway (`iam_auth.graphql`,
 * `iam_user.graphql`, `iam_rbac.graphql`, dan `schema.graphql`).
 */

export type UserKind = "CUSTOMER" | "EMPLOYEE" | "VENDOR" | "SUPPLIER" | "SYSTEM";

export type UserStatus = "PENDING" | "ACTIVE" | "SUSPENDED" | "DISABLED";

export type IdentityKind = "EMAIL" | "USERNAME";

export type SessionEndReason =
  | "LOGOUT"
  | "IDLE_TIMEOUT"
  | "ABSOLUTE_TIMEOUT"
  | "PASSWORD_CHANGED"
  | "ADMIN_REVOKED"
  | "USER_DISABLED";

export type ScopeType =
  | "GLOBAL"
  | "LEGAL_ENTITY"
  | "ORGANIZATION_UNIT"
  | "FACILITY"
  | "SUPPLIER_PARTY"
  | "VENDOR_PARTY"
  | "SELF";

export type GroupAssignableScopeType = Exclude<ScopeType, "SELF">;

export type Scope = {
  type: ScopeType;
  id?: string | null;
};

export type PageInfo = {
  endCursor?: string | null;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  startCursor?: string | null;
};

export type IamUser = {
  id: string;
  _id: string;
  kind: UserKind;
  status: UserStatus;
  display_name: string;
  avatar_url: string;
  locale: string;
  timezone: string;
  last_login_at?: string | null;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  auth_epoch: number;
  perm_revision: number;
};

export type IamUserIdentity = {
  id: string;
  _id: string;
  user_id: string;
  kind: IdentityKind;
  value: string;
  is_verified: boolean;
  is_primary: boolean;
  verified_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type IamSession = {
  id: string;
  _id: string;
  user_id: string;
  mfa_verified: boolean;
  mfa_verified_at?: string | null;
  ip_address: string;
  user_agent: string;
  authenticated_at: string;
  last_active_at: string;
  idle_expires_at: string;
  absolute_expires_at: string;
  ended_at?: string | null;
  end_reason?: SessionEndReason | null;
  created_at: string;
  generation: number;
};

export type IamRole = {
  id: string;
  _id: string;
  code: string;
  name: string;
  description: string;
  management_domain: string;
  default_redirect_path: string;
  redirect_priority: number;
  is_system: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type IamPermission = {
  id: string;
  _id: string;
  code: string;
  resource: string;
  action: string;
  description: string;
  is_sensitive: boolean;
  created_at: string;
};

export type IamGroup = {
  id: string;
  _id: string;
  code: string;
  name: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type ScopedPermission = {
  permission_code: string;
  is_sensitive: boolean;
  scope: Scope;
};

export type RoleGrant = {
  id: string;
  code: string;
  name: string;
  priority: number;
  default_redirect_path: string;
};

export type AccessSnapshot = {
  principal_id: string;
  user_kind: string;
  status: string;
  mfa_enabled: boolean;
  auth_epoch: number;
  perm_revision: number;
  recovery_epoch: number;
  can_access_identity: boolean;
  default_redirect_path: string;
  allowed_redirect_paths: string[];
  roles: RoleGrant[];
  permissions: ScopedPermission[];
  synced_at_unix: number;
  fresh_until_unix: number;
};

export type RelayEdge<TNode> = {
  cursor: string;
  node: TNode;
};

export type RelayConnection<TNode, TEdge = RelayEdge<TNode>> = {
  edges: TEdge[];
  nodes: TNode[];
  pageInfo: PageInfo;
  total_count: number;
};

export type UserRoleEdge = {
  cursor: string;
  node: IamRole;
  assignment_id: string;
  user_id: string;
  scope: Scope;
  granted_by?: string | null;
  expires_at?: string | null;
  created_at: string;
};

export type GroupMemberEdge = {
  cursor: string;
  node: IamUser;
  group_id: string;
  created_at: string;
};

export type GroupRoleEdge = {
  cursor: string;
  node: IamRole;
  assignment_id: string;
  group_id: string;
  scope: Scope;
  granted_by?: string | null;
  expires_at?: string | null;
  created_at: string;
};

export type RolePermissionEdge = {
  cursor: string;
  node: IamPermission;
  role_id: string;
  created_at: string;
};

export type UserConnection = RelayConnection<IamUser>;
export type UserIdentityConnection = RelayConnection<IamUserIdentity>;
export type SessionConnection = RelayConnection<IamSession>;
export type RoleConnection = RelayConnection<IamRole>;
export type PermissionConnection = RelayConnection<IamPermission>;
export type GroupConnection = RelayConnection<IamGroup>;
export type UserRoleConnection = RelayConnection<IamRole, UserRoleEdge>;
export type GroupMemberConnection = RelayConnection<IamUser, GroupMemberEdge>;
export type GroupRoleConnection = RelayConnection<IamRole, GroupRoleEdge>;
export type RolePermissionConnection = RelayConnection<IamPermission, RolePermissionEdge>;

export type RelayCursorVars = {
  first?: number | null;
  after?: string | null;
  last?: number | null;
  before?: string | null;
};

export type ProvisionUserKind = "EMPLOYEE" | "VENDOR" | "SUPPLIER" | "CUSTOMER";

export type AdminCreateUserInput = {
  kind: ProvisionUserKind;
  full_name: string;
  email: string;
  username: string;
  password: string;
  party_id?: string;
  contact_id?: string;
  contact_role?: string;
  job_title?: string;
  phone?: string;
  is_primary_contact?: boolean;
  worker_id?: string;
  provisioning_request_id?: string;
};

export type CreateRoleFormInput = {
  code: string;
  name: string;
  description?: string | null;
  management_domain: string;
  default_redirect_path: string;
  redirect_priority?: number | null;
  is_system?: boolean;
};

export type UpdateRoleFormInput = {
  id: string;
  name: string;
  description?: string | null;
  management_domain: string;
  default_redirect_path: string;
  redirect_priority?: number | null;
};

export type CreateGroupFormInput = {
  code: string;
  name: string;
  description?: string | null;
};

export type UpdateGroupFormInput = {
  id: string;
  name: string;
  description?: string | null;
};

export type AssignUserRoleFormInput = {
  user_id: string;
  role_id: string;
  scope_type: ScopeType;
  scope_id?: string | null;
  expires_at?: string | null;
};

export type AssignGroupRoleFormInput = {
  group_id: string;
  role_id: string;
  scope_type: GroupAssignableScopeType;
  scope_id?: string | null;
  expires_at?: string | null;
};
