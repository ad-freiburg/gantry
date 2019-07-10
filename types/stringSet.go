// Package types implements types which can be unmarshaled from json.
package types // import "github.com/ad-freiburg/gantry/types"

import "encoding/json"

// StringSet stores a list of strings as a map of bools.
type StringSet map[string]bool

// UnmarshalJSON sets *r to a copy of data.
func (r *StringSet) UnmarshalJSON(data []byte) error {
	result := make(map[string]bool, 0)

	var parsedJson []string
	err := json.Unmarshal(data, &parsedJson)
	if err == nil {
		for _, s := range parsedJson {
			result[s] = true
		}
	} else {
		var value string
		err := json.Unmarshal(data, &value)
		if err != nil {
			return err
		}
		result[value] = true
	}
	*r = result
	return nil
}
