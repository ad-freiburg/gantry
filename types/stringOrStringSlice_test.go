package types_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestStringOrStringSliceUnmarshalJSON(t *testing.T) {
	var cases = []struct {
		json   string
		err    string
		result types.StringOrStringSlice
	}{
		{"", "unexpected end of JSON input,", types.StringOrStringSlice{}},
		{"\"A\"", "", types.StringOrStringSlice{"A"}},
		{"[\"A\", \"B\"]", "", types.StringOrStringSlice{"A", "B"}},
		{"[\"A\", \"B\", \"A\"]", "", types.StringOrStringSlice{"A", "B", "A"}},
	}

	for _, c := range cases {
		s := types.StringOrStringSlice{}
		err := s.UnmarshalJSON([]byte(c.json))
		if (err != nil && c.err == "") || (err == nil && c.err != "") {
			t.Errorf("Incorrect error for '%s', got '%s', wanted '%s'", c.json, err, c.err)
		}
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: '%#v', wanted '%#v'", c.json, s, c.result)
		}
	}
}
