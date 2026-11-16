package service_test

import (
	"errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"testing"
)

func TestMissingDependenciesFailClosed(t *testing.T) {
	_, err := service.NewResource[entity.Worker, string, *entity.Worker](nil, service.Dependencies{}, service.Specification[entity.Worker, string]{})
	if !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing dependencies: %v", err)
	}
}
