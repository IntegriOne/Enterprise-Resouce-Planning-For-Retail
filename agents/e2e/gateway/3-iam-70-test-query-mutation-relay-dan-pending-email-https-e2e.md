# Laporan E2E Live HTTPS Gateway + IAM: 70 Test (138 Subtest) Query, Mutation, Relay, Pending Email & Fail-Closed

- **Tanggal Verifikasi:** 2026-11-15 WIB
- **Target Endpoint:** `https://erp.localhost:8443/api/graphql` (melalui Caddy `reverse_proxy round_robin` -> `gateway-1:8080`, `gateway-2:8080` -> `iam:9001` -> PostgreSQL `18.7` + Redis `8.10.3` TLS ACL)
- **Berkas Uji Utama:** [`iam_operations_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go), [`live_https_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/live_https_e2e_test.go)
- **Perintah Eksekusi:**
  ```bash
  cd /home/muriarfad/Documents/projects/erp-retail/backend/gateway && GOWORK=off go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e
  ```
- **Ringkasan Hasil:** **`PASS`** — **`70`** top-level test, **`138`** subtest, **`0` gagal**, **`0` skip**, durasi paket `60.582s`.

## 1. Status Lingkungan Docker Compose & Smoke HTTPS

Seluruh 8 layanan aktif dan sehat (`docker compose ps`):
- `caddy` (`integrione-retail-caddy:dev`, `Up`, `127.0.0.1:8443->443/tcp`)
- `gateway-1` (`integrione-retail-gateway:dev`, `healthy`, `127.0.0.1:5000->8080/tcp`)
- `gateway-2` (`integrione-retail-gateway:dev`, `healthy`, `127.0.0.1:5001->8080/tcp`)
- `iam` (`integrione-retail-iam:dev`, `healthy`, `127.0.0.1:9001->9001/tcp`)
- `frontend-1` & `frontend-2` (`integrione-retail-frontend:dev`, `healthy`, `127.0.0.1:3333-3334->3000/tcp`)
- `postgres` (`postgres:latest`, `healthy`, `127.0.0.1:6543->5432/tcp`)
- `redis` (`redis:latest`, `healthy`, `127.0.0.1:6379->6379/tcp`)

Hasil smoke HTTP langsung pasca-E2E:
- `GET https://erp.localhost:8443/login` -> `HTTP 200`
- `GET https://erp.localhost:8443/register` -> `HTTP 200`
- `GET https://erp.localhost:8443/verify-email` -> `HTTP 200`
- `GET https://erp.localhost:8443/api/graphql?query=%7Bhealth%7D` -> `HTTP 200`

Keadaan database `iam` bersih pasca-eksekusi: `users = 1`, `roles = 4`, `groups = 0` (`active_sessions = 1`).

## 2. Matriks Query IAM & Relay (`23` Test Top-Level, Seluruhnya `PASS`)

