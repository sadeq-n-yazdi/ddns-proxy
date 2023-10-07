package main

import (
	"fmt"
	"net/http"
)

// MergeMaps merges all maps and use the fist appear record for the result
func MergeMaps(reverseOrderMaps []map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Iterate through the array of maps in reverse order
	for i := len(reverseOrderMaps) - 1; i >= 0; i-- {
		currentMap := reverseOrderMaps[i]

		// Merge the current map into the merged map
		for key, value := range currentMap {
			merged[key] = value
		}
	}

	return merged
}

func GetParamsAsMap(r *http.Request) (result *map[string]interface{}, err error) {
	// Parse query parameters from the URL
	queryParams := r.URL.Query()

	// Parse form data from the request body
	err = r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("error parsing form data: %s", err)
	}
	formData := r.Form

	// Combine query parameters and form data into a single map
	resMap := make(map[string]interface{})
	result = &resMap
	for key, values := range queryParams {
		if len(values) > 0 {
			resMap[key] = values[0]
		}
	}
	for key, values := range formData {
		if len(values) > 0 {
			resMap[key] = values[0]
		}
	}

	return &resMap, nil

}
