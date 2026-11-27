# Katalog Roles & Permissions: Identity & Access Management (IAM)

Dokumen ini mendefinisikan daftar kanonik **15 *permission codes* domain `iam`**, metadata sensitivitas, aturan *scope* (`global` vs `self` vs eksternal), **7 *system template roles*** bawaan bootstrap IAM, aturan pemisahan `user_kind` (`human` vs `system` otomasi), serta pemetaan otorisasi untuk seluruh Query, Field Bersarang, dan Mutation GraphQL IAM pada Gateway.

Seluruh sumber acuan diambil langsung dari manifest kanonik [`iam.json`](../../../backend/shared/rbac/catalogs/iam.json), loader [`shared/rbac/catalog.go`](../../../backend/shared/rbac/catalog.go), [`shared/authz/snapshot.go`](../../../backend/shared/authz/snapshot.go), [`rbac_catalog_sync.go`](../../../backend/services/iam/services/rbac_catalog_sync.go), [`bootstrap.go`](../../../backend/services/iam/services/bootstrap.go), [`access_policy.go`](../../../backend/services/iam/domain/service/access_policy.go), [`role.go`](../../../backend/services/iam/domain/entity/role.go), [`assignment.go`](../../../backend/services/iam/domain/entity/assignment.go), [`runtime_sync.go`](../../../backend/services/iam/services/session/runtime_sync.go), serta [`internal/adapter/iam/`](../../../backend/gateway/internal/adapter/iam/adapter.go) (dengan test anti-drift [`permissions_manifest_test.go`](../../../backend/gateway/internal/adapter/iam/permissions_manifest_test.go)) dan [`graphql/iam/`](../../../backend/gateway/graphql/iam/iam_auth.graphql).

---

## 1. Daftar Permission Codes Domain `iam` (15 Kode Aktif)

Format kode permission mengikuti `<management>.<resource>.<action>` dan divalidasi ketergantungan metadata (`code == resource + "." + action`) pada saat bootstrap maupun pembuatan permission via IAM.

| Permission Code | `resource` | `action` | `is_sensitive` | Scope Target Valid | Deskripsi & Cakupan Utama |
| --- | --- | --- | --- | --- | --- |
| `iam.dashboard.view` | `iam.dashboard` | `view` | Tidak | `global` | Akses masuk portal administrasi `/identity` (`evaluateIdentityDashboardAccess` aktif jika redirect role `/identity` atau memiliki izin non-`*self` berawalan `iam.`). |
| `iam.users.read` | `iam.users` | `read` | Tidak | `global` | Membaca profil akun lintas pengguna dan daftar `userIdentities`. |
| `iam.users.write` | `iam.users` | `write` | **Ya** | `global` | Membuat akun staf/mitra (`employee`, `vendor`, `supplier`), aktivasi/suspend/disable/soft-delete akun, serta kelola identitas utama/cabut identitas. |
| `iam.users.readself` | `iam.users` | `readself` | Tidak | **`self` (hanya `scope_id == users.id`)** | Membaca profil akun sendiri (`me`), `UserIdentity.user`, `User.identities`, `User.roles`, `User.effective_roles`, `User.groups` milik sendiri. |
| `iam.users.updateself` | `iam.users` | `updateself` | Tidak *(sebagian aksi mewajibkan `RequireSelfSensitive`)* | **`self` (hanya `scope_id == users.id`)** | Mengubah profil, avatar, username, kirim ulang token verifikasi pending email; serta dengan **step-up 10 menit + MFA**: ubah email, ganti password, kelola TOTP MFA & recovery codes. |
| `iam.users.deleteself` | `iam.users` | `deleteself` | Tidak | **`self` (hanya `scope_id == users.id`)** | Membatalkan/menghapus pending email yang belum terverifikasi milik sendiri (`deletePendingUserEmail`). |
| `iam.roles.read` | `iam.roles` | `read` | Tidak | `global` | Membaca daftar/detail peran RBAC, izin per peran (`rolePermissions`), dan peran yang terpasang ke user (`userRoles`, `userEffectiveRoles`). |
| `iam.roles.write` | `iam.roles` | `write` | **Ya** | `global` | Membuat/mengubah/mengaktifkan/menonaktifkan/menghapus peran non-sistem, mengikat/mencabut izin ke peran, dan menetapkan/mencabut `user_roles`. |
| `iam.groups.read` | `iam.groups` | `read` | Tidak | `global` | Membaca daftar/detail grup, anggota grup (`groupMembers`), role grup (`groupRoles`), dan grup user (`userGroups`). |
| `iam.groups.write` | `iam.groups` | `write` | **Ya** | `global` | Membuat/mengubah/mengaktifkan/menonaktifkan/menghapus grup, menambah/mengeluarkan anggota, dan memasang/mencabut `group_roles`. |
| `iam.permissions.read` | `iam.permissions` | `read` | Tidak | `global` | Membaca katalog atomik `permissions` (`permissions`, `permission`, Relay `Node` permission). |
| `iam.sessions.read` | `iam.sessions` | `read` | Tidak | `global` | Membaca daftar dan node sesi aktif maupun riwayat lintas pengguna (`userSessions`, `Session.user`). |
| `iam.sessions.readself` | `iam.sessions` | `readself` | Tidak | **`self` (hanya `scope_id == users.id`)** | Membaca daftar dan detail node sesi milik akun sendiri (`mySessions`, `User.sessions`, `Session.user`, `Node` session). |
| `iam.sessions.revoke` | `iam.sessions` | `revoke` | **Ya** | `global` | Mencabut sesi aktif pengguna tertentu maupun seluruh sesi pengguna (`revokeSession`, `revokeAllUserSessions`). |
| `iam.sessions.deleteself` | `iam.sessions` | `deleteself` | **Ya** | **`self` (hanya `scope_id == users.id`)** | Mencabut sesi milik sendiri yang sedang berjalan atau perangkat lain via `revokeSession` (`RequireSelfPermission` + reauth sensitif). |

