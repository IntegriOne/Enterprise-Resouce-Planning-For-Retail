package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateDimensionAnchors memastikan setiap tingkat rubrik memiliki bukti teramati.
func ValidateDimensionAnchors(value entity.JSON) error {
	var anchors map[string]string
	if err := json.Unmarshal(value, &anchors); err != nil || len(anchors) != 5 {
		return hrerrors.Invalid("rating_anchors", "must define evidence for exactly five rating levels")
	}
	for level := 1; level <= 5; level++ {
		if strings.TrimSpace(anchors[fmt.Sprint(level)]) == "" {
			return hrerrors.Invalid("rating_anchors", "all five rating anchors are required")
		}
	}
	return nil
}
