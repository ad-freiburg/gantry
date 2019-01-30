package types // import "github.com/ad-freiburg/gantry/types"

import (
	"encoding/json"
	"fmt"
)

type StringMapOrStringSlice []string

func (l *StringMapOrStringSlice) UnmarshalJSON(data []byte) error {
	result := make([]string, 0)

	storage := make(map[string]string)
	err := json.Unmarshal(data, &storage)
	if err == nil {
		for k, v := range storage {
			result = append(result, fmt.Sprintf("%s=%s", k, v))
		}
	} else {
		err := json.Unmarshal(data, &result)
		if err != nil {
			return err
		}
	}
	*l = result
	return nil
}