> [!NOTE]
> Selain 15 izin domain `iam.*` di atas, manifest [`enterprise_foundation.json`](../../../backend/shared/rbac/catalogs/enterprise_foundation.json) dan sinkronisasi [`rbac_catalog_sync.go`](../../../backend/services/iam/services/rbac_catalog_sync.go) turut mendaftarkan **10 izin domain `enterprise_foundation.*`** (`addresses`, `legal_entities`, `locations`, `organization_units`, `parties` untuk `.read` dan `.write`; total **25 permission bootstrap**) agar akun human `superadmin` dapat membaca daftar mitra/pihak saat onboarding akun serta akun `auto_system_ef` memiliki hak layanan terkait. Rincian izin Enterprise Foundation mengikuti [katalog Enterprise Foundation](../enterprise_foundation/1-roles-permissions.md) (seluruh 10 entri bootstrap disemai dengan `is_sensitive: false`, sedangkan mutasi kritikalnya ditegakkan sensitif pada adapter Gateway).

---

## 2. Daftar System Template Roles (`is_system = true`)

Seluruh peran berikut dibentuk dan dijaga secara idempoten saat startup oleh `BootstrapPrincipalsAndRBAC` dengan `is_system = true` (immutable terhadap edit details, deactivate, delete, maupun edit role-permission lewat API publik). Prioritas redirect menggunakan **angka terkecil sebagai pemenang**.

