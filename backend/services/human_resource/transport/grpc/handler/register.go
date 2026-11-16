// Package handler mendaftarkan seluruh domain HRM ke transport privat.
package handler

import (
	"fmt"

	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/accessrequest"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/analytics"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/benefits"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/career"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/compensation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/competency"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/employeerelations"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/engagement"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/hrservice"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/learning"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/leave"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/offboarding"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/onboarding"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/performance"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/policy"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/promotion"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/succession"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/talent"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/workforceplanning"
	"google.golang.org/grpc"
)

// RegisterAll gagal bila salah satu domain tidak dapat didaftarkan.
func RegisterAll(registrar grpc.ServiceRegistrar, services *service.Registry) error {
	if operation.RequireDependency(registrar) != nil || services == nil {
		return fmt.Errorf("HRM registrar and services are required")
	}
	registrations := []struct {
		name     string
		register func() error
	}{
		{"workforceplanning", func() error { return workforceplanning.Register(registrar, services.WorkforcePlanning) }},
		{"personnel", func() error { return personnel.Register(registrar, services.Personnel) }},
		{"recruitment", func() error { return recruitment.Register(registrar, services.Recruitment) }},
		{"onboarding", func() error { return onboarding.Register(registrar, services.Onboarding) }},
		{"compensation", func() error { return compensation.Register(registrar, services.Compensation) }},
		{"benefits", func() error { return benefits.Register(registrar, services.Benefits) }},
		{"payroll", func() error { return payroll.Register(registrar, services.Payroll) }},
		{"leave", func() error { return leave.Register(registrar, services.Leave) }},
		{"performance", func() error { return performance.Register(registrar, services.Performance) }},
		{"learning", func() error { return learning.Register(registrar, services.Learning) }},
		{"competency", func() error { return competency.Register(registrar, services.Competency) }},
		{"succession", func() error { return succession.Register(registrar, services.Succession) }},
		{"employeerelations", func() error { return employeerelations.Register(registrar, services.EmployeeRelations) }},
		{"analytics", func() error { return analytics.Register(registrar, services.Analytics) }},
		{"offboarding", func() error { return offboarding.Register(registrar, services.Offboarding) }},
		{"career", func() error { return career.Register(registrar, services.Career) }},
		{"policy", func() error { return policy.Register(registrar, services.Policy) }},
		{"hrservice", func() error { return hrservice.Register(registrar, services.HRService) }},
		{"engagement", func() error { return engagement.Register(registrar, services.Engagement) }},
		{"accessrequest", func() error { return accessrequest.Register(registrar, services.AccessRequests) }},
		{"talent", func() error { return talent.Register(registrar, services.Talent) }},
		{"promotion", func() error { return promotion.Register(registrar, services.Promotion) }},
	}
	for _, registration := range registrations {
		if err := registration.register(); err != nil {
			return fmt.Errorf("register %s: %w", registration.name, err)
		}
	}
	return nil
}
