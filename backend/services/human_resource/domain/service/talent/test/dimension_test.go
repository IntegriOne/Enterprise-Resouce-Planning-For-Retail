package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/talent"
)

func TestValidateDimensionAnchors(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		anchors entity.JSON
		valid   bool
	}{
		{name: "five observed levels", anchors: entity.JSON(`{"1":"basic","2":"developing","3":"capable","4":"strong","5":"expert"}`), valid: true},
		{name: "missing level", anchors: entity.JSON(`{"1":"basic","2":"developing"}`)},
		{name: "blank level", anchors: entity.JSON(`{"1":"basic","2":"developing","3":" ","4":"strong","5":"expert"}`)},
		{name: "extra level", anchors: entity.JSON(`{"1":"basic","2":"developing","3":"capable","4":"strong","5":"expert","6":"extra"}`)},
		{name: "wrong value type", anchors: entity.JSON(`{"1":1,"2":"developing","3":"capable","4":"strong","5":"expert"}`)},
		{name: "json null", anchors: entity.JSON(`null`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := service.ValidateDimensionAnchors(test.anchors)
			if test.valid && err != nil {
				t.Fatalf("valid anchors: %v", err)
			}
			if !test.valid && !errors.Is(err, hrerrors.ErrInvalidInput) {
				t.Fatalf("invalid anchors: %v", err)
			}
		})
	}
}
