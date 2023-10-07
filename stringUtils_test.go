package main

import (
	"strconv"
	"testing"
)

func TestStringUtils(t *testing.T) {

	testCases := []map[string]interface{}{
		{
			"params": map[string]interface{}{
				"test": "test",
			},
			"pattern":  "simple {test} for test",
			"expected": "simple test for test",
		},
		{
			"params": map[string]interface{}{
				"number": 10,
			},
			"pattern":  "simple {number} for test",
			"expected": "simple " + strconv.Itoa(10) + " for test",
		},
		{
			"params": map[string]interface{}{
				"number": 12.34,
			},
			"pattern":  "simple {number} for test",
			"expected": "simple " + strconv.FormatFloat(12.34, 'G', -1, 64) + " for test",
		},
	}

	for _, testCase := range testCases {
		actual := Interpolate(testCase["pattern"].(string), testCase["params"].(map[string]interface{}))

		if actual != testCase["expected"].(string) {
			t.Errorf("Interpolation failed expected: '%s' but actual: '%s'", testCase["expected"].(string), actual)
		} else {
			t.Logf(testCase["pattern"].(string) + " test passed.")
		}
	}

}
