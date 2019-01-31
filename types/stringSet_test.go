package types // import "github.com/ad-freiburg/gantry/types"

import (
	"reflect"
	"testing"
)

func TestStringSetUnmarshalJSON(t *testing.T) {
	var cases = []struct {
		json   string
		result StringSet
	}{
		{"", StringSet{}},
		{"\"A\"", StringSet{"A": true}},
		{"[\"A\", \"B\"]", StringSet{"A": true, "B": true}},
	}

	for _, c := range cases {
		s := StringSet{}
		s.UnmarshalJSON([]byte(c.json))
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: %#v, wanted %#v", c.json, s, c.result)
		}
	}
}
