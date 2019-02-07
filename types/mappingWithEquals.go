package types

import (
	"encoding/json"
	"strings"
)

type MappingWithEquals map[string]*string

func (t *MappingWithEquals) UnmarshalJSON(data []byte) error {
	result := map[string]*string{}

	err := json.Unmarshal(data, &result)
	if err != nil {
		storage := []string{}
		err := json.Unmarshal(data, &storage)
		if err != nil {
			return err
		}
		for _, v := range storage {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 1 {
				result[parts[0]] = nil
			} else {
				result[parts[0]] = &parts[1]
			}
		}
	}
	*t = result
	return nil
}
