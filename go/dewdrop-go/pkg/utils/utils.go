package utils

import (
	"bytes"
	"regexp"
)

// IsValidResponse checks if Arduino acks with an "a" or gives data in a valid format
func IsValidResponse(response string) bool {
	return response != "" && (IsValidDataFormat(response) || response == "a")
}

// IsValidDataFormat checks if Arduino returns data in a valid format
func IsValidDataFormat(input string) bool {
	pattern := `^\d{1,20},\d{2}\.\d{2},\d{2}\.\d{2},[01]$`
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}

// SplitAndTrim is a helper function for parsing an Arduino data response
func SplitAndTrim(s string, sep rune) []string {
	var result []string
	for _, part := range bytes.Split([]byte(s), []byte{byte(sep)}) {
		result = append(result, string(bytes.TrimSpace(part)))
	}
	return result
}
