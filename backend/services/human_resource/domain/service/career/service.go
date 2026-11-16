// Package service memeriksa jalur dan aspirasi karier tanpa mengubah jabatan resmi.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateAspiration memeriksa target aspirasi terhadap transition dan path yang dipilih.
func ValidateAspiration(aspiration *entity.WorkerCareerAspiration, transition *entity.CareerPathTransition, path *entity.CareerPath) error {
	for _, record := range []entity.Validatable{aspiration, transition, path} {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	if aspiration.CareerPathTransitionID == nil || !strings.EqualFold(*aspiration.CareerPathTransitionID, transition.ID) || !strings.EqualFold(transition.CareerPathID, path.ID) || !strings.EqualFold(aspiration.TargetJobID, transition.ToJobID) {
		return hrerrors.Invalid("career_path_transition_id", "career path and target job must match aspiration")
	}
	return nil
}
