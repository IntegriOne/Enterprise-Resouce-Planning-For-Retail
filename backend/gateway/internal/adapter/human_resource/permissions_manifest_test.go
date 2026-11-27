package human_resource

import (
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	"github.com/integrione/erp-retail/backend/shared/rbac"
)

func TestHRMPermissionsInManifestMatchGRPCPolicy(t *testing.T) {
	t.Parallel()

	catalogs, err := rbac.LoadEmbeddedCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	grpcPermissions := security.PermissionCatalog()
	for _, code := range catalogs.PermissionCodesForDomain("human_resource") {
		permission := catalogs.PermissionByCode[code]
		resource := "hr_" + strings.TrimPrefix(permission.Resource, "human_resource.")
		mapped := false
		for key, actualCode := range grpcPermissions {
			if actualCode != code {
				continue
			}
			actualResource, action, _ := strings.Cut(key, ".")
			if actualResource != resource && (actualResource != "hr_promotion_recommendations" || resource != "hr_promotion_decisions") {
				t.Errorf("IAM permission %s uses unexpected gRPC resource %s", code, actualResource)
			}
			mapped = true
			if sensitive := security.IsSensitiveOperation(actualResource, action); permission.IsSensitive != sensitive {
				t.Errorf("IAM permission %s sensitivity = %t, gRPC action %s requires %t", code, permission.IsSensitive, action, sensitive)
			}
		}
		if !mapped {
			t.Errorf("IAM permission %s has no gRPC policy mapping", code)
		}
	}
}
