package types // import "github.com/ad-freiburg/gantry/types"

import "encoding/json"

// StringOrStringSlice stores a single string or a slice as a slice.
type StringOrStringSlice []string

// UnmarshalJSON sets *r to a copy of data.
func (r *StringOrStringSlice) UnmarshalJSON(data []byte) error {
	var result []string

	err := json.Unmarshal(data, &result)
	if err != nil {
		var value string
		err := json.Unmarshal(data, &value)
		if err != nil {
			return err
		}
		result = []string{value}
	}
	*r = result
	return nil
}
