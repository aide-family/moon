package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MatchStatusCodes checks if the given status code matches the patterns.
func MatchStatusCodes(patterns string, statusCode int) bool {
	// Split the patterns by commas
	patternList := strings.Split(patterns, ",")

	// Iterate through each pattern
	for _, pattern := range patternList {
		// Replace 'x' with regex pattern to match any digit
		regexPattern := strings.ReplaceAll(pattern, "x", "\\d")

		// Ensure the regex matches the full string
		regexPattern = "^" + regexPattern + "$"

		// Convert the status code to a string
		statusCodeStr := strconv.Itoa(statusCode)

		// Compile and match the regex
		matched, err := regexp.MatchString(regexPattern, statusCodeStr)
		if err != nil {
			fmt.Printf("Error compiling regex: %v\n", err)
			return false
		}
		if matched {
			return true
		}
	}

	return false
}
