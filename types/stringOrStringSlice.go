package types // import "github.com/ad-freiburg/gantry/types"

import "encoding/json"

type StringOrStringSlice []string

func (l *StringOrStringSlice) UnmarshalJSON(data []byte) error {
	result := make([]string, 0)

	err := json.Unmarshal(data, &result)
	if err != nil {
		var value string
		err := json.Unmarshal(data, &value)
		if err != nil {
			return err
		}
		result = []string{value}
	}
	*l = result
	return nil
}
