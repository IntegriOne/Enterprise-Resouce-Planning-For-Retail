// Package rbac menyediakan loader dan validator manifest deklaratif katalog permission serta template peran RBAC lintas management.
package rbac

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	// CurrentSchemaVersion menetapkan versi skema manifest katalog RBAC yang didukung saat ini.
	CurrentSchemaVersion = 1

	// MaxPermissionCodeLength membatasi panjang kode izin sesuai skema tabel permissions IAM.
	MaxPermissionCodeLength = 160
	// MaxPermissionResourceLength membatasi panjang kolom resource pada tabel permissions IAM.
	MaxPermissionResourceLength = 100
	// MaxPermissionActionLength membatasi panjang kolom action pada tabel permissions IAM.
	MaxPermissionActionLength = 60
	// MaxRoleCodeLength membatasi panjang kolom code pada tabel roles IAM.
	MaxRoleCodeLength = 100
	// MaxRoleNameLength membatasi panjang kolom name pada tabel roles IAM.
	MaxRoleNameLength = 160
	// MaxManagementDomainLength membatasi panjang kolom management_domain pada tabel roles IAM.
	MaxManagementDomainLength = 100
	// MaxRedirectPathLength membatasi panjang path redirect bawaan peran agar selaras dengan entitas Role IAM.
	MaxRedirectPathLength = 120
	// MaxRedirectPriority membatasi prioritas redirect pada rentang smallint PostgreSQL (0..32767).
	MaxRedirectPriority = 32767
)

var (
	domainPattern       = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	codePattern         = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*$`)
	internalPathPattern = regexp.MustCompile(`^/[a-zA-Z0-9/_-]*$`)

	// Daftar domain baseline superadmin yang disetujui; domain bisnis baru dilarang diwariskan ke superadmin secara otomatis.
	approvedSuperadminBaselineDomains = map[string]struct{}{
		"iam":                   {},
		"enterprise_foundation": {},
	}

	//go:embed catalogs/*.json
	embeddedCatalogFS embed.FS
)

// PermissionSpec mendefinisikan satu entri izin atomik deklaratif dalam manifest domain.
type PermissionSpec struct {
	Code        string `json:"code"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	IsSensitive bool   `json:"is_sensitive"`
}