| Role Code | Nama Peran | `management_domain` | `default_redirect_path` | `redirect_priority` | Kompatibilitas `user_kind` | Principal / Auto-Assignment & Ringkasan Hak Akses |
| --- | --- | --- | --- | --- | --- | --- |
| `superadmin` | `Super Administrator` | `iam` | `/identity` | `1` | **Hanya `employee`** (`IsSuperAdminRoleCode`) | Human admin perdana (`IAM_SUPERADMIN_*` di `.env`). Memegang seluruh **25 permission** (**15 `iam.*`** + **10 `enterprise_foundation.*`**) pada `ScopeGlobal`, ditambah baseline `iam_selfservice` pada `ScopeSelf`. |
| `iam_operator` | `IAM Operator` | `iam` | `/identity` | `10` | Manusia non-`system` (`employee`, `vendor`, `supplier`, `customer`) | **14 permission `iam.*`**: seluruh hak operasional IAM **kecuali `iam.roles.write`** (`dashboard.view`, `users.*`, `roles.read`, `groups.*`, `permissions.read`, `sessions.*`). Tidak otomatis membaca `enterprise_foundation.parties.read` kecuali diberi role tambahan. |
| `iam_readonly` | `IAM Auditor Read-Only` | `iam` | `/identity` | `20` | Manusia non-`system` | **8 permission `iam.*`**: `iam.dashboard.view`, `iam.users.read`, `iam.users.readself`, `iam.roles.read`, `iam.groups.read`, `iam.permissions.read`, `iam.sessions.read`, `iam.sessions.readself`. |
| `customer` | `Web Commerce Customer` | `commerce` | `/` | `100` | Manusia non-`system` | **Auto-assigned pada `CreateUser`** (`CUSTOMER`) di **`ScopeSelf{ID: users.id}`**. Memegang **5 izin `*self`**: `iam.users.readself`, `iam.users.updateself`, `iam.users.deleteself`, `iam.sessions.readself`, `iam.sessions.deleteself`. |
| `iam_selfservice` | `IAM Account Self-Service` | `iam` | `""` *(fallback kind: `/vendor`, `/supplier`, `/account`)* | `999` | Manusia non-`system` | **Auto-assigned pada pembuatan `EMPLOYEE`, `VENDOR`, `SUPPLIER`** serta `superadmin` di **`ScopeSelf{ID: users.id}`**. Memegang **5 izin `*self`** yang sama dengan `customer`. |
| `iam_automation` | `IAM Service Automation` | `iam` | `""` | `999` | **Hanya `system`** (`IsAutomationRoleCode`) | Akun layanan non-interaktif `auto_system_iam`. Memegang **7 izin non-`*self`**: `iam.users.read`, `iam.users.write`, `iam.roles.read`, `iam.groups.read`, `iam.permissions.read`, `iam.sessions.read`, `iam.sessions.revoke`. |
| `ef_automation` | `Enterprise Foundation Service Automation` | `enterprise_foundation` | `""` | `999` | **Hanya `system`** (`IsAutomationRoleCode`) | Akun layanan non-interaktif `auto_system_ef`. Memegang **10 izin `enterprise_foundation.*`** pada `ScopeGlobal`. |

### Matriks Keterikatan Permission ke Template Role IAM

| Permission Code | `superadmin` | `iam_operator` | `iam_readonly` | `customer` | `iam_selfservice` | `iam_automation` | `ef_automation` |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| `iam.dashboard.view` | ✓ | ✓ | ✓ | — | — | — | — |
| `iam.users.read` | ✓ | ✓ | ✓ | — | — | ✓ | — |
| `iam.users.write` | ✓ | ✓ | — | — | — | ✓ | — |
| `iam.users.readself` | ✓ | ✓ | ✓ | ✓ *(self)* | ✓ *(self)* | — | — |
| `iam.users.updateself` | ✓ | ✓ | — | ✓ *(self)* | ✓ *(self)* | — | — |
| `iam.users.deleteself` | ✓ | ✓ | — | ✓ *(self)* | ✓ *(self)* | — | — |
| `iam.roles.read` | ✓ | ✓ | ✓ | — | — | ✓ | — |
| `iam.roles.write` | ✓ | — | — | — | — | — | — |
| `iam.groups.read` | ✓ | ✓ | ✓ | — | — | ✓ | — |
| `iam.groups.write` | ✓ | ✓ | — | — | — | — | — |
| `iam.permissions.read` | ✓ | ✓ | ✓ | — | — | ✓ | — |
| `iam.sessions.read` | ✓ | ✓ | ✓ | — | — | ✓ | — |
| `iam.sessions.readself` | ✓ | ✓ | ✓ | ✓ *(self)* | ✓ *(self)* | — | — |
| `iam.sessions.revoke` | ✓ | ✓ | — | — | — | ✓ | — |
| `iam.sessions.deleteself` | ✓ | ✓ | — | ✓ *(self)* | ✓ *(self)* | — | — |
| `enterprise_foundation.*` (10 kode) | ✓ | — | — | — | — | — | ✓ |

---

## 3. Matriks Operasi GraphQL IAM ke Permission, Sensitivitas & Scope

