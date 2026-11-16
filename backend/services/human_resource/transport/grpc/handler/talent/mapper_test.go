package talent

import (
	"bytes"
	"testing"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	"google.golang.org/protobuf/proto"
)

func TestDimensionJSONAndDecimalRemainExactAcrossWire(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"threshold":9007199254740993123456789,"weight":0.1234567890123456789012345}`)
	value, err := dimension(&entity.TalentDimension{RatingAnchors: entity.JSON(raw), WeightPercent: entity.Decimal("99.999")})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := proto.Marshal(value.RatingAnchors)
	if err != nil {
		t.Fatal(err)
	}
	decoded := &commonv1.JsonDocument{}
	if err := proto.Unmarshal(encoded, decoded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, decoded.Utf8) || value.WeightPercent != "99.999" {
		t.Fatal("JSON or decimal precision lost")
	}
	value, err = dimension(&entity.TalentDimension{})
	if err != nil || value.RatingAnchors != nil {
		t.Fatal("SQL NULL must remain absent")
	}
	value, err = dimension(&entity.TalentDimension{RatingAnchors: entity.JSON("null")})
	if err != nil || value.RatingAnchors == nil || string(value.RatingAnchors.Utf8) != "null" {
		t.Fatal("JSON null must remain present")
	}
	if _, err = dimension(&entity.TalentDimension{RatingAnchors: entity.JSON("{")}); err == nil {
		t.Fatal("invalid stored JSON accepted")
	}
}
func TestScoreRangeRejectsNarrowingOverflow(t *testing.T) {
	t.Parallel()
	for _, value := range []int32{-32769, 32768, 65537} {
		if _, err := score(&value); err == nil {
			t.Fatalf("overflow accepted: %d", value)
		}
	}
	for _, value := range []int32{-32768, 0, 32767} {
		result, err := score(&value)
		if err != nil || result == nil || int32(*result) != value {
			t.Fatalf("boundary %d: %v %v", value, result, err)
		}
	}
}
