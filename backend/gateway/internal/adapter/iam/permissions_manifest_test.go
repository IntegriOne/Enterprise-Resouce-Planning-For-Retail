package iam

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/shared/authz"
	"github.com/integrione/erp-retail/backend/shared/rbac"
)

func TestIAMPermissionsAndGuardsMatchManifestCatalog(t *testing.T) {
	t.Parallel()

	catalogs, err := rbac.LoadEmbeddedCatalogs()
	if err != nil {
		t.Fatalf("LoadEmbeddedCatalogs() error = %v", err)
	}
	wantCodes := catalogs.PermissionCodesForDomain("iam")

	// 1. Pastikan daftar konstanta kanonik di shared/authz persis 1:1 dengan manifest iam.json.
	authzCodes := []string{
		authz.PermDashboardView,
		authz.PermUsersRead,
		authz.PermUsersWrite,
		authz.PermUsersReadSelf,
		authz.PermUsersUpdateSelf,
		authz.PermUsersDeleteSelf,
		authz.PermRolesRead,
		authz.PermRolesWrite,
		authz.PermGroupsRead,
		authz.PermGroupsWrite,
		authz.PermPermissionsRead,
		authz.PermSessionsRead,
		authz.PermSessionsReadSelf,
		authz.PermSessionsRevoke,
		authz.PermSessionsDeleteSelf,
	}
	slices.Sort(authzCodes)
	if !slices.Equal(authzCodes, wantCodes) {
		t.Fatalf("authz IAM constants drifted from shared/rbac iam.json:\ngot  = %v\nwant = %v", authzCodes, wantCodes)
	}

	// 2. Pastikan setiap literal kode permission dan argumen requireSensitive di adapter produksi IAM selaras dengan iam.json.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, wd, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("ParseDir(%s) error = %v", wd, err)
	}

	checkedCalls := 0
	for _, pkg := range pkgs {
		for fileName, fileAST := range pkg.Files {
			ast.Inspect(fileAST, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				switch sel.Sel.Name {
				case "requirePerm", "requireSelf":
					if len(call.Args) >= 3 {
						if code := extractStringLiteral(call.Args[1]); code != "" {
							spec, exists := catalogs.PermissionByCode[code]
							if !exists {
								t.Errorf("%s: adapter references undeclared permission %q", filepath.Base(fileName), code)
								return true
							}
							if bLit, ok := call.Args[2].(*ast.Ident); ok && (bLit.Name == "true" || bLit.Name == "false") {
								wantSensitive := bLit.Name == "true"
								if spec.IsSensitive && !wantSensitive {
									t.Errorf("%s: sensitive permission %q must require step-up reauth (true), got false", filepath.Base(fileName), code)
								}
							}
							checkedCalls++
						}
					}
				case "HasPermission":
					if len(call.Args) == 1 {
						if code := extractStringLiteral(call.Args[0]); code != "" {
							if _, exists := catalogs.PermissionByCode[code]; !exists {
								t.Errorf("%s: HasPermission checks undeclared permission %q", filepath.Base(fileName), code)
							}
							checkedCalls++
						}
					}
				}
				return true
			})
		}
	}
	if checkedCalls == 0 {
		t.Fatal("expected to inspect at least one permission guard call in IAM Gateway adapter")
	}
}

func extractStringLiteral(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return val
}