### 3.1 Query & Field Resolver Bersarang (`iam_auth.graphql`, `iam_user.graphql`, `iam_rbac.graphql`)

| Operasi / Field GraphQL | Gate Auth & Permission | Sensitif Reauth | Scope Evaluasi | Keterangan Perilaku |
| --- | --- | --- | --- | --- |
| `Query.health` | Publik | Tidak | — | Cek kesehatan HTTP/GraphQL Gateway tanpa menyentuh IAM/Redis. |
| `Query.browserSession` | Opsional (`AuthenticateOptional`) | Tidak | — | Anonim menerima pre-auth handle + CSRF dari Redis; terautentikasi menyinkronkan CSRF sesi aktif dari IAM (tidak menghapus cookie bila error S2S/outage). |
| `Query.myWorkspaceAccess` | `RequireAuthenticated` *(aktif)* | Tidak | `AccessSnapshot` | Mengembalikan `canAccessIdentity` (`evaluateIdentityDashboardAccess`) dan `defaultLocation` aman. |
| `Query.myAccessSnapshot` | `RequireAuthenticated` *(aktif)* | Tidak | `AccessSnapshot` | Membaca snapshot otorisasi segar (`<= 30s`) milik pemanggil. |
| `Query.me` | `iam.users.readself` | Tidak | `self` (`users.id`) | Mengembalikan profil user aktif pemilik sesi. |
| `Query.mySessions` | `iam.sessions.readself` | Tidak | `self` (`users.id`) | Daftar sesi aktif/riwayat milik sendiri terpaginasi Relay. |
| `Query.user`, `Query.users`, `Query.userIdentities` | `iam.users.read` | Tidak | `global` | Pencarian & inspeksi akun/identitas pengguna lintas sistem. |
| `Query.userRoles`, `Query.userEffectiveRoles` | `iam.roles.read` | Tidak | `global` | Inspeksi assignment langsung dan peran efektif pengguna. |
| `Query.userGroups` | `iam.groups.read` | Tidak | `global` | Inspeksi keanggotaan grup milik target pengguna. |
| `Query.userSessions` | `iam.sessions.read` | Tidak | `global` | Inspeksi sesi milik target pengguna oleh operator IAM. |
| `Query.roles`, `Query.role`, `Query.roleByCode`, `Query.rolePermissions`, `Role.permissions` | `iam.roles.read` | Tidak | `global` | Katalog dan relasi permission milik peran. |
| `Query.permissions`, `Query.permission` | `iam.permissions.read` | Tidak | `global` | Katalog atomik permission terdaftar. |
| `Query.groups`, `Query.group`, `Query.groupMembers`, `Query.groupRoles`, `Group.members`, `Group.roles` | `iam.groups.read` | Tidak | `global` | Katalog grup beserta anggota dan assignment perannya. |
| `Query.node(id: ID!)` (GID IAM) | `RequireAuthenticated` + dispatch resource | Tidak | `self` / `global` | `user` → `users.readself`/`users.read`; `role`/`permission`/`group` → masing-masing `.read`; `user-identity` & `session` → cek kepemilikan `self` dulu lalu fallback `.read` admin. |
| `User.identities`, `UserIdentity.user` | `iam.users.readself` **atau** `iam.users.read` | Tidak | `self` / `global` | Menggunakan `authorizeUserRef` (jika target == pemanggil wajib `*self`, fallback global admin). |
| `User.roles`, `User.effective_roles` | `iam.users.readself` **atau** `iam.roles.read` | Tidak | `self` / `global` | Pemilik akun dapat melihat perannya sendiri lewat `readself`; lintas akun wajib `iam.roles.read`. |
| `User.groups` | `iam.users.readself` **atau** `iam.groups.read` | Tidak | `self` / `global` | Pemilik akun dapat melihat grupnya sendiri lewat `readself`; lintas akun wajib `iam.groups.read`. |
| `User.sessions`, `Session.user` | `iam.sessions.readself` **atau** `iam.sessions.read` | Tidak | `self` / `global` | Sesi & relasi balik user dibatasi ke pemilik atau auditor sesi global. |

### 3.2 Mutation (`iam_auth.graphql`, `iam_user.graphql`, `iam_rbac.graphql`)