| Test Function | Subtest Terverifikasi | Durasi | Status |
| :--- | :--- | :--- | :--- |
| `TestQueryBrowserSession` | `PreAuthAnonymous`, `AuthenticatedAllFieldsNonNull`, `AfterLogoutReturnsUnauthenticated` | `0.40s` | `PASS` |
| `TestQueryMyWorkspaceAccess` | `SuperAdminIdentityWorkspace`, `CustomerStorefrontAccess`, `AnonymousUnauthenticated` | `0.35s` | `PASS` |
| `TestQueryMe` | `EmployeeProfileCompleteNonNull`, `AnonymousUnauthenticated` | `0.29s` | `PASS` |
| `TestQueryNodeAndNestedResolvers` | `NodeLookupNestedFieldsAndRBAC` (Relay Global ID `User`/`Role`/`Group`, nested fields, RBAC) | `0.70s` | `PASS` |
| `TestQueryMyAccessSnapshot` | `SuperAdminFullSnapshot`, `CustomerSnapshotDeniedIdentity`, `AnonymousUnauthenticated` | `0.35s` | `PASS` |
| `TestQueryMySessions` | `ActiveAndHistoricalSessionsComplete`, `AnonymousUnauthenticated` | `0.35s` | `PASS` |
| `TestQueryUser` | `ValidUserWithLoginNonNull`, `RBACCustomerForbidden`, `NotFoundError` | `0.79s` | `PASS` |
| `TestQueryUsers` | `FilterSearchKindStatusPaginationAndDeleted`, `RBACCustomerForbidden` | `0.69s` | `PASS` |
| `TestQueryUserIdentities` | `MultipleVerifiedIdentities`, `RBACCustomerForbidden` | `0.50s` | `PASS` |
| `TestQueryUserRoles` | `ScopedActiveAndExpiredFilter`, `RBACCustomerForbidden` | `0.67s` | `PASS` |
| `TestQueryUserEffectiveRoles` | `DirectPlusGroupInheritedRoles`, `RBACCustomerForbidden` | `0.78s` | `PASS` |
| `TestQueryUserGroups` | `MemberAndEmptySliceNonNull`, `RBACCustomerForbidden` | `0.54s` | `PASS` |
| `TestQueryUserSessions` | `AdminInspectsUserActiveAndEndedSessions`, `RBACCustomerForbidden` | `0.61s` | `PASS` |
| `TestQueryRoles` | `FilterActiveAndInactiveCustomRoles`, `RBACCustomerForbidden` | `0.67s` | `PASS` |
| `TestQueryRole` | `ValidRoleNonNull`, `RBACCustomerForbidden` | `0.66s` | `PASS` |
| `TestQueryRoleByCode` | `SystemAndCustomRoleByCode`, `RBACCustomerForbidden` | `0.48s` | `PASS` |
| `TestQueryRolePermissions` | `AssignedAndUnassignedPermissionsNonNull`, `RBACCustomerForbidden` | `0.60s` | `PASS` |
| `TestQueryPermissions` | `FilterByResourceSensitiveAndPagination`, `RBACCustomerForbidden` | `0.34s` | `PASS` |
| `TestQueryPermission` | `ValidPermissionByID`, `RBACCustomerForbidden` | `0.35s` | `PASS` |
| `TestQueryGroups` | `FilterActiveAndInactiveGroups`, `RBACCustomerForbidden` | `0.43s` | `PASS` |
| `TestQueryGroup` | `ValidGroupNonNull`, `RBACCustomerForbidden` | `0.43s` | `PASS` |
| `TestQueryGroupMembers` | `PaginatedMembersNonNull`, `RBACCustomerForbidden` | `0.76s` | `PASS` |
| `TestQueryGroupRoles` | `ScopedActiveAndExpiredGroupRoles`, `RBACCustomerForbidden` | `0.60s` | `PASS` |

## 3. Matriks Mutation IAM & Kontrak Email/Profil/RBAC (`46` Test Top-Level, Seluruhnya `PASS`)

