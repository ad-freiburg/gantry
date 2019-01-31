package types_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestStringOrStringSliceUnmarshalJSON(t *testing.T) {
	var cases = []struct {
		json   string
		result types.StringOrStringSlice
	}{
		{"", types.StringOrStringSlice{}},
		{"\"A\"", types.StringOrStringSlice{"A"}},
		{"[\"A\", \"B\"]", types.StringOrStringSlice{"A", "B"}},
		{"[\"A\", \"B\", \"A\"]", types.StringOrStringSlice{"A", "B", "A"}},
	}

	for _, c := range cases {
		s := types.StringOrStringSlice{}
		s.UnmarshalJSON([]byte(c.json))
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: %#v, wanted %#v", c.json, s, c.result)
		}
	}
}
