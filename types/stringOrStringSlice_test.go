package types // import "github.com/ad-freiburg/gantry/types"

import (
	"reflect"
	"testing"
)

func TestStringOrStringSliceUnmarshalJSON(t *testing.T) {
	var cases = []struct {
		json   string
		result StringOrStringSlice
	}{
		{"", StringOrStringSlice{}},
		{"\"A\"", StringOrStringSlice{"A"}},
		{"[\"A\", \"B\"]", StringOrStringSlice{"A", "B"}},
		{"[\"A\", \"B\", \"A\"]", StringOrStringSlice{"A", "B", "A"}},
	}

	for _, c := range cases {
		s := StringOrStringSlice{}
		s.UnmarshalJSON([]byte(c.json))
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: %#v, wanted %#v", c.json, s, c.result)
		}
	}
}