| Test Function | Subtest Terverifikasi | Durasi | Status |
| :--- | :--- | :--- | :--- |
| `TestMutationUserLogin` | `EmailAndUsernameLoginSuccess`, `MfaRequiredAndInvalidCredentials` | `0.72s` | `PASS` |
| `TestMutationValidateUserMfa` | `TotpAndRecoveryCodeWithReplayProtection` | `0.45s` | `PASS` |
| `TestMutationRegisterCustomer` | `RegisterSuccessAndDuplicateConflict` | `0.22s` | `PASS` |
| `TestMutationVerifyEmail` | `VerifiedAlreadyUsedExpiredAndInvalid` | `0.48s` | `PASS` |
| `TestMutationRequestPasswordReset` | `VerifiedEmailIssuesSingleActiveTokenAndUnverifiedOrGhostReturnsTrue` | `0.58s` | `PASS` |
| `TestMutationResetPassword` | `ValidTokenUpdatesPasswordAndInvalidRejected` | `0.37s` | `PASS` |
| `TestMutationUserLogout` | `AuthenticatedLogoutRevokesSession`, `AnonymousRejected` | `0.28s` | `PASS` |
| `TestMutationUpdateMyProfile` | `UpdateAllProfileFields`, `AnonymousRejected` | `0.28s` | `PASS` |
| `TestMutationUploadAvatarUser` | `ValidAvifStoresUrlAndUpdateProfilePreservesAvatar`, `AnonymousRejected` | `0.29s` | `PASS` |
| `TestMutationChangeUserPassword` | `FreshSuccessWrongCurrentAndStaleDenied` | `0.53s` | `PASS` |
| `TestMutationVerifyMySessionMfa` | `ValidTotpRotatesCsrfAndRejectsInvalidCode` | `0.29s` | `PASS` |
| `TestMutationEnrollMyTotp` | `FreshSessionSuccessAndStaleForbidden` | `0.56s` | `PASS` |
| `TestMutationConfirmMyTotp` | `ValidCodeEnablesAndInvalidRejected` | `0.29s` | `PASS` |
| `TestMutationDisableMyTotp` | `RequiresVerifiedSessionAndDisablesCleanly` | `0.29s` | `PASS` |
| `TestMutationRegenerateMyRecoveryCodes` | `GeneratesFreshCodesAndInvalidatesPrevious` | `0.34s` | `PASS` |
| `TestMutationCreateUser` | `CreateCustomerNonNull`, `RBACCustomerForbidden` | `0.50s` | `PASS` |
| `TestMutationCreateEmployeeUser` | `CreateEmployeeNonNullAndLoginWorks`, `RBACCustomerForbidden` | `0.56s` | `PASS` |
| `TestMutationCreateVendorUser` | `CreateVendorNonNullAndLoginWorks`, `RBACCustomerForbidden` | `0.55s` | `PASS` |
| `TestMutationCreateSupplierUser` | `CreateSupplierNonNullAndLoginWorks`, `RBACCustomerForbidden` | `0.57s` | `PASS` |
| `TestMutationActivateUser` | `ActivateSuspendedUser`, `RBACCustomerForbidden` | `0.57s` | `PASS` |
| `TestMutationSuspendUser` | `SuspendActiveUserBlocksSession`, `RBACCustomerForbidden` | `0.57s` | `PASS` |
| `TestMutationDisableUser` | `DisableActiveUserBlocksSession`, `RBACCustomerForbidden` | `0.58s` | `PASS` |
| `TestMutationSoftDeleteUser` | `SoftDeleteSetsDeletedAt`, `RBACCustomerForbidden` | `0.52s` | `PASS` |
| `TestMutationUpdateUserEmailDeletePendingAndResend` | `Case2PendingRegistrationResendSingleActiveToken`, `Case1UpdateUserEmailPreservesOldResendsDeletesAndVerifies`, `AnonymousUpdateOrDeleteRejected` | `1.07s` | `PASS` |
| `TestMutationSetPrimaryUserIdentity` | `SwitchPrimaryToVerifiedUsernameAndRejectUnverified`, `RBACCustomerForbidden` | `0.62s` | `PASS` |
| `TestMutationRemoveUserIdentity` | `RemoveSecondaryAndProtectPrimary`, `RBACCustomerForbidden` | `0.52s` | `PASS` |
| `TestMutationCreateRole` | `CreateValidAndRejectInvalidRedirectOrDuplicate`, `RBACCustomerForbidden` | `0.47s` | `PASS` |
| `TestMutationUpdateRoleDetails` | `UpdateAllMutableFields`, `RBACCustomerForbidden` | `0.59s` | `PASS` |
| `TestMutationActivateRole` | `ActivateDeactivatedRole`, `RBACCustomerForbidden` | `0.59s` | `PASS` |
| `TestMutationDeactivateRole` | `DeactivateCustomAndProtectSystemRole`, `RBACCustomerForbidden` | `0.58s` | `PASS` |
| `TestMutationDeleteRole` | `DeleteCustomAndProtectSystemRole`, `RBACCustomerForbidden` | `0.59s` | `PASS` |
| `TestMutationAssignRolePermission` | `AssignPermissionReturnsNonNull`, `RBACCustomerForbidden` | `0.58s` | `PASS` |
| `TestMutationRevokeRolePermission` | `RevokeAssignedPermissionReturnsTrue`, `RBACCustomerForbidden` | `0.59s` | `PASS` |
| `TestMutationAssignUserRole` | `AssignGlobalAndScopedRole`, `RBACCustomerForbidden` | `0.88s` | `PASS` |
| `TestMutationRevokeUserRole` | `RevokeReturnsTrue`, `RBACCustomerForbidden` | `0.63s` | `PASS` |
| `TestMutationCreateGroup` | `CreateValidAndDuplicateConflict`, `RBACCustomerForbidden` | `0.37s` | `PASS` |
| `TestMutationUpdateGroupDetails` | `UpdateNameAndDescription`, `RBACCustomerForbidden` | `0.42s` | `PASS` |
| `TestMutationActivateGroup` | `ActivateDeactivatedGroup`, `RBACCustomerForbidden` | `0.42s` | `PASS` |
| `TestMutationDeactivateGroup` | `DeactivateActiveGroup`, `RBACCustomerForbidden` | `0.41s` | `PASS` |
| `TestMutationDeleteGroup` | `DeleteGroupReturnsTrue`, `RBACCustomerForbidden` | `0.41s` | `PASS` |
| `TestMutationAddGroupMember` | `AddMemberReturnsNonNull`, `RBACCustomerForbidden` | `0.58s` | `PASS` |
| `TestMutationRemoveGroupMember` | `RemoveMemberReturnsTrue`, `RBACCustomerForbidden` | `0.58s` | `PASS` |
| `TestMutationAssignGroupRole` | `AssignScopedGroupRoleNonNullAndInherited`, `RBACCustomerForbidden` | `0.84s` | `PASS` |
| `TestMutationRevokeGroupRole` | `RevokeGroupRoleReturnsTrue`, `RBACCustomerForbidden` | `0.50s` | `PASS` |
| `TestMutationRevokeSession` | `RevokeTargetSessionInvalidatesClient`, `RBACCustomerForbidden` | `0.58s` | `PASS` |
| `TestMutationRevokeAllUserSessions` | `RevokeAllActiveSessionsReturnsCountAndInvalidatesAll`, `RBACCustomerForbidden` | `0.64s` | `PASS` |

