package types_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestStringMapUnmarshalJSON(t *testing.T) {
	bar := "Bar"
	var cases = []struct {
		json   string
		err    string
		result types.StringMap
	}{
		{"", "unexpected end of JSON input,", types.StringMap{}},
		{"{\"Foo\": \"Bar\"}", "", types.StringMap{"Foo": &bar}},
		{"[\"Foo=Bar\"]", "", types.StringMap{"Foo": &bar}},
		{"[\"Foo\"]", "", types.StringMap{"Foo": nil}},
	}

	for _, c := range cases {
		s := types.StringMap{}
		err := s.UnmarshalJSON([]byte(c.json))
		if (err != nil && c.err == "") || (err == nil && c.err != "") {
			t.Errorf("Incorrect error for '%s', got '%s', wanted '%s'", c.json, err, c.err)
		}
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: '%#v', wanted '%#v'", c.json, s, c.result)
		}
	}
}
