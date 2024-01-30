package virtualcluster

import (
	"fmt"
	"strings"
)

// stringSliceToMap converts key=value[,value] into map[string][]string.
func stringSliceToMap(strs []string) (map[string][]string, error) {
	res := make(map[string][]string, len(strs))
	for _, str := range strs {
		key, valuesInStr, ok := strings.Cut(str, "=")
		if !ok {
			return nil, fmt.Errorf("expected key=value[,value] format, but got %s", str)
		}
		values := strings.Split(valuesInStr, ",")
		res[key] = values
	}
	return res, nil
}