## 4. Matriks Skenario End-to-End Platform & Redis ACL (`TestLiveHTTPSIAMGatewayE2E`, `24.35s`, `PASS`)

| Subtest | Ruang Lingkup Verifikasi | Durasi | Status |
| :--- | :--- | :--- | :--- |
| `SingleEndpointAndFrontendRoutes` | Endpoint publik tunggal `/api/graphql`, penolakan route REST usang, serta kesiapan halaman Next.js | `0.02s` | `PASS` |
| `BrowserBootstrapOriginCSRFAndSuperAdminLogin` | Bootstrap `browserSession`, validasi header `Origin` & `X-CSRF-Token`, cookie `__Host-integrione_session`, redirect SuperAdmin ke `/identity` | `0.06s` | `PASS` |
| `UserVariantsSnakeCaseAndRBACLifecycle` | Kontrak snake_case, kreasi user lintas kind, role & group scoped, serta aktivasi | `0.40s` | `PASS` |
| `NegativeAuthorizationTestABAndZeroDBMutation` | Skenario negatif RBAC customer -> `FORBIDDEN` dan verifikasi zero mutation di PostgreSQL | `0.29s` | `PASS` |
| `EmployeeSelfServiceTOTPMFAAndStaleSensitiveDenial` | Enrollment TOTP 2FA, login ber-MFA, penolakan operasi sensitif tanpa re-auth segar | `5.47s` | `PASS` |
| `RegistrationVerificationPasswordResetAndRateLimit` | Registrasi customer, verifikasi email, reset kata sandi, dan pembatas laju Redis | `0.10s` | `PASS` |
| `RedisACLEnforcementAndFailClosedRecovery` | Penolakan write ACL `gw_runtime`, *fail-closed* saat IAM/Redis dimatikan, dan pemulihan konsekutif lintas replika | `17.72s` | `PASS` |
