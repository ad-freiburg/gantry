package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"os"
	"strings"
)

type BuildInfo struct {
	Context    string `json:"context"`
	Dockerfile string `json:"dockerfile"`
	Args       map[string]string
}

func (l *BuildInfo) UnmarshalJSON(data []byte) error {
	result := BuildInfo{}

	storage := map[string]interface{}{}
	err := json.Unmarshal(data, &storage)
	if err != nil {
		return err
	}
	for k, v := range storage {
		switch k {
		case "context":
			result.Context = v.(string)
		case "dockerfile":
			result.Dockerfile = v.(string)
		case "args":
			result.Args = parseBuildArgs(v)
		}
	}
	*l = result
	return nil
}

func parseBuildArgs(data interface{}) map[string]string {
	result := make(map[string]string)
	asSlice, ok := data.([]interface{})
	if ok {
		for _, v := range asSlice {
			parts := strings.SplitN(v.(string), "=", 2)
			if len(parts) < 2 {
				result[parts[0]] = os.Getenv(parts[0])
			} else {
				result[parts[0]] = parts[1]
			}
		}
	} else {
		asMap := data.(map[string]interface{})
		for k, v := range asMap {
			result[k] = v.(string)
		}
	}
	return result
}
