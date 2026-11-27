package iamerrors

// Sentinel error domain untuk peran, izin, grup, penugasan scope, otorisasi, dan validasi rute redirect.
var (
	// ErrRoleNotFound menandakan peran RBAC tidak ditemukan.
	ErrRoleNotFound = newSentinel(KindNotFound, "iam.role.not_found", "role not found")
	// ErrRoleAlreadyExists menandakan kode peran sudah terdaftar.
	ErrRoleAlreadyExists = newSentinel(KindAlreadyExists, "iam.role.already_exists", "role already exists")
	// ErrRoleInactive menandakan peran berstatus nonaktif.
	ErrRoleInactive = newSentinel(KindForbidden, "iam.role.inactive", "role is inactive")
	// ErrSystemRoleImmutable menandakan peran bawaan sistem tidak boleh diubah kode atau dihapus sembarangan.
	ErrSystemRoleImmutable = newSentinel(KindConflict, "iam.role.system_immutable", "system role is immutable")
	// ErrRoleInUse menandakan peran masih ditugaskan pada pengguna atau grup dan tidak dapat dihapus.
	ErrRoleInUse = newSentinel(KindConflict, "iam.role.in_use", "role is currently assigned to users or groups")
	// ErrInvalidRoleCode menandakan format atau panjang kode peran tidak valid.
	ErrInvalidRoleCode = newSentinel(KindInvalidInput, "iam.role.invalid_code", "invalid role code")
	// ErrInvalidRedirectPriority menandakan prioritas redirect berada di luar rentang smallint PostgreSQL (0..32767).
	ErrInvalidRedirectPriority = newSentinel(KindInvalidInput, "iam.role.invalid_redirect_priority", "redirect priority must be between 0 and 32767")

	// ErrPermissionNotFound menandakan izin atomik tidak ditemukan.
	ErrPermissionNotFound = newSentinel(KindNotFound, "iam.permission.not_found", "permission not found")
	// ErrPermissionAlreadyExists menandakan kode atau kombinasi resource/action izin sudah terdaftar.
	ErrPermissionAlreadyExists = newSentinel(KindAlreadyExists, "iam.permission.already_exists", "permission already exists")
	// ErrPermissionInUse menandakan izin masih terpasang pada peran dan tidak boleh dihapus.
	ErrPermissionInUse = newSentinel(KindConflict, "iam.permission.in_use", "permission is assigned to a role")
	// ErrInvalidPermissionCode menandakan format atau panjang kode izin tidak valid.
	ErrInvalidPermissionCode = newSentinel(KindInvalidInput, "iam.permission.invalid_code", "invalid permission code")
	// ErrRolePermissionExists menandakan izin sudah terpasang pada peran terkait.
	ErrRolePermissionExists = newSentinel(KindAlreadyExists, "iam.role_permission.already_exists", "role permission already exists")
	// ErrRolePermissionNotFound menandakan relasi izin pada peran tidak ditemukan.
	ErrRolePermissionNotFound = newSentinel(KindNotFound, "iam.role_permission.not_found", "role permission not found")

	// ErrGroupNotFound menandakan grup pengguna tidak ditemukan.
	ErrGroupNotFound = newSentinel(KindNotFound, "iam.group.not_found", "group not found")
	// ErrGroupAlreadyExists menandakan kode grup sudah terdaftar.
	ErrGroupAlreadyExists = newSentinel(KindAlreadyExists, "iam.group.already_exists", "group already exists")
	// ErrGroupInactive menandakan grup berada dalam status nonaktif.
	ErrGroupInactive = newSentinel(KindForbidden, "iam.group.inactive", "group is inactive")
	// ErrInvalidGroupCode menandakan format atau panjang kode grup tidak valid.
	ErrInvalidGroupCode = newSentinel(KindInvalidInput, "iam.group.invalid_code", "invalid group code")
	// ErrGroupMemberExists menandakan pengguna sudah menjadi anggota grup terkait.
	ErrGroupMemberExists = newSentinel(KindAlreadyExists, "iam.group_member.already_exists", "user is already a member of group")
	// ErrGroupMemberNotFound menandakan keanggotaan pengguna pada grup tidak ditemukan.
	ErrGroupMemberNotFound = newSentinel(KindNotFound, "iam.group_member.not_found", "group member not found")

	// ErrUserRoleExists menandakan penugasan peran pengguna pada scope yang sama sudah ada.
	ErrUserRoleExists = newSentinel(KindAlreadyExists, "iam.user_role.already_exists", "user role assignment already exists")
	// ErrUserRoleNotFound menandakan penugasan peran pengguna tidak ditemukan.
	ErrUserRoleNotFound = newSentinel(KindNotFound, "iam.user_role.not_found", "user role assignment not found")
	// ErrUserRoleExpired menandakan penugasan peran pengguna sudah melewati expires_at.
	ErrUserRoleExpired = newSentinel(KindForbidden, "iam.user_role.expired", "user role assignment has expired")
	// ErrGroupRoleExists menandakan penugasan peran grup pada scope yang sama sudah ada.
	ErrGroupRoleExists = newSentinel(KindAlreadyExists, "iam.group_role.already_exists", "group role assignment already exists")
	// ErrGroupRoleNotFound menandakan penugasan peran grup tidak ditemukan.
	ErrGroupRoleNotFound = newSentinel(KindNotFound, "iam.group_role.not_found", "group role assignment not found")

	// ErrInvalidScope menandakan kombinasi scope_type dan scope_id melanggar invarian global vs berlingkup ID.
	ErrInvalidScope = newSentinel(KindInvalidInput, "iam.rbac.invalid_scope", "invalid access scope")
	// ErrRedirectPathInvalid menandakan jalur redirect bukan path internal relatif yang aman.
	ErrRedirectPathInvalid = newSentinel(KindInvalidInput, "iam.rbac.invalid_redirect_path", "invalid internal redirect path")
	// ErrPermissionDenied menandakan principal tidak memiliki izin yang disyaratkan pada scope target.
	ErrPermissionDenied = newSentinel(KindForbidden, "iam.rbac.permission_denied", "permission denied")
	// ErrRoleUserKindMismatch menandakan peran tidak sesuai dengan kategori akun pengguna target.
	ErrRoleUserKindMismatch = newSentinel(KindForbidden, "iam.rbac.role_user_kind_mismatch", "role assignment is incompatible with target user kind")
	// ErrSensitiveReauthRequired menandakan izin sensitif memerlukan autentikasi ulang sesi terkini.
	ErrSensitiveReauthRequired = newSentinel(KindForbidden, "iam.rbac.sensitive_reauth_required", "sensitive permission requires recent authentication")
	// ErrScopeTargetNotFound menandakan entitas target scope di layanan pemilik tidak ditemukan.
	ErrScopeTargetNotFound = newSentinel(KindNotFound, "iam.scope.target_not_found", "scope target not found")
	// ErrScopeTargetNotEligible menandakan entitas target scope tidak aktif atau belum memenuhi syarat transaksi/peran.
	ErrScopeTargetNotEligible = newSentinel(KindConflict, "iam.scope.target_not_eligible", "scope target is not active or eligible")
	// ErrScopeOwnerVerificationUnavailable menandakan layanan pemilik objek scope tidak tersedia untuk memvalidasi target.
	ErrScopeOwnerVerificationUnavailable = newSentinel(KindUnavailable, "iam.scope.owner_verification_unavailable", "scope owner verification is unavailable")
)
