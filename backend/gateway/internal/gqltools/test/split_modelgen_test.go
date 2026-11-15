package gqltools_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/integrione/erp-retail/backend/gateway/internal/gqltools"
)

// TestUseSplitModelGenFromRealSchemas memverifikasi bahwa skema GraphQL Gateway hasil pemecahan
// menghasilkan file exec, model, dan resolver terpisah per file skema tanpa motherfile
// serta tetap terkompilasi penuh saat validasi ulang dijalankan.
func TestUseSplitModelGenFromRealSchemas(t *testing.T) {
	gatewayRoot := findGatewayRoot(t)
	tmpDir := t.TempDir()

	prepareTempGatewayWorkspace(t, gatewayRoot, tmpDir)

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to temp workspace %s: %v", tmpDir, err)
	}

	runGenerate := func() {
		cfg, err := config.LoadConfig("gqlgen.yml")
		if err != nil {
			t.Fatalf("load gqlgen.yml: %v", err)
		}
		cfg.SkipModTidy = true

		if err := api.Generate(cfg, gqltools.UseSplitModelGen()); err != nil {
			t.Fatalf("api.Generate failed: %v", err)
		}
	}

	// Eksekusi generasi pertama pada workspace bersih.
	runGenerate()

	assertExists(t,
		filepath.Join(tmpDir, "graph/generated/root_.generated.go"),
		filepath.Join(tmpDir, "graph/generated/schema.generated.go"),
		filepath.Join(tmpDir, "graph/generated/iam_auth.generated.go"),
		filepath.Join(tmpDir, "graph/generated/iam_user.generated.go"),
		filepath.Join(tmpDir, "graph/generated/iam_rbac.generated.go"),
		filepath.Join(tmpDir, "graph/model/schema_models_gen.go"),
		filepath.Join(tmpDir, "graph/model/iam_auth_models_gen.go"),
		filepath.Join(tmpDir, "graph/model/iam_user_models_gen.go"),
		filepath.Join(tmpDir, "graph/model/iam_rbac_models_gen.go"),
		filepath.Join(tmpDir, "graph/resolvers/resolver.go"),
		filepath.Join(tmpDir, "graph/resolvers/schema.resolvers.go"),
		filepath.Join(tmpDir, "graph/resolvers/iam_auth.resolvers.go"),
		filepath.Join(tmpDir, "graph/resolvers/iam_user.resolvers.go"),
		filepath.Join(tmpDir, "graph/resolvers/iam_rbac.resolvers.go"),
	)

	assertNotExists(t,
		filepath.Join(tmpDir, "graph/generated/generated.go"),
		filepath.Join(tmpDir, "graph/model/models_gen.go"),
	)

	assertFileDeclaresTypes(t, filepath.Join(tmpDir, "graph/model/schema_models_gen.go"),
		"Node", "PageInfo", "Scope", "ScopeType",
	)
	assertFileDeclaresTypes(t, filepath.Join(tmpDir, "graph/model/iam_auth_models_gen.go"),
		"UserLoginStatus", "VerifyEmailStatus", "BrowserSessionPayload",
		"UserLoginInput", "UserAvatarPayload", "UserAvatarUploadInput",
		"UserEmailInput", "UserEmailPayload", "ResendValidationEmailInput", "ResendValidationEmailPayload",
	)
	assertFileDeclaresTypes(t, filepath.Join(tmpDir, "graph/model/iam_user_models_gen.go"),
		"UserKind", "UserStatus", "User", "UserIdentity", "Session",
		"UserEdge", "UserConnection",
	)
	assertFileDeclaresMethod(t, filepath.Join(tmpDir, "graph/model/iam_user_models_gen.go"),
		"User", "IsNode",
	)
	assertStructDeclaresFields(t, filepath.Join(tmpDir, "graph/model/iam_user_models_gen.go"),
		"User", "ID", "RawID",
	)
	assertFileDeclaresMethod(t, filepath.Join(tmpDir, "graph/model/iam_user_models_gen.go"),
		"UserKind", "IsValid",
	)
	assertFileDeclaresTypes(t, filepath.Join(tmpDir, "graph/model/iam_rbac_models_gen.go"),
		"Role", "Permission", "Group", "AccessSnapshot",
		"RolePermissionEdge", "UserRoleEdge", "GroupMemberEdge", "GroupRoleEdge",
	)
	assertFileDeclaresMethod(t, filepath.Join(tmpDir, "graph/model/iam_rbac_models_gen.go"),
		"Role", "IsNode",
	)
	assertStructDeclaresFields(t, filepath.Join(tmpDir, "graph/model/iam_rbac_models_gen.go"),
		"Group", "ID", "RawID",
	)

	// Eksekusi ulang untuk memastikan file *_models_gen.go lama dibersihkan dan generasi ulang idempoten.
	staleFile := filepath.Join(tmpDir, "graph/model/stale_domain_models_gen.go")
	if err := os.WriteFile(staleFile, []byte("package model\n"), 0o644); err != nil {
		t.Fatalf("write stale model file: %v", err)
	}

	runGenerate()

	assertNotExists(t,
		staleFile,
		filepath.Join(tmpDir, "graph/model/models_gen.go"),
	)
	assertExists(t,
		filepath.Join(tmpDir, "graph/model/schema_models_gen.go"),
		filepath.Join(tmpDir, "graph/model/iam_auth_models_gen.go"),
		filepath.Join(tmpDir, "graph/model/iam_user_models_gen.go"),
		filepath.Join(tmpDir, "graph/model/iam_rbac_models_gen.go"),
	)
}

func findGatewayRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "gqlgen.yml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate backend/gateway root from %s", dir)
		}
		dir = parent
	}
}

func prepareTempGatewayWorkspace(t *testing.T, gatewayRoot, tmpDir string) {
	t.Helper()

	cfgBytes, err := os.ReadFile(filepath.Join(gatewayRoot, "gqlgen.yml"))
	if err != nil {
		t.Fatalf("read source gqlgen.yml: %v", err)
	}
	updatedCfg := strings.ReplaceAll(
		string(cfgBytes),
		"github.com/integrione/erp-retail/backend/gateway/graph/model",
		"tmp.test/gateway/graph/model",
	)
	if err := os.WriteFile(filepath.Join(tmpDir, "gqlgen.yml"), []byte(updatedCfg), 0o644); err != nil {
		t.Fatalf("write temp gqlgen.yml: %v", err)
	}

	srcGraphQL := filepath.Join(gatewayRoot, "graphql")
	dstGraphQL := filepath.Join(tmpDir, "graphql")
	if err := filepath.WalkDir(srcGraphQL, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcGraphQL, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstGraphQL, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy graphql directory: %v", err)
	}

	modelPkgDir := filepath.Join(tmpDir, "graph/model")
	if err := os.MkdirAll(modelPkgDir, 0o755); err != nil {
		t.Fatalf("mkdir temp graph/model: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(modelPkgDir, "package.go"),
		[]byte("// Package model menampung struct hasil generasi gqlgen.\npackage model\n"),
		0o644,
	); err != nil {
		t.Fatalf("write temp model package.go: %v", err)
	}

	goModBytes, err := os.ReadFile(filepath.Join(gatewayRoot, "go.mod"))
	if err != nil {
		t.Fatalf("read gateway go.mod: %v", err)
	}
	var modLines []string
	skipReplace := false
	for _, line := range strings.Split(string(goModBytes), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			modLines = append(modLines, "module tmp.test/gateway")
			continue
		}
		if strings.HasPrefix(trimmed, "replace (") {
			skipReplace = true
			continue
		}
		if skipReplace {
			if trimmed == ")" {
				skipReplace = false
			}
			continue
		}
		if strings.Contains(trimmed, "github.com/integrione/erp-retail/") {
			continue
		}
		modLines = append(modLines, line)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(strings.Join(modLines, "\n")), 0o644); err != nil {
		t.Fatalf("write temp go.mod: %v", err)
	}

	goSumBytes, err := os.ReadFile(filepath.Join(gatewayRoot, "go.sum"))
	if err != nil {
		t.Fatalf("read gateway go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "go.sum"), goSumBytes, 0o644); err != nil {
		t.Fatalf("write temp go.sum: %v", err)
	}
}

func assertExists(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected file to exist: %s (%v)", p, err)
		}
	}
}

func assertNotExists(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("expected file NOT to exist: %s (err=%v)", p, err)
		}
	}
}

func assertFileDeclaresTypes(t *testing.T, filePath string, expectedTypes ...string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filePath, err)
	}
	var declared []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if ok && ts.Name != nil {
				declared = append(declared, ts.Name.Name)
			}
		}
	}
	for _, want := range expectedTypes {
		if !slices.Contains(declared, want) {
			t.Errorf("%s missing expected type %q (declared=%v)", filepath.Base(filePath), want, declared)
		}
	}
}

func assertFileDeclaresMethod(t *testing.T, filePath, recvType, methodName string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filePath, err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 || fn.Name == nil {
			continue
		}
		if fn.Name.Name != methodName {
			continue
		}
		recv := fn.Recv.List[0].Type
		if star, isStar := recv.(*ast.StarExpr); isStar {
			recv = star.X
		}
		if ident, isIdent := recv.(*ast.Ident); isIdent && ident.Name == recvType {
			return
		}
	}
	t.Errorf("%s missing expected method (%s).%s", filepath.Base(filePath), recvType, methodName)
}

func assertStructDeclaresFields(t *testing.T, filePath, structName string, expectedFields ...string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filePath, err)
	}
	var foundFields []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil || ts.Name.Name != structName {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || st.Fields == nil {
				t.Fatalf("%s.%s is not a struct", filepath.Base(filePath), structName)
			}
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					foundFields = append(foundFields, name.Name)
				}
			}
		}
	}
	for _, want := range expectedFields {
		if !slices.Contains(foundFields, want) {
			t.Errorf("%s.%s missing field %q (fields=%v)", filepath.Base(filePath), structName, want, foundFields)
		}
	}
}
