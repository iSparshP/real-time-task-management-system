package common

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if the provided email is valid
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidateRequired checks if the provided string is not empty
func ValidateRequired(str string) bool {
	return strings.TrimSpace(str) != ""
}

// ValidateLength checks if the string length is within the specified range
func ValidateLength(str string, min, max int) bool {
	length := len(strings.TrimSpace(str))
	return length >= min && length <= max
}

// ValidateRange checks if the number is within the specified range
func ValidateRange(num, min, max int) bool {
	return num >= min && num <= max
}
