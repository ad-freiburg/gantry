package types_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestStringSetUnmarshalJSON(t *testing.T) {
	var cases = []struct {
		json   string
		result types.StringSet
	}{
		{"", types.StringSet{}},
		{"\"A\"", types.StringSet{"A": true}},
		{"[\"A\", \"B\"]", types.StringSet{"A": true, "B": true}},
		{"[\"A\", \"B\", \"A\"]", types.StringSet{"A": true, "B": true}},
	}

	for _, c := range cases {
		s := types.StringSet{}
		s.UnmarshalJSON([]byte(c.json))
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: %#v, wanted %#v", c.json, s, c.result)
		}
	}
}
