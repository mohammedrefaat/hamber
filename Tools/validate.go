package tools

import (
	"regexp"
)

func ValidatePhoneNumber(phoneNumber *string) bool {
	if phoneNumber == nil {
		return false
	}
	// Regular expression pattern for phone number validation
	pattern := `^\+?[1-9]\d{1,14}$`
	// Compile the pattern
	regex := regexp.MustCompile(pattern)
	// Match the phone number against the pattern
	match := regex.MatchString(Default(phoneNumber))
	return match
}

func ValidatePhoneNumberEg(phoneNumber *string) bool {
	if phoneNumber == nil {
		return false
	}
	// Regular expression pattern for phone number validation
	pattern := `^(011|012|010|015)\d{8}$`
	// Compile the pattern
	regex := regexp.MustCompile(pattern)
	// Match the phone number against the pattern
	match := regex.MatchString(Default(phoneNumber))
	return match
}

// validateEmail checks if the email format is valid using regex
func ValidateEmail(email *string) bool {
	if email == nil {
		return false
	}
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	// Match the email against the pattern
	match := re.MatchString(Default(email))
	return match
}
