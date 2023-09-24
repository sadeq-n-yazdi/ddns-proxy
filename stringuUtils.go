package main

import (
	"fmt"
	"regexp"
)

// Interpolate replaces placeholders in a string with values from a map
func Interpolate(format string, data map[string]interface{}) string {
	re := regexp.MustCompile(`{([^}]+)}`)
	result := re.ReplaceAllStringFunc(format, func(match string) string {
		key := match[1 : len(match)-1] // Remove curly braces
		if value, ok := data[key]; ok {
			return fmt.Sprintf("%v", value)
		}
		return match // Placeholder not found in map, return unchanged
	})
	return result
}
