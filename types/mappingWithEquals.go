package types

import (
	"encoding/json"
	"strings"
)

// MappingWithEquals stores a list of key=value or key: value as a map.
type MappingWithEquals map[string]*string

// UnmarshalJSON sets *r to a copy of data.
func (r *MappingWithEquals) UnmarshalJSON(data []byte) error {
	result := map[string]*string{}

	err := json.Unmarshal(data, &result)
	if err != nil {
		parsedJson := []string{}
		err := json.Unmarshal(data, &parsedJson)
		if err != nil {
			return err
		}
		for _, v := range parsedJson {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 1 {
				result[parts[0]] = nil
			} else {
				result[parts[0]] = &parts[1]
			}
		}
	}
	*r = result
	return nil
}