*(Seluruh mutasi HTTP POST pada `/api/graphql` juga wajib lolos `Origin`, rate limiter Redis, serta `X-CSRF-Token` pada `Guard` sebelum masuk ke resolver).*

| Mutasi GraphQL | Gate Auth & Permission | Sensitif Reauth (`10m` + MFA) | Scope & Aturan Keamanan Tambahan |
| --- | --- | --- | --- |
| `userLogin`, `validateUserMfa`, `verifyEmail`, `userResetPassword`, `resetPassword` | Publik + Rate Limit | Tidak | Akun `USER_KIND_SYSTEM` (`auto_system_*`) dan sandi kosong selalu ditolak **`system cannot login`** (`FORBIDDEN`). Akun `PENDING` belum terverifikasi ditolak login dengan anti-enumerasi. |
| `createUser` | Publik (anonim/customer) + Rate Limit **atau** Admin sesi | Tidak *(publik)* | Membuat akun `CUSTOMER` (`status: PENDING`). Mengembalikan field `user` hanya jika pemanggil terautentikasi; pendaftaran anonim email duplikat mengembalikan redirect netral ke `/login`. |
| `resendValidationEmailToken` | Opsional + Rate Limit (`iam.users.updateself` bila terautentikasi) | Tidak | Jika terautentikasi: memakai `user_id` dari cookie sesi (mengabaikan `input.email`). Jika anonim: memerlukan `input.email` dengan respons netral anti-enumerasi. |
| `userLogout` | `RequireAuthenticated` *(aktif)* | Tidak | Mengakhiri sesi aktif di IAM & Redis serta membersihkan cookie `__Host-integrione_session`. |
| `updateMyProfile`, `uploadAvatarUser`, `changeUsername` | `iam.users.updateself` | Tidak | `self` (`users.id`) | Memperbarui profil dasar, avatar AVIF, atau username akun sendiri. |
| `deletePendingUserEmail` | `iam.users.deleteself` | Tidak | `self` (`users.id`) | Menghapus pending email yang belum terverifikasi milik sendiri. |
| `verifyMySessionMfa` | `iam.users.updateself` | Tidak | `self` (`users.id`) | Meningkatkan status MFA pada sesi aktif (TOTP step-up) dan menyelaraskan `X-CSRF-Token` baru. |
| `updateUserEmail`, `changeUserPassword`, `enrollMyTotp`, `confirmMyTotp`, `disableMyTotp`, `regenerateMyRecoveryCodes` | `iam.users.updateself` | **Ya (`RequireSelfSensitive`)** | `self` (`users.id`) | Mewajibkan autentikasi segar `<= 10` menit & MFA step-up bila TOTP aktif. `changeUserPassword` mencabut seluruh sesi lama pengguna. |
| `createEmployeeUser`, `createVendorUser`, `createSupplierUser` | `iam.users.write` | **Ya** | `global` | Membuat akun staf/mitra berstatus `PENDING` + auto-assign `iam_selfservice` (`ScopeSelf`), wajib verifikasi email. |
| `activateUser` | `iam.users.write` | **Ya** | `global` | Mengizinkan self jika non-superadmin, namun tetap menolak mengubah akun `superadmin`/`*_automation` kecuali dilakukan oleh `superadmin`. |
| `suspendUser`, `disableUser`, `softDeleteUser` | `iam.users.write` | **Ya** | `global` | Dilarang melakukan aksi destruktif ke akun sendiri (`allowSelf = false`); non-`superadmin` dilarang menyentuh akun `superadmin` atau `*_automation`. |
| `setPrimaryUserIdentity`, `removeUserIdentity` | `iam.users.write` | **Ya** | `global` | Mengelola identitas non-primary/primary; akun `superadmin`/`*_automation` terlindungi; tidak boleh menghapus identitas terakhir atau identitas primary aktif. |
| `revokeSession` | `iam.sessions.deleteself` *(bila sesi sendiri)* **atau** `iam.sessions.revoke` | **Ya** | `self` / `global` | Memeriksa kepemilikan sesi milik sendiri via `findOwnedUserSession` tanpa membocorkan eksistensi sesi orang lain (*zero-oracle*). |
| `revokeAllUserSessions` | `iam.sessions.revoke` | **Ya** | `global` | Mencabut seluruh sesi target user dengan proteksi target `superadmin`/`*_automation`. |
| `createRole`, `updateRoleDetails`, `activateRole`, `deactivateRole`, `deleteRole`, `assignRolePermission`, `revokeRolePermission` | `iam.roles.write` | **Ya** | `global` | `is_system: true` ditolak pada `createRole`; peran `IsSystem` immutable di service IAM; non-`superadmin` tidak dapat memodifikasi/mendelegasikan izin yang tidak dimilikinya maupun peran `superadmin`/`*_automation`. |
| `assignUserRole` | `iam.roles.write` | **Ya** | `global` / `self` *(non-eksternal)* | Menegakkan anti-delegasi izin, kecocokan `user_kind`, validasi `scope_id == user_id` untuk `SELF`, dan **menolak `UNAVAILABLE`** pada scope eksternal hingga owner verifier aktif. |
| `revokeUserRole` | `iam.roles.write` | **Ya** | `global` | Saat ini Gateway hanya mengizinkan `superadmin` mencabut `user_roles` berdasarkan `assignment_id` (`callerIsSuperAdmin`). |
| `createGroup`, `updateGroupDetails`, `activateGroup`, `deactivateGroup`, `deleteGroup`, `addGroupMember`, `removeGroupMember`, `assignGroupRole`, `revokeGroupRole` | `iam.groups.write` | **Ya** | `global` *(kecuali `GroupRole` menolak `SELF`)* | Anti-delegasi memeriksa seluruh permission role di grup, mencegah pencampuran akun `system` ke grup peran manusia (`ErrRoleUserKindMismatch`), `revokeGroupRole` khusus `superadmin`, dan `GroupRole` menolak scope `SELF` (`NewGroupRole`). |

