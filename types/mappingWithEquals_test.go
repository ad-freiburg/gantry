package types_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestMappingWithEqualsUnmarshalJSON(t *testing.T) {
	bar := "Bar"
	var cases = []struct {
		json   string
		err    string
		result types.MappingWithEquals
	}{
		{"", "unexpected end of JSON input,", types.MappingWithEquals{}},
		{"{\"Foo\": \"Bar\"}", "", types.MappingWithEquals{"Foo": &bar}},
		{"[\"Foo=Bar\"]", "", types.MappingWithEquals{"Foo": &bar}},
		{"[\"Foo\"]", "", types.MappingWithEquals{"Foo": nil}},
	}

	for _, c := range cases {
		s := types.MappingWithEquals{}
		err := s.UnmarshalJSON([]byte(c.json))
		if (err != nil && c.err == "") || (err == nil && c.err != "") {
			t.Errorf("Incorrect error for '%s', got '%s', wanted '%s'", c.json, err, c.err)
		}
		if !reflect.DeepEqual(s, c.result) {
			t.Errorf("Incorrect result for '%s', got: '%#v', wanted '%#v'", c.json, s, c.result)
		}
	}
}
