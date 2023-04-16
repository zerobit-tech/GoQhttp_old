package stringutils

import (
	"regexp"
	"strings"
)

func IsNumericWithOutDecimal(word string) bool {
	if strings.TrimSpace(word) == "" {
		return false
	}
	return regexp.MustCompile(`\d`).MatchString(word)
	// calling regexp.MustCompile() function to create the regular expression.
	// calling MatchString() function that returns a bool that
	// indicates whether a pattern is matched by the string.
}

func IsNumeric(word string) bool {
	if strings.TrimSpace(word) == "" {
		return false
	}
	return regexp.MustCompile(`\d*\.?\d*`).MatchString(word)
	// calling regexp.MustCompile() function to create the regular expression.
	// calling MatchString() function that returns a bool that
	// indicates whether a pattern is matched by the string.
}
