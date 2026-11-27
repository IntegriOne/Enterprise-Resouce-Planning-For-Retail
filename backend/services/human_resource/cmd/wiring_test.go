package main

import (
	"sort"
	"strings"
	"testing"

	efintegration "github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/integration/enterprise_foundation"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBuildMutationValidatorsCoversAllCatalogMutatableResources(t *testing.T) {
	t.Parallel()
	offlineDB := &gorm.DB{Config: &gorm.Config{Dialector: postgres.New(postgres.Config{})}}
	repos, err := repository.New(offlineDB)
	if err != nil {
		t.Fatalf("initialize offline repositories: %v", err)
	}
	validators, err := buildMutationValidators(repos, efintegration.NewDisabledClient())
	if err != nil {
		t.Fatalf("buildMutationValidators: %v", err)
	}
	for resource, validator := range validators {
		if validator == nil {
			t.Fatalf("registered nil validator for %s", resource)
		}
	}

	catalog := security.PermissionCatalog()
	mutatable := map[string]struct{}{}
	readOnly := map[string]struct{}{}
	for op := range catalog {
		// Slip adalah resource virtual. Command cetak ulang memvalidasi hasil payroll sebelum menautkan PDF.
		if op == "hr_payslips.reissue" {
			continue
		}
		parts := strings.SplitN(op, ".", 2)
		if len(parts) != 2 {
			t.Fatalf("unexpected operation format %q", op)
		}
		resource, action := parts[0], parts[1]
		switch action {
		case "read", "list", "read_self":
			if _, ok := readOnly[resource]; !ok {
				readOnly[resource] = struct{}{}
			}
		default:
			mutatable[resource] = struct{}{}
		}
	}
	// Tabel buku besar cuti divalidasi melalui post_leave_entry meski tidak memiliki permission CRUD generik.
	mutatable["hr_leave_ledger_entries"] = struct{}{}

	var missing []string
	for resource := range mutatable {
		if _, ok := validators[resource]; !ok {
			missing = append(missing, resource)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("unvalidated mutatable resources (%d): %v", len(missing), missing)
	}

	expectedCount := len(mutatable)
	if len(validators) != expectedCount {
		t.Fatalf("validator count = %d, want %d", len(validators), expectedCount)
	}
}

func TestRegisterAllMatchesHRMProtoServices(t *testing.T) {
	t.Parallel()
	offlineDB := &gorm.DB{Config: &gorm.Config{Dialector: postgres.New(postgres.Config{})}}
	repos, err := repository.New(offlineDB)
	if err != nil {
		t.Fatalf("initialize offline repositories: %v", err)
	}
	services, err := appservice.NewRegistry(repos, operation.Dependencies{
		Policy: &security.Policy{}, Integrity: security.NewIntegrity(nil),
	})
	if err != nil {
		t.Fatalf("initialize service registry: %v", err)
	}
	server := grpc.NewServer()
	t.Cleanup(server.Stop)
	if err := handler.RegisterAll(server, services); err != nil {
		t.Fatalf("register HRM services: %v", err)
	}

	registered := server.GetServiceInfo()
	serviceCount := 0
	protoregistry.GlobalFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(string(file.Package()), "human_resource.") {
			return true
		}
		if strings.HasPrefix(string(file.Package()), "human_resource.workload.") {
			for i := 0; i < file.Services().Len(); i++ {
				if _, ok := registered[string(file.Services().Get(i).FullName())]; ok {
					t.Errorf("workload service registered on business listener: %s", file.Services().Get(i).FullName())
				}
			}
			return true
		}

		for i := 0; i < file.Services().Len(); i++ {
			service := file.Services().Get(i)
			serviceCount++
			info, ok := registered[string(service.FullName())]
			if !ok {
				t.Errorf("missing service %s", service.FullName())
				continue
			}
			methods := make(map[string]grpc.MethodInfo, len(info.Methods))
			for _, method := range info.Methods {
				methods[method.Name] = method
			}
			if len(methods) != service.Methods().Len() {
				t.Errorf("%s registered %d methods, want %d", service.FullName(), len(methods), service.Methods().Len())
			}
			for j := 0; j < service.Methods().Len(); j++ {
				method := service.Methods().Get(j)
				info, ok := methods[string(method.Name())]
				if !ok || info.IsClientStream != method.IsStreamingClient() || info.IsServerStream != method.IsStreamingServer() {
					t.Errorf("%s is missing or has incompatible streaming flags", method.FullName())
				}
			}
		}
		return true
	})
	if serviceCount < 22 || serviceCount != len(registered) {
		t.Errorf("found %d HRM proto services and %d registrations, expected matching counts of at least 22", serviceCount, len(registered))
	}
}
