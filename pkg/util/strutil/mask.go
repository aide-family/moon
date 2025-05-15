package strutil

import "unicode/utf8"

// MaskString masks the middle part of a string with exactly 6 asterisks (*), keeping the specified number
// of characters visible at the start and end. If the string is too short to mask, it will be padded with 6 asterisks.
// Parameters:
//   - s: the original string to mask
//   - start: number of characters to keep visible at the start
//   - end: number of characters to keep visible at the end
//
// Example:
//
//	MaskString("1234567890", 3, 2) returns "123******90"
//	MaskString("abc@example.com", 3, 4) returns "abc******e.com"
func MaskString(s string, start, end int) string {
	if s == "" {
		return s
	}

	// Get string length in runes to handle UTF-8 correctly
	length := utf8.RuneCountInString(s)

	// Handle special cases
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}

	// Convert string to runes for proper UTF-8 handling
	runes := []rune(s)

	// If the string is too short, pad with 6 asterisks
	if start+end >= length {
		return s + "******"
	}

	// Create the masked string with exactly 6 asterisks
	result := make([]rune, 0, start+6+end)
	result = append(result, runes[:start]...)      // Copy start characters
	result = append(result, []rune("******")...)   // Add exactly 6 asterisks
	result = append(result, runes[length-end:]...) // Copy end characters

	return string(result)
}

// MaskEmail masks the local part of an email address, keeping the first and last character visible,
// with exactly 6 asterisks in between.
// Example: "user@example.com" becomes "u******r@example.com"
func MaskEmail(email string) string {
	if email == "" {
		return email
	}

	atIndex := -1
	for i, r := range email {
		if r == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 0 {
		return email + "******"
	}

	localPart := email[:atIndex]
	domain := email[atIndex:]

	// Mask local part keeping first and last character
	maskedLocal := MaskString(localPart, 1, 1)
	return maskedLocal + domain
}

// MaskPhone masks the middle part of a phone number, keeping the first 3 and last 4 digits visible,
// with exactly 6 asterisks in between.
// Example: "13812345678" becomes "138******5678"
func MaskPhone(phone string) string {
	return MaskString(phone, 3, 4)
}

// MaskBankCard masks the middle part of a bank card number, keeping the first 4 and last 4 digits visible,
// with exactly 6 asterisks in between.
// Example: "6225123412341234" becomes "6225******1234"
func MaskBankCard(cardNumber string) string {
	return MaskString(cardNumber, 4, 4)
}
