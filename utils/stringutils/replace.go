package stringutils

import (
	"regexp"
)

func RemoveMultipleSpaces(str string) string {

	var re = regexp.MustCompile(`(?m)\s{2,}`)

	var substitution = " "

	return re.ReplaceAllString(str, substitution)
}

func RemoveSpecialChars(str string) string {

	return regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(str, "_")
}
