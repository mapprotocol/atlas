package ethconfig

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/naoina/toml"
)

func TestRangeLimitConfig(t *testing.T) {
	if Defaults.RangeLimit != 0 {
		t.Fatalf("expected default range limit 0, got %d", Defaults.RangeLimit)
	}

	const want = uint64(1234)
	settings := toml.Config{
		NormFieldName: func(_ reflect.Type, key string) string { return key },
		FieldToKey:    func(_ reflect.Type, field string) string { return field },
	}
	config := Config{RangeLimit: want}
	blob, err := settings.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if !bytes.Contains(blob, []byte("RangeLimit = 1234")) {
		t.Fatalf("marshaled config does not contain range limit: %s", blob)
	}
	var got Config
	if err := settings.Unmarshal([]byte("RangeLimit = 1234\n"), &got); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}
	if got.RangeLimit != want {
		t.Fatalf("expected range limit %d, got %d", want, got.RangeLimit)
	}
}
