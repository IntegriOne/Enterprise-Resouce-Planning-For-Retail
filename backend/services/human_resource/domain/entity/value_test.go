package entity_test

import (
	"encoding/json"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	"testing"
)

func TestDecimalPrecision(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"99999999999999.123456", "0.00000001", "-1234567890123456.12345678", "0"} {
		t.Run(raw, func(t *testing.T) {
			var d entity.Decimal
			if err := d.Scan([]byte(raw)); err != nil {
				t.Fatal(err)
			}
			v, err := d.Value()
			if err != nil || v != raw {
				t.Fatalf("decimal round trip: %v %v", v, err)
			}
		})
	}
	for _, raw := range []string{"", "NaN", "Infinity", "1);DROP TABLE hr_workers;--"} {
		t.Run("reject_"+raw, func(t *testing.T) {
			if _, err := entity.Decimal(raw).Value(); err == nil {
				t.Fatal("expected invalid decimal")
			}
		})
	}
	var d entity.Decimal
	if err := d.Scan(1.5); err == nil {
		t.Fatal("float conversion must not silently lose precision")
	}
}

func TestJSONNullAndBuffer(t *testing.T) {
	t.Parallel()
	var j entity.JSON
	if err := j.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if v, err := j.Value(); err != nil || v != nil {
		t.Fatalf("sql null: %v %v", v, err)
	}
	raw := []byte(`{"valid":true}`)
	if err := j.Scan(raw); err != nil {
		t.Fatal(err)
	}
	raw[2] = 'x'
	encoded, err := json.Marshal(j)
	if err != nil || string(encoded) != `{"valid":true}` {
		t.Fatalf("json buffer ownership: %s %v", encoded, err)
	}
	if err := json.Unmarshal([]byte(`null`), &j); err != nil {
		t.Fatal(err)
	}
	if v, err := j.Value(); err != nil || v != "null" {
		t.Fatalf("json null: %v %v", v, err)
	}
	if _, err := entity.JSON(`{`).Value(); err == nil {
		t.Fatal("invalid json accepted")
	}
}