// RoleSpec mendefinisikan satu template peran deklaratif beserta daftar kode izin yang terikat ke peran tersebut.
type RoleSpec struct {
	Code                string   `json:"code"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	ManagementDomain    string   `json:"management_domain"`
	DefaultRedirectPath string   `json:"default_redirect_path"`
	RedirectPriority    int      `json:"redirect_priority"`
	IsSystem            bool     `json:"is_system"`
	PermissionCodes     []string `json:"permission_codes"`
}

// CatalogManifest merepresentasikan dokumen JSON katalog RBAC milik satu management domain.
type CatalogManifest struct {
	SchemaVersion    int              `json:"schema_version"`
	ManagementDomain string           `json:"management_domain"`
	Permissions      []PermissionSpec `json:"permissions"`
	Roles            []RoleSpec       `json:"roles"`
}

// CatalogSet menyimpan hasil penggabungan tervalidasi dari seluruh manifest domain yang disematkan.
type CatalogSet struct {
	Manifests        []CatalogManifest
	Permissions      []PermissionSpec
	Roles            []RoleSpec
	PermissionByCode map[string]PermissionSpec
	RoleByCode       map[string]RoleSpec
}

// DomainPermissions mengembalikan salinan terurut daftar permission milik satu domain tertentu.
func (s *CatalogSet) DomainPermissions(domain string) []PermissionSpec {
	if s == nil {
		return nil
	}
	normDomain := normalizeLowerTrim(domain)
	if normDomain == "" {
		return nil
	}
	prefix := normDomain + "."
	var out []PermissionSpec
	for _, p := range s.Permissions {
		if strings.HasPrefix(p.Code, prefix) {
			out = append(out, p)
		}
	}
	slices.SortFunc(out, func(a, b PermissionSpec) int {
		return strings.Compare(a.Code, b.Code)
	})
	return out
}

// PermissionCodesForDomain mengembalikan daftar kode permission terurut milik satu domain tertentu.
func (s *CatalogSet) PermissionCodesForDomain(domain string) []string {
	perms := s.DomainPermissions(domain)
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		out = append(out, p.Code)
	}
	return out
}

// HasPermission memeriksa apakah kode permission terdaftar dalam kumpulan manifest katalog.
func (s *CatalogSet) HasPermission(code string) bool {
	if s == nil || len(s.PermissionByCode) == 0 {
		return false
	}
	_, ok := s.PermissionByCode[normalizeLowerTrim(code)]
	return ok
}

// LoadEmbeddedCatalogs memuat dan memvalidasi seluruh manifest katalog RBAC yang disematkan di binary.
func LoadEmbeddedCatalogs() (*CatalogSet, error) {
	return LoadCatalogsFromFS(embeddedCatalogFS, "catalogs")
}

// LoadCatalogsFromFS membaca seluruh file manifest *.json dalam direktori fs.FS, menormalkan, dan memvalidasinya secara ketat.
func LoadCatalogsFromFS(fsys fs.FS, dir string) (*CatalogSet, error) {
	if fsys == nil {
		return nil, errors.New("catalog filesystem must not be nil")
	}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("read catalog directory %q: %w", dir, err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			return nil, fmt.Errorf("catalog directory %q must not contain subdirectories such as %q", dir, entry.Name())
		}
		if path.Ext(entry.Name()) != ".json" {
			return nil, fmt.Errorf("unexpected non-json file %q in catalog directory %q", entry.Name(), dir)
		}
		names = append(names, entry.Name())
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no catalog manifest files found in %q", dir)
	}
	slices.Sort(names)

	manifests := make([]CatalogManifest, 0, len(names))
	for _, name := range names {
		fullPath := path.Join(dir, name)
		raw, readErr := fs.ReadFile(fsys, fullPath)
		if readErr != nil {
			return nil, fmt.Errorf("read catalog manifest %q: %w", fullPath, readErr)
		}
		manifest, decodeErr := decodeCatalogManifestStrict(raw)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode catalog manifest %q: %w", fullPath, decodeErr)
		}
		expectedDomain := strings.TrimSuffix(name, ".json")
		if normalizeLowerTrim(manifest.ManagementDomain) != expectedDomain {
			return nil, fmt.Errorf(
				"catalog manifest %q management_domain %q must match filename stem %q",
				fullPath,
				manifest.ManagementDomain,
				expectedDomain,
			)
		}
		manifests = append(manifests, manifest)
	}

	return ValidateCatalogSet(manifests)
}

// ValidateCatalogSet memvalidasi konsistensi seluruh manifest, isolasi namespace, dan kelengkapan referensi izin peran.
func ValidateCatalogSet(manifests []CatalogManifest) (*CatalogSet, error) {
	if len(manifests) == 0 {
		return nil, errors.New("catalog set must contain at least one manifest")
	}

	seenDomains := make(map[string]struct{}, len(manifests))
	permByCode := make(map[string]PermissionSpec)
	roleByCode := make(map[string]RoleSpec)
	normalizedManifests := make([]CatalogManifest, 0, len(manifests))
	var allPermissions []PermissionSpec
	var allRoles []RoleSpec

	for i := range manifests {
		m := cloneManifest(manifests[i])
		if err := ValidateManifest(&m); err != nil {
			return nil, err
		}
		if _, dupDomain := seenDomains[m.ManagementDomain]; dupDomain {
			return nil, fmt.Errorf("duplicate catalog manifest for management_domain %q", m.ManagementDomain)
		}
		seenDomains[m.ManagementDomain] = struct{}{}

		for _, p := range m.Permissions {
			if _, dup := permByCode[p.Code]; dup {
				return nil, fmt.Errorf("duplicate permission code %q across manifests", p.Code)
			}
			permByCode[p.Code] = p
			allPermissions = append(allPermissions, p)
		}
		for _, r := range m.Roles {
			if _, dup := roleByCode[r.Code]; dup {
				return nil, fmt.Errorf("duplicate role code %q across manifests", r.Code)
			}
			roleByCode[r.Code] = r
			allRoles = append(allRoles, r)
		}
		normalizedManifests = append(normalizedManifests, m)
	}

	for _, r := range allRoles {
		for _, pCode := range r.PermissionCodes {
			if _, ok := permByCode[pCode]; !ok {
				return nil, fmt.Errorf("role %q references unknown permission code %q", r.Code, pCode)
			}
		}
	}

	slices.SortFunc(normalizedManifests, func(a, b CatalogManifest) int {
		return strings.Compare(a.ManagementDomain, b.ManagementDomain)
	})

	return &CatalogSet{
		Manifests:        normalizedManifests,
		Permissions:      allPermissions,
		Roles:            allRoles,
		PermissionByCode: permByCode,
		RoleByCode:       roleByCode,
	}, nil
}

// ValidateManifest menormalkan dan memvalidasi aturan skema, namespace lock, dan batas keamanan satu manifest domain.
func ValidateManifest(m *CatalogManifest) error {
	if m == nil {
		return errors.New("catalog manifest must not be nil")
	}
	if m.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d: want %d", m.SchemaVersion, CurrentSchemaVersion)
	}
	domain := normalizeLowerTrim(m.ManagementDomain)
	if !isValidDomainCode(domain) {
		return fmt.Errorf("invalid management_domain %q", m.ManagementDomain)
	}
	m.ManagementDomain = domain

	seenPerms := make(map[string]struct{}, len(m.Permissions))
	for i := range m.Permissions {
		if err := normalizeAndValidatePermission(domain, &m.Permissions[i]); err != nil {
			return fmt.Errorf("manifest %q permission[%d]: %w", domain, i, err)
		}
		if _, dup := seenPerms[m.Permissions[i].Code]; dup {
			return fmt.Errorf("manifest %q contains duplicate permission %q", domain, m.Permissions[i].Code)
		}
		seenPerms[m.Permissions[i].Code] = struct{}{}
	}

	seenRoles := make(map[string]struct{}, len(m.Roles))
	for i := range m.Roles {
		if err := normalizeAndValidateRole(domain, &m.Roles[i]); err != nil {
			return fmt.Errorf("manifest %q role[%d]: %w", domain, i, err)
		}
		if _, dup := seenRoles[m.Roles[i].Code]; dup {
			return fmt.Errorf("manifest %q contains duplicate role %q", domain, m.Roles[i].Code)
		}
		seenRoles[m.Roles[i].Code] = struct{}{}
	}
	return nil
}

func decodeCatalogManifestStrict(raw []byte) (CatalogManifest, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()

	var manifest CatalogManifest
	if err := dec.Decode(&manifest); err != nil {
		return CatalogManifest{}, err
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return CatalogManifest{}, errors.New("unexpected trailing JSON tokens")
	}
	return manifest, nil
}

func normalizeAndValidatePermission(manifestDomain string, p *PermissionSpec) error {
	code := normalizeLowerTrim(p.Code)
	resource := normalizeLowerTrim(p.Resource)
	action := normalizeLowerTrim(p.Action)
	description := strings.TrimSpace(p.Description)

	if code == "" || utf8.RuneCountInString(code) > MaxPermissionCodeLength || !codePattern.MatchString(code) {
		return fmt.Errorf("invalid permission code %q", p.Code)
	}
	if resource == "" || utf8.RuneCountInString(resource) > MaxPermissionResourceLength || !codePattern.MatchString(resource) {
		return fmt.Errorf("invalid permission resource %q", p.Resource)
	}
	if action == "" || utf8.RuneCountInString(action) > MaxPermissionActionLength || !codePattern.MatchString(action) || strings.Contains(action, ".") {
		return fmt.Errorf("invalid permission action %q", p.Action)
	}
	if !strings.HasPrefix(resource, manifestDomain+".") {
		return fmt.Errorf("permission resource %q violates namespace lock for domain %q", resource, manifestDomain)
	}
	rest := strings.TrimPrefix(resource, manifestDomain+".")
	if rest == "" || strings.HasPrefix(rest, ".") || strings.HasSuffix(rest, ".") {
		return fmt.Errorf("permission resource %q must include non-empty resource segment under %q", resource, manifestDomain)
	}
	if code != resource+"."+action {
		return fmt.Errorf("permission code %q does not match resource %q + action %q", code, resource, action)
	}
	if description == "" {
		return fmt.Errorf("permission %q description must not be empty", code)
	}

	p.Code = code
	p.Resource = resource
	p.Action = action
	p.Description = description
	return nil
}

func normalizeAndValidateRole(manifestDomain string, r *RoleSpec) error {
	code := normalizeLowerTrim(r.Code)
	name := strings.TrimSpace(r.Name)
	description := strings.TrimSpace(r.Description)
	roleDomain := normalizeLowerTrim(r.ManagementDomain)
	redirectPath := strings.TrimSpace(r.DefaultRedirectPath)

	if code == "" || utf8.RuneCountInString(code) > MaxRoleCodeLength || !codePattern.MatchString(code) || strings.Contains(code, ".") {
		return fmt.Errorf("invalid role code %q", r.Code)
	}
	if name == "" || utf8.RuneCountInString(name) > MaxRoleNameLength {
		return fmt.Errorf("role %q name must be non-empty and within %d characters", code, MaxRoleNameLength)
	}
	if description == "" {
		return fmt.Errorf("role %q description must not be empty", code)
	}
	if !isValidDomainCode(roleDomain) {
		return fmt.Errorf("role %q has invalid management_domain %q", code, r.ManagementDomain)
	}
	if redirectPath != "" {
		if err := validateInternalRedirectPath(redirectPath); err != nil {
			return fmt.Errorf("role %q default_redirect_path %q invalid: %w", code, redirectPath, err)
		}
	}
	if r.RedirectPriority < 0 || r.RedirectPriority > MaxRedirectPriority {
		return fmt.Errorf("role %q redirect_priority %d out of bounds [0..%d]", code, r.RedirectPriority, MaxRedirectPriority)
	}
	if len(r.PermissionCodes) == 0 {
		return fmt.Errorf("role %q must reference at least one permission_code", code)
	}

	if err := validateRoleNamespaceLock(manifestDomain, code, roleDomain, r.IsSystem); err != nil {
		return err
	}

	normPerms := make([]string, 0, len(r.PermissionCodes))
	seenPerms := make(map[string]struct{}, len(r.PermissionCodes))
	for _, rawPerm := range r.PermissionCodes {
		pCode := normalizeLowerTrim(rawPerm)
		if pCode == "" || !codePattern.MatchString(pCode) {
			return fmt.Errorf("role %q has invalid permission code %q", code, rawPerm)
		}
		if _, dup := seenPerms[pCode]; dup {
			return fmt.Errorf("role %q lists duplicate permission code %q", code, pCode)
		}
		seenPerms[pCode] = struct{}{}

		if err := validateRolePermissionNamespace(manifestDomain, code, pCode); err != nil {
			return err
		}
		normPerms = append(normPerms, pCode)
	}

	r.Code = code
	r.Name = name
	r.Description = description
	r.ManagementDomain = roleDomain
	r.DefaultRedirectPath = redirectPath
	r.PermissionCodes = normPerms
	return nil
}

func validateRoleNamespaceLock(manifestDomain, roleCode, roleDomain string, isSystem bool) error {
	if manifestDomain == "iam" {
		if roleCode == "customer" {
			if roleDomain != "commerce" || !isSystem {
				return fmt.Errorf("core role %q in iam manifest must declare management_domain %q and is_system=true", roleCode, "commerce")
			}
			return nil
		}
		if roleDomain != "iam" {
			return fmt.Errorf("role %q in iam manifest must declare management_domain %q", roleCode, "iam")
		}
		return nil
	}

	if roleCode == "superadmin" || roleCode == "super_admin" || roleCode == "customer" || strings.HasPrefix(roleCode, "iam_") {
		return fmt.Errorf("manifest %q cannot declare reserved core role %q", manifestDomain, roleCode)
	}
	if roleDomain != manifestDomain {
		return fmt.Errorf("role %q in manifest %q violates domain lock with management_domain %q", roleCode, manifestDomain, roleDomain)
	}
	if isSystem && !strings.HasSuffix(roleCode, "_automation") {
		return fmt.Errorf("manifest %q role %q cannot set is_system=true unless it is a *_automation role", manifestDomain, roleCode)
	}
	return nil
}

func validateRolePermissionNamespace(manifestDomain, roleCode, permCode string) error {
	dotIdx := strings.IndexByte(permCode, '.')
	if dotIdx <= 0 {
		return fmt.Errorf("role %q references malformed permission code %q", roleCode, permCode)
	}
	permDomain := permCode[:dotIdx]

	if manifestDomain == "iam" {
		if roleCode == "superadmin" {
			if _, allowed := approvedSuperadminBaselineDomains[permDomain]; !allowed {
				return fmt.Errorf("superadmin role cannot auto-inherit non-baseline domain permission %q", permCode)
			}
			return nil
		}
		if roleCode == "customer" && !isSelfServicePermission(permCode) {
			return fmt.Errorf("core customer role can only hold self-service permissions, got %q", permCode)
		}
		if permDomain != "iam" {
			return fmt.Errorf("iam role %q cannot reference foreign domain permission %q", roleCode, permCode)
		}
		return nil
	}

	if permDomain != manifestDomain {
		return fmt.Errorf("manifest %q role %q cannot reference foreign domain permission %q", manifestDomain, roleCode, permCode)
	}
	return nil
}

func isValidDomainCode(domain string) bool {
	return domain != "" && utf8.RuneCountInString(domain) <= MaxManagementDomainLength && domainPattern.MatchString(domain)
}

func validateInternalRedirectPath(p string) error {
	if p == "" || len(p) > MaxRedirectPathLength {
		return errors.New("redirect path exceeds length bounds")
	}
	if !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		return errors.New("redirect path must be an internal relative route")
	}
	if strings.Contains(p, "..") || !internalPathPattern.MatchString(p) {
		return errors.New("redirect path contains forbidden characters")
	}
	return nil
}

func isSelfServicePermission(code string) bool {
	return strings.HasSuffix(code, ".readself") ||
		strings.HasSuffix(code, ".updateself") ||
		strings.HasSuffix(code, ".deleteself")
}

func normalizeLowerTrim(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func cloneManifest(m CatalogManifest) CatalogManifest {
	out := CatalogManifest{
		SchemaVersion:    m.SchemaVersion,
		ManagementDomain: m.ManagementDomain,
		Permissions:      slices.Clone(m.Permissions),
		Roles:            make([]RoleSpec, len(m.Roles)),
	}
	for i := range m.Roles {
		out.Roles[i] = m.Roles[i]
		out.Roles[i].PermissionCodes = slices.Clone(m.Roles[i].PermissionCodes)
	}
	return out
}