---

## 4. Prinsip Trust Boundary, Isolasi & Ketahanan

1. **Pemisahan Human `superadmin` vs Akun Otomasi `system`:**
   - Akun `user_kind = system` (`auto_system_iam`, `auto_system_ef`) dilarang memiliki email, kredensial sandi, TOTP/recovery code, sesi interaktif, maupun jalur dashboard (`ErrSystemCannotLogin` & `ErrSystemUserInteractiveForbidden`).
   - Penugasan peran antara `superadmin` (khusus `employee`), `*_automation` (khusus `system`), dan peran manusia biasa ditegakkan berlapis di service (`IsRoleCompatibleWithUserKind`) serta Gateway.
2. **Isolasi `*self` dan Pemblokiran Akun `PENDING`:**
   - Izin berakhiran `.readself`, `.updateself`, dan `.deleteself` tidak pernah valid pada `ScopeGlobal` maupun `ScopeSelf` dengan `scope_id != users.id`.
   - Meskipun peran baseline (`customer` / `iam_selfservice`) otomatis melekat sejak `PENDING`, `User.CanAuthenticate()` dan `AccessSnapshot.HasPermission(...)` menolak seluruh akses hingga identitas terverifikasi (`is_verified = true`) dan `status = active`.
3. **Verifikasi `ScopeType` Non-Global / Non-Self (`RequireScopeOwnerVerification`):**
   - Assignment role pada scope eksternal (`legal_entity`, `organization_unit`, `facility`, `vendor_party`, `supplier_party`) saat ini masih ditolak *fail-closed* dengan `UNAVAILABLE` (`iam.scope.owner_verification_unavailable`) di `role.Service` / `group.Service` sampai kontrak verifier objek pemilik di Enterprise Foundation resmi diaktifkan.
4. **Ketahanan Gateway saat IAM Unreachable:**
   - Gateway mengevaluasi sesi dan `AccessSnapshot` segar (`<= 30s`) dari Redis (`AuthManager` & `Authorizer`) dengan pendekatan *fail-closed* tanpa menyimpan fallback `ALLOW`; gangguan konektivitas IAM (`UNAVAILABLE`) maupun kegagalan verifikasi S2S internal HMAC (`codes.Unauthenticated` dari interceptor) dipetakan ke `UNAVAILABLE` tanpa menghapus cookie sesi browser pengguna yang masih sah.
